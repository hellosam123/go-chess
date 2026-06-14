package board

import (
	"math/bits"
)

func (b *Board) IsEndgame() bool {
	return b.PhaseValue <= 8
}

func (b *Board) InCheck() bool {
	if b.ActiveColor {
		return isSquareAttacked(b, bits.TrailingZeros64(b.Pieces[W_King]), false, false)
	}
	return isSquareAttacked(b, bits.TrailingZeros64(b.Pieces[B_King]), true, false)
}

func GetManhattanDistance(sq1 int, sq2 int) int {
	fDiff := (sq1 % 8) - (sq2 % 8)

	// absolute value
	if fDiff < 0 {
		fDiff = -fDiff
	}

	rDiff := (sq1 / 8) - (sq2 / 8)
	if rDiff < 0 {
		rDiff = -rDiff
	}

	return fDiff + rDiff
}

func GetKingDistance(sq1 int, sq2 int) int {
	fDiff := (sq1 % 8) - (sq2 % 8)

	if fDiff < 0 {
		fDiff = -fDiff
	}

	rDiff := (sq1 / 8) - (sq2 / 8)
	if rDiff < 0 {
		rDiff = -rDiff
	}

	if fDiff > rDiff {
		return fDiff
	}
	return rDiff
}

func GetAttackersMask(b *Board, occupancy uint64, sq int, attackingColor bool) uint64 {
	var attackersMask uint64 = 0

	var fileA uint64 = 0x101010101010101
	var fileH uint64 = 0x8080808080808080
	var sqMask uint64 = 1 << sq

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

	if KnightAttacks[sq]&attackingKnights != 0 {
		attackersMask |= KnightAttacks[sq] & attackingKnights
	}

	if KingAttacks[sq]&attackingKing != 0 {
		attackersMask |= KingAttacks[sq] & attackingKing
	}

	var pawnAttacks uint64
	if attackingColor {
		pawnAttacks = (sqMask&^fileA)>>9 | (sqMask&^fileH)>>7
	} else {
		pawnAttacks = (sqMask&^fileA)<<7 | (sqMask&^fileH)<<9
	}

	if pawnAttacks&attackingPawns != 0 {
		attackersMask |= pawnAttacks & attackingPawns
	}

	bishopAttacks := GetMagicBishopAttacksMask(occupancy, sq)

	if bishopAttacks&(attackingBishops|attackingQueens) != 0 {
		attackersMask |= bishopAttacks & (attackingBishops | attackingQueens)
	}

	rookAttacks := GetMagicRookAttacksMask(occupancy, sq)

	if rookAttacks&(attackingRooks|attackingQueens) != 0 {
		attackersMask |= rookAttacks & (attackingRooks | attackingQueens)
	}

	return attackersMask
}

func GetSmallestAttacker(b *Board, attackersMask uint64, color bool) (Piece, int) {
	var startPiece Piece
	var endPiece Piece

	if color {
		startPiece = W_Pawn
		endPiece = W_King
	} else {
		startPiece = B_Pawn
		endPiece = B_King
	}
	for p := startPiece; p <= endPiece; p++ {
		if b.Pieces[p]&attackersMask != 0 {
			return p, bits.TrailingZeros64(b.Pieces[p] & attackersMask)
		}
	}
	return Empty, 0
}

func IsSlider(piece Piece) bool {
	return (piece >= W_Bishop && piece <= W_Queen) || (piece >= B_Bishop || piece <= B_Queen)
}

// CheckFiftyMoveRule checks if the current board position triggers the fifty move rule
func (b *Board) CheckFiftyMoveRule() bool {
	if b.HalfMoveClock >= 100 {
		return true
	}
	return false
}

// CheckRepetition checks if the current board position has already happened in the game history
func (b *Board) CheckRepetition() bool {
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

// HasNonPawnPieces checks if a side has non pawn pieces. Used for zugzwang detection.
func (b *Board) HasNonPawnPieces(color bool) bool {
	if color {
		return (b.Pieces[W_Pawn] | b.Pieces[W_King]) != b.AllPieces
	}
	return (b.Pieces[B_Pawn] | b.Pieces[B_King]) != b.AllPieces
}
