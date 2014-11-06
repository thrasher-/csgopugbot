package main

import (
	"fmt"
	"net"
	"log"
	"strconv"
	"strings"
	"time"
	"regexp"
)

var irc *IRC

type Message struct {
	nickname, host, destination, message string
}

type IRC struct {
	server, password, nickname, username, internalIP, externalIP string
	ircChannels []Channels
	socket net.Conn
	connected, ProtocolDebug bool
	msg Message
	pingTime time.Time
	pingSent, pongReceived, joinedChannel bool
}

func (irc *IRC) SendToChannel(channel, data string, v ...interface{}) {
	buffer := fmt.Sprintf(data, v...)
	s := fmt.Sprintf("PRIVMSG %s :%s\r\n", channel, buffer)
	_, err := irc.socket.Write([]byte(s))

	if err != nil {
		log.Println("Error, unable to send data.")
		irc.connected = false
	}

	log.Printf("Sending channel message: %s\n", buffer)
}

func (irc *IRC) WriteData(data string, v ...interface{}) {
	buffer := fmt.Sprintf(data, v...)
	_, err := irc.socket.Write([]byte(buffer))

	if err != nil {
		log.Println("Error, unable to send data.")
		irc.connected = false
	}

	buffer = strings.Trim(buffer, "\r\n")
	log.Printf("Sending: %s\n", buffer)
}

func (irc *IRC) IRCLoop() {
	for {
		if (!irc.connected) {
			if irc.ConnectToServer() {
				log.Println("Connected to IRC server!")

				if (len(irc.password) > 0) {
					irc.WriteData("PASS %s\r\n", irc.password)
				}

				irc.WriteData("NICK %s\r\n", irc.nickname)
				irc.WriteData("USER %s %s %s %s\r\n", irc.username, irc.username, irc.username, irc.username)
				log.Println("Starting ping check loop...")

				go irc.PingLoop()

				for {
					if (!irc.connected) {
						log.Println("Connection to IRC server has been lost.")
						break
					} else {
						irc.RecvData()
					}
				} 
			} else {
				log.Println("Sleeping for 5 minutes before reconection.")
				time.Sleep(time.Minute * 5)
			}
		}
	}
}

