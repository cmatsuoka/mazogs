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
			case mazogs.PrisonerEyesOpen, mazogs.PrisonerEyesClosed:
				c = "😬"
			case mazogs.MazogEyesOpen, mazogs.MazogEyesClosed:
				c = "❌"
			}
			fmt.Printf("%s", c)
		}
		fmt.Printf("\n")
	}
}
func main() {
	rand.Seed(time.Now().UnixNano())
	m := mazogs.NewMaze()
	m.Generate(100 * time.Millisecond)
	count := m.CountEmpty()
	m.InsertEntrance()
	m.Populate()
	displayMap(m.Map())
	fmt.Printf("count = %d\n", count)
}
