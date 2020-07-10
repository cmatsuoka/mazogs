package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cmatsuoka/mazogs/graphics"
	"github.com/cmatsuoka/mazogs/maze"
)

const (
	LevelTryItOut = iota + 1
	LevelFaceAChallenge
	LevelManiacMobileMazogs
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
	player         *Player
}

func New() *Game {
	maze := maze.New()
	player := NewPlayer()
	return &Game{
		maze:   maze,
		player: player,
	}
}

// Intro displays the animated title screen until a key is pressed.
func (g *Game) Intro() {
	g.maze.IntroMaze()
	fillScreen(0x88)
	showSprites(g.maze, 4)

	animateTitle := func(num int) (keyPressed bool) {
		// scroll "MAZOGS" down the right hand side of the screen
		for i := 4; i <= 21; i++ {
			showSprites(g.maze, num)
			graphics.PrintAt(i, 25, "MAZOGS")
			if graphics.InKey() != "" {
				return true
			}
		}
		return false
	}

	for {
		graphics.PrintAt(1, 6, "A MAZE ADVENTURE GAME")
		if animateTitle(4) {
			break
		}
		showSprites(g.maze, 5)
		graphics.PrintAt(1, 6, "press a key to start")
		if animateTitle(5) {
			break
		}
	}
}

// WhichGame asks the user to select the game level.
func (g *Game) WhichGame() int {
	return LevelTryItOut
}

func (g *Game) Initialize(level int) {
	g.player.hasTreasure = false
	g.slowDown = false

	for i := 0; i < 10; i++ {
		timeout := 512 + rand.Intn(512)
		g.maze.Generate(time.Duration(timeout) * time.Millisecond)
		// Fetch the number of empty locations. Continue if the maze is complex enough.
		if g.maze.CountEmpty() >= 1200 {
			break
		}
	}

	g.maze.InsertEntrance()
	g.mazogTable = g.maze.Populate()
	g.level = level
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

func (g *Game) Map() []byte {
	return g.maze.Map()
}

func (g *Game) Report() SituationReport {
	return SituationReport{}
}

func fillScreen(code byte) {

}

// showSprites toggles maze codes and displays a screenful of sprites around player
// location. A maximum of num columns of sprites are displayed.
func showSprites(m *maze.Maze, num int) {
	fillScreen(0x80)
}
