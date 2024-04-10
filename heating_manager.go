package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
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
}

type HeatingManager struct {
	Config              Config
	TemperatureExceeded bool
	CheckInterval       time.Duration
	LastCheckFile       string
	Token               string
	TokenExpiry         time.Time
}

// NewHeatingManager creates a new HeatingManager instance.
func NewHeatingManager() (*HeatingManager, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}

	return &HeatingManager{
		Config:        config,
		CheckInterval: time.Duration(config.CheckInterval) * time.Minute,
		LastCheckFile: "lastCheck.txt",
	}, nil
}

// StartTemperatureMonitoring starts the temperature monitoring loop.
func (hm *HeatingManager) StartTemperatureMonitoring() {
	ticker := time.NewTicker(hm.CheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		hm.checkTemperature()
	}
}

// StartWeeklyCheck starts the weekly check loop.
func (hm *HeatingManager) StartWeeklyCheck() {
	weeklyCheckTimer := time.NewTimer(hm.nextWeeklyCheckDuration())
	defer weeklyCheckTimer.Stop()

	for range weeklyCheckTimer.C {
		hm.weeklyCheck()
		weeklyCheckTimer.Reset(hm.nextWeeklyCheckDuration())
	}
}

// loadConfig loads the application configuration from a JSON file.
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

// checkTemperature checks the temperature and sets the TemperatureExceeded flag accordingly.
func (hm *HeatingManager) checkTemperature() {
	temperature, err := hm.getTemperature()
	if err != nil {
		log.Printf("Failed to get temperature: %v", err)
		return
	}

	if temperature > hm.Config.TemperatureThreshold {
		fmt.Printf("Temperature has exceeded %.1f°C! Legionella heating will be rescheduled.\n", hm.Config.TemperatureThreshold)
		hm.TemperatureExceeded = true
	} else {
		fmt.Printf("Temperature is OK. Actual temperature: %.1f°C\n", temperature)
		hm.TemperatureExceeded = false
	}
}

// getTemperature gets the temperature from the Solar Manager API.
func (hm *HeatingManager) getTemperature() (float64, error) {
	if err := hm.getAuthToken(); err != nil {
		return 0, err
	}

	url := fmt.Sprintf("%s/%s", hm.Config.SolarManagerURL, hm.Config.SolarManagerSensorID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", hm.Token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to get temperature: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to get temperature: status code %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			CurrentWaterTemp float64 `json:"currentWaterTemp"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to unmarshal temperature response: %v", err)
	}

	return result.Data.CurrentWaterTemp, nil
}

// weeklyCheck checks if the temperature threshold has been exceeded and turns on the heating if necessary.
func (hm *HeatingManager) weeklyCheck() {
	if !hm.TemperatureExceeded {
		if err := hm.turnHeatingOn(); err != nil {
			log.Printf("Failed to turn on heating: %v", err)
		}

		// Schedule to turn off after a certain duration
		time.AfterFunc(4*time.Hour, func() {
			if err := hm.turnHeatingOff(); err != nil {
				log.Printf("Failed to turn off heating: %v", err)
			}
		})
	}
	hm.TemperatureExceeded = false
	hm.saveLastCheckTime()
}

// turnHeatingOn turns on the heating.
func (hm *HeatingManager) turnHeatingOn() error {
	if err := hm.getAuthToken(); err != nil {
		return fmt.Errorf("error getting auth token: %v", err)
	}

	url := fmt.Sprintf("https://cloud.solar-manager.ch/v1/control/heat-pump/%s", "xxx")
	requestBody, err := json.Marshal(map[string]interface{}{
		"heatPumpChargingMode": 1,
	})
	if err != nil {
		return fmt.Errorf("error marshalling request body: %v", err)
	}

	return hm.makeHeatingControlRequest(url, requestBody)
}

