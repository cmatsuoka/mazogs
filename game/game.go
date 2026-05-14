package game

import (
	"fmt"
	"math/rand"
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
	animIdleTicks   int // counts 10ms idle polls; advances animation every animIdleTicksMax ticks
	starved         bool
	exited          bool
	reportRequest   bool
	direction       int
	killed          bool
	enteredFromLeft bool
	moving          bool // true after the first step; resets when player stops
}

const (
	idlePollMs       = 20                        // ms per idle loop iteration
	smallDelayMs     = 200                       // ms per smallDelay (one player step)
	animIdleTicksMax = smallDelayMs / idlePollMs // ticks before advancing animation when idle
)

func New() *Game {
	maze := maze.New()
	return &Game{
		maze: maze,
	}
}

func (g *Game) Run() error {
	for {
		graphics.ClearKeys()
		showIntro(g)
		level := whichGame(g)
		initialize(g, level)
		chooseEntranceSide(g)
		situationReport(g)
		play(g)
	}
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
			g.tick()
			showSprites(g.maze, num)
			graphics.PrintAt(i, 25, "MAZOGS")
			// Use the same idle-loop constants as in-game so that changing
			// animIdleTicksMax or idlePollMs affects both screens equally.
			// Clear the latch so each scroll step requires a fresh key press.
			graphics.ClearLatch()
			for tick := 0; tick < animIdleTicksMax; tick++ {
				graphics.ProcessEvents()
				if graphics.InKey() != "" {
					return true
				}
				time.Sleep(idlePollMs * time.Millisecond)
			}
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
		fillScreen(0x80) // solid black, matches Assembly BB subroutine (BASIC 6037: POKE 17370,128)
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

	// BASIC 6110-6114: print key controls and "maze being drawn"
	// over the existing level confirmation background (wall tiles + level sprite).
	graphics.PrintAt(14, 2, " USE KEYS  W A D AND X   OR ")
	graphics.PrintAt(16, 2, " W S H AND J.  V=VIEW,Y=STOP")
	graphics.PrintAt(21, 2, "the_maze_is_now_being_drawn_")

	start := time.Now()

	for i := 0; i < 10; i++ {
		g.maze.Generate()
		if g.maze.CountEmpty() >= 1200 {
			break
		}
		// BASIC 6158: maze not complex enough, signal redraw.
		graphics.PrintAt(21, 2, " THE MAZE IS BEING REDRAWN  ")
	}

	// Keep the "being drawn" screen visible for at least 5 seconds total
	// (ZX-81 generation took 5–10s; on modern hardware we pad to match).
	// No extra delay is added if generation itself took longer.
	for remaining := 3*time.Second - time.Since(start); remaining > 0; remaining = 5*time.Second - time.Since(start) {
		graphics.ProcessEvents()
		sleep := remaining
		if sleep > 10*time.Millisecond {
			sleep = 10 * time.Millisecond
		}
		time.Sleep(sleep)
	}

	// BASIC 6298-6300: signal ready and wait for key press.
	graphics.PrintAt(21, 2, " MAZE READY - PRESS ANY KEY ")
	graphics.ClearKeys()
	graphics.WaitKey()

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
		if g.killed || g.starved || g.exited {
			break
		}
		if g.reportRequest {
			g.reportRequest = false
			situationReport(g)
		}
	}

	// BASIC 3012-3040: show end-game message based on cause of death/exit.
	fillScreen(0x80)
	if g.starved {
		// BASIC 3118-3124: flash "YOU HAVE STARVED TO DEATH" 35 times.
		for i := 0; i < 35; i++ {
			if i%2 == 0 {
				graphics.PrintAt(10, 4, "YOU HAVE STARVED TO DEATH")
			} else {
				graphics.PrintAt(10, 4, "you have starved to death")
			}
			graphics.Present()
			smallDelay()
		}
	} else if g.killed {
		// BASIC 3030-3034: flash "death  to all treasure seekers" 40 times.
		for i := 0; i < 40; i++ {
			if i%2 == 0 {
				graphics.PrintAt(18, 1, "death__to_all_treasure_seekers")
			} else {
				graphics.PrintAt(18, 1, "DEATH  TO ALL TREASURE SEEKERS")
			}
			graphics.Present()
			smallDelay()
		}
	} else if g.exited {
		// BASIC 3150-3196: player exited with the treasure.
		// Place companion and player-with-treasure at the exit cell, then
		// flash "welcome back" 60 times. Layout depends on which side of the
		// maze the player entered from (g.enteredFromLeft).
		ep := g.maze.ExitPos()
		area := g.maze.Map()
		area[ep-maze.MazeColumns] = maze.Empty // clear exit opening above
		area[ep] = maze.PlayerStanding         // companion at exit cell
		if g.enteredFromLeft {
			// BASIC 3180-3196: entered from left, player is to the left.
			area[ep-1] = maze.PlayerWithTreasure
			showSprites(g.maze, 5)
			graphics.PrintAt(10, 20, "EXIT")
			for i := 0; i < 60; i++ {
				if i%2 == 0 {
					graphics.PrintAt(18, 13, "welcome_back")
				} else {
					graphics.PrintAt(18, 13, "WELCOME BACK")
				}
				graphics.Present()
				smallDelay()
			}
		} else {
			// BASIC 3160-3178: entered from right, player is to the right.
			area[ep+1] = maze.PlayerWithTreasure
			showSprites(g.maze, 5)
			graphics.PrintAt(10, 8, "EXIT")
			for i := 0; i < 60; i++ {
				if i%2 == 0 {
					graphics.PrintAt(18, 7, "welcome_back")
				} else {
					graphics.PrintAt(18, 7, "WELCOME BACK")
				}
				graphics.Present()
				smallDelay()
			}
		}
	}
	graphics.WaitKey()
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
		// is picked up immediately. Advance animation at the same ~400ms
		// cadence as player steps so sprites don't freeze when idle.
		g.moving = false
		g.animIdleTicks++
		if g.animIdleTicks >= animIdleTicksMax {
			g.animIdleTicks = 0
			advanceAnimation(g.maze)
		}
		showPlayerStanding(g)
		graphics.ProcessEvents()
		time.Sleep(idlePollMs * time.Millisecond)
		moveAllMazogs(g)
		graphics.Present()
		return
	default:
		g.tick()
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
		g.tick()
		showPlayerStanding(g)
		smallDelay()
	case maze.Empty, maze.Trail, maze.ThisWay:
		movePlayer(g, pos)
	case maze.Mazog, maze.Mazog2:
		// Assembly L495C: credit movesKill before the fight (win or lose).
		g.movesRemaining += g.movesKill
		fightMazog(g, pos)
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
		g.tick()
		showPlayerStanding(g) // player stays at current position; assembly sets code at DE not HL
		smallDelay()
	case maze.Prisoner, maze.Prisoner2:
		g.tick()
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
		g.tick()
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
		// BASIC 3014 / Assembly L4BDF: player reached the exit with the treasure.
		g.exited = true
		return
	}

	moveAllMazogs(g)
	graphics.Present()
}

