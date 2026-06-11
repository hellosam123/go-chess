// Package eval handles evaluation of positions
package eval

import (
	"math/bits"

	"github.com/hellosam123/go-chess/internal/board"
)

var mgValue = [6]int{82, 337, 365, 477, 1025, 0}
var egValue = [6]int{94, 281, 297, 512, 936, 0}

var mgPawnTable = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	98, 134, 61, 95, 68, 126, 34, -11,
	-6, 7, 26, 31, 65, 56, 25, -20,
	-14, 13, 6, 21, 23, 12, 17, -23,
	-27, -2, -5, 12, 17, 6, 10, -25,
	-26, -4, -4, -10, 3, 3, 33, -12,
	-35, -1, -20, -23, -15, 24, 38, -22,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var egPawnTable = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	178, 173, 158, 134, 147, 132, 165, 187,
	94, 100, 85, 67, 56, 53, 82, 84,
	32, 24, 13, 5, -2, 4, 17, 17,
	13, 9, -3, -7, -7, -8, 3, -1,
	4, 7, -6, 1, 0, -5, -1, -8,
	13, 8, 8, 10, 13, 0, 2, -7,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var mgKnightTable = [64]int{
	-167, -89, -34, -49, 61, -97, -15, -107,
	-73, -41, 72, 36, 23, 62, 7, -17,
	-47, 60, 37, 65, 84, 129, 73, 44,
	-9, 17, 19, 53, 37, 69, 18, 22,
	-13, 4, 16, 13, 28, 19, 21, -8,
	-23, -9, 12, 10, 19, 17, 25, -16,
	-29, -53, -12, -3, -1, 18, -14, -19,
	-105, -21, -58, -33, -17, -28, -19, -23,
}

var egKnightTable = [64]int{
	-58, -38, -13, -28, -31, -27, -63, -99,
	-25, -8, -25, -2, -9, -25, -24, -52,
	-24, -20, 10, 9, -1, -9, -19, -41,
	-17, 3, 22, 22, 22, 11, 8, -18,
	-18, -6, 16, 25, 16, 17, 4, -18,
	-23, -3, -1, 15, 10, -3, -20, -22,
	-42, -20, -10, -5, -2, -20, -23, -44,
	-29, -51, -23, -15, -22, -18, -50, -64,
}

var mgBishopTable = [64]int{
	-29, 4, -82, -37, -25, -42, 7, -8,
	-26, 16, -18, -13, 30, 59, 18, -47,
	-16, 37, 43, 40, 35, 50, 37, -2,
	-4, 5, 19, 50, 37, 37, 7, -2,
	-6, 13, 13, 26, 34, 12, 10, 4,
	0, 15, 15, 15, 14, 27, 18, 10,
	4, 15, 16, 0, 7, 21, 33, 1,
	-33, -3, -14, -21, -13, -12, -39, -21,
}

var egBishopTable = [64]int{
	-14, -21, -11, -8, -7, -9, -17, -24,
	-8, -4, 7, -12, -3, -13, -4, -14,
	2, -8, 0, -1, -2, 6, 0, 4,
	-3, 9, 12, 9, 14, 10, 3, 2,
	-6, 3, 13, 19, 7, 10, -3, -9,
	-12, -3, 8, 10, 13, 3, -7, -15,
	-14, -18, -7, -1, 4, -9, -15, -27,
	-23, -9, -23, -5, -9, -16, -5, -17,
}

var mgRookTable = [64]int{
	32, 42, 32, 51, 63, 9, 31, 43,
	27, 32, 58, 62, 80, 67, 26, 44,
	-5, 19, 26, 36, 17, 45, 61, 16,
	-24, -11, 7, 26, 24, 35, -8, -20,
	-36, -26, -12, -1, 9, -7, 6, -23,
	-45, -25, -16, -17, 3, 0, -5, -33,
	-44, -16, -20, -9, -1, 11, -6, -71,
	-19, -13, 1, 17, 16, 7, -37, -26,
}

var egRookTable = [64]int{
	13, 10, 18, 15, 12, 12, 8, 5,
	11, 13, 13, 11, -3, 3, 8, 3,
	7, 7, 7, 5, 4, -3, -5, -3,
	4, 3, 13, 1, 2, 1, -1, 2,
	3, 5, 8, 4, -5, -6, -8, -11,
	-4, 0, -5, -1, -7, -12, -8, -16,
	-6, -6, 0, 2, -9, -9, -11, -3,
	-9, 2, 3, -1, -5, -13, 4, -20,
}

var mgQueenTable = [64]int{
	-28, 0, 29, 12, 59, 44, 43, 45,
	-24, -39, -5, 1, -16, 57, 28, 54,
	-13, -17, 7, 8, 29, 56, 47, 57,
	-27, -27, -16, -16, -1, 17, -2, 1,
	-9, -26, -9, -10, -2, -4, 3, -3,
	-14, 2, -11, -2, -5, 2, 14, 5,
	-35, -8, 11, 2, 8, 15, -3, 1,
	-1, -18, -9, 10, -15, -25, -31, -50,
}

