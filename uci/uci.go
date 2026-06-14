// Package uci handles text commands from chess GUIs using the UCI protocol
package uci

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hellosam123/go-chess/internal/board"
	"github.com/hellosam123/go-chess/internal/engine"
	"github.com/hellosam123/go-chess/internal/search"
)

func MatchUCIString(b *board.Board, str string) (board.Move, error) {
	var moves []board.Move
	moves, _ = b.GenerateLegalMoves()
	for _, move := range moves {
		if str == move.MoveToString() {
			return move, nil
		}
	}

	return 0, fmt.Errorf("Invalid or illegal UCI string: %s", str)
}

func HandlePosition(e *engine.Engine, args []string) error {
	if len(args) == 0 {
		return nil
	}

	currentIndex := 0

	if args[0] == "startpos" {
		e.Board.ParseFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
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
		e.Board.ParseFEN(fen)
	}

	if currentIndex < len(args) && args[currentIndex] == "moves" {
		for _, moveStr := range args[currentIndex+1:] {
			move, err := MatchUCIString(e.Board, moveStr)
			if err != nil {
				return err
			}

			e.Board.MakeMove(move)
		}
	}
	return nil
}

func HandleGo(e *engine.Engine, args []string) error {
	var timeLeft int  // in ms
	var increment int // in ms
	var err error

	if len(args) == 0 {
		timeLeft = 2000
		increment = 100
	} else {
		for i := 0; i < len(args); i++ {
			switch args[i] {
			case "wtime":
				if e.Board.ActiveColor {
					if i+1 < len(args) {
						i++
						timeLeft, err = strconv.Atoi(args[i])
						if err != nil {
							return fmt.Errorf("%v", err)
						}
					}
				}
			case "btime":
				if !e.Board.ActiveColor {
					if i+1 < len(args) {
						i++
						timeLeft, err = strconv.Atoi(args[i])
						if err != nil {
							return fmt.Errorf("%v", err)
						}
					}
				}
			case "winc":
				if e.Board.ActiveColor {
					if i+1 < len(args) {
						i++
						increment, err = strconv.Atoi(args[i])
						if err != nil {
							return fmt.Errorf("%v", err)
						}
					}
				}
			case "binc":
				if !e.Board.ActiveColor {
					if i+1 < len(args) {
						i++
						increment, err = strconv.Atoi(args[i])
						if err != nil {
							return fmt.Errorf("%v", err)
						}
					}
				}
			}
		}
	}

	searchTimeBudget := time.Duration(timeLeft/20+increment/2) * time.Millisecond

	s := search.NewSearch(e, searchTimeBudget, false)
	move, score, depth, nodes, elapsed := s.RootSearch()
	if !e.Board.ActiveColor {
		score = -score
	}

	moveStr := move.MoveToString()

	var mateThreshold int = 29000
	var mateScore int = 30000

	if score > mateThreshold {
		pliesToMate := mateScore - score
		movesToMate := (pliesToMate + 1) / 2
		fmt.Printf("info depth %d time %d nodes %d score mate %d\n", depth, elapsed, nodes, movesToMate)
	} else if score < -mateThreshold {
		pliesToMate := -mateScore - score
		movesToMate := (pliesToMate + 1) / 2
		fmt.Printf("info depth %d time %d nodes %d score mate %d\n", depth, elapsed, nodes, movesToMate)
	} else {
		fmt.Printf("info depth %d time %d nodes %d score cp %d\n", depth, elapsed, nodes, score)
	}
	fmt.Printf("bestmove %s\n", moveStr)

	os.Stdout.Sync()

	return nil
}
