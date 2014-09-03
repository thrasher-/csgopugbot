package main 

import (
	"fmt"
	"time"
)

func main() {
	irc := IRC{"irc.freenode.net:6667", "PugBotTest", "0 0 0 0", "#PugBotTest", nil, false}

	for {
		if (!irc.connected) {
			if (irc.ConnectToServer()) {
				fmt.Println("Connected!")
				irc.WriteData("NICK %s\r\n", irc.nickname);
				irc.WriteData("USER %s\r\n", irc.username);
				for {
					irc.RecvData()
				}
			} else {
				fmt.Println("Sleeping for 5 minutes before reconnection.")
				time.Sleep(time.Minute * 5)
			}
		}
	}
}