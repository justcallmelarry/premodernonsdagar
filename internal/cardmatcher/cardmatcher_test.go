package cardmatcher

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewCardMatcher(t *testing.T) {
	// Test with valid database file
	testDBPath := "../../files/db.json"

	matcher, err := NewCardMatcher(testDBPath)
	if err != nil {
		t.Fatalf("Expected no error creating CardMatcher, got: %v", err)
	}
	defer matcher.Close()

	if matcher == nil {
		t.Fatal("Expected CardMatcher to be created, got nil")
	}

	if matcher.db == nil {
		t.Fatal("Expected CardMatcher to have a database, got nil")
	}
}

func TestNewCardMatcher_InvalidPath(t *testing.T) {
	// Test with non-existent file
	_, err := NewCardMatcher("nonexistent.json")
	if err == nil {
		t.Fatal("Expected error creating CardMatcher with invalid path, got nil")
	}
}

func TestCardMatcher_FindCard(t *testing.T) {
	testDBPath := "../../files/db.json"
	matcher, err := NewCardMatcher(testDBPath)
	if err != nil {
		t.Fatalf("Failed to create CardMatcher: %v", err)
	}
	defer matcher.Close()

	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{
			name:     "exact match",
			query:    "Lightning Bolt",
			expected: "Lightning Bolt",
		},
		{
			name:     "case insensitive",
			query:    "lightning bolt",
			expected: "Lightning Bolt",
		},
		{
			name:     "partial match",
			query:    "Lightning",
			expected: "Arc Lightning",
		},
		{
			name:     "typo",
			query:    "Lightining Bolt",
			expected: "Lightning Bolt",
		},
		{
			name:     "partial word match",
			query:    "Birds of",
			expected: "Birds of Paradise",
		},
		{
			name:     "misspelling",
			query:    "Sight of Man",
			expected: "Light of Day",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := matcher.FindCard(tt.query)
			if err != nil {
				t.Fatalf("FindCard returned error: %v", err)
			}

			if result.Name != tt.expected {
				t.Errorf("FindCard(%q) = %q, want %q", tt.query, result.Name, tt.expected)
			}
		})
	}
}

func TestCardMatcher_FindCardWithInfo(t *testing.T) {
	testDBPath := "../../files/db.json"
	matcher, err := NewCardMatcher(testDBPath)
	if err != nil {
		t.Fatalf("Failed to create CardMatcher: %v", err)
	}
	defer matcher.Close()

	// Test exact match
	match, err := matcher.FindCardWithInfo("Lightning Bolt")
	if err != nil {
		t.Fatalf("FindCardWithInfo returned error: %v", err)
	}

	if match == nil {
		t.Fatal("Expected CardMatch result, got nil")
	}

	if match.Card.Name != "Lightning Bolt" {
		t.Errorf("Expected card name 'Lightning Bolt', got %q", match.Card.Name)
	}

	if match.Card.Legality != "legal" {
		t.Errorf("Expected legality 'legal', got %q", match.Card.Legality)
	}

	if match.Similarity != 1.0 {
		t.Errorf("Expected similarity 1.0 for exact match, got %f", match.Similarity)
	}

	if match.Distance != 0 {
		t.Errorf("Expected distance 0 for exact match, got %d", match.Distance)
	}
}

func TestCardMatcher_IsCardLegal(t *testing.T) {
	testDBPath := "../../files/db.json"
	matcher, err := NewCardMatcher(testDBPath)
	if err != nil {
		t.Fatalf("Failed to create CardMatcher: %v", err)
	}
	defer matcher.Close()

	tests := []struct {
		name     string
		cardName string
		expected bool
	}{
		{
			name:     "legal card",
			cardName: "Lightning Bolt",
			expected: true,
		},
		{
			name:     "banned card",
			cardName: "Tendrils of Agony",
			expected: false,
		},
		{
			name:     "case insensitive legal",
			cardName: "lightning bolt",
			expected: true,
		},
		{
			name:     "case insensitive banned",
			cardName: "tendrils of agony",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := matcher.IsCardLegal(tt.cardName)
			if err != nil {
				t.Fatalf("IsCardLegal returned error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("IsCardLegal(%q) = %v, want %v", tt.cardName, result, tt.expected)
			}
		})
	}
}

func TestCardMatcher_Close(t *testing.T) {
	testDBPath := "../../files/db.json"
	matcher, err := NewCardMatcher(testDBPath)
	if err != nil {
		t.Fatalf("Failed to create CardMatcher: %v", err)
	}

	// Should not panic
	matcher.Close()

	// Should be safe to call multiple times
	matcher.Close()
}

func TestCardMatcher_FindCard_EmptyQuery(t *testing.T) {
	testDBPath := "../../files/db.json"
	matcher, err := NewCardMatcher(testDBPath)
	if err != nil {
		t.Fatalf("Failed to create CardMatcher: %v", err)
	}
	defer matcher.Close()

	// Test empty query - should still return something (best match)
	result, err := matcher.FindCard("")
	if err != nil {
		t.Fatalf("FindCard with empty query returned error: %v", err)
	}

	if result.Name == "" {
		t.Error("FindCard with empty query returned empty string")
	}
}

// Helper function to create a temporary test file
func createTempTestDB(t *testing.T, content string) string {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_db.json")

	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp test file: %v", err)
	}

	return tmpFile
}

func TestCardMatcher_InvalidJSON(t *testing.T) {
	// Test with invalid JSON
	invalidJSON := `{"invalid": "json" "missing comma"}`
	tmpFile := createTempTestDB(t, invalidJSON)

	_, err := NewCardMatcher(tmpFile)
	if err == nil {
		t.Fatal("Expected error creating CardMatcher with invalid JSON, got nil")
	}
}

func TestCardMatcher_EmptyDatabase(t *testing.T) {
	// Test with empty array
	emptyDB := `[]`
	tmpFile := createTempTestDB(t, emptyDB)

	matcher, err := NewCardMatcher(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create CardMatcher with empty database: %v", err)
	}
	defer matcher.Close()

	// Should return error when trying to find cards in empty database
	_, err = matcher.FindCard("Lightning Bolt")
	if err == nil {
		t.Fatal("Expected error finding card in empty database, got nil")
	}
}
