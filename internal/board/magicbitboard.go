package board

type PRNG uint64

var RookBits = [64]int{
	12, 11, 11, 11, 11, 11, 11, 12,
	11, 10, 10, 10, 10, 10, 10, 11,
	11, 10, 10, 10, 10, 10, 10, 11,
	11, 10, 10, 10, 10, 10, 10, 11,
	11, 10, 10, 10, 10, 10, 10, 11,
	11, 10, 10, 10, 10, 10, 10, 11,
	11, 10, 10, 10, 10, 10, 10, 11,
	12, 11, 11, 11, 11, 11, 11, 12,
}

var BishopBits = [64]int{
	6, 5, 5, 5, 5, 5, 5, 6,
	5, 5, 5, 5, 5, 5, 5, 5,
	5, 5, 7, 7, 7, 7, 5, 5,
	5, 5, 7, 9, 9, 7, 5, 5,
	5, 5, 7, 9, 9, 7, 5, 5,
	5, 5, 7, 7, 7, 7, 5, 5,
	5, 5, 5, 5, 5, 5, 5, 5,
	6, 5, 5, 5, 5, 5, 5, 6,
}

var (
	// attack masks minus squares on the edge
	RookMasks   [64]uint64
	BishopMasks [64]uint64

	// magic numbers
	RookMagics   [64]uint64
	BishopMagics [64]uint64

	// shift values
	RookShifts   [64]int
	BishopShifts [64]int

	// blocker table
	RookTable   [64][]uint64
	BishopTable [64][]uint64
)

// generates the next random number
func (p *PRNG) next() uint64 {
	x := uint64(*p)
	x ^= x << 13
	x ^= x >> 7
	x ^= x << 17
	*p = PRNG(x)

	return x
}

// generates a random number with few 1 bits
func (p *PRNG) sparseRandom() uint64 {
	return p.next() & p.next() & p.next()
}

func initMask() {
	for sq := 0; sq < 64; sq++ {
		RookMasks[sq] = createRookMask(sq)
		BishopMasks[sq] = createBishopMask(sq)
	}
}

func createRookMask(sq int) uint64 {
	rank := sq / 8
	file := sq % 8

	var rookMask uint64 = 0

	for r := 1; r < 7; r++ {
		if r == rank {
			continue
		}

		rookMask |= 1 << (r*8 + file)
	}

	for f := 1; f < 7; f++ {
		if f == file {
			continue
		}

		rookMask |= 1 << (rank*8 + f)
	}

	return rookMask
}

func createBishopMask(sq int) uint64 {
	rank := sq / 8
	file := sq % 8

	var bishopMask uint64 = 0

	for r, f := rank+1, file+1; r < 7 && f < 7; r, f = r+1, f+1 {
		bishopMask |= 1 << (r*8 + f)
	}
	for r, f := rank-1, file+1; r > 0 && f < 7; r, f = r-1, f+1 {
		bishopMask |= 1 << (r*8 + f)
	}
	for r, f := rank-1, file-1; r > 0 && f > 0; r, f = r-1, f-1 {
		bishopMask |= 1 << (r*8 + f)
	}
	for r, f := rank+1, file-1; r < 7 && f > 0; r, f = r+1, f-1 {
		bishopMask |= 1 << (r*8 + f)
	}

	return bishopMask
}

func createRookAttacksMask(sq int, blockers uint64) uint64 {
	var RookAttacks uint64 = 0
	var RookDirections = [4]int{8, 1, -8, -1}
	for _, step := range RookDirections {
		currentSq := sq

		for {
			if step == 8 && currentSq >= 56 {
				break
			}
			if step == 1 && currentSq%8 >= 7 {
				break
			}
			if step == -8 && currentSq <= 7 {
				break
			}
			if step == -1 && currentSq%8 <= 0 {
				break
			}

			currentSq += step

			var sqMask uint64 = 1 << currentSq
			RookAttacks |= sqMask

			if sqMask&blockers != 0 {
				break
			}
		}
	}

	return RookAttacks
}

func createBishopAttacksMask(sq int, blockers uint64) uint64 {
	var BishopAttacks uint64 = 0
	var BishopDirections = [4]int{9, -7, -9, 7}
	for _, step := range BishopDirections {
		currentSq := sq

		for {
			if step == 9 && (currentSq >= 56 || currentSq%8 >= 7) {
				break
			}
			if step == -7 && (currentSq <= 7 || currentSq%8 >= 7) {
				break
			}
			if step == -9 && (currentSq <= 7 || currentSq%8 <= 0) {
				break
			}
			if step == 7 && (currentSq >= 56 || currentSq%8 <= 0) {
				break
			}

			currentSq += step

			var sqMask uint64 = 1 << currentSq
			BishopAttacks |= sqMask

			if sqMask&blockers != 0 {
				break
			}
		}
	}

	return BishopAttacks
}

// func generateRookBlockers(sq int) uint64 {
// 	var blockers uint64

// 	numBits := RookBits[sq]
// 	numPatterns := 1 << numBits

// 	for i := 0; i < numPatterns; i++ {

// 	}

// }

// func generateRookMagics(sq int) uint64 {
// 	var magicNumber uint64

// 	for sq := 0; sq < 64; sq++ {
// 		rookMask := RookMasks[sq]
// 		numBits := RookBits[sq]
// 		numPatterns := 1 << numBits

// 		blockers := make([]uint64, numPatterns)
// 		attacks := make([]uint64, numPatterns)
// 		usedTable := make([]uint64, numPatterns)

// 		for i := 0; i < numPatterns; i++ {

// 		}

// 	}
// }
