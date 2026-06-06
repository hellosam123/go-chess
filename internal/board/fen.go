package board

import (
	"fmt"
	"strings"
)

// ParseFEN sets the current board state to a FEN.
func (b *Board) ParseFEN(fen string) error {
	*b = Board{}

	fields := strings.Split(fen, " ")
	if len(fields) < 4 {
		return fmt.Errorf("invalid FEN: expected at least 4 fields, got %d", len(fields))
	}

	piecePlacement := fields[0]
	activeColor := fields[1]
	castlingRights := fields[2]
	enPassantSquare := fields[3]

	ranks := strings.Split(piecePlacement, "/")
	if len(ranks) != 8 {
		return fmt.Errorf("invalid FEN: expected 8 ranks, got %d", len(ranks))
	}

	for r := 0; r < 8; r++ {
		rank := 7 - r
		file := 0

		// loop through FEN 1st field (xxxx/xxxx/xxxx/xxxx...)
		for _, symbol := range ranks[r] {
			if symbol >= '1' && symbol <= '8' {
				emptySquares := int(symbol - '0')
				file += emptySquares
			} else {
				piece := symbolToPiece(symbol)
				if piece == Empty {
					return fmt.Errorf("invalid FEN: unknown symbol %c", symbol)
				}

				if file > 8 {
					return fmt.Errorf("invalid FEN: rank %d overflows 8 files", 8-r)
				}

				b.SetPiece(piece, rank*8+file)
				file++
			}
		}
	}

	b.SetGeneralBitboards()

	if activeColor != "w" && activeColor != "b" {
		return fmt.Errorf("invalid FEN active color: expected 'w' or 'b', got '%s'", activeColor)
	}

	b.ActiveColor = activeColor == "w"

	if castlingRights != "-" {
		for _, char := range castlingRights {
			switch char {
			case 'K':
				b.CastlingRights |= 1 << 3
			case 'Q':
				b.CastlingRights |= 1 << 2
			case 'k':
				b.CastlingRights |= 1 << 1
			case 'q':
				b.CastlingRights |= 1 << 0
			default:
				return fmt.Errorf("invalid FEN castling right: unknown character '%c'", char)
			}
		}
	}

	if enPassantSquare != "-" {
		if len(enPassantSquare) != 2 || enPassantSquare[0] < 'a' || enPassantSquare[0] > 'h' || enPassantSquare[1] < '1' || enPassantSquare[1] > '8' {
			return fmt.Errorf("invalid FEN en passant square: %s", enPassantSquare)
		}
		f := int(enPassantSquare[0] - 'a')
		r := int(enPassantSquare[1] - '1')
		b.EnPassantSquare = r*8 + f
	} else {
		b.EnPassantSquare = -1
	}

	b.HashKey = GenerateZobristKey(b)

	return nil
}
