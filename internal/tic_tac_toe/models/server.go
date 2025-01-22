package models

import "net"

type Server struct {
	ListenAddr string
	Ln         net.Listener
	Players    [2]*Player
}
