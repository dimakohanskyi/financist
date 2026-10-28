package service

import (
	"context"
	"fmt"
	"math"
	"slices"
	"sort"
	"time"

	"github.com/dimakohanskyi/financist/internal/model"
)

type LedgerSource interface {
	ContributionSource
	Ledger(ctx context.Context) ([]model.Transaction, error)
}

type AccountLister interface {
	List(ctx context.Context, includeArchived bool) ([]model.Account, error)
}

type AssetLister interface {
	List(ctx context.Context) ([]model.Asset, error)
	Valuations(ctx context.Context, assetID *int64) ([]model.AssetValuation, error)
}

type SettingsGetter interface {
	Get(ctx context.Context) (model.Settings, error)
}

// ReportService computes read-only summaries. Transfers between your own
// accounts are never income or spending.
// ponytail: reads whole tables into memory, fine for one household's
// history; move the sums into SQL if it gets slow.
type ReportService struct {
	txs        LedgerSource
	accounts   AccountLister
	categories CategoryLister
	assets     AssetLister
	settings   SettingsGetter
	budgets    *BudgetService
}

func NewReportService(txs LedgerSource, accounts AccountLister, categories CategoryLister, assets AssetLister, settings SettingsGetter, budgets *BudgetService) *ReportService {
	return &ReportService{txs, accounts, categories, assets, settings, budgets}
}

// chartSlices is how many categories the dashboard shows before folding the rest into one.
const chartSlices = 8

// Dashboard covers every account, or just accountID's own transactions when set.
// Transfers into that account aren't income, same as across all accounts.
func (s *ReportService) Dashboard(ctx context.Context, year, month int, accountID *int64) (model.Dashboard, error) {
	d := model.Dashboard{Year: year, Month: month}
	ledger, err := s.txs.Ledger(ctx)
	if err != nil {
		return d, err
	}
	if accountID != nil {
		ledger = slices.DeleteFunc(ledger, func(t model.Transaction) bool { return t.AccountID != *accountID })
	}
	totals := monthlyTotals(ledger)[[2]int{year, month}]
	d.Income, d.Spent, d.TransferredOut = totals.income, totals.expense, totals.transferred

	from, to := model.MonthRange(year, month)
	items, err := s.rollup(ctx, model.ContributionFilter{Type: model.TransactionExpense, From: &from, To: &to, AccountID: accountID})
	if err != nil {
		return d, err
	}
	if len(items) > chartSlices {
		var other model.CategoryAmount
		for _, it := range items[chartSlices:] {
			other.Amount += it.Amount
			other.TransactionCount += it.TransactionCount
		}
		items = append(items[:chartSlices], other)
	}
	for i := range items {
		items[i].Percent = model.Percent(items[i].Amount, d.Spent)
	}
	d.SpendingByCategory = items
	return d, nil
}

// CashFlow is income and expense per month; without bounds it spans the data.
func (s *ReportService) CashFlow(ctx context.Context, from, to *time.Time) (model.CashFlow, error) {
	cf := model.CashFlow{From: from, To: to}
	ledger, err := s.txs.Ledger(ctx)
	if err != nil {
		return cf, err
	}
	var kept []model.Transaction
	for _, t := range ledger {
		if t.Type != model.TransactionTransfer && within(t.Date, from, to) {
			kept = append(kept, t)
		}
	}
	if len(kept) == 0 {
		return cf, nil
	}
	if cf.From == nil {
		cf.From = &kept[0].Date
	}
	if cf.To == nil {
		cf.To = &kept[len(kept)-1].Date
	}
	totals := monthlyTotals(kept)
	eachMonth(*cf.From, *cf.To, func(y, m int) {
		t := totals[[2]int{y, m}]
		cf.Months = append(cf.Months, model.CashFlowMonth{Year: y, Month: m, Income: t.income, Expense: t.expense})
	})
	return cf, nil
}

