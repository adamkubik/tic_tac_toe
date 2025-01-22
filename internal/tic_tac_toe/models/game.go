package models

type Game struct {
	Player1       Player
	Player2       Player
	OnGoing       bool
	CurrentPlayer Player
	Board         [3][3]string
	Winner        Player
}
