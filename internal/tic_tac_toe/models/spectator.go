package models

import "net"

type Spectator struct {
	Conn     net.Conn
	NickName string
}
