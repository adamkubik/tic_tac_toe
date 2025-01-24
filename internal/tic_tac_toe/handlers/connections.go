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
	_, err := conn.Write([]byte("Enter 'login' to authenticate or 'spectator' to watch: "))
	if err != nil {
		LogAndClose(fmt.Sprintf("writing to connection error: %v", err), conn)
		return
	}

	reader := bufio.NewReader(conn)
	choice, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading choice: %v", err)
		conn.Close()
		return
	}
	choice = strings.TrimSpace(strings.ToLower(choice))

	if choice == "login" {
		handleLogin(s, conn, reader)
	} else if choice == "spectator" {
		handleSpectatorConnection(s, conn, reader)
	} else {
		_, err = conn.Write([]byte("Invalid choice. Disconnecting.\n"))
		if err != nil {
			LogAndClose(fmt.Sprintf("writing to connection error: %v", err), conn)
		}
	}
}

func handleLogin(s *models.Server, conn net.Conn, reader *bufio.Reader) {
	_, err := conn.Write([]byte("Enter your nickname: "))
	if err != nil {
		LogAndClose(fmt.Sprintf("writing to connection error: %v", err), conn)
		return
	}

	nickname, err := reader.ReadString('\n')
	if err != nil {
		LogAndClose(fmt.Sprintf("error reading nickname: %v", err), conn)
		return
	}
	nickname = strings.TrimSpace(nickname)

	log.Printf("received nickname %s from %s", nickname, conn.RemoteAddr())

	succ, err := ProcessNickname(s.DB, conn, reader, nickname)
	if err != nil {
		conn.Write([]byte("Error processing nickname. Disconnecting.\n"))
		conn.Close()
		return
	}

	if !succ {
		conn.Close()
		return
	}

	for {
		_, err = conn.Write([]byte("Enter 'play' to join a game or 'stats' to view your statistics: "))
		if err != nil {
			LogAndClose(fmt.Sprintf("writing to connection error: %v", err), conn)
			return
		}

		choice, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading choice: %v", err)
			conn.Close()
			return
		}
		choice = strings.TrimSpace(strings.ToLower(choice))

		if choice == "play" {
			handlePlayerConnection(s, conn, nickname)
			break
		} else if choice == "stats" {
			handleStatsRequest(s, conn, reader, nickname)
		} else {
			_, err = conn.Write([]byte("Invalid choice. Please enter 'play' or 'stats'.\n"))
			if err != nil {
				LogAndClose(fmt.Sprintf("writing to connection error: %v", err), conn)
				return
			}
		}
	}
}

func handleStatsRequest(s *models.Server, conn net.Conn, reader *bufio.Reader, username string) {
	err := PrintPlayerStats(s.DB, username, conn, reader)
	if err != nil {
		conn.Write([]byte("Error retrieving statistics. Disconnecting.\n"))
		conn.Close()
		return
	}
}

func handlePlayerConnection(s *models.Server, conn net.Conn, nickname string) {
	_, err := conn.Write([]byte("Waiting for an oponent...\n"))
	if err != nil {
		LogAndClose(fmt.Sprintf("writing to connection error: %v", err), conn)
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
		_, err := conn.Write([]byte("No games are currently active. Disconnecting.\n"))
		if err != nil {
			LogAndClose(fmt.Sprintf("writing to connection error: %v", err), conn)
		}
		return
	}

	_, err := conn.Write([]byte("Available games:\n"))
	if err != nil {
		LogAndClose(fmt.Sprintf("writing to connection error: %v", err), conn)
		return
	}

	for id, game := range s.Games {
		_, err = conn.Write([]byte(fmt.Sprintf("Game ID: %s (Players: %s vs %s)\n", id, game.Player1.NickName, game.Player2.NickName)))
		if err != nil {
			LogAndClose(fmt.Sprintf("writing to connection error: %v", err), conn)
			return
		}
	}

	_, err = conn.Write([]byte("Enter the ID of the game you want to spectate: "))
	if err != nil {
		LogAndClose(fmt.Sprintf("writing to connection error: %v", err), conn)
		return
	}

	gameID, err := reader.ReadString('\n')
	if err != nil {
		LogAndClose(fmt.Sprintf("Error reading game ID: %v", err), conn)
		return
	}
	gameID = strings.TrimSpace(gameID)

	game, ok := s.Games[gameID]
	if !ok {
		_, err = conn.Write([]byte("Invalid game ID. Disconnecting.\n"))
		if err != nil {
			LogAndClose(fmt.Sprintf("writing to connection error: %v", err), conn)
		}
		return
	}

	*game.Spectators = append(*game.Spectators, models.Spectator{Conn: conn})
	_, err = conn.Write([]byte(fmt.Sprintf("You are now spectating game %s.\n", gameID)))
	if err != nil {
		LogAndClose(fmt.Sprintf("writing to connection error: %v", err), conn)
		return
	}
}

func LogAndClose(errMsg string, conn net.Conn) {
	log.Print(errMsg)
	conn.Close()
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
