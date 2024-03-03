package question

type Question struct {
	Prompt    *QData
	BadPrompt *QData
	Filters   uint
}

type QData struct {
	Text   string
	Markup interface{}
}
