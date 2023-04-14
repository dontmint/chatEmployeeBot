package main

import (
	"strings"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	// Init config
	readConfig()

	// Cronjob
	ResetWorkTrackingJob()

	// Start API server
	go func() {
		if err := startGinServer(viper.GetString("api.listenAddr")); err != nil {
			log.Fatalf("Failed to start Gin server: %v", err)
		}
	}()

	// Create telegram bot api
	bot := initBotApi()
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, _ := bot.GetUpdatesChan(u)
	// Say hello to the team
	if viper.IsSet("tgbotapi.groupBroadCast.defaultMessage") {
		sendToGroup(bot,
			viper.GetString("tgbotapi.groupBroadCast.defaultMessage"))
	}

	// Connect to database
	db := connectToDB()
	defer db.Close()

	// Main thread
	for update := range updates {
		if update.Message != nil {
			if update.Message.Location != nil {
				status, err := getWorkTracking(db, update.Message.From.ID)
				if err != nil {
					log.Error(err)
				}
				if status == 0 {
					employeeCheckIn(bot,
						update.Message.Chat.ID,
						update.Message.MessageID,
						update.Message.Location,
						update.Message.From.ID,
						update.Message.Date)
				} else if status == 1 {
					employeeCheckOut(bot,
						update.Message.Chat.ID,
						update.Message.MessageID,
						update.Message.Location,
						update.Message.From.ID,
						update.Message.Date)
				}
			} else if update.Message.Text == "/start" {
				startMenu(bot,
					update.Message.Chat.ID,
					update.Message.From.ID,
					update.Message.From.FirstName+update.Message.From.LastName)
			} else if update.Message.Text == "/help" {
				helpMenu(bot, update.Message.Chat.ID)
			} else if update.Message.Text == "/checkin" || update.Message.Text == "/checkout" {
				checkIn(bot, &update)
			} else if update.Message.Text == "/salary" {
				employeeSalaryReport(bot, update.Message.Chat.ID, update.Message.From.ID)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"* This report is not included with other allowances or bonuses.\n* The final report will be announced directly to you.")
				_, err := bot.Send(msg)
				if err != nil {
					log.Error(err)
				}
			} else if strings.HasPrefix(update.Message.Text, "/report") {
				// Check if user is admin or not
				if authManager(update.Message.From.ID) {
					err := reportManager(bot,
						update.Message.Text,
						update.Message.Chat.ID)
					if err != nil {
						log.Error(err)
					}
				} else {
					helpMenu(bot, update.Message.Chat.ID)
				}
			}

		} else if update.EditedMessage != nil {
			log.Printf("Message edited")
		}
	}
}

func initBotApi() *tgbotapi.BotAPI {
	bot, err := tgbotapi.NewBotAPI(viper.GetString("tgbotapi.token"))
	if err != nil {
		log.Error(err)
	}

	bot.Debug = viper.GetBool("tgbotapi.debug")

	log.Printf("Authorized on account %s", bot.Self.UserName)

	return bot
}

func ResetWorkTrackingJob() {
	// This func use to refresh all dynamic vairables
	job := cron.New()
	job.AddFunc("0 0 * * *", func() {
		ResetWorkTracking()
	})
	job.Start()
}
