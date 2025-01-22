package models

import "net"

type Player struct {
	IP       string
	Conn     net.Conn
	NickName string
	Symbol   string
}
