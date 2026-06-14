package search

import (
	"sort"

	"github.com/hellosam123/go-chess/internal/board"
	eval "github.com/hellosam123/go-chess/internal/evaluation"
)

type scoredMove struct {
	move  board.Move
	score int
}

// OrderMoves orders a list of moves by making educated guesses at their scores
func (s *Search) OrderMoves(moves []board.Move, ply int) []board.Move {
	var scoredMoves []scoredMove = make([]scoredMove, len(moves))
	var sortedMoves []board.Move = make([]board.Move, len(moves))

	var ttBestMove board.Move

	ttIndex := s.Board.HashKey & s.TT.Mask
	ttEntry := s.TT.Entries[ttIndex]

	if ttEntry.HashKey == s.Board.HashKey {
		ttBestMove = ttEntry.BestMove
	}

	for i, m := range moves {
		if m == ttBestMove {
			scoredMoves[i] = scoredMove{move: m, score: 100000}
			continue
		}

		var moveScoreGuess int = 0

		movePiece := s.Board.GetPiece(m.GetFrom())

		if m.IsPromotion() {
			moveScoreGuess = 50000
		}

		if m.IsCapture() {
			captureEval := eval.StaticExchangeEval(s.Board, m.GetTo(), s.Board.ActiveColor)
			if captureEval >= 0 {
				moveScoreGuess = 30000 + captureEval
			} else {
				moveScoreGuess = -10000 + captureEval
			}
		} else {
			switch m {
			case s.KillerMoves[ply][0]:
				moveScoreGuess = 29000
			case s.KillerMoves[ply][1]:
				moveScoreGuess = 28000
			default:
				if s.Board.IsEndgame() && (movePiece == board.W_Pawn || movePiece == board.B_Pawn) && eval.IsPassedPawn(s.Board, m.GetFrom(), s.Board.ActiveColor) {
					targetRank := m.GetTo() / 8
					if !s.Board.ActiveColor {
						targetRank = 7 - targetRank
					}

					moveScoreGuess = 25000 + targetRank*100
				} else {
					side := 0
					if !s.Board.ActiveColor {
						side = 1
					}
					moveScoreGuess = s.HistoryTable[side][m.GetFrom()][m.GetTo()]
				}
			}
		}

		scoredMoves[i] = scoredMove{move: m, score: moveScoreGuess}
	}

	sort.Slice(scoredMoves, func(i int, j int) bool {
		return scoredMoves[i].score > scoredMoves[j].score
	})

	for i, sm := range scoredMoves {
		sortedMoves[i] = sm.move
	}

	return sortedMoves
}

// GetAndOrderSharpMoves takes a list of moves and filters for sharp moves and orders them
func (s *Search) GetAndOrderSharpMoves(moves []board.Move, ply int) []board.Move {
	var sharpMoves []board.Move = make([]board.Move, 0, len(moves))
	for _, m := range moves {
		if m.IsCapture() || m.IsPromotion() {
			sharpMoves = append(sharpMoves, m)
		}
	}
	return s.OrderMoves(sharpMoves, ply)
}
