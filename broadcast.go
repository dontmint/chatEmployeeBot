package main

import (
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func sendToGroup(bot *tgbotapi.BotAPI, message string) error {
	if viper.IsSet("tgbotapi.groupBroadCast.id") {
		msg := tgbotapi.NewMessage(viper.GetInt64("tgbotapi.groupBroadCast.id"), message)
		_, err := bot.Send(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func announceCheckInOut(bot *tgbotapi.BotAPI, status int, employeeId int, storeID int) {
	db := connectToDB()
	defer db.Close()
	query := "SELECT name FROM employee WHERE employeeid = " + strconv.Itoa(employeeId)
	row, err := readFromDB(db, query)
	if err != nil {
		log.Error(err)
	}
	var employeeName string
	err = row.Scan(&employeeName)
	if err != nil {
		log.Error(err)
	}

	query = "SELECT name FROM store WHERE storeid = " + strconv.Itoa(storeID)
	row, err = readFromDB(db, query)
	if err != nil {
		log.Error(err)
	}
	var storeName string
	err = row.Scan(&storeName)
	if err != nil {
		log.Error(err)
	}
	var message string
	if status == 1 {
		message = "🟢 " + employeeName + " 🟢 " + " Check In | " + storeName
	} else if status == 0 {
		message = "🟡 " + employeeName + " 🟡 " + " Check Out | " + storeName
	}
	sendToGroup(bot, message)
}

func newUserAlert(bot *tgbotapi.BotAPI, employeeId int, name string) {
	message := "New user has follow the bot\n"
	message = message + "ID \t- \t" + strconv.Itoa(employeeId) + "\n"
	message = message + "Name \t- \t" + name + "\n"
	sendToGroup(bot, message)
}
