// Package search handles all searching and pruning logic.
package search

import (
	"math/rand/v2"

	"github.com/hellosam123/go-chess/internal/board"
)

func RandomMove(b *board.Board) board.Move {
	moves := b.GenerateLegalMoves()
	moveIndex := rand.IntN(len(moves))
	return moves[moveIndex]
}
