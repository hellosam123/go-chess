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

	var kingSq int
	if b.ActiveColor {
		kingSq = bits.TrailingZeros64(b.Pieces[board.W_King])
	} else {
		kingSq = bits.TrailingZeros64(b.Pieces[board.B_King])
	}

	pinMasks := getPinMasks(b, kingSq)
	checkMask, checkers := getCheckmask(b, kingSq)

	generatePawnMoves(b, pinMasks, checkMask, checkers, &moves)
	generateKnightMoves(b, pinMasks, checkMask, checkers, &moves)
	generateBishopMoves(b, pinMasks, checkMask, checkers, &moves)
	generateRookMoves(b, pinMasks, checkMask, checkers, &moves)
	generateQueenMoves(b, pinMasks, checkMask, checkers, &moves)
	// generateKingMoves(b, pinMasks, checkMask, checkers, &moves)

}

// generatePawnMoves generates all pseudo legal pawn moves in a board
func generatePawnMoves(b *board.Board, pinMasks [64]uint64, checkMask uint64, checkers int, moves *[]Move) {
	if checkers >= 2 {
		return
	}

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
func generateKnightMoves(b *board.Board, pinMasks [64]uint64, checkMask uint64, checkers int, moves *[]Move) {
	if checkers >= 2 {
		return
	}

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

		if pinMasks[from] != 0 {
			continue
		}

		attacks := knightAttacks[from]
		if checkers == 1 {
			attacks &= checkMask
		}
		attacks &^= us

		for attacks != 0 {
			to := bits.TrailingZeros64(attacks)
			attacks &= attacks - 1

			var flag Flag = QuietMove
			if (1<<to)&them != 0 {
				flag = Capture
			}

			*moves = append(*moves, New(from, to, flag))
		}
	}
}

// generateBishopMoves generates all pseudo legal bishop moves in a board
func generateBishopMoves(b *board.Board, pinMasks [64]uint64, checkMask uint64, checkers int, moves *[]Move) {
	if checkers >= 2 {
		return
	}

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
func generateRookMoves(b *board.Board, pinMasks [64]uint64, checkMask uint64, checkers int, moves *[]Move) {
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
func generateQueenMoves(b *board.Board, pinMasks [64]uint64, checkMask uint64, checkers int, moves *[]Move) {
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

	for attacks != 0 {

		to := bits.TrailingZeros64(attacks)
		attacks &= attacks - 1

		var flag Flag = QuietMove
		if (1<<to)&them != 0 {
			flag = Capture
		}

		*moves = append(*moves, New(from, to, flag))
	}
}

// generateDirectionalRayMask creates a ray by taking an initial position and
// repeatedly incrementing by a step until a collision, returning a mask
func generateDirectionalRayMask(from int, step int, us uint64, them uint64) uint64 {
	var rayMask uint64 = 0
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
		rayMask |= sqMask

		if sqMask&them != 0 {
			break
		}
	}
	return rayMask
}

func getPinMasks(b *board.Board, kingSq int) [64]uint64 {
	var pinMasks [64]uint64
	var us uint64
	var themBishops uint64
	var themRooks uint64
	var themQueens uint64

	if b.ActiveColor {
		us = b.WPieces
		themBishops = b.Pieces[board.B_Bishop]
		themRooks = b.Pieces[board.B_Rook]
		themQueens = b.Pieces[board.B_Queen]
	} else {
		us = b.BPieces
		themBishops = b.Pieces[board.W_Bishop]
		themRooks = b.Pieces[board.W_Rook]
		themQueens = b.Pieces[board.W_Queen]
	}

	var allDirections = [8]int{8, 9, 1, -7, -8, -9, -1, 7}

	for _, step := range allDirections {
		currentSq := kingSq
		var rayMask uint64 = 0
		var friendlyPieceCount = 0
		var pinnedSq int
		var validSliders uint64

		if step == 8 || step == 1 || step == -8 || step == -1 {
			validSliders = themRooks | themQueens
		} else {
			validSliders = themBishops | themQueens
		}

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
			rayMask |= sqMask

			if sqMask&us != 0 {
				friendlyPieceCount++
				if friendlyPieceCount == 1 {
					pinnedSq = currentSq
				} else {
					break
				}
			} else if sqMask&b.AllPieces != 0 {
				if sqMask&validSliders != 0 {
					if friendlyPieceCount == 1 {
						pinMasks[pinnedSq] = rayMask
					}
					break
				}
				break
			}
		}
	}
	return pinMasks
}

func getCheckmask(b *board.Board, kingSq int) (uint64, int) {
	var checkMask uint64 = 0
	var checkers int = 0

	var themPawns uint64
	var themKnights uint64
	var themBishops uint64
	var themRooks uint64
	var themQueens uint64

	if b.ActiveColor {
		themPawns = b.Pieces[board.B_Pawn]
		themKnights = b.Pieces[board.B_Knight]
		themBishops = b.Pieces[board.B_Bishop]
		themRooks = b.Pieces[board.B_Rook]
		themQueens = b.Pieces[board.B_Queen]
	} else {
		themPawns = b.Pieces[board.W_Pawn]
		themKnights = b.Pieces[board.W_Knight]
		themBishops = b.Pieces[board.W_Bishop]
		themRooks = b.Pieces[board.W_Rook]
		themQueens = b.Pieces[board.W_Queen]
	}

	kingFile := kingSq % 8
	if b.ActiveColor {
		if kingFile > 0 && kingSq+7 < 64 && 1<<(kingSq+7)&themPawns != 0 {
			checkMask |= 1 << (kingSq + 7)
			checkers++
		}
		if kingFile < 7 && kingSq+9 < 64 && 1<<(kingSq+9)&themPawns != 0 {
			checkMask |= 1 << (kingSq + 9)
			checkers++
		}
	} else {
		if kingFile > 0 && kingSq-9 >= 0 && 1<<(kingSq-9)&themPawns != 0 {
			checkMask |= 1 << (kingSq - 9)
			checkers++
		}
		if kingFile < 7 && kingSq-7 >= 0 && 1<<(kingSq-7)&themPawns != 0 {
			checkMask |= 1 << (kingSq - 7)
			checkers++
		}
	}

	kingKnightAttacks := knightAttacks[kingSq]
	if kingKnightAttacks&themKnights != 0 {
		checkMask |= kingKnightAttacks & themKnights
		checkers++
	}

	if checkers >= 2 {
		return checkMask, checkers
	}

	var allDirections = [8]int{8, 9, 1, -7, -8, -9, -1, 7}

	for _, step := range allDirections {
		currentSq := kingSq
		var rayMask uint64 = 0
		var validSliders uint64

		if step == 8 || step == 1 || step == -8 || step == -1 {
			validSliders = themRooks | themQueens
		} else {
			validSliders = themBishops | themQueens
		}

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
			rayMask |= sqMask

			if sqMask&b.AllPieces != 0 {
				if sqMask&validSliders != 0 {
					checkMask |= rayMask
					checkers++
				}
				break
			}
		}
		if checkers >= 2 {
			return checkMask, checkers
		}
	}

	return checkMask, checkers
}
