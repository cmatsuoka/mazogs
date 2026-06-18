package game

import (
	"time"

	"github.com/cmatsuoka/mazogs/graphics"
	"github.com/cmatsuoka/mazogs/maze"
)

// displayMazeWindow renders a 16x16 cell window of raw maze codes starting at
// startPos, with the top-left corner placed at screen position (row=4, col=8).
// Cells outside the maze bounds are shown as InternalWall.
func displayMazeWindow(area []byte, startPos int) {
	const (
		viewRows  = 16
		viewCols  = 16
		screenRow = 4
		screenCol = 8
	)

	mazeSize := len(area)
	for r := range viewRows {
		for c := range viewCols {
			pos := startPos + r*maze.MazeColumns + c
			var code byte
			if pos < 0 || pos >= mazeSize {
				code = maze.InternalWall
			} else {
				code = area[pos]
			}
			graphics.PutZXChar(screenRow+r, screenCol+c, code)
		}
	}
}

// showMapViewer displays the post-game map explorer. The player can scroll the
// maze, toggle the route, and exit with Q. Mirrors BASIC 1000-1266.
func showMapViewer(g *Game) {
	fillScreen(zxInvChequerboard)

	// BASIC 1000-1002: mark relocated treasure with inverse asterisk.
	area := g.maze.Map()
	treasurePos := g.maze.TreasurePos()
	if area[treasurePos] == maze.Treasure || area[treasurePos] == maze.Treasure2 {
		area[treasurePos] = 0x97 // inverse asterisk
	}

	// BASIC 1006-1010: start at maze top-left.
	// BASIC 1041: clamp scroll to [MST, MST+1568]; 1568 = 24*64+32 keeps the
	// 16x16 viewport from wrapping across row boundaries.
	const maxScrollPos = 1568
	scrollPos := 0

	for {
		key := graphics.InKey()

		switch key {
		case "j", "Right":
			// BASIC 1045-1047: scroll right one column.
			scrollPos++
		case "h", "Left":
			// BASIC 1048-1050: scroll left one column.
			scrollPos--
		case "s", "Down":
			// BASIC 1051: scroll down one row.
			scrollPos += maze.MazeColumns
		case "w", "Up":
			// BASIC 1052: scroll up one row.
			scrollPos -= maze.MazeColumns
		case "p":
			// BASIC 1052 / 1100-1116: show route to treasure.
			showRoute(g)
			insertMazogsAndMarkPos(g)
			waitKeyRelease()
		case "0":
			// BASIC 1054 / 1200-1212: remove route.
			removeRoute(g)
			waitKeyRelease()
		case "g", "q":
			// BASIC 1053: quit map viewer.
			return
		}

		// BASIC 1055-1057: clamp scroll range.
		if scrollPos < 0 {
			scrollPos = 0
		} else if scrollPos > maxScrollPos {
			scrollPos = maxScrollPos
		}

		// Only render after a key action to match original behavior.
		if key != "" {
			fillScreen(zxInvChequerboard)
			displayMazeWindow(area, scrollPos)
			graphics.Present()
		}

		// Wait for next input.
		stepDelay()
	}
}

// showRoute resets the maze to show the path from entrance to treasure.
// Mirrors BASIC 1100-1116.
func showRoute(g *Game) {
	graphics.PrintAt(0, 8, "#####SOLVING####")
	graphics.Present()

	// Insert treasure at its original location.
	area := g.maze.Map()
	area[g.maze.TreasurePos()] = maze.Treasure

	// Reset player position to entrance.
	origPlayerPos := g.maze.PlayerPos
	g.maze.PlayerPos = g.maze.ExitPos()

	// Clear maze of mazogs, trails, and ThisWay codes.
	g.maze.ClearMaze()
	g.maze.TraceRoute()

	// Restore player position.
	g.maze.PlayerPos = origPlayerPos
}

// removeRoute clears the maze of route markers. Mirrors BASIC 1200-1206.
func removeRoute(g *Game) {
	// Reset player position to entrance for the clear operation.
	origPlayerPos := g.maze.PlayerPos
	g.maze.PlayerPos = g.maze.ExitPos()
	g.maze.ClearMaze()
	g.maze.PlayerPos = origPlayerPos
}

// insertMazogsAndMarkPos re-inserts all mazogs and marks the player's final
// position with a chequerboard character. Mirrors BASIC 1210-1212.
func insertMazogsAndMarkPos(g *Game) {
	area := g.maze.Map()
	g.maze.InsertMazogs(g.mazogTable)
	area[g.maze.PlayerPos] = maze.ExternalWall
}

// waitKeyRelease polls until the currently held key is released, pumping SDL
// events so KEYUP is processed promptly.
func waitKeyRelease() {
	for graphics.InKey() != "" {
		graphics.ProcessEvents()
		time.Sleep(idlePollMs * time.Millisecond)
	}
}
