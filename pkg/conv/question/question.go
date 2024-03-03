package question

// TODO: USE THIS IN INVALID CHECK IN conv.go
const (
	FILTER_TEXT = 1 << iota
	FILTER_CALLBACK
	FILTER_CMD
	FILTER_NUMBER //
	FILTER_DATE   //
	FILTER_TIME   //
)

type Question struct {
	Prompt    *QData
	BadPrompt *QData
	Filters   uint
}

type QData struct {
	Text   string
	Markup interface{}
}
