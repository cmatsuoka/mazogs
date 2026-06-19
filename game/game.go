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

type Game struct {
	hasCountdown   bool
	mazogsMove     bool
	slowDown       bool // adds an extra step delay after each move (TRY IT OUT only); mirrors BASIC 2170 CALL $426D patch
	movesRemaining int
	movesAllowed   int // initial move budget, set once at game start; used for the score calculation (BASIC MA)
	movesKill      int // moves added to the countdown when the player kills a mazog (player-initiated fights only)
	movesView      int // moves deducted from the countdown each time the player uses the View command
	level          int
	maze           *maze.Maze
	mazogTable     []int

	hasSword        bool
	hasTreasure     bool
	wayShown        bool
	wayShownAt      time.Time
	viewMode        bool
	viewModeAt      time.Time
	animIdleTicks   int // counts idle polls; advances animation every animIdleTicksMax ticks
	starved         bool
	exited          bool
	reportRequest   bool
	direction       int
	killed          bool
	enteredFromLeft bool
	moving          bool // true after the first step; resets when player stops
}

const (
	idlePollMs         = 50                       // ms per idle loop iteration
	stepDelayMs        = 350                      // ms per step during continuous walking
	firstStepDelayMs   = 100                      // ms after the first step; long enough for a single-step tap
	prisonerDelayMs    = 700                      // ms before prisoner reveals the route
	lineDelayTicks     = 10                       // ticks between situation report lines
	checkingDistanceMs = 1500                     // ms to display the CHECKING DISTANCE screen
	animIdleTicksMax   = stepDelayMs / idlePollMs // ticks before advancing animation when idle

	// ZX-81 frame-based timeouts. The original uses the FRAMES counter
	// (decrementing at 50Hz PAL) reset to $FFFF; timeout fires when the
	// high byte reaches a threshold.
	zx81FrameMs          = 20    // ZX-81 PAL frame period (1000/50)
	thisWayTimeoutFrames = 0x300 // high byte $FF->$FC = 768 frames
	viewTimeoutFrames    = 0x100 // high byte $FF->$FE = 256 frames

	// ZX-81 screen fill characters
	zxBlack           byte = 0x80 // solid black
	zxInvChequerboard byte = 0x88 // inverse checkerboard
)

func New() *Game {
	maze := maze.New()
	return &Game{
		maze: maze,
	}
}

// poll pumps SDL events and sleeps for idlePollMs to yield CPU in idle loops.
func poll() {
	graphics.ProcessEvents()
	time.Sleep(idlePollMs * time.Millisecond)
}

// pollFast is like poll but sleeps for a short duration for tight animation loops.
func pollFast() {
	graphics.ProcessEvents()
	time.Sleep(10 * time.Millisecond)
}

func (g *Game) Run() error {
	for {
		graphics.ClearKeys()
		showIntro(g)
		level := whichGame(g)
		initialize(g, level)
		chooseEntranceSide(g)
		key := situationReport(g)
		buySword(g, key)
		play(g)
	}
}

// buySword handles the sword purchase when the player presses T on the
// situation report screen. Mirrors BASIC 3062-3068.
// Conditions: T was pressed, player is not holding the treasure, player does
// not already have a sword, and the game is not TRY IT OUT (easy).
// Cost: the new movesRemaining is set to INT(ML/2)+1, where ML is the display
// value of movesRemaining (with the 10-move report cost already deducted).
// The purchase is skipped if the new value would be <= 2 (too few moves left).
func buySword(g *Game, key string) {
	if key != "t" || g.hasTreasure || g.hasSword || g.level == levelEasy {
		return
	}
	ml := g.movesRemaining
	if ml > 11 {
		ml -= 10
	}
	n := ml/2 + 1
	if n <= 2 {
		return // not enough moves remaining
	}
	g.movesRemaining = n
	g.hasSword = true
}

