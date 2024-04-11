package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

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
