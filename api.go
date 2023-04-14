package main

import (
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func startGinServer(listenAddr string) error {
	router := gin.Default()

	auth := router.Group("/")
	auth.Use(authMiddleware())

	// Add the authentication middleware
	//router.Use(authMiddleware())

	// Define your Gin routes here
	auth.GET("/", ApiWelcome)
	auth.GET("/api/v1/hours", ApiWorkingHours)
	auth.GET("/api/v1/salary", ApiSalary)
	router.GET("/api/v1/report", ApiDownloadReport)
	return router.Run(listenAddr)
}

func ApiWorkingHours(c *gin.Context) {
	// Get the query parameters from the URL
	employeeId, _ := strconv.Atoi(c.Query("employeeid"))
	day := c.Query("day")
	month := c.Query("month")
	year := c.Query("year")
	// Convert month from int to string if wrong format was sent to the API
	//re := (`^[01][0-9]$`)
	re := (`\d{1,2}`)
	if matchRegex(month, re) {
		log.Info("Found month send in digit format")
		monthInt, _ := strconv.Atoi(month)
		month = time.Month(monthInt).String()
	}
	// Start db connect and calculate time
	db := connectToDB()
	defer db.Close()
	workRecord, _ := getWorkRecords(db, employeeId, year, month, day)
	durations := sumDurations(calculateDurations(workRecord))
	// Get name
	query := "SELECT name FROM employee WHERE employeeid = " + strconv.Itoa(employeeId)
	row, err := readFromDB(db, query)
	if err != nil {
		log.Error(err)
	}
	var name string
	err = row.Scan(&name)
	if err != nil {
		log.Error(err)
	}
	// Return a response with the parameter values
	c.JSON(200, gin.H{
		"name":      name,
		"day":       day,
		"month":     month,
		"year":      year,
		"totalTime": durations,
	})
}

func ApiSalary(c *gin.Context) {
	// Start db connect and calculate time
	db := connectToDB()
	defer db.Close()
	// Get name
	employeeId, _ := strconv.Atoi(c.Query("employeeid"))
	query := "SELECT salaryrate FROM employee WHERE employeeid = " + strconv.Itoa(employeeId)
	row, err := readFromDB(db, query)
	if err != nil {
		log.Error(err)
	}
	var salary int
	err = row.Scan(&salary)
	// Return a response with the parameter values
	c.JSON(200, gin.H{
		"employeeid": employeeId,
		"salary":     salary,
	})
}

func ApiDownloadReport(c *gin.Context) {
	// Get the query parameters from the URL
	employeeId, _ := strconv.Atoi(c.Query("employeeid"))
	month := c.Query("month")
	year := c.Query("year")
	// Convert month from int to string if wrong format was sent to the API
	//re := (`^[01][0-9]$`)
	re := (`\d{1,2}`)
	if matchRegex(month, re) {
		log.Info("Found month send in digit format")
		monthInt, _ := strconv.Atoi(month)
		month = time.Month(monthInt).String()
	}
	monthlyReport := generatedMonthlyReport(employeeId, month, year)
	err := generateExcel(monthlyReport)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Generate report failed",
		})
	} else {
		filename := viper.GetString("reportPath")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", path.Base(filename)))
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.File(filename)

	}
}

func ApiWelcome(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Welcome to chatEmployee API",
	})
}

func matchRegex(str string, regex string) bool {
	match, err := regexp.MatchString(regex, str)
	if err != nil {
		// handle error
		return false
	}
	return match
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			return
		}

		tokenString := authHeader[7:]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// TODO: Replace this with your own secret key
			secret := []byte(viper.GetString("api.secret"))
			return secret, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		c.Next()
	}
}
