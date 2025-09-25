package aggregation

type EventData struct {
	Name       string                     `json:"name"`
	Date       string                     `json:"date"`
	Rounds     int                        `json:"rounds"`
	PlayerInfo map[string]PlayerEventInfo `json:"player_info"`
	Matches    []Match                    `json:"matches"`
	Results    map[string]string          `json:"results"`
}

type PlayerEventInfo struct {
	Deck     string `json:"deck"`
	Decklist string `json:"decklist,omitempty"`
}

// Match represents a match between two players
type Match struct {
	Player1 string `json:"player_1"`
	Player2 string `json:"player_2"`
	Result  string `json:"result"`
}

// MatchResult holds the processed result of a match
type MatchResult struct {
	Winner string
	Loser  string
	Draw   bool
	Score  string
}

type PlayerStats struct {
	Name               string
	GamesWon           int
	GamesLost          int
	MatchesWon         int
	MatchesLost        int
	MatchesDrawn       int
	WonAgainst         map[string]int
	LostAgainst        map[string]int
	TotalGamesPlayed   int
	TotalMatchesPlayed int
	UndefeatedEvents   int
	UnfinishedEvents   int
}

// GlickoOpponent implements the glicko2.Opponent interface
type GlickoOpponent struct {
	rating float64
	rd     float64
	sigma  float64
	score  float64
}

type EventListStats struct {
	Count             int             `json:"count"`
	AverageAttendance float64         `json:"average_attendance"`
	MaxAttendance     int             `json:"max_attendance"`
	MinAttendance     int             `json:"min_attendance"`
	Events            []EventListItem `json:"events"`
}

type EventListItem struct {
	Name string `json:"name"`
	Date string `json:"date"`
	URL  string `json:"url"`
}

// Player represents a player in the system
type Player struct {
	Name             string         `json:"name"`
	EloRating        float64        `json:"elo_rating"`
	GlickoRating     GlickoRating   `json:"glicko_rating"`
	DrawCounter      int            `json:"draw_counter"`
	GameWinRate      float64        `json:"game_win_rate"`
	MatchWinRate     float64        `json:"match_win_rate"`
	WonAgainst       map[string]int `json:"won_against"`
	LostAgainst      map[string]int `json:"lost_against"`
	UndefeatedEvents int            `json:"undefeated_events"`
	UnfinishedEvents int            `json:"unfinished_events"`
}

// GlickoRating represents the Glicko-2 rating for a player
type GlickoRating struct {
	Mu    float64 `json:"mu"`    // Rating
	Phi   float64 `json:"phi"`   // Rating Deviation
	Sigma float64 `json:"sigma"` // Volatility
}

// LeaderboardEntry represents a player's entry in a leaderboard
type LeaderboardEntry struct {
	PlayerName string
	Score      interface{} // Can be float64 or int depending on the leaderboard
}

// LeaderboardData contains all the leaderboard types
type LeaderboardData struct {
	Elo                  []LeaderboardEntry
	Glicko               []LeaderboardEntry
	MatchWinPercentage   []LeaderboardEntry
	GameWinPercentage    []LeaderboardEntry
	Draw                 []LeaderboardEntry
	Bye                  []LeaderboardEntry
	MostPlayedEvents     []LeaderboardEntry
	Undefeated           []LeaderboardEntry
	ModernLiganStandings []LeaderboardEntry
}

type PlayerListEntry struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
	URL  string `json:"url"`
}
