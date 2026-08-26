package service

import (
	"time"

	"duekeep/internal/model"
)

// StatusAtWrite считает active/expiring/expired. cancelled/archived не пересчитывает.
func StatusAtWrite(today, expires time.Time, notifyDays int, requested string) string {
	switch requested {
	case model.StatusCancelled, model.StatusArchived:
		return requested
	}
	today = today.UTC().Truncate(24 * time.Hour)
	expires = expires.UTC().Truncate(24 * time.Hour)
	if expires.Before(today) {
		return model.StatusExpired
	}
	until := today.AddDate(0, 0, notifyDays)
	if !expires.After(until) {
		return model.StatusExpiring
	}
	return model.StatusActive
}
