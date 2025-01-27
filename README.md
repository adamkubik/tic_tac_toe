# Tic-Tac-Toe multiplayer server

This project is a **TCP server** built in Go allowing players to connect and play a famous multiplayer game Tic-Tac-Toe.

## Features

- **Multiplayer Support**: Multiple players can connect to the server to play Tic-Tac-Toe.
- **Real-Time Gameplay**: Players make their moves in real-time, the server updates the game state accordingly and senss it to both players.
- **Game State Management**: The server ensures that game rules are followed, and it determines the winner or a draw.
- **Concurrency**: The server is designed to handle multiple players and games concurrently using Goroutines and Channels.
- **Spectator mode**: The players can choose to spectate one of the ongoing games and the server will then send them each move played in that game in real time.
- **Statistics**: Before starting a game, each player can view theis stats or the leaderboard.

## How It Works

1. Players connect to the server via a TCP client (e.g., Telnet or a custom client).
2. They can then choose to play (or view stats) or spectate one of the ongoing games.
3. The server pairs players into game sessions.
4. Each player takes turns making moves, with the server validating the input and updating the game state.
5. The game ends when one player wins or the game results in a draw. The server notifies both players and every spectator of the outcome.
6. Players are then automatically disconnected.

## Prerequisites

- Go 1.23.5 or later
- A TCP client to connect to the server (e.g., Telnet, Netcat, or a custom-built client)

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/adamkubik/tic_tac_toe.git
   cd tic_tac_toe
   ```
2. Build the server:
   ```bash
   go build -o tictactoe cmd/tic_tac_toe/main.go
   ```
3. Run the server:
   ```bash
   ./tictactoe
   ```

## Usage

1. Start the server as described in the installation section.
2. Connect to the server using a TCP client:
   ```bash
   telnet 34.118.38.74 23
   ```
3. Follow the prompts to join a game and start playing.
4. To quit, disconnect from the client.

## Game Rules

- The game is played on a 3x3 grid.
- Players take turns placing their symbol (`X` or `O`) in an empty cell by typing the cell coordinates (A1-C3).
- The first player to align three symbols horizontally, vertically, or diagonally wins.
- If all cells are filled without a winner or no possible combination results in a clear win, the game ends in a draw.


## Contributing

Contributions are welcome! Please fork the repository, create a new branch for your feature or bug fix, and submit a pull request.


## Authors

- [Adam Kubík](https://github.com/adamkubik)
- [Tomáš Karol Hőger](https://github.com/TomasKarolHoger)

---
Thank you for using our Tic-Tac-Toe multiplayer server! If you have any questions or suggestions, feel free to open an issue in the repository.

