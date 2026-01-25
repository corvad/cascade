package cascade

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

type LinkManager struct {
	db *DB
	kv *KVStore
}

var ErrShortUrlExists = fmt.Errorf("shortUrl already exists")

func NewLinkManager(db *DB, kv *KVStore) *LinkManager {
	return &LinkManager{db: db, kv: kv}
}

func (lm *LinkManager) CreateLink(url string, shortUrl string, accountID uint) error {
	link := &Link{}
	result := lm.db.DB.Where("short_url = ?", shortUrl).First(link)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return fmt.Errorf("db lookup failed: %w", result.Error)
	}
	if result.Error == nil {
		return ErrShortUrlExists
	}
	link.Url = url
	link.Owner = accountID
	link.Count = 0
	link.ShortUrl = shortUrl
	result = lm.db.DB.Create(link)
	if result.Error != nil {
		return fmt.Errorf("db create failed: %w", result.Error)
	}
	return nil
}

func (lm *LinkManager) GetLink(shortUrl string, queriedByAccountID uint) (string, error) {
	link := &Link{}
	result := lm.db.DB.Where("short_url = ?", shortUrl).First(link)
	if result.Error != nil {
		return "", fmt.Errorf("db lookup failed: %w", result.Error)
	}
	go func() {
		result := lm.db.DB.Model(link).Update("count", link.Count+1)
		if result.Error != nil {
			log.Printf("Failed to update link count for %s: %v", shortUrl, result.Error)
		}
		//create query log entry
		query := &Query{
			LinkID:    link.ID,
			QueriedBy: queriedByAccountID,
		}
		result = lm.db.DB.Create(query)
		if result.Error != nil {
			log.Printf("Failed to create query log entry for link %s: %v", shortUrl, result.Error)
		}
	}()
	return link.Url, nil
}
