package eval

import (
	"math/bits"

	"github.com/hellosam123/go-chess/internal/board"
)

var wKingAttackZones [64]uint64
var bKingAttackZones [64]uint64

var wKingShieldZones [64]uint64
var bKingShieldZones [64]uint64

type AttackMaps struct {
	WAll   uint64
	BAll   uint64
	WPawns uint64
	BPawns uint64

	WMGTotalScore int
	WEGTotalScore int
	BMGTotalScore int
	BEGTotalScore int

	WAttackersCount int
	WAttackWeight   int
	BAttackersCount int
	BAttackWeight   int
}

func init() {
	initKingAttackZones()
	initKingShieldZones()
}

func GenerateAttackMaps(b *board.Board) AttackMaps {
	var attackMaps AttackMaps

	var fileA uint64 = 0x101010101010101
	var fileH uint64 = 0x8080808080808080

	wKingSq := bits.TrailingZeros64(b.Pieces[board.W_King])
	attackMaps.WAll |= board.KingAttacks[wKingSq]
	wKingAttackZone := wKingAttackZones[wKingSq]

	bKingSq := bits.TrailingZeros64(b.Pieces[board.B_King])
	attackMaps.BAll |= board.KingAttacks[bKingSq]
	bKingAttackZone := bKingAttackZones[bKingSq]

	wPawns := b.Pieces[board.W_Pawn]
	attackMaps.WPawns = ((wPawns &^ fileA) << 7) | ((wPawns &^ fileH) << 9)
	attackMaps.WAll |= attackMaps.WPawns

	bPawns := b.Pieces[board.B_Pawn]
	attackMaps.BPawns = ((bPawns &^ fileA) >> 9) | ((bPawns &^ fileH) >> 7)
	attackMaps.BAll |= attackMaps.BPawns

	wKnights := b.Pieces[board.W_Knight]
	for wKnights != 0 {
		knightSq := bits.TrailingZeros64(wKnights)
		wKnights &= wKnights - 1
		knightAttacks := board.KnightAttacks[knightSq]
		attackMaps.WAll |= knightAttacks

		if knightAttacks&bKingAttackZone != 0 {
			attackMaps.WAttackersCount++
			attackMaps.WAttackWeight += Params[KingAttackerWeightN]
		}

		mobilityMask := knightAttacks &^ attackMaps.BPawns &^ b.WPieces
		attackMaps.WMGTotalScore += bits.OnesCount64(mobilityMask) * Params[MGMobilityWeightN]
		attackMaps.WEGTotalScore += bits.OnesCount64(mobilityMask) * Params[EGMobilityWeightN]
	}

	bKnights := b.Pieces[board.B_Knight]
	for bKnights != 0 {
		knightSq := bits.TrailingZeros64(bKnights)
		bKnights &= bKnights - 1
		knightAttacks := board.KnightAttacks[knightSq]
		attackMaps.BAll |= knightAttacks

		if knightAttacks&wKingAttackZone != 0 {
			attackMaps.BAttackersCount++
			attackMaps.BAttackWeight += Params[KingAttackerWeightN]
		}

		mobilityMask := knightAttacks &^ attackMaps.WPawns &^ b.BPieces
		attackMaps.BMGTotalScore += bits.OnesCount64(mobilityMask) * Params[MGMobilityWeightN]
		attackMaps.BEGTotalScore += bits.OnesCount64(mobilityMask) * Params[EGMobilityWeightN]
	}

	wRooks := b.Pieces[board.W_Rook]
	for wRooks != 0 {
		rookSq := bits.TrailingZeros64(wRooks)
		wRooks &= wRooks - 1
		rookAttacks := board.GetMagicRookAttacksMask(b.AllPieces, rookSq)
		attackMaps.WAll |= rookAttacks

		if rookAttacks&bKingAttackZone != 0 {
			attackMaps.WAttackersCount++
			attackMaps.WAttackWeight += Params[KingAttackerWeightR]
		}

		mobilityMask := rookAttacks &^ attackMaps.BPawns &^ b.WPieces
		attackMaps.WMGTotalScore += bits.OnesCount64(mobilityMask) * Params[MGMobilityWeightR]
		attackMaps.WEGTotalScore += bits.OnesCount64(mobilityMask) * Params[EGMobilityWeightR]
	}

	bRooks := b.Pieces[board.B_Rook]
	for bRooks != 0 {
		rookSq := bits.TrailingZeros64(bRooks)
		bRooks &= bRooks - 1
		rookAttacks := board.GetMagicRookAttacksMask(b.AllPieces, rookSq)
		attackMaps.BAll |= rookAttacks

		if rookAttacks&wKingAttackZone != 0 {
			attackMaps.BAttackersCount++
			attackMaps.BAttackWeight += Params[KingAttackerWeightR]
		}

		mobilityMask := rookAttacks &^ attackMaps.WPawns &^ b.BPieces
		attackMaps.BMGTotalScore += bits.OnesCount64(mobilityMask) * Params[MGMobilityWeightR]
		attackMaps.BEGTotalScore += bits.OnesCount64(mobilityMask) * Params[EGMobilityWeightR]
	}

	wBishops := b.Pieces[board.W_Bishop]
	for wBishops != 0 {
		bishopSq := bits.TrailingZeros64(wBishops)
		wBishops &= wBishops - 1
		bishopAttacks := board.GetMagicBishopAttacksMask(b.AllPieces, bishopSq)
		attackMaps.WAll |= bishopAttacks

		if bishopAttacks&bKingAttackZone != 0 {
			attackMaps.WAttackersCount++
			attackMaps.WAttackWeight += Params[KingAttackerWeightB]
		}

		mobilityMask := bishopAttacks &^ attackMaps.BPawns &^ b.WPieces
		attackMaps.WMGTotalScore += bits.OnesCount64(mobilityMask) * Params[MGMobilityWeightB]
		attackMaps.WEGTotalScore += bits.OnesCount64(mobilityMask) * Params[EGMobilityWeightB]
	}

	bBishops := b.Pieces[board.B_Bishop]
	for bBishops != 0 {
		bishopSq := bits.TrailingZeros64(bBishops)
		bBishops &= bBishops - 1
		bishopAttacks := board.GetMagicBishopAttacksMask(b.AllPieces, bishopSq)
		attackMaps.BAll |= bishopAttacks

		if bishopAttacks&wKingAttackZone != 0 {
			attackMaps.BAttackersCount++
			attackMaps.BAttackWeight += Params[KingAttackerWeightB]
		}

		mobilityMask := bishopAttacks &^ attackMaps.WPawns &^ b.BPieces
		attackMaps.BMGTotalScore += bits.OnesCount64(mobilityMask) * Params[MGMobilityWeightB]
		attackMaps.BEGTotalScore += bits.OnesCount64(mobilityMask) * Params[EGMobilityWeightB]
	}

	wQueens := b.Pieces[board.W_Queen]
	for wQueens != 0 {
		queenSq := bits.TrailingZeros64(wQueens)
		wQueens &= wQueens - 1
		queenAttacks := board.GetMagicRookAttacksMask(b.AllPieces, queenSq)
		queenAttacks |= board.GetMagicBishopAttacksMask(b.AllPieces, queenSq)
		attackMaps.WAll |= queenAttacks

		if queenAttacks&bKingAttackZone != 0 {
			attackMaps.WAttackersCount++
			attackMaps.WAttackWeight += Params[KingAttackerWeightQ]
		}

		mobilityMask := queenAttacks &^ attackMaps.BPawns &^ b.WPieces
		attackMaps.WMGTotalScore += bits.OnesCount64(mobilityMask) * Params[MGMobilityWeightQ]
		attackMaps.WEGTotalScore += bits.OnesCount64(mobilityMask) * Params[EGMobilityWeightQ]
	}

	bQueens := b.Pieces[board.B_Queen]
	for bQueens != 0 {
		queenSq := bits.TrailingZeros64(bQueens)
		bQueens &= bQueens - 1
		queenAttacks := board.GetMagicRookAttacksMask(b.AllPieces, queenSq)
		queenAttacks |= board.GetMagicBishopAttacksMask(b.AllPieces, queenSq)
		attackMaps.BAll |= queenAttacks

		if queenAttacks&wKingAttackZone != 0 {
			attackMaps.BAttackersCount++
			attackMaps.BAttackWeight += Params[KingAttackerWeightQ]
		}

		mobilityMask := queenAttacks &^ attackMaps.WPawns &^ b.BPieces
		attackMaps.BMGTotalScore += bits.OnesCount64(mobilityMask) * Params[MGMobilityWeightQ]
		attackMaps.BEGTotalScore += bits.OnesCount64(mobilityMask) * Params[EGMobilityWeightQ]
	}

	if attackMaps.WAttackersCount >= 2 {
		if attackMaps.WAttackersCount > 7 {
			attackMaps.WAttackersCount = 7
		}
		attackMaps.WMGTotalScore += attackMaps.WAttackWeight * KingNumAttackedWeightArr[attackMaps.WAttackersCount] / 100
	}

	if attackMaps.BAttackersCount >= 2 {
		if attackMaps.BAttackersCount > 7 {
			attackMaps.BAttackersCount = 7
		}
		attackMaps.BMGTotalScore += attackMaps.BAttackWeight * KingNumAttackedWeightArr[attackMaps.BAttackersCount] / 100
	}

	return attackMaps
}

