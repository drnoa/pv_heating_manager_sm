// heating_control.go

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HeatingManager represents the heating manager.
type HeatingManager struct {
	Config        Config
	Token         string
	TokenExpiry   time.Time
	CheckInterval time.Duration
	LastCheckFile string
}

// turnHeatingOn turns on the heating.
// It gets an authentication token, constructs the request URL and body,
// and makes a PUT request to control the heating system.
// Returns an error if any of the steps fail.
func (hm *HeatingManager) turnHeatingOn() error {
	// Get authentication token
	if err := hm.getAuthToken(); err != nil {
		return fmt.Errorf("error getting auth token: %v", err)
	}

	// Construct request URL and body
	url := fmt.Sprintf(hm.Config.HeatPumpControlURL, hm.Config.HeatPumpID)
	requestBody, err := json.Marshal(map[string]interface{}{
		"heatPumpChargingMode": 1,
	})
	if err != nil {
		return fmt.Errorf("error marshalling request body: %v", err)
	}

	// Make PUT request to control the heating system
	return hm.makeHeatingControlRequest(url, requestBody)
}

// turnHeatingOff turns off the heating.
// Steps are similar to turnHeatingOn().
func (hm *HeatingManager) turnHeatingOff() error {
	if err := hm.getAuthToken(); err != nil {
		return fmt.Errorf("error getting auth token: %v", err)
	}

	url := fmt.Sprintf(hm.Config.HeatPumpControlURL, hm.Config.HeatPumpID)
	requestBody, err := json.Marshal(map[string]interface{}{
		"heatPumpChargingMode": 2,
	})
	if err != nil {
		return fmt.Errorf("error marshalling request body: %v", err)
	}

	return hm.makeHeatingControlRequest(url, requestBody)
}

// makeHeatingControlRequest makes a request to control the heating system.
// The request is a PUT request with the provided URL and request body.
// Returns an error if the request fails.
func (hm *HeatingManager) makeHeatingControlRequest(url string, requestBody []byte) error {
	// Create request
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Set request headers
	req.Header.Set("Authorization", "Bearer "+hm.Token)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error executing request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to modify heat pump state, status code: %d", resp.StatusCode)
	}

	// Print success message if status code is OK
	if resp.StatusCode == http.StatusOK {
		fmt.Println("Heat pump state changed successfully.")
	}
	return nil
}
