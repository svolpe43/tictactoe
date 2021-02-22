package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/svolpe43/ttt/server/tictactoe"
)

const TicTacToeHost = "http://localhost:8080"

type Client interface {
	ListGames(ctx context.Context) ([]string, error)
	JoinGame(ctx context.Context, id string) (*tictactoe.JoinResponse, error)
	CreateGame(ctx context.Context, id string, symbol tictactoe.Symbol) error
	EndGame(ctx context.Context, id string) error
	GetGame(ctx context.Context, id, hash string) (*tictactoe.GameState, int, error)
	Move(ctx context.Context, id tictactoe.GameID, symbol tictactoe.Symbol, index int) (*tictactoe.GameState, error)
}

func NewClient() Client {
	return &client{
		host: TicTacToeHost,
	}
}

type client struct {
	host string
}

func (c *client) ListGames(ctx context.Context) ([]string, error) {

	url := c.host + "/"

	req, err := http.NewRequest(http.MethodGet, url, new(bytes.Buffer))
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Received status code - %d", resp.StatusCode)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(respBody), ","), nil
}

func (c *client) CreateGame(ctx context.Context, id string, symbol tictactoe.Symbol) error {

	sym := string('x')
	if symbol == tictactoe.O {
		sym = string('o')
	}

	url := c.host + "/" + id + "/create/" + sym

	req, err := http.NewRequest(http.MethodPost, url, new(bytes.Buffer))
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return errors.New(string(respBody))
	}

	return nil
}

func (c *client) JoinGame(ctx context.Context, id string) (*tictactoe.JoinResponse, error) {

	url := c.host + "/" + id + "/join"

	req, err := http.NewRequest(http.MethodPost, url, new(bytes.Buffer))
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(respBody))
	}

	joinResp := &tictactoe.JoinResponse{}

	if err := json.NewDecoder(bytes.NewReader(respBody)).Decode(&joinResp); err != nil {
		return nil, errors.New("could not decode join response")
	}

	return joinResp, nil
}

func (c *client) EndGame(ctx context.Context, id string) error {

	url := c.host + "/" + id + "/end"

	req, err := http.NewRequest(http.MethodDelete, url, new(bytes.Buffer))
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return errors.New(string(respBody))
	}

	return nil
}

func (c *client) GetGame(ctx context.Context, id, hash string) (*tictactoe.GameState, int, error) {

	url := c.host + "/" + id + "/" + hash

	req, err := http.NewRequest(http.MethodGet, url, new(bytes.Buffer))
	if err != nil {
		return nil, -1, err
	}

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, -1, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, -1, err
	}

	if resp.StatusCode == http.StatusOK {

		state := &tictactoe.GameState{}

		if err := json.NewDecoder(bytes.NewReader(respBody)).Decode(&state); err != nil {
			return nil, resp.StatusCode, errors.New("could not decode game state")
		}

		return state, http.StatusOK, nil
	}

	return nil, resp.StatusCode, nil
}

func (c *client) Move(ctx context.Context, id tictactoe.GameID, symbol tictactoe.Symbol, index int) (*tictactoe.GameState, error) {

	url := strings.Join([]string{
		c.host,
		string(id),
		"move",
		string(symbol),
		strconv.FormatInt(int64(index), 10),
	}, "/")

	req, err := http.NewRequest(http.MethodPost, url, new(bytes.Buffer))
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(respBody))
	}

	state := &tictactoe.GameState{}

	if err := json.NewDecoder(bytes.NewReader(respBody)).Decode(&state); err != nil {
		return nil, errors.New("could not decode game state")
	}

	return state, nil
}
