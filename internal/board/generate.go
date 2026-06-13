package board

import (
	"fmt"
	"math/bits"

	"github.com/hellosam123/go-chess/internal/squares"
)

var KnightAttacks [64]uint64
var KingAttacks [64]uint64

const WhiteKingsideClearMask = (1 << squares.F1) | (1 << squares.G1)
const WhiteQueensideClearMask = (1 << squares.B1) | (1 << squares.C1) | (1 << squares.D1)
const BlackKingsideClearMask = (1 << squares.F8) | (1 << squares.G8)
const BlackQueensideClearMask = (1 << squares.B8) | (1 << squares.C8) | (1 << squares.D8)

func init() {
	initKnightAttacks()
	initKingAttacks()
}

// GenerateLegalMoves generates all legal moves in a board,
// and returns a list of moves and number of checkers
func (b *Board) GenerateLegalMoves() ([]Move, int) {
	moves := make([]Move, 0, 64)

	var kingSq int
	if b.ActiveColor {
		kingSq = bits.TrailingZeros64(b.Pieces[W_King])
	} else {
		kingSq = bits.TrailingZeros64(b.Pieces[B_King])
	}

	if kingSq == 64 {
		fmt.Printf("WTF??")
		b.PrintBoard()
	}

	checkMask, pinMasks, checkers := getCheckAndPinMasks(b, kingSq)

	generatePawnMoves(b, pinMasks, checkMask, checkers, &moves)
	generateKnightMoves(b, pinMasks, checkMask, checkers, &moves)
	generateBishopMoves(b, pinMasks, checkMask, checkers, &moves)
	generateRookMoves(b, pinMasks, checkMask, checkers, &moves)
	generateQueenMoves(b, pinMasks, checkMask, checkers, &moves)
	generateKingMoves(b, checkers, &moves)

	return moves, checkers
}

