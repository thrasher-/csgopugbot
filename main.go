package main 

import (
	"fmt"
	"time"
)


func main() {

	irc := IRC{
		"irc.freenode.net:6667", //server
		"", //password
		"PugBotTest",  //nickname
		"0 0 0 0", //username
		"#PugBotTest", //channel
		nil,  // net.Conn
		false, //irc connected
		false, //irc protocol debug
		time.Now(), // pingTime
		false, //ping sent
		false, //pong received 
		false, //joined channel
		PUG{},
		CS{},
	}

	irc.pug.SetAllowedMaps([]string{"de_dust2", "de_inferno", "de_nuke", "de_train", "de_mirage", "de_overpass", "de_cobblestone"})
	irc.cs.rconPassword = "Gibson"
	irc.cs.csServer = "192.168.182.1:27016"
	irc.cs.listenAddress = ":1337"

	if (irc.cs.ConnectToRcon()) {
		if (irc.cs.StartUDPServer()) {
			go func() {
				for {
					r, err := irc.cs.RecvData();

					if !err {
						fmt.Printf("Error with receiving CS UDP server buffer.")
						break;
					}
					irc.HandleCSBuffer(r, irc.cs)
				}
			}()
			irc.cs.rc.WriteData("say PugBot connected")
			irc.cs.EnableLogging()
			irc.cs.ProtocolDebug = true
			irc.cs.pugPassword = "test123"

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


