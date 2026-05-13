package game

import (
	"fmt"
	"time"

	"github.com/cmatsuoka/mazogs/graphics"
	"github.com/cmatsuoka/mazogs/maze"
)

const (
	levelEasy = iota + 1
	levelMedium
	levelHard

	directionLeft
	directionRight
	directionUp
	directionDown
)

type SituationReport struct {
	// If the player is not holding the treasure then display the moves
	// to the treasure. If the player is holding the treasure then display
	// the moves to the exit.
	Moves          int
	MovesRemaining int
	MazogsToKill   int
	MovesKill      int
	MovesView      int
	MovesReport    int
}

type Game struct {
	hasCountdown   bool
	mazogsJump     bool
	mazogsMove     bool
	slowDown       bool
	movesRemaining int
	movesKill      int
	movesView      int
	movesReport    int
	level          int
	maze           *maze.Maze
	mazogTable     []int

	hasSword        bool
	hasTreasure     bool
	wayShown        bool
	wayShownAt      time.Time
	viewMode        bool
	viewModeAt      time.Time
	starved         bool
	exited          bool
	reportRequest   bool
	direction       int
	killed          bool
	enteredFromLeft bool
	moving          bool // true after the first step; resets when player stops
}

func New() *Game {
	maze := maze.New()
	return &Game{
		maze: maze,
	}
}

func (g *Game) Run() error {
	showIntro(g)
	level := whichGame(g)
	initialize(g, level)
	chooseEntranceSide(g)
	situationReport(g)
	play(g)
	return nil
}

// showIntro displays the animated title screen until a key is pressed.
func showIntro(g *Game) {
	g.maze.IntroMaze()
	fillScreen(0x88)
	smallDelay()
	g.maze.PlayerPos = 470
	showSprites(g.maze, 4)

	animateTitle := func(num int) (keyPressed bool) {
		// scroll "MAZOGS" down the right hand side of the screen
		for i := 4; i <= 21; i++ {
			showSprites(g.maze, num)
			graphics.PrintAt(i, 25, "MAZOGS")
			if graphics.InKey() != "" {
				return true
			}
			smallDelay()
		}
		showSprites(g.maze, 5)
		return false
	}

	for {
		graphics.PrintAt(1, 6, "A MAZE ADVENTURE GAME")
		if animateTitle(4) {
			break
		}
		showSprites(g.maze, 5)
		graphics.PrintAt(1, 6, "press_a_key_to_start`")
		if animateTitle(5) {
			break
		}
	}
}

// whichGame shows the level selection screen and sets game flags on g.
// The maze must be in IntroMaze state (set by showIntro) when called.
func whichGame(g *Game) int {
	for {
		fillScreen(0x80) // solid black, matches assembly BB subroutine (BASIC 6037: POKE 17370,128)
		renderWallBackground()
		sprites[maze.Mazog].render(0, 2) // BASIC 2082: POKE N-128,189 then USR 18602 toggles to eyes-open
		graphics.PrintAt(1, 10, "WHICH GAME ?")
		graphics.PrintAt(10, 5, "1. TRY IT OUT")
		graphics.PrintAt(12, 5, "2. FACE A CHALLENGE")
		graphics.PrintAt(14, 5, "3. MANIAC MOBILE MAZOGS")
		graphics.PrintAt(18, 5, "PRESS NUMBER TO CHOOSE")
		graphics.WaitKey()
		switch graphics.InKey() {
		case "1":
			fillScreen(0x80)
			renderWallBackground()
			sprites[maze.PlayerWithTreasureLeft].render(0, 2)
			graphics.PrintAt(1, 11, "TRY IT OUT")
			graphics.Present()
			g.hasCountdown = false
			g.mazogsMove = false
			g.mazogsJump = false
			g.slowDown = true
			return levelEasy
		case "2":
			fillScreen(0x80)
			renderWallBackground()
			sprites[maze.PlayerWithSwordRight].render(0, 2)
			graphics.PrintAt(1, 8, "FACE A CHALLENGE")
			graphics.Present()
			g.hasCountdown = true
			g.mazogsMove = false
			g.mazogsJump = false
			g.slowDown = false
			return levelMedium
		case "3":
			fillScreen(0x80)
			graphics.PrintAt(1, 6, "MANIAC MOBILE MAZOGS")
			renderWallBackground()
			for i := 0; i < 10; i++ {
				sprites[maze.PlayerStanding].render(0, 2)
				mazogCode := byte(maze.Mazog)
				if i%2 == 0 {
					mazogCode = maze.Mazog2 // toggle 0x3D-0xBD each iteration
				}
				sprites[mazogCode].render(0, 1)
				sprites[mazogCode].render(0, 3)
				graphics.Present()
			}
			g.hasCountdown = true
			g.mazogsMove = true
			g.mazogsJump = true
			g.slowDown = false
			return levelHard
		}
	}
}

