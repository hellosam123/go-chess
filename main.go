package main

import (
	"fmt"

	"github.com/hellosam123/go-chess/internal/board"
)

func main() {
	fmt.Println("A Golang chess engine")
	gameBoard := board.NewStartingBoard()

	gameBoard.PrintBoard()

	gameBoard.ParseFEN("r1bqkb1r/ppp2ppp/2n5/3pP3/P2N4/8/1PP2PPP/RNBQ1RK1 w kq d6 0 9")

	gameBoard.PrintBoard()
}
