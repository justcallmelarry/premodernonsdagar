package cardmatcher

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"
)

// Card represents a single card from the database
type Card struct {
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
	Legality string `json:"legality"`
}

// CardMatch represents a fuzzy match result
type CardMatch struct {
	Card       Card
	Similarity float64
	Distance   int
}

// CardDatabase provides fuzzy matching functionality for card names
type CardDatabase struct {
	cards []Card
}

// NewCardDatabase creates and initializes a new CardDatabase
func NewCardDatabase() *CardDatabase {
	return &CardDatabase{
		cards: make([]Card, 0),
	}
}

// LoadDatabase loads the card database from the JSON file
func (db *CardDatabase) LoadDatabase(filePath string) error {
	// Read the JSON file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read card database file: %w", err)
	}

	// Parse JSON into cards slice
	err = json.Unmarshal(data, &db.cards)
	if err != nil {
		return fmt.Errorf("failed to parse card database JSON: %w", err)
	}

	return nil
}

// GetCardCount returns the number of cards loaded in the database
func (db *CardDatabase) GetCardCount() int {
	return len(db.cards)
}

// normalizeString removes special characters, converts to lowercase, and trims whitespace
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

// levenshteinDistance calculates the Levenshtein distance between two strings
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	// Create a matrix to store distances
	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
	}

	// Initialize first row and column
	for i := 0; i <= len(a); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}

	// Fill the matrix
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

// min returns the minimum of three integers
func min(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}

// calculateSimilarity calculates a similarity score between 0 and 1
func calculateSimilarity(original, target string) (float64, int) {
	normalizedOriginal := normalizeString(original)
	normalizedTarget := normalizeString(target)

	// Exact match after normalization gets perfect score
	if normalizedOriginal == normalizedTarget {
		return 1.0, 0
	}

	// Calculate Levenshtein distance
	distance := levenshteinDistance(normalizedOriginal, normalizedTarget)
	maxLen := max(len(normalizedOriginal), len(normalizedTarget))

	if maxLen == 0 {
		return 1.0, 0
	}

	// Convert distance to similarity score (1.0 = perfect match, 0.0 = no similarity)
	similarity := 1.0 - float64(distance)/float64(maxLen)

	// Bonus for substring matches
	if strings.Contains(normalizedTarget, normalizedOriginal) || strings.Contains(normalizedOriginal, normalizedTarget) {
		similarity += 0.1
		if similarity > 1.0 {
			similarity = 1.0
		}
	}

	return similarity, distance
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// FindBestMatch finds the single best matching card for a given string
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

// FindMatches finds multiple matching cards above a similarity threshold
func (db *CardDatabase) FindMatches(query string, threshold float64, maxResults int) ([]CardMatch, error) {
	if len(db.cards) == 0 {
		return nil, fmt.Errorf("card database is empty")
	}

	if threshold < 0.0 || threshold > 1.0 {
		return nil, fmt.Errorf("threshold must be between 0.0 and 1.0")
	}

	var matches []CardMatch

	for _, card := range db.cards {
		similarity, distance := calculateSimilarity(query, card.Name)

		if similarity >= threshold {
			matches = append(matches, CardMatch{
				Card:       card,
				Similarity: similarity,
				Distance:   distance,
			})
		}
	}

	// Sort matches by similarity (highest first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Similarity > matches[j].Similarity
	})

	// Limit results if maxResults is specified
	if maxResults > 0 && len(matches) > maxResults {
		matches = matches[:maxResults]
	}

	return matches, nil
}

// FindExactMatch looks for an exact match (case-insensitive, normalized)
func (db *CardDatabase) FindExactMatch(query string) (*Card, error) {
	if len(db.cards) == 0 {
		return nil, fmt.Errorf("card database is empty")
	}

	normalizedQuery := normalizeString(query)

	for _, card := range db.cards {
		if normalizeString(card.Name) == normalizedQuery {
			return &card, nil
		}
	}

	return nil, fmt.Errorf("no exact match found for: %s", query)
}

// GetAllCards returns all cards in the database
func (db *CardDatabase) GetAllCards() []Card {
	return db.cards
}

// GetCardsByLegality returns all cards with a specific legality status
func (db *CardDatabase) GetCardsByLegality(legality string) []Card {
	var result []Card
	normalizedLegality := strings.ToLower(strings.TrimSpace(legality))

	for _, card := range db.cards {
		if strings.ToLower(strings.TrimSpace(card.Legality)) == normalizedLegality {
			result = append(result, card)
		}
	}

	return result
}

// Close cleans up the database resources
func (db *CardDatabase) Close() {
	db.cards = nil
}
