package main

import (
	"encoding/json"
	"log"
	"os"
)

type serviceType struct {
	Type   string `json:"type"`
	Local  string `json:"local"`
	Remote string `json:"remote"`
}

type serverType struct {
	Host     string        `json:"host"`
	Port     int           `json:"port"`
	User     string        `json:"user"`
	Password string        `json:"password"`
	Key      string        `json:"key"`
	Services []serviceType `json:"services"`
}

type configType struct {
	Server []serverType `json:"server"`
}

var Config configType

func init() {
	file, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatal("Failed to read config file: ", err)
	}

	err = json.Unmarshal(file, &Config)
	if err != nil {
		log.Fatal("Failed to parse config file: ", err)
	}
}
