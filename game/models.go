package game

// PlayerIndex identifies a player. Always 1 or 2.
type PlayerIndex int

const (
	Player1 PlayerIndex = 1
	Player2 PlayerIndex = 2
)

// Opponent returns the other player.
func (p PlayerIndex) Opponent() PlayerIndex {
	return 3 - p
}

// Keyword is a conqueror keyword ability.
type Keyword string

const (
	KeywordRush      Keyword = "rush"
	KeywordRend      Keyword = "rend"
	KeywordTaunt     Keyword = "taunt"
	KeywordIntercept Keyword = "intercept"
	KeywordSteadfast Keyword = "steadfast"
	KeywordFury      Keyword = "fury"
	KeywordOverwhelm Keyword = "overwhelm"
)