var egQueenTable = [64]int{
	-9, 22, 22, 27, 27, 19, 10, 20,
	-17, 20, 32, 41, 58, 25, 30, 0,
	-20, 6, 9, 49, 47, 35, 19, 9,
	3, 22, 24, 45, 57, 40, 57, 36,
	-18, 28, 19, 47, 31, 34, 39, 23,
	-16, -27, 15, 6, 9, 17, 10, 5,
	-22, -23, -30, -16, -16, -23, -36, -32,
	-33, -28, -22, -43, -5, -32, -20, -41,
}

var mgKingTable = [64]int{
	-65, 23, 16, -15, -56, -34, 2, 13,
	29, -1, -20, -7, -8, -4, -38, -29,
	-9, 24, 2, -16, -20, 6, 22, -22,
	-17, -20, -12, -27, -30, -25, -14, -36,
	-49, -1, -27, -39, -46, -44, -33, -51,
	-14, -14, -22, -46, -44, -30, -15, -27,
	1, 7, -8, -64, -43, -16, 9, 8,
	-15, 36, 12, -54, 8, -28, 24, 14,
}

var egKingTable = [64]int{
	-74, -35, -18, -18, -11, 15, 4, -17,
	-12, 17, 14, 17, 17, 38, 23, 11,
	10, 17, 23, 15, 20, 45, 44, 13,
	-8, 22, 24, 27, 26, 33, 26, 3,
	-18, -4, 21, 24, 27, 23, 9, -11,
	-19, -3, 11, 21, 23, 16, 7, -9,
	-27, -11, 4, 13, 14, 4, -5, -17,
	-53, -34, -21, -11, -28, -14, -24, -43,
}

var mgPassedPawnTable = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 10, 20, 25, 25, 20, 10, 0,
	0, 8, 15, 20, 20, 15, 8, 0,
	0, 5, 10, 15, 15, 10, 5, 0,
	0, 3, 7, 10, 10, 7, 3, 0,
	0, 2, 4, 6, 6, 4, 2, 0,
	0, 0, 1, 3, 3, 1, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var egPassedPawnTable = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	55, 35, 0, 0, 0, 0, 35, 55,
	40, 25, 0, 0, 0, 0, 25, 40,
	25, 15, 0, 0, 0, 0, 15, 25,
	15, 8, 0, 0, 0, 0, 8, 15,
	8, 4, 0, 0, 0, 0, 4, 8,
	4, 2, 0, 0, 0, 0, 2, 4,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var centerManhattanDistance = [64]int{
	6, 5, 4, 3, 3, 4, 5, 6,
	5, 4, 3, 2, 2, 3, 4, 5,
	4, 3, 2, 1, 1, 2, 3, 4,
	3, 2, 1, 0, 0, 1, 2, 3,
	3, 2, 1, 0, 0, 1, 2, 3,
	4, 3, 2, 1, 1, 2, 3, 4,
	5, 4, 3, 2, 2, 3, 4, 5,
	6, 5, 4, 3, 3, 4, 5, 6,
}

// Evaluate outputs a static score based on the evaluation of a position
func Evaluate(b *board.Board) int {
	var totalMGScore int
	var totalEGScore int

	attackMaps := GenerateAttackMaps(b)
	totalMGScore += attackMaps.WMGTotalScore - attackMaps.BMGTotalScore
	totalEGScore += attackMaps.WEGTotalScore - attackMaps.BEGTotalScore

	allPieceMask := b.AllPieces

	for allPieceMask != 0 {
		sq := bits.TrailingZeros64(allPieceMask)
		allPieceMask &= allPieceMask - 1
		piece := b.GetPiece(sq)
		if piece == board.Empty {
			continue
		}

		mgScore, egScore := GetPieceScore(b, piece, sq)
		if piece < board.B_Pawn {
			totalMGScore += mgScore
			totalEGScore += egScore
		} else {
			totalMGScore -= mgScore
			totalEGScore -= egScore
		}
	}

	totalScore := interpolateScore(totalMGScore, totalEGScore, b.PhaseValue)
	if b.PhaseValue <= 6 {
		if totalScore > 300 {
			totalScore += mopUpScore(b, board.W_King, board.B_King)
		} else if totalScore < -300 {
			totalScore -= mopUpScore(b, board.B_King, board.W_King)
		}
	}

	if !b.ActiveColor {
		totalScore = -totalScore
	}

	return totalScore
}

// mopUpScore outputs a score incentivizing bringing the king closer to the opponents king during mopups
func mopUpScore(b *board.Board, usKing board.Piece, themKing board.Piece) int {
	var usKingSq int = bits.TrailingZeros64(b.Pieces[usKing])
	var themKingSq int = bits.TrailingZeros64(b.Pieces[themKing])

	// weights are arbitrary
	cmdScore := centerManhattanDistance[themKingSq] * 10

	// 14 is maximum MD between two kings
	mdScore := (14 - getManhattanDistance(usKingSq, themKingSq)) * 4

	return cmdScore + mdScore
}

