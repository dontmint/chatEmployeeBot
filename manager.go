package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/spf13/viper"
)

func authManager(managerID int) bool {
	adminIds := viper.GetIntSlice("tgbotapi.adminId")
	if len(adminIds) == 0 {
		return false
	}
	for _, id := range adminIds {
		if id == managerID {
			return true
		}
	}
	return false
}

func reportManager(bot *tgbotapi.BotAPI, messageText string, chatID int64) error {
	args := strings.Split(messageText, " ")
	if len(args) == 3 {
		// Digit number for month and year
		month := args[1]
		year := args[2]
		monthlyReport := generatedMonthlyReport(0, month, year)
		err := generateExcel(monthlyReport)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "Error generating report: ")
			_, err := bot.Send(msg)
			if err != nil {
				return err
			}
		} else {
			log.Info("Sending report to admin")
			err := sendReportToUser(bot, fmt.Sprintf("%s-%s", month, year), chatID)
			if err != nil {
				return err
			}
		}
	} else {
		// Handle incorrect number of arguments
		msg := tgbotapi.NewMessage(chatID, "Invalid format. Use /report month year")
		_, err := bot.Send(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func createEmployee(db *sql.DB, workStatus WorkTrackingStatus, empl Employee) (int, error) {
	workStatusQuery := "INSERT INTO workTracking(employeeId, storeId, working) VALUES (?, ?, ?)"
	err := writeToDB(db, workStatusQuery,
		workStatus.EmployeeID,
		workStatus.StoreID,
		workStatus.Working)
	if err != nil {
		return empl.EmployeeID, err
	}
	emplQuery := "INSERT INTO employee (employeeId, name, storeId, salaryRate, active) VALUES (?, ?, ?, ?, ?)"
	err = writeToDB(db, emplQuery,
		empl.EmployeeID,
		empl.Name,
		empl.StoreID,
		empl.SalaryRate,
		empl.Active)
	if err != nil {
		return empl.EmployeeID, err
	}
	return empl.EmployeeID, nil
}

func createStore(db *sql.DB, store Store) error {
	storeQuery := "INSERT INTO store (name, address, latitude, longitude, active) VALUES (?, ?, ?, ?, ?)"
	err := writeToDB(db, storeQuery,
		store.Name,
		store.Address,
		store.Latitude,
		store.Longitude,
		store.Active,
	)
	if err != nil {
		return err
	}
	return nil
}

func sendReportToUser(bot *tgbotapi.BotAPI, caption string, chatID int64, optionalFileName ...string) error {
	var filename string
	if optionalFileName != nil {
		filename = optionalFileName[0]
	} else {
		filename = viper.GetString("reportPath")
	}
	// Open the Excel file
	file, err := os.Open(filename)
	if err != nil {
		log.Error(err)
	}
	defer file.Close()

	// Create a new message with document
	msg := tgbotapi.NewDocumentUpload(chatID, tgbotapi.FileReader{
		Name:   fmt.Sprintf("report-%s.xlsx", caption), // name of the file
		Reader: file,
		Size:   -1,
	})

	// Send the message
	if _, err := bot.Send(msg); err != nil {
		log.Fatal(err)
	}
	return nil
}
