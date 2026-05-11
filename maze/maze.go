package maze

import (
	"math/rand"
)

const (
	// Maze codes
	Empty                    byte = 0x00 // whitespace
	ExternalWall             byte = 0x08 // checkerboard
	DeadEnd                  byte = 0x88 // inverse checkerboard
	InternalWall             byte = 0x80 // black block
	Trail                    byte = 0x1b // period
	Exit                     byte = 0x0d // dollar
	ThisWay                  byte = 0x17 // asterisk
	PlayerStanding           byte = 0x1d // 1
	PlayerRight              byte = 0x1e // 2
	PlayerRight2             byte = 0x9e // inverse 2
	PlayerLeft               byte = 0x1f // 3
	PlayerLeft2              byte = 0x9f // inverse 3
	PlayerUpDown             byte = 0x20 // 4
	PlayerUpDown2            byte = 0xa0 // inverse 4
	PlayerWithTreasure       byte = 0x21 // 5
	PlayerWithTreasureRight  byte = 0x22 // 6
	PlayerWithTreasureRight2 byte = 0xa2 // inverse 6
	PlayerWithTreasureLeft   byte = 0x23 // 7
	PlayerWithTreasureLeft2  byte = 0xa3 // inverse 7
	PlayerWithTreasureDown   byte = 0x24 // 8
	PlayerWithTreasureDown2  byte = 0xa4 // inverse 8
	PlayerWithTreasureUp     byte = 0x25 // 9
	PlayerWithTreasureUp2    byte = 0xa5 // inverse 9
	PlayerWithSword          byte = 0x26 // A
	PlayerWithSwordRight     byte = 0x27 // B
	PlayerWithSwordRight2    byte = 0xa7 // inverse B
	PlayerWithSwordLeft      byte = 0x28 // C
	PlayerWithSwordLeft2     byte = 0xa8 // inverse C
	PlayerWithSwordUpDown    byte = 0x29 // D
	PlayerWithSwordUpDown2   byte = 0xa9 // inverse D
	Mazog                    byte = 0x3d // X
	Mazog2                   byte = 0xbd // inverse X
	Treasure                 byte = 0x39 // T
	Treasure2                byte = 0xb9 // inverse T
	Prisoner                 byte = 0x35 // P
	Prisoner2                byte = 0xb5 // inverse P
	Sword                    byte = 0xb8 // inverse S
	Fighting1                byte = 0x2b // F
	Fighting2                byte = 0x2c // G
	Fighting3                byte = 0x2d // H
)

const (
	// maze dimensions
	MazeRows    = 48
	MazeColumns = 64

	// indexes inside the maze
	entrancePos = 0x4ff

	// amount of items and creatures
	numMazogs    = 38
	numSwords    = 40
	numPrisoners = 30
)

type Maze struct {
	PlayerPos       int
	exitPos         int
	treasurePos     int
	area            []byte
	codeAtPlayerPos byte
}

func New() *Maze {
	return &Maze{
		area: make([]byte, MazeRows*MazeColumns),
	}
}

func (m *Maze) Map() []byte {
	return m.area
}

func (m *Maze) SetPlayerCode(code byte) {
	m.area[m.PlayerPos] = code
}

// ConstructMazeArea sets the border around the maze and fills all internal locations with
// internal wall maze codes.
func constructMazeArea(m *Maze) {
	for i := range m.area {
		wall := InternalWall
		if i < MazeColumns || i > (MazeRows-1)*MazeColumns {
			// top and bottom rows
			wall = ExternalWall
		} else {
			c := i % MazeColumns
			if c == 0 || c == MazeColumns-1 {
				// left and right columns
				wall = ExternalWall
			}
		}
		m.area[i] = wall
	}
}

