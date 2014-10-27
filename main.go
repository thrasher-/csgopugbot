package main 

import (
	"fmt"
	"time"
	"strings"
)

func main() {
	fmt.Println("Loading config file config.json..")
	config, err := ReadConfig("config.json")

	if err != nil {
		fmt.Println("Fatal error opening config.json file. Error: ", err)
		return;
	}

	fmt.Println("Config file loaded.")
	SetAllowedMaps(strings.Split(config.CSMaps, ","))
	fmt.Println("Set available maps: " + GetValidMaps())
	fmt.Println("Testing connectivity to CS server(s)..")
	
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

	fmt.Println("Starting main IRC loop..")
	irc.IRCLoop()
}

