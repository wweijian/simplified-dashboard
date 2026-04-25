package tasks

type Model struct{}

func New() Model { return Model{} }

func (m Model) LeftView(width, height int, focused bool) string {
	return "[ ] Buy milk\n[ ] PR review !\n[x] Deploy"
}

func (m Model) RightView(width, height int) string {
	return "PR review  ● High  Due: today\n─────────────────\n\nNotes:\n  Review auth module changes.\n  Check for SQL injection."
}
