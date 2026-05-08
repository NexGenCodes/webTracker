package dbutil

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

// ToNullString converts a string to sql.NullString.
func ToNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

// ToNullTime converts a time.Time to sql.NullTime.
func ToNullTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: !t.IsZero()}
}

// FloatToNullNumeric converts a float64 to sql.NullString for NUMERIC columns.
func FloatToNullNumeric(f float64) sql.NullString {
	return sql.NullString{String: fmt.Sprintf("%.2f", f), Valid: true}
}

// NullNumericToFloat converts a sql.NullString from a NUMERIC column back to float64.
func NullNumericToFloat(ns sql.NullString) float64 {
	if !ns.Valid || ns.String == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(ns.String, 64)
	return f
}
