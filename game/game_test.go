package game

import (
	"testing"

	"github.com/cmatsuoka/mazogs/maze"
)

// TestApplyMoveValues checks that movesKill, movesView, and movesRemaining are
// computed correctly for each level and that the movesKill cap is enforced.
// BASIC references: 6444 (movesKill = INT(moves/5*level)), 6448 (movesView = 30/level).
func TestApplyMoveValues(t *testing.T) {
	tests := []struct {
		name              string
		level             int
		moves             int
		wantMovesKill     int
		wantMovesView     int
		wantMovesRemaining int
	}{
		// Level 1 (TRY IT OUT): movesKill = moves/5*1, movesView = 30/1 = 30.
		{name: "easy typical", level: levelEasy, moves: 150,
			wantMovesKill: 30, wantMovesView: 30, wantMovesRemaining: 610},

		// Level 2 (FACE A CHALLENGE): movesKill = moves/5*2, movesView = 30/2 = 15.
		{name: "medium typical", level: levelMedium, moves: 150,
			wantMovesKill: 60, wantMovesView: 15, wantMovesRemaining: 610},

		// Level 3 (MANIAC): movesKill = moves/5*3, movesView = 30/3 = 10.
		{name: "hard typical", level: levelHard, moves: 150,
			wantMovesKill: 90, wantMovesView: 10, wantMovesRemaining: 610},

		// Integer division: moves not divisible by 5 is truncated before multiplying.
		// 152/5 = 30 (integer), 30*3 = 90.
		{name: "hard truncated division", level: levelHard, moves: 152,
			wantMovesKill: 90, wantMovesView: 10, wantMovesRemaining: 618},

		// Cap: movesKill must not exceed 255.
		// 430/5 = 86, 86*3 = 258 -> capped to 255.
		{name: "hard cap at 255", level: levelHard, moves: 430,
			wantMovesKill: 255, wantMovesView: 10, wantMovesRemaining: 1730},

		// Cap boundary: value exactly 255 should not be capped.
		// 425/5 = 85, 85*3 = 255 -> no cap needed.
		{name: "hard cap boundary exact", level: levelHard, moves: 425,
			wantMovesKill: 255, wantMovesView: 10, wantMovesRemaining: 1710},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Game{level: tt.level}
			applyMoveValues(g, tt.moves)

			if g.movesKill != tt.wantMovesKill {
				t.Errorf("movesKill = %d, want %d", g.movesKill, tt.wantMovesKill)
			}
			if g.movesView != tt.wantMovesView {
				t.Errorf("movesView = %d, want %d", g.movesView, tt.wantMovesView)
			}
			if g.movesRemaining != tt.wantMovesRemaining {
				t.Errorf("movesRemaining = %d, want %d", g.movesRemaining, tt.wantMovesRemaining)
			}
		})
	}
}

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