// CategorySpending is one category's total per month. A top-level category
// includes its subcategories.
func (s *ReportService) CategorySpending(ctx context.Context, categoryID int64, from, to *time.Time) (model.CategorySpending, error) {
	out := model.CategorySpending{CategoryID: categoryID, From: from, To: to}
	c, err := s.categories.GetByID(ctx, categoryID)
	if err != nil {
		return out, err
	}
	ids := map[int64]bool{categoryID: true}
	if c.ParentID == nil {
		all, err := s.categories.List(ctx, nil)
		if err != nil {
			return out, err
		}
		for _, child := range all {
			if child.ParentID != nil && *child.ParentID == categoryID {
				ids[child.ID] = true
			}
		}
	}
	contributions, err := s.txs.Contributions(ctx, model.ContributionFilter{From: from, To: to})
	if err != nil {
		return out, err
	}

	byMonth := map[[2]int]int64{}
	txs := map[int64]bool{} // a transaction split across two of these counts once
	var first, last time.Time
	for _, c := range contributions {
		if !ids[c.CategoryID] {
			continue
		}
		out.Total += c.Amount
		txs[c.TransactionID] = true
		byMonth[[2]int{c.Date.Year(), int(c.Date.Month())}] += c.Amount
		if first.IsZero() || c.Date.Before(first) {
			first = c.Date
		}
		if c.Date.After(last) {
			last = c.Date
		}
	}
	if len(txs) == 0 {
		return out, nil
	}
	if out.From == nil {
		out.From = &first
	}
	if out.To == nil {
		out.To = &last
	}
	eachMonth(*out.From, *out.To, func(y, m int) {
		out.Months = append(out.Months, model.MonthAmount{Year: y, Month: m, Amount: byMonth[[2]int{y, m}]})
	})
	out.TransactionCount = len(txs)
	out.AveragePerMonth = out.Total / int64(len(out.Months))
	return out, nil
}

// CategoryRanking ranks top-level categories of one kind by total over the period.
func (s *ReportService) CategoryRanking(ctx context.Context, kind string, from, to *time.Time) (model.CategoryRanking, error) {
	r := model.CategoryRanking{From: from, To: to}
	// a category's kind matches its transactions' type, so the type picks the kind
	items, err := s.rollup(ctx, model.ContributionFilter{Type: kind, From: from, To: to})
	if err != nil {
		return r, err
	}
	for _, it := range items {
		r.Total += it.Amount
	}
	for i := range items {
		items[i].Percent = model.Percent(items[i].Amount, r.Total)
	}
	r.Items = items
	return r, nil
}

// rollup totals contributions per top-level category, biggest first.
func (s *ReportService) rollup(ctx context.Context, f model.ContributionFilter) ([]model.CategoryAmount, error) {
	contributions, err := s.txs.Contributions(ctx, f)
	if err != nil {
		return nil, err
	}
	categories, err := s.categories.List(ctx, nil)
	if err != nil {
		return nil, err
	}
	byID := map[int64]model.Category{}
	for _, c := range categories {
		byID[c.ID] = c
	}

	type total struct {
		amount int64
		txs    map[int64]bool
		leaves map[int64]int64
	}
	totals := map[int64]*total{}
	for _, c := range contributions {
		top := c.CategoryID
		if parent := byID[top].ParentID; parent != nil {
			top = *parent
		}
		t := totals[top]
		if t == nil {
			t = &total{txs: map[int64]bool{}, leaves: map[int64]int64{}}
			totals[top] = t
		}
		t.amount += c.Amount
		t.txs[c.TransactionID] = true
		t.leaves[c.CategoryID] += c.Amount
	}

	out := make([]model.CategoryAmount, 0, len(totals))
	for id, t := range totals {
		item := model.CategoryAmount{CategoryID: &id, Amount: t.amount, TransactionCount: len(t.txs)}
		if len(t.leaves) > 1 {
			for leaf, amount := range t.leaves {
				item.Children = append(item.Children, model.CategoryAmount{CategoryID: &leaf, Amount: amount, Percent: model.Percent(amount, t.amount)})
			}
			sort.Slice(item.Children, func(i, j int) bool { return item.Children[i].Amount > item.Children[j].Amount })
		}
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Amount != out[j].Amount {
			return out[i].Amount > out[j].Amount
		}
		return byID[*out[i].CategoryID].SortOrder < byID[*out[j].CategoryID].SortOrder
	})
	return out, nil
}

