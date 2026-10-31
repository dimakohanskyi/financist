package model

import (
	"testing"
	"time"
)

func TestConvert(t *testing.T) {
	d := func(day int) time.Time { return time.Date(2024, 3, day, 0, 0, 0, 0, time.UTC) }
	rates := Rates{
		"NOK": {{Date: d(1), Quote: "NOK", Rate: 10}, {Date: d(5), Quote: "NOK", Rate: 20}},
		"EUR": {{Date: d(1), Quote: "EUR", Rate: 0.5}},
	}
	// 5000 NOK on the 3rd uses the rate from the 1st: 500 USD
	if got := rates.Convert(5000, "NOK", "USD", d(3)); got != 500 {
		t.Fatalf("NOK to USD: %d", got)
	}
	// 100 USD to EUR, and NOK to EUR through the USD pivot
	if got := rates.Convert(100, "USD", "EUR", d(9)); got != 50 {
		t.Fatalf("USD to EUR: %d", got)
	}
	if got := rates.Convert(2000, "NOK", "EUR", d(9)); got != 50 {
		t.Fatalf("NOK to EUR: %d", got)
	}
}