// generatePawnMoves generates all pseudo legal pawn moves in a board
func generatePawnMoves(b *Board, pinMasks [64]uint64, checkMask uint64, checkers int, moves *[]Move) {
	if checkers >= 2 {
		return
	}

	var pawns uint64
	var them uint64
	var rankStart uint64
	var rankEnd uint64
	var fileA uint64 = 0x101010101010101
	var fileH uint64 = 0x8080808080808080

	if b.ActiveColor {
		pawns = b.Pieces[W_Pawn]
		them = b.BPieces
		rankStart = 0xFF00
		rankEnd = 0xFF00000000000000
	} else {
		pawns = b.Pieces[B_Pawn]
		them = b.WPieces
		rankStart = 0xFF000000000000
		rankEnd = 0xFF
	}

	// a bitboard of all squares without pieces
	empty := ^b.AllPieces

	// pushMask is a bitboard of regular pawn pushes
	var pushMask uint64
	if b.ActiveColor {
		pushMask = (pawns << 8) & empty
	} else {
		pushMask = (pawns >> 8) & empty
	}

	var doublePushMask uint64
	if b.ActiveColor {
		doublePushMask = (pushMask << 8) & empty & ((pawns & rankStart) << 16)
	} else {
		doublePushMask = (pushMask >> 8) & empty & ((pawns & rankStart) >> 16)
	}

	var captureLeftMask uint64
	var captureRightMask uint64
	if b.ActiveColor {
		captureLeftMask = ((pawns &^ fileA) << 7) & them
		captureRightMask = ((pawns &^ fileH) << 9) & them
	} else {
		captureLeftMask = ((pawns &^ fileA) >> 9) & them
		captureRightMask = ((pawns &^ fileH) >> 7) & them
	}

	if checkers == 1 {
		pushMask &= checkMask
		doublePushMask &= checkMask
		captureLeftMask &= checkMask
		captureRightMask &= checkMask
	}

	// en passant mask
	var epLeftMask uint64
	var epRightMask uint64

	if b.EnPassantSquare != -1 {
		var epMask uint64 = 1 << b.EnPassantSquare
		if b.ActiveColor {
			epLeftMask = ((pawns &^ fileA) << 7) & epMask
			epRightMask = ((pawns &^ fileH) << 9) & epMask
		} else {
			epLeftMask = ((pawns &^ fileA) >> 9) & epMask
			epRightMask = ((pawns &^ fileH) >> 7) & epMask
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

		if pinMasks[from] == 0 || (1<<to)&pinMasks[from] != 0 {
			if (1<<to)&rankEnd != 0 {
				appendPromotions(from, to, moves)
			} else {
				*moves = append(*moves, New(from, to, QuietMove))
			}
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

		if pinMasks[from] == 0 || (1<<to)&pinMasks[from] != 0 {
			*moves = append(*moves, New(from, to, DoublePawnPush))
		}
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

		if pinMasks[from] == 0 || (1<<to)&pinMasks[from] != 0 {
			if (1<<to)&rankEnd != 0 {
				appendCapturePromotions(from, to, moves)
			} else {
				*moves = append(*moves, New(from, to, Capture))
			}
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

		if pinMasks[from] == 0 || (1<<to)&pinMasks[from] != 0 {
			if (1<<to)&rankEnd != 0 {
				appendCapturePromotions(from, to, moves)
			} else {
				*moves = append(*moves, New(from, to, Capture))
			}
		}
	}

	for epLeftMask != 0 {
		to := bits.TrailingZeros64(epLeftMask)
		epLeftMask = 0

		var from int
		var capturedPawnSq int
		if b.ActiveColor {
			from = to - 7
			capturedPawnSq = to - 8
		} else {
			from = to + 9
			capturedPawnSq = to + 8
		}

		checkValid := checkers == 0 || (1<<capturedPawnSq)&checkMask != 0
		pinValid := pinMasks[from] == 0 || (1<<to)&pinMasks[from] != 0

		if checkValid && pinValid && verifyHorizontalEPPin(b, from, capturedPawnSq) {
			*moves = append(*moves, New(from, to, EnPassant))
		}
	}

	for epRightMask != 0 {
		to := bits.TrailingZeros64(epRightMask)
		epRightMask = 0

		var from int
		var capturedPawnSq int
		if b.ActiveColor {
			from = to - 9
			capturedPawnSq = to - 8
		} else {
			from = to + 7
			capturedPawnSq = to + 8
		}

		checkValid := checkers == 0 || (1<<capturedPawnSq)&checkMask != 0
		pinValid := pinMasks[from] == 0 || (1<<to)&pinMasks[from] != 0

		if checkValid && pinValid && verifyHorizontalEPPin(b, from, capturedPawnSq) {
			*moves = append(*moves, New(from, to, EnPassant))
		}
	}
}

// generateKnightMoves generates all legal knight moves in a board
func generateKnightMoves(b *Board, pinMasks [64]uint64, checkMask uint64, checkers int, moves *[]Move) {
	if checkers >= 2 {
		return
	}

	var knights uint64
	var us uint64
	var them uint64
	// if White's turn
	if b.ActiveColor {
		knights = b.Pieces[W_Knight]
		us = b.WPieces
		them = b.BPieces
	} else {
		knights = b.Pieces[B_Knight]
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

		attacks := KnightAttacks[from]
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
func generateBishopMoves(b *Board, pinMasks [64]uint64, checkMask uint64, checkers int, moves *[]Move) {
	if checkers >= 2 {
		return
	}

	var bishops uint64
	var us uint64
	var them uint64
	if b.ActiveColor {
		bishops = b.Pieces[W_Bishop]
		us = b.WPieces
		them = b.BPieces
	} else {
		bishops = b.Pieces[B_Bishop]
		us = b.BPieces
		them = b.WPieces
	}

	for bishops != 0 {
		from := bits.TrailingZeros64(bishops)
		bishops &= bishops - 1

		var attacks uint64 = GetMagicBishopAttacksMask(b.AllPieces, from)
		attacks &^= us

		if checkers == 1 {
			attacks &= checkMask
		}

		if pinMasks[from] != 0 {
			attacks &= pinMasks[from]
		}

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

// generateRookMoves generates all pseudo legal rook moves in a board
func generateRookMoves(b *Board, pinMasks [64]uint64, checkMask uint64, checkers int, moves *[]Move) {
	if checkers >= 2 {
		return
	}

	var rooks uint64
	var us uint64
	var them uint64
	if b.ActiveColor {
		rooks = b.Pieces[W_Rook]
		us = b.WPieces
		them = b.BPieces
	} else {
		rooks = b.Pieces[B_Rook]
		us = b.BPieces
		them = b.WPieces
	}

	for rooks != 0 {
		from := bits.TrailingZeros64(rooks)
		rooks &= rooks - 1

		var attacks uint64 = GetMagicRookAttacksMask(b.AllPieces, from)
		attacks &^= us

		if checkers == 1 {
			attacks &= checkMask
		}

		if pinMasks[from] != 0 {
			attacks &= pinMasks[from]
		}

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

// generateQueenMoves generates all pseudo legal queen moves in a board
func generateQueenMoves(b *Board, pinMasks [64]uint64, checkMask uint64, checkers int, moves *[]Move) {
	if checkers >= 2 {
		return
	}
	var queens uint64
	var us uint64
	var them uint64
	if b.ActiveColor {
		queens = b.Pieces[W_Queen]
		us = b.WPieces
		them = b.BPieces
	} else {
		queens = b.Pieces[B_Queen]
		us = b.BPieces
		them = b.WPieces
	}

	for queens != 0 {
		from := bits.TrailingZeros64(queens)
		queens &= queens - 1

		var attacks uint64 = GetMagicRookAttacksMask(b.AllPieces, from) | GetMagicBishopAttacksMask(b.AllPieces, from)
		attacks &^= us

		if checkers == 1 {
			attacks &= checkMask
		}

		if pinMasks[from] != 0 {
			attacks &= pinMasks[from]
		}

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

// generateKingMoves generates all legal king moves in a board
func generateKingMoves(b *Board, checkers int, moves *[]Move) {
	var king uint64
	var us uint64
	var them uint64
	if b.ActiveColor {
		king = b.Pieces[W_King]
		us = b.WPieces
		them = b.BPieces
	} else {
		king = b.Pieces[B_King]
		us = b.BPieces
		them = b.WPieces
	}

	from := bits.TrailingZeros64(king)

	attacks := KingAttacks[from]
	attacks &^= us

	for attacks != 0 {

		to := bits.TrailingZeros64(attacks)
		attacks &= attacks - 1

		if isSquareAttacked(b, to, !b.ActiveColor, true) {
			continue
		}

		var flag Flag = QuietMove
		if (1<<to)&them != 0 {
			flag = Capture
		}

		*moves = append(*moves, New(from, to, flag))
	}

	if checkers == 0 {
		if b.ActiveColor {
			if from == squares.E1 {
				// White kingside castling
				if b.CastlingRights&0b1000 != 0 && b.AllPieces&WhiteKingsideClearMask == 0 {
					if !isSquareAttacked(b, squares.F1, false, true) && !isSquareAttacked(b, squares.G1, false, true) {
						*moves = append(*moves, New(from, squares.G1, KingCastle))
					}
				}
				// White queenside castling
				if b.CastlingRights&0b0100 != 0 && b.AllPieces&WhiteQueensideClearMask == 0 {
					if !isSquareAttacked(b, squares.C1, false, true) && !isSquareAttacked(b, squares.D1, false, true) {
						*moves = append(*moves, New(from, squares.C1, QueenCastle))
					}
				}
			}
		} else {
			if from == squares.E8 {
				// Black kingside castling
				if b.CastlingRights&0b0010 != 0 && b.AllPieces&BlackKingsideClearMask == 0 {
					if !isSquareAttacked(b, squares.F8, true, true) && !isSquareAttacked(b, squares.G8, true, true) {
						*moves = append(*moves, New(from, squares.G8, KingCastle))
					}
				}
				// Black queenside castling
				if b.CastlingRights&0b0001 != 0 && b.AllPieces&BlackQueensideClearMask == 0 {
					if !isSquareAttacked(b, squares.C8, true, true) && !isSquareAttacked(b, squares.D8, true, true) {
						*moves = append(*moves, New(from, squares.C8, QueenCastle))
					}
				}
			}
		}
	}
}

// initKnightAttacks calculates all knight attacks for each
// square and saves it in KnightAttacks as an array of bitboards
func initKnightAttacks() {
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
		KnightAttacks[sq] = attacks
	}
}

// initKingAttacks calculates all king attacks for each
// square and saves it in KingAttacks as an array of bitboards
func initKingAttacks() {
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
		KingAttacks[sq] = attacks
	}
}

// getCheckAndPinMasks returns a mask of squares between the king (exclusive)
// and pieces giving check (inclusive), an array of pin masks for each square,
// with each mask including the squares between the king (exclusive) and the
// pinner (inclusive), as well as the number of checkers
func getCheckAndPinMasks(b *Board, kingSq int) (uint64, [64]uint64, int) {
	var checkMask uint64 = 0
	var pinMasks [64]uint64
	var checkers int = 0

	var usPieces uint64
	var themPawns uint64
	var themKnights uint64
	var themBishops uint64
	var themRooks uint64
	var themQueens uint64

	if b.ActiveColor {
		usPieces = b.WPieces
		themPawns = b.Pieces[B_Pawn]
		themKnights = b.Pieces[B_Knight]
		themBishops = b.Pieces[B_Bishop]
		themRooks = b.Pieces[B_Rook]
		themQueens = b.Pieces[B_Queen]
	} else {
		usPieces = b.BPieces
		themPawns = b.Pieces[W_Pawn]
		themKnights = b.Pieces[W_Knight]
		themBishops = b.Pieces[W_Bishop]
		themRooks = b.Pieces[W_Rook]
		themQueens = b.Pieces[W_Queen]
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

	kingKnightAttacks := KnightAttacks[kingSq]
	if kingKnightAttacks&themKnights != 0 {
		checkMask |= kingKnightAttacks & themKnights
		checkers++
	}

	if checkers >= 2 {
		return checkMask, pinMasks, checkers
	}

	kingBishopAttacks := GetMagicBishopAttacksMask(b.AllPieces, kingSq)
	kingRookAttacks := GetMagicRookAttacksMask(b.AllPieces, kingSq)

	for themBishops != 0 {
		bishopSq := bits.TrailingZeros64(themBishops)
		themBishops &= themBishops - 1

		if fullBishopMasks[kingSq]&(1<<bishopSq) == 0 {
			continue
		}

		bishopAttacks := GetMagicBishopAttacksMask(b.AllPieces, bishopSq)
		rayCheck := kingBishopAttacks & (bishopAttacks | (1 << bishopSq))
		ray := bishopRayMasks[kingSq][bishopSq]
		if rayCheck == 0 {
			continue
		}
		if ray&b.AllPieces != 0 {
			rayPieces := ray & usPieces
			if rayPieces != 0 {
				if rayPieces&(rayPieces-1) == 0 {
					pinMasks[bits.TrailingZeros64(rayPieces)] = ray | (1 << bishopSq)
				}
			}
			continue
		}
		checkMask |= ray | (1 << bishopSq)
		checkers++

		if checkers >= 2 {
			return checkMask, pinMasks, checkers
		}

	}

	for themRooks != 0 {
		rookSq := bits.TrailingZeros64(themRooks)
		themRooks &= themRooks - 1

		if fullRookMasks[kingSq]&(1<<rookSq) == 0 {
			continue
		}

		rookAttacks := GetMagicRookAttacksMask(b.AllPieces, rookSq)
		rayCheck := kingRookAttacks & (rookAttacks | (1 << rookSq))
		ray := rookRayMasks[kingSq][rookSq]
		if rayCheck == 0 {
			continue
		}
		if ray&b.AllPieces != 0 {
			rayPieces := ray & usPieces
			if rayPieces != 0 {
				if rayPieces&(rayPieces-1) == 0 {
					pinMasks[bits.TrailingZeros64(rayPieces)] = ray | (1 << rookSq)
				}
			}
			continue
		}
		checkMask |= ray | (1 << rookSq)
		checkers++

		if checkers >= 2 {
			return checkMask, pinMasks, checkers
		}

	}

	for themQueens != 0 {
		queenSq := bits.TrailingZeros64(themQueens)
		themQueens &= themQueens - 1

		var rayCheck, ray uint64
		if fullRookMasks[kingSq]&(1<<queenSq) != 0 {
			rookAttacks := GetMagicRookAttacksMask(b.AllPieces, queenSq)
			rayCheck = kingRookAttacks & (rookAttacks | (1 << queenSq))
			ray = rookRayMasks[kingSq][queenSq]
		} else if fullBishopMasks[kingSq]&(1<<queenSq) != 0 {
			bishopAttacks := GetMagicBishopAttacksMask(b.AllPieces, queenSq)
			rayCheck = kingBishopAttacks & (bishopAttacks | (1 << queenSq))
			ray = bishopRayMasks[kingSq][queenSq]
		} else {
			continue
		}

		if rayCheck == 0 {
			continue
		}

		if ray&b.AllPieces != 0 {
			rayPieces := ray & usPieces
			if rayPieces != 0 {
				if rayPieces&(rayPieces-1) == 0 {
					pinMasks[bits.TrailingZeros64(rayPieces)] = ray | (1 << queenSq)
				}
			}
			continue
		}
		checkMask |= ray | (1 << queenSq)
		checkers++

		if checkers >= 2 {
			return checkMask, pinMasks, checkers
		}

	}

	return checkMask, pinMasks, checkers
}

// verifyHorizontalEPPin verifies that an en passant does not put the king in check
func verifyHorizontalEPPin(b *Board, from int, capturedPawnSq int) bool {
	var king uint64
	if b.ActiveColor {
		king = b.Pieces[W_King]
	} else {
		king = b.Pieces[B_King]
	}

	kingSq := bits.TrailingZeros64(king)
	if kingSq/8 != from/8 {
		return true
	}

	// a mask of all pieces except for the 2 pawns involved in en passant
	simMask := b.AllPieces &^ (1 << from) &^ (1 << capturedPawnSq)

	var themSliders uint64
	if b.ActiveColor {
		themSliders = b.Pieces[B_Rook] | b.Pieces[B_Queen]
	} else {
		themSliders = b.Pieces[W_Rook] | b.Pieces[W_Queen]
	}

	if themSliders == 0 {
		return true
	}

	if themSliders&GetMagicRookAttacksMask(simMask, kingSq) != 0 {
		return false
	}

	return true
}

// isSquareAttacked checks if a particular square is attacked by an attacking color,
// with an optional kingXray argument for valid king move detection
func isSquareAttacked(b *Board, square int, attackingColor bool, kingXray bool) bool {
	var fileA uint64 = 0x101010101010101
	var fileH uint64 = 0x8080808080808080
	var sqMask uint64 = 1 << square

	var occupancy uint64 = b.AllPieces

	var attackingPawns uint64
	var attackingKnights uint64
	var attackingBishops uint64
	var attackingRooks uint64
	var attackingQueens uint64
	var attackingKing uint64

	// White is attacking, check if Black king can enter the square
	if attackingColor {
		attackingPawns = b.Pieces[W_Pawn]
		attackingKnights = b.Pieces[W_Knight]
		attackingBishops = b.Pieces[W_Bishop]
		attackingRooks = b.Pieces[W_Rook]
		attackingQueens = b.Pieces[W_Queen]
		attackingKing = b.Pieces[W_King]
	} else {
		attackingPawns = b.Pieces[B_Pawn]
		attackingKnights = b.Pieces[B_Knight]
		attackingBishops = b.Pieces[B_Bishop]
		attackingRooks = b.Pieces[B_Rook]
		attackingQueens = b.Pieces[B_Queen]
		attackingKing = b.Pieces[B_King]
	}

	if KnightAttacks[square]&attackingKnights != 0 {
		return true
	}

	if KingAttacks[square]&attackingKing != 0 {
		return true
	}

	if attackingColor {

		var pawnAttacks uint64
		pawnAttacks = (sqMask&^fileA)>>9 | (sqMask&^fileH)>>7

		if pawnAttacks&attackingPawns != 0 {
			return true
		}

		if kingXray {
			occupancy &^= b.Pieces[B_King]
		}
	} else {
		var pawnAttacks uint64
		pawnAttacks = (sqMask&^fileA)<<7 | (sqMask&^fileH)<<9

		if pawnAttacks&attackingPawns != 0 {
			return true
		}

		if kingXray {
			occupancy &^= b.Pieces[W_King]
		}
	}

	bishopAttacks := GetMagicBishopAttacksMask(occupancy, square)

	if bishopAttacks&(attackingBishops|attackingQueens) != 0 {
		return true
	}

	rookAttacks := GetMagicRookAttacksMask(occupancy, square)

	if rookAttacks&(attackingRooks|attackingQueens) != 0 {
		return true
	}

	return false
}

// appendPromotions appends each promotion type to the move slice
func appendPromotions(from int, to int, moves *[]Move) {
	*moves = append(*moves,
		New(from, to, PromoteN),
		New(from, to, PromoteB),
		New(from, to, PromoteR),
		New(from, to, PromoteQ))
}

// appendCapturePromotions appends each capture promotion type to the move slice
func appendCapturePromotions(from int, to int, moves *[]Move) {
	*moves = append(*moves,
		New(from, to, PromoteCaptureN),
		New(from, to, PromoteCaptureB),
		New(from, to, PromoteCaptureR),
		New(from, to, PromoteCaptureQ))
}
