package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/cmatsuoka/mazogs/game"
	"github.com/cmatsuoka/mazogs/graphics"
)

func run() error {
	if err := graphics.Init("Mazogs", 800, 600); err != nil {
		return err
	}
	defer graphics.Deinit()

	g := game.New()
	return g.Run()
}

func main() {
	seed := time.Now().UnixNano()
	rand.Seed(seed)

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
}
