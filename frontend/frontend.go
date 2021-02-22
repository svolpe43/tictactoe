package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/svolpe43/ttt/server/tictactoe"
)

var game *Game

type Game struct {
	id     tictactoe.GameID
	board  []tictactoe.Symbol
	symbol tictactoe.Symbol
	turn   tictactoe.Symbol
}

func main() {

	var (
		ctx    = context.Background()
		client = NewClient()
		reader = bufio.NewReader(os.Stdin)
	)

	fmt.Println("Welcome to Tic Tak Toe!")

	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		args := strings.Split(text, " ")

		switch args[0] {
		case "list":
			if len(args) != 1 {
				fmt.Println("Usage: list")
				continue
			}

			games, err := client.ListGames(ctx)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(games)

		case "create":

			if game != nil {
				fmt.Println("There is already a game in progress, first end the game.")
				continue
			}

			if len(args) != 3 {
				fmt.Println("Usage: create <game name> <piece>")
				continue
			}

			symbol := tictactoe.X
			if args[2] == "O" {
				symbol = tictactoe.O
			}

			if err := client.CreateGame(ctx, args[1], symbol); err != nil {
				fmt.Println(err)
				continue
			}

			game = &Game{
				id:     tictactoe.GameID(args[1]),
				board:  make([]tictactoe.Symbol, 9),
				symbol: symbol,
				turn:   symbol,
			}

			render(game)
			fmt.Println()
			fmt.Println("Your turn!")
			fmt.Println()

		case "join":

			if game != nil {
				fmt.Println("You are already connected to a game")
				continue
			}

			if len(args) != 2 {
				fmt.Println("Usage: join <game name>")
				continue
			}

			resp, err := client.JoinGame(ctx, args[1])
			if err != nil {
				fmt.Println(err)
				continue
			}

			game = &Game{
				id:     resp.State.ID,
				board:  resp.State.Board,
				symbol: resp.Symbol,
				turn:   resp.State.Turn,
			}

			render(game)

			go longPoll(ctx, client)

		case "end":
			if len(args) != 2 {
				fmt.Println("Usage: end <game name>")
				continue
			}

			if err := client.EndGame(ctx, args[1]); err != nil {
				fmt.Println(err)
				continue
			}

			game = nil

			fmt.Println("Ended game", args[1])

		case "move":
			if len(args) != 2 {
				fmt.Println("Usage: move <index>")
				continue
			}

			if game == nil {
				fmt.Println("You must create or join a game first")
				continue
			}

			index, err := strconv.ParseInt(args[1], 10, 0)
			if err != nil {
				fmt.Println("Cannot parse index")
				continue
			}

			state, err := client.Move(ctx, game.id, game.symbol, int(index))
			if err != nil {
				fmt.Println(err)
				continue
			}

			game.board = state.Board
			game.turn = state.Turn

			if state.Winner != tictactoe.Empty {
				render(game)
				game = nil
				fmt.Println()
				fmt.Println("Winner!", state.Winner)
				continue
			}

			render(game)

			go longPoll(ctx, client)

		default:
			fmt.Println("Unknown command")
		}
	}
}

func longPoll(ctx context.Context, client Client) {

	for {

		state, code, err := client.GetGame(ctx, string(game.id), tictactoe.Hash(game.board))
		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Second * 3)
			continue
		}

		if code == http.StatusGatewayTimeout || code == http.StatusRequestTimeout {
			time.Sleep(time.Second * 3)
			continue
		}

		if code != http.StatusOK {
			time.Sleep(time.Second * 3)
			continue
		}

		game.board = state.Board
		game.turn = state.Turn

		render(game)

		if state.Winner != tictactoe.Empty {
			fmt.Println()
			fmt.Println("Winner!", state.Winner)
			game = nil
			break
		}

		fmt.Println()
		fmt.Println("Your turn!")
		fmt.Println()
		fmt.Print("-> ")
		break
	}
}

func render(game *Game) {
	if len(game.board) == 9 {

		b := []string{}
		for i := range game.board {
			b = append(b, empty(game.board[i]))
		}

		fmt.Println()
		fmt.Printf("Playing game \"%s\" as %s\n", game.id, game.symbol)
		fmt.Printf("Turn: %s\n", game.turn)
		fmt.Printf(" %s | %s | %s \n", b[0], b[1], b[2])
		fmt.Println("-----------")
		fmt.Printf(" %s | %s | %s \n", b[3], b[4], b[5])
		fmt.Println("-----------")
		fmt.Printf(" %s | %s | %s \n", b[6], b[7], b[8])
	}
}

func empty(s tictactoe.Symbol) string {
	sym := s
	if sym == "" {
		sym = " "
	}
	return string(sym)
}
