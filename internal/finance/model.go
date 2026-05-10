package finance

import (
	"database/sql"
	"time"
)

const (
	analyticsTransactionLimit = 8
	recentTransactionLimit    = 10
	topCategoryLimit          = 3
	trendMonthCount           = 6
)

var currentMonth = func() time.Time {
	return time.Now().In(time.Local)
}

type Model struct {
	store                 *Store
	selectedMonth         time.Time
	categoryIndex         int
	showComparison        bool
	categories            []Category
	transactionCategories []Category
	transactions          []Transaction
	selectedTransaction   int
	summary               MonthlySummary
	topCategories         []CategorySpend
	categoryBreakdown     []CategorySpend
	monthlyTrend          []MonthSpend
	err                   error
}

func New(store *Store) Model {
	model := Model{
		store:          store,
		selectedMonth:  monthStart(currentMonth()),
		showComparison: true,
	}
	model.load()
	return model
}

func (model *Model) load() {
	if err := model.store.EnsureDefaultCategories(); err != nil {
		model.err = err
		return
	}

	snapshot, err := model.loadSnapshot()
	if err != nil {
		model.err = err
		return
	}

	model.summary = snapshot.summary
	model.topCategories = snapshot.topCategories
	model.categories = snapshot.categories
	model.categoryBreakdown = snapshot.categoryBreakdown
	model.monthlyTrend = snapshot.monthlyTrend
	model.transactions = snapshot.transactions
	model.transactionCategories = snapshot.transactionCategories
	model.clampTransactionSelection()
	model.err = nil
}

type financeSnapshot struct {
	summary               MonthlySummary
	topCategories         []CategorySpend
	categories            []Category
	transactionCategories []Category
	categoryBreakdown     []CategorySpend
	monthlyTrend          []MonthSpend
	transactions          []Transaction
}

func (model Model) loadSnapshot() (financeSnapshot, error) {
	summary, err := model.store.MonthlySummary(model.selectedMonth)
	if err != nil {
		return financeSnapshot{}, err
	}

	topCategories, err := model.store.TopExpenseCategories(model.selectedMonth, topCategoryLimit)
	if err != nil {
		return financeSnapshot{}, err
	}

	return model.loadAnalyticsSnapshot(summary, topCategories)
}

func (model Model) loadAnalyticsSnapshot(summary MonthlySummary, topCategories []CategorySpend) (financeSnapshot, error) {
	transactionCategories, err := model.store.ListCategories()
	if err != nil {
		return financeSnapshot{}, err
	}

	categories, err := model.store.ListExpenseCategories()
	if err != nil {
		return financeSnapshot{}, err
	}

	category := model.selectedCategoryFrom(categories)
	breakdown, err := model.store.SpendingByCategory(model.selectedMonth)
	if err != nil {
		return financeSnapshot{}, err
	}

	trend, err := model.store.MonthlySpendTrend(model.selectedMonth, trendMonthCount, category)
	if err != nil {
		return financeSnapshot{}, err
	}

	transactions, err := model.store.TransactionsForMonth(model.selectedMonth, category, analyticsTransactionLimit)
	if err != nil {
		return financeSnapshot{}, err
	}
	return financeSnapshot{
		summary:               summary,
		topCategories:         topCategories,
		categories:            categories,
		transactionCategories: transactionCategories,
		categoryBreakdown:     breakdown,
		monthlyTrend:          trend,
		transactions:          transactions,
	}, nil
}

func (model Model) TransactionCategories() []Category {
	categories := make([]Category, len(model.transactionCategories))
	copy(categories, model.transactionCategories)
	return categories
}

func (model Model) DefaultTransactionDate(now time.Time) time.Time {
	if monthStart(now).Equal(model.selectedMonth) {
		return now
	}
	return model.selectedMonth
}

func (model Model) SelectedTransaction() (Transaction, bool) {
	if len(model.transactions) == 0 || model.selectedTransaction < 0 || model.selectedTransaction >= len(model.transactions) {
		return Transaction{}, false
	}
	return model.transactions[model.selectedTransaction], true
}

func (model Model) CreateTransaction(input CreateTransactionInput) Model {
	id, err := model.store.CreateTransaction(input)
	if err != nil {
		model.err = err
		return model
	}

	model.load()
	return model.selectTransaction(id)
}

func (model Model) UpdateTransaction(id int64, input UpdateTransactionInput) Model {
	if err := model.store.UpdateTransaction(id, input); err != nil {
		model.err = err
		return model
	}

	model.load()
	return model.selectTransaction(id)
}

func (model Model) DeleteSelectedTransaction() Model {
	transaction, ok := model.SelectedTransaction()
	if !ok {
		return model
	}

	if err := model.store.DeleteTransaction(transaction.ID); err != nil {
		model.err = err
		return model
	}

	model.load()
	model.clampTransactionSelection()
	return model
}

func (model *Model) clampTransactionSelection() {
	if model.selectedTransaction >= len(model.transactions) {
		model.selectedTransaction = len(model.transactions) - 1
	}
	if model.selectedTransaction < 0 {
		model.selectedTransaction = 0
	}
}

func (model Model) selectTransaction(id int64) Model {
	for i, transaction := range model.transactions {
		if transaction.ID == id {
			model.selectedTransaction = i
			break
		}
	}
	return model
}

func (model Model) SelectedCategory() CategoryFilter {
	return model.selectedCategoryFrom(model.categories)
}

func (model Model) selectedCategoryFrom(categories []Category) CategoryFilter {
	if model.categoryIndex <= 0 || model.categoryIndex > len(categories) {
		return CategoryFilter{Name: "All"}
	}

	category := categories[model.categoryIndex-1]
	return CategoryFilter{
		ID:   sqlNullInt64(category.ID),
		Name: category.Name,
	}
}

func sqlNullInt64(value int64) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: true}
}
