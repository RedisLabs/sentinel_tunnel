package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/RedisLabs/sentinel_tunnel/st_logger"
	"github.com/RedisLabs/sentinel_tunnel/st_sentinel_connection"
)

type SentinelTunnellingDbConfig struct {
	Name       string
	Local_port string
}

type SentinelTunnellingConfiguration struct {
	Sentinels_addresses_list []string
	Sentinel_dial_timeout    time.Duration
	Sentinel_command_timeout time.Duration
	Sentinel_proxy_timeout   time.Duration
	Databases                []SentinelTunnellingDbConfig
}

type SentinelTunnellingClient struct {
	configuration       SentinelTunnellingConfiguration
	sentinel_connection *st_sentinel_connection.Sentinel_connection
}

type get_db_address_by_name_function func(db_name string) (string, error)

func NewSentinelTunnellingClient(config_file_location string) *SentinelTunnellingClient {
	data, err := ioutil.ReadFile(config_file_location)
	if err != nil {
		st_logger.WriteLogMessage(st_logger.FATAL, "an error has occur during configuration read",
			err.Error())
	}

	Tunnelling_client := SentinelTunnellingClient{}
	err = json.Unmarshal(data, &(Tunnelling_client.configuration))
	if err != nil {
		st_logger.WriteLogMessage(st_logger.FATAL, "an error has occur during configuration read,",
			err.Error())
	}

	Tunnelling_client.sentinel_connection, err =
		st_sentinel_connection.NewSentinelConnection(Tunnelling_client.configuration.Sentinels_addresses_list, Tunnelling_client.configuration.Sentinel_dial_timeout, Tunnelling_client.configuration.Sentinel_command_timeout)
	if err != nil {
		st_logger.WriteLogMessage(st_logger.FATAL, "an error has occur, ",
			err.Error())
	}

	st_logger.WriteLogMessage(st_logger.INFO, "done initializing Tunnelling")

	return &Tunnelling_client
}

func createTunnelling(conn1 net.Conn, conn2 net.Conn) {
	_, err := io.Copy(conn1, conn2)
	if err != nil {
		st_logger.WriteLogMessage(st_logger.ERROR, "tunneling failed, ", err.Error())
	}
	conn1.Close()
	conn2.Close()
}

func handleConnection(c net.Conn, db_name string,
	get_db_address_by_name get_db_address_by_name_function, proxy_timeout time.Duration) {
	db_address, err := get_db_address_by_name(db_name)
	if err != nil {
		st_logger.WriteLogMessage(st_logger.ERROR, "cannot get db address for ", db_name,
			",", err.Error())
		c.Close()
		return
	}
	db_conn, err := net.DialTimeout("tcp", db_address, proxy_timeout)
	if err != nil {
		st_logger.WriteLogMessage(st_logger.ERROR, "cannot connect to db ", db_name,
			",", err.Error())
		c.Close()
		return
	}
	go createTunnelling(c, db_conn)
	go createTunnelling(db_conn, c)
}

func handleSigleDbConnections(listening_port string, db_name string,
	get_db_address_by_name get_db_address_by_name_function, proxy_timeout time.Duration) {

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", listening_port))
	if err != nil {
		st_logger.WriteLogMessage(st_logger.FATAL, "cannot listen to port ",
			listening_port, err.Error())
	}

	st_logger.WriteLogMessage(st_logger.INFO, "listening on port ", listening_port,
		" for connections to database: ", db_name)
	for {
		conn, err := listener.Accept()
		if err != nil {
			st_logger.WriteLogMessage(st_logger.FATAL, "cannot accept connections on port ",
				listening_port, err.Error())
		}
		go handleConnection(conn, db_name, get_db_address_by_name, proxy_timeout)
	}

}

func (st_client *SentinelTunnellingClient) Start() {
	for _, db_conf := range st_client.configuration.Databases {
		go handleSigleDbConnections(db_conf.Local_port, db_conf.Name,
			st_client.sentinel_connection.GetAddressByDbName, st_client.configuration.Sentinel_proxy_timeout*time.Millisecond)
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage : sentinel_tunnel <config_file_path> <log_file_path>")
		return
	}
	st_logger.InitializeLogger(os.Args[2])
	st_client := NewSentinelTunnellingClient(os.Args[1])
	st_client.Start()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
}
