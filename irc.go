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
	s := fmt.Sprintf("PRIVMSG %s :%s", i.channel, data)
	buffer := fmt.Sprintf(s, v...)
	_, err := i.testconn.Write([]byte(buffer))

	if err != nil {
		fmt.Println("Error, unable to send data.")
		i.connected = false
	}

	data = data[:len(data)-2]
	fmt.Printf("Sending channel message: %s\n", data)
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
			i.WriteData("PING :CHECKPNG\r\n")
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
		i.joinedChannel = true
	} else if (strings.Contains(data, "433")) {
		i.nickname = i.nickname + "`"
		i.WriteData("NICK %s\r\n", i.nickname)
	} else if (strings.Contains(data, "PING")) {
		s := strings.Split(data, " ")[1]
		i.WriteData("PONG %s\r\n", s)
	} else if (strings.Contains(data, "CHECKPNG")) {
		i.pongReceived = true
	}
	return;
}
