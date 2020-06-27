package mazogs

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	space = byte(iota)
	internalWall
	externalWall
	sword
	thePlayer
)

const (
	mazeRows      = 48
	mazeColumns   = 64
	startPosition = 0x1d6
	entrancePos   = 0x4ff
)

type Maze struct {
	area      []byte
	genTime   time.Time
	playerPos int
}

func NewMaze() *Maze {
	area := make([]byte, mazeRows*mazeColumns)
	for i := range area {
		wall := internalWall
		if i < mazeColumns || i > (mazeRows-1)*mazeColumns {
			// top and bottom rows
			wall = externalWall
		} else {
			c := i % mazeColumns
			if c == 0 || c == mazeColumns-1 {
				// left and right columns
				wall = externalWall
			}
		}
		area[i] = wall
	}
	return &Maze{
		area: area,
	}
}

// Generate generates the maze. The maze must already have been filled with internal walls,
// surrounded by an external wall. The routine creates a series of paths, with the initial
// path starting from the maze entrance. A direction is selected at random and an attempt
// made to progress the path in that direction. If this succeeds then a new random direction
// is selected and an attempt made to progress the path in the new direction. If the path
// could not be progressed in the selected direction then the other possible directions are
// checked in a fixed circular sequence of left-right-down-up until a free direction is found.
// The path is repeatedly progressed in this fashion until it is not possible to progress it
// further in any direction. When this occurs, a random location within the maze is selected
// to form the starting point of a new path and the process then continues from here. A path
// will be progressed in the selected direction unless it would cause it to intersect with
// an existing path. The routine will continue to attempt to route paths until a timeout
// expires.

func (m *Maze) Generate(genTimeout time.Duration) {
	m.genTime = time.Now()
	pos := startPosition

	for {
		direction := rand.Intn(4)
		switch direction {
		case 0:
			if m.canGoLeft(&pos) {
				continue
			}
			fallthrough
		case 1:
			if m.canGoRight(&pos) {
				continue
			}
			fallthrough
		case 2:
			if m.canGoUp(&pos) {
				continue
			}
			fallthrough
		case 3:
			if m.canGoDown(&pos) {
				continue
			}
		}
		if m.canGoLeft(&pos) {
			continue
		}
		if m.canGoRight(&pos) {
			continue
		}
		if m.canGoUp(&pos) {
			continue
		}

		// This point is reached when it is no longer possible to progress the path
		// in any direction, select a random location for the start of a new path.
		var timeout bool
		pos, timeout = m.newStartPosition(genTimeout)
		if timeout {
			break
		}
	}
}

func (m *Maze) newStartPosition(genTimeout time.Duration) (pos int, timeout bool) {
	for {
		if time.Since(m.genTime) > genTimeout {
			return 0, true
		}
		// The time to generate the maze has not yet expired, so select a new
		// random position as the start of the next path.
		pos = mazeColumns + rand.Intn(255)*11
		for i := 0; i < 6; i++ {
			if m.area[pos] == space {
				return pos, false
			}
			// The location is not empty, i.e. it contains an internal
			// wall. So check for an empty location within the 6 positions
			// to the right.If an empty location is found then this is
			// used as the starting position for a new path.
			pos++
		}
		// An empty location was not found near the randomly selected location
		// so jump back to select a new random location to try.
	}
}

func (m *Maze) canGoLeft(pos *int) bool {
	// Is there an internal wall to the left?
	if m.area[*pos-1] != internalWall {
		return false
	}
	// Is there an internal wall at the next left?
	if m.area[*pos-2] != internalWall {
		return false
	}
	// Is there an internal wall left-below?
	if m.area[*pos+mazeColumns-1] != internalWall {
		return false
	}
	// Is there an internal wall left-above?
	if m.area[*pos-mazeColumns-1] != internalWall {
		return false
	}
	// There is an internal wall above, below and the two positions to the left,
	// i.e. progressing left will not touch another pathway.
	*pos--
	m.area[*pos] = space
	return true
}

