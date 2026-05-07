package finance

func (model Model) currentSpend() float64 {
	if !model.SelectedCategory().ID.Valid {
		return model.summary.Expenses
	}
	if len(model.monthlyTrend) == 0 {
		return 0
	}
	return model.monthlyTrend[len(model.monthlyTrend)-1].Amount
}

func (model Model) previousSpend() float64 {
	if len(model.monthlyTrend) < 2 {
		return 0
	}
	return model.monthlyTrend[len(model.monthlyTrend)-2].Amount
}