// showIntro displays the animated title screen until a key is pressed.
func showIntro(g *Game) {
	g.maze.IntroMaze()
	fillScreen(zxInvChequerboard)
	stepDelay()
	g.maze.PlayerPos = 470
	showSprites(g.maze, 4)

	animateTitle := func(num int) (keyPressed bool) {
		// scroll "MAZOGS" down the right hand side of the screen
		for i := 4; i <= 21; i++ {
			g.tick()
			showSprites(g.maze, num)
			graphics.PrintAt(i, 25, "MAZOGS")
			graphics.Present()
			// Use the same idle-loop constants as in-game so that changing
			// animIdleTicksMax or idlePollMs affects both screens equally.
			// Clear the latch so each scroll step requires a fresh key press.
			graphics.ClearLatch()
			for range animIdleTicksMax {
				if graphics.InKey() != "" {
					return true
				}
				poll()
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
		fillScreen(zxBlack) // solid black, matches Assembly BB subroutine (BASIC 6037: POKE 17370,128)
		renderWallBackground()
		sprites[maze.Mazog].render(0, 2) // BASIC 2082: POKE N-128,189 then USR 18602 toggles to eyes-open
		graphics.PrintAt(1, 10, "WHICH GAME ?")
		graphics.PrintAt(10, 5, "1. TRY IT OUT")
		graphics.PrintAt(12, 5, "2. FACE A CHALLENGE")
		graphics.PrintAt(14, 5, "3. MANIAC MOBILE MAZOGS")
		graphics.PrintAt(18, 5, "PRESS NUMBER TO CHOOSE")
		graphics.Present()
		graphics.WaitKey()
		switch graphics.InKey() {
		case "1":
			fillScreen(zxBlack)
			renderWallBackground()
			sprites[maze.PlayerWithTreasureLeft].render(0, 2)
			graphics.PrintAt(1, 11, "TRY IT OUT")
			graphics.Present()
			applyLevelSelection(g, false, false, false) // don't use extra delay for now
			return levelEasy
		case "2":
			fillScreen(zxBlack)
			renderWallBackground()
			sprites[maze.PlayerWithSwordRight].render(0, 2)
			graphics.PrintAt(1, 8, "FACE A CHALLENGE")
			graphics.Present()
			applyLevelSelection(g, true, false, false)
			return levelMedium
		case "3":
			fillScreen(zxBlack)
			renderWallBackground()
			graphics.PrintAt(1, 6, "MANIAC MOBILE MAZOGS")
			for i := range 10 {
				sprites[maze.PlayerStanding].render(0, 2)
				mazogCode := byte(maze.Mazog)
				if i%2 == 0 {
					mazogCode = maze.Mazog2 // toggle 0x3D-0xBD each iteration
				}
				sprites[mazogCode].render(0, 1)
				sprites[mazogCode].render(0, 3)
				graphics.Present()
			}
			applyLevelSelection(g, true, true, false)
			return levelHard
		}
	}
}

func applyLevelSelection(g *Game, hasCountdown, mazogsMove, slowDown bool) {
	g.hasCountdown = hasCountdown
	g.mazogsMove = mazogsMove
	g.slowDown = slowDown
}

func initialize(g *Game, level int) {
	g.hasTreasure = false

	// BASIC 6110-6114: print key controls and "maze being drawn"
	// over the existing level confirmation background (wall tiles + level sprite).
	graphics.PrintAt(14, 2, " USE KEYS  W A D AND X   OR ")
	graphics.PrintAt(16, 2, " W S H AND J.  V=VIEW,Y=STOP")
	graphics.PrintAt(21, 2, "the_maze_is_now_being_drawn_")
	graphics.Present()

	start := time.Now()

	for i := 0; i < 10; i++ {
		g.maze.Generate()
		if g.maze.CountEmpty() >= 1200 {
			break
		}
		// BASIC 6158: maze not complex enough, signal redraw.
		graphics.PrintAt(21, 2, " THE MAZE IS BEING REDRAWN  ")
		graphics.Present()
	}

	// Keep the "being drawn" screen visible for at least 5 seconds total
	// (ZX-81 generation took 5-10s; on modern hardware we pad to match).
	// No extra delay is added if generation itself took longer.
	deadline := start.Add(5 * time.Second)
	for remaining := time.Until(deadline); remaining > 0; remaining = time.Until(deadline) {
		graphics.ProcessEvents()
		sleep := remaining
		if sleep > 10*time.Millisecond {
			sleep = 10 * time.Millisecond
		}
		time.Sleep(sleep)
	}

	// BASIC 6298-6300: signal ready and wait for key press.
	graphics.PrintAt(21, 2, " MAZE READY - PRESS ANY KEY ")
	graphics.Present()
	graphics.ClearKeys()
	graphics.WaitKey()

	g.maze.InsertEntrance()
	g.mazogTable = g.maze.Populate()
	g.level = level
}

func play(g *Game) {
	fillScreen(zxInvChequerboard)

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
			key := situationReport(g)
			buySword(g, key)
			fillScreen(zxInvChequerboard) // restore game background before rejoining game loop
		}
	}

	// BASIC 3012-3040: show end-game message based on cause of death/exit.
	fillScreen(zxBlack)
	if g.starved {
		// BASIC 3118-3124: 35 iterations, printing upper then lower each time.
		flashAlternatingMessage(35, 10, 4, "YOU HAVE STARVED TO DEATH", "you have starved to death")
	} else if g.killed {
		// BASIC 3026-3034: fill screen black, draw sprites (showing the
		// mazog that killed the player), then redraw and print the death
		// message 40 times.
		advanceAnimation(g.maze)
		showSprites(g.maze, 5)
		graphics.Present()
		flashDeathMessage(g)
	} else if g.exited {
		showExitScene(g)
	}
	scoreScreen(g)
}

