package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/USA-RedDragon/sentinel_tunnel/internal/logger"
	"github.com/USA-RedDragon/sentinel_tunnel/internal/sentinel_connection"
)

// https://goreleaser.com/cookbooks/using-main.version/
//
//nolint:golint,gochecknoglobals
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type SentinelTunnellingDbConfig struct {
	Name       string
	Local_port string
}

type SentinelTunnellingConfiguration struct {
	Sentinels_addresses_list []string
	Databases                []SentinelTunnellingDbConfig
}

type SentinelTunnellingClient struct {
	configuration       SentinelTunnellingConfiguration
	sentinel_connection *sentinel_connection.Sentinel_connection
}

type get_db_address_by_name_function func(db_name string) (string, error)

func NewSentinelTunnellingClient(config_file_location string) *SentinelTunnellingClient {
	data, err := os.ReadFile(config_file_location)
	if err != nil {
		logger.WriteLogMessage(logger.FATAL, "an error has occur during configuration read",
			err.Error())
	}

	Tunnelling_client := SentinelTunnellingClient{}
	err = json.Unmarshal(data, &(Tunnelling_client.configuration))
	if err != nil {
		logger.WriteLogMessage(logger.FATAL, "an error has occur during configuration read,",
			err.Error())
	}

	Tunnelling_client.sentinel_connection, err =
		sentinel_connection.NewSentinelConnection(Tunnelling_client.configuration.Sentinels_addresses_list)
	if err != nil {
		logger.WriteLogMessage(logger.FATAL, "an error has occur, ",
			err.Error())
	}

	logger.WriteLogMessage(logger.INFO, "done initializing Tunnelling")

	return &Tunnelling_client
}

func createTunnelling(conn1 net.Conn, conn2 net.Conn) {
	io.Copy(conn1, conn2)
	conn1.Close()
	conn2.Close()
}

func handleConnection(c net.Conn, db_name string,
	get_db_address_by_name get_db_address_by_name_function) {
	db_address, err := get_db_address_by_name(db_name)
	if err != nil {
		logger.WriteLogMessage(logger.ERROR, "cannot get db address for ", db_name,
			",", err.Error())
		c.Close()
		return
	}
	db_conn, err := net.Dial("tcp", db_address)
	if err != nil {
		logger.WriteLogMessage(logger.ERROR, "cannot connect to db ", db_name,
			",", err.Error())
		c.Close()
		return
	}
	go createTunnelling(c, db_conn)
	go createTunnelling(db_conn, c)
}

func handleSigleDbConnections(listening_port string, db_name string,
	get_db_address_by_name get_db_address_by_name_function) {

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", listening_port))
	if err != nil {
		logger.WriteLogMessage(logger.FATAL, "cannot listen to port ",
			listening_port, err.Error())
	}

	logger.WriteLogMessage(logger.INFO, "listening on port ", listening_port,
		" for connections to database: ", db_name)
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.WriteLogMessage(logger.FATAL, "cannot accept connections on port ",
				listening_port, err.Error())
		}
		go handleConnection(conn, db_name, get_db_address_by_name)
	}

}

func (st_client *SentinelTunnellingClient) Start() {
	for _, db_conf := range st_client.configuration.Databases {
		go handleSigleDbConnections(db_conf.Local_port, db_conf.Name,
			st_client.sentinel_connection.GetAddressByDbName)
	}
}

func main() {
	fmt.Printf("Redis Sentinel Tunnel %s (%s) <%s>\n", version, commit, date)
	if len(os.Args) < 2 {
		fmt.Printf("usage : %s <config_file_path>\n", os.Args[0])
		return
	}
	logger.InitializeLogger()
	st_client := NewSentinelTunnellingClient(os.Args[1])
	st_client.Start()
	for {
		time.Sleep(1000 * time.Millisecond)
	}
}
