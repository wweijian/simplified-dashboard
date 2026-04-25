package finance

type Model struct{}

func New() Model { return Model{} }

func (m Model) LeftView(width, height int, focused bool) string {
	return "Apr: -$1,234\nFood:   -$420\nIncome: +$5,000"
}

func (m Model) RightView(width, height int) string {
	return "Recent Transactions  /: filter  s: sort\n─────────────────\n\n25 Apr  Uber Eats      -$24.50  Food\n24 Apr  NTUC FairPrice -$67.20  Food\n24 Apr  Grab           -$12.00  Transport\n23 Apr  Netflix        -$15.98  Subscriptions\n22 Apr  Salary        +$5,000   Income"
}
