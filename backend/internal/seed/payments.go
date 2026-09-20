package seed

import "fmt"

// paymentSeed — факт оплаты вхождения. paid_on = якорь expires_at минус monthsBack календарных месяцев.
type paymentSeed struct {
	n          int
	itemN      int
	monthsBack int
}

func paymentID(n int) string {
	return fmt.Sprintf("99999999-9999-9999-9999-9999999999%02d", n)
}

// paymentSeeds — прошлые monthly-вхождения и якорь у status=paid. Даты от today.
func paymentSeeds() []paymentSeed {
	return []paymentSeed{
		{1, 1, 1},
		{2, 1, 2},
		{3, 2, 5},
		{4, 2, 6},
		{5, 5, 1},
		{6, 6, 1},
		{7, 14, 0},
		{8, 14, 1},
		{9, 53, 1},
		{10, 55, 0},
	}
}
