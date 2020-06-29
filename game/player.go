package game

type Player struct {
	pos         int
	hasSword    bool
	hasTreasure bool
	starved     bool
	exited      bool
	direction   byte
	killed      bool
}

func NewPlayer() *Player {
	return &Player{}
}
