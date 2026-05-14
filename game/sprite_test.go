package game

import (
	"testing"

	"github.com/cmatsuoka/mazogs/maze"
)

// TestAdvanceAnimationTogglesAnimatedCodes checks that each animated maze code
// in the viewport toggles to its paired frame and non-animated codes stay put.
func TestAdvanceAnimationTogglesAnimatedCodes(t *testing.T) {
	m := maze.New()
	m.PlayerPos = 10*maze.MazeColumns + 10
	area := m.Map()
	start := m.PlayerPos - 2*maze.MazeColumns - 2

	tests := []struct {
		name   string
		offset int
		from   byte
		to     byte
	}{
		{"prisoner->prisoner2", 0*maze.MazeColumns + 0, maze.Prisoner, maze.Prisoner2},
		{"prisoner2->prisoner", 0*maze.MazeColumns + 1, maze.Prisoner2, maze.Prisoner},
		{"treasure->treasure2", 1*maze.MazeColumns + 0, maze.Treasure, maze.Treasure2},
		{"treasure2->treasure", 1*maze.MazeColumns + 1, maze.Treasure2, maze.Treasure},
		{"mazog->mazog2", 2*maze.MazeColumns + 0, maze.Mazog, maze.Mazog2},
		{"mazog2->mazog", 2*maze.MazeColumns + 1, maze.Mazog2, maze.Mazog},
	}

	for _, tc := range tests {
		area[start+tc.offset] = tc.from
	}

	staticIndex := start + 3*maze.MazeColumns + 4
	area[staticIndex] = maze.InternalWall

	advanceAnimation(m)

	for _, tc := range tests {
		index := start + tc.offset
		if area[index] != tc.to {
			t.Fatalf("%s: got %#x want %#x", tc.name, area[index], tc.to)
		}
	}
	if area[staticIndex] != maze.InternalWall {
		t.Fatalf("non-animated code changed: got %#x want %#x", area[staticIndex], maze.InternalWall)
	}
}

// TestAdvanceAnimationOnlyTouchesViewport checks that animation updates are
// limited to the 5x4 viewport around the player.
func TestAdvanceAnimationOnlyTouchesViewport(t *testing.T) {
	m := maze.New()
	m.PlayerPos = 10*maze.MazeColumns + 10
	area := m.Map()
	start := m.PlayerPos - 2*maze.MazeColumns - 2

	inside := start
	outsideLeft := start - 1
	outsideBelow := start + 4*maze.MazeColumns

	area[inside] = maze.Prisoner
	area[outsideLeft] = maze.Treasure
	area[outsideBelow] = maze.Mazog

	advanceAnimation(m)

	if area[inside] != maze.Prisoner2 {
		t.Fatalf("inside viewport did not animate: got %#x want %#x", area[inside], maze.Prisoner2)
	}
	if area[outsideLeft] != maze.Treasure {
		t.Fatalf("left of viewport should not change: got %#x want %#x", area[outsideLeft], maze.Treasure)
	}
	if area[outsideBelow] != maze.Mazog {
		t.Fatalf("below viewport should not change: got %#x want %#x", area[outsideBelow], maze.Mazog)
	}
}

// TestAdvanceAnimationCodesHaveSprites checks that every animated code used by
// advanceAnimation has a sprite definition.
func TestAdvanceAnimationCodesHaveSprites(t *testing.T) {
	for _, code := range []byte{
		maze.Prisoner, maze.Prisoner2,
		maze.Treasure, maze.Treasure2,
		maze.Mazog, maze.Mazog2,
	} {
		if _, ok := sprites[code]; !ok {
			t.Fatalf("missing sprite for animated maze code %#x", code)
		}
	}
}
