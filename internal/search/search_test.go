package search

import (
	"testing"

	"github.com/hellosam123/go-chess/internal/board"
)

func TestSearch(t *testing.T) {
	b := board.NewStartingBoard()
	t.Log(Search(b, 5))
}
