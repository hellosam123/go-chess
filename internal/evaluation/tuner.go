package eval

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"

	"github.com/hellosam123/go-chess/internal/board"
)

type evalPosition struct {
	FEN    string
	Result float64
}

type parsedPosition struct {
	Board  *board.Board
	Result float64
}

func LoadEPDDataset(filepath string) ([]evalPosition, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var positions []evalPosition
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		parts := strings.Fields(line)
		fenStr := strings.Join(parts[:4], " ")

		resultStr := parts[5]
		resultStr = strings.ReplaceAll(resultStr, ";", "")
		resultStr = strings.ReplaceAll(resultStr, "\"", "")

		var result float64
		switch resultStr {
		case "1-0":
			result = 1
		case "1/2-1/2":
			result = 0.5
		case "0-1":
			result = 0
		default:
			return nil, fmt.Errorf("Invalid result: %s", line)
		}

		positions = append(positions, evalPosition{
			FEN:    fenStr,
			Result: result,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return positions, nil
}

func ParseEvalPositions(positions []evalPosition, numPositions int) []parsedPosition {
	var parsedPositions []parsedPosition
	var whiteWins []evalPosition
	var draws []evalPosition
	var blackWins []evalPosition
	var evalPositions []evalPosition

	for _, p := range positions {
		switch p.Result {
		case 1:
			whiteWins = append(whiteWins, p)
		case 0.5:
			draws = append(draws, p)
		case 0:
			blackWins = append(blackWins, p)
		}
	}

	numWhiteWin := numPositions * 4 / 10
	numDraw := numPositions * 2 / 10
	numBlackWin := numPositions * 4 / 10

	for i := 0; i < numWhiteWin; i++ {
		evalPositions = append(evalPositions, whiteWins[i])
	}
	for i := 0; i < numDraw; i++ {
		evalPositions = append(evalPositions, draws[i])
	}
	for i := 0; i < numBlackWin; i++ {
		evalPositions = append(evalPositions, blackWins[i])
	}

	for _, p := range evalPositions {
		b := board.NewStartingBoard()
		b.ParseFEN(p.FEN)

		parsedPositions = append(parsedPositions, parsedPosition{
			Board:  b,
			Result: p.Result,
		})
	}

	return parsedPositions
}

func CalculateMSE(positions []parsedPosition, K float64) float64 {
	var totalSquaredError float64

	numWorkers := 6
	chunkSize := len(positions) / numWorkers

	var wg sync.WaitGroup
	localSquaredErrors := make([]float64, numWorkers)

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			start := workerID * chunkSize
			end := start + chunkSize
			if workerID == numWorkers-1 {
				end = len(positions)
			}

			var localSquaredError float64
			for i := start; i < end; i++ {
				score := Evaluate(positions[i].Board)
				winProbability := 1 / (1 + math.Pow(10, -K*float64(score)/400))

				scoreError := positions[i].Result - winProbability
				localSquaredError += scoreError * scoreError
			}
			localSquaredErrors[workerID] = localSquaredError
		}(w)
	}

	wg.Wait()

	for _, localSquaredError := range localSquaredErrors {
		totalSquaredError += localSquaredError
	}

	return totalSquaredError / float64(len(positions))
}

func FindOptimalK(positions []parsedPosition) float64 {
	bestK := 1.42
	bestMSE := CalculateMSE(positions, bestK)

	for t := 0; t < 100; t++ {
		if MSE := CalculateMSE(positions, bestK+0.01); MSE < bestMSE {
			bestK += 0.01
			bestMSE = MSE
			continue
		}
		if MSE := CalculateMSE(positions, bestK-0.01); MSE < bestMSE {
			bestK -= 0.01
			bestMSE = MSE
			continue
		}
	}

	return bestK
}

func TexelTuner(positions []parsedPosition, maxCycles int) {
	var K = 1.42

	InitParams()
	bestMSE := CalculateMSE(positions, K)

	for cycle := 0; cycle < maxCycles; cycle++ {
		improved := false

		for index := 0; index < len(Params); index++ {
			if index == MGMaterialValues || index == EGMaterialValues {
				continue
			}

			Params[index] += 1
			InitParams()
			newMSE := CalculateMSE(positions, K)

			if newMSE < bestMSE {
				bestMSE = newMSE
				improved = true
				continue
			}

			Params[index] -= 2
			InitParams()
			newMSE = CalculateMSE(positions, K)
			if newMSE < bestMSE {
				bestMSE = newMSE
				improved = true
				continue
			}

			Params[index] += 1
		}

		if !improved {
			break
		}
	}

	fmt.Println(bestMSE)
}
