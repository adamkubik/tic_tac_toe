package models

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
	Spectators    *[]Spectator
}

type GameResult struct {
	Player1 Player
	Player2 Player
	Winner  *Player
	Loser   *Player
}