func moveAllMazogs(g *Game) {
	// Assembly L4EE7: restore all mazog cells (TraceRoute/Distance wipes them).
	g.maze.InsertMazogs(g.mazogTable)

	// Assembly L4B0C: the entire mazog loop is skipped when there is no
	// countdown timer (TryItOut mode). Only run when g.mazogsMove is true.
	if !g.mazogsMove {
		return
	}

	area := g.maze.Map()
	for i, mazogPos := range g.mazogTable {
		if mazogPos > 0xff00 { // dead sentinel 0xffff
			continue
		}

		// Assembly L4F6E: check horizontal adjacency only — mazogs never
		// initiate combat from above or below.
		playerPos := g.maze.PlayerPos
		if mazogPos-1 == playerPos || mazogPos+1 == playerPos {
			// Assembly L4FE3: mazog-initiated fight. movesKill is NOT
			// credited here (only credited on player-initiated path, L495C).
			fightMazog(g, mazogPos)
			if g.killed {
				return
			}
			continue
		}

		// Assembly L4FC0-L4FC9: pick a random direction using L40B4 4-way
		// split: 1=left, 2=right, 3=down, 4=up.
		quarter := rand.Intn(4) + 1
		var newPos int
		switch quarter {
		case 1:
			newPos = mazogPos - 1
		case 2:
			newPos = mazogPos + 1
		case 3:
			newPos = mazogPos + maze.MazeColumns
		case 4:
			newPos = mazogPos - maze.MazeColumns
		}

		// Assembly L4FCF: move only into Empty, Trail, or ThisWay cells.
		code := area[newPos]
		if code == maze.Empty || code == maze.Trail || code == maze.ThisWay {
			area[newPos] = maze.Mazog
			area[mazogPos] = maze.Empty
			g.mazogTable[i] = newPos
		}
	}
}

