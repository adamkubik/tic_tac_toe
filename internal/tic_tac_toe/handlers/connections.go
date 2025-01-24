package handlers

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"net"
	"strings"
	"tic_tac_toe/internal/tic_tac_toe/models"
)

func NewServer(address string, dB *sql.DB) *models.Server {
	return &models.Server{
		ListenAddr:  address,
		ConnsChan:   make(chan models.Player),
		ResultsChan: make(chan models.GameResult),
		Games:       make(map[string]*models.Game),
		DB:          dB,
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
	go ProcessGameResults(s.ResultsChan)

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

		go handleNewConn(s, conn)
	}
}

func handleNewConn(s *models.Server, conn net.Conn) {
	conn.Write([]byte("Enter 'play' to join a game or 'spectate' to watch: "))

	reader := bufio.NewReader(conn)
	choice, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading choice: %v", err)
		conn.Close()
		return
	}
	choice = strings.TrimSpace(strings.ToLower(choice))

	if choice == "play" {
		handlePlayerConnection(s, conn, reader)
	} else if choice == "spectate" {
		handleSpectatorConnection(s, conn, reader)
	} else {
		conn.Write([]byte("Invalid choice. Disconnecting.\n"))
		conn.Close()
	}
}

func handlePlayerConnection(s *models.Server, conn net.Conn, reader *bufio.Reader) {
	conn.Write([]byte("Enter your username: "))
	username, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("error reading username: %v", err)
		conn.Close()
		return
	}
	username = strings.TrimSpace(username)

	log.Printf("received username %s from %s", username, conn.RemoteAddr())

	succ, err := ProcessNickname(s.DB, conn, reader, username)
	if err != nil {
		conn.Write([]byte("Error processing username. Disconnecting.\n"))
		conn.Close()
		return
	}

	if !succ {
		conn.Close()
		return
	}

	conn.Write([]byte("Waiting for an oponent...\n"))

	player := models.Player{
		IP:       conn.RemoteAddr().String(),
		Conn:     conn,
		NickName: username,
	}

	s.ConnsChan <- player
}

func handleSpectatorConnection(s *models.Server, conn net.Conn, reader *bufio.Reader) {
	if len(s.Games) == 0 {
		conn.Write([]byte("No games are currently active. Disconnecting.\n"))
		conn.Close()
		return
	}

	conn.Write([]byte("Available games:\n"))
	for id, game := range s.Games {
		conn.Write([]byte(fmt.Sprintf("Game ID: %s (Players: %s vs %s)\n", id, game.Player1.NickName, game.Player2.NickName)))
	}

	conn.Write([]byte("Enter the ID of the game you want to spectate: "))
	gameID, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading game ID: %v", err)
		conn.Close()
		return
	}
	gameID = strings.TrimSpace(gameID)

	game, ok := s.Games[gameID]
	if !ok {
		conn.Write([]byte("Invalid game ID. Disconnecting.\n"))
		conn.Close()
		return
	}

	*game.Spectators = append(*game.Spectators, models.Spectator{Conn: conn})
	conn.Write([]byte(fmt.Sprintf("You are now spectating game %s.\n", gameID)))
}

func HandleConns(s *models.Server) {
	for {
		player1 := <-s.ConnsChan
		player2 := <-s.ConnsChan

		player1.Symbol = "X"
		player2.Symbol = "O"

		log.Printf("creating game with %s and %s", player1.NickName, player2.NickName)

		go StartGame(player1, player2, s)
	}
}