func (m *Maze) canGoRight(pos *int) bool {
	// Is there an internal wall to the right?
	if m.area[*pos+1] != internalWall {
		return false
	}
	// Is there an internal wall at the next right?
	if m.area[*pos+2] != internalWall {
		return false
	}
	// Is there an internal wall right-below?
	if m.area[*pos+mazeColumns+1] != internalWall {
		return false
	}
	// Is there an internal wall right-above?
	if m.area[*pos-mazeColumns+1] != internalWall {
		return false
	}
	// There is an internal wall above, below and the two positions to the right,
	// i.e. progressing right will not touch another pathway.
	*pos++
	m.area[*pos] = space
	return true
}

func (m *Maze) canGoUp(pos *int) bool {
	// Is there an internal wall to the left in the row above?
	if m.area[*pos-mazeColumns-1] != internalWall {
		return false
	}
	// Is there an internal wall in the original column in the row above?
	if m.area[*pos-mazeColumns] != internalWall {
		return false
	}
	// Is there an internal wall to the right in the row above?
	if m.area[*pos-mazeColumns+1] != internalWall {
		return false
	}
	// Is there an internal wall in the original column in the next row above?
	if m.area[*pos-2*mazeColumns] != internalWall {
		return false
	}
	// There is an internal wall above-left, above-right, immediately above for
	// the next two rows.
	*pos -= mazeColumns
	m.area[*pos] = space
	return true
}

func (m *Maze) canGoDown(pos *int) bool {
	// Is there an internal wall to the left in the row below?
	if m.area[*pos+mazeColumns-1] != internalWall {
		return false
	}
	// Is there an internal wall in the original column in the row below?
	if m.area[*pos+mazeColumns] != internalWall {
		return false
	}
	// Is there an internal wall to the right in the row below?
	if m.area[*pos+mazeColumns+1] != internalWall {
		return false
	}
	// Is there an internal wall in the original column in the next row below?
	if m.area[*pos+2*mazeColumns] != internalWall {
		return false
	}
	// There is an internal wall above-left, above-right, immediately above for
	// the next two rows.
	*pos += mazeColumns
	m.area[*pos] = space
	return true
}

// InsertEntrance creates the entrance passageways. The entrance/exit is located at
// row 19,63 (entering left) or 20,0 (entering right).A sword is always placed
// immediately above the player. An internal wall is placed left and right of the
// sword.
func (m *Maze) InsertEntrance() {
	p := entrancePos
	m.area[p] = internalWall
	p++
	m.area[p] = sword
	p += 2*mazeColumns - 1
	m.area[p] = internalWall
	p++
	m.area[p] = internalWall
	p -= mazeColumns
	startPos := p

	// A passageway will be inserted between the left and right sides of the
	// maze. The top of the passageway contains an internal wall followed by a
	// sword. The bottom of the passageway contains two internal walls. The
	// player is placed in the passageway below the sword.

	// Create a passageway into the maze to the left.
	for {
		m.area[p] = space
		p0 := p
		p--
		if m.area[p] == space {
			break
		}
		// The position to the left is not empty.
		p -= mazeColumns
		if m.area[p] == space {
			// Insert an empty location above to link up with the above-left
			// empty location.
			m.area[p+1] = space
			break
		}
		p += mazeColumns
		if m.area[p] == space {
			// Insert an empty location below to link up with the below-left
			// empty location.
			m.area[p+1] = space
			break
		}
		p = p0 - 1
	}

	p = startPos
	// Create a passageway into the maze to the right.
	for {
		m.area[p] = space
		p0 := p
		p++
		if m.area[p] == space {
			break
		}
		// The position to the right is not empty.
		p -= mazeColumns
		if m.area[p] == space {
			if m.area[p] == space {
				// Insert an empty location above to link up with the above-left
				// empty location.
				m.area[p-1] = space
				break
			}
			p += mazeColumns
			if m.area[p] == space {
				// Insert an empty location below to link up with the below-left
				// empty location.
				m.area[p-1] = space
				break
			}
		}
		p = p0 + 1
	}
	m.area[startPos] = thePlayer
	m.playerPos = startPos
}

func (m *Maze) Display() {
	for i := 0; i < mazeRows; i++ {
		for j := 0; j < mazeColumns; j++ {
			c := "  "
			switch m.area[i*mazeColumns+j] {
			case internalWall:
				c = "██"
			case externalWall:
				c = "▒▒"
			case sword:
				c = "🗡️ "
			case thePlayer:
				c = "🧍"
			}
			fmt.Printf("%s", c)
		}
		fmt.Printf("\n")
	}
}
