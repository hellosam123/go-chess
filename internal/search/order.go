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
func OrderMoves(b *board.Board, moves []board.Move, tt *eval.TranspositionTable) []board.Move {
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

		var capturedPiece board.Piece = board.Empty
		var isCapture bool = false

		if toMask&b.AllPieces != 0 {
			isCapture = true
			for p, pMask := range b.Pieces {
				if toMask&pMask != 0 {
					capturedPiece = board.Piece(p)
					break
				}
			}
		}

		if isCapture {
			moveScoreGuess = 10000 + (eval.GetPieceValue(capturedPiece)*10 - eval.GetPieceValue(movePiece))
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

// GetCaptures takes a list of moves and filters for captures
func GetCaptures(moves []board.Move) []board.Move {
	var captures []board.Move = make([]board.Move, 0, len(moves))
	for _, m := range moves {
		if m.IsCapture() {
			captures = append(captures, m)
		}
	}

	return captures
}
