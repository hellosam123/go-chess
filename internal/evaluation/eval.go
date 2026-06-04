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
