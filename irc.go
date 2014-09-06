package main

import (
	"fmt"
	"net"
	"strings"
	"time"
	"strconv"
)

type IRC struct {
	server, password, nickname, username, channel string
	testconn net.Conn
	connected bool

	pingTime time.Time
	pingSent, pongReceived, joinedChannel bool
}

func (i *IRC) SendToChannel(data string, v ...interface{}) {
	buffer := fmt.Sprintf(data, v...)
	s := fmt.Sprintf("PRIVMSG %s :%s\r\n", i.channel, buffer)
	_, err := i.testconn.Write([]byte(s))

	if err != nil {
		fmt.Println("Error, unable to send data.")
		i.connected = false
	}

	fmt.Printf("Sending channel message: %s\n", buffer)
}

func (i *IRC) WriteData(data string, v ...interface{}) {
	buffer := fmt.Sprintf(data, v...)
	_, err := i.testconn.Write([]byte(buffer))

	if err != nil {
		fmt.Println("Error, unable to send data.")
		i.connected = false
	}

	buffer = strings.Trim(buffer, "\r\n")
	fmt.Printf("Sending: %s\n", buffer)
}

func (i *IRC) PingLoop() {
	for {
		if (!i.connected) {
			fmt.Println("Exiting ping loop routine")
			break
		}

		if (!i.joinedChannel) {
			time.Sleep(time.Second * 30)
			continue
		}

		if (!i.pingSent) {
			i.pingTime = time.Now()
			i.WriteData("PING :TIMEOUTCHECK\r\n")
			i.pingSent = true
		} else {
			if (time.Since(i.pingTime) / time.Second >= 10 && !i.pongReceived) {
				fmt.Println("IRC connection has timed out.")
				i.connected = false
				break
			} else {
				fmt.Println("Received pong to our ping.")
				i.pingSent = false
				i.pongReceived = false
				time.Sleep(time.Minute)
			}
		}
	}
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

		if strings.Contains(completedBuffer, "\r") {
			break;
		} else {
			completedBuffer += s
		}
	}
	
	completedBuffer = strings.Trim(completedBuffer, "\r\n")
	fmt.Printf("Received: %s\n", completedBuffer)
	i.HandleIRCEvents(strings.Split(completedBuffer, " "))}

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

func (i *IRC) HandleIRCEvents(ircBuffer []string) {
	if (ircBuffer[1] == "001") {
		i.WriteData("JOIN %s\r\n", i.channel)
		i.joinedChannel = true
	} else if (ircBuffer[1] == "433") {
		i.nickname = i.nickname + "`"
		i.WriteData("NICK %s\r\n", i.nickname)
	} else if (ircBuffer[0] == "PING") {
		s := strings.Join(ircBuffer, " ")
		i.WriteData("%s\r\n", strings.Replace(s, "PING", "PONG", 1))
	} else if (ircBuffer[1] == "PONG") && (ircBuffer[3] == ":TIMEOUTCHECK") {
		i.pongReceived = true
	} else if (ircBuffer[1] == "PRIVMSG") {
		nickname := strings.Split(ircBuffer[0], "!")[0]
		nickname = strings.TrimPrefix(nickname, ":")
		host := strings.Split(ircBuffer[0], "@")[1]
		destination := ircBuffer[2]
		message := strings.Join(ircBuffer[3:], " ")
		message = strings.TrimPrefix(message, ":")
		fmt.Printf("Nickname: %s Host: %s Destination: %s Message: %s", nickname, host, destination, message)

		if (strings.Contains(message, "!pug") && strings.Contains(ircBuffer[3], "!pug")) {
			if (len(ircBuffer) == 5) {
				s := ircBuffer[4];
				i.SendToChannel("A PUG has been started on map %s, type !join to join the pug", s)
			} else {
				i.SendToChannel("A PUG has been started, type !join to join the pug")
			}
		} else if (message == "!join") {
			i.SendToChannel("%s has joined the pug!", nickname)
		} else if (message == "!cancelpug") {
			i.SendToChannel("The PUG has been cancelled, type !pug to create one")
		} else if (message == "!stats") {
			i.SendToChannel("Stats for player %s can be visited here: http://www.cs-stats.com/player/xxxxxx", nickname)
		} else if (message == "!status") {
			i.SendToChannel("Meow")
		}
	}
	return;
}