// scoreScreen fills the screen and shows the score if the player exited with
// the treasure on a non-easy level, then waits for G (new game) or M (examine
// maze, not yet implemented). BASIC 3200-4534.
func scoreScreen(g *Game) {
	// BASIC 3200: GOSUB SUR fills with whatever 17370 holds at this point.
	// For killed and exited paths, BASIC 6041 has restored 17370 to 136
	// (chequerboard). For the starved path, 17370 was set to 128 (black) at
	// BASIC 3100 and 3140 jumps past 3200, so the black fill from BASIC 3116
	// stays in effect.
	if g.starved {
		fillScreen(zxBlack) // BASIC 3116: black; starved path skips 3200
	} else {
		fillScreen(zxInvChequerboard) // BASIC 3200: chequerboard; killed/exited paths
	}

	// BASIC 3216: show score only when exited with treasure and not TRY IT OUT.
	if g.exited && g.level != levelEasy {
		graphics.PrintAt(4, 13, "score")
		graphics.PrintAt(7, 2, fmt.Sprintf("MOVES ALLOWED = %d", g.movesAllowed))
		graphics.PrintAt(9, 0, fmt.Sprintf("``MOVES LEFT = %d", g.movesRemaining))
		graphics.PrintAt(11, 0, fmt.Sprintf("``SCORE = %d PER CENT", g.movesRemaining*100/g.movesAllowed))
	}

	// BASIC 4500-4520.
	graphics.PrintAt(18, 2, `PRESS "M" TO EXAMINE THE MAZE`)
	graphics.PrintAt(20, 2, `PRESS "G" FOR ANOTHER GAME`)
	graphics.Present()

	// BASIC 4522-4534: wait for M or G.
	for {
		graphics.WaitKey()
		switch graphics.InKey() {
		case "g", "G":
			return
		case "m", "M":
			showMapViewer(g)
			return
		}
	}
}

