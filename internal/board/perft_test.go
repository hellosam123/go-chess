package board

func Perft(b *Board, depth int) int {
	if depth == 0 {
		return 1
	}

	var nodes uint64 = 0

	moves := GenerateLegalMoves(b)
	for _, m := range moves {

	}

	return 0
}
