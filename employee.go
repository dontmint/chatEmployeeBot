package main

import (
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func employeeCheckIn(bot *tgbotapi.BotAPI, chatID int64, messageID int, userLocation *tgbotapi.Location, employeeID int, date int) error {
	distance, storeName, storeID := (validateDistance(userLocation))
	var message string
	msg := tgbotapi.NewMessage(chatID, message)
	if int(math.Round(distance)) <= viper.GetInt("checkin.maxDistance") {
		msg.Text = "🟢 CheckIn successful 🟢\nStore: " + storeName
		checkInOut := tgbotapi.NewKeyboardButtonLocation("CheckOut")
		checkInOut.RequestLocation = true
		// Create a new reply keyboard with the buttons
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(checkInOut),
		)
		msg.ReplyMarkup = keyboard
		workStatus := WorkTrackingStatus{
			EmployeeID: employeeID,
			Working:    1,
			StoreID:    storeID,
		}
		db := connectToDB()
		defer db.Close()
		err := updateWorkTrackingStatus(db, workStatus)
		if err != nil {
			log.Error(err)
		}
		err = createWorkRecord(db, employeeID, "checkin", date)
		if err != nil {
			log.Error(err)
		}
		announceCheckInOut(bot, 1, employeeID, storeID)
	} else {
		msg.Text = "🔴 Checkin failed 🔴\nNot found any store"
	}

	msg.ReplyToMessageID = messageID
	_, err := bot.Send(msg)
	return err
}

func employeeCheckOut(bot *tgbotapi.BotAPI, chatID int64, messageID int, userLocation *tgbotapi.Location, employeeID int, date int) error {
	distance, storeName, storeID := (validateDistance(userLocation))
	var message string
	msg := tgbotapi.NewMessage(chatID, message)
	if int(math.Round(distance)) <= viper.GetInt("checkin.maxDistance") {
		checkInOut := tgbotapi.NewKeyboardButtonLocation("CheckIn")
		checkInOut.RequestLocation = true
		// Create a new reply keyboard with the buttons
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(checkInOut),
		)
		msg.ReplyMarkup = keyboard
		workStatus := WorkTrackingStatus{
			EmployeeID: employeeID,
			Working:    0,
			StoreID:    storeID,
		}
		db := connectToDB()
		defer db.Close()
		err := updateWorkTrackingStatus(db, workStatus)
		if err != nil {
			log.Debug(err)
		}
		err = createWorkRecord(db, employeeID, "checkout", date)
		if err != nil {
			log.Error(err)
		}
		announceCheckInOut(bot, 0, employeeID, storeID)
		totalTime := getTodayHours(employeeID)
		msg.Text = "🟢 CheckOut completed 🟢\nStore: " + storeName + "\nTotal time today: " + totalTime
	} else {
		msg.Text = "🔴 CheckOut failed 🔴\nNot found any store"
	}
	msg.ReplyToMessageID = messageID
	_, err := bot.Send(msg)
	return err
}

func employeeSalaryReport(bot *tgbotapi.BotAPI, chatID int64, employeeId int) {
	// Use time module to detect current system month and year
	now := time.Now()
	month := now.Month().String()
	year := strconv.Itoa(now.Year())
	log.Info("Report :" + strconv.Itoa(employeeId) + " -- " + month + " -- " + year)

	monthlyReport := generatedMonthlyReport(employeeId, month, year)
	fileName := "/tmp/report-" + strconv.Itoa(employeeId) + "-" + month + "-" + year + ".xlsx"
	err := generateExcel(monthlyReport, fileName)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Error generating report: ")
		_, err := bot.Send(msg)
		if err != nil {
			log.Error(err)
		}
	} else {
		log.Info("Sending report to User")
		err := sendReportToUser(bot, fmt.Sprintf("%s-%s", month, year), chatID, fileName)
		if err != nil {
			log.Error(err)
		}
	}
}

func getWorkTracking(db *sql.DB, employeeId int) (int, error) {
	// Prepare query statement
	query := "SELECT working FROM workTracking WHERE employeeId = " + strconv.Itoa(employeeId)
	row, err := readFromDB(db, query)
	if err != nil {
		log.Error(err)
	}
	var working int
	err = row.Scan(&working)
	if err != nil {
		log.Error(err)
	}
	// Return results
	return working, nil
}

func createWorkTrackingStatus(db *sql.DB, workStatus WorkTrackingStatus) error {
	query := "INSERT INTO workTracking(employeeId, storeId, working) VALUES ($1, $2, $3)"
	err := writeToDB(db, query,
		workStatus.EmployeeID,
		workStatus.StoreID,
		workStatus.Working)
	if err != nil {
		return err
	}
	return nil
}

