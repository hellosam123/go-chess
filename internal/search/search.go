// Package search handles all searching and pruning logic.
package search

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/hellosam123/go-chess/internal/board"
	"github.com/hellosam123/go-chess/internal/engine"
	eval "github.com/hellosam123/go-chess/internal/evaluation"
)

const MaxSearchPly = 20
const MaxAbsolutePly = 64

const checkmateThreshold = 29000

type ScoredMove struct {
	move  board.Move
	score int
}

// I've tried experimenting with an OOP approach for the engine
type Search struct {
	Board        *board.Board
	TT           *eval.TranspositionTable
	HistoryTable *[2][64][64]int

	Nodes     int
	SelDepth  int
	StartTime time.Time

	SearchTimeBudget time.Duration
	AnalysisMode     bool
	Aborted          bool

	ScoredMoveBuffer [218]ScoredMove

	KillerMoves [MaxAbsolutePly][2]board.Move

	PVTable  [MaxAbsolutePly][MaxAbsolutePly]board.Move
	PVLength [MaxAbsolutePly]int
}

func NewSearch(e *engine.Engine, searchTimeBudget time.Duration, analysisMode bool) *Search {
	return &Search{
		Board:            e.Board,
		TT:               e.TT,
		HistoryTable:     e.HistoryTable,
		StartTime:        time.Now(),
		SearchTimeBudget: searchTimeBudget,
		AnalysisMode:     analysisMode,
	}
}

func RandomMove(b *board.Board) board.Move {
	moves, _ := b.GenerateLegalMoves()
	moveIndex := rand.IntN(len(moves))
	return moves[moveIndex]
}

// RootSearch uses iterative deepening and initializes the alpha-beta search.
func (s *Search) RootSearch() (move board.Move, score int, depth int8, nodes int, elapsed time.Duration) {
	s.clearSearchHeuristics()

	moves, checkers := s.Board.GenerateLegalMoves()

	if len(moves) == 0 {
		if checkers > 0 {
			return 0, -30000, 0, 0, 0
		} else {
			return 0, 0, 0, 0, 0
		}
	}

	moves = s.OrderMoves(moves, 0)

	var finalBestEval int = -30000
	var finalBestMove board.Move = moves[0]

	var currentDepth int8
	var delta int = 40

	for currentDepth = 1; currentDepth <= MaxSearchPly; currentDepth++ {
		s.SelDepth = 0
		s.PVLength[0] = 0

		if s.Aborted {
			break
		}

		var alpha int = -30001
		var beta int = 30001

		if currentDepth >= 3 {
			alpha = finalBestEval - delta
			beta = finalBestEval + delta

			if s.Board.IsEndgame() {
				s.iterateEndgameHistoryHeuristics()
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
				s.PVLength[1] = 0

				unMove := s.Board.MakeMove(m)
				var score int

				// principle variation search
				if i == 0 {
					score = -s.alphaBetaSearch(0, currentDepth-1, -beta, -alpha, true)
				} else {
					score = -s.alphaBetaSearch(0, currentDepth-1, -alpha-1, -alpha, false)

					if score > alpha && !s.Aborted {
						score = -s.alphaBetaSearch(0, currentDepth-1, -beta, -alpha, true)
					}
				}
				s.Board.UnMakeMove(m, unMove)

				if s.Aborted {
					break
				}

				if score > alpha {
					alpha = score
					bestMove = m

					s.PVTable[0][0] = m
					copy(s.PVTable[0][1:MaxAbsolutePly], s.PVTable[1][1:MaxAbsolutePly])
					s.PVLength[0] = s.PVLength[1] + 1
				}
			}

			if s.Aborted {
				break
			}

			if alpha <= finalBestEval-delta {
				alpha = finalBestEval - delta*2
				beta = finalBestEval + delta
				delta *= 2
				continue
			} else if alpha >= beta {
				// alpha = finalBestEval - delta
				beta = finalBestEval + delta*2
				delta *= 2
				continue
			}

			break
		}

		if !s.Aborted {
			finalBestEval = alpha
			finalBestMove = bestMove
		}

		s.iterateHistoryHeuristics()

		if s.AnalysisMode {
			timeElapsed := time.Duration(time.Since(s.StartTime).Milliseconds())
			if timeElapsed == 0 {
				timeElapsed = 1
			}
			nps := s.Nodes * 1000 / int(timeElapsed)
			pvStr := ""
			for i := 0; i < s.PVLength[0]; i++ {
				pvStr += s.PVTable[0][i].MoveToString() + " "
			}

			fmt.Printf("info depth %d seldepth %d score cp %d nodes %d nps %d time %d pv %s\n", currentDepth, s.SelDepth, finalBestEval, s.Nodes, nps, timeElapsed, pvStr)
		}
	}

	return finalBestMove, finalBestEval, currentDepth, s.Nodes, time.Duration(time.Since(s.StartTime).Milliseconds())
}

