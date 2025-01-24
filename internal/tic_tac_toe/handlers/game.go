package handlers

import (
	"fmt"
	"strings"
	"tic_tac_toe/internal/tic_tac_toe/models"

	"github.com/google/uuid"
)

func StartGame(p1 models.Player, p2 models.Player, s *models.Server) {
	board := [3][3]string{
		{" ", " ", " "},
		{" ", " ", " "},
		{" ", " ", " "},
	}

	gameId := uuid.New().String()

	g := models.Game{
		ID:            gameId,
		Player1:       p1,
		Player2:       p2,
		OnGoing:       true,
		CurrentPlayer: &p1,
		WaitingPlayer: &p2,
		Board:         &board,
		Winner:        nil,
		Loser:         nil,
		Spectators:    &map[models.Spectator]struct{}{},
	}

	s.Games[gameId] = &g

	sendMessage(g.CurrentPlayer, "The game is starting... you're player 'X'\n")
	sendMessage(g.WaitingPlayer, "The game is starting... you're player 'O'\n")
	playGame(&g, s)
}

func playGame(g *models.Game, s *models.Server) {
	for g.OnGoing {
		board := getBoard(g.Board)
		if err := sendMessage(g.CurrentPlayer, board); err != nil {
			handleError(g, s, err)
			return
		}
		if err := sendMessage(g.WaitingPlayer, "Waiting for your oponent's turn...\n"); err != nil {
			handleError(g, s, err)
			return
		}
		sendToSpectators(g, board)

		var row, col int
		for {
			move, err := requestMove(g.CurrentPlayer)
			if err != nil {
				handleError(g, s, err)
				return
			}

			row, col, err = validateMove(move, g.Board)
			if err != nil {
				if err := sendMessage(g.CurrentPlayer, fmt.Sprintf("Invalid move: %s. Try again.\n", err.Error())); err != nil {
					handleError(g, s, err)
					return
				}
				continue
			}

			break
		}

		g.Board[row][col] = g.CurrentPlayer.Symbol

		board = getBoard(g.Board)
		if checkWin(g.Board, g.CurrentPlayer.Symbol) {
			g.Winner = g.CurrentPlayer
			g.Loser = g.WaitingPlayer
			g.OnGoing = false
			if err := sendMessage(g.CurrentPlayer, board); err != nil {
				handleError(g, s, err)
				return
			}
			if err := sendMessage(g.WaitingPlayer, board); err != nil {
				handleError(g, s, err)
				return
			}
			sendToSpectators(g, board)
			break
		} else if isDraw(g.Board) {
			g.OnGoing = false
			break
		}

		if err := sendMessage(g.CurrentPlayer, board); err != nil {
			handleError(g, s, err)
			return
		}
		g.CurrentPlayer, g.WaitingPlayer = g.WaitingPlayer, g.CurrentPlayer
	}

	announceResult(g, s)
}

func getBoard(board *[3][3]string) string {
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

	return boardStr.String()
}

func sendToSpectators(game *models.Game, msg string) {
	if game.Spectators != nil {
		for spectator := range *game.Spectators {
			_, err := spectator.Conn.Write([]byte(msg))
			if err != nil {
				spectator.Conn.Close()
				removeSpectator(game.Spectators, &spectator)
				continue
			}
			if game.OnGoing {
				_, err = spectator.Conn.Write([]byte(fmt.Sprintf("%s's turn:\n", game.CurrentPlayer.NickName)))
				if err != nil {
					spectator.Conn.Close()
					removeSpectator(game.Spectators, &spectator)
				}
			}
		}
	}
}

func removeSpectator(spectators *map[models.Spectator]struct{}, spectator *models.Spectator) {
	if spectators != nil {
		delete(*spectators, *spectator)
	}
}

func disconnectSpectators(spectators *map[models.Spectator]struct{}) {
	if spectators != nil {
		for s := range *spectators {
			s.Conn.Close()
		}
	}
}

func requestMove(player *models.Player) (string, error) {
	if err := sendMessage(player, "Your move (format: A1, B3, etc.): "); err != nil {
		return "", err
	}

	buffer := make([]byte, 1024)
	n, err := player.Conn.Read(buffer)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(buffer[:n])), nil
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
	resultMessage := ""
	result := models.GameResult{
		GameID:  g.ID,
		Player1: g.Player1,
		Player2: g.Player2,
		Winner:  nil,
		Loser:   nil,
		Error:   nil,
	}

	if g.Winner != nil {
		result.Winner = g.Winner
		result.Loser = g.Loser
		resultMessage = fmt.Sprintf("Game Over. %s wins!\n", g.Winner.NickName)
	} else {
		resultMessage = "Game Over. It's a draw!\n"
	}

	if err := sendMessage(&g.Player1, resultMessage); err != nil {
		handleError(g, s, err)
		return
	}
	if err := sendMessage(&g.Player2, resultMessage); err != nil {
		handleError(g, s, err)
		return
	}

	sendToSpectators(g, resultMessage)

	disconnectSpectators(g.Spectators)
	g.Player1.Conn.Close()
	g.Player2.Conn.Close()

	delete(s.Games, g.ID)
	delete(s.ActiveUsers, g.Player1.NickName)
	delete(s.ActiveUsers, g.Player2.NickName)

	s.ResultsChan <- result
}

func sendMessage(player *models.Player, message string) error {
	_, err := player.Conn.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to send message to %s: %w", player.NickName, err)
	}
	return nil
}

func handleError(g *models.Game, s *models.Server, err error) {
	fmt.Printf("Error occurred in game %s: %s\n", g.ID, err)

	errorMessage := fmt.Sprintf("Game Over due to an error: %s\n", err.Error())

	sendMessage(g.CurrentPlayer, errorMessage)
	sendMessage(g.WaitingPlayer, errorMessage)
	for spectator := range *g.Spectators {
		spectator.Conn.Write([]byte(errorMessage))
	}

	disconnectSpectators(g.Spectators)
	g.Player1.Conn.Close()
	g.Player2.Conn.Close()

	delete(s.Games, g.ID)
	delete(s.ActiveUsers, g.Player1.NickName)
	delete(s.ActiveUsers, g.Player2.NickName)

	result := models.GameResult{
		GameID:  g.ID,
		Player1: g.Player1,
		Player2: g.Player2,
		Winner:  nil,
		Loser:   nil,
		Error:   err,
	}
	s.ResultsChan <- result
}
