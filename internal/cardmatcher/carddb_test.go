package cardmatcher

import (
	"os"
	"path/filepath"
	"testing"
)

const PREMODERN_CARD_COUNT = 5408

func TestCardDatabase_LoadDatabase(t *testing.T) {
	db := NewCardDatabase()

	// Test loading valid database
	err := db.LoadDatabase("../../files/db.json")
	if err != nil {
		t.Fatalf("Failed to load database: %v", err)
	}

	if db.GetCardCount() != PREMODERN_CARD_COUNT {
		t.Errorf("Expected %d cards, got %d", PREMODERN_CARD_COUNT, db.GetCardCount())
	}
}

func TestCardDatabase_LoadDatabase_InvalidFile(t *testing.T) {
	db := NewCardDatabase()

	// Test loading non-existent file
	err := db.LoadDatabase("nonexistent.json")
	if err == nil {
		t.Fatal("Expected error loading non-existent file, got nil")
	}
}

func TestCardDatabase_LoadDatabase_InvalidJSON(t *testing.T) {
	db := NewCardDatabase()

	// Create temporary invalid JSON file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.json")
	err := os.WriteFile(tmpFile, []byte(`{"invalid": json}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	err = db.LoadDatabase(tmpFile)
	if err == nil {
		t.Fatal("Expected error loading invalid JSON, got nil")
	}
}

func TestCardDatabase_FindBestMatch(t *testing.T) {
	db := NewCardDatabase()
	err := db.LoadDatabase("../../files/db.json")
	if err != nil {
		t.Fatalf("Failed to load database: %v", err)
	}

	tests := []struct {
		name               string
		query              string
		expectedCard       string
		expectedSimilarity float64
		expectError        bool
	}{
		{
			name:               "exact match",
			query:              "Lightning Bolt",
			expectedCard:       "Lightning Bolt",
			expectedSimilarity: 1.0,
		},
		{
			name:               "case insensitive",
			query:              "lightning bolt",
			expectedCard:       "Lightning Bolt",
			expectedSimilarity: 1.0,
		},
		{
			name:               "with punctuation",
			query:              "Swords to Plowshares",
			expectedCard:       "Swords to Plowshares",
			expectedSimilarity: 1.0,
		},
		{
			name:         "partial match",
			query:        "Lightning",
			expectedCard: "Arc Lightning",
		},
		{
			name:         "typo",
			query:        "Lightining Bolt",
			expectedCard: "Lightning Bolt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := db.FindBestMatch(tt.query)

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if match.Card.Name != tt.expectedCard {
				t.Errorf("Expected card %q, got %q", tt.expectedCard, match.Card.Name)
			}

			if tt.expectedSimilarity > 0 && match.Similarity != tt.expectedSimilarity {
				t.Errorf("Expected similarity %f, got %f", tt.expectedSimilarity, match.Similarity)
			}
		})
	}
}

func TestCardDatabase_FindBestMatch_EmptyDatabase(t *testing.T) {
	db := NewCardDatabase()

	_, err := db.FindBestMatch("Lightning Bolt")
	if err == nil {
		t.Fatal("Expected error finding match in empty database, got nil")
	}
}

func TestCardDatabase_FindMatches(t *testing.T) {
	db := NewCardDatabase()
	err := db.LoadDatabase("../../files/db.json")
	if err != nil {
		t.Fatalf("Failed to load database: %v", err)
	}

	// Test finding multiple matches
	matches, err := db.FindMatches("Lightning", 0.3, 5)
	if err != nil {
		t.Fatalf("FindMatches returned error: %v", err)
	}

	if len(matches) == 0 {
		t.Fatal("Expected at least one match for 'Lightning'")
	}

	// Should find Arc Lightning as best match
	if matches[0].Card.Name != "Arc Lightning" {
		t.Errorf("Expected first match to be 'Arc Lightning', got %q", matches[0].Card.Name)
	}

	// Verify matches are sorted by similarity (descending)
	for i := 1; i < len(matches); i++ {
		if matches[i-1].Similarity < matches[i].Similarity {
			t.Errorf("Matches not sorted by similarity: %f < %f", matches[i-1].Similarity, matches[i].Similarity)
		}
	}
}

func TestCardDatabase_FindMatches_InvalidThreshold(t *testing.T) {
	db := NewCardDatabase()
	err := db.LoadDatabase("../../files/db.json")
	if err != nil {
		t.Fatalf("Failed to load database: %v", err)
	}

	// Test invalid threshold values
	_, err = db.FindMatches("Lightning", -0.1, 5)
	if err == nil {
		t.Fatal("Expected error with negative threshold, got nil")
	}

	_, err = db.FindMatches("Lightning", 1.1, 5)
	if err == nil {
		t.Fatal("Expected error with threshold > 1, got nil")
	}
}

func TestCardDatabase_FindMatches_MaxResults(t *testing.T) {
	db := NewCardDatabase()
	err := db.LoadDatabase("../../files/db.json")
	if err != nil {
		t.Fatalf("Failed to load database: %v", err)
	}

	// Test maxResults limiting
	matches, err := db.FindMatches("", 0.0, 3) // Very low threshold should match many cards
	if err != nil {
		t.Fatalf("FindMatches returned error: %v", err)
	}

	if len(matches) > 3 {
		t.Errorf("Expected at most 3 matches, got %d", len(matches))
	}
}

func TestCardDatabase_FindExactMatch(t *testing.T) {
	db := NewCardDatabase()
	err := db.LoadDatabase("../../files/db.json")
	if err != nil {
		t.Fatalf("Failed to load database: %v", err)
	}

	tests := []struct {
		name         string
		query        string
		expectedCard string
		expectError  bool
	}{
		{
			name:         "exact match",
			query:        "Lightning Bolt",
			expectedCard: "Lightning Bolt",
		},
		{
			name:         "case insensitive",
			query:        "lightning bolt",
			expectedCard: "Lightning Bolt",
		},
		{
			name:         "with punctuation normalization",
			query:        "Swords to Plowshares",
			expectedCard: "Swords to Plowshares",
		},
		{
			name:        "no match",
			query:       "Nonexistent Card",
			expectError: true,
		},
		{
			name:        "partial match should fail",
			query:       "Lightning",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card, err := db.FindExactMatch(tt.query)

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if card.Name != tt.expectedCard {
				t.Errorf("Expected card %q, got %q", tt.expectedCard, card.Name)
			}
		})
	}
}

func TestCardDatabase_GetCardsByLegality(t *testing.T) {
	db := NewCardDatabase()
	err := db.LoadDatabase("../../files/db.json")
	if err != nil {
		t.Fatalf("Failed to load database: %v", err)
	}

	// Test getting legal cards
	legalCards := db.GetCardsByLegality("legal")
	if len(legalCards) == 0 {
		t.Fatal("Expected at least one legal card")
	}

	// Verify all returned cards are legal
	for _, card := range legalCards {
		if card.Legality != "legal" {
			t.Errorf("Expected legal card, got card with legality: %q", card.Legality)
		}
	}

	// Test getting banned cards
	bannedCards := db.GetCardsByLegality("banned")
	if len(bannedCards) == 0 {
		t.Fatal("Expected at least one banned card")
	}

	// Verify all returned cards are banned
	for _, card := range bannedCards {
		if card.Legality != "banned" {
			t.Errorf("Expected banned card, got card with legality: %q", card.Legality)
		}
	}

	// Test case insensitive
	legalCardsCI := db.GetCardsByLegality("LEGAL")
	if len(legalCardsCI) != len(legalCards) {
		t.Errorf("Case insensitive search returned different results: %d vs %d", len(legalCardsCI), len(legalCards))
	}

	// Test non-existent legality
	nonexistent := db.GetCardsByLegality("nonexistent")
	if len(nonexistent) != 0 {
		t.Errorf("Expected no cards for nonexistent legality, got %d", len(nonexistent))
	}
}

func TestCardDatabase_GetAllCards(t *testing.T) {
	db := NewCardDatabase()
	err := db.LoadDatabase("../../files/db.json")
	if err != nil {
		t.Fatalf("Failed to load database: %v", err)
	}

	allCards := db.GetAllCards()
	if len(allCards) != PREMODERN_CARD_COUNT {
		t.Errorf("Expected %d cards, got %d", PREMODERN_CARD_COUNT, len(allCards))
	}

	// Verify some expected cards are present
	cardNames := make(map[string]bool)
	for _, card := range allCards {
		cardNames[card.Name] = true
	}

	expectedCards := []string{"Lightning Bolt", "Survival of the Fittest", "Force of Will"}
	for _, expected := range expectedCards {
		if !cardNames[expected] {
			t.Errorf("Expected card %q not found in results", expected)
		}
	}
}

func TestCardDatabase_GetCardCount(t *testing.T) {
	db := NewCardDatabase()

	// Empty database should have 0 cards
	if db.GetCardCount() != 0 {
		t.Errorf("Expected 0 cards in empty database, got %d", db.GetCardCount())
	}

	err := db.LoadDatabase("../../files/db.json")
	if err != nil {
		t.Fatalf("Failed to load database: %v", err)
	}

	if db.GetCardCount() != PREMODERN_CARD_COUNT {
		t.Errorf("Expected %d cards after loading, got %d", PREMODERN_CARD_COUNT, db.GetCardCount())
	}
}

func TestCardDatabase_Close(t *testing.T) {
	db := NewCardDatabase()
	err := db.LoadDatabase("../../files/db.json")
	if err != nil {
		t.Fatalf("Failed to load database: %v", err)
	}

	if db.GetCardCount() != PREMODERN_CARD_COUNT {
		t.Errorf("Expected %d cards before close, got %d", PREMODERN_CARD_COUNT, db.GetCardCount())
	}

	db.Close()

	// After close, should have no cards
	if db.GetCardCount() != 0 {
		t.Errorf("Expected 0 cards after close, got %d", db.GetCardCount())
	}

	// Should be safe to call multiple times
	db.Close()
}
