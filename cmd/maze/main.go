package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cmatsuoka/mazogs"
)

func displayMap(themap []byte) {
	for i := 0; i < mazogs.MazeRows; i++ {
		for j := 0; j < mazogs.MazeColumns; j++ {
			c := "  "
			switch themap[i*mazogs.MazeColumns+j] {
			case mazogs.InternalWall:
				c = "██"
			case mazogs.ExternalWall:
				c = "▒▒"
			case mazogs.Sword:
				c = "🗡️ "
			case mazogs.PlayerStanding:
				c = "🧍"
			case mazogs.Prisoner, mazogs.Prisoner2:
				c = "😬"
			case mazogs.Mazog, mazogs.Mazog2:
				c = "❌"
			case mazogs.Treasure, mazogs.Treasure2:
				c = "💰"
			case mazogs.ThisWay:
				c = "TW"
			}
			fmt.Printf("%s", c)
		}
		fmt.Printf("\n")
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	m := mazogs.NewMaze()
	m.AddTreasure()

	count := func() int {
		for {
			timeout := 512 + rand.Intn(512)
			m.Generate(time.Duration(timeout) * time.Millisecond)
			// Fetch the number of empty locations. Continue if the maze is complex enough.
			count := m.CountEmpty()
			if count >= 1200 {
				return count
			}
		}
	}()

	m.InsertEntrance()
	m.Populate()
	m.ChooseEntrance(1)
	moves := m.Distance()
	displayMap(m.Map())
	fmt.Printf("count = %d, moves = %d\n", count, moves)
}
