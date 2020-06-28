package mazogs

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	// Maze codes
	empty                       byte = 0x00 // whitespace
	externalWall                byte = 0x08 // checkerboard
	deadEnd                     byte = 0x88 // inverse checkerboard
	internalWall                byte = 0x7f // black block
	trail                       byte = 0x1b // period
	exit                        byte = 0x0d // dollar
	thisWay                     byte = 0x17 // asterisk
	playerStanding              byte = 0x1d // 1
	playerRight                 byte = 0x1e // 2
	playerRight2                byte = 0x9e // inverse 2
	playerLeft                  byte = 0x1f // 3
	playerLeft2                 byte = 0x9f // inverse 3
	playerUpDown                byte = 0x20 // 4
	playerUpDown2               byte = 0xa0 // inverse 4
	playerHoldingTreasure       byte = 0x21 // 5
	playerHoldingTreasureRight  byte = 0x22 // 6
	playerHoldingTreasureRight2 byte = 0xa2 // inverse 6
	playerHoldingTreasureLeft   byte = 0x23 // 7
	playerHoldingTreasureLeft2  byte = 0xa3 // inverse 7
	playerHoldingTreasureDown   byte = 0x24 // 8
	playerHoldingTreasureDown2  byte = 0xa4 // inverse 8
	playerHoldingTreasureUp     byte = 0x25 // 9
	playerHoldingTreasureUp2    byte = 0xa5 // inverse 9
	playerHoldingSword          byte = 0x26 // A
	playerHoldingSwordRight     byte = 0x27 // B
	playerHoldingSwordRight2    byte = 0xa7 // inverse B
	playerHoldingSwordLeft      byte = 0x28 // C
	playerHoldingSwordLeft2     byte = 0xa8 // inverse C
	playerHoldingSwordUpDown    byte = 0x29 // D
	playerHoldingSwordUpDown2   byte = 0xa9 // inverse D
	mazogEyesOpen               byte = 0x3d // X
	mazogEyesClosed             byte = 0xbd // inverse X
	theTrasure                  byte = 0x39 // T
	theTrasure2                 byte = 0xb9 // inverse T
	prisonerEyesOpen            byte = 0x35 // P
	prisonerEyesClosed          byte = 0xb5 // inverse P
	sword                       byte = 0xb8 // inverse S
	fighting1                   byte = 0x2b // F
	fighting2                   byte = 0x2c // G
	fighting3                   byte = 0x2d // H
)

const (
	// maze dimensions
	mazeRows    = 48
	mazeColumns = 64

	// indexes inside the maze
	startPosition = 0x1d6
	entrancePos   = 0x4ff

	numMazogs    = 38
	numSwords    = 40
	numPrisoners = 30
)

