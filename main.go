package main

import "fmt"
import "net"
import "time"
import "strings"
import "strconv"

type Config struct {
	IRCServer, Nickname, Username, Channel string
}

type IRC struct {
	server, nickname, username, channel string
	testconn net.Conn
	connected bool
}

func (i *IRC) WriteData(data string, v ...interface{}) {
	buffer := fmt.Sprintf(data, v...)
	_, err := i.testconn.Write([]byte(buffer))

	if err != nil {
		fmt.Println("Error, unable to send data.")
		i.connected = false
	}

	buffer = buffer[:len(buffer)-2]
	fmt.Printf("Sending: %s\n", buffer)
}

func (i *IRC) CloseConnection() {
	i.testconn.Close();
}

func (i *IRC) RecvData() {
	var completedBuffer = ""

	for {
		readBuffer := make([]byte, 1);
		_, err := i.testconn.Read(readBuffer);

		if err != nil {
			fmt.Println("Error, unable to receive data.")
			i.connected = false
			return;
		}

		s := string(readBuffer);

		if strings.Contains(completedBuffer, "\r") || strings.Contains(completedBuffer, "\n") {
			break;
		} else {
			completedBuffer += s
		}

		if err != nil {
			fmt.Println("Error, unable to recv data.")
		}
	}
		
	fmt.Printf("Received: %s\n", completedBuffer);
	i.HandleIRCEvents(completedBuffer)
}

func (i *IRC) ConnectToServer() bool { 
	ports := strings.Split(i.server, ":")[1]
	port, err := strconv.Atoi(ports)

	if err != nil {
		return false
	}
	if !((port >= 0) && (port <= 65535)) {
		return false
	}

	fmt.Printf("Attempting connection to %s..\n", i.server)
	i.testconn, err = net.Dial("tcp", i.server)
	
	if err != nil {
		fmt.Println("Error, unable to connect.")
		return false
	}

	i.connected = true
	return true
}

func (i *IRC) HandleIRCEvents(data string) {
	if (strings.Contains(data, "001")) {
		i.WriteData("JOIN %s\r\n", i.channel)
	} else if (strings.Contains(data, "433")) {
		i.nickname = i.nickname + "`"
		i.WriteData("NICK %s\r\n", i.nickname)
	} else if (strings.Contains(data, "PING")) {
		s := strings.Split(data, " ")[1]
		i.WriteData("PONG %s\r\n", s)
	}
	return;
}

func main() {
	cfg := Config{"irc.freenode.net:6667", "PugBotTestAA", "0 0 0 0", "#PugBotTest"}
	irc := IRC{cfg.IRCServer, cfg.Nickname, cfg.Username, cfg.Channel, nil, false}

	for {
		if (!irc.connected) {
			if (irc.ConnectToServer()) {
				fmt.Println("Connected!")
				irc.WriteData("NICK %s\r\n", cfg.Nickname);
				irc.WriteData("USER %s\r\n", cfg.Username);
				for {
					if (!irc.connected) {
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