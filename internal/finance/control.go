package finance

func (model Model) Update(key string) Model {
	switch key {
	case "left":
		return model.moveMonth(-1)
	case "right":
		return model.moveMonth(1)
	case "[":
		return model.cycleCategory(-1)
	case "]":
		return model.cycleCategory(1)
	case "0":
		return model.clearCategory()
	case "c":
		model.showComparison = !model.showComparison
		return model
	case "r":
		model.load()
		return model
	default:
		return model
	}
}

func (model Model) moveMonth(distance int) Model {
	model.selectedMonth = monthStart(model.selectedMonth).AddDate(0, distance, 0)
	model.load()
	return model
}

func (model Model) cycleCategory(distance int) Model {
	count := len(model.categories) + 1
	model.categoryIndex = (model.categoryIndex + distance + count) % count
	model.load()
	return model
}

func (model Model) clearCategory() Model {
	model.categoryIndex = 0
	model.load()
	return model
}