// NetWorthRanges maps a range name to its length in days; any other name means all history.
var NetWorthRanges = map[string]int{"last_30_days": 30, "last_90_days": 90, "last_year": 365, "last_5_years": 5 * 365}

// cashAccountTypes count towards net worth as cash. An investment account's
// balance is money not yet invested; what's invested is valued as an asset.
var cashAccountTypes = map[string]bool{"checking": true, "debit_card": true, "savings": true, "cash": true, "investment": true}

var netWorthParts = []struct{ key, name string }{
	{"cash", "Cash"},
	{"investments", "Investments"},
	{"real_estate", "Real estate"},
	{"vehicles", "Vehicles"},
	{"precious_metals", "Precious metals"},
	{"other", "Other"},
}

// NetWorth is cash (from cash account balances) plus the latest value of
// every asset, as a daily series over the range and as breakdowns of today.
func (s *ReportService) NetWorth(ctx context.Context, rangeName string) (model.NetWorth, error) {
	var nw model.NetWorth
	accounts, err := s.accounts.List(ctx, true)
	if err != nil {
		return nw, err
	}
	ledger, err := s.txs.Ledger(ctx)
	if err != nil {
		return nw, err
	}
	assets, err := s.assets.List(ctx)
	if err != nil {
		return nw, err
	}
	valuations, err := s.assets.Valuations(ctx, nil)
	if err != nil {
		return nw, err
	}

	cash := map[int64]bool{}
	for _, a := range accounts {
		if cashAccountTypes[a.AccountType] {
			cash[a.ID] = true
		}
	}
	delta := map[time.Time]int64{}
	for _, t := range ledger {
		day := model.Day(t.Date)
		switch t.Type {
		case model.TransactionIncome:
			if cash[t.AccountID] {
				delta[day] += t.Amount
			}
		case model.TransactionExpense:
			if cash[t.AccountID] {
				delta[day] -= t.Amount
			}
		case model.TransactionTransfer:
			if cash[t.AccountID] {
				delta[day] -= t.Amount
			}
			if t.ToAccountID != nil && cash[*t.ToAccountID] {
				delta[day] += t.Amount
			}
		}
	}
	cashEvents := runningTotals(delta)

	// after each valuation day: the sum of every asset's latest value so far
	latest := map[int64]int64{}
	var assetEvents []model.DatedAmount
	for _, v := range valuations {
		latest[v.AssetID] = v.Value
		var sum int64
		for _, value := range latest {
			sum += value
		}
		day := model.Day(v.AsOfDate)
		if n := len(assetEvents); n > 0 && assetEvents[n-1].Date.Equal(day) {
			assetEvents[n-1].Amount = sum
		} else {
			assetEvents = append(assetEvents, model.DatedAmount{Date: day, Amount: sum})
		}
	}

	today := model.Day(time.Now())
	start := today
	if days, ok := NetWorthRanges[rangeName]; ok {
		start = today.AddDate(0, 0, -(days - 1))
	} else {
		for _, e := range append(append([]model.DatedAmount{}, cashEvents...), assetEvents...) {
			if e.Date.Before(start) {
				start = e.Date
			}
		}
	}
	cashSeries, assetSeries := daily(cashEvents, start, today), daily(assetEvents, start, today)
	nw.Series = make([]model.DatedAmount, len(cashSeries))
	for i := range cashSeries {
		nw.Series[i] = model.DatedAmount{Date: cashSeries[i].Date, Amount: cashSeries[i].Amount + assetSeries[i].Amount}
	}
	nw.Current = nw.Series[len(nw.Series)-1].Amount
	opening := nw.Series[0].Amount
	nw.ChangeAmount = nw.Current - opening
	if opening != 0 {
		p := model.Percent(nw.ChangeAmount, opening)
		nw.ChangePercent = &p
	}

	var cashToday int64
	if len(cashEvents) > 0 {
		cashToday = cashEvents[len(cashEvents)-1].Amount
	}
	byClass := map[string]int64{"cash": cashToday}
	total := cashToday
	for _, a := range assets {
		byClass[a.AssetClass] += a.CurrentValue
		total += a.CurrentValue
	}
	for _, part := range netWorthParts {
		nw.Breakdown = append(nw.Breakdown, model.NetWorthPart{Key: part.key, Name: part.name, Amount: byClass[part.key], Percent: model.Percent(byClass[part.key], total)})
	}

	for _, role := range []string{model.CapitalIncome, model.CapitalNeutral, model.CapitalDrain} {
		sum := model.CapitalRoleSummary{Role: role}
		for _, a := range assets {
			if a.CapitalRole == role {
				sum.TotalValue += a.CurrentValue
				if a.MonthlyCashFlow != nil {
					sum.MonthlyCashFlow += *a.MonthlyCashFlow
				}
				sum.Count++
			}
		}
		nw.CapitalRoles = append(nw.CapitalRoles, sum)
	}

	// cash is the zero-risk anchor an 80/20-style rule measures against
	tiers := map[string]*model.RiskLevelSummary{}
	for _, level := range []string{model.RiskLow, model.RiskMedium, model.RiskHigh} {
		tiers[level] = &model.RiskLevelSummary{Level: level}
	}
	if cashToday != 0 {
		tiers[model.RiskLow].Items = append(tiers[model.RiskLow].Items, model.NetWorthPart{Key: "cash", Name: "Cash", Amount: cashToday})
	}
	for _, a := range assets {
		if a.CurrentValue != 0 {
			tiers[a.RiskLevel].Items = append(tiers[a.RiskLevel].Items, model.NetWorthPart{Key: fmt.Sprintf("asset:%d", a.ID), Name: a.Name, Amount: a.CurrentValue})
		}
	}
	var capital int64
	for _, tier := range tiers {
		for _, it := range tier.Items {
			tier.TotalValue += it.Amount
		}
		capital += tier.TotalValue
	}
	for _, level := range []string{model.RiskLow, model.RiskMedium, model.RiskHigh} {
		tier := tiers[level]
		tier.Percent = model.Percent(tier.TotalValue, capital)
		sort.Slice(tier.Items, func(i, j int) bool { return tier.Items[i].Amount > tier.Items[j].Amount })
		for i := range tier.Items {
			tier.Items[i].Percent = model.Percent(tier.Items[i].Amount, tier.TotalValue)
		}
		nw.RiskLevels = append(nw.RiskLevels, *tier)
	}
	return nw, nil
}

