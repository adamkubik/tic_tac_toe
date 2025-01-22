package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"tic_tac_toe/internal/tic_tac_toe/models"
	"time"
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
	s.Ln = listener
	defer listener.Close()

	log.Printf("server is currently running on %s", s.ListenAddr)

	go AcceptConnections(s)
	waitForPlayers(s)
	return nil
}

func AcceptConnections(s *models.Server) {
	for s.Players[0] == nil || s.Players[1] == nil {
		conn, err := s.Ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		log.Printf("server now accepted new connection from: %s", conn.RemoteAddr().String())

		player := models.Player{
			Conn: conn,
		}
		s.Mu.Lock()
		if s.Players[0] == nil {
			player.Symbol = "X"
			s.Players[0] = &player
		} else {
			player.Symbol = "O"
			s.Players[1] = &player
		}
		s.Mu.Unlock()
	}
}

func waitForPlayers(s *models.Server) {
	for {
		s.Mu.Lock()
		if s.Players[0] != nil && s.Players[1] != nil {
			s.Mu.Unlock()
			break
		}
		s.Mu.Unlock()
		time.Sleep(1 * time.Second)
	}

	game := models.Game{
		Player1:       *s.Players[0],
		Player2:       *s.Players[1],
		OnGoing:       true,
		CurrentPlayer: *s.Players[0],
	}

	startGame(&game)
}

func startGame(g *models.Game) {
	sendMessage(g.Player1.Conn, "Both players are connected. The game will start now!")
	sendMessage(g.Player2.Conn, "Both players are connected. The game will start now!")

	fmt.Println("Starting the game...")
}

func sendMessage(conn net.Conn, message string) {
	if conn == nil {
		return
	}
	writer := bufio.NewWriter(conn)
	_, err := writer.WriteString(message + "\n")
	if err == nil {
		writer.Flush()
	}
}
