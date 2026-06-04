// Package search handles all searching and pruning logic.
package search

import (
	"math/rand/v2"

	"github.com/hellosam123/go-chess/internal/board"
	eval "github.com/hellosam123/go-chess/internal/evaluation"
)

func RandomMove(b *board.Board) board.Move {
	moves := b.GenerateLegalMoves()
	moveIndex := rand.IntN(len(moves))
	return moves[moveIndex]
}

func Search(b *board.Board, depth int) (int, board.Move) {
	var bestEval int = -100000
	var bestMove board.Move

	if depth == 0 {
		bestEval = eval.Evaluate(b)
		return bestEval, 0
	}

	moves := b.GenerateLegalMoves()
	for _, m := range moves {
		unMove := b.MakeMove(m)

		score, _ := Search(b, depth-1)
		currentEval := -score
		b.UnMakeMove(m, unMove)

		if currentEval > bestEval {
			bestEval = currentEval
			bestMove = m
		}
	}

	return bestEval, bestMove
}
