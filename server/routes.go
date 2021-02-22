package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	uuid "github.com/satori/go.uuid"
	"github.com/svolpe43/ttt/server/tictactoe"
)

const LongPollMaxWait = 30

type Server interface {
	Start()
	ListGames(w http.ResponseWriter, r *http.Request)
	JoinGame(w http.ResponseWriter, r *http.Request)
	CreateGame(w http.ResponseWriter, r *http.Request)
	EndGame(w http.ResponseWriter, r *http.Request)
	GetGame(w http.ResponseWriter, r *http.Request)
	Move(w http.ResponseWriter, r *http.Request)
}

func NewServer() Server {
	return &server{
		tictactoe: tictactoe.NewTicTacToe(),
	}
}

type server struct {
	tictactoe tictactoe.TicTacToe
}

func (s *server) Start() {

	r := chi.NewRouter()

	r.Get("/", s.ListGames)
	r.Post("/{id}/create/{symbol}", s.CreateGame)
	r.Get("/{id}/{hash}", s.GetGame)
	r.Post("/{id}/join", s.JoinGame)
	r.Post("/{id}/move/{symbol}/{index}", s.Move)
	r.Delete("/{id}/end", s.EndGame)

	fmt.Printf("Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}

func (s *server) ListGames(w http.ResponseWriter, r *http.Request) {

	games := s.tictactoe.ListGames()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(strings.Join(games, ",")))
}

func (s *server) CreateGame(w http.ResponseWriter, r *http.Request) {

	gameID := tictactoe.GameID(chi.URLParam(r, "id"))

	symbol := tictactoe.X
	if chi.URLParam(r, "symbol") == string('o') {
		symbol = tictactoe.O
	}

	if err := s.tictactoe.CreateGame(gameID, symbol); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *server) JoinGame(w http.ResponseWriter, r *http.Request) {

	gameID := tictactoe.GameID(chi.URLParam(r, "id"))

	resp, err := s.tictactoe.JoinGame(gameID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (s *server) EndGame(w http.ResponseWriter, r *http.Request) {

	gameID := tictactoe.GameID(chi.URLParam(r, "id"))

	if err := s.tictactoe.EndGame(gameID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *server) Move(w http.ResponseWriter, r *http.Request) {

	gameID := tictactoe.GameID(chi.URLParam(r, "id"))
	symbol := chi.URLParam(r, "symbol")

	indexStr := chi.URLParam(r, "index")
	index, err := strconv.ParseInt(indexStr, 10, 0)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Could not parse symbol"))
		return
	}

	state, err := s.tictactoe.Move(gameID, tictactoe.Symbol(symbol), int(index))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(state)
}

// GetGame is a long polling request that will listen to the
// event stream of a particular game and respond with the result.
func (s *server) GetGame(w http.ResponseWriter, r *http.Request) {

	gameID := tictactoe.GameID(chi.URLParam(r, "id"))
	hash := chi.URLParam(r, "hash")

	game, err := s.tictactoe.GetGame(gameID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// If the hash are not equal then just return the game
	// state as is. If not, we wait for an event to fire
	if hash != tictactoe.Hash(game.Board) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(game)
		return
	}

	id := uuid.NewV4().String()

	stream, err := s.tictactoe.GameStream(gameID, id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	ctx := r.Context()
	timeout := time.NewTimer(LongPollMaxWait * time.Second)

POLL:
	for {
		select {
		case <-ctx.Done():
			w.WriteHeader(http.StatusGatewayTimeout)
			w.Write([]byte("Request context has timed out"))
			break POLL
		case <-timeout.C:
			w.WriteHeader(http.StatusRequestTimeout)
			w.Write([]byte("Long poll request timed out"))
			break POLL
		case state := <-stream:

			// throw away messages with the same hash
			// these are sent from our moves
			if tictactoe.Hash(state.Board) == hash {
				continue
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(state)
			break POLL
		}
	}

	if err := s.tictactoe.DeleteStream(gameID, id); err != nil {
		fmt.Println("WARN: could not delete stream")
	}
}
