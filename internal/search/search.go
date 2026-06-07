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
func RootSearch(b *board.Board, searchTimeBudget time.Duration, tt *eval.TranspositionTable) (move board.Move, eval int, depth int8, nodes int, elapsed time.Duration) {
	var totalNodes int
	var searchTimeStart time.Time
	var abort bool = false

	moves, checkers := b.GenerateLegalMoves()

	if len(moves) == 0 {
		if checkers > 0 {
			return 0, -30000, 0, 0, 0
		} else {
			return 0, 0, 0, 0, 0
		}
	}

	moves = OrderMoves(b, moves, tt)

	var finalBestEval int = -30000
	var finalBestMove board.Move
	searchTimeStart = time.Now()

	var currentDepth int8
	for currentDepth = 1; currentDepth < 20; currentDepth++ {
		if abort {
			break
		}

		var alpha int = -30000
		var beta int = 30000
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
func alphaBetaSearch(b *board.Board, ply int, depth int8, alpha int, beta int, tt *eval.TranspositionTable, nodes *int, searchTimeBudget *time.Duration, searchTimeStart *time.Time, abort *bool) int {
	if depth <= 0 {
		return quiescenceSearch(b, ply, alpha, beta, tt, nodes)
	}

	if *abort {
		return 0
	}

	ply++

	TTIndex := b.HashKey & tt.Mask
	currTTEntry := tt.Entries[TTIndex]

	if currTTEntry.HashKey == b.HashKey && currTTEntry.Depth >= depth {
		currEval := int(currTTEntry.Eval)
		if currTTEntry.Flag == eval.Exact {
			return currEval
		}
		if currTTEntry.Flag == eval.Alpha && currEval <= alpha {
			return alpha
		}
		if currTTEntry.Flag == eval.Beta && currEval >= beta {
			return beta
		}
	}

	moves, checkers := b.GenerateLegalMoves()

	if len(moves) == 0 {
		if checkers > 0 {
			return -30000 + ply
		} else {
			return 0
		}
	}

	if checkFiftyMoveRule(b) || checkRepetition(b) {
		return 0
	}

	moves = OrderMoves(b, moves, tt)

	var bestEval int = -30000
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
			bestMove = m
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

	if currTTEntry.HashKey != b.HashKey || currTTEntry.Depth < depth {

		newTTEntry := eval.TTEntry{
			HashKey:  b.HashKey,
			Depth:    depth,
			Eval:     int16(bestEval),
			Flag:     flag,
			BestMove: bestMove,
		}

		tt.Entries[TTIndex] = newTTEntry
	}

	return bestEval
}

// quiescenceSearch recursively searches a position until there are no more captures
func quiescenceSearch(b *board.Board, ply int, alpha int, beta int, tt *eval.TranspositionTable, nodes *int) int {
	var staticEval int = eval.Evaluate(b)
	var originalAlpha int = alpha

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
			return -30000 + ply
		} else {
			return 0
		}
	}
	sharpMoves := GetAndOrderSharpMoves(b, moves, tt)
	for _, m := range sharpMoves {
		*nodes++

		// delta pruning
		var moveValue int
		if m.IsCapture() {
			moveValue = eval.GetPieceValue(b.GetPiece(m.GetTo()))
		}
		if m.IsPromotion() {
			moveValue += eval.GetPieceValue(board.W_Queen) - eval.GetPieceValue(board.W_Pawn)
		}
		// 200 point buffer for positional considerations
		if originalAlpha > -29000 && staticEval+moveValue+200 < originalAlpha {
			continue
		}

		unMove := b.MakeMove(m)
		score := -quiescenceSearch(b, ply, -beta, -alpha, tt, nodes)
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

// checkFiftyMoveRule checks if the current board position triggers the fifty move rule
func checkFiftyMoveRule(b *board.Board) bool {
	if b.HalfMoveClock >= 100 {
		return true
	}
	return false
}

// checkRepetition checks if the current board position has already happened in the game history
func checkRepetition(b *board.Board) bool {
	if b.HalfMoveClock < 2 {
		return false
	}

	limit := len(b.History) - b.HalfMoveClock
	if limit < 0 {
		limit = 0
	}

	for i := len(b.History) - 2; i >= limit; i-- {
		if b.History[i] == b.HashKey {
			return true
		}
	}
	return false
}
