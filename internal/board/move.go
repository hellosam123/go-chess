package board

import (
	"github.com/hellosam123/go-chess/internal/squares"
)

type Move uint16
type Flag uint16

type UnMove struct {
	hashKey         uint64
	enPassantSquare int
	HalfMoveClock   int
	capturedPiece   Piece
	castlingRights  byte
}

// bitmasks to extract data from Move object
const (
	FromMask uint16 = 0x3F   // 0000 0000 0011 1111 (bits 0-5)
	ToMask   uint16 = 0xFC0  // 0000 1111 1100 0000 (bits 6-11)
	FlagMask uint16 = 0xF000 // 1111 0000 0000 0000 (bits 12-15)
)

// move flags, stored in bits 12-15 of Move object
const (
	QuietMove       Flag = 0 << 12
	DoublePawnPush  Flag = 1 << 12
	KingCastle      Flag = 2 << 12
	QueenCastle     Flag = 3 << 12
	Capture         Flag = 4 << 12
	EnPassant       Flag = 5 << 12
	PromoteN        Flag = 8 << 12
	PromoteB        Flag = 9 << 12
	PromoteR        Flag = 10 << 12
	PromoteQ        Flag = 11 << 12
	PromoteCaptureN Flag = 12 << 12
	PromoteCaptureB Flag = 13 << 12
	PromoteCaptureR Flag = 14 << 12
	PromoteCaptureQ Flag = 15 << 12
)

// New creates a new Move stored in uint16 format
func New(from int, to int, flag Flag) Move {

	return Move(uint16(from) | uint16(to)<<6 | uint16(flag))
}

// GetFrom returns the starting square as an int (0-63)
func (m Move) GetFrom() int {
	return int(uint16(m) & FromMask)
}

// GetTo returns the ending square as an int (0-63)
func (m Move) GetTo() int {
	return int((uint16(m) & ToMask) >> 6)
}

// GetFlag returns the flag as a Flag (uint16)
func (m Move) GetFlag() Flag {
	return Flag(uint16(m) & FlagMask)
}

// IsCapture returns if the move is a capture
func (m Move) IsCapture() bool {
	return uint16(m)&(4<<12) != 0
}

// IsPromotion returns if the move is a promotion
func (m Move) IsPromotion() bool {
	return uint16(m)&(8<<12) != 0
}

// MoveToString returns a Move formatted as a string in long algebraic notation (e.g. e2e4, a7a8q)
func (m Move) MoveToString() string {
	fromStr := squares.IndexToSquareArray[m.GetFrom()]
	toStr := squares.IndexToSquareArray[m.GetTo()]
	promotionStr := ""

	if m.IsPromotion() {
		switch m.GetFlag() {
		case PromoteN, PromoteCaptureN:
			promotionStr = "n"
		case PromoteB, PromoteCaptureB:
			promotionStr = "b"
		case PromoteR, PromoteCaptureR:
			promotionStr = "r"
		case PromoteQ, PromoteCaptureQ:
			promotionStr = "q"
		}
	}

	return fromStr + toStr + promotionStr
}

