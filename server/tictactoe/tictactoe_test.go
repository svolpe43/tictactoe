package tictactoe

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTicTacToe(t *testing.T) {

	ttt := NewTicTacToe()

	ttt.CreateGame("sucker", X)

	ids := ttt.ListGames()
	fmt.Println(ids)

	stream, err := ttt.GameStream("sucker", "asdf")
	require.NoError(t, err)

	ticker := time.NewTicker(500 * time.Millisecond)

	go func() {
		for {
			select {
			case data := <-stream:
				fmt.Println("Received game data", data)
			case t := <-ticker.C:
				fmt.Println("Tick at", t)
			}
		}
	}()

	_, err = ttt.Move("sucker", "O", 0)
	require.NoError(t, err)

	_, err = ttt.Move("sucker", "X", 6)
	require.NoError(t, err)

	_, err = ttt.Move("sucker", "O", 1)
	require.NoError(t, err)

	_, err = ttt.Move("sucker", "X", 7)
	require.NoError(t, err)

	_, err = ttt.Move("sucker", "O", 2)
	require.NoError(t, err)

}

func TestHash(t *testing.T) {
	a := []Symbol{"X", "X", "X", "", "", "", "X", "X", "X"}
	require.Equal(t, Hash(a), "111---111")

	b := []Symbol{"X", "O", "X", "-", "O", "-", "X", "X", "X"}
	require.Equal(t, Hash(b), "121-2-111")
}