func (irc *IRC) PingLoop() {
	for {
		if !irc.connected {
			log.Println("Exiting ping loop routine")
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
			irc.pongReceived = false
		} else {
			if time.Since(irc.pingTime) / time.Second >= 10 && !irc.pongReceived {
				log.Println("IRC connection has timed out.")
				irc.connected = false
				break
			} else if irc.pongReceived {
				log.Println("Received pong to our ping.")
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
			log.Println("Error, unable to receive data.")
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
	log.Printf("Received: %s\n", completedBuffer)
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

	log.Printf("Attempting connection to %s..\n", irc.server)
	irc.socket, err = net.Dial("tcp", irc.server)

	if err != nil {
		log.Println("Error, unable to connect.")
		return false
	}

	irc.connected = true
	return true
}

func (irc *IRC) HandleIRCEvents(ircBuffer string) {
	if irc.ProtocolDebug {
		log.Printf("ircBuffer size: %d\n", len(ircBuffer))
		log.Printf("ircBuffer = %s\n", ircBuffer)
	}

	r, err := regexp.Compile(`^(?:[:](\S+) )?(\S+)(?: ([^:].+?))?(?: [:](.+))?$`)
	if err != nil {
		log.Println("Error, regex string wasn't able to compile")
		return
	}

	matches := r.FindAllStringSubmatch(ircBuffer, -1)
	if len(matches) == 0 {
		log.Printf("No match found for: %s\n", ircBuffer)
		return
	}

	for _, match := range matches {
		switch match[2] {
		case "001":
			for i := 0; i < len(irc.ircChannels); i++ {
				irc.WriteData("JOIN %s %s\r\n", irc.ircChannels[i].Channel, irc.ircChannels[i].Password)
			}
			irc.joinedChannel = true
			irc.WriteData("USERHOST %s\r\n", irc.nickname)
		case "433":
			irc.nickname = irc.nickname + "`"
			irc.WriteData("NICK %s\r\n", irc.nickname)
		case "302":
			irc.externalIP = strings.Split(match[4], "@")[1]
			log.Printf("Received external IP: %s\n", irc.externalIP)
		case "PING":
			irc.WriteData("%s\r\n", strings.Replace(match[0], "PING", "PONG", 1))
		case "PONG":
			if match[4] == "TIMEOUTCHECK" {
				irc.pongReceived = true
			}
		case "NICK":
			nickname := strings.Split(match[0], "!")[0]
			nickname = nickname[1:]
			pug, success := GetPugByPlayer(nickname)

			if !success {
				return
			}

			pug.UpdatePlayerNickname(nickname, match[4])
		case "PART", "QUIT":
			nickname := strings.Split(match[0], "!")[0]
			nickname = nickname[1:]
			pug, success := GetPugByPlayer(nickname)

			if !success {
				return
			}

			channel := pug.GetIRCChannel()

			if pug.LeavePug(nickname) && !pug.PugActive() {
				if pug.GetPlayerCount() == 0 {
					irc.SendToChannel(channel, "The PUG admin has left the PUG and there are no other plays to assign the admin rights to. Type !pug <map> to start a new one.")
					pug.EndPug()
					DeletePug(pug.GetPugID())
					cs, _ := GetServerByChannel(channel)
					cs.SetInUseStatus(false)
					cs.SetIRCChannel("")
				} else {
					if pug.GetAdmin() == nickname {
						pug.AssignNewAdmin()
						irc.SendToChannel(channel, "The PUG administrator has left the pug and %s has been asigned as the PUG admin.", pug.GetAdmin())
					} else {
						irc.SendToChannel(channel, "%s has left the pug, [%d/10]", nickname, pug.GetPlayerCount())
					}
				}
			}
		case "PRIVMSG":
			nickname := strings.Split(match[1], "!")[0]
			host := strings.Split(match[1], "@")[1]
			destination := match[3]
			msgBuf := match[4]
			message := strings.Split(msgBuf, " ")
			irc.msg = Message{nickname, host, destination, msgBuf}

			if !strings.Contains(destination, "#") {
				return;
			}

			_, pugStarted := GetPugByChannel(destination)

			if message[0] == "!pug" {
				if pugStarted {
					irc.SendToChannel(destination, "A PUG has already been started, please wait until the next PUG has started.")
					return
				}

				cs, success := GetFreeServer("Sydney") // to-do - obtain channel region
					
				if !success {
					irc.SendToChannel(destination, "Unable to discover any available servers with specified region.")
					return
				}

				cs.SetInUseStatus(true)
				cs.SetIRCChannel(destination)
				
				p := &PUG{}

				if len(message) > 1 {
					s := message[1]
					if !IsValidMap(s) {
						p.SetMap("de_dust2")
					} else {
						p.SetMap(s)
					}
				}

				irc.SendToChannel(destination, "A PUG has been started on map %s, type !join to join the pug", p.GetMap())
				log.Printf("Assigned server ID, region %s to pug ID %d with channel %s\n", cs.GetRegion(), p.GetPugID(), destination)
				cs.rc.WriteData("changelevel %s", p.GetMap())
				cs.serverPassword = p.GenerateRandomPassword("pug")
				cs.rc.WriteData("sv_password %s", cs.serverPassword)
				cs.pugAdminPassword = p.GenerateRandomPassword("admin")
				p.StartPug()
				p.SetIRCChannel(destination)
				p.JoinPug(nickname)
				NewPug(p)
			} else if message[0] == "!join" {
				if pugStarted {
					pug, _ := GetPugByChannel(destination)
					if pug.JoinPug(nickname) && !pug.PugActive() {
						irc.SendToChannel(destination, "%s has joined the pug! [%d/10]", nickname, pug.GetPlayerCount())
						if pug.GetPlayerCount() < 10 {
							return
						}
						irc.SendToChannel(destination, "The PUG is now full! The server information will be messaged to you.")
						pug.RandomisePlayerList()
						pug.SetPugActive(true)
						
						players := pug.GetPlayers()
						irc.SendToChannel(destination, "The teams are as follows. Terrorists: %s Counter-Terrorists: %s", strings.Join(players[0:5], " "), strings.Join(players[5:10], " "))
						cs, _ := GetServerByChannel(destination)

						for i := range players {
							if players[i] == pug.GetAdmin() {
								irc.WriteData("PRIVMSG %s :PUG details are: connect %s; password %s. PUG Admin password: %s (type !login <password> in game and !lo3 once all players are ready).\r\n", players[i], cs.serverIP, cs.serverPassword, cs.pugAdminPassword)
							} else {
								irc.WriteData("PRIVMSG %s :PUG details are: connect %s; password %s.\r\n", players[i], cs.serverIP, cs.serverPassword)
							}
						}
					}
				} else {
					irc.SendToChannel(destination, "A PUG has not been started, type !pug <map> to start a new one.")
					return
				}
			} else if message[0] == "!leave" {
				if pugStarted {
					pug, _ := GetPugByChannel(destination)
					if pug.LeavePug(nickname) && !pug.PugActive() {
						if pug.GetPlayerCount() == 0 {
							irc.SendToChannel(destination, "The PUG admin has left the PUG and there are no other plays to assign the admin rights to. Type !pug <map> to start a new one.")
							pug.EndPug()
							DeletePug(pug.GetPugID())
							cs, _ := GetServerByChannel(destination)
							cs.SetInUseStatus(false)
							cs.SetIRCChannel("")
						} else {
							if pug.GetAdmin() == nickname {
								pug.AssignNewAdmin()
								irc.SendToChannel(destination, "The PUG administrator has left the pug and %s has been asigned as the PUG admin.", pug.GetAdmin())
							} else {
								irc.SendToChannel(destination, "%s has left the pug, [%d/10]", nickname, pug.GetPlayerCount())
							}
						}
					}
				} else {
					irc.SendToChannel(destination, "A PUG has not been started, type !pug <map> to start a new one.")
					return
				}
			} else if message[0] == "!stats" {
				irc.SendToChannel(destination, "Stats for player %s can be visited here: http://www.cs-stats.com/player/xxxxxx", nickname)
				return
			} else if message[0] == "!players" {
				if pugStarted {
					pug, _ := GetPugByChannel(destination)
					irc.SendToChannel(destination, "Player list: %s [%d/10]", strings.Join(pug.GetPlayers(), " "), pug.GetPlayerCount())
					return
				}
			} else if message[0] == "!say" {
				if len(message) > 1 {
					cs, success := GetServerByChannel(destination)
					
					if !success {
						log.Println("Unable to find server")
						return
					}

					s := strings.Join(message[1:], " ")
					cs.rc.WriteData("say [IRC] %s", s)
					irc.SendToChannel(destination, "Sent message to CS server.")
				}
			}
		}
	}
}
