package aggregation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

func generateLeaderboards() error {
	// Directory containing player JSON files
	playerDir := "files/players/"
	outputFile := "files/lists/leaderboards.json"

	// Read all player files
	files, err := os.ReadDir(playerDir)
	if err != nil {
		return err
	}

	var players []Player

	// Parse each player file
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			filePath := filepath.Join(playerDir, file.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				return err
			}

			var player Player
			if err := json.Unmarshal(data, &player); err != nil {
				return err
			}
			players = append(players, player)
		}
	}

	// Generate leaderboards
	leaderboards := []LeaderboardContainer{
		{
			Title:   "Elo Rating",
			Entries: topN(players, func(p Player) float64 { return p.EloRating }, 8),
			Type:    "float",
		},
		{
			Title:   "Glicko2 Rating",
			Entries: topN(players, func(p Player) float64 { return p.GlickoRating.Mu }, 8),
			Type:    "float",
		},
		{
			Title:   "Match Win Percentage",
			Entries: topN(players, func(p Player) float64 { return p.MatchWinRate }, 8),
			Type:    "float",
			Suffix:  "%",
		},
		{
			Title:   "Game Win Percentage",
			Entries: topN(players, func(p Player) float64 { return p.GameWinRate }, 8),
			Type:    "float",
			Suffix:  "%",
		},
		{
			Title:   "Played Events",
			Entries: topN(players, func(p Player) float64 { return float64(p.EventsAttended) }, 8),
			Type:    "int",
		},
		{
			Title:   "Undefeated Events",
			Entries: topN(players, func(p Player) float64 { return float64(p.UndefeatedEvents) }, 8),
			Type:    "int",
		},
		{
			Title:   "Unfinished Events",
			Entries: topN(players, func(p Player) float64 { return float64(p.UnfinishedEvents) }, 8),
			Type:    "int",
		},
	}

	// Write leaderboards to file
	output, err := json.MarshalIndent(leaderboards, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputFile, output, 0644); err != nil {
		return err
	}

	return nil
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
			})
		}
	}

	return topPlayers
}
