package main

import (
	"fmt"
	"net"
	"strings"
	"time"
	"strconv"
)

type Message struct {
	nickname, host, destination, message string
}

type IRC struct {
	server, password, nickname, username, channel string
	socket net.Conn
	connected bool

	pingTime time.Time
	pingSent, pongReceived, joinedChannel bool
	pug PUG
	cs CS
}

func (irc *IRC) SendToChannel(data string, v ...interface{}) {
	buffer := fmt.Sprintf(data, v...)
	s := fmt.Sprintf("PRIVMSG %s :%s\r\n", irc.channel, buffer)
	_, err := irc.socket.Write([]byte(s))

	if err != nil {
		fmt.Println("Error, unable to send data.")
		irc.connected = false
	}

	fmt.Printf("Sending channel message: %s\n", buffer)
}

func (irc *IRC) WriteData(data string, v ...interface{}) {
	buffer := fmt.Sprintf(data, v...)
	_, err := irc.socket.Write([]byte(buffer))

	if err != nil {
		fmt.Println("Error, unable to send data.")
		irc.connected = false
	}

	buffer = strings.Trim(buffer, "\r\n")
	fmt.Printf("Sending: %s\n", buffer)
}

func (irc *IRC) PingLoop() {
	for {
		if (!irc.connected) {
			fmt.Println("Exiting ping loop routine")
			break
		}

		if (!irc.joinedChannel) {
			time.Sleep(time.Second * 30)
			continue
		}

		if (!irc.pingSent) {
			irc.pingTime = time.Now()
			irc.WriteData("PING :TIMEOUTCHECK\r\n")
			irc.pingSent = true
		} else {
			if (time.Since(irc.pingTime) / time.Second >= 10 && !irc.pongReceived) {
				fmt.Println("IRC connection has timed out.")
				irc.connected = false
				break
			} else {
				fmt.Println("Received pong to our ping.")
				irc.pingSent = false
				irc.pongReceived = false
				time.Sleep(time.Minute)
			}
		}
	}
}

func (irc *IRC) CloseConnection() {
	irc.socket.Close();
}

func (irc *IRC) RecvData() {
	var completedBuffer = ""

	for {
		readBuffer := make([]byte, 1);
		_, err := irc.socket.Read(readBuffer);

		if err != nil {
			fmt.Println("Error, unable to receive data.")
			irc.connected = false
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
	irc.HandleIRCEvents(strings.Split(completedBuffer, " "))}

func (irc *IRC) ConnectToServer() bool { 
	ports := strings.Split(irc.server, ":")[1]
	port, err := strconv.Atoi(ports)

	if err != nil {
		return false
	}
	if !((port >= 0) && (port <= 65535)) {
		return false
	}

	fmt.Printf("Attempting connection to %s..\n", irc.server)
	irc.socket, err = net.Dial("tcp", irc.server)
	
	if err != nil {
		fmt.Println("Error, unable to connect.")
		return false
	}

	irc.connected = true
	return true
}

func (irc *IRC) HandleIRCEvents(ircBuffer []string) {
	if (ircBuffer[1] == "001") {
		irc.WriteData("JOIN %s\r\n", irc.channel)
		irc.joinedChannel = true
		irc.WriteData("USERHOST %s\r\n", irc.nickname)
	} else if (ircBuffer[1] == "433") {
		irc.nickname = irc.nickname + "`"
		irc.WriteData("NICK %s\r\n", irc.nickname)
	} else if (ircBuffer[1] == "302") {
		irc.cs.externalIP = strings.Split(ircBuffer[3], "@")[1]
		fmt.Printf("Received external IP: %s\n", irc.cs.externalIP)
	} else if (ircBuffer[0] == "PING") {
		s := strings.Join(ircBuffer, " ")
		irc.WriteData("%s\r\n", strings.Replace(s, "PING", "PONG", 1))
	} else if (ircBuffer[1] == "PONG") && (ircBuffer[3] == ":TIMEOUTCHECK") {
		irc.pongReceived = true
	} else if (ircBuffer[1] == "PRIVMSG") {
		nickname := strings.Split(ircBuffer[0], "!")[0]
		nickname = strings.TrimPrefix(nickname, ":")
		host := strings.Split(ircBuffer[0], "@")[1]
		destination := ircBuffer[2]
		message := strings.Join(ircBuffer[3:], " ")
		message = strings.TrimPrefix(message, ":")
		msg := Message{nickname, host, destination, message}
		fmt.Println(msg)

		if (strings.Contains(message, "!pug") && strings.Contains(ircBuffer[3], "!pug")) {
			if (irc.pug.PugStarted()) {
				irc.SendToChannel("A PUG has already been started, please wait until the next PUG has started.")
				return;
			}
			if (len(ircBuffer) == 5) {
				s := ircBuffer[4];
				if (!irc.pug.IsValidMap(s)) {
					irc.pug.SetMap(s)
				}			
			} 
			irc.pug.StartPug()
			irc.pug.JoinPug(nickname)
			irc.SendToChannel("A PUG has been started on map %s, type !join to join the pug", irc.pug.GetMap())
		} else if (message == "!join") {
			if (irc.pug.PugStarted()) {
				if (irc.pug.JoinPug(nickname)) {
					irc.SendToChannel("%s has joined the pug! [%d/10]", nickname, irc.pug.GetPlayerCount())
					return;
				}
			} else {
				irc.SendToChannel("A PUG has not been started, type !pug <map> to start a new one.")
				return;
			}
		} else if (message == "!leave") {
			if (irc.pug.PugStarted()) {
				if (irc.pug.LeavePug(nickname)) {
					irc.SendToChannel("%s has left the pug [%d/10]", nickname, irc.pug.GetPlayerCount())
					return;
				}
			} else {
				irc.SendToChannel("A PUG has not been started, type !pug <map> to start a new one.")
				return;
			}
		} else if (message == "!cancelpug") {
			if (irc.pug.PugStarted()) {
				if (nickname == irc.pug.GetAdmin()) {
					irc.pug.EndPug();
					irc.SendToChannel("The PUG administrator has cancelled the pug. Type !pug <map> to start a new one.")
					return;
				}
			} else {
				irc.SendToChannel("A PUG has not been started, type !pug <map> to start a new one.")
				return;
			}
		} else if (message == "!stats") {
			irc.SendToChannel("Stats for player %s can be visited here: http://www.cs-stats.com/player/xxxxxx", nickname)
			return;
		} else if (message == "!players") { 
			if (irc.pug.PugStarted()) {
				irc.SendToChannel("Player list: %s [%d/10]", strings.Join(irc.pug.GetPlayers(), " "), irc.pug.GetPlayerCount())
				return;
			}
		} else if (strings.Contains(message, "!say") && strings.Contains(ircBuffer[3], "!say")) {
			if (irc.cs.rconConnected) {
				s := strings.Join(ircBuffer[4:], " ")
				irc.cs.rc.WriteData("say [IRC] %s", s)
			}
		}
	}
	return;
}
