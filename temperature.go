// temperature.go

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

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
