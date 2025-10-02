package cardmatcher

import (
	"os"
	"path/filepath"
	"testing"
)

const PremodernCardCount = 5408

func TestCardDatabase_LoadDatabase(t *testing.T) {
	db := NewCardDatabase()

	// Test loading valid database
	err := db.LoadDatabase("../../files/db.json")
	if err != nil {
		t.Fatalf("Failed to load database: %v", err)
	}

	if db.GetCardCount() != PremodernCardCount {
		t.Errorf("Expected %d cards, got %d", PremodernCardCount, db.GetCardCount())
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

	if db.GetCardCount() != PremodernCardCount {
		t.Errorf("Expected %d cards after loading, got %d", PremodernCardCount, db.GetCardCount())
	}
}

func TestCardDatabase_Close(t *testing.T) {
	db := NewCardDatabase()
	err := db.LoadDatabase("../../files/db.json")
	if err != nil {
		t.Fatalf("Failed to load database: %v", err)
	}

	if db.GetCardCount() != PremodernCardCount {
		t.Errorf("Expected %d cards before close, got %d", PremodernCardCount, db.GetCardCount())
	}

	db.Close()

	// After close, should have no cards
	if db.GetCardCount() != 0 {
		t.Errorf("Expected 0 cards after close, got %d", db.GetCardCount())
	}

	// Should be safe to call multiple times
	db.Close()
}
