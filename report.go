package main

import (
	"fmt"
	"math"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/xuri/excelize/v2"
)

func generateExcel(monthlyReport []MonthlyReport, optionalFileName ...string) error {
	// Create a new Excel file
	file := excelize.NewFile()

	// Generate Salary sheet
	if err := salarySheet(file, monthlyReport); err != nil {
		log.Error("Generate salary sheet failed: %v", err)
		return err
	}

	// Generate Checkin Checkout time for all user passed in
	employeeIDs := extractEmployeeIDs(monthlyReport)
	month := monthlyReport[0].Month
	year := monthlyReport[0].Year
	for _, employee := range employeeIDs {
		if err := checkinSheet(file, employee, month, year); err != nil {
			log.Error("Generate checkin sheet failed for: %v - %v", employee, err)
			return err
		}
	}

	// If optionalFileName var is not nil, create a report to that specific path
	var fileSave string
	if optionalFileName != nil {
		fileSave = optionalFileName[0]
	} else {
		fileSave = viper.GetString("reportPath")
	}
	// Save file
	if err := file.SaveAs(fileSave); err != nil {
		log.Error("Canot save file: %v", err)
		return err
	}
	log.Info("Excel file generated successfully!")
	return nil
}

func generateReport(employeeId int, month, year string) []MonthlyReport {
	// Start db connect and calculate time
	db := connectToDB()
	defer db.Close()
	// Get name
	query := "SELECT name, salaryrate FROM employee WHERE employeeid = " + strconv.Itoa(employeeId)
	row, err := readFromDB(db, query)
	if err != nil {
		log.Error(err)
	}
	var name string
	var salaryRate int
	err = row.Scan(&name, &salaryRate)
	if err != nil {
		log.Error(err)
	}
	workRecords, err := getWorkRecords(db, employeeId, year, month, "")
	// Create slice of unique days
	var days []string
	for _, record := range workRecords {
		if !contains(days, record.Day) {
			days = append(days, record.Day)
		}
	}
	var dayWorkRecord []WorkRecord
	var monthlyReport []MonthlyReport
	for _, day := range days {
		for _, record := range workRecords {
			if record.Day == day {
				dayWorkRecord = append(dayWorkRecord, WorkRecord{
					EmployeeID: record.EmployeeID,
					Time:       record.Time,
					Year:       record.Year,
					Month:      record.Month,
					Day:        record.Day,
					Checkin:    record.Checkin,
					Checkout:   record.Checkout})
			}
		}
		dayDuration := calculateDurations(dayWorkRecord)
		monthlyReport = append(monthlyReport, MonthlyReport{
			EmployeeID:   employeeId,
			EmployeeName: name,
			Year:         year,
			Month:        month,
			Day:          day,
			Time:         timeToFloat(sumDurations(dayDuration)),
			SalaryRate:   salaryRate,
		})
		// Empty the dayWorkRecord after calculate complete
		dayWorkRecord = []WorkRecord{}
	}
	return monthlyReport
}

func generatedMonthlyReport(employeeId int, month, year string) []MonthlyReport {
	re := (`\d{1,2}`)
	if matchRegex(month, re) {
		log.Info("Found month send in digit format")
		monthInt, _ := strconv.Atoi(month)
		month = time.Month(monthInt).String()
	}
	var monthlyReport []MonthlyReport
	// Check if var employeeId int is empty or not
	if employeeId == 0 {
		db := connectToDB()
		defer db.Close()

		query := "SELECT employeeId FROM employee ORDER BY employeeId"
		rows, err := db.Query(query)
		if err != nil {
			log.Error(err)
		}
		defer rows.Close()
		log.Info("Generating montly report for all user")
		for rows.Next() {
			var id int
			err := rows.Scan(&id)
			if err != nil {
				log.Error(err)
			}
			report := generateReport(id, month, year)
			log.Debug("Report generated: ", report)
			monthlyReport = append(monthlyReport, report...)
		}

		if err = rows.Err(); err != nil {
			log.Error(err)
		}

	} else {
		// Generate for 1 employee
		monthlyReport = generateReport(employeeId, month, year)
	}
	log.Info("Generated monthly report: %v", monthlyReport)
	return monthlyReport
}

