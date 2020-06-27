package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cmatsuoka/mazogs"
)

func main() {
	fmt.Println("mazogs")
	rand.Seed(time.Now().UnixNano())
	m := mazogs.NewMaze()
	m.Generate(100 * time.Millisecond)
	m.InsertEntrance()
	m.Populate()
	m.Display()
}
