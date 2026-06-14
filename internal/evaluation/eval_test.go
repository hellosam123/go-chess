package eval

import (
	"testing"

	"github.com/hellosam123/go-chess/internal/board"
	"github.com/hellosam123/go-chess/internal/squares"
)

func TestStaticExchangeEval(t *testing.T) {
	b := board.NewStartingBoard()
	b.ParseFEN("1k1r3q/1ppn3p/p4b2/4p3/8/P2N2P1/1PP1R1BP/2K1Q3 w - -")
	b.PrintBoard()

	t.Log(StaticExchangeEval(b, squares.E5, true))
}
