package eval

import "github.com/hellosam123/go-chess/internal/board"

// an array of bitboards for the squares in front of a pawn, for each color and every square
var pawnFrontSpanMasks [2][64]uint64
var pawnFrontMasks [2][64]uint64

// includes all squares surrounding the pawn, not just those in front of it
var pawnFileMasks [64]uint64
var pawnSideMasks [64]uint64

func init() {
	initPawnMasks()
}

func initPawnMasks() {
	var fileA uint64 = 0x101010101010101

	for sq := 0; sq < 64; sq++ {
		file := sq % 8
		rank := sq / 8

		fileMask := fileA << file
		var fileSideMask uint64
		if file-1 >= 0 {
			fileSideMask |= fileA << (file - 1)
		}
		if file+1 < 8 {
			fileSideMask |= fileA << (file + 1)
		}

		var rankMask uint64 = 0xFFFFFFFFFFFFFFFF << ((rank + 1) * 8)

		var inverseRankMask uint64 = (1 << (rank * 8)) - 1

		pfsMask := (fileMask | fileSideMask) & rankMask
		pfMask := fileMask & rankMask
		inversePFSMask := (fileMask | fileSideMask) & inverseRankMask
		inversePFMask := fileMask & inverseRankMask

		pawnFrontSpanMasks[0][sq] = pfsMask
		pawnFrontMasks[0][sq] = pfMask
		pawnFrontSpanMasks[1][sq] = inversePFSMask
		pawnFrontMasks[1][sq] = inversePFMask

		pawnFileMasks[sq] = fileMask &^ (1 << sq)
		pawnSideMasks[sq] = fileSideMask
	}
}

func IsPassedPawn(b *board.Board, sq int, color bool) bool {
	if color {
		return pawnFrontSpanMasks[0][sq]&b.Pieces[board.B_Pawn] == 0 && pawnFrontMasks[0][sq]&b.Pieces[board.W_Pawn] == 0
	}
	return pawnFrontSpanMasks[1][sq]&b.Pieces[board.W_Pawn] == 0 && pawnFrontMasks[1][sq]&b.Pieces[board.B_Pawn] == 0
}

func isIsolatedPawn(b *board.Board, sq int, color bool) bool {
	if color {
		return pawnSideMasks[sq]&b.Pieces[board.W_Pawn] == 0
	}
	return pawnSideMasks[sq]&b.Pieces[board.B_Pawn] == 0
}

func isDoubledPawn(b *board.Board, sq int, color bool) bool {
	if color {
		return pawnFileMasks[sq]&b.Pieces[board.W_Pawn] != 0
	}
	return pawnFileMasks[sq]&b.Pieces[board.B_Pawn] != 0
}
