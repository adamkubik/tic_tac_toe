package models

import (
	"database/sql"
	"net"
	"sync"
)

type Server struct {
	ListenAddr  string
	Listener    net.Listener
	ConnsChan   chan Player
	ResultsChan chan GameResult
	DB          *sql.DB

	ActiveGamesMu sync.Mutex
	Games         map[string]*Game

	ActiveUsersMu sync.Mutex
	ActiveUsers   map[string]net.Conn
}
