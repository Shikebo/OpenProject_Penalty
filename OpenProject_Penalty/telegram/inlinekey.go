package telegram

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// var Menu = tgbotapi.NewInlineKeyboardMarkup(
//
//	tgbotapi.NewInlineKeyboardRow(
//		tgbotapi.NewInlineKeyboardButtonData("📰 Мои задачи", "callback_mytasks"),
//	),
var Menu = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("📋 Просроченные задачи", "callback_penalties"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("💵 Виды штрафов", "callback_mypenalty"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("📰 Новости", "callback_notifications"),
	),
)

var Authorization = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("/Commands"),
	))
