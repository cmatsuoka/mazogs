package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cmatsuoka/mazogs/game"
	"github.com/cmatsuoka/mazogs/maze"
)

func displayMap(themap []byte) {
	for i := 0; i < maze.MazeRows; i++ {
		for j := 0; j < maze.MazeColumns; j++ {
			c := "  "
			switch themap[i*maze.MazeColumns+j] {
			case maze.InternalWall:
				c = "██"
			case maze.ExternalWall:
				c = "▒▒"
			case maze.Sword:
				c = "🗡️ "
			case maze.PlayerStanding:
				c = "🧍"
			case maze.Prisoner, maze.Prisoner2:
				c = "😬"
			case maze.Mazog, maze.Mazog2:
				c = "❌"
			case maze.Treasure, maze.Treasure2:
				c = "💰"
			case maze.ThisWay:
				c = "**"
			case maze.DeadEnd:
				c = "xx"
			case maze.Exit:
				c = ">>"
			}
			fmt.Printf("%s", c)
		}
		fmt.Printf("\n")
	}
}

func main() {
	seed := time.Now().UnixNano()
	rand.Seed(seed)

	g := game.New()
	g.Initialize(1)
	g.ChooseEntrance(1)

	displayMap(g.Map())
}
