package main

import (
	"database/sql"
	"math"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/spf13/viper"
)

func haversine(defaultLat, defaultLon, lat, lon float64) float64 {
	const R = 6371 // Earth radius in km

	// Convert latitude and longitude to radians
	defaultLat = defaultLat * math.Pi / 180
	defaultLon = defaultLon * math.Pi / 180
	lat = lat * math.Pi / 180
	lon = lon * math.Pi / 180

	// Calculate the difference between the latitudes and longitudes
	dLat := lat - defaultLat
	dLon := lon - defaultLon

	// Apply the Haversine formula
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(defaultLat)*math.Cos(lat)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	d := R * c * 1000

	return d
}

func validateDistance(userLocation *tgbotapi.Location) (float64, string, int) {
	db := connectToDB()
	defer db.Close()
	stores := getStoreLocation(db)
	for _, store := range stores {
		d := (haversine(store.Latitude, store.Longitude,
			userLocation.Latitude, userLocation.Longitude))
		if int(math.Round(d)) <= viper.GetInt("checkin.maxDistance") {
			return d - 1, store.Name, store.StoreID
		}
	}
	return float64(viper.GetInt("checkin.maxDistance")) + 1, "", 0
}

func getStoreLocation(db *sql.DB) []Store {
	// Prepare query statement
	rows, err := db.Query("SELECT storeID, name, address, latitude, longitude FROM store WHERE active = 1")
	if err != nil {
		return nil
	}
	var stores []Store
	for rows.Next() {
		var store Store
		err = rows.Scan(&store.StoreID, &store.Name, &store.Address, &store.Latitude, &store.Longitude)
		if err != nil {
			return nil
		}
		stores = append(stores, store)
	}
	// Return results
	return stores
}
