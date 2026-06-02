package movegen

import (
	"math/bits"

	"github.com/hellosam123/go-chess/internal/board"
)

var knightAttacks [64]uint64
var kingAttacks [64]uint64
var RookDirections = [4]int{8, 1, -8, -1}
var BishopDirections = [4]int{9, -7, -9, 7}

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
	moves := make([]Move, 0, 64)

	generatePawnMoves(b, &moves)
	generateKnightMoves(b, &moves)
	generateKingMoves(b, &moves)
	generateBishopMoves(b, &moves)
	generateRookMoves(b, &moves)
	generateQueenMoves(b, &moves)

}

// generatePawnMoves generates all pseudo legal pawn moves in a board
func generatePawnMoves(b *board.Board, moves *[]Move) {
	var pawns uint64
	var them uint64
	var rankStart uint64
	var rankEnd uint64
	var fileA uint64 = 0x101010101010101
	var fileH uint64 = 0x8080808080808080

	var shiftFunc func(uint64, int) uint64

	if b.ActiveColor {
		pawns = b.Pieces[board.W_Pawn]
		them = b.BPieces
		rankStart = 0xFF00
		rankEnd = 0xFF00000000000000
		shiftFunc = func(bb uint64, shift int) uint64 { return bb << shift }
	} else {
		pawns = b.Pieces[board.B_Pawn]
		them = b.WPieces
		rankStart = 0xFF000000000000
		rankEnd = 0xFF
		shiftFunc = func(bb uint64, shift int) uint64 { return bb >> shift }
	}

	// a bitboard of all squares without pieces
	empty := ^b.AllPieces

	// pushMask is a bitboard of regular pawn pushes
	pushMask := (shiftFunc(pawns, 8)) & empty

	doublePushMask := (shiftFunc(pushMask, 8)) & empty & (shiftFunc(pawns&rankStart, 16))

	var captureLeftMask uint64
	var captureRightMask uint64
	if b.ActiveColor {
		captureLeftMask = (shiftFunc(pawns&^fileA, 7)) & them
		captureRightMask = (shiftFunc(pawns&^fileH, 9)) & them
	} else {
		captureLeftMask = (shiftFunc(pawns&^fileA, 9)) & them
		captureRightMask = (shiftFunc(pawns&^fileH, 7)) & them
	}

	// en passant mask
	var epLeftMask uint64
	var epRightMask uint64

	if b.EnPassantSquare != -1 {
		var epBB uint64 = 1 << b.EnPassantSquare
		if b.ActiveColor {
			epLeftMask = ((pawns &^ fileA) << 7) & epBB
			epRightMask = ((pawns &^ fileH) << 9) & epBB
		} else {
			epLeftMask = ((pawns &^ fileA) >> 9) & epBB
			epRightMask = ((pawns &^ fileH) >> 7) & epBB
		}
	}

	for pushMask != 0 {
		to := bits.TrailingZeros64(pushMask)
		pushMask &= pushMask - 1

		var from int
		if b.ActiveColor {
			from = to - 8
		} else {
			from = to + 8
		}

		if (1<<to)&rankEnd != 0 {
			*moves = append(*moves, New(from, to, PromoteN))
			*moves = append(*moves, New(from, to, PromoteB))
			*moves = append(*moves, New(from, to, PromoteR))
			*moves = append(*moves, New(from, to, PromoteQ))
		} else {
			*moves = append(*moves, New(from, to, QuietMove))
		}
	}

	for doublePushMask != 0 {
		to := bits.TrailingZeros64(doublePushMask)
		doublePushMask &= doublePushMask - 1

		var from int
		if b.ActiveColor {
			from = to - 16
		} else {
			from = to + 16
		}
		*moves = append(*moves, New(from, to, DoublePawnPush))
	}

	for captureLeftMask != 0 {
		to := bits.TrailingZeros64(captureLeftMask)
		captureLeftMask &= captureLeftMask - 1

		var from int
		if b.ActiveColor {
			from = to - 7
		} else {
			from = to + 9
		}

		if (1<<to)&rankEnd != 0 {
			*moves = append(*moves, New(from, to, PromoteCaptureN))
			*moves = append(*moves, New(from, to, PromoteCaptureB))
			*moves = append(*moves, New(from, to, PromoteCaptureR))
			*moves = append(*moves, New(from, to, PromoteCaptureQ))
		} else {
			*moves = append(*moves, New(from, to, Capture))
		}
	}

	for captureRightMask != 0 {
		to := bits.TrailingZeros64(captureRightMask)
		captureRightMask &= captureRightMask - 1

		var from int
		if b.ActiveColor {
			from = to - 9
		} else {
			from = to + 7
		}

		if (1<<to)&rankEnd != 0 {
			*moves = append(*moves, New(from, to, PromoteCaptureN))
			*moves = append(*moves, New(from, to, PromoteCaptureB))
			*moves = append(*moves, New(from, to, PromoteCaptureR))
			*moves = append(*moves, New(from, to, PromoteCaptureQ))
		} else {
			*moves = append(*moves, New(from, to, Capture))
		}
	}

	for epLeftMask != 0 {
		to := bits.TrailingZeros64(epLeftMask)
		epLeftMask = 0

		var from int
		if b.ActiveColor {
			from = to - 7
		} else {
			from = to + 9
		}
		*moves = append(*moves, New(from, to, EnPassant))
	}

	for epRightMask != 0 {
		to := bits.TrailingZeros64(epRightMask)
		epRightMask = 0

		var from int
		if b.ActiveColor {
			from = to - 9
		} else {
			from = to + 7
		}
		*moves = append(*moves, New(from, to, EnPassant))
	}
}

