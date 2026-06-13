package eval

import (
	"testing"

	"github.com/hellosam123/go-chess/internal/board"
	"github.com/hellosam123/go-chess/internal/squares"
)

func TestStaticExchangeEval(t *testing.T) {
	b := board.NewStartingBoard()
	b.ParseFEN("1k1r4/1pp4p/p7/4p3/8/P5P1/1PP4P/2K1R3 w - -")
	b.PrintBoard()

	t.Log(StaticExchangeEval(b, squares.E5, true))
}
