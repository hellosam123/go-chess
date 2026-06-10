package board

import (
	"fmt"
	"math/bits"
	"os"
	"strings"
)

type PRNG uint64

var (
	rookMasks   [64]uint64
	bishopMasks [64]uint64

	fullRookMasks   [64]uint64
	fullBishopMasks [64]uint64

	rookRayMasks   [64][64]uint64
	bishopRayMasks [64][64]uint64

	rookTable   [64][]uint64
	bishopTable [64][]uint64
)

var rookMagics = [64]uint64{
	0x2480004000108020, 0xC040002000401000, 0x3100200300407008, 0x100082100100005,
	0x2100080050030064, 0x660010A200040801, 0x660010A200040801, 0x11000A2040820100,
	0x2C30800281400020, 0x280200081400C, 0x210180100080A000, 0x2801000800805,
	0x4100800400180180, 0x2002001024020098, 0x3152000408011200, 0x2000DA60300C4,
	0x208004C001201040, 0x182810040010020, 0x8200898010002000, 0x890020100100,
	0x2011010004901800, 0x102008004005280, 0x240008820710, 0x10020001608504,
	0x2C822180004001, 0x4C8400880200280, 0x1AD14100200100, 0x1100100080080084,
	0x80080040080, 0x200C040080420080, 0x2402100400020801, 0xA198A00104401,
	0xC400032800580, 0x11400084802002, 0x2200412005001300, 0x811009823003000,
	0x1000800C00800801, 0x10A00048A001008, 0x180080204009001, 0x1800030182000044,
	0x8040008A208000, 0x182810040010020, 0x200100490010, 0x800900100E10008,
	0x400100A498010010, 0x2002001024020098, 0x4100881241040030, 0x4110820004,
	0xC400032800580, 0x80140008109A100, 0x19300420008080, 0x890020100100,
	0x4100800400180180, 0x102008004005280, 0x3152000408011200, 0x88800243000080,
	0x10200410014802A, 0x200400080210051, 0xA81082002042, 0x300101880421,
	0x1842004820500C02, 0x451000204001803, 0x101001084020001, 0x1312300804C0072,
}

var rookBitShifts = [64]int{
	52, 53, 53, 53, 53, 53, 53, 52,
	53, 54, 54, 54, 54, 54, 54, 53,
	53, 54, 54, 54, 54, 54, 54, 53,
	53, 54, 54, 54, 54, 54, 54, 53,
	53, 54, 54, 54, 54, 54, 54, 53,
	53, 54, 54, 54, 54, 54, 54, 53,
	53, 54, 54, 54, 54, 54, 54, 53,
	52, 53, 53, 53, 53, 53, 53, 52,
}

var bishopMagics = [64]uint64{
	0xC002423212060200, 0x4010C10600820412, 0x1408008400838000, 0x4404081081002,
	0x41104080000420, 0xC2882008010000, 0x100C020110A98180, 0x1088084200210,
	0x200404481040, 0x211202809490020, 0x804800C9020000, 0x200040435800004,
	0x808041044122000, 0x100C020110A98180, 0x4002088805482080, 0xA104201900880,
	0x6510802820013402, 0x20040202220201, 0x4B40A8800421200, 0x4844004641020000,
	0x180800400A00441, 0x8082000100410444, 0x42040202C1041001, 0x4601015202490C80,
	0x220300004500201, 0x8010C0008900404, 0x10220114040400, 0x4040008410200,
	0x73080409004008, 0x200082012101088C, 0x5404024085861082, 0x882024008804800,
	0x4048080C4040C401, 0x828141002040101, 0x2400209000080A60, 0x8400040400080211,
	0x1010220020020028, 0x401202A200050044, 0x108024041009800, 0x92004910614402,
	0x42040202C1041001, 0x140E021004800280, 0x81084050000800, 0x100060212002C00,
	0x10400493040600, 0x204100070401200, 0x4010C10600820412, 0x18100400A48000A0,
	0x100C020110A98180, 0x2800B20104200000, 0xA104201900880, 0x1080042021000,
	0x2900003006020000, 0xC20408100900E4, 0xA02076840820A028, 0x4010C10600820412,
	0x1088084200210, 0xA104201900880, 0x8400000021280800, 0xD00502801420221,
	0x400924010020A20, 0x2280001082104908, 0x200404481040, 0xC002423212060200,
}

var bishopBitShifts = [64]int{
	58, 59, 59, 59, 59, 59, 59, 58,
	59, 59, 59, 59, 59, 59, 59, 59,
	59, 59, 57, 57, 57, 57, 59, 59,
	59, 59, 57, 55, 55, 57, 59, 59,
	59, 59, 57, 55, 55, 57, 59, 59,
	59, 59, 57, 57, 57, 57, 59, 59,
	59, 59, 59, 59, 59, 59, 59, 59,
	58, 59, 59, 59, 59, 59, 59, 58,
}

func init() {
	InitTables()
}

