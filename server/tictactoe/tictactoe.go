package tictactoe

import (
	"sync"
)

type Symbol string

const (
	Empty Symbol = ""
	X     Symbol = "X"
	O     Symbol = "O"
)

type EventType int

const (
	NoEvent    EventType = 0
	WinEvent   EventType = 1
	MoveEvent  EventType = 2
	EndedEvent EventType = 3
)

type GameState struct {
	ID     GameID    `json:"id"`
	Event  EventType `json:"event"`
	Board  []Symbol  `json:"board"`
	Turn   Symbol    `json:"turn"`
	Winner Symbol    `json:"winner"`
}

type GameID string

type TicTacToe interface {
	ListGames() []string
	CreateGame(id GameID, symbol Symbol) error
	JoinGame(id GameID) (*JoinResponse, error)
	EndGame(id GameID) error
	GetGame(id GameID) (*GameState, error)
	GameStream(gameID GameID, id string) (chan GameState, error)
	DeleteStream(gameID GameID, id string) error
	Move(id GameID, symbol Symbol, index int) (*GameState, error)
}

func NewTicTacToe() TicTacToe {
	return &ttt{
		games: map[GameID]*game{},
	}
}

type game struct {
	board       []Symbol
	streams     map[string]chan GameState
	streamMutex sync.Mutex
	turn        Symbol
	playerX     bool
	playerO     bool
}

type ttt struct {
	games map[GameID]*game
}

func (t *ttt) ListGames() []string {
	ids := []string{}
	for i := range t.games {
		ids = append(ids, string(i))
	}
	return ids
}

func (t *ttt) CreateGame(id GameID, symbol Symbol) error {

	if _, ok := t.games[GameID(id)]; ok {
		return &GameExistsErr{}
	}

	t.games[GameID(id)] = &game{
		board:   make([]Symbol, 9),
		streams: map[string]chan GameState{},
		// player who created the game goes first
		turn: symbol,
	}

	if symbol == X {
		t.games[GameID(id)].playerX = true
	} else {
		t.games[GameID(id)].playerO = true
	}

	return nil
}

type JoinResponse struct {
	Symbol Symbol    `json:"symbol"`
	State  GameState `json:"state"`
}

func (t *ttt) JoinGame(id GameID) (*JoinResponse, error) {

	game, ok := t.games[GameID(id)]
	if !ok {
		return nil, &GameNotFoundErr{}
	}

	sym := Empty
	if !game.playerX {
		game.playerX = true
		sym = X
	} else if !game.playerO {
		game.playerO = true
		sym = O
	} else {
		return nil, &TooManyPlayersErr{}
	}

	return &JoinResponse{
		Symbol: sym,
		State: GameState{
			ID:    id,
			Event: NoEvent,
			Board: game.board,
			Turn:  game.turn,
		},
	}, nil
}

func (t *ttt) EndGame(id GameID) error {

	game, ok := t.games[GameID(id)]
	if !ok {
		return &GameNotFoundErr{}
	}

	t.send(game, GameState{
		Event: EndedEvent,
	})

	delete(t.games, GameID(id))

	return nil
}

func (t *ttt) GetGame(id GameID) (*GameState, error) {

	game, ok := t.games[GameID(id)]
	if !ok {
		return nil, &GameNotFoundErr{}
	}

	return &GameState{
		ID:    id,
		Event: NoEvent,
		Board: game.board,
		Turn:  game.turn,
	}, nil
}

func (t *ttt) GameStream(gameID GameID, id string) (chan GameState, error) {

	game, ok := t.games[gameID]
	if !ok {
		return nil, &GameNotFoundErr{}
	}

	game.streamMutex.Lock()
	defer game.streamMutex.Unlock()

	stream := make(chan GameState, 1)
	game.streams[id] = stream

	return stream, nil
}

func (t *ttt) DeleteStream(gameID GameID, id string) error {

	game, ok := t.games[gameID]
	if !ok {
		return &GameNotFoundErr{}
	}

	game.streamMutex.Lock()
	defer game.streamMutex.Unlock()

	delete(game.streams, id)
	return nil
}

func (t *ttt) Move(gameID GameID, symbol Symbol, index int) (*GameState, error) {

	game, ok := t.games[gameID]
	if !ok {
		return nil, &GameNotFoundErr{}
	}

	if game.turn != symbol {
		return nil, &NotYourTurnErr{}
	}

	if game.board[index] != Empty {
		return nil, &IllegalMoveErr{}
	}

	game.board[index] = symbol

	if game.IsWon(symbol, index) {

		state := &GameState{
			Event:  WinEvent,
			Winner: symbol,
			Board:  game.board,
		}

		t.send(game, *state)
		return state, nil
	}

	if game.turn == X {
		game.turn = O
	} else if game.turn == O {
		game.turn = X
	}

	state := &GameState{
		Event: MoveEvent,
		Turn:  game.turn,
		Board: game.board,
	}
	t.send(game, *state)

	return state, nil
}

func (g *game) IsWon(symbol Symbol, i int) bool {

	b := g.board

	// if center is last symbol
	if b[4] == symbol {

		// middle horizontal
		if b[3] == b[4] && b[4] == b[5] {
			return true
		}

		// middle vertical
		if b[1] == b[4] && b[4] == b[7] {
			return true
		}

		// diag
		if b[0] == b[4] && b[4] == b[8] {
			return true
		}

		// anti diag
		if b[2] == b[4] && b[4] == b[6] {
			return true
		}
	}

	// if top left is last symbol
	if b[0] == symbol {

		// top horizontal
		if b[0] == b[1] && b[1] == b[2] {
			return true
		}

		// left vertical
		if b[0] == b[3] && b[3] == b[6] {
			return true
		}
	}

	// if bottom right is last symbol
	if b[8] == symbol {

		// bottom horizontal
		if b[6] == b[7] && b[7] == b[8] {
			return true
		}

		// right vertical
		if b[2] == b[5] && b[5] == b[8] {
			return true
		}
	}

	return false
}

func (t *ttt) send(game *game, state GameState) {
	for i := range game.streams {
		game.streams[i] <- state
	}
}

func Hash(b []Symbol) string {
	h := ""
	for _, v := range b {
		if v == "" {
			h += "-"
		} else {
			h += string(v)
		}
	}
	return h
}
