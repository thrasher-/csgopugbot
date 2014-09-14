package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
	"regexp"
)

type Message struct {
	nickname, host, destination, message string
}

type IRC struct {
	server, password, nickname, username, channel string
	socket net.Conn
	connected, ProtocolDebug bool

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
		if !irc.connected {
			fmt.Println("Exiting ping loop routine")
			break
		}

		if !irc.joinedChannel {
			time.Sleep(time.Second * 30)
			continue
		}

		if !irc.pingSent {
			irc.pingTime = time.Now()
			irc.WriteData("PING :TIMEOUTCHECK\r\n")
			irc.pingSent = true
		} else {
			if time.Since(irc.pingTime) / time.Second >= 10 && !irc.pongReceived {
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
	irc.socket.Close()
}

func (irc *IRC) RecvData() {
	var completedBuffer = ""

	for {
		readBuffer := make([]byte, 1)
		_, err := irc.socket.Read(readBuffer)

		if err != nil {
			fmt.Println("Error, unable to receive data.")
			irc.connected = false
			return
		}

		s := string(readBuffer)

		if strings.Contains(completedBuffer, "\r") {
			break
		} else {
			completedBuffer += s
		}
	}

	completedBuffer = strings.Trim(completedBuffer, "\r\n")
	fmt.Printf("Received: %s\n", completedBuffer)
	irc.HandleIRCEvents(completedBuffer)
}

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

func (irc *IRC) HandleIRCEvents(ircBuffer string) {
	if irc.ProtocolDebug {
		fmt.Printf("ircBuffer size: %d\n", len(ircBuffer))
		fmt.Printf("ircBuffer = %s\n", ircBuffer)
	}

	r, err := regexp.Compile(`^(?:[:](\S+) )?(\S+)(?: ([^:].+?))?(?: [:](.+))?$`)
	if err != nil {
		fmt.Println("Error, regex string wasn't able to compile")
		return
	}

	matches := r.FindAllStringSubmatch(ircBuffer, -1)
	if len(matches) == 0 {
		fmt.Printf("No match found for: %s\n", ircBuffer)
		return
	}

	for _, match := range matches {
		switch match[2] {
		case "001":
			irc.WriteData("JOIN %s\r\n", irc.channel)
			irc.joinedChannel = true
			irc.WriteData("USERHOST %s\r\n", irc.nickname)
		case "433":
			irc.nickname = irc.nickname + "`"
			irc.WriteData("NICK %s\r\n", irc.nickname)
		case "302":
			irc.cs.externalIP = strings.Split(match[4], "@")[1]
			fmt.Printf("Received external IP: %s\n", irc.cs.externalIP)
		case "PING":
			irc.WriteData("%s\r\n", strings.Replace(match[0], "PING", "PONG", 1))
		case "PONG":
			if match[4] == "TIMEOUTCHECK" {
				irc.pongReceived = true
			}
		case "NICK":
			if !irc.pug.PugStarted() {
				return
			}

			nickname := strings.Split(match[1], "!")[0]
			irc.pug.UpdatePlayerNickname(nickname, match[4])
		case "PART", "QUIT":
			if !irc.pug.PugStarted() {
				return
			}

			nickname := strings.Split(match[1], "!")[0]
			if irc.pug.LeavePug(nickname) {
				if irc.pug.GetPlayerCount() == 0 {
					irc.SendToChannel("The PUG admin has left the PUG and there are no other plays to assign the admin rights to. Type !pug <map> to start a new one.")
					irc.pug.EndPug()
				} else {
					if irc.pug.GetAdmin() == nickname {
						irc.pug.AssignNewAdmin()
						irc.SendToChannel("The PUG administrator has left the pug and %s has been asigned as the PUG admin.", irc.pug.GetAdmin())
					} else {
						irc.SendToChannel("%s has left the pug, [%d/10]", nickname, irc.pug.GetPlayerCount())
					}
				}
			}
		case "PRIVMSG":
			nickname := strings.Split(match[1], "!")[0]
			host := strings.Split(match[1], "@")[1]
			destination := match[2]
			msgBuf := match[4]
			message := strings.Split(msgBuf, " ")
			msg := Message{nickname, host, destination, msgBuf}
			fmt.Println(msg)

			if message[0] == "!pug" {
				if irc.pug.PugStarted() {
					irc.SendToChannel("A PUG has already been started, please wait until the next PUG has started.")
					return
				}
				if len(message) > 1 {
					s := message[1]
					if !irc.pug.IsValidMap(s) {
						irc.pug.SetMap("de_dust2")
					} else {
						irc.pug.SetMap(s)
					}
				}
				irc.pug.StartPug()
				irc.pug.JoinPug(nickname)
				irc.SendToChannel("A PUG has been started on map %s, type !join to join the pug", irc.pug.GetMap())
			} else if message[0] == "!join" {
				if irc.pug.PugStarted() {
					if irc.pug.JoinPug(nickname) {
						irc.SendToChannel("%s has joined the pug! [%d/10]", nickname, irc.pug.GetPlayerCount())
						return
					}
				} else {
					irc.SendToChannel("A PUG has not been started, type !pug <map> to start a new one.")
					return
				}
			} else if message[0] == "!leave" {
				if irc.pug.PugStarted() {
					if irc.pug.LeavePug(nickname) {
						if irc.pug.GetPlayerCount() == 0 {
							irc.SendToChannel("The PUG admin has left the PUG and there are no other plays to assign the admin rights to. Type !pug <map> to start a new one.")
							irc.pug.EndPug()
						} else {
							if irc.pug.GetAdmin() == nickname {
								irc.pug.AssignNewAdmin()
								irc.SendToChannel("The PUG administrator has left the pug and %s has been asigned as the PUG admin.", irc.pug.GetAdmin())
							} else {
								irc.SendToChannel("%s has left the pug, [%d/10]", nickname, irc.pug.GetPlayerCount())
							}
						}
					}
				} else {
					irc.SendToChannel("A PUG has not been started, type !pug <map> to start a new one.")
					return
				}
			} else if message[0] == "!stats" {
				irc.SendToChannel("Stats for player %s can be visited here: http://www.cs-stats.com/player/xxxxxx", nickname)
				return
			} else if message[0] == "!players" {
				if irc.pug.PugStarted() {
					irc.SendToChannel("Player list: %s [%d/10]", strings.Join(irc.pug.GetPlayers(), " "), irc.pug.GetPlayerCount())
					return
				}
			} else if message[0] == "!say" {
				if irc.cs.rconConnected {
					if len(message) > 1 {
						s := strings.Join(message[1:], " ")
						irc.cs.rc.WriteData("say [IRC] %s", s)
						irc.SendToChannel("Sent message to CS server.")
					}
				}
			}
		}
	}
}
