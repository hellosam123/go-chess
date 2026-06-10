package search

import (
	"testing"
	"time"

	"github.com/hellosam123/go-chess/internal/board"
	eval "github.com/hellosam123/go-chess/internal/evaluation"
)

func TestSearch(t *testing.T) {
	b := board.NewStartingBoard()
	tt := *eval.NewTranspositionTable(32)
	time, _ := time.ParseDuration("1s")
	t.Log(RootSearch(b, time, &tt))
}

func BenchmarkRandomMove(b *testing.B) {
	gameBoard := board.NewStartingBoard()
	for i := 0; i < b.N; i++ {
		RandomMove(gameBoard)
	}
}