// generateKnightMoves generates all pseudo legal knight moves in a board
func generateKnightMoves(b *board.Board, moves *[]Move) {
	var knights uint64
	var us uint64
	var them uint64
	// if White's turn
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

// generateDirectionalRay creates a ray by taking an initial position and
// repeatedly incrementing by a step until a collision, saving to moves
func generateDirectionalRay(from int, step int, us uint64, them uint64, moves *[]Move) {
	currentSq := from
	for {
		currentFile := currentSq % 8

		currentSq += step

		if currentSq < 0 || currentSq > 63 {
			break
		}

		newFile := currentSq % 8
		if step == 1 || step == 9 || step == -7 {
			if newFile <= currentFile {
				break
			}
		}

		if step == -1 || step == 7 || step == -9 {
			if newFile >= currentFile {
				break
			}
		}

		var sqMask uint64 = 1 << currentSq
		if sqMask&us != 0 {
			break
		}
		if sqMask&them != 0 {
			*moves = append(*moves, New(from, currentSq, Capture))
			break
		}

		*moves = append(*moves, New(from, currentSq, QuietMove))
	}
}

// generateBishopMoves generates all pseudo legal bishop moves in a board
func generateBishopMoves(b *board.Board, moves *[]Move) {
	var bishops uint64
	var us uint64
	var them uint64
	if b.ActiveColor {
		bishops = b.Pieces[board.W_Bishop]
		us = b.WPieces
		them = b.BPieces
	} else {
		bishops = b.Pieces[board.B_Bishop]
		us = b.BPieces
		them = b.WPieces
	}

	for bishops != 0 {
		from := bits.TrailingZeros64(bishops)
		bishops &= bishops - 1

		// cast a ray in all 4 directions
		for _, step := range BishopDirections {
			generateDirectionalRay(from, step, us, them, moves)
		}

	}
}

// generateRookMoves generates all pseudo legal rook moves in a board
func generateRookMoves(b *board.Board, moves *[]Move) {
	var rooks uint64
	var us uint64
	var them uint64
	if b.ActiveColor {
		rooks = b.Pieces[board.W_Rook]
		us = b.WPieces
		them = b.BPieces
	} else {
		rooks = b.Pieces[board.B_Rook]
		us = b.BPieces
		them = b.WPieces
	}

	for rooks != 0 {
		from := bits.TrailingZeros64(rooks)
		rooks &= rooks - 1

		// cast a ray in all 4 directions
		for _, step := range RookDirections {
			generateDirectionalRay(from, step, us, them, moves)
		}

	}
}

// generateQueenMoves generates all pseudo legal queen moves in a board
func generateQueenMoves(b *board.Board, moves *[]Move) {
	var queens uint64
	var us uint64
	var them uint64
	if b.ActiveColor {
		queens = b.Pieces[board.W_Queen]
		us = b.WPieces
		them = b.BPieces
	} else {
		queens = b.Pieces[board.B_Queen]
		us = b.BPieces
		them = b.WPieces
	}

	for queens != 0 {
		from := bits.TrailingZeros64(queens)
		queens &= queens - 1

		// cast a ray in all 4 directions
		for _, step := range RookDirections {
			generateDirectionalRay(from, step, us, them, moves)
		}

		for _, step := range BishopDirections {
			generateDirectionalRay(from, step, us, them, moves)
		}

	}
}