// alphaBetaSearch recursively searches a position until a given depth, using alpha-beta pruning
func (s *Search) alphaBetaSearch(ply int, depth int8, alpha int, beta int, isPV bool) int {
	if ply >= MaxAbsolutePly-1 {
		return eval.Evaluate(s.Board)
	}

	if depth <= 0 && s.Board.InCheck() {
		depth++
	}

	if depth <= 0 {
		return s.quiescenceSearch(ply, alpha, beta)
	}

	if s.Aborted {
		return 0
	}

	if s.Board.CheckFiftyMoveRule() || s.Board.CheckRepetition() {
		return 0
	}

	ply++

	TTIndex := s.Board.HashKey & s.TT.Mask
	currTTEntry := s.TT.Entries[TTIndex]

	if currTTEntry.HashKey == s.Board.HashKey && currTTEntry.Depth >= depth {
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

	moves, checkers := s.Board.GenerateLegalMoves()

	// static null move pruning
	if depth >= 3 && checkers == 0 {
		margin := 120 * int(depth)
		if eval.Evaluate(s.Board)-margin >= beta {
			return beta
		}
	}

	// null move pruning
	if depth >= 3 && !isPV && checkers == 0 && s.Board.HasNonPawnPieces(s.Board.ActiveColor) {
		s.Board.ActiveColor = !s.Board.ActiveColor
		savedEnPassantSquare := s.Board.EnPassantSquare
		savedHashKey := s.Board.HashKey

		s.Board.HashKey ^= board.ActiveColorMask
		if s.Board.EnPassantSquare != -1 {
			s.Board.HashKey ^= board.EnPassantTable[s.Board.EnPassantSquare%8]
			s.Board.EnPassantSquare = -1
		}

		var reduction int8 = 2
		nullScore := -s.alphaBetaSearch(ply, depth-1-reduction, -beta, -beta+1, false)
		s.Board.ActiveColor = !s.Board.ActiveColor

		s.Board.EnPassantSquare = savedEnPassantSquare
		s.Board.HashKey = savedHashKey

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

	moves = s.OrderMoves(moves, ply)

	var bestEval int = -30001
	var bestMove board.Move
	var flag eval.HashFlag = eval.Alpha
	for i, m := range moves {
		s.Nodes++

		if !s.AnalysisMode && s.Nodes%8192 == 0 {
			if time.Since(s.StartTime) > s.SearchTimeBudget {
				s.Aborted = true
				return 0
			}
		}

		if ply < MaxAbsolutePly-1 {
			s.PVLength[ply+1] = 0
		}

		var score int
		unMove := s.Board.MakeMove(m)

		// principle variation search
		if i == 0 {
			score = -s.alphaBetaSearch(ply, depth-1, -beta, -alpha, isPV)
		} else {
			// late move reduction
			if depth >= 3 && i >= 3 && !m.IsCapture() && !m.IsPromotion() && checkers == 0 {
				var lmrReduction int8 = 1
				if depth > 4 {
					lmrReduction = 2
				}
				score = -s.alphaBetaSearch(ply, depth-1-lmrReduction, -alpha-1, -alpha, false)
			} else {
				score = alpha + 1
			}

			// if lmr didn't run or lmr score above alpha
			if score > alpha {
				score = -s.alphaBetaSearch(ply, depth-1, -alpha-1, -alpha, false)
			}

			if score > alpha && score < beta && !s.Aborted {
				// full width search
				score = -s.alphaBetaSearch(ply, depth-1, -beta, -alpha, true)
			}
		}
		s.Board.UnMakeMove(m, unMove)

		if s.Aborted {
			return 0
		}

		if score > bestEval {
			bestEval = score
			bestMove = m
		}

		if score >= beta {
			if !m.IsCapture() {
				if s.KillerMoves[ply][0] != m {
					s.KillerMoves[ply][1] = s.KillerMoves[ply][0]
					s.KillerMoves[ply][0] = m
				}

				side := 0
				if !s.Board.ActiveColor {
					side = 1
				}

				// history bonus
				maxHistory := 20000
				bonus := int(depth) * int(depth)
				currentScore := s.HistoryTable[side][m.GetFrom()][m.GetTo()]
				s.HistoryTable[side][m.GetFrom()][m.GetTo()] += bonus - (bonus * currentScore / maxHistory)
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

			s.PVTable[ply][ply] = m
			if ply < MaxAbsolutePly-1 {
				nextPly := ply + 1
				copy(s.PVTable[ply][nextPly:MaxAbsolutePly], s.PVTable[nextPly][nextPly:MaxAbsolutePly])
				s.PVLength[ply] = s.PVLength[nextPly] + 1
			} else {
				s.PVLength[ply] = 1
			}

		} else {
			// history malus
			if !m.IsCapture() {
				side := 0
				if !s.Board.ActiveColor {
					side = 1
				}

				s.HistoryTable[side][m.GetFrom()][m.GetTo()] -= int(depth) * int(depth)
				if s.HistoryTable[side][m.GetFrom()][m.GetTo()] < -20000 {
					s.HistoryTable[side][m.GetFrom()][m.GetTo()] = -20000
				}
			}
		}
	}

	if currTTEntry.HashKey != s.Board.HashKey || currTTEntry.Depth < depth {

		newTTEntry := eval.TTEntry{
			HashKey:  s.Board.HashKey,
			Depth:    depth,
			Eval:     scoreToTT(bestEval, ply),
			Flag:     flag,
			BestMove: bestMove,
		}

		s.TT.Entries[TTIndex] = newTTEntry
	}

	return bestEval
}

// quiescenceSearch recursively searches a position until there are no more captures
func (s *Search) quiescenceSearch(ply int, alpha int, beta int) int {
	var staticEval int = eval.Evaluate(s.Board)
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
	if ply > s.SelDepth {
		s.SelDepth = ply
	}

	moves, checkers := s.Board.GenerateLegalMoves()
	if len(moves) == 0 {
		if checkers > 0 {
			return -30000 + ply
		} else {
			return 0
		}
	}
	sharpMoves := s.GetAndOrderSharpMoves(moves, ply)
	for _, m := range sharpMoves {
		s.Nodes++

		// delta pruning
		var moveValue int
		if m.IsCapture() {
			// seems like SEE pruning loses elo here
			// SEE pruning
			// if eval.StaticExchangeEval(b, m.GetTo(), b.ActiveColor) < 0 {
			// 	continue
			// }

			moveValue = eval.GetPieceValue(s.Board.GetPiece(m.GetTo()))
		}
		if m.IsPromotion() {
			moveValue += eval.GetPieceValue(board.W_Queen) - eval.GetPieceValue(board.W_Pawn)
		}
		// 200 point buffer for positional considerations
		if originalAlpha > -29000 && staticEval+moveValue+200 < originalAlpha {
			continue
		}

		unMove := s.Board.MakeMove(m)
		score := -s.quiescenceSearch(ply, -beta, -alpha)
		s.Board.UnMakeMove(m, unMove)

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

// clearSearchHeuristics clears KillerMoves and divides all values in HistoryTable by 2.
func (s *Search) clearSearchHeuristics() {
	s.KillerMoves = [MaxAbsolutePly][2]board.Move{}

	for side := 0; side < 2; side++ {
		for from := 0; from < 64; from++ {
			for to := 0; to < 64; to++ {
				// age old search data
				s.HistoryTable[side][from][to] /= 2
			}
		}
	}
}

// iterateHistoryHeuristics multiplies all values in HistoryTable by 4/5.
func (s *Search) iterateHistoryHeuristics() {
	for side := 0; side < 2; side++ {
		for from := 0; from < 64; from++ {
			for to := 0; to < 64; to++ {
				// age old search data
				s.HistoryTable[side][from][to] = s.HistoryTable[side][from][to] * 4 / 5
			}
		}
	}
}

// iterateEndgameHistoryHeuristics multiples all values except for king moves by 1/2.
func (s *Search) iterateEndgameHistoryHeuristics() {
	for side := 0; side < 2; side++ {
		for from := 0; from < 64; from++ {
			fromPiece := s.Board.GetPiece(from)
			if fromPiece == board.W_King || fromPiece == board.B_King {
				continue
			}
			for to := 0; to < 64; to++ {
				s.HistoryTable[side][from][to] = s.HistoryTable[side][from][to] * 1 / 2
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
