package cardmatcher

import (
	"fmt"
)

// CardMatcher provides a simple interface for card matching operations
type CardMatcher struct {
	db *CardDatabase
}

// NewCardMatcher creates a new CardMatcher and loads the database
func NewCardMatcher(dbPath string) (*CardMatcher, error) {
	db := NewCardDatabase()
	err := db.LoadDatabase(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load card database: %w", err)
	}

	return &CardMatcher{db: db}, nil
}

// FindCard is a convenience method that finds the best match for a card name and returns just the card name
func (cm *CardMatcher) FindCard(query string) (*Card, error) {
	match, err := cm.db.FindBestMatch(query)
	if err != nil {
		return nil, err
	}

	return &match.Card, nil

}

// FindCardWithInfo returns the best match with additional information
func (cm *CardMatcher) FindCardWithInfo(query string) (*CardMatch, error) {
	return cm.db.FindBestMatch(query)
}

// IsCardLegal checks if a card is legal (not banned)
func (cm *CardMatcher) IsCardLegal(cardName string) (bool, error) {
	match, err := cm.db.FindBestMatch(cardName)
	if err != nil {
		return false, err
	}

	return match.Card.Legality == "legal", nil
}

// Close cleans up the matcher resources
func (cm *CardMatcher) Close() {
	if cm.db != nil {
		cm.db.Close()
	}
}
