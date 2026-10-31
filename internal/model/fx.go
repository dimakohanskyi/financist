package model

import (
	"math"
	"sort"
	"time"
)

// BaseCurrency is what every total is converted to.
const BaseCurrency = "USD"

// FxRate is the price of 1 USD in Quote on a day.
type FxRate struct {
	Date  time.Time
	Quote string
	Rate  float64
}

// Rates holds each quote's rates oldest first.
type Rates map[string][]FxRate

// rate is the quote's rate on or before date. Before its first rate (or
// offline since) the nearest one stands in; a currency with none counts 1:1.
func (r Rates) rate(cur string, date time.Time) float64 {
	pts := r[cur]
	if cur == BaseCurrency || len(pts) == 0 {
		return 1
	}
	i := sort.Search(len(pts), func(i int) bool { return pts[i].Date.After(date) })
	return pts[max(i-1, 0)].Rate
}

// Convert turns minor units of from into minor units of to, pivoting through USD.
func (r Rates) Convert(amount int64, from, to string, date time.Time) int64 {
	if from == to {
		return amount
	}
	return int64(math.Round(float64(amount) / r.rate(from, date) * r.rate(to, date)))
}

// ToBase converts to USD.
func (r Rates) ToBase(amount int64, from string, date time.Time) int64 {
	return r.Convert(amount, from, BaseCurrency, date)
}

// FxStatus is the newest rate of each quote and when rates were last fetched.
type FxStatus struct {
	Source    string
	UpdatedAt *time.Time // nil before the first successful fetch
	Rates     []FxRate
}
