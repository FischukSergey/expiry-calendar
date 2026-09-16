package service

import (
	"time"

	"duekeep/internal/model"
)

// StatusAtWrite считает active/expiring/expired. cancelled/archived/paid не пересчитывает.
// notifyDays == nil — без окна expiring (active до дня срока, затем expired).
func StatusAtWrite(today, expires time.Time, notifyDays *int, requested string) string {
	switch requested {
	case model.StatusCancelled, model.StatusArchived, model.StatusPaid:
		return requested
	}
	today = today.UTC().Truncate(24 * time.Hour)
	expires = expires.UTC().Truncate(24 * time.Hour)
	if expires.Before(today) {
		return model.StatusExpired
	}
	if notifyDays == nil {
		return model.StatusActive
	}
	until := today.AddDate(0, 0, *notifyDays)
	if !expires.After(until) {
		return model.StatusExpiring
	}
	return model.StatusActive
}

// statusFromOccurrences — active/expiring/expired от ближайшего open, не сырой expires_at.
func statusFromOccurrences(it model.Item, today time.Time, paid map[string]struct{}) (string, error) {
	switch it.Status {
	case model.StatusCancelled, model.StatusArchived, model.StatusPaid:
		return it.Status, nil
	}
	due, ok, err := nextUnpaidOccurrence(it, today, paid)
	if err != nil {
		return "", err
	}
	if !ok {
		return model.StatusActive, nil
	}
	return StatusAtWrite(today, due, it.NotifyBeforeDays, ""), nil
}
