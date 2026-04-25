package calendar

type Model struct{}

func New() Model { return Model{} }

func (m Model) LeftView(width, height int, focused bool) string {
	return "09:00 Standup\n14:00 1:1\n16:00 Team sync"
}

func (m Model) RightView(width, height int) string {
	return "Fri 25 Apr 2026\n─────────────────\n\n09:00  Standup (30m)\n10:00  ·\n14:00  1:1 with manager (1h)\n15:00  ·\n16:00  Team sync"
}
