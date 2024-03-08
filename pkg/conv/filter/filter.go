package filter

import (
	"main/pkg/conv/updtypes"
	"strconv"
	"strings"
	"time"
)

// TODO: USE THIS IN INVALID CHECK IN conv.go
const (
	FILTER_TEXT = 1 << iota
	FILTER_CALLBACK
	FILTER_CMD
	FILTER_NUMBER //
	FILTER_DATE
	FILTER_TIME
)

// TODO: Перепроверить и переделать фильтры на бит маски, чтобы CMD и TEXT выключали друг друга при активации
// IsValid - Func validates is response of type updateType passes all filters
// Returns TRUE if valid is OK, and false if not
func IsValid(filters uint, updateType uint, response string) bool {
	if (((filters & FILTER_CALLBACK) == 0) && updateType == updtypes.TYPE_CALLBACK) ||
		((response[0] != '/') && (filters&FILTER_CMD) != 0) ||
		((response[0] == '/') && (filters&FILTER_TEXT) != 0) ||
		((filters&FILTER_TEXT) == 0 && updateType == updtypes.TYPE_MESSAGE) ||
		((filters&FILTER_TIME) != 0 && !isTime(response)) ||
		((filters&FILTER_DATE) != 0 && !isDate(response)) {
		return false
	}
	return true
}

func isTime(text string) bool {
	delimiters := []rune{':', ' '}
	for _, delimiter := range delimiters {
		if splitted := strings.Split(text, string(delimiter)); len(splitted) == 2 {
			hours, hErr := strconv.Atoi(splitted[0])
			minutes, mErr := strconv.Atoi(splitted[1])
			return (hours >= 0 && hours <= 24) && hErr == nil && mErr == nil && (minutes >= 0 && minutes <= 60)
		}
	}
	return false
}

// TODO:
// Add exceptions and handling them to understand why is date or time is invalid
func isDate(text string) bool {
	delimiters := []rune{':', '/', ' ', '\\'}

	isDateValid := func(dateParts []string) bool {
		isYearFound := false
		mdCounter := 0
		for i := 0; i < 3; i++ {
			if len(dateParts[i]) == 2 {
				mdCounter++
			} else if len(dateParts[i]) == 3 {
				isYearFound = true
			}
		}
		return isYearFound && mdCounter == 2
	}

	for _, delimiter := range delimiters {
		if splitted := strings.Split(text, string(delimiter)); len(splitted) == 3 && isDateValid(splitted) {

			// is year valid
			year, err := strconv.Atoi(splitted[2])
			if err != nil {
				return false
			}

			// is month valid
			month, err := strconv.Atoi(splitted[1])
			if err != nil || month < 0 || month > 12 {
				return false
			}

			// is days valid
			day, err := strconv.Atoi(splitted[0])
			if err != nil {
				return false
			}

			// is month have these days
			t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
			return t.Day() == day
		}
	}
	return false
}
