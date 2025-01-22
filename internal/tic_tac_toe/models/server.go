package models

import (
	"net"
)

type Server struct {
	ListenAddr string
	Listener   net.Listener
	ConnsChan  chan net.Conn
}
