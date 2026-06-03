package main

import (
	"fmt"

	"github.com/hellosam123/go-chess/internal/board"
)

func main() {
	fmt.Println("A Golang chess engine")
	gameBoard := board.NewStartingBoard()
	gameBoard.ParseFEN("r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1")

	gameBoard.PrintBoard()

	moves := board.GenerateLegalMoves(gameBoard)
	fmt.Println(len(moves))
	for _, move := range moves {
		fmt.Println(move.ToString())
	}
}