func initialize(g *Game, level int) {
	g.hasTreasure = false

	for i := 0; i < 10; i++ {
		g.maze.Generate()
		// Fetch the number of empty locations. Continue if the maze is complex enough.
		if g.maze.CountEmpty() >= 1200 {
			break
		}
	}

	g.maze.InsertEntrance()
	g.mazogTable = g.maze.Populate()
	g.level = level
}

func play(g *Game) {
	fillScreen(0x88)

	g.hasSword = false
	g.hasTreasure = false
	g.wayShown = false
	g.viewMode = false
	g.starved = false
	g.exited = false
	g.reportRequest = false
	g.direction = 0
	g.killed = false
	g.moving = false

	graphics.ClearKeys() // discard any key state from previous screens
	for {
		gameLoop(g)
	}
}

func gameLoop(g *Game) {
	graphics.ProcessEvents() // pump SDL events every iteration to handle KEYUP reliably

	if g.wayShown {
		// The assembly resets FRAMES to $FFFF and clears when the high byte
		// reaches $FC (~15 seconds at 50Hz). Use a 15-second timeout here.
		if time.Since(g.wayShownAt) >= 15360*time.Millisecond {
			g.maze.ClearMaze()
			g.wayShown = false
		}
	}

	if g.viewMode {
		// Has the view timeout expired, i.e. after 5.12 seconds?
		if time.Since(g.viewModeAt) >= 5120*time.Millisecond {
			g.viewMode = false
			fillScreen(0x88) // restore game background before normal rendering
		} else {
			// Use smallDelay (not time.Sleep) so SDL events are pumped,
			// ensuring the KEYUP for 'V' is processed before the timer
			// fires and control returns to InKey().
			smallDelay()
			displayView(g)
			moveAllMazogs(g)
			graphics.Present()
			return
		}
	}

	pos := g.maze.PlayerPos
	g.direction = 0

	switch graphics.InKey() {
	case "a", "h", "Left":
		pos--
		g.direction = directionLeft
	case "d", "j", "Right":
		pos++
		g.direction = directionRight
	case "w", "Up":
		pos -= maze.MazeColumns
		g.direction = directionUp
	case "x", "s", "Down":
		pos += maze.MazeColumns
		g.direction = directionDown
	case "v":
		g.viewMode = true
		g.viewModeAt = time.Now()
		fillScreen(0x88)  // inverse checkerboard, matches assembly L43D5 (_INVCHEQUERBOARD=$88)
		decrementTimer(g) // deduct movesView once on entry
		if g.hasCountdown {
			g.maze.ClearMaze()
		}
		displayView(g)
		moveAllMazogs(g)
		graphics.Present()
		return
	case "y":
		g.reportRequest = true
		return
	case "":
		// No key held — show idle sprite and poll quickly so a new press
		// is picked up immediately.
		g.moving = false
		showPlayerStanding(g)
		graphics.ProcessEvents()
		time.Sleep(10 * time.Millisecond)
		moveAllMazogs(g)
		graphics.Present()
		return
	default:
		showPlayerStanding(g)
		smallDelay()
		moveAllMazogs(g)
		graphics.Present()
		return
	}

	code := g.maze.Map()[pos]

	switch code {
	case maze.InternalWall:
		g.moving = false
		showPlayerStanding(g)
	case maze.Empty, maze.Trail, maze.ThisWay:
		movePlayer(g, pos)
	case maze.Mazog, maze.Mazog2:
	case maze.Treasure, maze.Treasure2:
		m := g.maze
		m.Map()[m.ExitPos()] = maze.Exit // place exit in the maze
		if g.hasSword {
			m.Map()[pos] = maze.Sword // drop sword at treasure location
			g.hasSword = false
		} else {
			m.Map()[pos] = maze.InternalWall
		}
		g.hasTreasure = true
		g.moving = false
		showPlayerStanding(g) // player stays at current position; assembly sets code at DE not HL
		smallDelay()
	case maze.Prisoner, maze.Prisoner2:
		showPlayerStanding(g)
		if g.mazogsMove {
			g.maze.Map()[pos] = maze.InternalWall
		}
		g.maze.ClearMaze()
		g.maze.TraceRoute()
		showSprites(g.maze, 5) // render route markers
		graphics.Present()     // show immediately so the route is visible during the delay
		g.wayShown = true
		g.wayShownAt = time.Now()
		g.moving = false
		smallDelay()
	case maze.Sword:
		m := g.maze
		if g.hasSword {
			// already armed, treat as a wall
			showPlayerStanding(g)
		} else if g.hasTreasure {
			m.Map()[pos] = maze.Treasure // drop treasure at sword location
			g.hasTreasure = false
			m.Map()[m.ExitPos()] = maze.Empty // remove exit
			g.hasSword = true
			showPlayerStanding(g)
		} else {
			m.Map()[pos] = maze.InternalWall
			g.hasSword = true
			showPlayerStanding(g)
		}
		g.moving = false
		smallDelay()
	case maze.Exit:
	}

	moveAllMazogs(g)
	graphics.Present()
}

