package server

import (
	"log"
	"net"
	"tic_tac_toe/internal/tic_tac_toe/models"
)

func NewServer(address string) *models.Server {
	return &models.Server{
		ListenAddr: address,
		ConnsChan:  make(chan net.Conn),
	}
}

func ListenAndPair(s *models.Server) error {
	listener, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	defer listener.Close()

	s.Listener = listener

	log.Printf("server is listening on %s", s.ListenAddr)

	go AcceptNewConns(s)

	HandleConns(s)

	return nil
}

func AcceptNewConns(s *models.Server) {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			log.Printf("accepting connection error: %v", err)
			continue
		}

		log.Printf("new success connection from %s", conn.RemoteAddr())

		s.ConnsChan <- conn
	}
}

func HandleConns(s *models.Server) {
	for {
		conn1 := <-s.ConnsChan
		conn2 := <-s.ConnsChan

		log.Printf("creating game with %s and %s", conn1.RemoteAddr(), conn2.RemoteAddr())

		go StartGame(conn1, conn2)
	}
}

func StartGame(first net.Conn, second net.Conn) {
	log.Printf("Game started for players: %s and %s", first.RemoteAddr().String(), second.RemoteAddr().String())
}
