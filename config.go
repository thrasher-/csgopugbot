package main

import (
	"io/ioutil"
	"encoding/json"
)

type Config struct {
	IRCServer string
	IRCPassword string
	IRCChannels []Channels
	IRCNickname string
	IRCUsername string
	CSServers []CSServers
	CSMaps string
	TeamNames string
	CSDefaultPugAdminPassword string
}

type Channels struct {
	Channel string
	Password string
	Region string
}

type CSServers struct {
	Server string
	RconPassword string
	ListenAddress string
	DefaultPugAdminPassword string
	Region string
	Log bool
}

func ReadConfig(path string) (Config, error) {
	file, err := ioutil.ReadFile(path)

	if err != nil {
		return Config{}, err
	}

	cfg := Config{}
	json.Unmarshal(file, &cfg)
	return cfg, nil
}
