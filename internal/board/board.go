// Package board handles the internal representation of a chess board,
// including piece placement, game state tracking, and FEN string conversion.
package board

import "fmt"

type Piece int

const Empty Piece = 99
const (
	// iota starts at 0 and increments 1 for each const declaration
	// White pieces
	W_Pawn Piece = iota
	W_Knight
	W_Bishop
	W_Rook
	W_Queen
	W_King

	// Black pieces
	B_Pawn
	B_Knight
	B_Bishop
	B_Rook
	B_Queen
	B_King
)

var pieceToSymbolMap = map[Piece]rune{
	W_Pawn: 'P', W_Knight: 'N', W_Bishop: 'B', W_Rook: 'R', W_Queen: 'Q', W_King: 'K',
	B_Pawn: 'p', B_Knight: 'n', B_Bishop: 'b', B_Rook: 'r', B_Queen: 'q', B_King: 'k',
}

var symbolToPieceMap = map[rune]Piece{
	'P': W_Pawn, 'N': W_Knight, 'B': W_Bishop, 'R': W_Rook, 'Q': W_Queen, 'K': W_King,
	'p': B_Pawn, 'n': B_Knight, 'b': B_Bishop, 'r': B_Rook, 'q': B_Queen, 'k': B_King,
}

type Board struct {
	// an array of 12 bitboards. Index 0-5
	Pieces [12]uint64

	// true for White, false for Black
	ActiveColor bool

	// bits 0-3, where 0: BQ, 1: BK, 2: WQ, 3: WK
	// in binary: 0000 WK-WQ-BK-BQ
	CastlingRights uint8

	// int (0-63), -1 for None
	EnPassantSquare int
}

func symbolToPiece(symbol rune) Piece {
	if piece, ok := symbolToPieceMap[symbol]; ok {
		return piece
	} else {
		return Empty
	}
}

// SetPiece turns on the corresponding bit for a specific piece at a square index (0-63)
func (b *Board) SetPiece(piece Piece, square int) {
	var bit uint64 = 1 << square
	b.Pieces[piece] |= bit
}

// SetPiece turns off the corresponding bit for a specific piece at a square index (0-63)
func (b *Board) ClearPiece(piece Piece, square int) {
	var bit uint64 = 1 << square
	b.Pieces[piece] &^= bit
}

// PrintBoard prints the current board state to the console
func (b *Board) PrintBoard() {
	fmt.Println("  +-----------------+")
	for rank := 7; rank >= 0; rank-- {
		fmt.Printf("%d | ", rank+1)
		for file := 0; file <= 7; file++ {
			square := rank*8 + file
			var bit uint64 = 1 << square

			char := '.'
			for p := W_Pawn; p <= B_King; p++ {
				if (b.Pieces[p] & bit) != 0 {
					char = pieceToSymbolMap[p]
					break
				}
			}
			fmt.Printf("%c ", char)
		}
		fmt.Println("|")
	}
	fmt.Println("  +-----------------+")
	fmt.Println("    a b c d e f g h")

}

// NewStartingBoard creates a new Board object initialized with the starting position
func NewStartingBoard() *Board {
	b := &Board{
		ActiveColor:     true,
		CastlingRights:  0b1111,
		EnPassantSquare: -1,
	}

	b.Pieces[W_Pawn] = 0x000000000000FF00   // a2-h2 (bits 8-15)
	b.Pieces[W_Knight] = 0x0000000000000042 // b1,g1 (bits 1,7)
	b.Pieces[W_Bishop] = 0x0000000000000024 // c1,f1 (bits 2,6)
	b.Pieces[W_Rook] = 0x0000000000000081   // a1,h1 (bits 0,5)
	b.Pieces[W_Queen] = 0x0000000000000008  // d1 (bit 3)
	b.Pieces[W_King] = 0x0000000000000010   // e1 (bit 4)

	b.Pieces[B_Pawn] = 0x00FF000000000000   // a7-h7 (bits 48-55)
	b.Pieces[B_Knight] = 0x4200000000000000 // b8,g8 (bits 57,62)
	b.Pieces[B_Bishop] = 0x2400000000000000 // c8,f8 (bits 58,61)
	b.Pieces[B_Rook] = 0x8100000000000000   // a8,h8 (bits 56,63)
	b.Pieces[B_Queen] = 0x0800000000000000  // d8 (bit 59)
	b.Pieces[B_King] = 0x1000000000000000   // e8 (bit 60)

	return b
}
