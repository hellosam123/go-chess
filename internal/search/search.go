// Package search handles all searching and pruning logic.
package search

import (
	"math/rand/v2"
	"time"

	"github.com/hellosam123/go-chess/internal/board"
	eval "github.com/hellosam123/go-chess/internal/evaluation"
)

const MaxSearchPly = 20
const MaxAbsolutePly = 64

const checkmateThreshold = 29000

// KillerMoves stores the ply and 2 killer moves
var KillerMoves [MaxAbsolutePly][2]board.Move

// HistoryTable stores the number of times a certain move triggers beta
var HistoryTable [2][64][64]int

func RandomMove(b *board.Board) board.Move {
	moves, _ := b.GenerateLegalMoves()
	moveIndex := rand.IntN(len(moves))
	return moves[moveIndex]
}

// RootSearch uses iterative deepening and initializes the alpha-beta search.
func RootSearch(b *board.Board, searchTimeBudget time.Duration, tt *eval.TranspositionTable) (move board.Move, score int, depth int8, nodes int, elapsed time.Duration) {
	var totalNodes int
	var searchTimeStart time.Time
	var abort bool = false

	clearSearchHeuristics()

	moves, checkers := b.GenerateLegalMoves()

	if len(moves) == 0 {
		if checkers > 0 {
			return 0, -30000, 0, 0, 0
		} else {
			return 0, 0, 0, 0, 0
		}
	}

	moves = OrderMoves(b, moves, 0, tt)

	var finalBestEval int = -30000
	var finalBestMove board.Move = moves[0]
	searchTimeStart = time.Now()

	var currentDepth int8
	var delta int = 40

	for currentDepth = 1; currentDepth <= MaxSearchPly; currentDepth++ {
		if abort {
			break
		}

		var alpha int = -30001
		var beta int = 30001

		if currentDepth >= 3 {
			alpha = finalBestEval - delta
			beta = finalBestEval + delta

			if eval.IsEndgame(b) {
				iterateEndgameHistoryHeuristics(b)
			}
		}

		var bestMove board.Move

		// aspiration windows
		for {
			if alpha < -30001 {
				alpha = -30001
			}
			if beta > 30001 {
				beta = 30001
			}

			for i, m := range moves {
				unMove := b.MakeMove(m)

				var score int

				// principle variation search
				if i == 0 {
					score = -alphaBetaSearch(b, 0, currentDepth-1, -beta, -alpha, true, tt, &totalNodes, &searchTimeBudget, &searchTimeStart, &abort)
				} else {
					score = -alphaBetaSearch(b, 0, currentDepth-1, -alpha-1, -alpha, false, tt, &totalNodes, &searchTimeBudget, &searchTimeStart, &abort)

					if score > alpha && !abort {
						score = -alphaBetaSearch(b, 0, currentDepth-1, -beta, -alpha, true, tt, &totalNodes, &searchTimeBudget, &searchTimeStart, &abort)
					}
				}
				b.UnMakeMove(m, unMove)

				if abort {
					break
				}

				if score > alpha {
					alpha = score
					bestMove = m
				}
			}

			if abort {
				break
			}

			if alpha <= finalBestEval-delta {
				alpha = finalBestEval - delta*2
				beta = finalBestEval + delta
				delta *= 2
				continue
			} else if alpha >= beta {
				beta = finalBestEval + delta*2
				delta *= 2
				continue
			}

			break
		}

		if !abort {
			finalBestEval = alpha
			finalBestMove = bestMove
		}

		iterateHistoryHeuristics()
	}

	return finalBestMove, finalBestEval, currentDepth, totalNodes, time.Duration(time.Since(searchTimeStart).Milliseconds())
}

