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

func ProcessNickname(db *sql.DB, conn net.Conn, reader *bufio.Reader, nickname string) (bool, error) {
	exists, err := ExistsNickname(db, nickname)
	if err != nil {
		return false, err
	}

	if exists {
		conn.Write([]byte("Enter your password: "))
		password, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("error reading password: %v", err)
			return false, err
		}
		password = strings.TrimSpace(password)
		valid, err := VerifyPassword(db, nickname, password)
		if err != nil {
			return false, err
		}
		if !valid {
			return false, nil
		}
		conn.Write([]byte("\r\nWelcome back!\r\n"))
		return true, nil
	} else {
		conn.Write([]byte("Enter your password to register: "))
		password, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("error reading password: %v", err)
			return false, err
		}
		password = strings.TrimSpace(password)
		err = CreateUser(db, nickname, password)
		if err != nil {
			return false, err
		}
		conn.Write([]byte("You have now registered into the game.\r\n"))
		return true, nil
	}
}

func ExistsNickname(db *sql.DB, nickname string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM players WHERE nickname=$1)"
	err := db.QueryRow(query, nickname).Scan(&exists)
	if err != nil {
		log.Printf("error checking if username exists in database: %v", err)
		return false, err
	}
	return exists, nil
}

func CreateUser(db *sql.DB, nickname, password string) error {
	query := "INSERT INTO players (nickname, password) VALUES ($1, $2)"
	_, err := db.Exec(query, nickname, password)
	if err != nil {
		log.Printf("error creating new user in database: %v", err)
		return err
	}
	return nil
}

func VerifyPassword(db *sql.DB, nickname, password string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM players WHERE nickname=$1 AND password=$2)"
	err := db.QueryRow(query, nickname, password).Scan(&exists)
	if err != nil {
		log.Printf("error verifying password: %v", err)
		return false, err
	}
	return exists, nil
}

func MonitorResults(s *models.Server) {
	for {
		result := <-s.ResultsChan
		if result.Error != nil {
			log.Printf("game %s will not update the database: %v", result.GameID, result.Error)
			continue
		}

		err := UpdatePlayerStats(s.DB, result)
		if err != nil {
			log.Printf("error updating player stats in database: %v", err)
		}
	}
}

func UpdatePlayerStats(db *sql.DB, result models.GameResult) error {
	if result.Winner != nil {
		query := "UPDATE players SET wins = wins + 1, all_games = all_games + 1 WHERE nickname = $1"
		_, err := db.Exec(query, result.Winner.NickName)
		if err != nil {
			log.Printf("error updating winner: %v", err)
			return err
		}
	}

	if result.Loser != nil {
		query := "UPDATE players SET losses = losses + 1, all_games = all_games + 1 WHERE nickname = $1"
		_, err := db.Exec(query, result.Loser.NickName)
		if err != nil {
			log.Printf("error updating loser: %v", err)
			return err
		}
	}

	if result.Winner == nil {
		query := "UPDATE players SET draws = draws + 1, all_games = all_games + 1 WHERE nickname = $1 OR nickname = $2"
		_, err := db.Exec(query, result.Player1.NickName, result.Player2.NickName)
		if err != nil {
			log.Printf("error updating both players due to draw: %v", err)
			return err
		}
	}

	return nil
}

func PrintPlayerStats(dB *sql.DB, nickname string, conn net.Conn) error {
	var numberOfGames, wins, losses, draws int
	query := "SELECT all_games, wins, losses, draws FROM players WHERE nickname=$1"
	err := dB.QueryRow(query, nickname).Scan(&numberOfGames, &wins, &losses, &draws)
	if err != nil {
		log.Printf("error retrieving player stats: %v", err)
		return err
	}

	winRate := float64(wins) / float64(numberOfGames) * 100

	stats := fmt.Sprintf("%s'stats:\r\nAll games: %2d\r\nWins: %7d\r\nLosses: %5d\r\nDraws: %6d\r\nWinrate: %7.1f%%\r\n", nickname, numberOfGames, wins, losses, draws, winRate)
	_, err = conn.Write([]byte(stats))
	if err != nil {
		log.Printf("error writing player stats to connection: %v", err)
		return err
	}

	return nil
}

func PrintTopPlayers(db *sql.DB, conn net.Conn) error {
	query := `
        SELECT nickname, wins, all_games, 
        CASE 
            WHEN all_games = 0 THEN 0
            ELSE CAST(wins AS FLOAT) / CAST(all_games AS FLOAT)
        END AS winrate
        FROM players
        ORDER BY winrate DESC
        LIMIT 10
    `
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("error retrieving top players: %v", err)
		return err
	}
	defer rows.Close()

	var builder strings.Builder
	builder.WriteString("\r\nTop 10 Players:\r\n")
	builder.WriteString(fmt.Sprintf("    %-20s %-10s\r\n", "Nickname", "Winrate"))

	for player := 1; rows.Next(); player++ {
		var nickname string
		var wins, numberOfGames int
		var winRate float64
		err := rows.Scan(&nickname, &wins, &numberOfGames, &winRate)
		if err != nil {
			log.Printf("error scanning top player: %v", err)
			return err
		}
		builder.WriteString(fmt.Sprintf("%2d. %-20s %6.1f%%\r\n", player, nickname, winRate*100))
	}

	if err = rows.Err(); err != nil {
		log.Printf("error iterating over top players: %v", err)
		return err
	}

	_, err = conn.Write([]byte(builder.String()))
	if err != nil {
		log.Printf("error writing top players to connection: %v", err)
		return err
	}

	return nil
}
