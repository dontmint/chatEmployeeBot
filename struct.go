package main

import "github.com/xuri/excelize/v2"

type Store struct {
	StoreID   int     `json:"storeId"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Active    int64   `json:"active"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Employee struct {
	EmployeeID int    `json:"employeeId"`
	Name       string `json:"name"`
	StoreID    int    `json:"storeId"`
	SalaryRate int    `json:"salaryRate"`
	Active     int    `json:"active"`
}

type Manager struct {
	ManagerID  int    `json:"managerId"`
	Name       string `json:"name"`
	StoreID    int    `json:"storeId"`
	SalaryRate int    `json:"salaryRate"`
	Active     int    `json:"active"`
}

type WorkTrackingStatus struct {
	EmployeeID int
	StoreID    int
	Working    int
}

type WorkRecord struct {
	EmployeeID int    `json:"employeeId"`
	Time       string `json:"time"`
	Year       string `json:"year"`
	Month      string `json:"month"`
	Day        string `json:"day"`
	Checkin    bool   `json:"checkin"`
	Checkout   bool   `json:"checkout"`
}

type MonthlyReport struct {
	EmployeeID   int     `json:"employeeId"`
	EmployeeName string  `json:"employeeName"`
	Year         string  `json:"year"`
	Month        string  `json:"month"`
	Day          string  `json:"day"`
	Time         float64 `json:"time"`
	SalaryRate   int     `json:"salaryRate"`
}

// Define the base style
var baseStyle = excelize.Style{
	Alignment: &excelize.Alignment{
		Vertical:   "center",
		Horizontal: "center",
	},
	Border: []excelize.Border{
		{
			Type:  "left",
			Color: "000000",
			Style: 1,
		},
		{
			Type:  "top",
			Color: "000000",
			Style: 1,
		},
		{
			Type:  "bottom",
			Color: "000000",
			Style: 1,
		},
		{
			Type:  "right",
			Color: "000000",
			Style: 1,
		},
	},
	Font: &excelize.Font{
		Bold:   false,
		Family: "Calibri",
		Size:   18,
		Color:  "000000",
	},
}
var regularStyle = baseStyle
var headerStyle = baseStyle
var criticalStyle = baseStyle
var checkInStyle = baseStyle
var checkOutStyle = baseStyle

func init() {
	// Override the defualt styles to use in cell
	regularStyle.Font.Size = 18
	regularStyle.Font.Bold = false

	// Override the specific attributes for headerStyle
	headerStyle.Font.Size = 22
	headerStyle.Font.Bold = true
	headerStyle.Fill = excelize.Fill{
		Type:    "pattern",
		Color:   []string{"#FFFFE0"}, // Light yellow
		Pattern: 1,
	}

	// Define a custom number format for currency shell
	customFormat := "#,##0 \"VND\""
	criticalStyle.CustomNumFmt = &customFormat
}
