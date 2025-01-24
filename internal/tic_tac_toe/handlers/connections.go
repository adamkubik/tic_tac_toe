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
		ActiveUsers: make(map[string]net.Conn),
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
	go MonitorResults(s)

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
	if err := trySendMessage(conn, "\nEnter: 'login' to authenticate,\n       'spectate' to watch or\n       'quit' to quit: "); err != nil {
		return
	}

	reader := bufio.NewReader(conn)
	choice, err := tryReadMessage(conn, reader)
	if err != nil {
		return
	}
	choice = strings.TrimSpace(strings.ToLower(choice))

	if choice == "login" {
		handleLogin(s, conn, reader)
	} else if choice == "spectate" {
		handleSpectatorConnection(s, conn, reader)
	} else if choice == "quit" {
		conn.Close()
	} else {
		trySendMessage(conn, "Invalid choice. Disconnecting.\n")
		conn.Close()
	}
}

func handleLogin(s *models.Server, conn net.Conn, reader *bufio.Reader) {
	if err := trySendMessage(conn, "Enter your nickname: "); err != nil {
		return
	}

	nickname, err := tryReadMessage(conn, reader)
	if err != nil {
		return
	}
	nickname = strings.TrimSpace(nickname)

	log.Printf("received nickname %s from %s", nickname, conn.RemoteAddr())

	if _, exists := s.ActiveUsers[nickname]; exists {
		trySendMessage(conn, "User already logged in. Disconnecting.\n")
		conn.Close()
		return
	}

	succ, err := ProcessNickname(s.DB, conn, reader, nickname)
	if err != nil {
		trySendMessage(conn, "Error processing nickname. Disconnecting.\n")
		conn.Close()
		return
	}

	if !succ {
		conn.Close()
		return
	}

	s.ActiveUsers[nickname] = conn

	for {
		if err := trySendMessage(conn, "\nEnter: 'play' to join a game,\n       'stats' to view your statistics,\n       'top10' to view top 10 players or\n       'quit' to quit: "); err != nil {
			delete(s.ActiveUsers, nickname)
			return
		}

		choice, err := tryReadMessage(conn, reader)
		if err != nil {
			delete(s.ActiveUsers, nickname)
			return
		}
		choice = strings.TrimSpace(strings.ToLower(choice))

		if choice == "play" {
			handlePlayerConnection(s, conn, nickname)
			break
		} else if choice == "stats" {
			handleStatsRequest(s, conn, nickname)
		} else if choice == "top10" {
			err := PrintTopPlayers(s.DB, conn)
			if err != nil {
				conn.Write([]byte("kasfnskjnvks"))
				conn.Close()
				delete(s.ActiveUsers, nickname)
				return
			}
		} else if choice == "quit" {
			conn.Close()
			delete(s.ActiveUsers, nickname)
			break
		} else {
			if err := trySendMessage(conn, "Invalid choice. Please enter 'play', 'stats' 'top10' or 'quit': \n"); err != nil {
				conn.Close()
				delete(s.ActiveUsers, nickname)
				return
			}
		}
	}
}

func handleStatsRequest(s *models.Server, conn net.Conn, username string) {
	err := PrintPlayerStats(s.DB, username, conn)
	if err != nil {
		trySendMessage(conn, "Error retrieving statistics. Disconnecting.\n")
		conn.Close()
		delete(s.ActiveUsers, username)
		return
	}
}

func handlePlayerConnection(s *models.Server, conn net.Conn, nickname string) {
	if err := trySendMessage(conn, "Waiting for an oponent...\n"); err != nil {
		delete(s.ActiveUsers, nickname)
		return
	}

	player := models.Player{
		IP:       conn.RemoteAddr().String(),
		Conn:     conn,
		NickName: nickname,
	}

	s.ConnsChan <- player
}

func handleSpectatorConnection(s *models.Server, conn net.Conn, reader *bufio.Reader) {
	if len(s.Games) == 0 {
		trySendMessage(conn, "No games are currently active. Disconnecting.\n")
		conn.Close()
		return
	}

	if err := trySendMessage(conn, "Available games:\n"); err != nil {
		return
	}

	for id, game := range s.Games {
		if err := trySendMessage(conn, fmt.Sprintf("Game ID: %s (Players: %s vs %s)\n", id, game.Player1.NickName, game.Player2.NickName)); err != nil {
			return
		}
	}

	if err := trySendMessage(conn, "Enter the ID of the game you want to spectate: "); err != nil {
		return
	}

	gameID, err := tryReadMessage(conn, reader)
	if err != nil {
		return
	}
	gameID = strings.TrimSpace(gameID)

	game, ok := s.Games[gameID]
	if !ok {
		trySendMessage(conn, "Invalid game ID. Disconnecting.\n")
		conn.Close()
		return
	}

	if err := trySendMessage(conn, "You are now spectating game %s.\n"); err != nil {
		return
	}
	spectator := models.Spectator{Conn: conn}
	(*game.Spectators)[spectator] = struct{}{}
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

func trySendMessage(conn net.Conn, message string) error {
	_, err := conn.Write([]byte(message))
	if err != nil {
		log.Print(fmt.Errorf("failed to send message to %s: %w", conn.RemoteAddr(), err))
		conn.Close()
		return err
	}

	return nil
}

func tryReadMessage(conn net.Conn, reader *bufio.Reader) (string, error) {
	msg, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading message from user: %v", err)
		conn.Close()
		return "", err
	}
	return msg, nil
}