// maxLookbackMonths caps how far back a streak is counted.
const maxLookbackMonths = 24

// Alerts are early warnings. Streaks look at complete months only, so rent
// paid before salary lands never alarms; budgets look at the current month,
// while there's still time to react.
func (s *ReportService) Alerts(ctx context.Context) ([]model.Notice, error) {
	st, err := s.settings.Get(ctx)
	if err != nil {
		return nil, err
	}
	ledger, err := s.txs.Ledger(ctx)
	if err != nil {
		return nil, err
	}
	var alerts []model.Notice
	warn := func(key string, params map[string]any) {
		alerts = append(alerts, model.Notice{Key: key, Level: "warning", Params: params})
	}
	now := time.Now().UTC()
	year, month := now.Year(), int(now.Month())

	monthly := monthlyTotals(ledger)
	streak := 0
	for y, m, i := year, month, 0; i < maxLookbackMonths; i++ {
		y, m = previousMonth(y, m)
		if t := monthly[[2]int{y, m}]; t.income-t.expense >= 0 {
			break
		}
		streak++
	}
	if streak >= st.NegativeCashFlowThresholdMonths {
		warn("negative_cash_flow_streak", map[string]any{"months": streak})
	}

	nw, err := s.NetWorth(ctx, "all")
	if err != nil {
		return nil, err
	}
	if decline := declineStreak(nw.Series, year, month); decline >= st.NetWorthDeclineThresholdMonths {
		warn("net_worth_decline_streak", map[string]any{"months": decline})
	}
	var risky float64
	for _, tier := range nw.RiskLevels {
		if tier.Level != model.RiskLow {
			risky += tier.Percent
		}
	}
	if risky > float64(st.RiskyAllocationThresholdPercent) {
		warn("risky_allocation_exceeded", map[string]any{"percent": int(math.Round(risky)), "threshold": st.RiskyAllocationThresholdPercent})
	}

	statuses, err := s.budgets.Status(ctx, year, month)
	if err != nil {
		return nil, err
	}
	over := 0
	for _, b := range statuses {
		if b.Spent > b.MonthlyLimit {
			over++
		}
	}
	if over > 0 {
		warn("budget_exceeded", map[string]any{"count": over})
	}

	accounts, err := s.accounts.List(ctx, false)
	if err != nil {
		return nil, err
	}
	if idle := idleCashAccounts(accounts, ledger, st); idle > 0 {
		warn("idle_cash", map[string]any{"count": idle, "days": st.IdleCashThresholdDays})
	}
	return alerts, nil
}

