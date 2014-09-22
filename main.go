package main 

import (
	"fmt"
	"time"
	"strings"
)


func main() {

	fmt.Println("Loading config file config.ini..")
	config, err :=ProcessConfig("config.ini")

	if err != nil {
		fmt.Println("Fatal error opening config.ini file. Please ensure that the config file exists in the same directory as the PUGbot executable.")
		return;
	}

	fmt.Println("Config file loaded.")
	
	irc := IRC{
		config["ircServer"], //server
		config["ircPassword"], //password
		config["ircNickname"],  //nickname
		config["ircUsername"], //username
		config["ircChannel"], //channel
		nil,  // net.Conn
		false, //irc connected
		false, //irc protocol debug
		time.Now(), // pingTime
		false, //ping sent
		false, //pong received 
		false, //joined channel
		PUG{},
		CS{},
		ScoreManager{},
	}

	irc.pug.SetAllowedMaps(strings.Split(config["csMaps"], ","))
	irc.cs.rconPassword = config["csRconPassword"]
	irc.cs.csServer = config["csServer"]
	irc.cs.listenAddress = ":1337"
	irc.cs.pugAdminPassword = config["csPugAdminPassword"]

	if (irc.cs.ConnectToRcon()) {
		if (irc.cs.StartUDPServer()) {
			go func() {
				for {
					r, err := irc.cs.RecvData();

					if !err {
						fmt.Printf("Error with receiving CS UDP server buffer.")
						break;
					}
					irc.HandleCSBuffer(r, &irc.cs)
				}
			}()
			irc.cs.rc.WriteData("say PugBot connected")
			irc.cs.EnableLogging()
			irc.cs.relayGameEvents = false
			irc.cs.ProtocolDebug = false

			if (!irc.connected) {
				if (irc.ConnectToServer()) {
					fmt.Println("Connected to IRC server!")

					if (len(irc.password) > 0) {
						irc.WriteData("PASS %s\r\n", irc.password)
					}

					irc.WriteData("NICK %s\r\n", irc.nickname)
					irc.WriteData("USER %s\r\n", irc.username)

					fmt.Println("Starting ping check loop...")
					go irc.PingLoop()

					for {
						if (!irc.connected) {
							fmt.Println("Connection to IRC server has been lost.")
							break;
						}
						irc.RecvData()
					}
				} else {
					fmt.Println("Sleeping for 5 minutes before reconnection.")
					time.Sleep(time.Minute * 5)
				}
			}
		} else {
			fmt.Println("Fatal error, unable to start UDP server. Exiting.")
			return;
		}
	} else {
		fmt.Println("Fatal error, unable to connect to RCON. Exiting.")
		return;
	}
}

