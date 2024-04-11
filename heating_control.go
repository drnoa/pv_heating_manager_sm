// heating_control.go

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// turnHeatingOn turns on the heating.
func (hm *HeatingManager) turnHeatingOn() error {
	if err := hm.getAuthToken(); err != nil {
		return fmt.Errorf("error getting auth token: %v", err)
	}

	url := fmt.Sprintf(hm.Config.HeatPumpControlURL, hm.Config.HeatPumpID)
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