func (m *Maze) IntroMaze() {
	constructMazeArea(m)
	p := 7*MazeColumns + 22

	// A section of the maze is filled as follows:
	// 5T###
	// ..BX#
	// #*P #
	// x*#s#

	for k, v := range map[int]byte{
		p - 2*MazeColumns - 2: PlayerWithTreasure,
		p - 2*MazeColumns - 1: Treasure,
		p - MazeColumns - 2:   Trail,
		p - MazeColumns - 1:   Trail,
		p - MazeColumns:       PlayerWithSwordRight,
		p - MazeColumns + 1:   Mazog,
		p - 1:                 ThisWay,
		p:                     Prisoner2,
		p + 1:                 Empty,
		p + MazeColumns + 1:   Sword,
		p + MazeColumns - 1:   ThisWay,
		p + MazeColumns - 2:   Mazog2,
	} {
		m.area[k] = v
	}
}

// maxConsecutiveDeadEnds is the number of consecutive dead ends without
// carving a single cell before declaring the maze saturated.
const maxConsecutiveDeadEnds = 512

// maxStartSearchRetries is the maximum number of random candidates tried in
// newStartPosition before giving up (i.e. no empty cells remain).
const maxStartSearchRetries = 1024

// Generate generates the maze. The maze must already have been filled with internal walls,
// surrounded by an external wall. The routine creates a series of paths, with the initial
// path starting from the maze entrance. A direction is selected at random and an attempt
// made to progress the path in that direction. If this succeeds then a new random direction
// is selected and an attempt made to progress the path in the new direction. If the path
// could not be progressed in the selected direction then the other possible directions are
// checked in a fixed circular sequence of left-right-down-up until a free direction is found.
// The path is repeatedly progressed in this fashion until it is not possible to progress it
// further in any direction. When this occurs, a random location within the maze is selected
// to form the starting point of a new path and the process then continues from here.
// Generation stops when maxConsecutiveDeadEnds dead ends occur without any cell being carved,
// which indicates the maze is fully saturated.
func (m *Maze) Generate() {
	constructMazeArea(m)
	addTreasure(m)

	pos := m.PlayerPos
	deadEnds := 0

	for {
		direction := rand.Intn(4)
		switch direction {
		case 0:
			if canGoLeft(m, &pos) {
				deadEnds = 0
				continue
			}
			fallthrough
		case 1:
			if canGoRight(m, &pos) {
				deadEnds = 0
				continue
			}
			fallthrough
		case 2:
			if canGoUp(m, &pos) {
				deadEnds = 0
				continue
			}
			fallthrough
		case 3:
			if canGoDown(m, &pos) {
				deadEnds = 0
				continue
			}
		}
		if canGoLeft(m, &pos) {
			deadEnds = 0
			continue
		}
		if canGoRight(m, &pos) {
			deadEnds = 0
			continue
		}
		if canGoUp(m, &pos) {
			deadEnds = 0
			continue
		}

		// This point is reached when it is no longer possible to progress the path
		// in any direction, select a random location for the start of a new path.
		deadEnds++
		if deadEnds >= maxConsecutiveDeadEnds {
			break
		}
		var found bool
		pos, found = newStartPosition(m)
		if !found {
			break
		}
	}
}

func newStartPosition(m *Maze) (pos int, found bool) {
	for i := 0; i < maxStartSearchRetries; i++ {
		// The time to generate the maze has not yet expired, so select a new
		// random position as the start of the next path.
		p := 2*MazeColumns + rand.Intn(256)*11
		for j := 0; j < 7; j++ {
			if m.area[p] == Empty {
				return p, true
			}
			// The location is not empty, i.e. it contains an internal
			// wall. So check for an empty location within the 6 positions
			// to the right. If an empty location is found then this is
			// used as the starting position for a new path.
			p++
		}
		// An empty location was not found near the randomly selected location
		// so jump back to select a new random location to try.
	}
	return 0, false
}

func canGoLeft(m *Maze, pos *int) bool {
	// Is there an internal wall to the left?
	if m.area[*pos-1] != InternalWall {
		return false
	}
	// Is there an internal wall at the next left?
	if m.area[*pos-2] != InternalWall {
		return false
	}
	// Is there an internal wall left-below?
	if m.area[*pos+MazeColumns-1] != InternalWall {
		return false
	}
	// Is there an internal wall left-above?
	if m.area[*pos-MazeColumns-1] != InternalWall {
		return false
	}
	// There is an internal wall above, below and the two positions to the left,
	// i.e. progressing left will not touch another pathway.
	*pos--
	m.area[*pos] = Empty
	return true
}

