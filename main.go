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
		time.Now(), // pingTime
		false, //ping sent
		false, //pong received 
		false, //joined channel
	}

	for {
		if (!irc.connected) {
			if (irc.ConnectToServer()) {
				fmt.Println("Connected!")

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
	}
}