const (
	trailingMonths          = 3
	risingThresholdPercent  = 25
	savingsRateSwingPercent = 5
)

// Advice is a few non-urgent observations about this month's spending.
func (s *ReportService) Advice(ctx context.Context) ([]model.Notice, error) {
	now := time.Now().UTC()
	year, month := now.Year(), int(now.Month())
	categories, err := s.categories.List(ctx, nil)
	if err != nil {
		return nil, err
	}
	names := map[int64]string{}
	for _, c := range categories {
		names[c.ID] = c.Name
	}
	spending := func(y, m int) (map[int64]int64, error) {
		from, to := model.MonthRange(y, m)
		cs, err := s.txs.Contributions(ctx, model.ContributionFilter{Type: model.TransactionExpense, From: &from, To: &to})
		out := map[int64]int64{}
		for _, c := range cs {
			out[c.CategoryID] += c.Amount
		}
		return out, err
	}
	current, err := spending(year, month)
	if err != nil {
		return nil, err
	}
	var advice []model.Notice

	// the category furthest above its own trailing average
	trailing := map[int64]int64{}
	for y, m, i := year, month, 0; i < trailingMonths; i++ {
		y, m = previousMonth(y, m)
		past, err := spending(y, m)
		if err != nil {
			return nil, err
		}
		for id, amount := range past {
			trailing[id] += amount
		}
	}
	var rising *model.Notice
	var best float64
	for id, amount := range current {
		average := trailing[id] / trailingMonths
		if average <= 0 {
			continue
		}
		if increase := model.Percent(amount-average, average); increase >= risingThresholdPercent && increase > best {
			best = increase
			rising = &model.Notice{Key: "rising_category", Level: "warning", Params: map[string]any{
				"category": names[id], "percent": int(math.Round(increase)), "current": amount, "average": average,
			}}
		}
	}
	if rising != nil {
		advice = append(advice, *rising)
	}

	// the month's biggest category with no budget
	budgets, err := s.budgets.List(ctx)
	if err != nil {
		return nil, err
	}
	budgeted := map[int64]bool{}
	for _, b := range budgets {
		budgeted[b.CategoryID] = true
	}
	var top int64
	var topAmount int64
	for id, amount := range current {
		if !budgeted[id] && (amount > topAmount || amount == topAmount && id < top) {
			top, topAmount = id, amount
		}
	}
	if topAmount > 0 {
		advice = append(advice, model.Notice{Key: "unbudgeted_top_category", Level: "neutral", Params: map[string]any{"category": names[top], "amount": topAmount}})
	}

	// savings rate against last month, when it moved enough to mention
	ledger, err := s.txs.Ledger(ctx)
	if err != nil {
		return nil, err
	}
	monthly := monthlyTotals(ledger)
	py, pm := previousMonth(year, month)
	this, last := monthly[[2]int{year, month}], monthly[[2]int{py, pm}]
	if this.income > 0 && last.income > 0 {
		rate := model.Percent(this.income-this.expense, this.income)
		diff := rate - model.Percent(last.income-last.expense, last.income)
		if math.Abs(diff) >= savingsRateSwingPercent {
			tone := "positive"
			if diff < 0 {
				tone = "warning"
			}
			advice = append(advice, model.Notice{Key: "savings_rate_trend", Level: tone, Params: map[string]any{"rate": int(math.Round(rate)), "diff": int(math.Round(diff))}})
		}
	}
	return advice, nil
}

type monthTotal struct{ income, expense, transferred int64 }

