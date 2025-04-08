package utils

import (
	"database/sql/driver"
	"errors"
	"strings"
)

type CSV []string

func (csv *CSV) Scan(src any) error {
	if src == nil {
		*csv = nil
		return nil
	}

	switch src.(type) {
	case CSV:
		*csv = src.(CSV)
		return nil
	case string:
		*csv = strings.Split(src.(string), ",")
		return nil
	default:
		return errors.New("invalid type")
	}
}

func (csv CSV) Value() (driver.Value, error) {
	if csv == nil {
		return "", nil
	}
	return strings.Join(csv, ","), nil
}

func (csv *CSV) Includes(s string) bool {
	for _, v := range *csv {
		if v == s {
			return true
		}
	}
	return false
}
