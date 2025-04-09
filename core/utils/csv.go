package utils

import (
	"database/sql/driver"
	"errors"
	"strings"

	"golang.org/x/exp/slices"
)

// CSV represents a slice of strings for CSV data.
type CSV []string

// Scan implements the sql.Scanner interface, allowing CSV to be scanned from a database.
func (csv *CSV) Scan(src any) error {
	if src == nil {
		*csv = nil
		return nil
	}

	switch v := src.(type) {
	case CSV:
		*csv = v
		return nil
	case string:
		*csv = strings.Split(v, ",")
		return nil
	default:
		return errors.New("invalid type")
	}
}

// Value implements the driver.Valuer interface, allowing CSV to be converted to a database value.
func (csv CSV) Value() (driver.Value, error) {
	if csv == nil {
		return "", nil
	}
	return strings.Join(csv, ","), nil
}

// Includes checks if the given string is present in the CSV slice.
func (csv *CSV) Includes(s string) bool {
	return slices.Contains(*csv, s)
}
