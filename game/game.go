package game

import (
	"fmt"
	"time"

	"github.com/cmatsuoka/mazogs/graphics"
	"github.com/cmatsuoka/mazogs/maze"
)

const (
	levelTryItOut = iota + 1
	levelFaceAChallenge
	levelManiacMobileMazogs

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

	hasSword      bool
	hasTreasure   bool
	wayShown      bool
	viewMode      bool
	starved       bool
	exited        bool
	reportRequest bool
	direction     int
	killed        bool
}

func New() *Game {
	maze := maze.New()
	return &Game{
		maze: maze,
	}
}

func (g *Game) Run() error {
	showIntro(g)
	level := whichGame()
	initialize(g, level)
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

// whichGame asks the user to select the game level.
func whichGame() int {
	return levelTryItOut
}

func initialize(g *Game, level int) {
	g.hasTreasure = false
	g.slowDown = false

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

	for {
		gameLoop(g)
	}
}

func gameLoop(g *Game) {
	if g.wayShown {
		// Has the 'This Way' timer expired?

		// The 'This Way' timer has expired so remove the 'This Way' markers.
	}

	if g.viewMode {
		// Has the view timeout expired, i.e. after 5.1 seconds?
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
	case "y":
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
		showPlayerStanding(g)
	case maze.Empty, maze.Trail, maze.ThisWay:
		movePlayer(g, pos)
	case maze.Mazog, maze.Mazog2:
	case maze.Treasure, maze.Treasure2:
	case maze.Prisoner, maze.Prisoner2:
	case maze.Sword:
	case maze.Exit:
	}

	moveAllMazogs(g)
	graphics.Present()
}

func moveAllMazogs(g *Game) {
}

func movePlayer(g *Game, pos int) {
	code := g.maze.Map()[pos]

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
	smallDelay()

	// Did the player starve to death?
	if g.starved {
	}

	if g.level == levelTryItOut {
		smallDelay()
	}

	showSprites(g.maze, 5)
}

func decrementTimer(g *Game) {
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

func (g *Game) ChooseEntrance(dir int) {
	g.maze.ChooseEntrance(dir)

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

	if g.level > levelTryItOut {
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
	t0 := time.Now()
	for time.Since(t0) < 400*time.Millisecond {
		graphics.ProcessEvents()
		time.Sleep(10 * time.Millisecond)
	}
}
