package models

import "sync"

type Game struct {
	ID            string
	Player1       Player
	Player2       Player
	Board         *[3][3]string
	OnGoing       bool
	CurrentPlayer *Player
	WaitingPlayer *Player
	Winner        *Player
	Loser         *Player

	SpectatorsMu sync.Mutex
	Spectators   *map[Spectator]struct{}

	Error error
}

type GameResult struct {
	GameID  string
	Player1 Player
	Player2 Player
	Winner  *Player
	Loser   *Player
	Error   error
}
