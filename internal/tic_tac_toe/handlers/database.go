package handlers

import (
	"bufio"
	"database/sql"
	"log"
	"net"
	"strings"
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
			conn.Write([]byte("Invalid password. Disconnecting.\n"))
			return false, nil
		}
		conn.Write([]byte("Welcome back!\n"))
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
		conn.Write([]byte("You have now registered into the game.\n"))
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