// MakeMove takes a Move input and updates the board state
func (b *Board) MakeMove(m Move) UnMove {
	unMove := UnMove{
		enPassantSquare: b.EnPassantSquare,
		castlingRights:  b.CastlingRights,
		capturedPiece:   0,
		hashKey:         b.HashKey,
	}

	if b.EnPassantSquare != -1 {
		b.HashKey ^= EnPassantTable[b.EnPassantSquare%8]
		b.EnPassantSquare = -1
	}

	from := m.GetFrom()
	to := m.GetTo()
	flag := m.GetFlag()

	switch flag {
	case DoublePawnPush:
		if b.ActiveColor {
			b.EnPassantSquare = to - 8
		} else {
			b.EnPassantSquare = to + 8
		}
	case KingCastle:
		if b.ActiveColor {
			b.RemovePiece(W_Rook, squares.H1)
			b.PutPiece(W_Rook, squares.F1)
		} else {
			b.RemovePiece(B_Rook, squares.H8)
			b.PutPiece(B_Rook, squares.F8)
		}
	case QueenCastle:
		if b.ActiveColor {
			b.RemovePiece(W_Rook, squares.A1)
			b.PutPiece(W_Rook, squares.D1)
		} else {
			b.RemovePiece(B_Rook, squares.A8)
			b.PutPiece(B_Rook, squares.D8)
		}
	case EnPassant:
		if b.ActiveColor {
			b.RemovePiece(B_Pawn, to-8)
		} else {
			b.RemovePiece(W_Pawn, to+8)
		}
	}

	if m.IsCapture() && flag != EnPassant {
		toPiece := b.GetPiece(to)
		b.RemovePiece(toPiece, to)
		unMove.capturedPiece = toPiece
	}

	fromPiece := b.GetPiece(from)
	b.RemovePiece(fromPiece, from)
	b.PutPiece(fromPiece, to)

	if m.IsPromotion() {
		b.RemovePiece(fromPiece, to)
		switch flag {
		case PromoteN, PromoteCaptureN:
			if b.ActiveColor {
				b.PutPiece(W_Knight, to)
			} else {
				b.PutPiece(B_Knight, to)
			}
		case PromoteB, PromoteCaptureB:
			if b.ActiveColor {
				b.PutPiece(W_Bishop, to)
			} else {
				b.PutPiece(B_Bishop, to)
			}
		case PromoteR, PromoteCaptureR:
			if b.ActiveColor {
				b.PutPiece(W_Rook, to)
			} else {
				b.PutPiece(B_Rook, to)
			}
		case PromoteQ, PromoteCaptureQ:
			if b.ActiveColor {
				b.PutPiece(W_Queen, to)
			} else {
				b.PutPiece(B_Queen, to)
			}
		}
	}

	if fromPiece == W_King || fromPiece == B_King || fromPiece == W_Rook || fromPiece == B_Rook {
		switch from {
		case squares.E1:
			b.HashKey ^= CastlingTable[b.CastlingRights]
			b.CastlingRights &^= 0b1100
			b.HashKey ^= CastlingTable[b.CastlingRights]
		case squares.H1:
			b.HashKey ^= CastlingTable[b.CastlingRights]
			b.CastlingRights &^= 0b1000
			b.HashKey ^= CastlingTable[b.CastlingRights]
		case squares.A1:
			b.HashKey ^= CastlingTable[b.CastlingRights]
			b.CastlingRights &^= 0b0100
			b.HashKey ^= CastlingTable[b.CastlingRights]
		case squares.E8:
			b.HashKey ^= CastlingTable[b.CastlingRights]
			b.CastlingRights &^= 0b0011
			b.HashKey ^= CastlingTable[b.CastlingRights]
		case squares.H8:
			b.HashKey ^= CastlingTable[b.CastlingRights]
			b.CastlingRights &^= 0b0010
			b.HashKey ^= CastlingTable[b.CastlingRights]
		case squares.A8:
			b.HashKey ^= CastlingTable[b.CastlingRights]
			b.CastlingRights &^= 0b0001
			b.HashKey ^= CastlingTable[b.CastlingRights]
		}
	}

	switch to {
	case squares.H1:
		b.HashKey ^= CastlingTable[b.CastlingRights]
		b.CastlingRights &^= 0b1000
		b.HashKey ^= CastlingTable[b.CastlingRights]
	case squares.A1:
		b.HashKey ^= CastlingTable[b.CastlingRights]
		b.CastlingRights &^= 0b0100
		b.HashKey ^= CastlingTable[b.CastlingRights]
	case squares.H8:
		b.HashKey ^= CastlingTable[b.CastlingRights]
		b.CastlingRights &^= 0b0010
		b.HashKey ^= CastlingTable[b.CastlingRights]
	case squares.A8:
		b.HashKey ^= CastlingTable[b.CastlingRights]
		b.CastlingRights &^= 0b0001
		b.HashKey ^= CastlingTable[b.CastlingRights]
	}

	b.ActiveColor = !b.ActiveColor
	b.HashKey ^= ActiveColorMask
	if b.EnPassantSquare != -1 {
		b.HashKey ^= EnPassantTable[b.EnPassantSquare%8]
	}
	b.SetGeneralBitboards()
	b.History = append(b.History, b.HashKey)

	unMove.HalfMoveClock = b.HalfMoveClock

	if fromPiece == W_Pawn || fromPiece == B_Pawn || m.IsCapture() {
		b.HalfMoveClock = 0
	} else {
		b.HalfMoveClock++
	}

	return unMove
}

// unMakeMove is the inverse of MakeMove
func (b *Board) UnMakeMove(m Move, unMove UnMove) {
	from := m.GetFrom()
	to := m.GetTo()
	flag := m.GetFlag()
	b.ActiveColor = !b.ActiveColor

	switch flag {
	case KingCastle:
		if b.ActiveColor {
			b.ClearPiece(W_Rook, squares.F1)
			b.PutPiece(W_Rook, squares.H1)
		} else {
			b.ClearPiece(B_Rook, squares.F8)
			b.PutPiece(B_Rook, squares.H8)
		}
	case QueenCastle:
		if b.ActiveColor {
			b.ClearPiece(W_Rook, squares.D1)
			b.PutPiece(W_Rook, squares.A1)
		} else {
			b.ClearPiece(B_Rook, squares.D8)
			b.PutPiece(B_Rook, squares.A8)
		}
	case EnPassant:
		if b.ActiveColor {
			b.PutPiece(B_Pawn, to-8)
		} else {
			b.PutPiece(W_Pawn, to+8)
		}
	}

	if m.IsPromotion() {
		switch flag {
		case PromoteN, PromoteCaptureN:
			if b.ActiveColor {
				b.ClearPiece(W_Knight, to)
			} else {
				b.ClearPiece(B_Knight, to)
			}
		case PromoteB, PromoteCaptureB:
			if b.ActiveColor {
				b.ClearPiece(W_Bishop, to)
			} else {
				b.ClearPiece(B_Bishop, to)
			}
		case PromoteR, PromoteCaptureR:
			if b.ActiveColor {
				b.ClearPiece(W_Rook, to)
			} else {
				b.ClearPiece(B_Rook, to)
			}
		case PromoteQ, PromoteCaptureQ:
			if b.ActiveColor {
				b.ClearPiece(W_Queen, to)
			} else {
				b.ClearPiece(B_Queen, to)
			}
		}
		if b.ActiveColor {
			b.SetPiece(W_Pawn, from)
		} else {
			b.SetPiece(B_Pawn, from)
		}
	} else {
		fromPiece := b.GetPiece(to)
		b.ClearPiece(fromPiece, to)
		b.SetPiece(fromPiece, from)
	}

	if m.IsCapture() && flag != EnPassant {
		b.SetPiece(unMove.capturedPiece, to)
	}

	b.EnPassantSquare = unMove.enPassantSquare
	b.CastlingRights = unMove.castlingRights
	b.HashKey = unMove.hashKey
	b.SetGeneralBitboards()
	b.History = b.History[:len(b.History)-1]
	b.HalfMoveClock = unMove.HalfMoveClock
}
