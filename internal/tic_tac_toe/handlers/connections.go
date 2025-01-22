package server

import (
	"fmt"
	"log"
	"net"
	"tic_tac_toe/internal/tic_tac_toe/models"
)

func NewServer(address string) *models.Server {
	return &models.Server{
		ListenAddr: address,
	}
}

func ListenNewConn(s *models.Server) error {
	listener, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return fmt.Errorf("listening server error: %v", err)
	}
	defer listener.Close()

	log.Printf("server is currently running on %s", s.ListenAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("accepting server error: %v", err)
		}

		log.Printf("server now accepted new connection from: %s", conn.RemoteAddr().String())
	}
}
