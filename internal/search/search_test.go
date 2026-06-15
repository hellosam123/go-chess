package search

import (
	"testing"
	"time"

	"github.com/hellosam123/go-chess/internal/board"
	"github.com/hellosam123/go-chess/internal/engine"
)

func TestSearch(t *testing.T) {
	e := engine.NewEngine(32)

	time, _ := time.ParseDuration("1.000s")
	s := NewSearch(e.Board, e.TT, e.HistoryTable, &e.SearchAbort, time, false)
	t.Log(s.RootSearch())
}

func BenchmarkRandomMove(b *testing.B) {
	gameBoard := board.NewStartingBoard()
	gameBoard.ParseFEN("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq -")
	for i := 0; i < b.N; i++ {
		RandomMove(gameBoard)
	}
}