func salarySheet(file *excelize.File, monthlyReport []MonthlyReport) error {
	sheetName := "Salary"
	index, _ := file.NewSheet(sheetName)

	// Define styles to use in cell
	style, _ := file.NewStyle(&regularStyle)

	// Override the specific attributes for headerStyle
	headStyle, _ := file.NewStyle(&headerStyle)
	// Define the headers and their positions in a map
	headers := map[string]string{
		"B2": "Employee Name",
		"C2": "Working day",
		"D2": "Hours",
		"E2": "Salary Rate",
		"F2": "Total",
		"G2": "Estimate Salary",
	}
	// Define the money/currency shell
	finalStyle, _ := file.NewStyle(&criticalStyle)

	// Loop through the headers and set their values and style
	for pos, val := range headers {
		file.SetCellValue(sheetName, pos, val)
		file.SetCellStyle(sheetName, pos, pos, headStyle)
	}
	totalTimeByEmployee := make(map[string]float64)
	// Loop over the data and set the values
	for i, d := range monthlyReport {
		rowNum := i + 3 // Start at row 3
		// Set the User ID value
		employeeIDCell := fmt.Sprintf("B%d", rowNum)
		file.SetCellValue(sheetName, employeeIDCell, d.EmployeeName)

		// Set the Working day and Total hours values
		dayCell := fmt.Sprintf("C%d", rowNum)
		hoursCell := fmt.Sprintf("D%d", rowNum)
		salaryCell := fmt.Sprintf("E%d", rowNum)
		totalCell := fmt.Sprintf("F%d", rowNum)
		estimateSalaryCell := fmt.Sprintf("G%d", rowNum)

		// Add the time for this employee to the total time for their name
		totalTimeByEmployee[d.EmployeeName] += d.Time

		file.SetCellValue(sheetName, dayCell, d.Day+"/"+d.Month+"/"+d.Year)
		file.SetCellValue(sheetName, hoursCell, d.Time)
		file.SetCellValue(sheetName, salaryCell, d.SalaryRate)

		file.SetCellStyle(sheetName, employeeIDCell, employeeIDCell, style)
		file.SetCellStyle(sheetName, dayCell, dayCell, style)
		file.SetCellStyle(sheetName, hoursCell, hoursCell, style)
		file.SetCellStyle(sheetName, salaryCell, salaryCell, finalStyle)
		file.SetCellStyle(sheetName, totalCell, totalCell, style)
		file.SetCellStyle(sheetName, estimateSalaryCell, estimateSalaryCell, finalStyle)

	}

	// Merge the User ID cells with the same value
	numRows := len(monthlyReport) + 2 // Add 2 for the header row
	for i := 2; i <= numRows; i++ {
		// Create a map to store the total time for each employee

		curemployeeID, _ := file.GetCellValue(sheetName, fmt.Sprintf("B%d", i))
		prevemployeeID, _ := file.GetCellValue(sheetName, fmt.Sprintf("B%d", i-1))
		//
		if curemployeeID == prevemployeeID {
			// Merge User ID and Salary cell
			file.MergeCell(sheetName, fmt.Sprintf("B%d", i), fmt.Sprintf("B%d", i-1))
			file.MergeCell(sheetName, fmt.Sprintf("E%d", i), fmt.Sprintf("E%d", i-1))
			file.MergeCell(sheetName, fmt.Sprintf("F%d", i), fmt.Sprintf("F%d", i-1))
			file.MergeCell(sheetName, fmt.Sprintf("G%d", i), fmt.Sprintf("G%d", i-1))
		}
	}

	// Insert total hours and salary
	for i := 3; i <= numRows; i++ {
		curemployeeID, _ := file.GetCellValue(sheetName, fmt.Sprintf("B%d", i))
		salaryRateStr, _ := file.GetCellValue(sheetName, fmt.Sprintf("E%d", i))
		salaryRate, _ := strconv.ParseFloat(salaryRateStr, 64)
		salary := math.Round(totalTimeByEmployee[curemployeeID]) * salaryRate
		//
		file.SetCellValue(sheetName, fmt.Sprintf("F%d", i), totalTimeByEmployee[curemployeeID])
		file.SetCellValue(sheetName, fmt.Sprintf("G%d", i), salary)
	}

	// Calculate totalHours by employeeId

	// Set the active sheet and save the file
	file.SetActiveSheet(index)

	// set row height based on maximum character length
	file.AutoFilter(sheetName, "B:E", nil)
	file.SetColWidth(sheetName, "B", "G", 30)
	return nil
}