func canGoRight(m *Maze, pos *int) bool {
	// Is there an internal wall to the right?
	if m.area[*pos+1] != InternalWall {
		return false
	}
	// Is there an internal wall at the next right?
	if m.area[*pos+2] != InternalWall {
		return false
	}
	// Is there an internal wall right-below?
	if m.area[*pos+MazeColumns+1] != InternalWall {
		return false
	}
	// Is there an internal wall right-above?
	if m.area[*pos-MazeColumns+1] != InternalWall {
		return false
	}
	// There is an internal wall above, below and the two positions to the right,
	// i.e. progressing right will not touch another pathway.
	*pos++
	m.area[*pos] = Empty
	return true
}

func canGoUp(m *Maze, pos *int) bool {
	// Is there an internal wall to the left in the row above?
	if m.area[*pos-MazeColumns-1] != InternalWall {
		return false
	}
	// Is there an internal wall in the original column in the row above?
	if m.area[*pos-MazeColumns] != InternalWall {
		return false
	}
	// Is there an internal wall to the right in the row above?
	if m.area[*pos-MazeColumns+1] != InternalWall {
		return false
	}
	// Is there an internal wall in the original column in the next row above?
	if m.area[*pos-2*MazeColumns] != InternalWall {
		return false
	}
	// There is an internal wall above-left, above-right, immediately above for
	// the next two rows.
	*pos -= MazeColumns
	m.area[*pos] = Empty
	return true
}

func canGoDown(m *Maze, pos *int) bool {
	// Is there an internal wall to the left in the row below?
	if m.area[*pos+MazeColumns-1] != InternalWall {
		return false
	}
	// Is there an internal wall in the original column in the row below?
	if m.area[*pos+MazeColumns] != InternalWall {
		return false
	}
	// Is there an internal wall to the right in the row below?
	if m.area[*pos+MazeColumns+1] != InternalWall {
		return false
	}
	// Is there an internal wall in the original column in the next row below?
	if m.area[*pos+2*MazeColumns] != InternalWall {
		return false
	}
	// There is an internal wall above-left, above-right, immediately above for
	// the next two rows.
	*pos += MazeColumns
	m.area[*pos] = Empty
	return true
}

func (m *Maze) CountEmpty() int {
	return countCode(m, Empty)
}

// countCode determines how many locations there are in the maze with the specified
// code, to check whether there are a sufficient number of movement possibilities.
// The routine is also used to calculate the number of locations to reach the treasure
// or exit. In this case the route is first populated with 'This Way' codes and then
// this function is used to count how many 'This Way' codes there are.
func countCode(m *Maze, what byte) (count int) {
	for _, code := range m.area {
		if code == what {
			count++
		}
	}
	return count
}

// clearMaze clears the maze of mazogs, trails, and 'This Way's
func clearMaze(m *Maze) {
	p := m.PlayerPos
	m.codeAtPlayerPos = m.area[p]
	for i, code := range m.area {
		switch code {
		case Trail, ThisWay, Mazog, Mazog2:
			m.area[i] = Empty
		}
	}
}

