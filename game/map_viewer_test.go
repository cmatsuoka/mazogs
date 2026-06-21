package game

import (
	"testing"

	"github.com/cmatsuoka/mazogs/maze"
)

// TestInsertMazogsAndMarkPosMarksFinalPos checks that insertMazogsAndMarkPos
// marks the player's final position (playerFinalPos) with ExternalWall, not
// the current PlayerPos.
func TestInsertMazogsAndMarkPosMarksFinalPos(t *testing.T) {
	m := maze.New()
	area := m.Map()

	// Set up cells in the maze interior.
	finalPos := 2*maze.MazeColumns + 10
	currentPos := 2*maze.MazeColumns + 20
	liveMazogPos := 2*maze.MazeColumns + 30

	area[finalPos] = maze.Empty
	area[currentPos] = maze.Empty
	area[liveMazogPos] = maze.Empty
	m.PlayerPos = currentPos

	table := deadMazogTable()
	table[0] = liveMazogPos

	g := &Game{
		maze:           m,
		mazogTable:     table,
		playerFinalPos: finalPos,
	}

	insertMazogsAndMarkPos(g)

	// The player's final position should be marked with ExternalWall.
	if area[finalPos] != maze.ExternalWall {
		t.Fatalf("playerFinalPos (%d) should be ExternalWall: got %#x want %#x",
			finalPos, area[finalPos], maze.ExternalWall)
	}

	// The current PlayerPos must NOT be marked (it's a different cell).
	if area[currentPos] != maze.Empty {
		t.Fatalf("current PlayerPos (%d) should remain Empty: got %#x want %#x",
			currentPos, area[currentPos], maze.Empty)
	}

	// The live mazog should be restored.
	if area[liveMazogPos] != maze.Mazog {
		t.Fatalf("live mazog should be restored: got %#x want %#x",
			area[liveMazogPos], maze.Mazog)
	}
}

// TestInsertMazogsAndMarkPosWithSamePos checks the common case where
// PlayerPos and playerFinalPos happen to be the same cell.
func TestInsertMazogsAndMarkPosWithSamePos(t *testing.T) {
	m := maze.New()
	area := m.Map()

	pos := 2*maze.MazeColumns + 10
	area[pos] = maze.Trail
	m.PlayerPos = pos

	g := &Game{
		maze:           m,
		mazogTable:     deadMazogTable(),
		playerFinalPos: pos,
	}

	insertMazogsAndMarkPos(g)

	if area[pos] != maze.ExternalWall {
		t.Fatalf("position should be ExternalWall: got %#x want %#x",
			area[pos], maze.ExternalWall)
	}
}

// TestInsertMazogsAndMarkPosOverwritesTrail checks that the ExternalWall
// marker overwrites whatever was at playerFinalPos.
func TestInsertMazogsAndMarkPosOverwritesTrail(t *testing.T) {
	m := maze.New()
	area := m.Map()

	trailPos := 2*maze.MazeColumns + 10
	thisWayPos := 2*maze.MazeColumns + 20

	area[trailPos] = maze.Trail
	area[thisWayPos] = maze.ThisWay
	m.PlayerPos = thisWayPos

	g := &Game{
		maze:           m,
		mazogTable:     deadMazogTable(),
		playerFinalPos: trailPos,
	}

	insertMazogsAndMarkPos(g)

	if area[trailPos] != maze.ExternalWall {
		t.Fatalf("trail at playerFinalPos should be overwritten: got %#x want %#x",
			area[trailPos], maze.ExternalWall)
	}

	// thisWayPos should be untouched (not mazog, not cleared).
	if area[thisWayPos] != maze.ThisWay {
		t.Fatalf("thisWayPos should be untouched: got %#x want %#x",
			area[thisWayPos], maze.ThisWay)
	}
}

// TestRemoveRouteClearsMazeAndRestoresPlayerPos checks that removeRoute
// clears maze markers and restores the original player position.
func TestRemoveRouteClearsMazeAndRestoresPlayerPos(t *testing.T) {
	m := maze.New()
	area := m.Map()

	// Set up maze with trail, thisway, and mazog markers.
	trailPos := 2*maze.MazeColumns + 10
	thisWayPos := 2*maze.MazeColumns + 11
	mazogPos := 2*maze.MazeColumns + 12

	area[trailPos] = maze.Trail
	area[thisWayPos] = maze.ThisWay
	area[mazogPos] = maze.Mazog
	m.PlayerPos = trailPos

	origPlayerPos := m.PlayerPos

	g := &Game{
		maze:       m,
		mazogTable: deadMazogTable(),
	}

	removeRoute(g)

	// Trail, ThisWay, and Mazog should be cleared to Empty.
	if area[trailPos] != maze.Empty {
		t.Fatalf("trail should be cleared: got %#x want %#x",
			area[trailPos], maze.Empty)
	}
	if area[thisWayPos] != maze.Empty {
		t.Fatalf("thisway should be cleared: got %#x want %#x",
			area[thisWayPos], maze.Empty)
	}
	if area[mazogPos] != maze.Empty {
		t.Fatalf("mazog should be cleared: got %#x want %#x",
			area[mazogPos], maze.Empty)
	}

	// PlayerPos should be restored to its original value.
	if m.PlayerPos != origPlayerPos {
		t.Fatalf("PlayerPos not restored: got %d want %d", m.PlayerPos, origPlayerPos)
	}
}

// TestRemoveRoutePreservesExitPosCell checks that ClearMaze preserves the
// code at the exit position (the temporary PlayerPos during the operation).
func TestRemoveRoutePreservesExitPosCell(t *testing.T) {
	m := maze.New()
	area := m.Map()

	// Position 0 is the default ExitPos (uninitialized int = 0).
	// ClearMaze saves and restores the code at the current PlayerPos
	// (which is set to ExitPos during removeRoute).
	area[0] = maze.ExternalWall

	testPos := 2*maze.MazeColumns + 10
	area[testPos] = maze.Mazog2 // a code that ClearMaze WILL clear
	m.PlayerPos = testPos

	g := &Game{
		maze:       m,
		mazogTable: deadMazogTable(),
	}

	removeRoute(g)

	// PlayerPos restored.
	if m.PlayerPos != testPos {
		t.Fatalf("PlayerPos not restored: got %d want %d", m.PlayerPos, testPos)
	}

	// Mazog2 at testPos should be cleared.
	if area[testPos] != maze.Empty {
		t.Fatalf("mazog2 cell should be cleared: got %#x want %#x",
			area[testPos], maze.Empty)
	}

	// ExitPos cell (0) should be preserved through ClearMaze.
	if area[0] != maze.ExternalWall {
		t.Fatalf("exitPos cell should be preserved: got %#x want %#x",
			area[0], maze.ExternalWall)
	}
}