func moveAllMazogs(g *Game) {
}

func movePlayer(g *Game, pos int) {
	code := g.maze.Map()[g.maze.PlayerPos]

	updatePlayer := func(code, playerWithTreasure, playerWithTreasure2, playerWithSword, playerWithSword2, player, player2 byte) byte {
		if g.hasTreasure {
			if code == playerWithTreasure {
				code = playerWithTreasure2
			} else {
				code = playerWithTreasure
			}
		} else if g.hasSword {
			if code == playerWithSword {
				code = playerWithSword2
			} else {
				code = playerWithSword
			}
		} else {
			if code == player {
				code = player2
			} else {
				code = player
			}
		}
		return code
	}

	switch g.direction {
	case directionRight:
		code = updatePlayer(code,
			maze.PlayerWithTreasureRight, maze.PlayerWithTreasureRight2,
			maze.PlayerWithSwordRight, maze.PlayerWithSwordRight2,
			maze.PlayerRight, maze.PlayerRight2)
	case directionUp:
		code = updatePlayer(code,
			maze.PlayerWithTreasureUp, maze.PlayerWithTreasureUp2,
			maze.PlayerWithSwordUpDown, maze.PlayerWithSwordUpDown2,
			maze.PlayerUpDown, maze.PlayerUpDown2)
	case directionDown:
		code = updatePlayer(code,
			maze.PlayerWithTreasureDown, maze.PlayerWithTreasureDown2,
			maze.PlayerWithSwordUpDown, maze.PlayerWithSwordUpDown2,
			maze.PlayerUpDown, maze.PlayerUpDown2)
	case directionLeft:
		code = updatePlayer(code,
			maze.PlayerWithTreasureLeft, maze.PlayerWithTreasureLeft2,
			maze.PlayerWithSwordLeft, maze.PlayerWithSwordLeft2,
			maze.PlayerLeft, maze.PlayerLeft2)
	}

	mazeMap := g.maze.Map()
	// Set the maze code for the player at the new maze position.
	mazeMap[pos] = code
	// Place a trail marker in the older player location within the maze.
	mazeMap[g.maze.PlayerPos] = maze.Trail
	// Save the new location of the player in the maze.
	g.maze.PlayerPos = pos

	decrementTimer(g)
	if g.moving {
		smallDelay()
	} else {
		shortDelay()
		g.moving = true
	}
	if g.slowDown {
		smallDelay()
	}

	// Did the player starve to death?
	if g.starved {
	}

	showSprites(g.maze, 5)
}

func decrementTimer(g *Game) {
	if !g.hasCountdown {
		return
	}
	if g.viewMode {
		// Assembly L50BC: test (movesRemaining - 1 - movesView) for underflow.
		// If negative, keep movesRemaining unchanged (can't afford the view).
		// If non-negative, deduct movesView (the -1 is only used for the test).
		if g.movesRemaining-1-g.movesView < 0 {
			return
		}
		g.movesRemaining -= g.movesView
		return
	}
	// Assembly L4B56/L50DF: if movesRemaining is already 0, set starved and
	// keep it at 0. Otherwise decrement by 1 (starvation is checked on the
	// next call when 0 is found, not immediately after hitting 0).
	if g.movesRemaining == 0 {
		g.starved = true
		return
	}
	g.movesRemaining--
}