func updateWorkTrackingStatus(db *sql.DB, workStatus WorkTrackingStatus) error {
	query := "UPDATE workTracking SET working = $1, storeId = $2 WHERE employeeId = $3 "
	err := writeToDB(db, query,
		workStatus.Working,
		workStatus.StoreID,
		workStatus.EmployeeID)
	if err != nil {
		return err
	}
	return nil
}

func createWorkRecord(db *sql.DB, employeeId int, recordType string, date int) error {
	var err error
	//Test new log time method
	time, day, month, year := extractDate(date)
	query := "INSERT INTO workRecord(employeeId, time, day, month, year, checkin, checkout) VALUES($1, $2, $3, $4, $5, $6, $7)"
	if recordType == "checkin" {
		err = writeToDB(db, query, employeeId, time, day, month, year, 1, 0)
	} else if recordType == "checkout" {
		err = writeToDB(db, query, employeeId, time, day, month, year, 0, 1)
	}
	if err != nil {
		return err
	}
	return nil
}

func getWorkRecords(db *sql.DB, employeeId int, year, month, day string) ([]WorkRecord, error) {
	// format the current time as "15:04 02 Jan 2006"
	// prepare the query statement
	var query string
	var rows *sql.Rows
	var err error
	if day == "" {
		query = `
		SELECT time, year, month, day, checkin, checkout
		FROM workRecord
		WHERE (employeeId = $1 AND year = $2 AND month = $3 )
		ORDER BY year, month, day, time;
		`
		rows, err = db.Query(query, employeeId, year, month)
	} else {
		query = `
		SELECT time, year, month, day, checkin, checkout
		FROM workRecord
		WHERE (employeeId = $1 AND year = $2 AND month = $3 AND day = $4)
		ORDER BY year, month, day, time
		`
		rows, err = db.Query(query, employeeId, year, month, day)
	}
	// execute the query with today's date as the parameter
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// iterate over the query results and build a slice of WorkRecord structs
	var workRecords []WorkRecord
	for rows.Next() {
		var wr WorkRecord
		wr.EmployeeID = employeeId
		err := rows.Scan(&wr.Time,
			&wr.Year,
			&wr.Month,
			&wr.Day,
			&wr.Checkin, &wr.Checkout)
		if err != nil {
			return nil, err
		}
		workRecords = append(workRecords, wr)
	}
	days := make([]time.Time, 0, len(workRecords))
	for _, p := range workRecords {
		days = append(days, parseTime(p.Day))
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return workRecords, nil
}

func calculateDurations(records []WorkRecord) []time.Duration {
	var lastCheckInTime time.Time
	location, err := time.LoadLocation(viper.GetString("timezone"))
	if err != nil {
		log.Error("Cannot load timzezone: ", err)
	}
	durations := make([]time.Duration, 0)
	employeeTimes := make(map[int][]time.Time)

	for _, r := range records {
		// Parse check-in/check-out time
		t, err := time.ParseInLocation("15:04 2 January 2006", fmt.Sprintf("%s %s %s %s", r.Time, r.Day, r.Month, r.Year), location)
		if err != nil {
			log.Error("Error parsing time: ", err)
			continue
		}

		// Store check-in/check-out time for each employee
		if r.Checkin {
			employeeTimes[r.EmployeeID] = []time.Time{t}
		} else if r.Checkout {
			checkInTime, exists := employeeTimes[r.EmployeeID]
			if !exists {
				log.Info("User did not checkin but checkout: ", r.EmployeeID)
				continue
			}
			lastCheckInTime = checkInTime[0]
			delete(employeeTimes, r.EmployeeID)

			// Calculate duration and append to durations
			duration := t.Sub(lastCheckInTime)
			durations = append(durations, duration)
		} else {
			log.Warn("Invalid work record type: ", r)
		}
	}

	return durations
}

func getTodayHours(employeeId int) string {
	db := connectToDB()
	defer db.Close()
	_, day, month, year := getCurrentTime()
	workRecord, _ := getWorkRecords(db, employeeId, year, month, day)
	durations := calculateDurations(workRecord)
	return sumDurations(durations)
}
func sumDurations(durations []time.Duration) string {
	total := time.Duration(0)
	for _, d := range durations {
		total += d
	}
	hours := int(total.Hours())
	minutes := int(total.Minutes()) % 60
	seconds := int(total.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func ResetWorkTracking() bool {
	db := connectToDB()
	defer db.Close()
	query := "UPDATE workTracking SET working = 0;"
	err := writeToDB(db, query)
	if err != nil {
		log.Error("Cannot reset work tracking: ", err)
		return false
	}
	log.Info("Reset work tracking successful")
	return true
}
