package handlers

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"tic_tac_toe/internal/tic_tac_toe/models"
)

func NewServer(address string, dB *sql.DB) *models.Server {
	return &models.Server{
		ListenAddr:  address,
		ConnsChan:   make(chan models.Player),
		ResultsChan: make(chan models.GameResult),
		DB:          dB,

		ActiveGamesMu: sync.Mutex{},
		Games:         make(map[string]*models.Game),

		ActiveUsersMu: sync.Mutex{},
		ActiveUsers:   make(map[string]net.Conn),
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
	if err := trySendMessage(conn, "\r\nEnter: 'login' to authenticate,\r\n       'spectate' to watch or\r\n       'quit' to quit: "); err != nil {
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
		if err := trySendMessage(conn, "Invalid choice. Disconnecting.\r\n"); err != nil {
			log.Printf("error sending message: %v", err)
		}
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

	s.ActiveUsersMu.Lock()

	if _, exists := s.ActiveUsers[nickname]; exists {
		if err := trySendMessage(conn, "User already logged in. Disconnecting.\r\n"); err != nil {
			log.Printf("error sending message: %v", err)
		}
		conn.Close()
		s.ActiveUsersMu.Unlock()
		return
	}

	for attempt := 0; attempt < 3; attempt++ {
		succ, err := ProcessNickname(s.DB, conn, reader, nickname)
		if err != nil {
			if err := trySendMessage(conn, "Error processing nickname. Disconnecting.\r\n"); err != nil {
				log.Printf("error sending message: %v", err)
			}
			conn.Close()
			s.ActiveUsersMu.Unlock()
			handleLogout(s, nickname)
			return
		}

		if succ {
			break
		}

		if attempt == 2 {
			if err := trySendMessage(conn, "Invalid password. Disconnecting.\r\n"); err != nil {
				log.Printf("error sending message: %v", err)
			}
			conn.Close()
			s.ActiveUsersMu.Unlock()
			handleLogout(s, nickname)
			return
		}
		if err := trySendMessage(conn, fmt.Sprintf("Invalid password. Try again. %d attempt(s) left.\r\n", 2-attempt)); err != nil {
			log.Printf("error sending message: %v", err)
		}
	}

	s.ActiveUsers[nickname] = conn
	s.ActiveUsersMu.Unlock()

	if err := handleBasicCommands(s, conn, reader, nickname); err != nil {
		conn.Close()
		handleLogout(s, nickname)
	}
}

func handleBasicCommands(s *models.Server, conn net.Conn, reader *bufio.Reader, nickname string) error {
	for {
		if err := trySendMessage(conn, "\r\nEnter: 'play' to join a game,\r\n       'stats' to view your statistics,\r\n       'top10' to view top 10 players or\r\n       'quit' to quit: "); err != nil {
			return err
		}

		choice, err := tryReadMessage(conn, reader)
		if err != nil {
			return err
		}
		if err := trySendMessage(conn, "\r\n"); err != nil {
			return err
		}

		choice = strings.TrimSpace(strings.ToLower(choice))

		switch choice {
		case "play":
			handlePlayerConnection(s, conn, nickname)
			return nil
		case "stats":
			handleStatsRequest(s, conn, nickname)
		case "top10":
			if err := PrintTopPlayers(s.DB, conn); err != nil {
				if err := trySendMessage(conn, "Error printing top10 players.\r\n"); err != nil {
					return err
				}
				return err
			}
		case "quit":
			conn.Close()
			handleLogout(s, nickname)
			return nil
		default:
			if err := trySendMessage(conn, "Invalid choice. Please enter 'play', 'stats', 'top10' or 'quit': \r\n"); err != nil {
				return err
			}
		}
	}
}

func handleStatsRequest(s *models.Server, conn net.Conn, username string) {
	err := PrintPlayerStats(s.DB, username, conn)
	if err != nil {
		if err := trySendMessage(conn, "Error retrieving statistics. Disconnecting.\r\n"); err != nil {
			log.Printf("error sending message: %v", err)
		}
		conn.Close()
		handleLogout(s, username)
		return
	}
}

func handlePlayerConnection(s *models.Server, conn net.Conn, nickname string) {
	if err := trySendMessage(conn, "Waiting for an oponent...\r\n"); err != nil {
		handleLogout(s, nickname)
		return
	}

	player := models.Player{
		IP:       conn.RemoteAddr().String(),
		Conn:     conn,
		NickName: nickname,
	}

	s.ConnsChan <- player
}

func handleLogout(s *models.Server, nickname string) {
	s.ActiveUsersMu.Lock()
	defer s.ActiveUsersMu.Unlock()

	delete(s.ActiveUsers, nickname)
}

func handleSpectatorConnection(s *models.Server, conn net.Conn, reader *bufio.Reader) {
	s.ActiveGamesMu.Lock()
	if len(s.Games) == 0 {
		if err := trySendMessage(conn, "No games are currently active. Disconnecting.\r\n"); err != nil {
			log.Printf("error sending message: %v", err)
		}
		conn.Close()
		s.ActiveGamesMu.Unlock()
		return
	}
	s.ActiveGamesMu.Unlock()

	if err := trySendMessage(conn, "Available games:\r\n"); err != nil {
		return
	}

	s.ActiveGamesMu.Lock()
	for id, game := range s.Games {
		if err := trySendMessage(conn, fmt.Sprintf("Game ID: %s (Players: %s vs %s)\r\n", id, game.Player1.NickName, game.Player2.NickName)); err != nil {
			s.ActiveGamesMu.Unlock()
			return
		}
	}
	s.ActiveGamesMu.Unlock()

	if err := trySendMessage(conn, "Enter the ID of the game you want to spectate: "); err != nil {
		return
	}

	gameID, err := tryReadMessage(conn, reader)
	if err != nil {
		return
	}
	gameID = strings.TrimSpace(gameID)

	s.ActiveGamesMu.Lock()
	game, ok := s.Games[gameID]
	if !ok {
		if err := trySendMessage(conn, "Invalid game ID or the game has finished in the meantime. Disconnecting.\n"); err != nil {
			log.Printf("error sending message: %v", err)
		}
		conn.Close()
		s.ActiveGamesMu.Unlock()
		return
	}

	spectator := models.Spectator{Conn: conn}
	(*game.Spectators)[spectator] = struct{}{}

	if err := trySendMessage(conn, fmt.Sprintf("You are now spectating game %s.\r\n", gameID)); err != nil {
		return
	}

	s.ActiveGamesMu.Unlock()

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
