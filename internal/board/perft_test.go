package board

import (
	"testing"
	"time"
)

func Perft(b *Board, depth int) int {
	if depth == 0 {
		return 1
	}

	var nodes int = 0

	moves, _ := b.GenerateLegalMoves()
	for _, m := range moves {
		unMove := b.MakeMove(m)
		nodes += Perft(b, depth-1)
		b.UnMakeMove(m, unMove)
	}

	return nodes
}

func TestPerft(t *testing.T) {
	gameBoard := NewStartingBoard()
	gameBoard.ParseFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	t.Log(Perft(gameBoard, 6))
}

func TestPerftAll(t *testing.T) {
	type perftPosition struct {
		position string
		depth    []int
		nodes    []int
	}

	var perftPositions []perftPosition
	// starting position
	perftPositions = append(perftPositions, perftPosition{
		position: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		depth:    []int{1, 2, 3, 4, 5},
		nodes:    []int{20, 400, 8902, 197281, 4865609},
	})

	// kiwipete
	perftPositions = append(perftPositions, perftPosition{
		position: "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq -",
		depth:    []int{1, 2, 3, 4},
		nodes:    []int{48, 2039, 97862, 4085603},
	})

	perftPositions = append(perftPositions, perftPosition{
		position: "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		depth:    []int{1, 2, 3, 4, 5, 6},
		nodes:    []int{14, 191, 2812, 43238, 674624, 11030083},
	})

	perftPositions = append(perftPositions, perftPosition{
		position: "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
		depth:    []int{1, 2, 3, 4, 5},
		nodes:    []int{6, 264, 9467, 422333, 15833292},
	})

	perftPositions = append(perftPositions, perftPosition{
		position: "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
		depth:    []int{1, 2, 3, 4},
		nodes:    []int{44, 1486, 62379, 2103487},
	})

	perftPositions = append(perftPositions, perftPosition{
		position: "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
		depth:    []int{1, 2, 3, 4},
		nodes:    []int{46, 2079, 89890, 3894594},
	})

	initialStartTime := time.Now()
	for _, p := range perftPositions {
		b := NewStartingBoard()
		b.ParseFEN(p.position)
		for i, d := range p.depth {
			startTime := time.Now()
			perftNodes := Perft(b, d)
			var successMsg string
			if perftNodes == p.nodes[i] {
				successMsg = "SUCCESS"
			} else {
				successMsg = "FAIL"
			}
			t.Logf("%v %s: position %s at depth %d expected %d nodes, got %d\n", time.Since(startTime), successMsg, p.position, d, p.nodes[i], perftNodes)
		}
	}
	t.Logf("Total time for all tests: %v", time.Since(initialStartTime))
}
