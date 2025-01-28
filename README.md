# Tic-Tac-Toe multiplayer server

This project is a **TCP server** built in Go allowing players to connect and play a famous multiplayer game Tic-Tac-Toe.

## Features

- **Multiplayer support**: Multiple players can connect to the server to play Tic-Tac-Toe.
- **Real-Time gameplay**: Players make their moves in real-time, the server updates the game state accordingly and sends it to both players.
- **Game state management**: The server ensures that game rules are followed, and it determines the winner or a draw.
- **Concurrency**: The server is designed to handle multiple players and games concurrently using Goroutines and Channels.
- **Spectator mode**: The players can choose to spectate one of the ongoing games and the server will be sending them each move played in that game in real time.
- **Player statistics**: A connected database stores player statistics, including wins, losses, and draws.

## How it works

1. Players connect to the server via a TCP client (e.g., Telnet or a custom client).
2. They can then choose to play (or view statistics) or spectate one of the ongoing games.
3. The server pairs players into game sessions.
4. Each player takes turns making moves, with the server validating the input and updating the game state.
5. The game ends when one player wins or the game results in a draw. The server notifies both players and every spectator of the outcome.
6. Players are then automatically disconnected.

## Deployment

The server is deployed as a **Docker image** and runs on Google Cloud Platform. It can be accessed at:

```
34.118.38.74:23
```

## Prerequisites

- A TCP client to connect to the server (e.g., Telnet, Netcat, or a custom-built client)
- Go version 1.23.5 or later (if run locally)

## Usage

If you wish to run the server locally, follow the Installation steps below, if not, skip to step 4:

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

1. Connect to the server using a TCP client (if you're running the server locally, connect to 0.0.0.0:23):
   ```bash
   telnet 34.118.38.74 23
   ```
2. Follow the prompts to join a game and start playing.
3. To quit, disconnect from the client.


## Game Rules

- The game is played on a 3x3 grid.
- Players take turns placing their symbol (`X` or `O`) in an empty cell by typing the cell coordinates (A1-C3).
- The first player to align three symbols horizontally, vertically, or diagonally wins.
- If all cells are filled without a winner the game ends in a draw.


## Contributing

Contributions are welcome! Please fork the repository, create a new branch for your feature or bug fix, and submit a pull request.


## Authors

- [Adam Kubík](https://github.com/adamkubik)
- [Tomáš Karol Hőger](https://github.com/TomasKarolHoger)

---
Thank you for using our Tic-Tac-Toe multiplayer server! If you have any questions or suggestions, feel free to open an issue in the repository.
