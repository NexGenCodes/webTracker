package dbutil

import (
	"database/sql"
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

// ToNullFloat64 converts a float64 to sql.NullFloat64.
func ToNullFloat64(f float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: f, Valid: true}
}