func checkinSheet(file *excelize.File, employeeId int, month, year string) error {
	// Start db connect and calculate time
	db := connectToDB()
	defer db.Close()
	// Get name
	query := "SELECT name FROM employee WHERE employeeid = " + strconv.Itoa(employeeId)
	row, err := readFromDB(db, query)
	if err != nil {
		return err
	}
	var employeeName string
	err = row.Scan(&employeeName)
	if err != nil {
		return err
	}
	workRecords, err := getWorkRecords(db, employeeId, year, month, "")
	sheetName := employeeName
	index, _ := file.NewSheet(sheetName)

	// Define styles to use in cell
	style, _ := file.NewStyle(&regularStyle)
	checkInStyle.Alignment.Horizontal = "left"
	inStyle, _ := file.NewStyle(&checkInStyle)
	checkOutStyle.Alignment.Horizontal = "right"
	outStyle, _ := file.NewStyle(&checkOutStyle)

	// Override the specific attributes for headerStyle
	headStyle, _ := file.NewStyle(&headerStyle)
	// Define the headers and their positions in a map
	headers := map[string]string{
		"B2": "Employee Name",
		"C2": "Working day",
		"D2": "CheckIn/Out",
	}
	// Loop through the headers and set their values and style
	for pos, val := range headers {
		file.SetCellValue(sheetName, pos, val)
		file.SetCellStyle(sheetName, pos, pos, headStyle)
	}
	for i, d := range workRecords {
		rowNum := i + 3 // Start at row 3
		employeeCell := fmt.Sprintf("B%d", rowNum)
		file.SetCellValue(sheetName, employeeCell, employeeName)
		// Set the Working day and Total hours values
		dayCell := fmt.Sprintf("C%d", rowNum)
		checkinCell := fmt.Sprintf("D%d", rowNum)
		file.SetCellValue(sheetName, dayCell, d.Day+"/"+d.Month+"/"+d.Year)

		file.SetCellStyle(sheetName, employeeCell, employeeCell, style)
		file.SetCellStyle(sheetName, dayCell, dayCell, style)

		if d.Checkin == true {
			file.SetCellValue(sheetName, checkinCell, "🟢"+d.Time)
			file.SetCellStyle(sheetName, checkinCell, checkinCell, inStyle)
		} else if d.Checkout == true {
			file.SetCellValue(sheetName, checkinCell, d.Time+"🟡")
			file.SetCellStyle(sheetName, checkinCell, checkinCell, outStyle)
		}

	}

	// Merge the User ID cells with the same value
	numRows := len(workRecords) + 2 // Add 2 for the header row
	for i := 2; i <= numRows; i++ {
		// Create a map to store the total time for each employee

		curemployeeID, _ := file.GetCellValue(sheetName, fmt.Sprintf("B%d", i))
		prevemployeeID, _ := file.GetCellValue(sheetName, fmt.Sprintf("B%d", i-1))
		curDay, _ := file.GetCellValue(sheetName, fmt.Sprintf("C%d", i))
		preDay, _ := file.GetCellValue(sheetName, fmt.Sprintf("C%d", i-1))

		if curemployeeID == prevemployeeID {
			// Merge User ID and Salary cell
			file.MergeCell(sheetName, fmt.Sprintf("B%d", i), fmt.Sprintf("B%d", i-1))
		}

		if curDay == preDay {
			// Merge User ID and Salary cell
			file.MergeCell(sheetName, fmt.Sprintf("C%d", i), fmt.Sprintf("C%d", i-1))
		}
	}

	// Set the active sheet and save the file
	file.SetActiveSheet(index)
	// set row height based on maximum character length
	file.AutoFilter(sheetName, "B:D", nil)
	file.SetColWidth(sheetName, "B", "D", 30)
	return nil
}

func extractEmployeeIDs(reports []MonthlyReport) []int {
	ids := make([]int, len(reports))
	for i, r := range reports {
		ids[i] = r.EmployeeID
	}
	return ids
}
