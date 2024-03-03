package filter

import "main/pkg/conv/updtypes"

// TODO: USE THIS IN INVALID CHECK IN conv.go
const (
	FILTER_TEXT = 1 << iota
	FILTER_CALLBACK
	FILTER_CMD
	FILTER_NUMBER //
	FILTER_DATE   //
	FILTER_TIME   //
)

// TODO: Перепроверить и переделать фильтры на бит маски, чтобы CMD и TEXT выключали друг друга при активации
// IsValid - Func validates is response of type updateType passes all filters
// Returns TRUE if valid is OK, and false if not
func IsValid(filters uint, updateType uint, response string) bool {
	if (((filters & FILTER_CALLBACK) == 0) && updateType == updtypes.TYPE_CALLBACK) ||
		((response[0] != '/') && (filters&FILTER_CMD) != 0) ||
		((response[0] == '/') && (filters&FILTER_TEXT) != 0) ||
		((filters&FILTER_TEXT) == 0 && updateType == updtypes.TYPE_MESSAGE) {
		return false
	}
	return true
}
