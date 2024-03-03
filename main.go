package main

import (
	"bufio"
	"fmt"
	"log"
	"main/pkg/conv"
	"main/pkg/conv/question"
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
		if strings.Contains(scanner.Text(), "token") == false {
			continue
		}
		token = strings.Split(scanner.Text(), "=")[1]
		break
	}

	if scanner.Err() != nil {
		panic("Errors while read file")
	}

	bot, _ := tgbotapi.NewBotAPI(token)

	c := conv.New(bot)

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
				Filters: question.FILTER_CALLBACK,
			},
		},
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
				Filters: question.FILTER_TEXT | question.FILTER_CALLBACK,
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
				Filters: question.FILTER_TEXT,
			},
		},
		startHandler,
	)
	c.CommandObserver.Subscribe("/addevent", addEventHandler)
	c.CommandObserver.Subscribe("/start", startHandler)

	// bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 3
	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		c.HandleUpdate(&update)
	}
}

func HandleCallbackQuery(bot *tgbotapi.BotAPI, cq *tgbotapi.CallbackQuery) {
	bot.AnswerCallbackQuery(tgbotapi.CallbackConfig{CallbackQueryID: cq.ID, ShowAlert: false})
}

func HandleCommand(bot *tgbotapi.BotAPI, chatId int64, cmd string) {
	switch cmd {
	case "start", "menu":
		// start conv
		startKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Список всех ваших событий", "callback_all_dates")),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Добавить", "callback_add_date"),
				tgbotapi.NewInlineKeyboardButtonData("Ближайшие события", "callback_nearest_dates"),
			),
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Настройки", "callback_settings")),
		)
		msg := tgbotapi.NewMessage(chatId, "Главное меню")
		msg.ReplyMarkup = startKeyboard
		bot.Send(msg)
	case "help":
		bot.Send(tgbotapi.NewMessage(chatId,
			` Основная команда для работы - это меню ( /menu ).
В меню есть несколько кнопок:
Список ваших дат - Показывает все даты, которые вы добавили для напоминания
Добавить - Добавляет дату для напоминания
Ближайшие даты - показывает 4 (по-умолчанию) ближайших события	
`))
	default:
		bot.Send(
			tgbotapi.NewMessage(
				chatId,
				fmt.Sprintf("К сожалению, я не умею использовать команду /%s\nИспользуйте /help для помощи или воспользуйтесь командами из быстрого доступа слева снизу.", cmd),
			),
		)
	}
}