// TraceRoute cleverly determines the route to the treasure, or if the player has
// already collected the treasure then it will determine the route to the exit. It
// first searches around the player's location for the treasure #1 maze code. If this
// is not found then it searches for treasure #2 maze code. If this is not found then
// it searches for the exit maze code (which will only ever be in the maze once the
// player has collected the treasure). If the exit is not found then it searches for
// an empty location. If an empty location is found then it is replaced with 'This Way'.
// The search is now repeated from this new location. As each empty location is found
// it is replaced with 'This Way'. This continues until either the treasure / exit is
// found or there is no empty location found, i.e. a dead end was reached. When a dead
// end is reached, a search is now made for 'This Way' and replaced with a special
// 'Dead End' maze code. If an empty location is found adjacent to the 'Dead End' location
// then it is explored. If an empty location is not found then there must be another
// 'This Way' that lead to this location originally and so it is now replaced with 'Dead
// End'. The net result is that all dead ends are back-tracked and labelled as dead ends.
// This leaves only those 'This Way' codes that form part of the actual route to the
// treasure or exit (as appropriate). Once the route to the treasure / exit has been
// established, all 'Dead End' codes are replaced as empty locations to leave just the
// route through the maze in place.
func (m *Maze) TraceRoute() {
	p := m.PlayerPos

	tryNewPos := func() (stopTrying bool) {
		for _, what := range []byte{Treasure, Treasure2, Exit, Empty, ThisWay} {
			newPos, found := checkSurroundings(m, p, what)
			if found {
				switch what {
				case Treasure, Treasure2, Exit:
					return true
				case Empty:
					// An empty location was found to the left, right, above or below.
					m.area[p] = ThisWay
					p = newPos
					return false
				default:
					// 'This Way' was found to the left, right, above or below. The
					// search will always favour an empty location and so if an empty
					// location was not found but a 'This Way' was found then it means
					// that a dead end was reached. The search is then back-tracked
					// along the route that was previously followed. The location is
					// marked as a dead end, which eventually will leave only those
					// 'This Way' entries that form part of the route.
					m.area[p] = DeadEnd
					p = newPos
					return false
				}
			}
		}
		// Was searching for 'This Way' but failed to find it. Must be at the exit.
		return true
	}

	// Enter a loop to search for a specific maze code (treasure #1, treasure #2, exit,
	// empty location or 'This Way') at the current location.
	for {
		if tryNewPos() {
			break
		}
	}
	// Reached the treasure #1, treasure #2 or exit. The exit will only be in the maze after
	// the player has collected the treasure and so will not be found when searching for
	// the treasure.
	m.area[p] = ThisWay
	// Now remove all dead end markers and replace with empty locations. This will leave
	// 'This Way' markers along the route to the treasure or exit.
	p = m.PlayerPos
	m.area[p] = m.codeAtPlayerPos
	for i, code := range m.area {
		if code == DeadEnd {
			m.area[i] = Empty
		}
	}
}

func checkSurroundings(m *Maze, p int, what byte) (newPos int, found bool) {
	for _, newPos := range []int{p - 1, p + 1, p - MazeColumns, p + MazeColumns} {
		if m.area[newPos] == what {
			return newPos, true
		}
	}
	return p, false
}

// InsertEntrance creates the entrance passageways. The entrance/exit is located at
// row 19,63 (entering left) or 20,0 (entering right).A sword is always placed
// immediately above the player. An internal wall is placed left and right of the
// sword.
func (m *Maze) InsertEntrance() {
	p := entrancePos
	m.area[p] = InternalWall
	p++
	m.area[p] = Sword
	p += 2*MazeColumns - 1
	m.area[p] = InternalWall
	p++
	m.area[p] = InternalWall
	p -= MazeColumns
	startPos := p

	// A passageway will be inserted between the left and right sides of the
	// maze. The top of the passageway contains an internal wall followed by a
	// sword. The bottom of the passageway contains two internal walls. The
	// player is placed in the passageway below the sword.

	// Create a passageway into the maze to the left.
	for {
		m.area[p] = Empty
		p0 := p
		p--
		if m.area[p] == Empty {
			break
		}
		// The position to the left is not empty.
		p -= MazeColumns
		if m.area[p] == Empty {
			// Insert an empty location above to link up with the
			// above-left empty location.
			m.area[p+1] = Empty
			break
		}
		// The position above-left is not empty.
		p += 2 * MazeColumns
		if m.area[p] == Empty {
			// Insert an empty location below to link up with the
			// below-left empty location.
			m.area[p+1] = Empty
			break
		}
		// The position below-left is not empty.
		p = p0 - 1
	}

	p = startPos
	// Create a passageway into the maze to the right.
	for {
		m.area[p] = Empty
		p0 := p
		p++
		if m.area[p] == Empty {
			break
		}
		// The position to the right is not empty.
		p -= MazeColumns
		if m.area[p] == Empty {
			// Insert an empty location above to link up with the
			// above-left empty location.
			m.area[p-1] = Empty
			break
		}
		// The position above-right is not empty.
		p += 2 * MazeColumns
		if m.area[p] == Empty {
			// Insert an empty location below to link up with the
			// below-left empty location.
			m.area[p-1] = Empty
			break
		}
		// The position below-right is not empty.
		p = p0 + 1
	}
	m.PlayerPos = startPos
}

