package ui

func panelWidths(total int) (left, right int) {
	left = total * 40 / 100
	right = total - left - 1
	return
}
