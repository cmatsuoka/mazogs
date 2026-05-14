package maze

import (
	"math/rand"
	"testing"
)

func deadMazogTable() []int {
	table := make([]int, numMazogs)
	for i := range table {
		table[i] = 0xffff
	}
	return table
}

func TestPopulateMazogsNotAdjacentToExternalWall(t *testing.T) {
	// Regression for mazog placement: candidates must be empty and not
	// adjacent to an external wall on either side (assembly L4EC9).
	for seed := int64(0); seed < 64; seed++ {
		rand.Seed(seed)

		m := New()
		m.Generate()
		m.InsertEntrance()
		mazogs := m.Populate()

		if len(mazogs) != numMazogs {
			t.Fatalf("seed=%d: got %d mazogs, want %d", seed, len(mazogs), numMazogs)
		}

		for i, pos := range mazogs {
			if pos <= 0 || pos >= len(m.area)-1 {
				t.Fatalf("seed=%d: mazog[%d] out of bounds at %d", seed, i, pos)
			}
			if m.area[pos-1] == ExternalWall || m.area[pos+1] == ExternalWall {
				t.Fatalf("seed=%d: mazog[%d] at %d is adjacent to external wall", seed, i, pos)
			}
		}
	}
}

func TestInsertMazogsSkipsDeadEntries(t *testing.T) {
	m := New()
	constructMazeArea(m)

	livePos := 2*MazeColumns + 10
	deadPos := 2*MazeColumns + 20

	m.area[livePos] = Empty
	m.area[deadPos] = Trail

	table := deadMazogTable()
	table[0] = livePos
	table[1] = 0xffff

	m.InsertMazogs(table)

	if m.area[livePos] != Mazog {
		t.Fatalf("live mazog was not restored: got %#x want %#x", m.area[livePos], Mazog)
	}
	if m.area[deadPos] != Trail {
		t.Fatalf("dead mazog entry should not modify cell: got %#x want %#x", m.area[deadPos], Trail)
	}
}

func TestInsertMazogsPreservesMazog2(t *testing.T) {
	m := New()
	constructMazeArea(m)

	pos := 2*MazeColumns + 30
	m.area[pos] = Mazog2

	table := deadMazogTable()
	table[0] = pos

	m.InsertMazogs(table)

	if m.area[pos] != Mazog2 {
		t.Fatalf("expected mazog2 to be preserved, got %#x want %#x", m.area[pos], Mazog2)
	}
}
