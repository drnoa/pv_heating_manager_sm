// config.go

package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config contains the configuration settings for the application.
type Config struct {
	// SolarManagerURL is the URL of the Solar Manager API.
	SolarManagerURL string `json:"solarManagerURL"`
	// SolarManagerSensorID is the ID of the Solar Manager sensor.
	SolarManagerSensorID string `json:"solarManagerSensorID"`
	// TemperatureThreshold is the threshold temperature at which the heating should turn on.
	TemperatureThreshold float64 `json:"temperatureThreshold"`
	// TemperatureTurnOff is the temperature at which the heating should turn off.
	TemperatureTurnOff float64 `json:"temperatureTurnOff"`
	// CheckInterval is the interval in minutes at which the temperature should be checked.
	CheckInterval int `json:"checkInterval"`
	// WeeklyCheckInterval is the interval in minutes at which a weekly check should be performed.
	WeeklyCheckInterval int `json:"weeklyCheckInterval"`
	// Username is the username for authentication with the Solar Manager API.
	Username string `json:"username"`
	// Password is the password for authentication with the Solar Manager API.
	Password string `json:"password"`
	// HeatPumpID is the ID of the heat pump to control.
	HeatPumpID string `json:"heatPumpID"`
	// HeatPumpControlURL is the URL of the heat pump control API.
	HeatPumpControlURL string `json:"heatPumpControlURL"`
}

// loadConfig loads the configuration from the config.json file.
// It returns the configuration and an error if any.
func loadConfig() (Config, error) {
	var config Config
	configFile, err := os.Open("config.json")
	if err != nil {
		return config, fmt.Errorf("failed to open config file: %v", err)
	}
	defer configFile.Close()

	err = json.NewDecoder(configFile).Decode(&config)
	if err != nil {
		return config, fmt.Errorf("failed to parse config file: %v", err)
	}

	return config, nil
}
