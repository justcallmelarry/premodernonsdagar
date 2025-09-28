package aggregation

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
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

func aggregatePlayerStats() error {
	eloCalc := elogo.NewElo()

	players := make(map[string]*PlayerStats)
	eloRatings := make(map[string]int)
	glickoRatings := make(map[string]struct {
		Rating float64
		RD     float64
		Sigma  float64
	})

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

	// First pass: Identify all players
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
					}
					eloRatings[name] = 1500 // Starting ELO
					glickoRatings[name] = struct {
						Rating float64
						RD     float64
						Sigma  float64
					}{
						Rating: 1500, // Starting Glicko-2 rating
						RD:     350,  // Starting Glicko-2 rating deviation
						Sigma:  0.06, // Starting Glicko-2 volatility
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error in first pass: %w", err)
	}

	// Second pass: Process all events in chronological order
	eventFiles := []string{}
	err = filepath.WalkDir("files/events", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".json") {
			return nil
		}

		eventFiles = append(eventFiles, path)
		return nil
	})

	if err != nil {
		return fmt.Errorf("error collecting event files: %w", err)
	}

	sort.Strings(eventFiles)

	// Process each event
	for _, eventPath := range eventFiles {
		eventData, err := readEventFile(eventPath)
		if err != nil {
			return err
		}

		eventPlayers := make(map[string]bool)
		for name := range eventData.PlayerInfo {
			eventPlayers[name] = true
		}

		for _, match := range eventData.Matches {
			eventPlayers[match.Player1] = true
			eventPlayers[match.Player2] = true

			// Ensure players are initialized
			for _, player := range []string{match.Player1, match.Player2} {
				if _, exists := players[player]; !exists {
					players[player] = &PlayerStats{
						Name:        player,
						WonAgainst:  make(map[string]int),
						LostAgainst: make(map[string]int),
					}
					eloRatings[player] = 1500 // Starting ELO
					glickoRatings[player] = struct {
						Rating float64
						RD     float64
						Sigma  float64
					}{
						Rating: 1500, // Starting Glicko-2 rating
						RD:     350,  // Starting Glicko-2 rating deviation
						Sigma:  0.06, // Starting Glicko-2 volatility
					}
				}
			}

			result := ParseMatchResult(match)

			if result.Draw {
				players[match.Player1].MatchesDrawn++
				players[match.Player2].MatchesDrawn++
			} else {
				players[result.Winner].MatchesWon++
				players[result.Loser].MatchesLost++

				players[result.Winner].WonAgainst[result.Loser]++
				players[result.Loser].LostAgainst[result.Winner]++
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

			p1OutcomeElo, p2OutcomeElo := eloCalc.Outcome(eloRatings[match.Player1], eloRatings[match.Player2], eloScore)
			eloRatings[match.Player1] = p1OutcomeElo.Rating
			eloRatings[match.Player2] = p2OutcomeElo.Rating

			// Glicko-2: needs to be done after processing all matches
		}

		for name := range eventPlayers {
			players[name].AttendedEvents++

			if players[name].TotalMatchesPlayed < eventData.Rounds {
				players[name].UnfinishedEvents++

			} else if players[name].TotalMatchesPlayed > 0 &&
				players[name].MatchesLost == 0 &&
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
				rating: glickoRatings[match.Player2].Rating,
				rd:     glickoRatings[match.Player2].RD,
				sigma:  glickoRatings[match.Player2].Sigma,
				score:  scoreP1,
			})

			playerMatchesInEvent[match.Player2] = append(playerMatchesInEvent[match.Player2], GlickoOpponent{
				rating: glickoRatings[match.Player1].Rating,
				rd:     glickoRatings[match.Player1].RD,
				sigma:  glickoRatings[match.Player1].Sigma,
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
					glickoRatings[player].Rating,
					glickoRatings[player].RD,
					glickoRatings[player].Sigma,
					glickoOpponents,
					0.6, // Tau value (recommended between 0.3 and 1.2)
				)

				glickoRatings[player] = struct {
					Rating float64
					RD     float64
					Sigma  float64
				}{
					Rating: nr,
					RD:     nrd,
					Sigma:  nsigma,
				}
			} else {
				// If player skipped this tournament, update RD
				newRD := glicko2.Skip(
					glickoRatings[player].Rating,
					glickoRatings[player].RD,
					glickoRatings[player].Sigma,
				)
				glickoRatings[player] = struct {
					Rating float64
					RD     float64
					Sigma  float64
				}{
					Rating: glickoRatings[player].Rating,
					RD:     newRD,
					Sigma:  glickoRatings[player].Sigma,
				}
			}
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
			Name:           name,
			AttendedEvents: stats.AttendedEvents,
			EloRating:      math.Round(float64(eloRatings[name])*100) / 100,
			GlickoRating: GlickoRating{
				Mu:    math.Round(glickoRatings[name].Rating*100) / 100,
				Phi:   math.Round(glickoRatings[name].RD*100) / 100,
				Sigma: math.Round(glickoRatings[name].Sigma*100) / 100,
			},
			DrawCounter:      stats.MatchesDrawn,
			GameWinRate:      math.Round(gameWinRate*100) / 100,
			MatchWinRate:     math.Round(matchWinRate*100) / 100,
			WonAgainst:       stats.WonAgainst,
			LostAgainst:      stats.LostAgainst,
			UndefeatedEvents: stats.UndefeatedEvents,
			UnfinishedEvents: stats.UnfinishedEvents,
		}

		// Write player to JSON file
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

		// Mark this file as generated
		delete(existingFiles, filePath)

		// Add to the players list
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

	// Sort the players list alphabetically by slug
	sort.Slice(playersList, func(i, j int) bool {
		return playersList[i].Slug < playersList[j].Slug
	})

	// Write the players list to a JSON file
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

// readEventFile reads and parses an event JSON file
func readEventFile(path string) (*EventData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read event file %s: %w", path, err)
	}

	var eventData EventData
	if err := json.Unmarshal(data, &eventData); err != nil {
		return nil, fmt.Errorf("failed to parse event file %s: %w", path, err)
	}

	return &eventData, nil
}

// ParseMatchResult processes a match result string
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