func FindAllMagics() {
	var (
		// magic numbers
		rookMagicsArray   [64]string
		bishopMagicsArray [64]string

		// shift values
		rookBitShiftsArray   [64]string
		bishopBitShiftsArray [64]string
	)

	fileName := "magics.txt"
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("Error creating file: %v", err)
		return
	}
	defer file.Close()

	for sq := 0; sq < 64; sq++ {
		rookMagic, rookBitShift := generateRookMagic(sq)
		bishopMagic, bishopBitShift := generateBishopMagic(sq)

		rookMagicsArray[sq] = fmt.Sprintf("%#X", rookMagic)
		rookBitShiftsArray[sq] = fmt.Sprintf("%d", rookBitShift)
		bishopMagicsArray[sq] = fmt.Sprintf("%#X", bishopMagic)
		bishopBitShiftsArray[sq] = fmt.Sprintf("%d", bishopBitShift)
	}

	rookMagicsStr := formatSqArray(rookMagicsArray, 4)
	rookBitShiftsStr := formatSqArray(rookBitShiftsArray, 8)
	bishopMagicsStr := formatSqArray(bishopMagicsArray, 4)
	bishopBitShiftsStr := formatSqArray(bishopBitShiftsArray, 8)

	fmt.Fprintln(file, "var rookMagics = [64]uint64{")
	fmt.Fprintf(file, "%s", rookMagicsStr)
	fmt.Fprintln(file, "}")

	fmt.Fprintln(file, "")

	fmt.Fprintln(file, "var rookBitShifts = [64]int{")
	fmt.Fprintf(file, "%s", rookBitShiftsStr)
	fmt.Fprintln(file, "}")

	fmt.Fprintln(file, "")

	fmt.Fprintln(file, "var bishopMagics = [64]uint64{")
	fmt.Fprintf(file, "%s", bishopMagicsStr)
	fmt.Fprintln(file, "}")

	fmt.Fprintln(file, "")

	fmt.Fprintln(file, "var bishopBitShifts = [64]int{")
	fmt.Fprintf(file, "%s", bishopBitShiftsStr)
	fmt.Fprintln(file, "}")
}

func GetMagicRookAttacksMask(b *Board, from int) uint64 {
	blockers := rookMasks[from] & b.AllPieces
	tableIndex := (blockers * rookMagics[from]) >> rookBitShifts[from]
	return rookTable[from][tableIndex]
}

func GetMagicBishopAttacksMask(b *Board, from int) uint64 {
	blockers := bishopMasks[from] & b.AllPieces
	tableIndex := (blockers * bishopMagics[from]) >> bishopBitShifts[from]
	return bishopTable[from][tableIndex]
}

func InitTables() {
	for sq := 0; sq < 64; sq++ {
		rookMasks[sq] = generateRookMask(sq)
		bishopMasks[sq] = generateBishopMask(sq)

		fullRookMasks[sq] = generateFullRookMask(sq)
		fullBishopMasks[sq] = generateFullBishopMask(sq)

		rookRayMasks[sq] = generateRookRayMasks(sq)
		bishopRayMasks[sq] = generateBishopRayMasks(sq)

		rookTableSize := 1 << (64 - rookBitShifts[sq])
		rookTable[sq] = make([]uint64, rookTableSize)

		bishopTableSize := 1 << (64 - bishopBitShifts[sq])
		bishopTable[sq] = make([]uint64, bishopTableSize)

		rookBlockers := generateRookBlockers(sq)
		bishopBlockers := generateBishopBlockers(sq)

		for _, b := range rookBlockers {
			tableIndex := (b * rookMagics[sq]) >> rookBitShifts[sq]
			rookTable[sq][tableIndex] = generateRookAttacksMask(sq, b)
		}

		for _, b := range bishopBlockers {
			tableIndex := (b * bishopMagics[sq]) >> bishopBitShifts[sq]
			bishopTable[sq][tableIndex] = generateBishopAttacksMask(sq, b)
		}
	}
}

