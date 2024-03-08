package filter

import (
	"main/pkg/conv/updtypes"
	"strings"
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
			return true
		}
	}
	return false
}

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
			return true
		}
	}
	return false
}
