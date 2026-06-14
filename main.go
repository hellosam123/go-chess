package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hellosam123/go-chess/internal/engine"
	"github.com/hellosam123/go-chess/uci"
)

func main() {
	fmt.Println("A Golang chess engine")

	engine := engine.NewEngine(32)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		tokens := strings.Fields(line)
		command := tokens[0]

		switch command {
		case "uci":
			fmt.Println("id name GoChess v1.0")
			fmt.Println("id author isfsam")
			fmt.Println("uciok")
		case "isready":
			fmt.Println("readyok")
		case "ucinewgame":
			engine.Board.ParseFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
			engine.ResetEngine()
		case "position":
			err := uci.HandlePosition(engine, tokens[1:])
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		case "go":
			err := uci.HandleGo(engine, tokens[1:])
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		case "print":
			engine.Board.PrintBoard()
		case "exit":
			return
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "UCI loop read error: %v\n", err)
			os.Exit(1)
		}
	}

}
