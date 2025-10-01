package cardmatcher

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"unicode"

	"premodernonsdagar/pkg/levenshtein"
)

func NewCardDatabase() *CardDatabase {
	return &CardDatabase{
		cards: make([]Card, 0),
	}
}

func (db *CardDatabase) LoadDatabase(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read card database file: %w", err)
	}

	err = json.Unmarshal(data, &db.cards)
	if err != nil {
		return fmt.Errorf("failed to parse card database JSON: %w", err)
	}

	return nil
}

func (db *CardDatabase) GetCardCount() int {
	return len(db.cards)
}

func normalizeString(s string) string {
	// Remove leading digits
	s = strings.TrimLeftFunc(s, unicode.IsDigit)

	// Convert to lowercase and remove leading/trailing whitespace
	s = strings.ToLower(strings.TrimSpace(s))

	// Remove punctuation and special characters, keep only letters, numbers, and spaces
	var result strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			result.WriteRune(r)
		}
	}

	// Replace multiple spaces with single space
	normalized := strings.Join(strings.Fields(result.String()), " ")

	return normalized
}

func calculateSimilarity(original, target string) (float64, int) {
	normalizedOriginal := normalizeString(original)
	normalizedTarget := normalizeString(target)

	if normalizedOriginal == normalizedTarget {
		return 1.0, 0
	}

	maxLen := max(len(normalizedOriginal), len(normalizedTarget))
	distance := levenshtein.Distance(normalizedOriginal, normalizedTarget)
	similarity := 1.0 - float64(distance)/float64(maxLen)

	if strings.Contains(normalizedTarget, normalizedOriginal) || strings.Contains(normalizedOriginal, normalizedTarget) {
		similarity += 0.1
		if similarity > 1.0 {
			similarity = 1.0
		}
	}

	return similarity, distance
}

func (db *CardDatabase) FindBestMatch(query string) (*CardMatch, error) {
	if len(db.cards) == 0 {
		return nil, fmt.Errorf("card database is empty")
	}

	var bestMatch *CardMatch
	bestSimilarity := 0.0

	for _, card := range db.cards {
		similarity, distance := calculateSimilarity(query, card.Name)

		if similarity > bestSimilarity {
			bestSimilarity = similarity
			bestMatch = &CardMatch{
				Card:       card,
				Similarity: similarity,
				Distance:   distance,
			}
		}
	}

	if bestMatch == nil {
		return nil, fmt.Errorf("no matches found")
	}

	return bestMatch, nil
}

func (db *CardDatabase) Close() {
	db.cards = nil
}
