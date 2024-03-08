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
								tgbotapi.NewInlineKeyboardButtonData("Список всех ваших событий", "/incoming"),
							),
							tgbotapi.NewInlineKeyboardRow(
								tgbotapi.NewInlineKeyboardButtonData("Добавить новое событие", "/addevent"),
								tgbotapi.NewInlineKeyboardButtonData("Ближайшие события", "/nearest"),
							),
							tgbotapi.NewInlineKeyboardRow(
								tgbotapi.NewInlineKeyboardButtonData("Настройки", "/settings"),
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
					Text: "Как назовем ваше событие? Например: 'День рождение начальника', 'Запись к терапевту'",
				},
				Filters: filter.FILTER_TEXT,
				BadPrompt: &question.QData{
					Text: "Пожалуйста, введите корректное название текстом",
				},
			},
			{
				Prompt: &question.QData{
					Text: "Добавьте описание событию, чтобы не забыть детали, когда наступит момент X",
				},
				Filters: filter.FILTER_TEXT,
			},
			{
				Prompt: &question.QData{
					Text: "На какую дату планируется данное событие? Можете использовать любой из представленных шаблонов для даты:\n25.01.2024\n25/01/2024\n25\\01\\2024\n25:01:2024\n25 01 2024",
				},
			},
		},
		startHandler,
		func(client *subscriber.Client, bot *tgbotapi.BotAPI) {
			bot.Send(tgbotapi.NewMessage(client.ChatId, "Date successfully added"))
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
