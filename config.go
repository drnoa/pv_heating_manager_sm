// config.go

package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	SolarManagerURL      string  `json:"solarManagerURL"`
	SolarManagerSensorID string  `json:"solarManagerSensorID"`
	TemperatureThreshold float64 `json:"temperatureThreshold"`
	TemperatureTurnOff   float64 `json:"temperatureTurnOff"`
	CheckInterval        int     `json:"checkInterval"`
	WeeklyCheckInterval  int     `json:"weeklyCheckInterval"`
	Username             string  `json:"username"`
	Password             string  `json:"password"`
	HeatPumpID           string  `json:"heatPumpID"`
	HeatPumpControlURL   string  `json:"heatPumpControlURL"`
}

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
