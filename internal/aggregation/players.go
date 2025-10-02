package aggregation

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"premodernonsdagar/internal/utils"
	elogo "premodernonsdagar/pkg/elo"
	"premodernonsdagar/pkg/glicko2"
)

func (o *GlickoOpponent) R() float64     { return o.rating }
func (o *GlickoOpponent) RD() float64    { return o.rd }
func (o *GlickoOpponent) Sigma() float64 { return o.sigma }
func (o *GlickoOpponent) SJ() float64    { return o.score }

// sortMatchupsByValue sorts a map[string]int by value in descending order
func sortMatchupsByValue(matchups map[string]int) []MatchupRecord {
	// Convert map to slice of MatchupRecord
	records := make([]MatchupRecord, 0, len(matchups))
	for opponent, count := range matchups {
		records = append(records, MatchupRecord{
			Opponent: opponent,
			Count:    count,
		})
	}

	// Sort by count in descending order
	sort.Slice(records, func(i, j int) bool {
		return records[i].Count > records[j].Count
	})

	return records
}

func aggregatePlayerStats() error {
	eloCalc := elogo.NewElo()

	players := make(map[string]*PlayerStats)

	err := os.MkdirAll("files/players", 0755)
	if err != nil {
		return fmt.Errorf("failed to create players directory: %w", err)
	}

	err = os.MkdirAll("files/events", 0755)
	if err != nil {
		return fmt.Errorf("failed to create events directory: %w", err)
	}

	err = os.MkdirAll("files/lists", 0755)
	if err != nil {
		return fmt.Errorf("failed to create lists directory: %w", err)
	}

	// Collect existing player JSON files
	existingFiles := make(map[string]bool)
	err = filepath.WalkDir("files/players", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(path), ".json") {
			existingFiles[path] = true
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to collect existing player files: %w", err)
	}

	eventFiles := []string{}
	err = filepath.WalkDir("files/events", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".json") {
			return nil
		}

		eventData, err := readEventFile(path)
		if err != nil {
			return err
		}

		for _, match := range eventData.Matches {
			for _, name := range []string{match.Player1, match.Player2} {
				if _, exists := players[name]; !exists {
					players[name] = &PlayerStats{
						Name:        name,
						WonAgainst:  make(map[string]int),
						LostAgainst: make(map[string]int),
						EloRating:   1500,
						GlickoRating: GlickoStats{
							Rating: 1500,
							RD:     350,
							Sigma:  0.06,
						},
						EloHistory: []HistoryEntry{
							{Date: "Unranked", Score: 1500},
						},
						GlickoHistory: []HistoryEntry{
							{Date: "Unranked", Score: 1500},
						},
					}
				}
			}
		}

		eventFiles = append(eventFiles, path)

		return nil
	})

	if err != nil {
		return fmt.Errorf("error in first pass: %w", err)
	}

	sort.Strings(eventFiles)

	// Process each event
	for _, eventPath := range eventFiles {
		eventData, err := readEventFile(eventPath)
		if err != nil {
			return err
		}

		eventPlayerData := make(map[string]*PlayerEventData)

		for _, match := range eventData.Matches {
			if _, exists := eventPlayerData[match.Player1]; !exists {
				eventPlayerData[match.Player1] = &PlayerEventData{
					TotalMatchesPlayed: 0,
					TotalWins:          0,
				}
			}
			if _, exists := eventPlayerData[match.Player2]; !exists {
				eventPlayerData[match.Player2] = &PlayerEventData{
					TotalMatchesPlayed: 0,
					TotalWins:          0,
				}
			}
			eventPlayerData[match.Player1].TotalMatchesPlayed++
			eventPlayerData[match.Player2].TotalMatchesPlayed++

			result := ParseMatchResult(match)

			if result.Draw {
				players[match.Player1].MatchesDrawn++
				players[match.Player2].MatchesDrawn++
			} else {
				eventPlayerData[result.Winner].TotalWins++
				players[result.Winner].MatchesWon++
				players[result.Loser].MatchesLost++

				players[result.Winner].WonAgainst[result.Loser]++
				players[result.Loser].LostAgainst[result.Winner]++
			}

			for _, p := range []string{match.Player1, match.Player2} {
				if slices.Contains(match.ExtraMatch, p) {
					players[p].ExtraMatchesPlayed++
				}
			}

			parts := strings.Split(result.Score, "-")
			if len(parts) == 2 {
				p1Games, p2Games := 0, 0
				fmt.Sscanf(result.Score, "%d-%d", &p1Games, &p2Games)

				players[match.Player1].GamesWon += p1Games
				players[match.Player1].GamesLost += p2Games
				players[match.Player2].GamesWon += p2Games
				players[match.Player2].GamesLost += p1Games

				players[match.Player1].TotalGamesPlayed += p1Games + p2Games
				players[match.Player2].TotalGamesPlayed += p1Games + p2Games
			}

			players[match.Player1].TotalMatchesPlayed++
			players[match.Player2].TotalMatchesPlayed++

			// Update ELO ratings
			eloScore := 0.5 // Draw by default
			if !result.Draw {
				if result.Winner == match.Player1 {
					eloScore = 1.0
				} else {
					eloScore = 0.0
				}
			}

			p1OutcomeElo, p2OutcomeElo := eloCalc.Outcome(players[match.Player1].EloRating, players[match.Player2].EloRating, eloScore)
			players[match.Player1].EloRating = p1OutcomeElo.Rating
			players[match.Player2].EloRating = p2OutcomeElo.Rating

			// Glicko-2: needs to be done after processing all matches
		}

		for name := range eventPlayerData {
			players[name].AttendedEvents++

			if eventPlayerData[name].TotalMatchesPlayed < eventData.Rounds {
				players[name].UnfinishedEvents++

			} else if eventPlayerData[name].TotalWins == eventData.Rounds &&
				players[name].MatchesDrawn == 0 {
				players[name].UndefeatedEvents++
			}
		}

		playerMatchesInEvent := make(map[string][]GlickoOpponent)

		for _, match := range eventData.Matches {
			result := ParseMatchResult(match)

			// Score for player 1
			scoreP1 := 0.5 // Draw by default
			if !result.Draw {
				if result.Winner == match.Player1 {
					scoreP1 = 1.0
				} else {
					scoreP1 = 0.0
				}
			}

			playerMatchesInEvent[match.Player1] = append(playerMatchesInEvent[match.Player1], GlickoOpponent{
				rating: players[match.Player2].GlickoRating.Rating,
				rd:     players[match.Player2].GlickoRating.RD,
				sigma:  players[match.Player2].GlickoRating.Sigma,
				score:  scoreP1,
			})

			playerMatchesInEvent[match.Player2] = append(playerMatchesInEvent[match.Player2], GlickoOpponent{
				rating: players[match.Player1].GlickoRating.Rating,
				rd:     players[match.Player1].GlickoRating.RD,
				sigma:  players[match.Player1].GlickoRating.Sigma,
				score:  1.0 - scoreP1, // Reverse score for player 2
			})
		}

		// Update Glicko-2 ratings for each player
		for player, opponents := range playerMatchesInEvent {
			if len(opponents) > 0 {
				// Convert our GlickoOpponent to the glicko2.Opponent interface
				glickoOpponents := make([]glicko2.Opponent, len(opponents))
				for i, opp := range opponents {
					oppCopy := opp // Create a copy to avoid issues with pointer reuse
					glickoOpponents[i] = &oppCopy
				}

				// Update rating
				nr, nrd, nsigma := glicko2.Rank(
					players[player].GlickoRating.Rating,
					players[player].GlickoRating.RD,
					players[player].GlickoRating.Sigma,
					glickoOpponents,
					0.6, // Tau value (recommended between 0.3 and 1.2)
				)

				players[player].GlickoRating.Rating = nr
				players[player].GlickoRating.RD = nrd
				players[player].GlickoRating.Sigma = nsigma
			} else {
				// If player skipped this tournament, update RD
				newRD := glicko2.Skip(
					players[player].GlickoRating.Rating,
					players[player].GlickoRating.RD,
					players[player].GlickoRating.Sigma,
				)
				players[player].GlickoRating = GlickoStats{
					Rating: players[player].GlickoRating.Rating,
					RD:     newRD,
					Sigma:  players[player].GlickoRating.Sigma,
				}
			}
		}
		for name := range eventPlayerData {
			players[name].EloHistory = append(players[name].EloHistory, HistoryEntry{
				Date:  eventData.Date,
				Score: float64(players[name].EloRating),
			})
			players[name].GlickoHistory = append(players[name].GlickoHistory, HistoryEntry{
				Date:  eventData.Date,
				Score: math.Round(players[name].GlickoRating.Rating*100) / 100,
			})
			players[name].WinRateHistory = append(players[name].WinRateHistory, HistoryEntry{
				Date:  eventData.Date,
				Score: math.Round(float64(players[name].MatchesWon)/float64(players[name].MatchesWon+players[name].MatchesLost+players[name].MatchesDrawn)*10000) / 100,
			})
		}
	}

	playersList := []PlayerListEntry{}

	for name, stats := range players {
		if stats.TotalMatchesPlayed == 0 {
			continue
		}

		gameWinRate := 0.0
		if stats.GamesWon+stats.GamesLost > 0 {
			gameWinRate = float64(stats.GamesWon) / float64(stats.GamesWon+stats.GamesLost) * 100.0
		}

		matchWinRate := 0.0
		if stats.MatchesWon+stats.MatchesLost+stats.MatchesDrawn > 0 {
			matchWinRate = float64(stats.MatchesWon) / float64(stats.MatchesWon+stats.MatchesLost+stats.MatchesDrawn) * 100.0
		}

		player := &Player{
			Name:             name,
			AttendedEvents:   stats.AttendedEvents,
			UndefeatedEvents: stats.UndefeatedEvents,
			UnfinishedEvents: stats.UnfinishedEvents,
			EloRating:        players[name].EloRating,
			GlickoRating: GlickoRating{
				Mu:    math.Round(players[name].GlickoRating.Rating*100) / 100,
				Phi:   math.Round(players[name].GlickoRating.RD*100) / 100,
				Sigma: math.Round(players[name].GlickoRating.Sigma*100) / 100,
			},
			DrawCounter:        stats.MatchesDrawn,
			GameWinRate:        math.Round(gameWinRate*100) / 100,
			MatchWinRate:       math.Round(matchWinRate*100) / 100,
			ExtraMatchesPlayed: stats.ExtraMatchesPlayed,
			WonAgainst:         sortMatchupsByValue(stats.WonAgainst),
			LostAgainst:        sortMatchupsByValue(stats.LostAgainst),
			EloHistory:         stats.EloHistory,
			GlickoHistory:      stats.GlickoHistory,
			WinRateHistory:     stats.WinRateHistory,
		}

		playerJSON, err := json.MarshalIndent(player, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal player %s: %w", name, err)
		}

		slug := utils.Slugify(name)
		filePath := fmt.Sprintf("files/players/%s.json", slug)
		err = os.WriteFile(filePath, playerJSON, 0644)
		if err != nil {
			return fmt.Errorf("failed to write player file for %s: %w", name, err)
		}

		delete(existingFiles, filePath)

		playersList = append(playersList, PlayerListEntry{
			Name: name,
			Slug: slug,
			URL:  "/players/" + slug,
		})
	}

	// Remove any leftover files that were not regenerated
	for filePath := range existingFiles {
		err := os.Remove(filePath)
		if err != nil {
			return fmt.Errorf("failed to remove old player file %s: %w", filePath, err)
		}
	}

	sort.Slice(playersList, func(i, j int) bool {
		return playersList[i].Slug < playersList[j].Slug
	})

	playersListJSON, err := json.MarshalIndent(playersList, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal players list: %w", err)
	}

	err = os.WriteFile("files/lists/players.json", playersListJSON, 0644)
	if err != nil {
		return fmt.Errorf("failed to write players list file: %w", err)
	}

	return nil
}

func readEventFile(path string) (*Event, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read event file %s: %w", path, err)
	}

	var eventData Event
	if err := json.Unmarshal(data, &eventData); err != nil {
		return nil, fmt.Errorf("failed to parse event file %s: %w", path, err)
	}

	return &eventData, nil
}

func ParseMatchResult(match Match) MatchResult {
	parts := strings.Split(match.Result, "-")
	if len(parts) != 2 {
		return MatchResult{Draw: true, Score: match.Result}
	}

	var p1Score, p2Score int
	_, err := fmt.Sscanf(match.Result, "%d-%d", &p1Score, &p2Score)
	if err != nil {
		return MatchResult{Draw: true, Score: match.Result}
	}

	if p1Score == p2Score {
		return MatchResult{
			Draw:  true,
			Score: match.Result,
		}
	}

	if p1Score > p2Score {
		return MatchResult{
			Winner: match.Player1,
			Loser:  match.Player2,
			Draw:   false,
			Score:  match.Result,
		}
	} else {
		return MatchResult{
			Winner: match.Player2,
			Loser:  match.Player1,
			Draw:   false,
			Score:  match.Result,
		}
	}
}
