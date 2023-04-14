package main

import (
	"fmt"
	"math"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func readConfig() {
	viper.SetConfigName("config")        // name of config file (without extension)
	viper.SetConfigType("yaml")          // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("$HOME/.config") // call multiple times to add many search paths
	viper.AddConfigPath(".")             // optionally look for config in the working directory
	err := viper.ReadInConfig()          // Find and read the config file
	if err != nil {                      // Handle errors reading the config file
		log.Error("Fatal error config file: %w \n", err)
	} else {
		log.Info("Config and secret are successfully loaded")
	}
}

// Helper function to check if slice contains a value
func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// This function takes a string representing a time in HH:MM:SS format and returns a float64 value that represents the total number of hours
// represented by that time string, rounded to one decimal place.
func timeToFloat(timeStr string) float64 {
	// Attempt to parse the given time string into a Go Time object using the specified format string "15:04:05"
	// If there is an error while parsing the time string, log the error and return 0
	t, err := time.Parse("15:04:05", timeStr)
	if err != nil {
		// Log the error message along with the error itself
		log.Error("Canot parse time: ", err)
		return 0 // Return 0 if there was an error while parsing the time string
	}

	// Calculate the total number of seconds represented by the given time string
	totalSeconds := float64(t.Hour()*3600 + t.Minute()*60 + t.Second())

	// Round the total seconds to one decimal place by dividing it by 36 (to convert to tenths of hours),
	// rounding to the nearest tenth using the math.Round() function, then dividing by 100 to convert back to hours.
	return math.Round(totalSeconds/36) / 100
}

// This function takes a Unix timestamp as input (an integer representing the number of seconds since January 1, 1970 UTC)
// It extracts and returns the hour and minute information as a string in "HH:MM" format, the day of the month as an integer,
// the name of the month as a string, and the year as an integer.
func extractDate(unixTime int) (hours string, day int, month string, year int) {
	// Convert the given Unix timestamp to a Go Time object
	t := time.Unix(int64(unixTime), 0)

	// Format the hour and minute as a string with leading zeros if necessary
	hours = fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute())

	// Get the day of the month as an integer
	day = t.Day()

	// Get the name of the month as a string
	month = t.Month().String()

	// Get the year as an integer
	year = t.Year()

	// Return the extracted date and time information
	return hours, day, month, year
}

// parseTime parses a string representing a time in RFC3339 format in a given timezone
// and returns a Time value with the parsed time
func parseTime(str string) time.Time {
	// Load location based on the configured timezone
	loc, err := time.LoadLocation(viper.GetString("timzeone"))
	if err != nil {
		return time.Time{} // return an empty time value if error occurs
	}

	// Parse the input string into a Time value using the specified timezone
	t, _ := time.ParseInLocation(time.RFC3339, str, loc)

	// Return the parsed time value
	return t
}

// getCurrentTime returns four strings representing the current time in a given timezone
func getCurrentTime() (string, string, string, string) {
	// Load location based on the configured timezone
	loc, err := time.LoadLocation(viper.GetString("timezone"))
	if err != nil {
		log.Error(err)
	}

	// Get the current time in the specified location
	currentTime := time.Now().In(loc)

	// Return four formatted strings representing the time
	return currentTime.Format("15:04"), // hours and minutes as "15:04"
		currentTime.Format("2"), // day of the month as "02" -> "2" fix bug
		currentTime.Format("January"), // Month as full name, e.g. "January"
		currentTime.Format("2006") // Year as four digits, e.g. "2006"
}
