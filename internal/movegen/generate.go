package movegen

import (
	"math/bits"

	"github.com/hellosam123/go-chess/internal/board"
)

var knightAttacks [64]uint64
var kingAttacks [64]uint64

func init() {
	precalculateKnightAttacks()
	precalculateKingAttacks()
}

// precalculateKnightAttacks calculates all knight attacks for each
// square and saves it in knightAttacks as an array of bitboards
func precalculateKnightAttacks() {
	// offsets in squares clockwise from North
	rankOffsets := []int{2, 1, -1, -2, -2, -1, 1, 2}
	fileOffsets := []int{1, 2, 2, 1, -1, -2, -2, -1}

	for sq := 0; sq < 64; sq++ {
		rank := sq / 8
		file := sq % 8

		var attacks uint64
		for i := 0; i < 8; i++ {
			attackRank := rank + rankOffsets[i]
			attackFile := file + fileOffsets[i]

			if attackRank >= 0 && attackRank < 8 && attackFile >= 0 && attackFile < 8 {
				attackSq := attackRank*8 + attackFile
				attacks |= 1 << attackSq
			}
		}
		knightAttacks[sq] = attacks
	}
}

// precalculateKingAttacks calculates all king attacks for each
// square and saves it in kingAttacks as an array of bitboards
func precalculateKingAttacks() {
	// offsets in squares clockwise from North
	rankOffsets := []int{1, 1, 0, -1, -1, -1, 0, 1}
	fileOffsets := []int{0, 1, 1, 1, 0, -1, -1, -1}

	for sq := 0; sq < 64; sq++ {
		rank := sq / 8
		file := sq % 8

		var attacks uint64
		for i := 0; i < 8; i++ {
			attackRank := rank + rankOffsets[i]
			attackFile := file + fileOffsets[i]

			if attackRank >= 0 && attackRank < 8 && attackFile >= 0 && attackFile < 8 {
				attackSq := attackRank*8 + attackFile
				attacks |= 1 << attackSq
			}
		}
		kingAttacks[sq] = attacks
	}
}

// GeneratePseudoLegalMoves generates all pseudo legal moves in a board
func GeneratePseudoLegalMoves(b *board.Board) {
	moves := make([]Move, 0, 48)

	generateKnightMoves(b, &moves)
	generateKingMoves(b, &moves)

}

// generatePawnMoves generates all pseudo legal pawn moves in a board
// func generatePawnMoves(b *board.Board, moves *[]Move) {
// 	// if White's turn
// 	var pawns uint64
// 	var us uint64
// 	var them uint64
// 	activeColor := b.ActiveColor
// 	if activeColor {
// 		pawns = b.Pieces[board.W_Pawn]
// 		us = b.WPieces
// 		them = b.BPieces

// 		toMask := pawns << 8

// 	} else {
// 		pawns = b.Pieces[board.B_Pawn]
// 		us = b.BPieces
// 		them = b.WPieces

// 		toMask := pawns >> 8
// 	}
// }

// generateKnightMoves generates all pseudo legal knight moves in a board
func generateKnightMoves(b *board.Board, moves *[]Move) {
	// if White's turn
	var knights uint64
	var us uint64
	var them uint64
	if b.ActiveColor {
		knights = b.Pieces[board.W_Knight]
		us = b.WPieces
		them = b.BPieces
	} else {
		knights = b.Pieces[board.B_Knight]
		us = b.BPieces
		them = b.WPieces
	}

	for knights != 0 {
		from := bits.TrailingZeros64(knights)
		// Brian Kernighan's flip rightmost 1 method (e.g. 1001 1100 -> 1001 1000)
		knights &= knights - 1

		attacks := knightAttacks[from]
		attacks &^= us
		captures := attacks & them

		for attacks != 0 {
			// magic bit manipulation hack to isolate only rightmost bit
			// (e.g. 1001 1100 -> 0000 0100)
			toMask := attacks & -attacks

			to := bits.TrailingZeros64(attacks)
			attacks &= attacks - 1

			var flag Flag = QuietMove
			if (toMask)&captures != 0 {
				flag = Capture
			}

			*moves = append(*moves, New(from, to, flag))
		}
	}
}

// generateKingMoves generates all pseudo legal king moves in a board
func generateKingMoves(b *board.Board, moves *[]Move) {
	// if White's turn
	var king uint64
	var us uint64
	var them uint64
	if b.ActiveColor {
		king = b.Pieces[board.W_King]
		us = b.WPieces
		them = b.BPieces
	} else {
		king = b.Pieces[board.B_King]
		us = b.BPieces
		them = b.WPieces
	}

	from := bits.TrailingZeros64(king)

	attacks := kingAttacks[from]
	attacks &^= us
	captures := attacks & them

	for attacks != 0 {
		toMask := attacks & -attacks

		to := bits.TrailingZeros64(attacks)
		attacks &= attacks - 1

		var flag Flag = QuietMove
		if (toMask)&captures != 0 {
			flag = Capture
		}

		*moves = append(*moves, New(from, to, flag))
	}
}

// func generateBishopMoves(b *board.Board, moves *[]Move) {
// 	var bishops uint64
// 	var us uint64
// 	var them uint64
// 	if b.ActiveColor {
// 		bishops = b.Pieces[board.W_King]
// 		us = b.WPieces
// 		them = b.BPieces
// 	} else {
// 		bishops = b.Pieces[board.B_King]
// 		us = b.BPieces
// 		them = b.WPieces
// 	}

// 	for bishops != 0 {
// 		from := bits.TrailingZeros64(bishops)
// 		bishops &= bishops - 1

// 	}
// }

// func generateRookMoves(b *board.Board, moves *[]Move) {
// 	var rooks uint64
// 	var us uint64
// 	var them uint64
// 	if b.ActiveColor {
// 		rooks = b.Pieces[board.W_King]
// 		us = b.WPieces
// 		them = b.BPieces
// 	} else {
// 		rooks = b.Pieces[board.B_King]
// 		us = b.BPieces
// 		them = b.WPieces
// 	}

// 	for rooks != 0 {
// 		from := bits.TrailingZeros64(rooks)
// 		rooks &= rooks - 1

// 	}
// }
