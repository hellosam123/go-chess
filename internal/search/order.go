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
func OrderMoves(b *board.Board, moves []board.Move, ply int, tt *eval.TranspositionTable) []board.Move {
	var scoredMoves []scoredMove = make([]scoredMove, len(moves))
	var sortedMoves []board.Move = make([]board.Move, len(moves))

	var ttBestMove board.Move

	ttIndex := b.HashKey & tt.Mask
	ttEntry := tt.Entries[ttIndex]

	if ttEntry.HashKey == b.HashKey {
		ttBestMove = ttEntry.BestMove
	}

	for i, m := range moves {
		if m == ttBestMove {
			scoredMoves[i] = scoredMove{move: m, score: 100000}
			continue
		}

		var moveScoreGuess int = 0
		var fromMask uint64 = 1 << m.GetFrom()
		var toMask uint64 = 1 << m.GetTo()

		var movePiece board.Piece
		for p, pMask := range b.Pieces {
			if fromMask&pMask != 0 {
				movePiece = board.Piece(p)
				break
			}
		}

		if m.IsPromotion() {
			moveScoreGuess = 50000
		}

		if m.IsCapture() {
			for p, pMask := range b.Pieces {
				if toMask&pMask != 0 {
					capturedPiece := board.Piece(p)
					// adds score on top of promotion value
					moveScoreGuess += 30000 + (eval.GetPieceValue(capturedPiece)*2 - eval.GetPieceValue(movePiece))
					break
				}
			}
		} else {
			switch m {
			case KillerMoves[ply][0]:
				moveScoreGuess = 29000
			case KillerMoves[ply][1]:
				moveScoreGuess = 28000
			default:
				if eval.IsEndgame(b) && (movePiece == board.W_Pawn || movePiece == board.B_Pawn) && eval.IsPassedPawn(b, m.GetFrom(), b.ActiveColor) {
					targetRank := m.GetTo() / 8
					if !b.ActiveColor {
						targetRank = 7 - targetRank
					}

					moveScoreGuess = 25000 + targetRank*100
				} else {
					side := 0
					if !b.ActiveColor {
						side = 1
					}
					moveScoreGuess = HistoryTable[side][m.GetFrom()][m.GetTo()]
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
func GetAndOrderSharpMoves(b *board.Board, moves []board.Move, ply int, tt *eval.TranspositionTable) []board.Move {
	var sharpMoves []board.Move = make([]board.Move, 0, len(moves))
	for _, m := range moves {
		if m.IsCapture() || m.IsPromotion() {
			sharpMoves = append(sharpMoves, m)
		}
	}
	return OrderMoves(b, sharpMoves, ply, tt)
}