// alphaBetaSearch recursively searches a position until a given depth, using alpha-beta pruning
func alphaBetaSearch(b *board.Board, ply int, depth int8, alpha int, beta int, isPV bool, tt *eval.TranspositionTable, nodes *int, searchTimeBudget *time.Duration, searchTimeStart *time.Time, abort *bool) int {
	if ply >= MaxAbsolutePly-1 {
		return eval.Evaluate(b)
	}

	if depth <= 0 {
		return quiescenceSearch(b, ply, alpha, beta, tt, nodes)
	}

	if *abort {
		return 0
	}

	if checkFiftyMoveRule(b) || checkRepetition(b) {
		return 0
	}

	ply++

	TTIndex := b.HashKey & tt.Mask
	currTTEntry := tt.Entries[TTIndex]

	if currTTEntry.HashKey == b.HashKey && currTTEntry.Depth >= depth {
		currEval := scoreFromTT(currTTEntry.Eval, ply)
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

	// static null move pruning
	if depth >= 3 && checkers == 0 {
		margin := 120 * int(depth)
		if eval.Evaluate(b)-margin >= beta {
			return beta
		}
	}

	// null move pruning
	if depth >= 3 && !isPV && checkers == 0 && hasNonPawnPieces(b, b.ActiveColor) {
		b.ActiveColor = !b.ActiveColor
		savedEnPassantSquare := b.EnPassantSquare
		savedHashKey := b.HashKey

		b.HashKey ^= board.ActiveColorMask
		if b.EnPassantSquare != -1 {
			b.HashKey ^= board.EnPassantTable[b.EnPassantSquare%8]
			b.EnPassantSquare = -1
		}

		var reduction int8 = 2
		nullScore := -alphaBetaSearch(b, ply, depth-1-reduction, -beta, -beta+1, false, tt, nodes, searchTimeBudget, searchTimeStart, abort)
		b.ActiveColor = !b.ActiveColor

		b.EnPassantSquare = savedEnPassantSquare
		b.HashKey = savedHashKey

		if nullScore >= beta {
			return beta
		}
	}

	if len(moves) == 0 {
		if checkers > 0 {
			return -30000 + ply
		} else {
			return 0
		}
	}

	moves = OrderMoves(b, moves, ply, tt)

	var bestEval int = -30001
	var bestMove board.Move
	var flag eval.HashFlag = eval.Alpha
	for i, m := range moves {
		*nodes++

		if *nodes%2048 == 0 {
			if time.Since(*searchTimeStart) > *searchTimeBudget {
				*abort = true
				return 0
			}
		}

		var score int
		unMove := b.MakeMove(m)

		// principle variation search
		if i == 0 {
			score = -alphaBetaSearch(b, ply, depth-1, -beta, -alpha, isPV, tt, nodes, searchTimeBudget, searchTimeStart, abort)
		} else {
			// late move reduction
			if depth >= 3 && i >= 3 && !m.IsCapture() && !m.IsPromotion() && checkers == 0 {
				var lmrReduction int8 = 1
				if depth > 4 {
					lmrReduction = 2
				}
				score = -alphaBetaSearch(b, ply, depth-1-lmrReduction, -alpha-1, -alpha, false, tt, nodes, searchTimeBudget, searchTimeStart, abort)
			} else {
				score = alpha + 1
			}

			// if lmr didn't run or lmr score above alpha
			if score > alpha {
				score = -alphaBetaSearch(b, ply, depth-1, -alpha-1, -alpha, false, tt, nodes, searchTimeBudget, searchTimeStart, abort)
			}

			if score > alpha && score < beta && !*abort {
				// full width search
				score = -alphaBetaSearch(b, ply, depth-1, -beta, -alpha, true, tt, nodes, searchTimeBudget, searchTimeStart, abort)
			}
		}
		b.UnMakeMove(m, unMove)

		if *abort {
			return 0
		}

		if score > bestEval {
			bestEval = score
			bestMove = m
		}

		if score >= beta {
			if !m.IsCapture() {
				if KillerMoves[ply][0] != m {
					KillerMoves[ply][1] = KillerMoves[ply][0]
					KillerMoves[ply][0] = m
				}

				side := 0
				if !b.ActiveColor {
					side = 1
				}

				// history bonus
				maxHistory := 20000
				bonus := int(depth) * int(depth)
				currentScore := HistoryTable[side][m.GetFrom()][m.GetTo()]
				HistoryTable[side][m.GetFrom()][m.GetTo()] += bonus - (bonus * currentScore / maxHistory)
			}

			flag = eval.Beta
			bestEval = score
			bestMove = m
			break
		}

		if score > alpha {
			flag = eval.Exact
			alpha = score
			bestMove = m
		} else {
			// history malus
			if !m.IsCapture() {
				side := 0
				if !b.ActiveColor {
					side = 1
				}

				if HistoryTable[side][m.GetFrom()][m.GetTo()] > -20000 {
					HistoryTable[side][m.GetFrom()][m.GetTo()] -= int(depth) * int(depth)
				}
			}
		}
	}

	if currTTEntry.HashKey != b.HashKey || currTTEntry.Depth < depth {

		newTTEntry := eval.TTEntry{
			HashKey:  b.HashKey,
			Depth:    depth,
			Eval:     scoreToTT(bestEval, ply),
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

	if ply >= MaxAbsolutePly-1 {
		return staticEval
	}

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
	sharpMoves := GetAndOrderSharpMoves(b, moves, ply, tt)
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

// hasNonPawnPieces checks if a side has non pawn pieces. Used for zugzwang detection.
func hasNonPawnPieces(b *board.Board, color bool) bool {
	if color {
		return (b.Pieces[board.W_Pawn] | b.Pieces[board.W_King]) != b.AllPieces
	}
	return (b.Pieces[board.B_Pawn] | b.Pieces[board.B_King]) != b.AllPieces
}

// clearSearchHeuristics clears KillerMoves and divides all values in HistoryTable by 2.
func clearSearchHeuristics() {
	KillerMoves = [MaxAbsolutePly][2]board.Move{}

	for side := 0; side < 2; side++ {
		for from := 0; from < 64; from++ {
			for to := 0; to < 64; to++ {
				// age old search data
				HistoryTable[side][from][to] /= 2
			}
		}
	}
}

// iterateHistoryHeuristics multiplies all values in HistoryTable by 4/5.
func iterateHistoryHeuristics() {
	for side := 0; side < 2; side++ {
		for from := 0; from < 64; from++ {
			for to := 0; to < 64; to++ {
				// age old search data
				HistoryTable[side][from][to] = HistoryTable[side][from][to] * 4 / 5
			}
		}
	}
}

// iterateEndgameHistoryHeuristics multiples all values except for king moves by 1/2.
func iterateEndgameHistoryHeuristics(b *board.Board) {
	for side := 0; side < 2; side++ {
		for from := 0; from < 64; from++ {
			for to := 0; to < 64; to++ {

				if b.GetPiece(from) != board.W_King || b.GetPiece(from) != board.B_King {
					HistoryTable[side][from][to] = HistoryTable[side][from][to] * 1 / 2
				}
			}
		}
	}
}

func scoreToTT(score int, ply int) int16 {
	if score > checkmateThreshold {
		return int16(score + ply)
	}
	if score < -checkmateThreshold {
		return int16(score - ply)
	}
	return int16(score)
}

func scoreFromTT(score int16, ply int) int {
	if score > checkmateThreshold {
		return int(score) - ply
	}
	if score < -checkmateThreshold {
		return int(score) + ply
	}
	return int(score)
}
