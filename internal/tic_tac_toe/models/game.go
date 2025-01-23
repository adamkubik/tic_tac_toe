package models

type Game struct {
	Player1       Player
	Player2       Player
	OnGoing       bool
	CurrentPlayer *Player
	WaitingPlayer *Player
	Board         *[3][3]string
	Winner        *Player
	Loser         *Player
}

type GameResult struct {
	Player1 Player
	Player2 Player
	Winner  *Player
	Loser   *Player
}