// fightMazog performs the fight sequence between the player and a mazog at
// mazogPos. Assembly L4E2A-L4EA3. The caller is responsible for crediting
// movesKill when appropriate (player-initiated only, Assembly L495C).
func fightMazog(g *Game, mazogPos int) {
	// Assembly L4E2A: leave trail at old position, move player to mazog's cell.
	g.maze.Map()[g.maze.PlayerPos] = maze.Trail
	g.maze.PlayerPos = mazogPos

	// 7-iteration fight animation: Fighting1, Mazog, Fighting2, Mazog2, Fighting3, Mazog.
	// On real ZX-81 each render takes ~90ms; we replicate that here.
	// Events are pumped during each sleep so KEYUP is processed.
	for i := 0; i < 7; i++ {
		for _, frame := range []byte{
			maze.Fighting1, maze.Mazog,
			maze.Fighting2, maze.Mazog2,
			maze.Fighting3, maze.Mazog,
		} {
			g.maze.Map()[mazogPos] = frame
			showSprites(g.maze, 5)
			graphics.Present()
			t0 := time.Now()
			for time.Since(t0) < 90*time.Millisecond {
				graphics.ProcessEvents()
				time.Sleep(10 * time.Millisecond)
			}
		}
	}

	// Assembly L4E5C: sword wins unconditionally; no sword = 50/50 random.
	// Assembly L4E83 uses a 4-way random split (L40B4) and wins on quarters 1 & 3.
	quarter := rand.Intn(4) + 1 // 1, 2, 3, or 4
	won := g.hasSword || quarter == 1 || quarter == 3
	if won {
		// Find this mazog in the table by position and mark it dead.
		for i, mp := range g.mazogTable {
			if mp == mazogPos {
				g.mazogTable[i] = 0xffff
				break
			}
		}
		g.hasSword = false
		g.moving = false
		graphics.ClearKeys()
		g.tick()
		showPlayerStanding(g)
	} else {
		// Assembly L4E97: mazog wins — 25 rapid blink cycles (~90ms each),
		// then player is killed.
		g.maze.Map()[mazogPos] = maze.Mazog
		for i := 0; i < 25; i++ {
			g.tick()
			showSprites(g.maze, 5)
			graphics.Present()
			t0 := time.Now()
			for time.Since(t0) < 90*time.Millisecond {
				graphics.ProcessEvents()
				time.Sleep(idlePollMs * time.Millisecond)
			}
		}
		g.killed = true
	}
}

// tick advances animation by one step and resets the idle animation counter
// so the idle loop doesn't fire again immediately after an action.
func (g *Game) tick() {
	advanceAnimation(g.maze)
	g.animIdleTicks = 0
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

	g.tick()
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
// Assembly L517C.
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
	g.maze.InsertMazogs(g.mazogTable) // Distance() in ChooseEntrance clears mazogs; restore them.

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
	for time.Since(t0) < smallDelayMs*time.Millisecond {
		graphics.ProcessEvents()
		time.Sleep(idlePollMs * time.Millisecond)
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