func PawnShieldEval(b *board.Board, color bool) int {
	var usPawns uint64
	var kingShieldZone uint64
	if color {
		usPawns = b.Pieces[board.W_Pawn]
		kingShieldZone = wKingShieldZones[bits.TrailingZeros64(b.Pieces[board.W_King])]
	} else {
		usPawns = b.Pieces[board.B_Pawn]
		kingShieldZone = bKingShieldZones[bits.TrailingZeros64(b.Pieces[board.B_King])]
	}

	pawnCount := bits.OnesCount64(usPawns & kingShieldZone)
	if pawnCount > 4 {
		pawnCount = 4
	}
	return KingPawnShieldWeightArr[pawnCount]
}

func initKingAttackZones() {
	for sq := 0; sq < 64; sq++ {
		rank := sq / 8
		file := sq % 8

		var kingAttackRow uint64 = 1 << sq
		if file > 0 {
			kingAttackRow |= 1 << (sq - 1)
		}
		if file < 7 {
			kingAttackRow |= 1 << (sq + 1)
		}

		wKingAttackZone := kingAttackRow
		if rank > 0 {
			wKingAttackZone |= kingAttackRow >> 8
		}

		for i, r := 1, rank; r < 7 && i <= 2; i, r = i+1, r+1 {
			wKingAttackZone |= kingAttackRow << (8 * i)
		}

		wKingAttackZones[sq] = wKingAttackZone

		bKingAttackZone := kingAttackRow
		if rank < 7 {
			bKingAttackZone |= kingAttackRow << 8
		}

		for i, r := 1, rank; r > 0 && i <= 2; i, r = i+1, r-1 {
			bKingAttackZone |= kingAttackRow >> (8 * i)
		}

		bKingAttackZones[sq] = bKingAttackZone
	}
}

func initKingShieldZones() {
	for sq := 0; sq < 64; sq++ {
		rank := sq / 8
		file := sq % 8

		var kingShieldRow uint64 = 1 << sq
		if file > 0 {
			kingShieldRow |= 1 << (sq - 1)
		}
		if file < 7 {
			kingShieldRow |= 1 << (sq + 1)
		}

		wKingShieldZone := kingShieldRow

		for i, r := 1, rank; r < 7 && i <= 2; i, r = i+1, r+1 {
			wKingShieldZone |= kingShieldRow << (8 * i)
		}

		wKingShieldZones[sq] = wKingShieldZone

		bKingShieldZone := kingShieldRow
		if rank < 7 {
			bKingShieldZone |= kingShieldRow << 8
		}

		for i, r := 1, rank; r > 0 && i <= 2; i, r = i+1, r-1 {
			bKingShieldZone |= kingShieldRow >> (8 * i)
		}

		bKingShieldZones[sq] = bKingShieldZone
	}
}
