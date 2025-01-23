package handlers

import (
	"fmt"
	"strings"
	"tic_tac_toe/internal/tic_tac_toe/models"
)

func StartGame(p1 models.Player, p2 models.Player, s *models.Server) {
	board := [3][3]string{
		{" ", " ", " "},
		{" ", " ", " "},
		{" ", " ", " "},
	}

	g := models.Game{
		Player1:       p1,
		Player2:       p2,
		OnGoing:       true,
		CurrentPlayer: &p1,
		WaitingPlayer: &p2,
		Board:         &board,
		Winner:        nil,
		Loser:         nil,
	}

	sendMessage(g.CurrentPlayer, "The game is starting... you're player 'X'\n")
	sendMessage(g.WaitingPlayer, "The game is starting... you're player 'O'\n")
	playGame(&g, s)
}

func playGame(g *models.Game, s *models.Server) {
	for g.OnGoing {
		sendBoard(g.CurrentPlayer, g.Board)
		sendMessage(g.WaitingPlayer, "Waiting for your oponent's turn...\n")

		var row, col int
		for {
			move := requestMove(g.CurrentPlayer)

			var err error
			row, col, err = validateMove(move, g.Board)
			if err != nil {
				sendMessage(g.CurrentPlayer, fmt.Sprintf("Invalid move: %s. Try again.\n", err.Error()))
				continue
			}

			break
		}

		g.Board[row][col] = g.CurrentPlayer.Symbol

		if checkWin(g.Board, g.CurrentPlayer.Symbol) {
			g.Winner = g.CurrentPlayer
			g.Loser = g.WaitingPlayer
			g.OnGoing = false
			sendBoard(g.CurrentPlayer, g.Board)
			sendBoard(g.WaitingPlayer, g.Board)
			break
		} else if isDraw(g.Board) {
			g.OnGoing = false
			break
		}

		sendBoard(g.CurrentPlayer, g.Board)
		g.CurrentPlayer, g.WaitingPlayer = g.WaitingPlayer, g.CurrentPlayer
	}

	announceResult(g, s)
}

func sendBoard(player *models.Player, board *[3][3]string) {
	var boardStr strings.Builder

	boardStr.WriteString("\n")
	boardStr.WriteString("   1   2   3\n")
	rows := []string{"A ", "B ", "C "}

	for i, row := range board {
		boardStr.WriteString(rows[i])
		for j, cell := range row {
			if cell == "" {
				boardStr.WriteString("   ")
			} else {
				boardStr.WriteString(" " + cell + " ")
			}
			if j < 2 {
				boardStr.WriteString("|")
			}
		}
		boardStr.WriteString("\n")

		if i < 2 {
			boardStr.WriteString("  -----------\n")
		}
	}
	boardStr.WriteString("\n")

	sendMessage(player, boardStr.String())
}

func requestMove(player *models.Player) string {
	sendMessage(player, "Your move (format: A1, B3, etc.): ")
	buffer := make([]byte, 1024)
	n, _ := player.Conn.Read(buffer)
	return strings.TrimSpace(string(buffer[:n]))
}

func validateMove(move string, board *[3][3]string) (int, int, error) {
	if len(move) != 2 {
		return -1, -1, fmt.Errorf("move must be 2 characters (e.g., A1)")
	}
	row := int(move[0] - 'A')
	col := int(move[1] - '1')
	if row < 0 || row >= 3 {
		return -1, -1, fmt.Errorf("row out of bounds")
	}
	if col < 0 || col >= 3 {
		return -1, -1, fmt.Errorf("column out of bounds")
	}
	if board[row][col] != " " {
		return -1, -1, fmt.Errorf("cell already occupied")
	}
	return row, col, nil
}

func checkWin(board *[3][3]string, symbol string) bool {
	for i := 0; i < 3; i++ {
		if board[i][0] == symbol && board[i][1] == symbol && board[i][2] == symbol {
			return true
		}
		if board[0][i] == symbol && board[1][i] == symbol && board[2][i] == symbol {
			return true
		}
	}
	if board[0][0] == symbol && board[1][1] == symbol && board[2][2] == symbol {
		return true
	}
	if board[0][2] == symbol && board[1][1] == symbol && board[2][0] == symbol {
		return true
	}
	return false
}

func isDraw(board *[3][3]string) bool {
	for _, row := range board {
		for _, cell := range row {
			if cell == " " {
				return false
			}
		}
	}
	return true
}

func announceResult(g *models.Game, s *models.Server) {
	result := models.GameResult{
		Player1: g.Player1,
		Player2: g.Player2,
		Winner:  nil,
		Loser:   nil,
	}

	if g.Winner != nil {
		sendMessage(g.Winner, "Game Over. You won!\n")
		sendMessage(g.Loser, "Game Over. You lost!\n")
		result.Winner = g.Winner
		result.Loser = g.Loser
	} else {
		sendMessage(&g.Player1, "Game Over. It's a draw!\n")
		sendMessage(&g.Player2, "Game Over. It's a draw!\n")
	}

	g.Player1.Conn.Close()
	g.Player2.Conn.Close()
	s.ResultsChan <- result
}

func sendMessage(player *models.Player, message string) {
	player.Conn.Write([]byte(message))
}

func ProcessGameResults(resultsChan <-chan models.GameResult) {
	for result := range resultsChan {
		fmt.Printf("Game finished!\n")
		fmt.Printf("Player 1: %s (%s)\n", result.Player1.NickName, result.Player1.Symbol)
		fmt.Printf("Player 2: %s (%s)\n", result.Player2.NickName, result.Player2.Symbol)

		if result.Winner != nil {
			fmt.Printf("Winner: %s\n", result.Winner.NickName)
		} else {
			fmt.Printf("It's a draw!\n")
		}

		fmt.Println("----------------------")
	}
}
