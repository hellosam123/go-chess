package eval

import (
	"fmt"
	"testing"
)

func TestLoadEPDDataset(t *testing.T) {
	positions, _ := LoadEPDDataset("quiet-labeled.epd")
	fmt.Println(positions[:10])

}

func TestCalculateMSE(t *testing.T) {
	positions, err := LoadEPDDataset("quiet-labeled.epd")
	if err != nil {
		t.Log(err)
	}

	parsedPositions := ParseEvalPositions(positions, 100000)

	t.Log(CalculateMSE(parsedPositions, 1.42))
}

func TestFindOptimalK(t *testing.T) {
	positions, err := LoadEPDDataset("quiet-labeled.epd")
	if err != nil {
		t.Log(err)
	}

	parsedPositions := ParseEvalPositions(positions, 100000)

	t.Log(FindOptimalK(parsedPositions))
}

func TestTexelTuner(t *testing.T) {
	positions, err := LoadEPDDataset("quiet-labeled.epd")
	if err != nil {
		t.Log(err)
	}

	parsedPositions := ParseEvalPositions(positions, 100000)

	TexelTuner(parsedPositions, 50)
	fmt.Printf("var Params = []int{\n")
	for i, val := range Params {
		fmt.Printf("%d, ", val)
		if (i+1)%5 == 0 {
			fmt.Printf("\n")
		}
	}
	fmt.Printf("\n}\n")
}
