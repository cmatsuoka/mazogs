package game

import (
	"testing"

	"github.com/cmatsuoka/mazogs/maze"
)

func deadMazogTable() []int {
	table := make([]int, 38)
	for i := range table {
		table[i] = 0xffff
	}
	return table
}

// TestDecrementTimerViewModeUnderflowDoesNotDeduct checks that view mode skips
// deduction when the view cost would underflow the timer.
func TestDecrementTimerViewModeUnderflowDoesNotDeduct(t *testing.T) {
	g := &Game{
		hasCountdown:   true,
		viewMode:       true,
		movesRemaining: 10,
		movesView:      10,
	}

	decrementTimer(g)

	if g.movesRemaining != 10 {
		t.Fatalf("movesRemaining changed on underflow check: got %d want %d", g.movesRemaining, 10)
	}
	if g.starved {
		t.Fatalf("starved should remain false during view underflow check")
	}
}

// TestDecrementTimerViewModeDeductsWhenAffordable checks that view mode
// subtracts the configured view cost when there is enough time remaining.
func TestDecrementTimerViewModeDeductsWhenAffordable(t *testing.T) {
	g := &Game{
		hasCountdown:   true,
		viewMode:       true,
		movesRemaining: 12,
		movesView:      10,
	}

	decrementTimer(g)

	if g.movesRemaining != 2 {
		t.Fatalf("unexpected view deduction: got %d want %d", g.movesRemaining, 2)
	}
}

// TestDecrementTimerViewModeExactBoundary checks that view mode allows
// deduction when the remaining moves hit the exact zero boundary.
func TestDecrementTimerViewModeExactBoundary(t *testing.T) {
	g := &Game{
		hasCountdown:   true,
		viewMode:       true,
		movesRemaining: 11,
		movesView:      10,
	}

	decrementTimer(g)

	if g.movesRemaining != 1 {
		t.Fatalf("boundary case should allow deduction: got %d want %d", g.movesRemaining, 1)
	}
}

// TestDecrementTimerStarvationOnNextTickAtZero checks that starvation is only
// set on a tick after the timer has already reached zero.
func TestDecrementTimerStarvationOnNextTickAtZero(t *testing.T) {
	g := &Game{
		hasCountdown:   true,
		movesRemaining: 1,
	}

	decrementTimer(g)
	if g.movesRemaining != 0 {
		t.Fatalf("first tick should decrement to zero: got %d want 0", g.movesRemaining)
	}
	if g.starved {
		t.Fatalf("starved should still be false immediately after reaching zero")
	}

	decrementTimer(g)
	if !g.starved {
		t.Fatalf("starved should be set when timer is already zero")
	}
	if g.movesRemaining != 0 {
		t.Fatalf("movesRemaining should stay at zero after starvation: got %d", g.movesRemaining)
	}
}

// TestMoveAllMazogsRestoresWhenMovementDisabled checks that disabled movement
// still restores live mazogs to the maze without changing their positions.
func TestMoveAllMazogsRestoresWhenMovementDisabled(t *testing.T) {
	m := maze.New()
	pos := 2*maze.MazeColumns + 10
	m.Map()[pos] = maze.Empty

	table := deadMazogTable()
	table[0] = pos

	g := &Game{
		maze:       m,
		mazogTable: table,
		mazogsMove: false, // TRY IT OUT mode
	}

	moveAllMazogs(g)

	if g.maze.Map()[pos] != maze.Mazog {
		t.Fatalf("expected mazog to be restored even when movement is disabled")
	}
	if g.mazogTable[0] != pos {
		t.Fatalf("mazog position should not change when movement is disabled: got %d want %d", g.mazogTable[0], pos)
	}
}
