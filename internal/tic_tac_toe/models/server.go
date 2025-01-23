package models

import (
	"net"
)

type Server struct {
	ListenAddr  string
	Listener    net.Listener
	ConnsChan   chan Player
	ResultsChan chan GameResult
}
