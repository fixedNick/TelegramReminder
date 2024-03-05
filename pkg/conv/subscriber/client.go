package subscriber

type Client struct {
	ChatId         int64
	Responses      []string
	CurrentHandler *Subscriber
	HandlerCommand string
}
