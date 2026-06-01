// Package movegen handles move generation, including move representation
// and generation of pseudo-legal and legal moves.
package movegen

import (
	"github.com/hellosam123/go-chess/internal/squares"
)

type Move uint16
type Flag uint16

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

func New(from int, to int, flag Flag) Move {

	return Move(uint16(from) | uint16(to)<<6 | uint16(flag))
}

func (m Move) GetFrom() int {
	return int(uint16(m) & FromMask)
}

func (m Move) GetTo() int {
	return int((uint16(m) & ToMask) >> 6)
}

func (m Move) GetFlag() Flag {
	return Flag(uint16(m) & FlagMask)
}

func (m Move) IsCapture() bool {
	return uint16(m)&(4<<12) != 0
}

func (m Move) IsPromotion() bool {
	return uint16(m)&(8<<12) != 0
}

func (m Move) String() string {
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
