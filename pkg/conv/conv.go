package conv

import (
	"fmt"
	"log"
	"main/pkg/conv/filter"
	"main/pkg/conv/sql"
	"main/pkg/conv/subscriber"
	"main/pkg/conv/updtypes"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type ConversationsManager struct {
	CommandObserver *Command
	Clients         map[int64]*subscriber.Client
	Bot             *tgbotapi.BotAPI
	DB              *sql.Sql
}

func New(bot *tgbotapi.BotAPI, db *sql.Sql) *ConversationsManager {
	sql.Init(db)

	return &ConversationsManager{
		CommandObserver: &Command{
			Subscribers: make(map[string]*subscriber.Subscriber),
		},
		Clients: *db.GetClientsFromDB(),
		Bot:     bot,
		DB:      db,
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

	// checks is client in localdb

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

			// New client
			// Add client to sql
			// Add client ot local storage (cm.Clients)

			var (
				fullName string
				domain   string
			)

			if update.Message != nil {
				fullName = strings.TrimSpace(fmt.Sprintf("%s %s", update.Message.From.FirstName, update.Message.From.LastName))
				domain = update.Message.From.UserName
			} else {
				fullName = strings.TrimSpace(fmt.Sprintf("%s %s", update.CallbackQuery.From.FirstName, update.CallbackQuery.From.LastName))
				domain = update.CallbackQuery.From.UserName
			}
			cm.DB.AddClient(chatId, fullName, &domain, &cmd)

			activeClient = &subscriber.Client{ChatId: chatId, CurrentHandler: handler}
			cm.Clients[chatId] = activeClient
		}

		// TODO: Refactor Handlers to one object
		// TODO: Add client funcs to manipulate states
		activeClient.HandlerCommand = cmd
		cm.DB.UpdateClientHandler(chatId, cmd)
		cm.DB.DeleteClientResponses(chatId)

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

	// Response is valid, update local and db storages
	client.Responses = append(client.Responses, responseText)
	cm.DB.AddClientResponse(client.ChatId, responseText)

	// check is it last question
	if len(client.CurrentHandler.Questions) == responseIdx+1 {

		mmm := tgbotapi.NewMessage(client.ChatId, fmt.Sprintf("Отлично! Ваши данные получены!\n%v", client.Responses))
		cm.Bot.Send(mmm)
		// provide to next handler
		if client.CurrentHandler.ProvideTo != nil {
			client.CurrentHandler.HandleCommand(client, cm.Bot)
			// TODO: Proceed operation, ex: Add Important Date
			if client.CurrentHandler.FinishFunc != nil {
				client.CurrentHandler.FinishFunc(client, cm.Bot)
			}
			client.CurrentHandler = client.CurrentHandler.ProvideTo
			cm.DB.DeleteClientResponses(client.ChatId)
			// TODO: send finish message
		}
		clear(client.Responses)
		client.Responses = []string{}
		return
	}

	nextQuestion := client.CurrentHandler.Questions[responseIdx+1]

	nextMessage := tgbotapi.NewMessage(client.ChatId, nextQuestion.Prompt.Text)
	nextMessage.ReplyMarkup = nextQuestion.Prompt.Markup
	cm.Bot.Send(nextMessage)
}

func (cm *ConversationsManager) AssociateClienthWithHandlers() {
	for _, client := range cm.Clients {
		if client.HandlerCommand == "" {
			continue
		}

		if associatedHandler, exist := cm.CommandObserver.Subscribers[client.HandlerCommand]; exist {
			client.CurrentHandler = associatedHandler
			continue
		}

		log.Printf("Cannot associate current handler for client %d. Command handler is %s.", client.ChatId, client.HandlerCommand)
	}
}
