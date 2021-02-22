package main

import (
	"fmt"
)

func main() {
	fmt.Println("Tic Tac Toe Server has started.")
	s := NewServer()
	s.Start()
}
