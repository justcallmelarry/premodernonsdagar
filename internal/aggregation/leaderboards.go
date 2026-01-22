package aggregation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"premodernonsdagar/internal/utils"
	"sort"
	"strings"
)

func generateLeaderboards() error {
	// Create leaderboards directory
	leaderboardsDir := "files/lists/leaderboards"
	if err := os.MkdirAll(leaderboardsDir, 0755); err != nil {
		return fmt.Errorf("failed to create leaderboards directory: %w", err)
	}

	// Read all event files to get seasons
	eventFiles, err := filepath.Glob("files/events/*.json")
	if err != nil {
		return fmt.Errorf("failed to read event files: %w", err)
	}

	// Collect all seasons and events by season
	eventsBySeason := make(map[string][]Event)
	allEventDates := []string{}

	for _, eventFile := range eventFiles {
		data, err := os.ReadFile(eventFile)
		if err != nil {
			return fmt.Errorf("failed to read event file %s: %w", eventFile, err)
		}

		var event Event
		if err := json.Unmarshal(data, &event); err != nil {
			return fmt.Errorf("failed to parse event file %s: %w", eventFile, err)
		}

		allEventDates = append(allEventDates, event.Date)
		eventsBySeason[event.Season] = append(eventsBySeason[event.Season], event)
	}

	// Get current season
	currentSeason, err := GetCurrentSeason(allEventDates)
	if err != nil {
		return fmt.Errorf("failed to get current season: %w", err)
	}

	// Get all seasons
	seasons, err := GetAllSeasons(allEventDates)
	if err != nil {
		return fmt.Errorf("failed to get all seasons: %w", err)
	}

	// Read all player files
	playerDir := "files/players/"
	playerFiles, err := os.ReadDir(playerDir)
	if err != nil {
		return fmt.Errorf("failed to read player directory: %w", err)
	}

	allPlayers := []Player{}
	for _, file := range playerFiles {
		if filepath.Ext(file.Name()) == ".json" {
			filePath := filepath.Join(playerDir, file.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read player file %s: %w", filePath, err)
			}

			var player Player
			if err := json.Unmarshal(data, &player); err != nil {
				return fmt.Errorf("failed to parse player file %s: %w", filePath, err)
			}
			allPlayers = append(allPlayers, player)
		}
	}

	displaySeasons := []LeaderboardSeasonEntry{}
	for _, season := range seasons {
		url := "/leaderboards/" + season
		if season == currentSeason {
			url = "/leaderboards"
		}
		displaySeasons = append(displaySeasons, LeaderboardSeasonEntry{
			Season: strings.ToUpper(season),
			URL:    url,
		})
	}

	// Save seasons list

	// Generate leaderboards for each season (except current season)
	for _, season := range seasons {
		// Skip current season - it will be handled separately
		if season == currentSeason {
			continue
		}

		eventsInSeason := eventsBySeason[season]

		// Calculate season-specific stats for each player
		seasonPlayers := calculateSeasonStats(allPlayers, eventsInSeason)

		leaderboards := LeaderbardsInformation{
			Season:     strings.ToUpper(season),
			AllSeasons: displaySeasons,
			Leaderboards: []LeaderboardContainer{
				{
					Title:   "Match Win Percentage",
					Entries: topN(seasonPlayers, func(p Player) float64 { return p.MatchWinRate }, 32),
					Type:    "float",
					Suffix:  "%",
				},
				{
					Title:   "Game Win Percentage",
					Entries: topN(seasonPlayers, func(p Player) float64 { return p.GameWinRate }, 32),
					Type:    "float",
					Suffix:  "%",
				},
				{
					Title:   "Played Events",
					Entries: topN(seasonPlayers, func(p Player) float64 { return float64(p.AttendedEvents) }, 32),
					Type:    "int",
				},
				{
					Title:   "Undefeated Events",
					Entries: topN(seasonPlayers, func(p Player) float64 { return float64(p.UndefeatedEvents) }, 32),
					Type:    "int",
				},
				{
					Title:   "Extra Matches Played",
					Entries: topN(seasonPlayers, func(p Player) float64 { return float64(p.ExtraMatchesPlayed) }, 32),
					Type:    "int",
				},
				{
					Title:   "Unfinished Events",
					Entries: topN(seasonPlayers, func(p Player) float64 { return float64(p.UnfinishedEvents) }, 32),
					Type:    "int",
				},
			}}

		// Write season leaderboard file
		output, err := json.MarshalIndent(leaderboards, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal leaderboards for season %s: %w", season, err)
		}

		seasonFile := filepath.Join(leaderboardsDir, season+".json")
		if err := os.WriteFile(seasonFile, output, 0644); err != nil {
			return fmt.Errorf("failed to write leaderboards for season %s: %w", season, err)
		}
	}

	// Generate current season leaderboards with ELO and Glicko2
	currentSeasonEvents := eventsBySeason[currentSeason]
	currentSeasonPlayers := calculateSeasonStats(allPlayers, currentSeasonEvents)

	currentLeaderboards := LeaderbardsInformation{
		Season:     strings.ToUpper(currentSeason),
		AllSeasons: displaySeasons,
		Leaderboards: []LeaderboardContainer{
			{
				Title:   "Elo Rating",
				Entries: topN(allPlayers, func(p Player) float64 { return float64(p.EloRating) }, 32),
				Type:    "int",
			},
			{
				Title:   "Glicko2 Rating",
				Entries: topN(allPlayers, func(p Player) float64 { return p.GlickoRating.Mu }, 32),
				Type:    "float",
			},
			{
				Title:   "Match Win Percentage",
				Entries: topN(currentSeasonPlayers, func(p Player) float64 { return p.MatchWinRate }, 32),
				Type:    "float",
				Suffix:  "%",
			},
			{
				Title:   "Game Win Percentage",
				Entries: topN(currentSeasonPlayers, func(p Player) float64 { return p.GameWinRate }, 32),
				Type:    "float",
				Suffix:  "%",
			},
			{
				Title:   "Played Events",
				Entries: topN(currentSeasonPlayers, func(p Player) float64 { return float64(p.AttendedEvents) }, 32),
				Type:    "int",
			},
			{
				Title:   "Undefeated Events",
				Entries: topN(currentSeasonPlayers, func(p Player) float64 { return float64(p.UndefeatedEvents) }, 32),
				Type:    "int",
			},
			{
				Title:   "Extra Matches Played",
				Entries: topN(currentSeasonPlayers, func(p Player) float64 { return float64(p.ExtraMatchesPlayed) }, 32),
				Type:    "int",
			},
			{
				Title:   "Unfinished Events",
				Entries: topN(currentSeasonPlayers, func(p Player) float64 { return float64(p.UnfinishedEvents) }, 32),
				Type:    "int",
			},
		}}

	// Write current.json
	currentOutput, err := json.MarshalIndent(currentLeaderboards, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal current leaderboards: %w", err)
	}

	currentFile := filepath.Join(leaderboardsDir, "current.json")
	if err := os.WriteFile(currentFile, currentOutput, 0644); err != nil {
		return fmt.Errorf("failed to write current leaderboards: %w", err)
	}

	// Clean up old season files that are no longer needed
	// (e.g., if current season changed, remove the old current season file)
	leaderboardFiles, err := filepath.Glob(filepath.Join(leaderboardsDir, "s*.json"))
	if err != nil {
		return fmt.Errorf("failed to list leaderboard files: %w", err)
	}

	// Create a set of valid season files (all seasons except current)
	validSeasonFiles := make(map[string]bool)
	for _, season := range seasons {
		if season != currentSeason {
			validSeasonFiles[filepath.Join(leaderboardsDir, season+".json")] = true
		}
	}

	// Remove files that are not in the valid set
	for _, file := range leaderboardFiles {
		if !validSeasonFiles[file] {
			if err := os.Remove(file); err != nil {
				return fmt.Errorf("failed to remove old leaderboard file %s: %w", file, err)
			}
		}
	}

	return nil
}

