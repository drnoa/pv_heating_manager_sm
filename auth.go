package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HeatingManager represents the main application struct.
type HeatingManager struct {
	Config      Config
	Token       string
	TokenExpiry time.Time
}

// getAuthToken gets or refreshes the authentication token as necessary.
//
// This function first checks if a token is present, and if not, performs a login.
// If a token is present, it checks if it has expired. If it has, the function performs a token refresh.
// If a valid token is present, the function does nothing.
//
// Returns:
// - error: If there was an error during the process of obtaining or refreshing the token.
func (hm *HeatingManager) getAuthToken() error {
	// Check if a token is present
	if hm.Token == "" {
		// If no token is present, perform a login
		return hm.login()
	}

	// Check if the token has expired
	if time.Now().After(hm.TokenExpiry) {
		// If the token has expired, perform a token refresh
		return hm.refreshToken()
	}

	// If a valid token is present, do nothing
	return nil
}

// login performs a login to obtain a new authentication token.
//
// The function sends a POST request to the Solar Manager API with the user's credentials.
// If the request is successful, a new token is obtained and stored in the HeatingManager struct.
//
// Returns:
// - error: If there was an error during the login process.
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
		ExpiresIn   int    `json:"expiresIn"` // Duration until the token expires in seconds
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("error decoding auth response: %v", err)
	}

	hm.Token = result.AccessToken
	// Set the expiry date of the token based on the current time plus the token's duration
	hm.TokenExpiry = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)

	return nil
}

// refreshToken refreshes the authentication token.
//
// The function sends a POST request to the Solar Manager API with the current token.
// If the request is successful, a new token is obtained and stored in the HeatingManager struct.
//
// Returns:
// - error: If there was an error during the refresh process.
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
		ExpiresIn   int    `json:"expiresIn"` // Duration until the token expires in seconds
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("error decoding refresh response: %v", err)
	}

	hm.Token = result.AccessToken
	hm.TokenExpiry = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)

	return nil
}
