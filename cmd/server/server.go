package server

import (
	"fmt"
	"github.com/DivPro/sentinel_tunnel/cmd/config"
	"github.com/DivPro/sentinel_tunnel/cmd/resolver"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"io"
	"net"
)

type Server interface {
	Start() error
	Stop()
}

type server struct {
	resolver  resolver.Resolver
	databases []*config.DatabaseConfig
	listeners []net.Listener
}

func NewServer(resolver resolver.Resolver, databases []*config.DatabaseConfig) Server {
	return &server{
		resolver:  resolver,
		databases: databases,
	}
}

func (s *server) Start() error {
	var eg errgroup.Group
	fn := func(db *config.DatabaseConfig) func() error {
		return func() error {
			return s.startDatabase(db)
		}
	}
	for _, db := range s.databases {
		eg.Go(fn(db))
	}

	return eg.Wait()
}

func (s *server) Stop() {
	for _, listener := range s.listeners {
		_ = listener.Close()
	}
}

func (s *server) startDatabase(conf *config.DatabaseConfig) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", conf.Port))
	if err != nil {
		log.Error().Err(err).Msgf("error listening port: %d", conf.Port)
		return err
	}

	log.Info().Msgf("listening started '%s' on port %d", conf.Name, conf.Port)
	s.listeners = append(s.listeners, listener)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error().Err(err).Msgf("error accepting connection to: %s on port %d", conf.Name, conf.Port)
			continue
		}

		go s.handleConnection(conn, conf.Name)
	}
}

func (s *server) handleConnection(conn net.Conn, dbName string) {
	log.Debug().Msgf("resolving: %s", dbName)
	dbAddr, err := s.resolver.Resolve(dbName)
	if err != nil {
		log.Error().Err(err).Msgf("failed resolving db: %s", dbName)
		_ = conn.Close()
		return
	}

	dbConn, err := net.Dial("tcp", dbAddr)
	if err != nil {
		log.Error().Err(err).Msgf("error tunneling to: %s", dbName)
		_ = conn.Close()
		return
	}
	go createTunnelling(conn, dbConn)
	go createTunnelling(dbConn, conn)
}

func createTunnelling(conn1 net.Conn, conn2 net.Conn) {
	_, err := io.Copy(conn1, conn2)
	if err != nil {
		log.Error().Err(err).Msg("tunneling")
	}
	_ = conn1.Close()
	_ = conn2.Close()
}
