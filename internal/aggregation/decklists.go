package aggregation

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"premodernonsdagar/internal/cardmatcher"
)

func processDecklistFile(cm *cardmatcher.CardMatcher, filePath string) (*Decklist, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	decklist := &Decklist{
		MainDeck:  make([]DecklistCard, 0),
		Sideboard: make([]DecklistCard, 0),
	}

	scanner := bufio.NewScanner(file)
	inSideboard := false
	lineNum := 0

	// Extract the date from the base filename (first 10 characters)
	baseName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	if len(baseName) < 10 {
		return nil, fmt.Errorf("filename too short to extract date: %s", filePath)
	}
	date := baseName[:10]

	// Open the event JSON file to get the event name
	eventFilePath := filepath.Join("input/events", date+".json")
	eventFile, err := os.Open(eventFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open event file: %w", err)
	}
	defer eventFile.Close()

	var eventData InputEvent
	decoder := json.NewDecoder(eventFile)
	if err := decoder.Decode(&eventData); err != nil {
		return nil, fmt.Errorf("failed to decode event file: %w", err)
	}

	decklist.EventName = eventData.Name
	for playerName, playerInfo := range eventData.PlayerInfo {
		if playerInfo.Decklist == baseName {
			decklist.DeckName = playerInfo.Deck
			decklist.PlayerName = playerName
			break
		}
	}

	// Regex to match card lines: number followed by space(s) and card name
	cardLineRegex := regexp.MustCompile(`^(\d+)\s+(.+)$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNum++

		// Skip empty lines
		if line == "" {
			continue
		}

		lc := strings.ToLower(strings.TrimSpace(line))
		if strings.Contains(lc, "sideboard") {
			inSideboard = true
			continue
		}

		// Parse card line
		matches := cardLineRegex.FindStringSubmatch(line)
		if len(matches) != 3 {
			fmt.Printf("Warning: Skipping invalid line %d in %s: %s\n", lineNum, filePath, line)
			continue
		}

		countStr := matches[1]
		cardName := strings.TrimSpace(matches[2])

		count, err := strconv.Atoi(countStr)
		if err != nil {
			fmt.Printf("Warning: Invalid count on line %d in %s: %s\n", lineNum, filePath, line)
			continue
		}

		// Find the card using the card matcher
		card, err := cm.FindCard(cardName)
		if err != nil {
			fmt.Printf("Warning: Could not find card '%s' on line %d in %s: %v\n", cardName, lineNum, filePath, err)
			// Still add the card with the original name
			card = &cardmatcher.Card{
				Name:     cardName,
				ImageURL: "",
				Legality: "unknown",
			}
		}

		decklistCard := DecklistCard{
			Count:    count,
			Name:     card.Name,
			URL:      card.ImageURL,
			Legality: card.Legality,
			CardType: card.CardType,
		}

		if inSideboard {
			decklist.Sideboard = append(decklist.Sideboard, decklistCard)
		} else {
			decklist.MainDeck = append(decklist.MainDeck, decklistCard)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	for _, card := range decklist.MainDeck {
		decklist.MainDeckCount += card.Count
	}

	for _, card := range decklist.Sideboard {
		decklist.SideboardCount += card.Count
	}

	// Sort main deck by card type (Creature, Other, Land) then alphabetically
	sort.Slice(decklist.MainDeck, func(i, j int) bool {
		typeI := getCardTypePriority(decklist.MainDeck[i].CardType)
		typeJ := getCardTypePriority(decklist.MainDeck[j].CardType)
		if typeI != typeJ {
			return typeI < typeJ
		}
		return decklist.MainDeck[i].Name < decklist.MainDeck[j].Name
	})

	// Sort sideboard alphabetically
	sort.Slice(decklist.Sideboard, func(i, j int) bool {
		return decklist.Sideboard[i].Name < decklist.Sideboard[j].Name
	})

	return decklist, nil
}

func saveDecklistAsJSON(baseName string, decklist *Decklist) error {
	// Marshal to JSON with proper formatting
	jsonData, err := json.MarshalIndent(decklist, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal decklist to JSON: %w", err)
	}

	// Write to file
	outputFilePath := "files/decklists/" + baseName + ".json"
	err = os.WriteFile(outputFilePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}

func cleanupOldFiles(generatedFiles map[string]bool) error {
	return filepath.WalkDir("files/decklists", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		baseName := strings.TrimSuffix(filepath.Base(path), ".json")

		// If this file wasn't generated in this run, delete it
		if !generatedFiles[baseName] {
			return os.Remove(path)
		}

		return nil
	})
}

func generateDecklists() error {
	err := os.MkdirAll("files/decklists", 0755)
	if err != nil {
		return fmt.Errorf("failed to create decklists directory: %w", err)
	}

	cm, err := cardmatcher.NewCardMatcher("files/db.json")
	if err != nil {
		return fmt.Errorf("failed to initialize card matcher: %w", err)
	}
	// First, get list of files we'll create so we can clean up old ones
	inputFiles, err := filepath.Glob("input/decklists/*.txt")
	if err != nil {
		return fmt.Errorf("failed to list input files: %w", err)
	}

	generatedFiles := make(map[string]bool)

	// Process each input file
	for _, inputFile := range inputFiles {
		// Get base filename without extension
		baseName := strings.TrimSuffix(filepath.Base(inputFile), ".txt")

		generatedFiles[baseName] = true

		decklist, err := processDecklistFile(cm, inputFile)
		if err != nil {
			return fmt.Errorf("failed to process %s: %w", inputFile, err)
		}

		err = saveDecklistAsJSON(baseName, decklist)
		if err != nil {
			return fmt.Errorf("failed to save %s: %w", baseName, err)
		}
	}

	err = cleanupOldFiles(generatedFiles)
	if err != nil {
		return fmt.Errorf("failed to cleanup old files: %w", err)
	}

	return nil
}

func getCardTypePriority(cardType string) int {
	switch strings.ToLower(cardType) {
	case "land":
		return 2
	case "creature":
		return 0
	case "other":
		return 1
	default:
		return 3 // Unknown types go last
	}
}
