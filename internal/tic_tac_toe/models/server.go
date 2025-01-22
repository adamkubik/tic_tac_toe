package models

import (
	"net"
	"sync"
)

type Server struct {
	ListenAddr string
	Ln         net.Listener

	mu      sync.Mutex
	Players [2]*Player
}