// GetPieceScore takes the position of a piece and outputs a middlegame
// and endgame score based on PeSTO piece square tables
func GetPieceScore(b *board.Board, p board.Piece, sq int) (int, int) {
	var mgScore int
	var egScore int
	var sqIndex int = sq

	// PeSTO tables start index with 0=A8
	if p <= board.W_King {
		sqIndex = flipSquare(sqIndex)
	}

	switch p {
	case board.W_Pawn, board.B_Pawn:
		mgScore, egScore = GetPawnScore(b, p, sq)
	case board.W_Knight, board.B_Knight:
		mgScore = mgValue[1] + mgKnightTable[sqIndex]
		egScore = egValue[1] + egKnightTable[sqIndex]
	case board.W_Bishop, board.B_Bishop:
		mgScore = mgValue[2] + mgBishopTable[sqIndex]
		egScore = egValue[2] + egBishopTable[sqIndex]
	case board.W_Rook, board.B_Rook:
		mgScore = mgValue[3] + mgRookTable[sqIndex]
		egScore = egValue[3] + egRookTable[sqIndex]
	case board.W_Queen, board.B_Queen:
		mgScore = mgValue[4] + mgQueenTable[sqIndex]
		egScore = egValue[4] + egQueenTable[sqIndex]
	case board.W_King, board.B_King:
		mgScore = mgKingTable[sqIndex]
		egScore = egKingTable[sqIndex]
	default:
		return 0, 0
	}

	return mgScore, egScore
}

func GetPawnScore(b *board.Board, p board.Piece, sq int) (int, int) {
	var mgScore int
	var egScore int
	var sqIndex = sq

	if p == board.W_Pawn {
		sqIndex = flipSquare(sqIndex)
	}

	mgScore = mgValue[0] + mgPawnTable[sqIndex]
	egScore = egValue[0] + egPawnTable[sqIndex]

	var isPassed bool
	var isIsolated bool
	var isDoubled bool

	switch p {
	case board.W_Pawn:
		isPassed = IsPassedPawn(b, sq, true)
		isIsolated = isIsolatedPawn(b, sq, true)
		isDoubled = isDoubledPawn(b, sq, true)

		if isPassed {
			mgScore += mgPassedPawnTable[sqIndex]
			egScore += egPassedPawnTable[sqIndex]

			wKingSq := bits.TrailingZeros64(b.Pieces[board.W_King])
			bKingSq := bits.TrailingZeros64(b.Pieces[board.B_King])
			egScore += (7 - getKingDistance(sq, wKingSq)) * 4
			egScore += (getKingDistance(sq, bKingSq) - 1) * 3

			// promotion square
			egScore += (7 - getKingDistance(56+sq%8, wKingSq)) * 2
		}

	case board.B_Pawn:
		isPassed = IsPassedPawn(b, sq, false)
		isIsolated = isIsolatedPawn(b, sq, false)
		isDoubled = isDoubledPawn(b, sq, false)

		if isPassed {
			mgScore += mgPassedPawnTable[sqIndex]
			egScore += egPassedPawnTable[sqIndex]

			wKingSq := bits.TrailingZeros64(b.Pieces[board.W_King])
			bKingSq := bits.TrailingZeros64(b.Pieces[board.B_King])
			egScore += (7 - getKingDistance(sq, bKingSq)) * 4
			egScore += (getKingDistance(sq, wKingSq) - 1) * 3
			egScore += (7 - getKingDistance(sq%8, bKingSq)) * 2
		}

	}

	if isIsolated && isDoubled {
		mgScore -= 20
		egScore -= 20
	} else if isIsolated {
		mgScore -= 15
		// for some reason an egScore penalty here reduces elo
		if !isPassed {
			egScore -= 10
		}
	} else if isDoubled {
		mgScore -= 10
		egScore -= 10
	}

	return mgScore, egScore
}

// GetPieceValue returns fixed material values for move ordering
func GetPieceValue(p board.Piece) int {
	const (
		pawnValue   = 100
		knightValue = 305
		bishopValue = 333
		rookValue   = 563
		queenValue  = 950
	)

	switch p {
	case board.W_Pawn, board.B_Pawn:
		return pawnValue
	case board.W_Knight, board.B_Knight:
		return knightValue
	case board.W_Bishop, board.B_Bishop:
		return bishopValue
	case board.W_Rook, board.B_Rook:
		return rookValue
	case board.W_Queen, board.B_Queen:
		return queenValue
	default:
		return 0
	}
}

func IsEndgame(b *board.Board) bool {
	return b.PhaseValue <= 8
}

// flipSquare takes a square and reflects it vertically
func flipSquare(square int) int {
	return square ^ 0b111000
}

// interpolateScore takes middlegame and endgame scores and outputs a single score
func interpolateScore(mgScore int, egScore int, phase int) int {
	if phase > 24 {
		phase = 24
	}

	egPhase := 24 - phase

	return (mgScore*int(phase) + egScore*int(egPhase)) / 24
}

func getManhattanDistance(sq1 int, sq2 int) int {
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

func getKingDistance(sq1 int, sq2 int) int {
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