func formatSqArray(sqArray [64]string, elementsPerLine int) string {
	var sb strings.Builder
	for i := 0; i < 64; i++ {
		sb.WriteString(sqArray[i])
		sb.WriteString(", ")
		if (i+1)%elementsPerLine == 0 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

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
	p1 := p.next()
	p2 := p.next()
	p3 := p.next()
	return p1 & p2 & p3
}

func generateRookMask(sq int) uint64 {
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

func generateFullRookMask(sq int) uint64 {
	rank := sq / 8
	file := sq % 8

	var rookMask uint64 = 0

	for r := 0; r < 8; r++ {
		if r == rank {
			continue
		}

		rookMask |= 1 << (r*8 + file)
	}

	for f := 0; f < 8; f++ {
		if f == file {
			continue
		}

		rookMask |= 1 << (rank*8 + f)
	}

	return rookMask
}

func generateBishopMask(sq int) uint64 {
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

func generateFullBishopMask(sq int) uint64 {
	rank := sq / 8
	file := sq % 8

	var bishopMask uint64 = 0

	for r, f := rank+1, file+1; r < 8 && f < 8; r, f = r+1, f+1 {
		bishopMask |= 1 << (r*8 + f)
	}
	for r, f := rank-1, file+1; r >= 0 && f < 8; r, f = r-1, f+1 {
		bishopMask |= 1 << (r*8 + f)
	}
	for r, f := rank-1, file-1; r >= 0 && f >= 0; r, f = r-1, f-1 {
		bishopMask |= 1 << (r*8 + f)
	}
	for r, f := rank+1, file-1; r < 8 && f >= 0; r, f = r+1, f-1 {
		bishopMask |= 1 << (r*8 + f)
	}

	return bishopMask
}

func generateRookAttacksMask(sq int, blockers uint64) uint64 {
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

func generateBishopAttacksMask(sq int, blockers uint64) uint64 {
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

func generateRookBlockers(sq int) []uint64 {
	var blockers []uint64

	rookMask := generateRookMask(sq)
	blockerMask := rookMask
	for {
		blockers = append(blockers, blockerMask)
		if blockerMask == 0 {
			break
		}
		blockerMask = (blockerMask - 1) & rookMask
	}

	return blockers
}

func generateBishopBlockers(sq int) []uint64 {
	var blockers []uint64

	bishopMask := generateBishopMask(sq)
	blockerMask := bishopMask
	for {
		blockers = append(blockers, blockerMask)
		if blockerMask == 0 {
			break
		}
		blockerMask = (blockerMask - 1) & bishopMask
	}

	return blockers
}

func generateRookMagic(sq int) (uint64, int) {
	var blockers []uint64 = generateRookBlockers(sq)

	// rook attacks for each blocker configuration
	attacks := make([]uint64, len(blockers))

	for i, b := range blockers {
		attacks[i] = generateRookAttacksMask(sq, b)
	}

	p := PRNG(31415926535)
	for bitShift := 54; bitShift > 50; bitShift-- {
		tableSize := 1 << (64 - bitShift)
		var magicIndexes = make([]uint64, tableSize)
		for i := 0; i < 1048576; i++ {
			clear(magicIndexes)

			magic := p.sparseRandom()
			var fail bool = false

			for j, b := range blockers {
				magicIndex := magic * b >> bitShift
				if magicIndexes[magicIndex] == 0 {
					magicIndexes[magicIndex] = attacks[j]
				} else if magicIndexes[magicIndex] != attacks[j] {
					fail = true
					break
				}

			}
			if !fail {
				return magic, bitShift
			}
		}
	}
	return 0, 0
}

func generateBishopMagic(sq int) (uint64, int) {
	var blockers []uint64 = generateBishopBlockers(sq)

	// rook attacks for each blocker configuration
	attacks := make([]uint64, len(blockers))

	for i, b := range blockers {
		attacks[i] = generateBishopAttacksMask(sq, b)
	}

	p := PRNG(31415926535)
	for bitShift := 59; bitShift > 54; bitShift-- {
		tableSize := 1 << (64 - bitShift)
		var magicIndexes = make([]uint64, tableSize)
		for i := 0; i < 131072; i++ {
			clear(magicIndexes)

			magic := p.sparseRandom()
			var fail bool = false

			for j, b := range blockers {
				magicIndex := magic * b >> bitShift
				if magicIndexes[magicIndex] == 0 {
					magicIndexes[magicIndex] = attacks[j]
				} else if magicIndexes[magicIndex] != attacks[j] {
					fail = true
					break
				}

			}
			if !fail {
				return magic, bitShift
			}
		}
	}
	return 0, 0
}

func generateRookRayMasks(sq int) [64]uint64 {
	fullRookMask := fullRookMasks[sq]
	var rayMasks [64]uint64
	for fullRookMask != 0 {
		to := bits.TrailingZeros64(fullRookMask)
		fullRookMask &= fullRookMask - 1

		var rayMask uint64

		if sq%8 == to%8 {
			if to > sq {
				for i := sq + 8; i < to; i += 8 {
					rayMask |= 1 << i
				}
			} else if to < sq {
				for i := sq - 8; i > to; i -= 8 {
					rayMask |= 1 << i
				}
			}
		} else if sq/8 == to/8 {
			if to > sq {
				for i := sq + 1; i < to; i++ {
					rayMask |= 1 << i
				}
			} else if to < sq {
				for i := sq - 1; i > to; i-- {
					rayMask |= 1 << i
				}
			}
		}

		rayMasks[to] = rayMask
	}

	return rayMasks
}

func generateBishopRayMasks(sq int) [64]uint64 {
	fullBishopMask := fullBishopMasks[sq]
	var rayMasks [64]uint64
	for fullBishopMask != 0 {
		to := bits.TrailingZeros64(fullBishopMask)
		fullBishopMask &= fullBishopMask - 1

		var rayMask uint64

		if sq%9 == to%9 {
			if to > sq {
				for i := sq + 9; i < to; i += 9 {
					rayMask |= 1 << i
				}
			} else if to < sq {
				for i := sq - 9; i > to; i -= 9 {
					rayMask |= 1 << i
				}
			}
		} else if sq%7 == to%7 {
			if to > sq {
				for i := sq + 7; i < to; i += 7 {
					rayMask |= 1 << i
				}
			} else if to < sq {
				for i := sq - 7; i > to; i -= 7 {
					rayMask |= 1 << i
				}
			}
		}

		rayMasks[to] = rayMask
	}

	return rayMasks
}
