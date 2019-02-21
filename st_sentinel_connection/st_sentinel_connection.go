package st_sentinel_connection

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"
)

type Get_master_addr_reply struct {
	reply string
	err   error
}

type Sentinel_connection struct {
	sentinels_addresses              []string
	sentinel_dial_timeout            time.Duration
	sentinel_command_timeout         time.Duration
	current_sentinel_connection      net.Conn
	reader                           *bufio.Reader
	writer                           *bufio.Writer
	get_master_address_by_name_reply chan *Get_master_addr_reply
	get_master_address_by_name       chan string
}

const (
	client_closed     = true
	client_not_closed = false
)

func (c *Sentinel_connection) parseResponse() (request []string, err error, is_client_closed bool) {
	var ret []string
	buf, _, e := c.reader.ReadLine()
	if e != nil {
		return nil, fmt.Errorf("failed read line from client: %v", e), client_closed
	}
	if len(buf) == 0 {
		return nil, errors.New("failed read line from client"), client_closed
	}
	if buf[0] != '*' {
		return nil, errors.New("first char in mbulk is not *"), client_not_closed
	}
	mbulk_size, _ := strconv.Atoi(string(buf[1:]))
	if mbulk_size == -1 {
		return nil, errors.New("null request"), client_not_closed
	}
	ret = make([]string, mbulk_size)
	for i := 0; i < mbulk_size; i++ {
		buf1, _, e1 := c.reader.ReadLine()
		if e1 != nil {
			return nil, fmt.Errorf("failed read line from client: %v", e1), client_closed
		}
		if len(buf1) == 0 {
			return nil, errors.New("failed read line from client"), client_closed
		}
		if buf1[0] != '$' {
			return nil, errors.New("first char in bulk is not $"), client_not_closed
		}
		bulk_size, _ := strconv.Atoi(string(buf1[1:]))
		buf2, _, e2 := c.reader.ReadLine()
		if e2 != nil {
			return nil, fmt.Errorf("failed read line from client: %v", e2), client_closed
		}
		bulk := string(buf2)
		if len(bulk) != bulk_size {
			return nil, errors.New("wrong bulk size"), client_not_closed
		}
		ret[i] = bulk
	}
	return ret, nil, client_not_closed
}

func (c *Sentinel_connection) getMasterAddrByNameFromSentinel(db_name string) (addr []string, returned_err error, is_client_closed bool) {
	err := c.current_sentinel_connection.SetDeadline(time.Now().Add(c.sentinel_command_timeout))
	if err != nil {
		return nil, err, false
	}
	_, err = fmt.Fprintf(c.writer, "*3\r\n$8\r\nsentinel\r\n$23\r\nget-master-addr-by-name\r\n$%d\r\n%s\r\n", len(db_name), db_name)
	if err != nil {
		return nil, err, false
	}
	err = c.writer.Flush()
	if err != nil {
		return nil, err, false
	}
	err = c.current_sentinel_connection.SetDeadline(time.Time{})
	if err != nil {
		return nil, err, false
	}

	return c.parseResponse()
}

func (c *Sentinel_connection) retrieveAddressByDbName() {
	for db_name := range c.get_master_address_by_name {
		addr, err, is_client_closed := c.getMasterAddrByNameFromSentinel(db_name)
		if err != nil {
			fmt.Println("failed to get master addresses: ", err.Error())
			if !is_client_closed {
				c.get_master_address_by_name_reply <- &Get_master_addr_reply{
					reply: "",
					err:   errors.New("failed to retrieve db name from the sentinel, db_name:" + db_name),
				}
			}
			if !c.reconnectToSentinel() {
				c.get_master_address_by_name_reply <- &Get_master_addr_reply{
					reply: "",
					err:   errors.New("failed to connect to any of the sentinel services"),
				}
			}
			continue
		}
		c.get_master_address_by_name_reply <- &Get_master_addr_reply{
			reply: net.JoinHostPort(addr[0], addr[1]),
			err:   nil,
		}
	}
}

func (c *Sentinel_connection) reconnectToSentinel() bool {
	for _, sentinelAddr := range c.sentinels_addresses {

		if c.current_sentinel_connection != nil {
			c.current_sentinel_connection.Close()
			c.reader = nil
			c.writer = nil
			c.current_sentinel_connection = nil
		}

		u, err := url.Parse("redis://" + sentinelAddr)
		if err != nil {
			fmt.Printf("failed to parse address %s: %v\n", sentinelAddr, err)
			return false
		}

		c.current_sentinel_connection, err = net.DialTimeout("tcp", u.Host, c.sentinel_dial_timeout)
		if err == nil {
			c.reader = bufio.NewReader(c.current_sentinel_connection)
			c.writer = bufio.NewWriter(c.current_sentinel_connection)

			pass, ok := u.User.Password()
			if ok {
				err = c.auth(pass)
				if err != nil {
					fmt.Printf("failed to auth: %v\n", err)
					return false
				}
			}

			return true
		}
		fmt.Printf("failed to dial sentinel %s: %v\n", u.Host, err)
	}
	return false
}

func (c *Sentinel_connection) auth(pass string) error {
	if pass == "" {
		return errors.New("password is not supplied")
	}
	err := c.current_sentinel_connection.SetDeadline(time.Now().Add(c.sentinel_command_timeout))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(c.writer, "*2\r\n$4\r\nauth\r\n$%d\r\n%s\r\n", len(pass), pass)
	if err != nil {
		return err
	}
	err = c.writer.Flush()
	if err != nil {
		return err
	}
	err = c.current_sentinel_connection.SetDeadline(time.Time{})
	if err != nil {
		return err
	}

	err = c.parseAuthResponse()

	return err
}

func (c *Sentinel_connection) parseAuthResponse() error {
	buf, _, err := c.reader.ReadLine()
	if err != nil {
		return fmt.Errorf("failed read line from client: %v", err)
	}

	if !bytes.Equal([]byte("+OK"), buf) {
		return errors.New("failed to authenticate")
	}

	return nil
}

func (c *Sentinel_connection) GetAddressByDbName(name string) (string, error) {
	c.get_master_address_by_name <- name
	reply := <-c.get_master_address_by_name_reply
	return reply.reply, reply.err
}

func NewSentinelConnection(addresses []string, dialTimeout, commandTimeout time.Duration) (*Sentinel_connection, error) {
	connection := Sentinel_connection{
		sentinels_addresses:              addresses,
		get_master_address_by_name:       make(chan string),
		get_master_address_by_name_reply: make(chan *Get_master_addr_reply),
		current_sentinel_connection:      nil,
		reader:                           nil,
		writer:                           nil,
		sentinel_dial_timeout:            time.Millisecond * dialTimeout,
		sentinel_command_timeout:         time.Millisecond * commandTimeout,
	}

	if !connection.reconnectToSentinel() {
		return nil, errors.New("could not connect to any sentinels")
	}

	go connection.retrieveAddressByDbName()

	return &connection, nil
}