// calculateSeasonStats calculates player stats filtered by season events
func calculateSeasonStats(allPlayers []Player, eventsInSeason []Event) []Player {
	// Create a map of player names to season-specific stats
	seasonStats := make(map[string]*Player)

	// Get list of player names that participated in this season
	playersInSeason := make(map[string]bool)
	for _, event := range eventsInSeason {
		for _, result := range event.Results {
			playersInSeason[result.Name] = true
		}
	}

	// Initialize season stats for each player
	for _, player := range allPlayers {
		if !playersInSeason[player.Name] {
			continue
		}

		seasonPlayer := Player{
			Name:               player.Name,
			AttendedEvents:     0,
			UndefeatedEvents:   0,
			UnfinishedEvents:   0,
			ExtraMatchesPlayed: 0,
			MatchWinRate:       0,
			GameWinRate:        0,
		}
		seasonStats[player.Name] = &seasonPlayer
	}

	// Calculate stats from events in this season
	for _, event := range eventsInSeason {
		// Track players in this event
		playersInEvent := make(map[string]bool)
		wins := make(map[string]int)
		losses := make(map[string]int)
		draws := make(map[string]int)
		gamesWon := make(map[string]int)
		gamesLost := make(map[string]int)
		extraMatches := make(map[string]int)

		for _, result := range event.Results {
			playersInEvent[result.Name] = true
		}

		// Parse matches
		for _, match := range event.Matches {
			result := ParseMatchResult(match)

			// Count extra matches
			if len(match.ExtraMatch) > 0 {
				for _, player := range match.ExtraMatch {
					extraMatches[player]++
				}
			}

			// Parse the score to get games won/lost
			// Result format is like "2-1-0" meaning player1 won 2-1
			if result.Draw {
				draws[match.Player1]++
				draws[match.Player2]++
				// For draws, assume 1-1 game score
				gamesWon[match.Player1]++
				gamesLost[match.Player1]++
				gamesWon[match.Player2]++
				gamesLost[match.Player2]++
			} else {
				wins[result.Winner]++
				losses[result.Loser]++

				// Parse game score from result string
				// Result.Score is like "2-1" or "2-0"
				var winnerGames, loserGames int
				fmt.Sscanf(result.Score, "%d-%d", &winnerGames, &loserGames)
				gamesWon[result.Winner] += winnerGames
				gamesLost[result.Winner] += loserGames
				gamesWon[result.Loser] += loserGames
				gamesLost[result.Loser] += winnerGames
			}
		}

		// Update season stats for each player in this event
		for playerName := range playersInEvent {
			if stats, ok := seasonStats[playerName]; ok {
				stats.AttendedEvents++

				// Check if undefeated (no losses or unfinished)
				if losses[playerName] == 0 {
					// Check if they finished the event (played expected number of matches)
					totalMatches := wins[playerName] + losses[playerName] + draws[playerName]
					if totalMatches == event.Rounds {
						stats.UndefeatedEvents++
					}
				}

				// Check if unfinished
				totalMatches := wins[playerName] + losses[playerName] + draws[playerName]
				if totalMatches < event.Rounds {
					stats.UnfinishedEvents++
				}

				stats.ExtraMatchesPlayed += extraMatches[playerName]
			}
		}
	}

	// Calculate win rates
	for _, stats := range seasonStats {
		// Calculate match win rate
		totalGames := 0
		totalWins := 0
		totalMatches := 0
		totalMatchWins := 0

		for _, event := range eventsInSeason {
			for _, match := range event.Matches {
				result := ParseMatchResult(match)

				if match.Player1 == stats.Name || match.Player2 == stats.Name {
					// Count matches
					isExtraMatch := false
					for _, extra := range match.ExtraMatch {
						if extra == stats.Name {
							isExtraMatch = true
							break
						}
					}

					if !isExtraMatch {
						totalMatches++

						if result.Draw {
							// Draws count as 0.5 wins for win rate
							// For games: 1-1
							totalGames += 2
							totalWins += 1
						} else if result.Winner == stats.Name {
							totalMatchWins++
							// Parse game score
							var winnerGames, loserGames int
							fmt.Sscanf(result.Score, "%d-%d", &winnerGames, &loserGames)
							totalGames += winnerGames + loserGames
							totalWins += winnerGames
						} else {
							// Loss
							var winnerGames, loserGames int
							fmt.Sscanf(result.Score, "%d-%d", &winnerGames, &loserGames)
							totalGames += winnerGames + loserGames
							totalWins += loserGames
						}
					}
				}
			}
		}

		if totalMatches > 0 {
			stats.MatchWinRate = float64(totalMatchWins) / float64(totalMatches) * 100
		}

		if totalGames > 0 {
			stats.GameWinRate = float64(totalWins) / float64(totalGames) * 100
		}
	}

	// Convert map to slice
	result := make([]Player, 0, len(seasonStats))
	for _, stats := range seasonStats {
		result = append(result, *stats)
	}

	return result
}

func topN(players []Player, scoreFunc func(Player) float64, n int) []LeaderboardEntry {
	sort.Slice(players, func(i, j int) bool {
		scoreI, scoreJ := scoreFunc(players[i]), scoreFunc(players[j])
		if scoreI == scoreJ {
			return players[i].Name < players[j].Name // Alphabetical order if scores are tied
		}
		return scoreI > scoreJ
	})

	var topPlayers []LeaderboardEntry
	for i := 0; i < len(players) && len(topPlayers) < n; i++ {
		score := scoreFunc(players[i])
		if score > 0 {
			topPlayers = append(topPlayers, LeaderboardEntry{
				Name:  players[i].Name,
				Score: score,
				URL:   "/players/" + utils.Slugify(players[i].Name),
			})
		}
	}

	return topPlayers
}
