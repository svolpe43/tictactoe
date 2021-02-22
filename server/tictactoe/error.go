package tictactoe

type GameNotFoundErr struct {
}

func (g *GameNotFoundErr) Error() string {
	return "Game not found"
}

type GameExistsErr struct {
}

func (g *GameExistsErr) Error() string {
	return "Game with that name already exists"
}

type NotYourTurnErr struct {
}

func (g *NotYourTurnErr) Error() string {
	return "Not your turn"
}

type TooManyPlayersErr struct {
}

func (g *TooManyPlayersErr) Error() string {
	return "Too many players in this game to join"
}

type IllegalMoveErr struct {
}

func (g *IllegalMoveErr) Error() string {
	return "Illegal move."
}
