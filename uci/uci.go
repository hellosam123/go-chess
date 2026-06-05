// Package uci handles text commands from chess GUIs using the UCI protocol
package uci

import (
	"fmt"
	"os"
	"strings"

	"github.com/hellosam123/go-chess/internal/board"
	"github.com/hellosam123/go-chess/internal/search"
)

func MatchUCIString(b *board.Board, str string) (board.Move, error) {
	var moves []board.Move = b.GenerateLegalMoves()
	for _, move := range moves {
		if str == move.MoveToString() {
			return move, nil
		}
	}

	return 0, fmt.Errorf("Invalid or illegal UCI string: %s", str)
}

func HandlePosition(b *board.Board, args []string) error {
	if len(args) == 0 {
		return nil
	}

	currentIndex := 0

	if args[0] == "startpos" {
		b.ParseFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
		currentIndex = 1
	} else if args[0] == "fen" {
		fenParts := []string{}
		for currentIndex = 1; currentIndex < len(args); currentIndex++ {
			if args[currentIndex] == "moves" {
				break
			}
			fenParts = append(fenParts, args[currentIndex])
		}
		fen := strings.Join(fenParts, " ")
		b.ParseFEN(fen)
	}

	if currentIndex < len(args) && args[currentIndex] == "moves" {
		for _, moveStr := range args[currentIndex+1:] {
			move, err := MatchUCIString(b, moveStr)
			if err != nil {
				return err
			}

			b.MakeMove(move)
		}
	}
	return nil
}

func HandleGo(b *board.Board, args []string) {
	score, move := search.Search(b, 5)
	moveStr := move.MoveToString()
	fmt.Printf("info score cp %d\n", score)
	fmt.Printf("bestmove %s\n", moveStr)

	os.Stdout.Sync()
}
