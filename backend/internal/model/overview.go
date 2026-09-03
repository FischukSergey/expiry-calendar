package model

// Dashboard — GET /dashboard. Суммы по валютам отдельно, без конвертации.
type Dashboard struct {
	Counts             DashboardCounts `json:"counts"`
	UpcomingCost       []UpcomingCost  `json:"upcoming_cost"`
	ExpirationsByMonth []MonthCount    `json:"expirations_by_month"`
	CostByKind         []KindCost      `json:"cost_by_kind"`
	Soonest            []DashboardItem `json:"soonest"`
}

// DashboardCounts — KPI. expiring_7/30 по дате, active/expired по полю status.
type DashboardCounts struct {
	Active     int `json:"active"`
	Expiring7  int `json:"expiring_7"`
	Expiring30 int `json:"expiring_30"`
	Expired    int `json:"expired"`
}

// UpcomingCost — run-rate одной валюты (monthly/yearly из billing_period).
type UpcomingCost struct {
	Currency string `json:"currency"`
	Monthly  int    `json:"monthly"`
	Yearly   int    `json:"yearly"`
}

// MonthCount — сроки в календарном месяце UTC: число записей и суммы по валютам.
type MonthCount struct {
	Month   string           `json:"month"`
	Count   int              `json:"count"`
	Amounts []CurrencyAmount `json:"amounts"`
}

// CurrencyAmount — сумма cost_amount в одной валюте, без конвертации.
type CurrencyAmount struct {
	Currency string `json:"currency"`
	Amount   int    `json:"amount"`
}

// KindCost — сумма cost_amount по типу и валюте (как записано, без периода).
type KindCost struct {
	KindID   string `json:"kind_id"`
	Currency string `json:"currency"`
	Amount   int    `json:"amount"`
}

// DashboardItem — краткая карточка soonest.
type DashboardItem struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	ExpiresAt string `json:"expires_at"`
	Status    string `json:"status"`
	KindID    string `json:"kind_id"`
}

// Calendar — GET /calendar. Пустые дни не включаем.
type Calendar struct {
	Year  int           `json:"year"`
	Month int           `json:"month"`
	Days  []CalendarDay `json:"days"`
}

// CalendarDay — обязательства одного дня.
type CalendarDay struct {
	Date  string         `json:"date"`
	Items []CalendarItem `json:"items"`
}

// CalendarItem — точка на дне календаря. Status — запись; OccurrenceStatus — этот день.
type CalendarItem struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Status           string `json:"status"`
	OccurrenceStatus string `json:"occurrence_status"`
	CostAmount       int    `json:"cost_amount"`
	Currency         string `json:"currency"`
}
