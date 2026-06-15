package eval

// var Params = []int{
// 	82, 337, 365, 477, 1025, // mgValue (material) 0-4
// 	94, 281, 297, 512, 936, // egValue (material) 5-9
// 	10, 10, 5, 5, // mg weights for knights, bishops, rooks, and queens 10-13
// 	5, 5, 10, 5, // eg weights for kbrq 14-17
// 	10, 10, 20, 40, // king attack weight constants 18-21
// 	0, 0, 50, 75, 88, 94, 97, 99, // king num attacked weight 22-29
// 	0, 40, 70, 90, 95, // king pawn shield weight 30-34
// 	20, 20, // isolated and doubled penalty 35-36
// 	15, 10, // isolated penalty 37-38
// 	10, 10, // doubled penalty 39-40
// }

var Params = []int{
	82, 361, 415, 527, 1075,
	94, 306, 298, 543, 986,
	5, 4, 2, 0, 0,
	1, 4, 6, -4, -1,
	31, 47, 0, 0, 40,
	100, 112, 131, 97, 99,
	30, 42, 58, 81, 67,
	19, 15, 20, -1, 3,
	11,
}

const (
	MGMaterialValues            = 0
	EGMaterialValues            = 5
	MGMobilityWeightN           = 10
	MGMobilityWeightB           = 11
	MGMobilityWeightR           = 12
	MGMobilityWeightQ           = 13
	EGMobilityWeightN           = 14
	EGMobilityWeightB           = 15
	EGMobilityWeightR           = 16
	EGMobilityWeightQ           = 17
	KingAttackerWeightN         = 18
	KingAttackerWeightB         = 19
	KingAttackerWeightR         = 20
	KingAttackerWeightQ         = 21
	KingNumAttackedWeights      = 22
	KingPawnShieldWeights       = 30
	MGIsolatedAndDoubledPenalty = 35
	EGIsolatedAndDoubledPenalty = 36
	MGIsolatedPenalty           = 37
	EGIsolatedPenalty           = 38
	MGDoubledPenalty            = 39
	EGDoubledPenalty            = 40
)

var (
	MGValue                  [6]int
	EGValue                  [6]int
	KingNumAttackedWeightArr [8]int
	KingPawnShieldWeightArr  [5]int
)

func init() {
	InitParams()
}

func InitParams() {
	copy(MGValue[:5], Params[MGMaterialValues:MGMaterialValues+5])
	MGValue[5] = 0

	copy(EGValue[:5], Params[EGMaterialValues:EGMaterialValues+5])
	EGValue[5] = 0

	copy(KingNumAttackedWeightArr[:8], Params[KingNumAttackedWeights:KingNumAttackedWeights+8])
	copy(KingPawnShieldWeightArr[:5], Params[KingPawnShieldWeights:KingPawnShieldWeights+5])
}
