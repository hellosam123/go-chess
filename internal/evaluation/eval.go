// Package eval handles evaluation of positions
package eval

import (
	"math/bits"

	"github.com/hellosam123/go-chess/internal/board"
)

const (
	pawnValue   = 100
	knightValue = 305
	bishopValue = 333
	rookValue   = 563
	queenValue  = 950
)

func Evaluate(b *board.Board) int {
	var evaluation int
	var usMaterial int
	var themMaterial int
	if b.ActiveColor {
		usMaterial = countMaterial(b, true)
		themMaterial = countMaterial(b, false)
		evaluation = usMaterial - themMaterial
	} else {
		usMaterial = countMaterial(b, false)
		themMaterial = countMaterial(b, true)
		evaluation = usMaterial - themMaterial
	}

	return evaluation
}

func GetPieceValue(p board.Piece) int {
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

func countMaterial(b *board.Board, color bool) int {
	var material int
	if color {
		material += bits.OnesCount64(b.Pieces[board.W_Pawn]) * pawnValue
		material += bits.OnesCount64(b.Pieces[board.W_Knight]) * knightValue
		material += bits.OnesCount64(b.Pieces[board.W_Bishop]) * bishopValue
		material += bits.OnesCount64(b.Pieces[board.W_Rook]) * rookValue
		material += bits.OnesCount64(b.Pieces[board.W_Queen]) * queenValue
	} else {
		material += bits.OnesCount64(b.Pieces[board.B_Pawn]) * pawnValue
		material += bits.OnesCount64(b.Pieces[board.B_Knight]) * knightValue
		material += bits.OnesCount64(b.Pieces[board.B_Bishop]) * bishopValue
		material += bits.OnesCount64(b.Pieces[board.B_Rook]) * rookValue
		material += bits.OnesCount64(b.Pieces[board.B_Queen]) * queenValue
	}

	return material
}
