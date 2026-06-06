// Package search handles all searching and pruning logic.
package search

import (
	"math/rand/v2"
	"time"

	"github.com/hellosam123/go-chess/internal/board"
	eval "github.com/hellosam123/go-chess/internal/evaluation"
)

func RandomMove(b *board.Board) board.Move {
	moves, _ := b.GenerateLegalMoves()
	moveIndex := rand.IntN(len(moves))
	return moves[moveIndex]
}

// RootSearch uses iterative deepening and initializes the alpha-beta search.
func RootSearch(b *board.Board, searchTimeBudget time.Duration, tt *eval.TranspositionTable) (move board.Move, eval int, depth int, nodes int, elapsed time.Duration) {
	var totalNodes int
	var searchTimeStart time.Time
	var abort bool = false

	moves, checkers := b.GenerateLegalMoves()

	if len(moves) == 0 {
		if checkers > 0 {
			return 0, -100000, 0, 0, 0
		} else {
			return 0, 0, 0, 0, 0
		}
	}

	moves = OrderMoves(b, moves, tt)

	var finalBestEval int = -100000
	var finalBestMove board.Move
	searchTimeStart = time.Now()

	var currentDepth int
	for currentDepth = 1; currentDepth < 20; currentDepth++ {
		if abort {
			break
		}

		var alpha int = -100000
		var beta int = 100000
		var bestMove board.Move

		for _, m := range moves {
			unMove := b.MakeMove(m)
			score := -alphaBetaSearch(b, 0, currentDepth-1, -beta, -alpha, tt, &totalNodes, &searchTimeBudget, &searchTimeStart, &abort)
			b.UnMakeMove(m, unMove)

			if abort {
				break
			}

			if score > alpha {
				alpha = score
				bestMove = m
			}
		}

		if !abort {
			finalBestEval = alpha
			finalBestMove = bestMove
		}
	}

	if finalBestMove == 0 {
		finalBestMove = moves[0]
	}

	return finalBestMove, finalBestEval, currentDepth, totalNodes, time.Duration(time.Since(searchTimeStart).Milliseconds())
}

// alphaBetaSearch recursively searches a position until a given depth, using alpha-beta pruning
func alphaBetaSearch(b *board.Board, ply int, depth int, alpha int, beta int, tt *eval.TranspositionTable, nodes *int, searchTimeBudget *time.Duration, searchTimeStart *time.Time, abort *bool) int {
	if depth <= 0 {
		return quiescenceSearch(b, ply, alpha, beta, nodes)
	}

	if *abort {
		return 0
	}

	ply++

	TTIndex := b.HashKey & tt.Mask
	currTTEntry := tt.Entries[TTIndex]

	if currTTEntry.HashKey == b.HashKey && currTTEntry.Depth >= depth {

		if currTTEntry.Flag == eval.Exact {
			return currTTEntry.Eval
		}
		if currTTEntry.Flag == eval.Alpha && currTTEntry.Eval <= alpha {
			return alpha
		}
		if currTTEntry.Flag == eval.Beta && currTTEntry.Eval >= beta {
			return beta
		}
	}

	moves, checkers := b.GenerateLegalMoves()

	if len(moves) == 0 {
		if checkers > 0 {
			return -100000 + ply
		} else {
			return 0
		}
	}

	moves = OrderMoves(b, moves, tt)

	var bestEval int = -100000
	var bestMove board.Move
	var flag eval.HashFlag = eval.Alpha
	for _, m := range moves {
		*nodes++

		if *nodes%2048 == 0 {
			if time.Since(*searchTimeStart) > *searchTimeBudget {
				*abort = true
				return 0
			}
		}

		unMove := b.MakeMove(m)
		score := -alphaBetaSearch(b, ply, depth-1, -beta, -alpha, tt, nodes, searchTimeBudget, searchTimeStart, abort)
		b.UnMakeMove(m, unMove)

		if *abort {
			return 0
		}

		if score > bestEval {
			bestEval = score
		}

		if score >= beta {
			flag = eval.Beta
			bestEval = score
			bestMove = m
			break
		}

		if score > alpha {
			flag = eval.Exact
			alpha = score
			bestMove = m
		}
	}

	if currTTEntry.HashKey == 0 || currTTEntry.Depth < depth {
		newTTEntry := eval.TTEntry{
			HashKey:  b.HashKey,
			Depth:    depth,
			Eval:     bestEval,
			Flag:     flag,
			BestMove: bestMove,
		}

		tt.Entries[TTIndex] = newTTEntry
	}

	return bestEval
}

// quiescenceSearch recursively searches a position until there are no more captures
func quiescenceSearch(b *board.Board, ply int, alpha int, beta int, nodes *int) int {
	var staticEval int = eval.Evaluate(b)

	bestEval := staticEval
	if bestEval >= beta {
		return bestEval
	}
	if bestEval > alpha {
		alpha = bestEval
	}

	ply++

	moves, checkers := b.GenerateLegalMoves()
	if len(moves) == 0 {
		if checkers > 0 {
			return -100000 + ply
		} else {
			return 0
		}
	}
	captures := GetCaptures(moves)
	for _, m := range captures {
		*nodes++

		unMove := b.MakeMove(m)
		score := -quiescenceSearch(b, ply, -beta, -alpha, nodes)
		b.UnMakeMove(m, unMove)

		if score >= beta {
			return beta
		}

		if score > bestEval {
			bestEval = score
		}

		if score > alpha {
			alpha = score
		}
	}

	return bestEval
}