type Maze struct {
	area       []byte
	genTime    time.Time
	playerPos  int
	mazogTable []int
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
		area:       area,
		mazogTable: make([]int, 40),
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
		pos = mazeColumns + rand.Intn(256)*11
		for i := 0; i < 6; i++ {
			if m.area[pos] == empty {
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
	m.area[*pos] = empty
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
	m.area[*pos] = empty
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
	m.area[*pos] = empty
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
	m.area[*pos] = empty
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
		m.area[p] = empty
		p0 := p
		p--
		if m.area[p] == empty {
			break
		}
		// The position to the left is not empty.
		p -= mazeColumns
		if m.area[p] == empty {
			// Insert an empty location above to link up with the
			// above-left empty location.
			m.area[p+1] = empty
			break
		}
		p += mazeColumns
		if m.area[p] == empty {
			// Insert an empty location below to link up with the
			// below-left empty location.
			m.area[p+1] = empty
			break
		}
		p = p0 - 1
	}

	p = startPos
	// Create a passageway into the maze to the right.
	for {
		m.area[p] = empty
		p0 := p
		p++
		if m.area[p] == empty {
			break
		}
		// The position to the right is not empty.
		p -= mazeColumns
		if m.area[p] == empty {
			if m.area[p] == empty {
				// Insert an empty location above to link up with the
				// above-left empty location.
				m.area[p-1] = empty
				break
			}
			p += mazeColumns
			if m.area[p] == empty {
				// Insert an empty location below to link up with the
				// below-left empty location.
				m.area[p-1] = empty
				break
			}
		}
		p = p0 + 1
	}
	m.area[startPos] = playerStanding
	m.playerPos = startPos
}

// Populate inserts mazogs, swords and prisoners randomly within the maze. Mazogs can
// only be placed at empty locations. Swords and prisoners can only be placed where
// there is an internal wall. No items are placed next to an external wall.
func (m *Maze) Populate() {
	// Insert swords at random locations within the maze.
	for i := 0; i < numSwords; i++ {
		p := m.randomInternalWallPos()
		m.area[p] = sword
	}

	// Insert prisoners at random locations within the maze.
	for i := 0; i < numPrisoners; i++ {
		p := m.randomInternalWallPos()
		m.area[p] = prisonerEyesOpen
	}

	// Determine random locations within the maze for the mazogs. The addresses
	// of the mazogs are placed within the mazogs table rather than codes the
	// mazogs inserted with the maze (this is done afterwards). The routine will
	// only locate a mazog at the position of an empty location within the maze.
	// Note that there is nothing to prevent multiple mazogs being placed at
	// the same location.
	for i := 0; i < numMazogs; i++ {
		p := func() int {
			for {
				p := randomMazePos()
				if m.area[p] == empty && m.area[p] != externalWall && m.area[p+1] != externalWall {
					return p
				}
			}
		}()
		m.mazogTable[i] = p
	}

	m.insertMazogs()
}

// insertMazogs inserts all alive mazogs into the maze. It is used from above to
// insert the mazogs for the first time, and also after clearing the maze when
// calculating the distance to the treasure/exit. It is called repeatedly from the main
// loop and if the mazog codes are already in the maze then they are simply overwritten,
// causing no change. It is also used when examining the maze at the end of the game
// after clearing the maze and then populating the route to the treasure.
func (m *Maze) insertMazogs() {
	for i := 0; i < numMazogs; i++ {
		mp := m.mazogTable[i]
		if mp > 0xff00 {
			// Mazog has been killed.
			continue
		}
		if m.area[mp] == mazogEyesClosed {
			m.area[mp] = mazogEyesOpen
		} else {
			m.area[mp] = mazogEyesClosed
		}
	}
}

// randomInteralWallPos selects a random maze location containing an internal wall. This
// routine is used when selecting the locations to place swords and prisoners within
// the maze. It prevents placing a sword or prisoner next to an external wall.
func (m *Maze) randomInternalWallPos() int {
	for {
		p := randomMazePos()
		if m.area[p] == internalWall && m.area[p-1] != externalWall && m.area[p+1] != externalWall {
			return p
		}
	}
}

// randomMazePos selects a random maze location. This routine selects a random location
// within the maze by choosing a random 8-bit number, multiplying it by 11, and offsetting
// into the maze from the first row that can contain empty location (i.e. the 3rd row).
// This means that not all positions within the maze can be returned, but they are
// uniformly distributed throughout the maze.
//
// The multiplication by 11 distributes the random number throughout the maze. The maze
// is 48 rows by 64 columns = 3072 positions. The top and bottom rows are external walls,
// and the next inner rows are always internal walls. Assuming the maze coordinates are
// 0 index based then the rows spans 0 to 47, and the first row that can contains game
// characters is row 2 and the last is row 45. The distribution range is:
//   0   x 11 = 0               => (0, 0)
//   255 x 11 = 2805 = 43 re 53 => (45, 53)
func randomMazePos() int {
	return rand.Intn(256)*11 + 2*mazeColumns
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
			case playerStanding:
				c = "🧍"
			case prisonerEyesOpen, prisonerEyesClosed:
				c = "😬"
			case mazogEyesOpen, mazogEyesClosed:
				c = "❌"
			}
			fmt.Printf("%s", c)
		}
		fmt.Printf("\n")
	}
}
