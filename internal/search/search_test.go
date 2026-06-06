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
	time, _ := time.ParseDuration("500ms")
	t.Log(RootSearch(b, time, &tt))
}
