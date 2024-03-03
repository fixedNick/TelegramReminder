package subscriber

import (
	"fmt"
	"main/pkg/conv/question"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type CommandSubscriber interface {
	HandleCommand(activeClient *Client, bot *tgbotapi.BotAPI)
}

type Subscriber struct {
	Questions []question.Question
	ProvideTo *Subscriber
	// TODO:
	// Добавить поле Message, Которое будет отослано после завершения диалога с данным подсписчиком
}

func NewHandler(questions []question.Question, provideTo *Subscriber) *Subscriber {
	return &Subscriber{
		Questions: questions,
		ProvideTo: provideTo,
	}
}

func (s *Subscriber) HandleCommand(client *Client, bot *tgbotapi.BotAPI) {

	// send first question if it exist [should be exist]
	if len(s.Questions) == 0 {
		panic(fmt.Sprintf("There are no questions for command %v", s))
	}
	firstQuestion := s.Questions[0]
	msg := tgbotapi.NewMessage(client.ChatId, firstQuestion.Prompt.Text)
	msg.ReplyMarkup = firstQuestion.Prompt.Markup

	bot.Send(msg)
}