// turnHeatingOff turns off the heating.
func (hm *HeatingManager) turnHeatingOff() error {
	if err := hm.getAuthToken(); err != nil {
		return fmt.Errorf("error getting auth token: %v", err)
	}

	url := fmt.Sprintf("https://cloud.solar-manager.ch/v1/control/heat-pump/%s", "xxx")
	requestBody, err := json.Marshal(map[string]interface{}{
		"heatPumpChargingMode": 2,
	})
	if err != nil {
		return fmt.Errorf("error marshalling request body: %v", err)
	}

	return hm.makeHeatingControlRequest(url, requestBody)
}

// makeHeatingControlRequest makes a request to control the heating system.
func (hm *HeatingManager) makeHeatingControlRequest(url string, requestBody []byte) error {
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+hm.Token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error executing request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to modify heat pump state, status code: %d", resp.StatusCode)
	}

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Heat pump state changed successfully.")
	}
	return nil
}

// getAuthToken gets or refreshes the authentication token as necessary.
func (hm *HeatingManager) getAuthToken() error {
	// Wenn kein Token vorhanden ist, führe einen Login durch
	if hm.Token == "" {
		return hm.login()
	}

	// Wenn das Token vorhanden ist, aber abgelaufen, führe einen Refresh durch
	if time.Now().After(hm.TokenExpiry) {
		return hm.refreshToken()
	}

	// Wenn das Token vorhanden und gültig ist, mache nichts
	return nil
}

func (hm *HeatingManager) login() error {
	url := "https://cloud.solar-manager.ch/v1/oauth/login"
	credentials := map[string]string{
		"email":    hm.Config.Username,
		"password": hm.Config.Password,
	}
	credentialsJSON, err := json.Marshal(credentials)
	if err != nil {
		return fmt.Errorf("error marshalling credentials: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(credentialsJSON))
	if err != nil {
		return fmt.Errorf("error making auth request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication failed: status code %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken string `json:"accessToken"`
		ExpiresIn   int    `json:"expiresIn"` // Die Dauer bis zum Ablauf des Tokens in Sekunden
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("error decoding auth response: %v", err)
	}

	hm.Token = result.AccessToken
	// Setze das Ablaufdatum basierend auf dem aktuellen Zeitpunkt plus der Gültigkeitsdauer des Tokens
	hm.TokenExpiry = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)

	return nil
}

// refreshToken refreshes the authentication token.
func (hm *HeatingManager) refreshToken() error {
	url := "https://cloud.solar-manager.ch/v1/oauth/refresh"
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("error creating refresh request: %v", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", hm.Token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error executing refresh request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to refresh token: status code %d", resp.StatusCode)
	}

	var result struct {
		AccessToken string `json:"accessToken"`
		ExpiresIn   int    `json:"expiresIn"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("error decoding refresh response: %v", err)
	}

	hm.Token = result.AccessToken
	hm.TokenExpiry = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)

	return nil
}

// saveLastCheckTime saves the last check time to a file.
func (hm *HeatingManager) saveLastCheckTime() {
	now := time.Now()
	err := os.WriteFile(hm.LastCheckFile, []byte(now.Format(time.RFC3339)), 0644)
	if err != nil {
		log.Printf("Failed to save last check time: %v", err)
	}
}

// nextWeeklyCheckDuration calculates the duration until the next weekly check.
func (hm *HeatingManager) nextWeeklyCheckDuration() time.Duration {
	lastCheck, err := hm.readLastCheckTime()
	if err != nil {
		return 0
	}
	nextCheck := lastCheck.Add(time.Duration(hm.Config.WeeklyCheckInterval) * time.Hour)
	if time.Now().After(nextCheck) {
		return 0
	}
	return time.Until(nextCheck)
}

// readLastCheckTime reads the last check time from a file.
func (hm *HeatingManager) readLastCheckTime() (time.Time, error) {
	data, err := os.ReadFile(hm.LastCheckFile)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to read last check time: %w", err)
	}

	lastCheck, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse last check time: %w", err)
	}

	return lastCheck, nil
}
