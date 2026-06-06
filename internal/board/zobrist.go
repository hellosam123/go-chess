package board

import (
	"math/rand"
)

var PieceSquareTable [12][64]uint64
var ActiveColorMask uint64
var CastlingTable [16]uint64
var EnPassantTable [8]uint64

func init() {
	initZobrist()
}

func initZobrist() {
	source := rand.NewSource(31415926535)
	r := rand.New(source)

	for piece := 0; piece < 12; piece++ {
		for sq := 0; sq < 64; sq++ {
			PieceSquareTable[piece][sq] = r.Uint64()
		}
	}

	ActiveColorMask = r.Uint64()

	for i := 0; i < 16; i++ {
		CastlingTable[i] = r.Uint64()
	}

	for i := 0; i < 8; i++ {
		EnPassantTable[i] = r.Uint64()
	}
}

func GenerateZobristKey(b *Board) uint64 {
	var finalKey uint64 = 0

	for sq := 0; sq < 64; sq++ {
		piece := b.GetPiece(sq)
		if piece != Empty {
			finalKey ^= PieceSquareTable[piece][sq]
		}
	}

	if !b.ActiveColor {
		finalKey ^= ActiveColorMask
	}

	finalKey ^= CastlingTable[b.CastlingRights]

	if b.EnPassantSquare != -1 {
		file := b.EnPassantSquare % 8
		finalKey ^= EnPassantTable[file]
	}

	return finalKey
}
