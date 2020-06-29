package main

import (
	"fmt"
	"math/rand"
	"time"

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

	m := maze.New()

	count := func() int {
		var count int
		for i := 0; i < 10; i++ {
			timeout := 512 + rand.Intn(512)
			m.Generate(time.Duration(timeout) * time.Millisecond)
			// Fetch the number of empty locations. Continue if the maze is complex enough.
			count = m.CountEmpty()
			if count >= 1200 {
				return count
			}
			fmt.Println("Maze is not complex enough, try again.")
		}
		return count
	}()

	m.InsertEntrance()
	m.Populate()
	m.ChooseEntrance(1)

	var moves int
	for {
		moves = m.Distance()
		if moves > 120 {
			// The treasure is sufficiently far away.
			break
		}

		// The treasure is not far enough away.
		fmt.Printf("Treasure is too close (%d moves), relocating...\n", moves)
		m.RelocateTreasure()
	}
	displayMap(m.Map())
	fmt.Printf("count = %d, moves = %d\n", count, moves)
}