// Animate updates the animation for prisoners, treasure and mazog.
func (m *Maze) Animate(p int) {
	switch m.area[p] {
	case Prisoner, Prisoner2, Treasure, Treasure2, Mazog, Mazog2:
		m.area[p] += 0x80
	}
}

func addTreasure(m *Maze) int {
	for {
		// Select a random location between row 2 column 0 and row 45 column 58.
		p := 2*MazeColumns + rand.Intn(44*MazeColumns-6)
		if m.area[p] != ExternalWall && m.area[p-1] != ExternalWall && m.area[p+1] != ExternalWall {
			// Set the player at the location of the treasure
			m.PlayerPos = p
			// Insert the treasure ['T'] into the maze.
			m.area[p] = Treasure
			m.treasurePos = p
			return p
		}
	}
}

func (m *Maze) RelocateTreasure() {
	m.area[m.treasurePos] = Empty
	for {
		// Select a random location between row 2 column 0 and row 45 column 58.
		p := 2*MazeColumns + rand.Intn(44*MazeColumns-6)
		if m.area[p] == Empty {
			m.treasurePos = p
			m.area[p] = Treasure
			break
		}
	}
}

func (m *Maze) ChooseEntrance(dir int) {
	p := m.PlayerPos
	m.exitPos = p
	if dir > 0 {
		// Enter maze from the right
		m.area[p-1] = InternalWall
	} else {
		// Enter maze from the left
		m.area[p+1] = InternalWall
	}
}

func (m *Maze) Distance() int {
	clearMaze(m)
	m.TraceRoute()
	moves := countCode(m, ThisWay)
	clearMaze(m)
	return moves
}

// Populate inserts mazogs, swords and prisoners randomly within the maze. Mazogs can
// only be placed at empty locations. Swords and prisoners can only be placed where
// there is an internal wall. No items are placed next to an external wall.
func (m *Maze) Populate() (mazogTable []int) {
	mazogTable = make([]int, numMazogs)

	// Insert swords at random locations within the maze.
	for i := 0; i < numSwords; i++ {
		p := m.randomInternalWallPos()
		m.area[p] = Sword
	}

	// Insert prisoners at random locations within the maze.
	for i := 0; i < numPrisoners; i++ {
		p := m.randomInternalWallPos()
		m.area[p] = Prisoner
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
				if m.area[p] == Empty && m.area[p] != ExternalWall && m.area[p+1] != ExternalWall {
					return p
				}
			}
		}()
		mazogTable[i] = p
	}

	m.insertMazogs(mazogTable)

	return mazogTable
}

// insertMazogs inserts all alive mazogs into the maze. It is used from above to
// insert the mazogs for the first time, and also after clearing the maze when
// calculating the distance to the treasure/exit. It is called repeatedly from the main
// loop and if the mazog codes are already in the maze then they are simply overwritten,
// causing no change. It is also used when examining the maze at the end of the game
// after clearing the maze and then populating the route to the treasure.
func (m *Maze) insertMazogs(mazogTable []int) {
	for i := 0; i < numMazogs; i++ {
		mp := mazogTable[i]
		if mp > 0xff00 {
			// Mazog has been killed.
			continue
		}
		if m.area[mp] == Mazog2 {
			m.area[mp] = Mazog
		} else {
			m.area[mp] = Mazog2
		}
	}
}

// randomInteralWallPos selects a random maze location containing an internal wall. This
// routine is used when selecting the locations to place swords and prisoners within
// the maze. It prevents placing a sword or prisoner next to an external wall.
func (m *Maze) randomInternalWallPos() int {
	for {
		p := randomMazePos()
		if m.area[p] == InternalWall && m.area[p-1] != ExternalWall && m.area[p+1] != ExternalWall {
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
	return rand.Intn(256)*11 + 2*MazeColumns
}
