package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	ClientId      string  `json:"client_id"`
	ClientSecret  string  `json:"client_secret"`
	RefreshToken  string  `json:"refresh_token"`
	CorporationId int32   `json:"corporation_id"`
	RegionIds     []int32 `json:"region_ids"`
	LocationIds   []int64 `json:"location_ids"`
}

func LoadConfig() (config Config, err error) {
	file, err := os.Open("config.json")
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
