package board

import (
	"testing"
)

func TestFindAllMagics(t *testing.T) {
	FindAllMagics()
}

func TestInitTables(t *testing.T) {
	InitTables()
}

func TestGetMagicRookAttacksMask(t *testing.T) {
	b := NewStartingBoard()
	t.Logf("%#X", GetMagicRookAttacksMask(b.AllPieces, 16))
}
