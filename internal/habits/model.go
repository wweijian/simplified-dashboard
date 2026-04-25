package habits

type Model struct{}

func New() Model { return Model{} }

func (m Model) LeftView(width, height int, focused bool) string {
	return "[x] Exercise  🔥3\n[ ] Read\n[x] Meditate  🔥7"
}

func (m Model) RightView(width, height int) string {
	return "Read  daily · 0-day streak\n─────────────────\n\nLast 14 days:\n12 █ █ ░ █ █ ░ ░ █ ░ █ ░ ░ █ ░ 25\n\n8/14 days (57%)\nBest streak: 4 days (Apr 12–15)"
}
