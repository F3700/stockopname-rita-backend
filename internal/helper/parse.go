package helper

import (
	"fmt"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

func ParseDate(dateStr string) (*time.Time, error) {
	if dateStr == "" {
		return nil, nil
	}
	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}
	return &date, nil
}

func ParseIntParam(params httprouter.Params, key string) (int, error) {
	valueStr := params.ByName(key)
	if valueStr == "" {
		return 0, fmt.Errorf("missing parameter: %s", key)
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("invalid parameter format: %s", key)
	}
	return value, nil
}

func ParseDateNullable(endDate *time.Time) string {
	if endDate == nil {
		return ""
	}
	return endDate.Format(time.RFC3339)
}
