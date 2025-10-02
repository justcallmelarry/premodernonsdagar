package aggregation

type InputEvent struct {
	Name       string                     `json:"name"`
	Date       string                     `json:"date"`
	Rounds     int                        `json:"rounds"`
	PlayerInfo map[string]PlayerEventInfo `json:"player_info"` // player name -> deck
	Matches    []Match                    `json:"matches"`
}

type PlayerResult struct {
	Name     string
	Result   string
	Deck     string
	Decklist string
	URL      string
}

type Event struct {
	Name    string  `json:"name"`
	Date    string  `json:"date"`
	Rounds  int     `json:"rounds"`
	Matches []Match `json:"matches"`
	// Results       map[string]string `json:"results"`
	Results []PlayerResult `json:"results"`
}

type PlayerEventInfo struct {
	Deck     string `json:"deck"`
	Decklist string `json:"decklist,omitempty"`
}

type Match struct {
	Player1    string   `json:"player_1"`
	Player2    string   `json:"player_2"`
	Result     string   `json:"result"`
	ExtraMatch []string `json:"extra_match,omitempty"`
}

type MatchResult struct {
	Winner string
	Loser  string
	Draw   bool
	Score  string
}

type GlickoStats struct {
	Rating float64
	RD     float64
	Sigma  float64
}

type PlayerStats struct {
	Name               string
	AttendedEvents     int
	UndefeatedEvents   int
	UnfinishedEvents   int
	GamesWon           int
	GamesLost          int
	MatchesWon         int
	MatchesLost        int
	MatchesDrawn       int
	WonAgainst         map[string]int
	LostAgainst        map[string]int
	TotalGamesPlayed   int
	TotalMatchesPlayed int
	EloRating          int
	ExtraMatchesPlayed int
	GlickoRating       GlickoStats
	EloHistory         []HistoryEntry
	GlickoHistory      []HistoryEntry
	WinRateHistory     []HistoryEntry
}

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

type MatchupRecord struct {
	Opponent string `json:"opponent"`
	Count    int    `json:"count"`
}

type PlayerEventData struct {
	TotalMatchesPlayed int
	TotalWins          int
}

type HistoryEntry struct {
	Date  string  `json:"date"`
	Score float64 `json:"score"`
}

type Player struct {
	Name               string          `json:"name"`
	UndefeatedEvents   int             `json:"undefeated_events"`
	UnfinishedEvents   int             `json:"unfinished_events"`
	AttendedEvents     int             `json:"attended_events"`
	EloRating          int             `json:"elo_rating"`
	GlickoRating       GlickoRating    `json:"glicko_rating"`
	DrawCounter        int             `json:"draw_counter"`
	GameWinRate        float64         `json:"game_win_rate"`
	MatchWinRate       float64         `json:"match_win_rate"`
	WonAgainst         []MatchupRecord `json:"won_against"`
	ExtraMatchesPlayed int             `json:"extra_matches_played"`
	LostAgainst        []MatchupRecord `json:"lost_against"`
	EloHistory         []HistoryEntry  `json:"elo_history"`
	GlickoHistory      []HistoryEntry  `json:"glicko_history"`
	WinRateHistory     []HistoryEntry  `json:"win_rate_history"`
}

type GlickoRating struct {
	Mu    float64 `json:"mu"`    // Rating
	Phi   float64 `json:"phi"`   // Rating Deviation
	Sigma float64 `json:"sigma"` // Volatility
}

type LeaderboardContainer struct {
	Title   string             `json:"title"`
	Entries []LeaderboardEntry `json:"entries"`
	Type    string             `json:"type,omitempty"` // "int" or "float", optional
	Suffix  string             `json:"suffix,omitempty"`
}

type LeaderboardEntry struct {
	Name  string      `json:"name"`
	Score interface{} `json:"score"` // Can be float64 or int depending on the leaderboard
	URL   string      `json:"url"`
}

type PlayerListEntry struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
	URL  string `json:"url"`
}

type DecklistCard struct {
	Count    int    `json:"count"`
	Name     string `json:"name"`
	URL      string `json:"url,omitempty"`
	Legality string `json:"legality,omitempty"`
	CardType string `json:"card_type,omitempty"`
}

type Decklist struct {
	EventName      string         `json:"event_name"`
	PlayerName     string         `json:"player_name"`
	MainDeck       []DecklistCard `json:"main_deck"`
	MainDeckCount  int            `json:"main_deck_count"`
	Sideboard      []DecklistCard `json:"sideboard,omitempty"`
	SideboardCount int            `json:"sideboard_count"`
}
