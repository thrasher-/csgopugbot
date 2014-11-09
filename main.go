package main 

import (
	"time"
	"log"
	"strings"
)

func main() {
	log.Println("Loading config file config.json..")
	config, err := ReadConfig("config.json")

	if err != nil {
		log.Println("Fatal error opening config.json file. Error: ", err)
		return;
	}

	log.Println("Config file loaded.")
	SetAllowedMaps(strings.Split(config.CSMaps, ","))
	SetTeamName(strings.Split(config.TeamNames, ","))
	log.Println("Set available maps: " + GetValidMaps())
	log.Println("Testing connectivity to CS server(s)..")
	
	if !SetupAndTestCSServers(config.CSServers) {
		return;
	}

	irc = &IRC{
		config.IRCServer, //server
		config.IRCPassword, //password
		config.IRCNickname,  //nickname
		config.IRCUsername, //username
		"", //internal IP
		"", //external IP
		config.IRCChannels, //channel
		nil,  // net.Conn
		false, //irc connected
		false, //irc protocol debug
		Message{},
		time.Now(), // pingTime
		false, //ping sent
		false, //pong received 
		false, //joined channel
	}

	log.Println("Starting main IRC loop..")
	irc.IRCLoop()
}