// displayView renders a 16×16 cell window of raw maze codes centred on the
// player, starting at screen position (row=4, col=8). Cells outside the maze
// bounds are shown as InternalWall. This mirrors the ZX-81 View routine at
// L517C.
func displayView(g *Game) {
	const (
		viewRows  = 16
		viewCols  = 16
		screenRow = 4
		screenCol = 8
		halfRows  = viewRows / 2
		halfCols  = viewCols / 2
	)

	area := g.maze.Map()
	mazeSize := len(area)
	startPos := g.maze.PlayerPos - halfRows*maze.MazeColumns - halfCols

	for r := 0; r < viewRows; r++ {
		for c := 0; c < viewCols; c++ {
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

func showPlayerStanding(g *Game) {
	var code byte
	if g.hasTreasure {
		code = maze.PlayerWithTreasure
	} else if g.hasSword {
		code = maze.PlayerWithSword
	} else {
		code = maze.PlayerStanding
	}
	g.maze.SetPlayerCode(code)
	showSprites(g.maze, 5)
}

// chooseEntranceSide shows the entrance view and waits for the player to press
// 'L' or 'R' to select which side of the maze to enter from. It then inserts
// a wall on the blocked side and completes entrance setup. Mirrors BASIC lines
// 6310–6366.
func chooseEntranceSide(g *Game) {
	fillScreen(0x88)
	displayView(g)
	graphics.PrintAt(2, 1, `WHICH WAY ?   PRESS "L" OR "R"`)
	graphics.Present()

	var dir int
	for {
		graphics.WaitKey()
		key := graphics.InKey()
		if key == "l" {
			dir = directionLeft
			g.enteredFromLeft = true
			break
		}
		if key == "r" {
			dir = directionRight
			g.enteredFromLeft = false
			break
		}
	}

	g.ChooseEntrance(dir)

	fillScreen(0x88)
	displayView(g)
	graphics.PrintAt(21, 2, "PRESS ANY KEY FOR INFORMATION")
	graphics.Present()
	graphics.WaitKey()
}

func (g *Game) ChooseEntrance(dir int) {
	// maze.ChooseEntrance uses dir > 0 for "enter from right"; translate our
	// direction constants (all positive) to its 0/1 convention.
	mazeDir := 0
	if dir == directionRight {
		mazeDir = 1
	}
	g.maze.ChooseEntrance(mazeDir)

	var moves int
	for {
		moves = g.maze.Distance()
		if moves > 120 {
			// The treasure is sufficiently far away.
			break
		}

		// The treasure is not far enough away.
		fmt.Printf("Treasure is too close (%d moves), relocating...\n", moves)
		g.maze.RelocateTreasure()
	}

	// Calculate the number of moves allowed. Include an extra 10 to compensate
	// for the cost of the situation report that is displayed before the game begins.
	// Note that this extra 10 moves is not included in the figure show to the user
	// in the initial situation report (it has already been deducted by then) but it
	// is used in calculating the score should the player exit with the treasure.
	g.movesRemaining = moves*4 + 10

	// Calculate the number of moves gained for killing a mazog.
	g.movesKill = moves / 5 * g.level
	if g.movesKill > 255 {
		g.movesKill = 255
	}

	// Store the number of moves lost for requesting a View.
	g.movesView = 30 / g.level
}

func situationReport(g *Game) {
	fillScreen(0x88)
	graphics.PrintAt(2, 7, "situation_report")
	moves := g.maze.Distance()
	if g.hasTreasure {
		graphics.PrintAt(5, 2, fmt.Sprintf(`MOVES BACK TO "BASE" = %d`, moves))
	} else {
		graphics.PrintAt(5, 2, fmt.Sprintf("MOVES TO THE TREASURE = %d", moves))
	}

	if g.level > levelEasy {
	}

	graphics.PrintAt(21, 2, "PRESS ANY KEY FOR THE GAME")

	graphics.Present()
	graphics.WaitKey()
}

func fillScreen(code byte) {
	for i := 0; i < 24; i++ {
		for j := 0; j < 32; j++ {
			graphics.PutZXChar(i, j, code)
		}
	}
	graphics.Present()
}

func smallDelay() {
	// Clear the latch so the next move requires a fresh key press.
	// A physically held key (keyValue) still works for continuous movement.
	graphics.ClearLatch()
	t0 := time.Now()
	for time.Since(t0) < 400*time.Millisecond {
		graphics.ProcessEvents()
		time.Sleep(10 * time.Millisecond)
	}
}

// shortDelay is used after the first step of a movement sequence to give a
// snappier start-of-walking feel. Subsequent steps use the full smallDelay.
func shortDelay() {
	graphics.ClearLatch()
	t0 := time.Now()
	for time.Since(t0) < 50*time.Millisecond {
		graphics.ProcessEvents()
		time.Sleep(10 * time.Millisecond)
	}
}
