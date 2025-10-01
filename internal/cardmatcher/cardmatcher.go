package cardmatcher

import (
	"fmt"
)

type CardMatcher struct {
	db *CardDatabase
}

func NewCardMatcher(dbPath string) (*CardMatcher, error) {
	db := NewCardDatabase()
	err := db.LoadDatabase(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load card database: %w", err)
	}

	return &CardMatcher{db: db}, nil
}

func (cm *CardMatcher) FindCard(query string) (*Card, error) {
	match, err := cm.db.FindBestMatch(query)
	if err != nil {
		return nil, err
	}

	return &match.Card, nil

}

func (cm *CardMatcher) FindCardWithInfo(query string) (*CardMatch, error) {
	return cm.db.FindBestMatch(query)
}

func (cm *CardMatcher) IsCardLegal(cardName string) (bool, error) {
	match, err := cm.db.FindBestMatch(cardName)
	if err != nil {
		return false, err
	}

	return match.Card.Legality == "legal", nil
}

func (cm *CardMatcher) Close() {
	if cm.db != nil {
		cm.db.Close()
	}
}
