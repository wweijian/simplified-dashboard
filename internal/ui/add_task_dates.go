package ui

import (
	"database/sql"
	"strconv"
	"time"
)

func dueDaysFromDate(value sql.NullString, now time.Time) string {
	if !value.Valid || value.String == "" {
		return ""
	}

	dueDate, err := time.ParseInLocation("2006-01-02", value.String, time.Local)
	if err != nil {
		return ""
	}

	today := now.In(time.Local).Truncate(24 * time.Hour)
	return strconv.Itoa(int(dueDate.Sub(today).Hours() / 24))
}
