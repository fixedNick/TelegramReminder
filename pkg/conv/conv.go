package conv

import (
	"fmt"
	"main/pkg/conv/filter"
	"main/pkg/conv/subscriber"
	"main/pkg/conv/updtypes"

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

func (cm *ConversationsManager) HandleUpdate(update *tgbotapi.Update) {
	// if received command or data of callback
	// check do we handle this command
	// -> use notify to start handling command

	var (
		updateType   uint
		chatId       int64
		responseText string
		client       *subscriber.Client
	)

	switch {
	case update.Message != nil && update.Message.IsCommand():
		updateType = updtypes.TYPE_MESSAGE

		chatId = update.Message.Chat.ID
		responseText = update.Message.Text

		cmd := fmt.Sprintf("/%s", update.Message.Command())
		if _, exist := cm.CommandObserver.Subscribers[cmd]; exist {
			cm.HandleCommand(cmd, chatId, update)
			return
		}
	case update.CallbackQuery != nil:

		updateType = updtypes.TYPE_CALLBACK
		chatId = update.CallbackQuery.Message.Chat.ID
		responseText = update.CallbackQuery.Data

		cm.Bot.AnswerCallbackQuery(tgbotapi.CallbackConfig{CallbackQueryID: update.CallbackQuery.ID, ShowAlert: false})
		if _, exist := cm.CommandObserver.Subscribers[update.CallbackQuery.Data]; exist {
			cm.HandleCommand(update.CallbackQuery.Data, chatId, update)
			return
		}
	case update.Message != nil:
		chatId = update.Message.Chat.ID
		responseText = update.Message.Text
	}

	if _, exist := cm.Clients[chatId]; !exist {
		fmt.Println("Received non-command response from client as first message?")
		return
	}

	client = cm.Clients[chatId]

	cm.handleResponse(client, responseText, updateType)
}

func (cm *ConversationsManager) HandleCommand(cmd string, chatId int64, update *tgbotapi.Update) {
	if handler, exist := cm.CommandObserver.Subscribers[cmd]; exist {

		var activeClient *subscriber.Client
		if activeClient, exist = cm.Clients[chatId]; exist {
			activeClient.CurrentHandler = handler
			clear(activeClient.Responses)
		} else {
			activeClient = &subscriber.Client{ChatId: chatId, CurrentHandler: handler}
			cm.Clients[chatId] = activeClient
		}

		handler.HandleCommand(activeClient, cm.Bot)
		// get response from gorutine and set next step or send bad feedback
		return
	}
}

func (cm *ConversationsManager) handleResponse(client *subscriber.Client, responseText string, updateType uint) {
	// handle responses to questions

	responseIdx := len(client.Responses)
	currentQuestion := client.CurrentHandler.Questions[responseIdx]

	// check is response valid
	if !filter.IsValid(currentQuestion.Filters, updateType, responseText) {
		if currentQuestion.BadPrompt != nil {
			badPrompt := tgbotapi.NewMessage(client.ChatId, currentQuestion.BadPrompt.Text)
			badPrompt.ReplyMarkup = currentQuestion.BadPrompt.Markup
			cm.Bot.Send(badPrompt)
		}
		return
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
}
