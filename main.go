package main 

import (
	"fmt"
	"time"
	"strings"
	"log"
)


func main() {

	mconfig := make(map[string]string)
	fmt.Printf("Loading Config\n")
	lines, err := readConfig("config.ini")
	if err != nil {
		log.Fatalf("Error Reading Config File")
	}
	for _, line := range lines {
		if strings.Contains(line, "server="){
			mconfig["server"] = strings.TrimPrefix(line, "server=")
		}
		if strings.Contains(line, "password="){
			mconfig["password"] = strings.TrimPrefix(line, "password=")
		}
		if strings.Contains(line, "nickname="){
			mconfig["nickname"] = strings.TrimPrefix(line, "nickname=")
		}
		if strings.Contains(line, "username="){
			mconfig["username"] = strings.TrimPrefix(line, "username=")
		}
		if strings.Contains(line, "channel="){
			mconfig["channel"] = strings.TrimPrefix(line, "channel=")
		}
	}
	fmt.Println(mconfig["server"], mconfig["password"], mconfig["nickname"], mconfig["channel"]) //for ze testing
	
	irc := IRC{
		mconfig["server"], //server
		mconfig["password"], //password
		mconfig["nickname"],  //nickname
		mconfig["username"], //username
		mconfig["channel"], //channel
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

	irc.pug.SetAllowedMaps([]string{"de_dust2", "de_inferno", "de_nuke", "de_train", "de_mirage", "de_overpass", "de_cobblestone"})
	irc.cs.rconPassword = "Gibson"
	irc.cs.csServer = "192.168.182.1:27015"
	irc.cs.serverPassword = "Gibson"
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

