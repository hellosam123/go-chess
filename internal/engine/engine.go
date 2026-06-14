// Package engine handles the initialization of the engine struct
package engine

import (
	"github.com/hellosam123/go-chess/internal/board"
	eval "github.com/hellosam123/go-chess/internal/evaluation"
)

type Engine struct {
	HistoryTable *[2][64][64]int
	Board        *board.Board
	TT           *eval.TranspositionTable
}

func NewEngine(ttSizeMB int) *Engine {
	return &Engine{
		HistoryTable: new([2][64][64]int),
		Board:        board.NewStartingBoard(),
		TT:           eval.NewTranspositionTable(ttSizeMB),
	}
}

func (e *Engine) ResetEngine() {
	*e.HistoryTable = [2][64][64]int{}
	e.TT.ResetTable()
	e.Board.ResetBoard()
}
