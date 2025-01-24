package main

import (
	"fmt"
	"log"
	config "tic_tac_toe/db"
	"tic_tac_toe/internal/tic_tac_toe/handlers"
)

func main() {
	cfg := config.LoadConfig()
	dB := config.InitDB(cfg)
	defer dB.Close()

	s := handlers.NewServer("127.0.0.1:8080", dB)
	fmt.Printf("starting server on %s\n", s.ListenAddr)
	if err := handlers.ListenAndPair(s); err != nil {
		log.Printf("server failed: %v", err)
	}
}
