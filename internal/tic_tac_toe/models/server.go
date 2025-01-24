package models

import (
	"database/sql"
	"net"
)

type Server struct {
	ListenAddr  string
	Listener    net.Listener
	ConnsChan   chan Player
	ResultsChan chan GameResult
	Games       map[string]*Game
	DB          *sql.DB
	ActiveUsers map[string]net.Conn
}