func gameLoop(g *Game) {
	graphics.ProcessEvents() // pump SDL events every iteration to handle KEYUP reliably

	if g.wayShown {
		// The assembly resets FRAMES to $FFFF and clears when the high byte
		// reaches $FC, i.e. after $300 frames (768 * 20ms = 15.36s).
		if time.Since(g.wayShownAt) >= thisWayTimeoutFrames*zx81FrameMs*time.Millisecond {
			g.maze.ClearMaze()
			g.wayShown = false
		}
	}

	if g.viewMode {
		// Has the view timeout expired? Assembly checks high byte $FF->$FE,
		// i.e. $100 frames (256 * 20ms = 5.12s).
		if time.Since(g.viewModeAt) >= viewTimeoutFrames*zx81FrameMs*time.Millisecond {
			g.viewMode = false
			fillScreen(zxInvChequerboard) // restore game background before normal rendering
		} else {
			// Use stepDelay (not time.Sleep) so SDL events are pumped,
			// ensuring the KEYUP for 'V' is processed before the timer
			// fires and control returns to InKey().
			stepDelay()
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
		fillScreen(zxInvChequerboard) // inverse checkerboard, matches assembly L43D5 (_INVCHEQUERBOARD=$88)
		decrementTimer(g)             // deduct movesView once on entry
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
		// is picked up immediately. Advance animation and move mazogs at
		// the same ~400ms cadence as player steps so sprites don't freeze
		// when idle and mazogs don't outrun the player.
		g.moving = false
		g.animIdleTicks++
		if g.animIdleTicks >= animIdleTicksMax {
			g.animIdleTicks = 0
			advanceAnimation(g.maze)
			moveAllMazogs(g)
		}
		showPlayerStanding(g)
		poll()
		graphics.Present()
		return
	default:
		g.tick()
		showPlayerStanding(g)
		stepDelay()
		moveAllMazogs(g)
		graphics.Present()
		return
	}

	code := g.maze.Map()[pos]

	switch code {
	case maze.InternalWall:
		stopAtBlockedTile(g, false)
	case maze.Empty, maze.Trail, maze.ThisWay:
		movePlayer(g, pos)
	case maze.Mazog, maze.Mazog2:
		// Assembly L495C: credit movesKill before the fight (win or lose).
		g.movesRemaining += g.movesKill
		fightMazog(g, pos)
	case maze.Treasure, maze.Treasure2:
		if !beginBlockedInteraction(g, code) {
			return
		}
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
		graphics.ClearKeys()
		g.tick()
		showPlayerStanding(g) // player stays at current position; assembly sets code at DE not HL
		stepDelay()
	case maze.Prisoner, maze.Prisoner2:
		if !beginBlockedInteraction(g, code) {
			return
		}
		graphics.ClearKeys()
		g.tick()
		showPlayerStanding(g)
		prisonerAnswerDelay()
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
		stepDelay()
	case maze.Sword:
		if !beginBlockedInteraction(g, code) {
			return
		}
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
		graphics.ClearKeys()
		stepDelay()
	case maze.Exit:
		// BASIC 3014 / Assembly L4BDF: player reached the exit with the treasure.
		g.exited = true
		return
	}

	moveAllMazogs(g)
	graphics.Present()
}

// canProcessBlockedInteraction allows prisoner, sword, and treasure actions
// only after the player has already stopped next to the blocked tile. This
// prevents unintentional interactions with the tile.
func canProcessBlockedInteraction(g *Game, code byte) bool {
	if g.moving {
		return false
	}

	switch code {
	case maze.Treasure, maze.Treasure2, maze.Prisoner, maze.Prisoner2, maze.Sword:
		return true
	default:
		return false
	}
}

func flashAlternatingMessage(count, row, col int, firstMsg, secondMsg string) {
	for i := 0; i < count; i++ {
		graphics.PrintAt(row, col, firstMsg)
		graphics.Present()
		fastFlashDelay()
		graphics.PrintAt(row, col, secondMsg)
		graphics.Present()
		fastFlashDelay()
	}
}

func flashDeathMessage(g *Game) {
	for range 40 {
		advanceAnimation(g.maze)
		showSprites(g.maze, 5)
		graphics.PrintAt(18, 1, "death__to_all_treasure_seekers")
		graphics.Present()
		fastFlashDelay()
	}
}

func showExitScene(g *Game) {
	// BASIC 3150-3196: player exited with the treasure.
	// Place companion and player-with-treasure at the exit cell, then
	// flash "welcome back" 60 times. Layout depends on which side of the
	// maze the player entered from (g.enteredFromLeft).
	ep := g.maze.ExitPos()
	area := g.maze.Map()
	area[ep-maze.MazeColumns] = maze.Empty // clear exit opening above
	area[ep] = maze.PlayerStanding         // companion at exit cell

	messageCol := 7
	exitCol := 8
	playerPos := ep + 1
	if g.enteredFromLeft {
		// BASIC 3180-3196: entered from left, player is to the left.
		messageCol = 13
		exitCol = 20
		playerPos = ep - 1
	}
	area[playerPos] = maze.PlayerWithTreasure

	showSprites(g.maze, 5)
	graphics.PrintAt(10, exitCol, "EXIT")
	graphics.Present()
	flashAlternatingMessage(60, 18, messageCol, "welcome_back", "WELCOME BACK")
}

// beginBlockedInteraction applies the shared "must stop, then press again"
// rule for adjacent interactions like treasure, prisoner, and sword.
func beginBlockedInteraction(g *Game, code byte) bool {
	if canProcessBlockedInteraction(g, code) {
		return true
	}
	stopAtBlockedTile(g, true)
	graphics.Present()
	return false
}

// stopAtBlockedTile handles a blocked move attempt and can optionally discard
// the current input so the next interaction requires a fresh directional press.
func stopAtBlockedTile(g *Game, requireFreshPress bool) {
	g.moving = false
	if requireFreshPress {
		// Discard the current held direction so a blocked approach does not
		// turn into an interaction unless the player presses again.
		graphics.ClearKeys()
	}
	g.tick()
	showPlayerStanding(g)
	stepDelay()
}

func moveAllMazogs(g *Game) {
	// Assembly L4EE7: restore all mazog cells (TraceRoute/Distance wipes them).
	g.maze.InsertMazogs(g.mazogTable)

	// Assembly L4B0C: the entire mazog loop (including fight checks) is skipped
	// when there is no countdown timer (TryItOut mode).
	if !g.hasCountdown {
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

		// Assembly L4F56: mobility flag ($00=Moving, $01=Static). Level 2
		// mazogs are static -- they can still fight when adjacent but do not move.
		if !g.mazogsMove {
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
	for range 7 {
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
				pollFast()
			}
		}
	}

	// Assembly L4E5C: sword wins unconditionally; no sword = 50/50 random.
	// Assembly L4E83 uses a 4-way random split (L40B4) and wins on quarters 1 & 3.
	quarter := rand.Intn(4) + 1 // 1, 2, 3, or 4
	won := g.hasSword || quarter == 1 || quarter == 3
	if won {
		// Find this mazog in the table by position and mark it dead.
		// Unlike the assembly (L4E6B-L4E7B) which stops at the first match,
		// we kill all entries at this position. The original RNG (256 slots
		// for 38 mazogs) makes duplicate placements very likely, and leaving
		// a surviving duplicate causes InsertMazogs to respawn the mazog.
		for i, mp := range g.mazogTable {
			if mp == mazogPos {
				g.mazogTable[i] = 0xffff
			}
		}
		g.hasSword = false
		g.moving = false
		// Only clear the latch (buffered presses during the fight animation)
		// so that keyValue (the physically held key) is preserved and
		// continuous movement resumes without requiring a re-press.
		graphics.ClearLatch()
		g.tick()
		showPlayerStanding(g)
	} else {
		// Assembly L4E97: mazog wins — 25 rapid blink cycles (~90ms each),
		// then player is killed.
		g.maze.Map()[mazogPos] = maze.Mazog
		for range 25 {
			g.tick()
			showSprites(g.maze, 5)
			graphics.Present()
			fastFlashDelay()
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
	current := g.maze.Map()[g.maze.PlayerPos]
	first, second := playerMoveSprites(g.direction, g.hasTreasure, g.hasSword)
	code := toggleSprite(current, first, second)

	mazeMap := g.maze.Map()
	// Set the maze code for the player at the new maze position.
	mazeMap[pos] = code
	// Place a trail marker in the older player location within the maze.
	mazeMap[g.maze.PlayerPos] = maze.Trail
	// Save the new location of the player in the maze.
	g.maze.PlayerPos = pos

	decrementTimer(g)
	if g.starved {
		return // assembly L4389: RET Z -- skip tick/showSprites on starvation tick
	}
	if g.moving {
		stepDelay()
	} else {
		firstStepDelay()
		g.moving = true
	}
	if g.slowDown {
		stepDelay()
	}

	g.tick()
	showSprites(g.maze, 5)
}

func playerMoveSprites(direction int, hasTreasure, hasSword bool) (byte, byte) {
	switch direction {
	case directionRight:
		if hasTreasure {
			return maze.PlayerWithTreasureRight, maze.PlayerWithTreasureRight2
		}
		if hasSword {
			return maze.PlayerWithSwordRight, maze.PlayerWithSwordRight2
		}
		return maze.PlayerRight, maze.PlayerRight2
	case directionUp:
		if hasTreasure {
			return maze.PlayerWithTreasureUp, maze.PlayerWithTreasureUp2
		}
		if hasSword {
			return maze.PlayerWithSwordUpDown, maze.PlayerWithSwordUpDown2
		}
		return maze.PlayerUpDown, maze.PlayerUpDown2
	case directionDown:
		if hasTreasure {
			return maze.PlayerWithTreasureDown, maze.PlayerWithTreasureDown2
		}
		if hasSword {
			return maze.PlayerWithSwordUpDown, maze.PlayerWithSwordUpDown2
		}
		return maze.PlayerUpDown, maze.PlayerUpDown2
	case directionLeft:
		if hasTreasure {
			return maze.PlayerWithTreasureLeft, maze.PlayerWithTreasureLeft2
		}
		if hasSword {
			return maze.PlayerWithSwordLeft, maze.PlayerWithSwordLeft2
		}
		return maze.PlayerLeft, maze.PlayerLeft2
	default:
		return maze.PlayerStanding, maze.PlayerStanding
	}
}

func toggleSprite(current, first, second byte) byte {
	if current == first {
		return second
	}
	return first
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

// displayView renders a 16x16 cell window of raw maze codes centred on the
// player, starting at screen position (row=4, col=8). Cells outside the maze
// bounds are shown as InternalWall. This mirrors the ZX-81 View routine at
// Assembly L517C.
func displayView(g *Game) {
	const (
		halfRows = 8
		halfCols = 8
	)
	startPos := g.maze.PlayerPos - halfRows*maze.MazeColumns - halfCols
	displayMazeWindow(g.maze.Map(), startPos)
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
	fillScreen(zxInvChequerboard)
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

	fillScreen(zxInvChequerboard)
	displayView(g)
	graphics.PrintAt(21, 2, "PRESS ANY KEY FOR INFORMATION")
	graphics.Present()
	graphics.WaitKey()

	fillScreen(zxInvChequerboard)
	graphics.PrintAt(10, 7, "CHECKING DISTANCE")
	graphics.Present()
	time.Sleep(checkingDistanceMs * time.Millisecond)
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
		g.maze.RelocateTreasure()
	}

	applyMoveValues(g, moves)
}

// applyMoveValues sets the move budget and per-action costs derived from the
// distance to the treasure and the current level. Extracted so it can be
// tested independently of maze generation.
// Assembly / BASIC references: 6444 (movesKill), 6448 (movesView), 6450 (movesRemaining).
func applyMoveValues(g *Game, moves int) {
	// Calculate the number of moves allowed. Include an extra 10 to compensate
	// for the cost of the situation report that is displayed before the game begins.
	// Note that this extra 10 moves is not included in the figure shown to the user
	// in the initial situation report (it has already been deducted by then) but it
	// is used in calculating the score should the player exit with the treasure.
	g.movesRemaining = moves*4 + 10
	g.movesAllowed = g.movesRemaining // BASIC 6441: LET MA=N

	// Calculate the number of moves gained for killing a mazog.
	g.movesKill = moves / 5 * g.level
	if g.movesKill > 255 {
		g.movesKill = 255
	}

	// Store the number of moves lost for requesting a View.
	g.movesView = 30 / g.level
}

func situationReport(g *Game) string {
	fillScreen(zxInvChequerboard)

	// Render each line with a delay between lines to match the original
	// BASIC interpreter's per-line processing time. BASIC 6456-6500.
	renderLine := func(row, col int, msg string) {
		graphics.PrintAt(row, col, msg)
		graphics.Present()
		lineDelay()
	}

	renderLine(2, 7, "situation_report")
	moves := g.maze.Distance()
	if g.hasTreasure {
		renderLine(5, 2, fmt.Sprintf(`MOVES BACK TO "BASE" = %d`, moves))
	} else {
		renderLine(5, 2, fmt.Sprintf("MOVES TO THE TREASURE = %d", moves))
	}

	if g.level > levelEasy {
		// The situation report costs 10 moves; deduct them from the display
		// value so the player sees their true remaining budget.
		// BASIC 6477: IF ML>11 THEN LET ML=ML-10
		ml := g.movesRemaining
		if ml > 11 {
			ml -= 10
		}
		renderLine(7, 2, fmt.Sprintf("YOU HAVE %d MOVES TO GO", ml))

		if ml < moves {
			// Insufficient moves to reach the goal; show how many mazogs
			// the player must kill to make up the deficit.
			// BASIC 6482: LET X=INT((MOVES-ML)/PEEK 18779)+1
			mazogs := (moves-ml)/g.movesKill + 1
			renderLine(9, 2, "YOU NEED TO KILL AT LEAST")
			renderLine(10, 2, fmt.Sprintf("%d MAZOG", mazogs))
		}

		renderLine(13, 2, fmt.Sprintf(`YOU GAIN %d FOR A "KILL"`, g.movesKill))
		renderLine(15, 2, fmt.Sprintf(`YOU LOSE %d FOR EACH "VIEW"`, g.movesView))
		renderLine(17, 2, "THIS REPORT TAKES 10 MOVES")

		// Only shown when the player can actually buy (no sword, no treasure).
		if !g.hasSword && !g.hasTreasure {
			renderLine(19, 2, `PRESS  T  TO "BUY" A SWORD`)
		}
	}

	renderLine(21, 2, "PRESS ANY KEY FOR THE GAME")
	return graphics.WaitKey()
}

func fillScreen(code byte) {
	for i := 0; i < 24; i++ {
		for j := 0; j < 32; j++ {
			graphics.PutZXChar(i, j, code)
		}
	}
}

func stepDelay() {
	// Clear the latch so the next move requires a fresh key press.
	// A physically held key (keyValue) still works for continuous movement.
	graphics.ClearLatch()
	t0 := time.Now()
	for time.Since(t0) < stepDelayMs*time.Millisecond {
		poll()
	}
}

func fastFlashDelay() {
	waitDelay(90 * time.Millisecond)
}

func waitDelay(delay time.Duration) {
	t0 := time.Now()
	for time.Since(t0) < delay {
		pollFast()
	}
}

// prisonerAnswerDelay simulates the original route-computation pause before
// the prisoner answer becomes visible.
func prisonerAnswerDelay() {
	t0 := time.Now()
	for time.Since(t0) < prisonerDelayMs*time.Millisecond {
		poll()
	}
}

// lineDelay adds a short delay between situation report lines to simulate the
// BASIC interpreter's per-line processing time.
func lineDelay() {
	for range lineDelayTicks {
		poll()
	}
}

// firstStepDelay is used after the first step of a movement sequence. It is
// long enough for the player to release the key when only a single step
// is intended, while still feeling snappier than the full stepDelay used
// for subsequent (continuous) steps.
func firstStepDelay() {
	graphics.ClearLatch()
	t0 := time.Now()
	for time.Since(t0) < firstStepDelayMs*time.Millisecond {
		poll()
	}
}
