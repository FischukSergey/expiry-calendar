package service

import (
	"context"
	"net"
	"time"

	"duekeep/internal/model"
)

// NextUnpaidForTest открывает nextUnpaidOccurrence тестам внешнего пакета.
func NextUnpaidForTest(it model.Item, from time.Time, paid map[string]struct{}) (time.Time, bool, error) {
	return nextUnpaidOccurrence(it, from, paid)
}

// StatusFromOccurrencesForTest открывает пересчёт статуса тестам внешнего пакета.
func StatusFromOccurrencesForTest(it model.Item, today time.Time, paid map[string]struct{}) (string, error) {
	return statusFromOccurrences(it, today, paid)
}

// CSVSafeForTest открывает экранирование ячейки тестам внешнего пакета.
func CSVSafeForTest(s string) string {
	return csvSafe(s)
}

// SetLookupForTest подменяет резолв хоста в тесте SSRF.
func (p *Push) SetLookupForTest(fn func(context.Context, string) ([]net.IP, error)) {
	p.lookup = fn
}
