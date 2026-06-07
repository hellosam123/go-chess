package eval

import (
	"math/bits"
	"unsafe"

	"github.com/hellosam123/go-chess/internal/board"
)

type HashFlag byte

const (
	Exact HashFlag = iota
	Alpha
	Beta
)

type TTEntry struct {
	HashKey  uint64
	Eval     int16
	BestMove board.Move
	Depth    int8
	Flag     HashFlag
}

type TranspositionTable struct {
	Entries []TTEntry
	Mask    uint64
}

// NewTranspositionTable returns a pointer to a transposition table with size sizeMB
func NewTranspositionTable(sizeMB int) *TranspositionTable {
	entrySize := int(unsafe.Sizeof(TTEntry{}))
	sizeB := sizeMB * 1024 * 1024

	numEntries := sizeB / entrySize

	if numEntries <= 0 {
		return nil
	}

	// floors to power of 2
	roundedNumEntries := 1 << (bits.Len(uint(numEntries)) - 1)

	return &TranspositionTable{
		Entries: make([]TTEntry, roundedNumEntries),
		Mask:    uint64(roundedNumEntries - 1),
	}
}

func (tt TranspositionTable) CountEntries() int {
	count := 0
	for i := 0; i < len(tt.Entries); i++ {
		if tt.Entries[i].HashKey != 0 {
			count++
		}
	}

	return count
}
