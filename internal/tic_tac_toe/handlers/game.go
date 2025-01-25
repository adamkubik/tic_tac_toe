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

	s.ActiveGamesMu.Lock()
	s.Games[gameId] = &g
	s.ActiveGamesMu.Unlock()

	if err := sendMessageToPlayer(g.CurrentPlayer, "The game is starting... you're player 'X'\n"); err != nil {
		handleError(&g, s, err)
		return
	}
	if err := sendMessageToPlayer(g.WaitingPlayer, "The game is starting... you're player '0'\n"); err != nil {
		handleError(&g, s, err)
		return
	}

	playGame(&g, s)
}

func playGame(g *models.Game, s *models.Server) {
	for g.OnGoing {
		board := getBoard(g.Board)
		if err := sendMessageToPlayer(g.CurrentPlayer, board); err != nil {
			handleError(g, s, err)
			return
		}
		if err := sendMessageToPlayer(g.WaitingPlayer, "Waiting for your oponent's turn...\n"); err != nil {
			handleError(g, s, err)
			return
		}
		sendToSpectators(g, board)

		if err := tryGetMove(g); err != nil {
			handleError(g, s, err)
			return
		}

		board = getBoard(g.Board)
		if checkWin(g.Board, g.CurrentPlayer.Symbol) {
			g.Winner = g.CurrentPlayer
			g.Loser = g.WaitingPlayer
			if err := sendFinalBoard(g, board, s); err != nil {
				return
			}
			break
		} else if isDraw(g.Board) {
			if err := sendFinalBoard(g, board, s); err != nil {
				return
			}
			break
		}

		if err := sendMessageToPlayer(g.CurrentPlayer, board); err != nil {
			handleError(g, s, err)
			return
		}
		g.CurrentPlayer, g.WaitingPlayer = g.WaitingPlayer, g.CurrentPlayer
	}

	announceResult(g, s)
}

func sendFinalBoard(g *models.Game, board string, s *models.Server) error {
	g.OnGoing = false
	if err := sendMessageToPlayer(g.CurrentPlayer, board); err != nil {
		handleError(g, s, err)
		return err
	}
	if err := sendMessageToPlayer(g.WaitingPlayer, board); err != nil {
		handleError(g, s, err)
		return err
	}
	sendToSpectators(g, board)
	return nil
}

func tryGetMove(g *models.Game) error {
	var row, col int
	for {
		move, err := requestMove(g.CurrentPlayer)
		if err != nil {
			return err
		}

		row, col, err = validateMove(move, g.Board)
		if err != nil {
			if err := sendMessageToPlayer(g.CurrentPlayer, fmt.Sprintf("Invalid move: %s. Try again.\n", err.Error())); err != nil {
				return err
			}
			continue
		}
		updateBoard(g, row, col)
		break
	}

	return nil
}

func updateBoard(g *models.Game, row int, col int) {
	g.Board[row][col] = g.CurrentPlayer.Symbol
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
	if game.Spectators == nil {
		return
	}

	for spectator := range *game.Spectators {
		_, err := spectator.Conn.Write([]byte(msg))
		if err != nil {
			spectator.Conn.Close()
			removeSpectator(game, &spectator)
			continue
		}
		if game.OnGoing {
			_, err = spectator.Conn.Write([]byte(fmt.Sprintf("%s's turn:\n", game.CurrentPlayer.NickName)))
			if err != nil {
				spectator.Conn.Close()
				removeSpectator(game, &spectator)
			}
		}
	}
}

func removeSpectator(game *models.Game, spectator *models.Spectator) {
	if game.Spectators != nil {
		delete(*game.Spectators, *spectator)
	}
}

func disconnectSpectators(game *models.Game) {
	if game.Spectators != nil {
		for s := range *game.Spectators {
			s.Conn.Close()
		}
	}
}

func requestMove(player *models.Player) (string, error) {
	if err := sendMessageToPlayer(player, "Your move (format: A1, B3, etc.): "); err != nil {
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
	winPatterns := [][][2]int{
		// Rows
		{{0, 0}, {0, 1}, {0, 2}},
		{{1, 0}, {1, 1}, {1, 2}},
		{{2, 0}, {2, 1}, {2, 2}},
		// Columns
		{{0, 0}, {1, 0}, {2, 0}},
		{{0, 1}, {1, 1}, {2, 1}},
		{{0, 2}, {1, 2}, {2, 2}},
		// Diagonals
		{{0, 0}, {1, 1}, {2, 2}},
		{{0, 2}, {1, 1}, {2, 0}},
	}

	canWin := func(player string) bool {
		for _, pattern := range winPatterns {
			count := 0
			empty := 0
			for _, pos := range pattern {
				row, col := pos[0], pos[1]
				if board[row][col] == player {
					count++
				} else if board[row][col] == " " {
					empty++
				}
			}
			if count > 0 && count+empty == 3 {
				return true
			}
		}
		return false
	}

	if !canWin("X") && !canWin("O") {
		return true
	}

	return false
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

	if err := sendMessageToPlayer(&g.Player1, resultMessage); err != nil {
		handleError(g, s, err)
		return
	}
	if err := sendMessageToPlayer(&g.Player2, resultMessage); err != nil {
		handleError(g, s, err)
		return
	}

	sendToSpectators(g, resultMessage)

	disconnectSpectators(g)
	g.Player1.Conn.Close()
	g.Player2.Conn.Close()

	s.ActiveGamesMu.Lock()
	delete(s.Games, g.ID)
	s.ActiveGamesMu.Unlock()

	s.ActiveUsersMu.Lock()
	delete(s.ActiveUsers, g.Player1.NickName)
	delete(s.ActiveUsers, g.Player2.NickName)
	s.ActiveUsersMu.Unlock()

	s.ResultsChan <- result
}

func sendMessageToPlayer(player *models.Player, message string) error {
	_, err := player.Conn.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to send message to %s: %w", player.NickName, err)
	}
	return nil
}

func handleError(g *models.Game, s *models.Server, err error) {
	fmt.Printf("Error occurred in game %s: %s\n", g.ID, err)

	errorMessage := fmt.Sprintf("Game Over due to an error: %s\n", err.Error())

	sendMessageToPlayer(g.CurrentPlayer, errorMessage)
	sendMessageToPlayer(g.WaitingPlayer, errorMessage)
	for spectator := range *g.Spectators {
		spectator.Conn.Write([]byte(errorMessage))
	}

	disconnectSpectators(g)
	g.Player1.Conn.Close()
	g.Player2.Conn.Close()

	s.ActiveGamesMu.Lock()
	delete(s.Games, g.ID)
	s.ActiveGamesMu.Unlock()

	s.ActiveUsersMu.Lock()
	delete(s.ActiveUsers, g.Player1.NickName)
	delete(s.ActiveUsers, g.Player2.NickName)
	s.ActiveUsersMu.Unlock()

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
