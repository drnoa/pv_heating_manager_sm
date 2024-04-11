// auth.go

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

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
