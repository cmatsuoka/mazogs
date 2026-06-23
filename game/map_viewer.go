package game

import (
	"github.com/cmatsuoka/mazogs/graphics"
	"github.com/cmatsuoka/mazogs/maze"
)

// showMapViewer displays the post-game maze in full screen.
// Mirrors BASIC 3200-4534.
func showMapViewer(g *Game) {
	area := g.maze.Map()
	scrollRow, scrollCol := 0, 0

	// Render initial frame before entering the input loop.
	fillScreen(zxInvChequerboard)
	for dr := 0; dr < 24; dr++ {
		for dc := 0; dc < 32; dc++ {
			pos := (scrollRow+dr)*maze.MazeColumns + (scrollCol + dc)
			if pos >= 0 && pos < len(area) {
				graphics.PutZXChar(dr, dc, area[pos])
			}
		}
	}
	graphics.Present()

	for {
		if graphics.QuitRequested() {
			return
		}
		key := graphics.InKey()
		moved := false
		switch key {
		case "d", "j", "Right":
			if scrollCol < 32 {
				scrollCol++
				moved = true
			}
		case "a", "h", "Left":
			if scrollCol > 0 {
				scrollCol--
				moved = true
			}
		case "x", "s", "Down":
			if scrollRow < 24 {
				scrollRow++
				moved = true
			}
		case "w", "Up":
			if scrollRow > 0 {
				scrollRow--
				moved = true
			}
		case "p":
			scrollRow, scrollCol = 0, 0
			showRoute(g)
			insertMazogsAndMarkPos(g)
			waitKeyRelease()
			moved = true
		case "0":
			scrollRow, scrollCol = 0, 0
			removeRoute(g)
			insertMazogsAndMarkPos(g)
			waitKeyRelease()
			moved = true
		case "g", "G", "Escape":
			return
		}

		if moved {
			fillScreen(zxInvChequerboard)
			for dr := 0; dr < 24; dr++ {
				for dc := 0; dc < 32; dc++ {
					pos := (scrollRow+dr)*maze.MazeColumns + (scrollCol + dc)
					if pos >= 0 && pos < len(area) {
						graphics.PutZXChar(dr, dc, area[pos])
					}
				}
			}
			graphics.Present()
			if key != "p" && key != "0" {
				stepDelay()
			}
		} else {
			poll()
		}
	}
}

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

// showRoute resets the maze to show the path from entrance to treasure.
// Mirrors BASIC 1100-1116.
func showRoute(g *Game) {
	graphics.PrintAt(0, 8, "_____SOLVING____")
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
	area[g.playerFinalPos] = maze.ExternalWall
}

// waitKeyRelease polls until the currently held key is released, pumping SDL
// events so KEYUP is processed promptly.
func waitKeyRelease() {
	for graphics.InKey() != "" {
		poll()
	}
}