func monthlyTotals(ledger []model.Transaction) map[[2]int]monthTotal {
	out := map[[2]int]monthTotal{}
	for _, t := range ledger {
		d := t.Date.UTC()
		key := [2]int{d.Year(), int(d.Month())}
		total := out[key]
		switch t.Type {
		case model.TransactionIncome:
			total.income += t.Amount
		case model.TransactionExpense:
			total.expense += t.Amount
		case model.TransactionTransfer:
			total.transferred += t.Amount
		}
		out[key] = total
	}
	return out
}

// idleCashAccounts counts cash accounts at or above the threshold balance
// with nothing touching them for the threshold's days.
func idleCashAccounts(accounts []model.Account, ledger []model.Transaction, st model.Settings) int {
	balances := map[int64]int64{}
	lastActivity := map[int64]time.Time{}
	for _, t := range ledger {
		switch t.Type {
		case model.TransactionIncome:
			balances[t.AccountID] += t.Amount
		default:
			balances[t.AccountID] -= t.Amount
		}
		lastActivity[t.AccountID] = t.Date // the ledger is oldest first
		if t.Type == model.TransactionTransfer && t.ToAccountID != nil {
			balances[*t.ToAccountID] += t.Amount
			lastActivity[*t.ToAccountID] = t.Date
		}
	}
	cutoff := model.Day(time.Now()).AddDate(0, 0, -st.IdleCashThresholdDays)
	idle := 0
	for _, a := range accounts {
		switch a.AccountType {
		case "checking", "debit_card", "savings", "cash":
		default:
			continue // a card balance isn't your cash, an investment account is already working
		}
		last, ok := lastActivity[a.ID]
		if balances[a.ID] >= st.IdleCashThresholdAmount && (!ok || !model.Day(last).After(cutoff)) {
			idle++
		}
	}
	return idle
}

// declineStreak counts the complete months in a row whose month-end value fell.
func declineStreak(series []model.DatedAmount, year, month int) int {
	var ends []int64
	var lastKey [2]int
	for _, p := range series {
		key := [2]int{p.Date.Year(), int(p.Date.Month())}
		if key == [2]int{year, month} {
			continue
		}
		if key != lastKey || len(ends) == 0 {
			ends = append(ends, p.Amount)
			lastKey = key
		} else {
			ends[len(ends)-1] = p.Amount
		}
	}
	streak := 0
	for i := len(ends) - 1; i > 0 && ends[i] < ends[i-1]; i-- {
		streak++
	}
	return streak
}

// runningTotals turns per-day changes into a sorted running total.
func runningTotals(delta map[time.Time]int64) []model.DatedAmount {
	days := make([]time.Time, 0, len(delta))
	for d := range delta {
		days = append(days, d)
	}
	sort.Slice(days, func(i, j int) bool { return days[i].Before(days[j]) })
	out := make([]model.DatedAmount, len(days))
	var running int64
	for i, d := range days {
		running += delta[d]
		out[i] = model.DatedAmount{Date: d, Amount: running}
	}
	return out
}

// daily forward-fills sorted running totals into one point per day; events
// before start become start's opening value.
func daily(events []model.DatedAmount, start, end time.Time) []model.DatedAmount {
	var out []model.DatedAmount
	var current int64
	i := 0
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		for i < len(events) && !events[i].Date.After(d) {
			current = events[i].Amount
			i++
		}
		out = append(out, model.DatedAmount{Date: d, Amount: current})
	}
	return out
}

func within(t time.Time, from, to *time.Time) bool {
	return (from == nil || !t.Before(*from)) && (to == nil || !t.After(*to))
}

// eachMonth calls fn for every month from from through to.
func eachMonth(from, to time.Time, fn func(year, month int)) {
	from, to = from.UTC(), to.UTC()
	for y, m := from.Year(), int(from.Month()); y < to.Year() || y == to.Year() && m <= int(to.Month()); {
		fn(y, m)
		if m++; m > 12 {
			y, m = y+1, 1
		}
	}
}

func previousMonth(year, month int) (int, int) {
	if month == 1 {
		return year - 1, 12
	}
	return year, month - 1
}
