package board

import (
	"testing"
)

func Perft(b *Board, depth int) int {
	if depth == 0 {
		return 1
	}

	var nodes int = 0

	moves, _ := b.GenerateLegalMoves()
	for _, m := range moves {
		unMove := b.MakeMove(m)
		nodes += Perft(b, depth-1)
		b.UnMakeMove(m, unMove)
	}

	return nodes
}

func TestPerft(t *testing.T) {
	gameBoard := NewStartingBoard()
	gameBoard.ParseFEN("n1n5/PPPk4/8/8/8/8/4Kppp/5N1N b - - 0 1")
	t.Log(Perft(gameBoard, 5))
}
