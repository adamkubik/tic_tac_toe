package main

import (
	"fmt"
	"log"
	server "tic_tac_toe/internal/tic_tac_toe/handlers"
)

func main() {
	s := server.NewServer("127.0.0.1:8080")
	fmt.Printf("starting server on %s\n", s.ListenAddr)
	if err := server.ListenAndPair(s); err != nil {
		log.Printf("server failed: %v", err)
	}
}
