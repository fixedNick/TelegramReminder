package main

import (
	"bufio"
	"log"
	"main/pkg/conv"
	"main/pkg/conv/filter"
	"main/pkg/conv/question"
	"main/pkg/conv/sql"
	"main/pkg/conv/subscriber"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {

	// TODO: use YAML file to manipulate config
	// TODO: use config as flag or panic

	// get token from txt file
	config, err := os.Open("config.txt")
	if err != nil {
		log.Fatal("Add file config.txt into your main directory")
	}
	defer config.Close()

	scanner := bufio.NewScanner(config)

	var token string
	for scanner.Scan() {
		if !strings.Contains(scanner.Text(), "token") {
			continue
		}
		token = strings.Split(scanner.Text(), "=")[1]
		break
	}

	if scanner.Err() != nil {
		panic("Errors while read file")
	}

	bot, _ := tgbotapi.NewBotAPI(token)

	c := conv.New(
		bot,
		&sql.Sql{
			Host:     "",
			User:     "root",
			Password: "",
			DBName:   "reminder",
		},
	)

	// create handlers
	startHandler := subscriber.NewHandler(
		[]question.Question{
			{
				Prompt: &question.QData{
					Text: "Главное меню",
					Markup: &tgbotapi.InlineKeyboardMarkup{
						InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
							tgbotapi.NewInlineKeyboardRow(
								tgbotapi.NewInlineKeyboardButtonData("Добавить дату", "/addevent"),
							),
						},
					},
				},
				BadPrompt: &question.QData{
					Text: "К сожалению я не могу распознать ваше сообщение. Для дальнейшей работы, пожалуйста, выбери пункт из меню.",
				},
				Filters: filter.FILTER_CALLBACK,
			},
		},
		nil,
		nil,
	)

	addEventHandler := subscriber.NewHandler(
		[]question.Question{
			{
				Prompt: &question.QData{
					Text: "Как у тебя дела? Выбери одно или просто напиши в чат.",
					Markup: &tgbotapi.InlineKeyboardMarkup{
						InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
							tgbotapi.NewInlineKeyboardRow(
								tgbotapi.NewInlineKeyboardButtonData("Хорошо", "callback_good"),
								tgbotapi.NewInlineKeyboardButtonData("Неоч", "callback_bad"),
							),
						},
					},
				},
				Filters: filter.FILTER_CALLBACK,
				BadPrompt: &question.QData{
					Text: "Не могу понять как твои дела. Пожалуйста, нажми на одну из кнопок выше",
				},
			},
			{
				Prompt: &question.QData{
					Text: "С делами разобрались, чо по чем (text only)",
					Markup: &tgbotapi.ReplyKeyboardMarkup{
						Keyboard: [][]tgbotapi.KeyboardButton{
							tgbotapi.NewKeyboardButtonRow(
								tgbotapi.NewKeyboardButton("Ну кирпичОм"),
								tgbotapi.NewKeyboardButton("Фиг с эти вРачОм"),
							),
						},
					},
				},
				Filters: filter.FILTER_TEXT,
			},
		},
		startHandler,
		func(client *subscriber.Client, bot *tgbotapi.BotAPI) {
			log.Printf("Done for client %d, with responses: %v", client.ChatId, client.Responses)
		},
	)
	c.CommandObserver.Subscribe("/addevent", addEventHandler)
	c.CommandObserver.Subscribe("/start", startHandler)

	// bot.Debug = true

	c.AssociateClienthWithHandlers()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 3
	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		c.HandleUpdate(&update)
	}
}
