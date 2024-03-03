package conv

import (
	"fmt"
	"main/pkg/conv/question"
	"main/pkg/conv/subscriber"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type ConversationsManager struct {
	CommandObserver *Command
	Clients         map[int64]*subscriber.Client
	Bot             *tgbotapi.BotAPI
}

func New(bot *tgbotapi.BotAPI) *ConversationsManager {
	return &ConversationsManager{
		CommandObserver: &Command{
			Subscribers: make(map[string]*subscriber.Subscriber),
		},
		Clients: make(map[int64]*subscriber.Client),
		Bot:     bot,
	}
}

type Command struct {
	Subscribers map[string]*subscriber.Subscriber
}

func (c *Command) Subscribe(command string, subscriber *subscriber.Subscriber) {
	c.Subscribers[command] = subscriber
}

const (
	_TYPE_MESSAGE  = 0
	_TYPE_CALLBACK = 1
)

func (cm *ConversationsManager) HandleUpdate(update *tgbotapi.Update) {
	// if received command or data of callback
	// check do we handle this command
	// -> use notify to start handling command

	var updateType uint

	switch {
	case update.Message != nil && update.Message.IsCommand():
		updateType = _TYPE_MESSAGE
		cmd := fmt.Sprintf("/%s", update.Message.Command())
		if _, exist := cm.CommandObserver.Subscribers[cmd]; exist {
			cm.HandleCommand(cmd, update.Message.Chat.ID, update)
			return
		}
	case update.CallbackQuery != nil:
		updateType = _TYPE_CALLBACK
		cm.Bot.AnswerCallbackQuery(tgbotapi.CallbackConfig{CallbackQueryID: update.CallbackQuery.ID, ShowAlert: false})
		if _, exist := cm.CommandObserver.Subscribers[update.CallbackQuery.Data]; exist {
			cm.HandleCommand(update.CallbackQuery.Data, update.CallbackQuery.Message.Chat.ID, update)
			return
		}
	}

	// else
	// handle message as response to client active command
	cm.handleResponse(update, updateType)
}

func (cm *ConversationsManager) HandleCommand(cmd string, chatId int64, update *tgbotapi.Update) {
	if handler, exist := cm.CommandObserver.Subscribers[cmd]; exist {

		var activeClient *subscriber.Client
		if activeClient, exist = cm.Clients[chatId]; exist {
			activeClient.CurrentHandler = handler
			clear(activeClient.Responses)
		} else {
			activeClient = &subscriber.Client{ChatId: chatId}
			cm.Clients[chatId] = activeClient
		}

		handler.HandleCommand(activeClient, cm.Bot)
		// get response from gorutine and set next step or send bad feedback
		return
	}
}

func (cm *ConversationsManager) handleResponse(update *tgbotapi.Update, updateType uint) {
	// handle responses to questions
	// ...

	var chatId int64
	var responseText string

	if updateType == _TYPE_MESSAGE {
		chatId = update.Message.Chat.ID
		responseText = update.Message.Text
	} else {
		chatId = update.CallbackQuery.Message.Chat.ID
		responseText = update.CallbackQuery.Data
	}

	if _, exist := cm.Clients[chatId]; !exist {
		fmt.Println("Received non-command response from client as first message?")
		return
	}
	client := cm.Clients[chatId]

	responseIdx := len(client.Responses)
	currentQuestion := client.CurrentHandler.Questions[responseIdx]

	// check is response valid

	// invalid response
	// TODO: Перепроверить и переделать фильтры на бит маски, чтобы CMD и TEXT выключали друг друга при активации
	if (((currentQuestion.Filters & question.FILTER_CALLBACK) == 0) && updateType == _TYPE_CALLBACK) ||
		((responseText[0] != '/') && (currentQuestion.Filters&question.FILTER_CMD) != 0) ||
		((responseText[0] == '/') && (currentQuestion.Filters&question.FILTER_TEXT) != 0) {
		badPrompt := tgbotapi.NewMessage(client.ChatId, currentQuestion.BadPrompt.Text)
		badPrompt.ReplyMarkup = currentQuestion.BadPrompt.Markup
		cm.Bot.Send(badPrompt)
	}
	client.Responses = append(client.Responses, responseText)

	// check is it last question
	if len(client.CurrentHandler.Questions) == responseIdx+1 {

		mmm := tgbotapi.NewMessage(client.ChatId, fmt.Sprintf("Отлично! Ваши данные получены!\n%v", client.Responses))
		cm.Bot.Send(mmm)
		clear(client.Responses)
		client.Responses = []string{}
		// provide to next handler
		if client.CurrentHandler.ProvideTo != nil {
			// TODO: save responses
			client.CurrentHandler = client.CurrentHandler.ProvideTo
			client.CurrentHandler.HandleCommand(client, cm.Bot)
			// TODO: send finish message
		}
		return
	}

	nextQuestion := client.CurrentHandler.Questions[responseIdx+1]

	nextMessage := tgbotapi.NewMessage(client.ChatId, nextQuestion.Prompt.Text)
	nextMessage.ReplyMarkup = nextQuestion.Prompt.Markup
	cm.Bot.Send(nextMessage)
	// - get handler
	// - is received answer good?
	// - save response
	// get next question | SWITCH on Provided handler and send handle his command
	// update data about current step
	// send next question

}
