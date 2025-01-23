package main

import (
	"fmt"
	"log"
	"tic_tac_toe/internal/tic_tac_toe/handlers"
)

func main() {
	s := handlers.NewServer("127.0.0.1:8080")
	fmt.Printf("starting server on %s\n", s.ListenAddr)
	if err := handlers.ListenAndPair(s); err != nil {
		log.Printf("server failed: %v", err)
	}
}
