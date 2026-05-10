package ui

import (
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"

	finance "simplified-dashboard/internal/finance"

	bbt "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type addFinanceTransactionField int

const (
	addFinanceTransactionDate addFinanceTransactionField = iota
	addFinanceTransactionAmount
	addFinanceTransactionCategory
	addFinanceTransactionDescription
)

const (
	addFinanceTransactionFieldCount = 4
	maxFinanceDescriptionChars      = 200
)

type addFinanceTransactionForm struct {
	date          string
	amount        string
	description   string
	categories    []finance.Category
	categoryIndex int
	field         addFinanceTransactionField
	errorText     string
}

func newAddFinanceTransactionForm(categories []finance.Category, now time.Time) addFinanceTransactionForm {
	return addFinanceTransactionForm{
		date:       now.In(time.Local).Format("2006-01-02"),
		categories: categories,
	}
}

func newEditFinanceTransactionForm(transaction finance.Transaction, categories []finance.Category) addFinanceTransactionForm {
	form := addFinanceTransactionForm{
		date:        transaction.Date,
		amount:      formatFormAmount(transaction.Amount),
		description: nullableStringValue(transaction.Description),
		categories:  categories,
	}
	form.categoryIndex = form.categoryIndexFor(transaction.CategoryID)
	return form
}

func (form addFinanceTransactionForm) Update(msg bbt.KeyMsg) addFinanceTransactionForm {
	form.errorText = ""

	switch msg.String() {
	case "tab", "down":
		form.field = (form.field + 1) % addFinanceTransactionFieldCount
	case "shift+tab", "up":
		form.field = (form.field + addFinanceTransactionFieldCount - 1) % addFinanceTransactionFieldCount
	case "left":
		form = form.moveCategory(-1)
	case "right":
		form = form.moveCategory(1)
	case "backspace", "ctrl+h":
		form = form.deleteLastRune()
	case "ctrl+u":
		form = form.clearField()
	default:
		form = form.appendRunes(msg.Runes)
	}
	return form
}

func (form addFinanceTransactionForm) Submit() (addFinanceTransactionForm, finance.CreateTransactionInput, bool) {
	date := strings.TrimSpace(form.date)
	if !form.validateDate(date) {
		return form, finance.CreateTransactionInput{}, false
	}

	amount, ok := form.amountValue()
	if !ok {
		return form, finance.CreateTransactionInput{}, false
	}

	category, ok := form.selectedCategory()
	if !ok {
		form.errorText = "Category is required"
		return form, finance.CreateTransactionInput{}, false
	}

	if category.Type == "expense" {
		amount = -math.Abs(amount)
	} else {
		amount = math.Abs(amount)
	}

	description := strings.TrimSpace(form.description)
	if len([]rune(description)) > maxFinanceDescriptionChars {
		form.errorText = fmt.Sprintf("Description must be %d characters or fewer", maxFinanceDescriptionChars)
		return form, finance.CreateTransactionInput{}, false
	}

	return form, finance.CreateTransactionInput{
		Date:        date,
		Amount:      amount,
		CategoryID:  sql.NullInt64{Int64: category.ID, Valid: true},
		Description: sql.NullString{String: description, Valid: description != ""},
	}, true
}

func (form *addFinanceTransactionForm) validateDate(date string) bool {
	parsed, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil || parsed.Format("2006-01-02") != date {
		form.errorText = "Date must use YYYY-MM-DD"
		return false
	}
	return true
}

func (form *addFinanceTransactionForm) amountValue() (float64, bool) {
	if strings.TrimSpace(form.amount) == "" {
		form.errorText = "Amount is required"
		return 0, false
	}

	amount, err := strconv.ParseFloat(form.amount, 64)
	if err != nil {
		form.errorText = "Amount must be a number"
		return 0, false
	}
	if amount <= 0 {
		form.errorText = "Amount must be greater than zero"
		return 0, false
	}
	return amount, true
}

func (form addFinanceTransactionForm) View(width int) string {
	return form.ViewWithTitle(width, "Add transaction")
}

func (form addFinanceTransactionForm) ViewWithTitle(width int, title string) string {
	modalWidth := modalWidth(width)
	fieldWidth := max(modalWidth-12, 12)

	lines := []string{
		formTitleStyle().Render(title),
		fieldLabel("Date", form.field == addFinanceTransactionDate),
		inputField(form.date, fieldWidth, form.field == addFinanceTransactionDate),
		fieldLabel("Amount", form.field == addFinanceTransactionAmount),
		inputField(form.amount, fieldWidth, form.field == addFinanceTransactionAmount),
		fieldLabel("Category", form.field == addFinanceTransactionCategory),
		categoryField(form.categories, form.categoryIndex, fieldWidth, form.field == addFinanceTransactionCategory),
		fieldLabel("Description", form.field == addFinanceTransactionDescription),
		textAreaField(form.description, fieldWidth, descriptionVisibleLines, form.field == addFinanceTransactionDescription),
	}
	if form.errorText != "" {
		lines = append(lines, errorStyle().Render(form.errorText))
	}
	return renderModal(modalWidth, lines)
}

func (form addFinanceTransactionForm) appendRunes(runes []rune) addFinanceTransactionForm {
	for _, r := range runes {
		form = form.appendRune(r)
	}
	return form
}

func (form addFinanceTransactionForm) appendRune(r rune) addFinanceTransactionForm {
	switch form.field {
	case addFinanceTransactionDate:
		if unicode.IsDigit(r) || r == '-' {
			form.date += string(r)
		} else {
			form.errorText = "Date accepts numbers and hyphens"
		}
	case addFinanceTransactionAmount:
		if unicode.IsDigit(r) || r == '.' {
			form.amount += string(r)
		} else {
			form.errorText = "Amount accepts numbers and decimals"
		}
	case addFinanceTransactionCategory:
		form = form.selectCategoryByRune(r)
	case addFinanceTransactionDescription:
		if len([]rune(form.description)) >= maxFinanceDescriptionChars {
			form.errorText = fmt.Sprintf("Description must be %d characters or fewer", maxFinanceDescriptionChars)
			return form
		}
		form.description += string(r)
	}
	return form
}

func (form addFinanceTransactionForm) deleteLastRune() addFinanceTransactionForm {
	switch form.field {
	case addFinanceTransactionDate:
		form.date = trimLastRune(form.date)
	case addFinanceTransactionAmount:
		form.amount = trimLastRune(form.amount)
	case addFinanceTransactionDescription:
		form.description = trimLastRune(form.description)
	}
	return form
}

func (form addFinanceTransactionForm) clearField() addFinanceTransactionForm {
	switch form.field {
	case addFinanceTransactionDate:
		form.date = ""
	case addFinanceTransactionAmount:
		form.amount = ""
	case addFinanceTransactionDescription:
		form.description = ""
	}
	return form
}

func (form addFinanceTransactionForm) moveCategory(distance int) addFinanceTransactionForm {
	if form.field != addFinanceTransactionCategory || len(form.categories) == 0 {
		return form
	}
	form.categoryIndex = (form.categoryIndex + distance + len(form.categories)) % len(form.categories)
	return form
}

func (form addFinanceTransactionForm) selectCategoryByRune(r rune) addFinanceTransactionForm {
	index, ok := categoryIndexFromRune(r)
	if !ok {
		form.errorText = "Category uses arrows, 1-9, or 0"
		return form
	}
	if index >= len(form.categories) {
		form.errorText = "No category at that number"
		return form
	}
	form.categoryIndex = index
	return form
}

func (form addFinanceTransactionForm) selectedCategory() (finance.Category, bool) {
	if len(form.categories) == 0 || form.categoryIndex < 0 || form.categoryIndex >= len(form.categories) {
		return finance.Category{}, false
	}
	return form.categories[form.categoryIndex], true
}

func (form addFinanceTransactionForm) categoryIndexFor(categoryID sql.NullInt64) int {
	if !categoryID.Valid {
		return 0
	}
	for i, category := range form.categories {
		if category.ID == categoryID.Int64 {
			return i
		}
	}
	return 0
}

func categoryField(categories []finance.Category, selected, width int, active bool) string {
	if len(categories) == 0 {
		return inputField("No categories", width, active)
	}

	contentWidth := max(width-2, 1)
	segments := make([]string, 0, len(categories))
	for i, category := range categories {
		label := fmt.Sprintf("%s %s", categoryShortcut(i), category.Name)
		if i == selected {
			label = "[" + label + "]"
		}
		segments = append(segments, fitFieldValue(label, contentWidth))
	}
	return borderedField(width, active).Render(strings.Join(wrapCategorySegments(segments, contentWidth), "\n"))
}

func categoryShortcut(index int) string {
	if index == 9 {
		return "0"
	}
	return strconv.Itoa(index + 1)
}

func categoryIndexFromRune(r rune) (int, bool) {
	if r == '0' {
		return 9, true
	}
	if r >= '1' && r <= '9' {
		return int(r - '1'), true
	}
	return 0, false
}

func wrapCategorySegments(segments []string, width int) []string {
	if len(segments) == 0 {
		return []string{""}
	}

	lines := []string{segments[0]}
	for _, segment := range segments[1:] {
		last := len(lines) - 1
		next := lines[last] + "  " + segment
		if lipgloss.Width(next) <= width {
			lines[last] = next
		} else {
			lines = append(lines, segment)
		}
	}
	return lines
}

func formatFormAmount(amount float64) string {
	amount = math.Abs(amount)
	if amount == math.Trunc(amount) {
		return strconv.FormatFloat(amount, 'f', 0, 64)
	}
	return strconv.FormatFloat(amount, 'f', 2, 64)
}
