// Package board handles the internal representation of a chess board,
// including piece placement, game state tracking, FEN string conversion,
// move representation, and move generation.
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
	// an array of 12 bitboards, one for each piece.
	// index 0-5 for White, 6-11 for Black.
	// In order: pawn, knight, bishop, rook, queen, king.
	Pieces [12]uint64

	// bitboards for white, black, and all pieces
	WPieces   uint64
	BPieces   uint64
	AllPieces uint64

	// true for White, false for Black
	ActiveColor bool

	// bits 0-3, where 0: BQ, 1: BK, 2: WQ, 3: WK
	// in binary: 0000 WK-WQ-BK-BQ
	CastlingRights uint8

	// int (0-63), -1 for None
	EnPassantSquare int
}

// symbolToPiece converts a symbol (e.g. 'q') to a Piece (e.g. B_Queen)
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

// GetPiece gets the piece (type Piece) at a specific square index (0-63)
func (b *Board) GetPiece(square int) Piece {
	for piece, pieceMask := range b.Pieces {
		if (1<<square)&pieceMask != 0 {
			return Piece(piece)
		}
	}

	return Empty
}

// SetGeneralBitboards sets the bitboard values for WPieces, BPieces, and AllPieces
func (b *Board) SetGeneralBitboards() {
	// reset bitboards to 0
	b.WPieces = 0
	b.BPieces = 0
	b.AllPieces = 0

	for p := W_Pawn; p <= W_King; p++ {
		b.WPieces |= b.Pieces[p]
	}

	for p := B_Pawn; p <= B_King; p++ {
		b.BPieces |= b.Pieces[p]
	}

	b.AllPieces = b.WPieces | b.BPieces
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
	b := &Board{}
	b.ParseFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")

	return b
}
