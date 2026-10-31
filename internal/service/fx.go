package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dimakohanskyi/financist/internal/model"
)

// fxQuotes are the currencies rates are kept for, besides the USD base.
var fxQuotes = []string{"EUR", "NOK", "PLN", "UAH"}

const (
	// Frankfurter serves ECB reference rates by day, history included.
	frankfurterURL = "https://api.frankfurter.dev/v1/"
	// The ECB doesn't publish UAH, so it comes from here: latest only, no history.
	erAPIURL  = "https://open.er-api.com/v6/latest/USD"
	fxSource  = "Frankfurter (ECB); UAH from open.er-api.com"
	fxGapDays = 7 // a transaction is covered by a rate this recent: weekends and holidays have none
)

type FxStore interface {
	Save(ctx context.Context, rates []model.FxRate, fetchedAt time.Time) error
	All(ctx context.Context) ([]model.FxRate, *time.Time, error)
}

type FxService struct {
	store FxStore
	txs   interface {
		Ledger(ctx context.Context) ([]model.Transaction, error)
	}
	client *http.Client

	mu        sync.Mutex // guards the cache below
	cache     model.Rates
	updatedAt *time.Time
}

func NewFxService(store FxStore, txs LedgerSource) *FxService {
	return &FxService{store: store, txs: txs, client: &http.Client{Timeout: 15 * time.Second}}
}

// Rates is a snapshot of every stored rate; offline it is whatever was fetched last.
func (s *FxService) Rates(ctx context.Context) (model.Rates, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cache == nil {
		if err := s.load(ctx); err != nil {
			return nil, err
		}
	}
	return s.cache, nil
}

func (s *FxService) load(ctx context.Context) error {
	all, updated, err := s.store.All(ctx)
	if err != nil {
		return err
	}
	s.cache = model.Rates{}
	for _, r := range all {
		s.cache[r.Quote] = append(s.cache[r.Quote], r)
	}
	s.updatedAt = updated
	return nil
}

// Status is the newest rate of each quote.
func (s *FxService) Status(ctx context.Context) (model.FxStatus, error) {
	rates, err := s.Rates(ctx)
	st := model.FxStatus{Source: fxSource}
	if err != nil {
		return st, err
	}
	s.mu.Lock()
	st.UpdatedAt = s.updatedAt
	s.mu.Unlock()
	for _, q := range fxQuotes {
		if pts := rates[q]; len(pts) > 0 {
			st.Rates = append(st.Rates, pts[len(pts)-1])
		}
	}
	return st, nil
}

// Refresh fetches the latest rates plus the history behind transactions that
// have no recent rate yet. It keeps what it could fetch and reports the rest.
func (s *FxService) Refresh(ctx context.Context) error {
	today := model.Day(time.Now())
	start, err := s.historyStart(ctx, today)
	if err != nil {
		return err
	}
	var errs []error
	rates, err := s.frankfurter(ctx, start)
	if err != nil {
		errs = append(errs, err)
	}
	have := map[string]bool{}
	for _, r := range rates {
		have[r.Quote] = true
	}
	var missing []string
	for _, q := range fxQuotes {
		if !have[q] {
			missing = append(missing, q)
		}
	}
	if len(missing) > 0 {
		more, err := s.erAPI(ctx, missing, today)
		if err != nil {
			errs = append(errs, err)
		}
		rates = append(rates, more...)
	}
	if err := s.store.Save(ctx, rates, time.Now()); err != nil {
		return err
	}
	s.mu.Lock()
	err = s.load(ctx)
	s.mu.Unlock()
	return errors.Join(append(errs, err)...)
}

// historyStart is the day to fetch from: the oldest transaction without a rate
// within fxGapDays before it, else a week back (the latest rate, past a weekend).
func (s *FxService) historyStart(ctx context.Context, today time.Time) (time.Time, error) {
	start := today.AddDate(0, 0, -fxGapDays)
	rates, err := s.Rates(ctx)
	if err != nil {
		return start, err
	}
	ledger, err := s.txs.Ledger(ctx)
	if err != nil {
		return start, err
	}
	pts := rates[fxQuotes[0]] // quotes arrive together, so one stands for all
	for _, t := range ledger {
		day := model.Day(t.Date)
		if day.After(today) || !day.Before(start) {
			continue
		}
		i := sort.Search(len(pts), func(i int) bool { return pts[i].Date.After(day) })
		if i == 0 || day.Sub(pts[i-1].Date) > fxGapDays*24*time.Hour {
			start = day
		}
	}
	return start, nil
}

func (s *FxService) get(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", url, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// frankfurter returns every published day from start through the latest.
func (s *FxService) frankfurter(ctx context.Context, start time.Time) ([]model.FxRate, error) {
	var body struct {
		Rates map[string]map[string]float64
	}
	url := fmt.Sprintf("%s%s..?base=USD&symbols=%s", frankfurterURL, start.Format(time.DateOnly), strings.Join(fxQuotes, ","))
	if err := s.get(ctx, url, &body); err != nil {
		return nil, err
	}
	var out []model.FxRate
	for day, quotes := range body.Rates {
		d, err := time.Parse(time.DateOnly, day)
		if err != nil {
			return nil, err
		}
		for q, rate := range quotes {
			out = append(out, model.FxRate{Date: d, Quote: q, Rate: rate})
		}
	}
	return out, nil
}

// erAPI returns today's rate of each wanted quote.
func (s *FxService) erAPI(ctx context.Context, quotes []string, today time.Time) ([]model.FxRate, error) {
	var body struct {
		Rates map[string]float64
	}
	if err := s.get(ctx, erAPIURL, &body); err != nil {
		return nil, err
	}
	var out []model.FxRate
	for _, q := range quotes {
		if rate := body.Rates[q]; rate > 0 {
			out = append(out, model.FxRate{Date: today, Quote: q, Rate: rate})
		}
	}
	return out, nil
}

type usdKey struct{}

// WithUSD marks a request whose totals should come back in USD.
func WithUSD(ctx context.Context) context.Context { return context.WithValue(ctx, usdKey{}, true) }

func wantsUSD(ctx context.Context) bool {
	on, _ := ctx.Value(usdKey{}).(bool)
	return on
}

// Wrap returns src with amounts converted to USD at each transaction's date,
// but only for requests marked with WithUSD. Otherwise amounts stay as stored.
func (s *FxService) Wrap(src LedgerSource) LedgerSource { return usdSource{src, s} }

type usdSource struct {
	src LedgerSource
	fx  *FxService
}

func (w usdSource) Ledger(ctx context.Context) ([]model.Transaction, error) {
	ts, err := w.src.Ledger(ctx)
	if err != nil || !wantsUSD(ctx) {
		return ts, err
	}
	rates, err := w.fx.Rates(ctx)
	if err != nil {
		return nil, err
	}
	for i := range ts {
		ts[i].Amount = rates.ToBase(ts[i].Amount, ts[i].Currency, ts[i].Date)
	}
	return ts, nil
}

func (w usdSource) Contributions(ctx context.Context, f model.ContributionFilter) ([]model.Contribution, error) {
	cs, err := w.src.Contributions(ctx, f)
	if err != nil || !wantsUSD(ctx) {
		return cs, err
	}
	rates, err := w.fx.Rates(ctx)
	if err != nil {
		return nil, err
	}
	for i := range cs {
		cs[i].Amount = rates.ToBase(cs[i].Amount, cs[i].Currency, cs[i].Date)
	}
	return cs, nil
}
