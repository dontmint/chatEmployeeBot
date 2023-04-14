package main

import (
	log "github.com/sirupsen/logrus"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func startMenu(bot *tgbotapi.BotAPI, chatID int64, employeeId int, name string) error {
	//Define the buttons for the custom keyboard
	help := tgbotapi.NewKeyboardButtonLocation("/help")
	// Create a new reply keyboard with the buttons
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(help),
	)
	// Create a new message with the custom keyboard and send it to the user
	msg := tgbotapi.NewMessage(chatID, "Welcome to your amazing workspace\n Use /help for the instruction")
	msg.ReplyMarkup = keyboard
	_, err := bot.Send(msg)
	newUserAlert(bot, employeeId, name)
	return err
}

func helpMenu(bot *tgbotapi.BotAPI, chatID int64) error {
	//Define the buttons for the custom keyboard
	message := `Hello and welcome to our company's chat bot!,
	/checkin
	/checkout - Checkin or checkout at your store
	/salary - Show your salary and work record this month`
	msg := tgbotapi.NewMessage(chatID, message)
	_, err := bot.Send(msg)
	return err
}

func checkIn(bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
	checkInOut := tgbotapi.NewKeyboardButtonLocation("")
	db := connectToDB()
	defer db.Close()
	status, err := getWorkTracking(db, update.Message.From.ID)
	if err != nil {
		log.Error(err)
	}
	if status == 0 {
		checkInOut.Text = "CheckIn"
	} else if status == 1 {
		checkInOut.Text = "CheckOut"
	}
	//Define the buttons for the custom keyboard
	message := `Share your location to continue`
	checkInOut.RequestLocation = true
	// Create a new reply keyboard with the buttons
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(checkInOut),
	)
	// Create a new message with the custom keyboard and send it to the user
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
	msg.ReplyMarkup = keyboard
	_, err = bot.Send(msg)
	return err
}
