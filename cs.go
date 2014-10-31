package main

import (
	"errors"
	"net"
	"fmt"
	"strings"
	"strconv"
	"time"
)

var csManager []*CS

type CS struct {
	pugID int
	serverID int
	InUse, log bool
	SrvSocket *net.UDPConn
	rc RemoteConsole
	sm ScoreManager
	serverIP, rconPassword, localIP, serverPassword, listenAddress, region, pugAdminPassword, authSteamID, ircChannel, externalIP string
}

func GetFreeServer(region string) (*CS, bool) {
	for i := range csManager {
		if !csManager[i].InUse && csManager[i].region == region {
			return csManager[i], true
		}
	}
	return nil, false
}

func GetServerByID(serverID int) (*CS, bool) {
	for i := range csManager {
		if csManager[i].serverID == serverID {
			return csManager[i], true
		}
	}
	return nil, false
}

func GetServerByChannel(channel string) (*CS, bool) {
	for i := range csManager {
		if csManager[i].ircChannel == channel {
			return csManager[i], true
		}
	}
	return nil, false
}

func SetupAndTestCSServers(csServers []CSServers) (bool) {
	success := 0

	for i := 0; i < len(csServers); i++ {
		if NewCSServer(csServers[i].Server, csServers[i].RconPassword, "", csServers[i].ListenAddress, csServers[i].Region, "", csServers[i].Log) {
			success++
		}
	}

	if success == 0 {
		fmt.Println("Fatal error, unable to connect to any servers.")
		return false
	}
	return true
}

func GetCSServerCount() (int) {
	return len(csManager)
}

func NewCSServer(serverIP, rconPassword, serverPassword, listenAddress, region, ircChannel string, log bool) (bool) {
	cs := &CS{}

	if (len(csManager) == 0) {
		cs.serverID = 0
	} else {
		cs.serverID = len(csManager)
	}

	cs.serverIP = serverIP
	cs.rconPassword = rconPassword

	if !cs.ConnectToRcon() {
		fmt.Printf("Fatal error, unable to connect to CS server %s\n", serverIP)
		return false
	}

	cs.listenAddress = listenAddress
	if !cs.StartUDPServer() {
		fmt.Printf("Fatal error, unable to bind UDP server %s", listenAddress)
		return false
	}

	cs.serverPassword = serverPassword
	cs.InUse = false
	cs.log = log
	cs.ircChannel = ircChannel

	cs.EnableLogging()
	go cs.RecvData()
	csManager = append(csManager, cs)
	return true
}

func GetServer(serverID int) (*CS) {
	return csManager[serverID]
}

func DeleteServer(serverID int) {
	for i := range csManager {
		if (serverID == csManager[i].GetServerID()) {
			csManager = append(csManager[:i], csManager[i+1:]...)
			return
		}
	}
}

func (cs *CS) GetRegion() (string) {
	return cs.region
}

func (cs *CS) SetInUseStatus(inUse bool) {
	cs.InUse = inUse
}

func (cs *CS) GetServerIP() (string) {
	return cs.serverIP
}

func (cs *CS) SetIRCChannel(channel string) {
	cs.ircChannel = channel
}

func (cs *CS) GetIRCChannel() (string) {
	return cs.ircChannel
}

func (cs *CS) GetServerID() (int) {
	return cs.serverID
}

func (cs *CS) StartUDPServer() bool {
	fmt.Printf("Starting UDP server on %s", cs.listenAddress)
	addr, err := net.ResolveUDPAddr("udp", cs.listenAddress)

	if err != nil {
		fmt.Printf("Unable to resolve listen address")
		return false
	}

	cs.SrvSocket, err = net.ListenUDP("udp", addr)

	if err != nil {
		fmt.Printf("Unable to listen.")
		return false
	}

	fmt.Println("UDP server started successfully")
	return true
}

func (cs *CS) EnableLogging() {
	// to-do: check if cs server IP is private and use internal ip, otherwise use external ip
	port, _ := strconv.Atoi(strings.Split(cs.listenAddress, ":")[1]) 
	cs.rc.WriteData("logaddress_add %s:%d", cs.localIP, port)
	cs.rc.WriteData("log on")
}

func (cs *CS) ConnectToRcon() bool  {
	const timeout = 10 * time.Second
	err := errors.New("")
	cs.rc.conn, err = net.Dial("tcp", cs.serverIP)
	
	if err != nil {
		fmt.Println("Unable to connect to server RCON.")
		return false
	}

	cs.rc.reqid = 0x7fffffff
	var reqid int
	reqid, err = cs.rc.writeCmd(SERVERDATA_AUTH, cs.rconPassword)
	
	if err != nil {
		fmt.Printf("Error authenticating: %s\n", err)
		return false
	}

	cs.rc.readbuf = make([]byte, readBufferSize)

	var respType, requestId int
	respType, requestId, _, err = cs.rc.readResponse(timeout)
	if err != nil {
		fmt.Printf("Error reading response: %s\n", err)
		return false
	}

	if respType != SERVERDATA_AUTH_RESPONSE {
		respType, requestId, _, err = cs.rc.readResponse(timeout)
	}

	if err != nil {
		fmt.Printf("Error: %s", err)
		return false
	}

	if respType != SERVERDATA_AUTH_RESPONSE {
		fmt.Println(ErrInvalidAuthResponse)
		return false
	}
	if requestId != reqid {
		fmt.Println(ErrAuthFailed)
		return false
	}

	cs.localIP = strings.Split(cs.rc.conn.LocalAddr().String(), ":")[0]
	fmt.Printf("Internal IP: %s\n", cs.localIP)
	return true
}

func (cs *CS) RecvData() {
	for {
		buffer := make([]byte, 1024)
		rlen, _, err := cs.SrvSocket.ReadFromUDP(buffer)

		if err != nil {
			fmt.Printf("Unable to read data from UDP socket. Error: %s\n", err)
			break;
		}

		s := string(buffer)
		s = s[5:rlen-2]
		fmt.Printf("Received %d bytes: (%s)\n", rlen, s)
		cs.HandleCSBuffer(strings.Split(s, " "))
	}
}

func GetPlayerInfo(playerInfo string) (string, string, string, string) {
	if strings.Count(playerInfo, "<") == 0 {
		return "", "", "", ""
	}

	playerInfo = playerInfo[1:len(playerInfo)-1];
	splitBuffer := strings.SplitAfter(playerInfo, "<")
	player := strings.Trim(splitBuffer[0], "<")
	playerID := strings.TrimSuffix(splitBuffer[1], "><")
	steamID := strings.TrimSuffix(splitBuffer[2], "><")
	team := strings.TrimSuffix(splitBuffer[3], ">")
	return player, playerID, steamID, team
}

func (cs *CS) HandleCSBuffer(csBuffer []string) {
	if (cs.log) {
		fmt.Printf("csBuffer size: %d\n", len(csBuffer))
		for i, _ := range csBuffer {
			fmt.Printf("csBuffer[%d] = %s\n", i, csBuffer[i])
		}
	}
	if (csBuffer[5] == "entered") {
		player,_,steamID,_ := GetPlayerInfo(csBuffer[4])
		irc.SendToChannel("%s (%s) has entered the game.", player, steamID)
	} else if (csBuffer[5] == "purchased") {
		player,_,_,_ := GetPlayerInfo(csBuffer[4])
		item := csBuffer[6][1:len(csBuffer[6])-1];
		fmt.Printf("Player %s purchased %s\n", player, item)
		return;
	} else if csBuffer[5] == "triggered" && cs.log {
		player,_,_,_ := GetPlayerInfo(csBuffer[4])
		event := csBuffer[6][1:len(csBuffer[6])-1];

		switch event {
			case "Begin_Bomb_Defuse_Without_Kit": 
				irc.SendToChannel(irc.msg.destination, " %s started bomb defuse without kit.", player)
			case "Begin_Bomb_Defuse_With_Kit":
				irc.SendToChannel(irc.msg.destination," %s started bomb defuse with kit.", player)
			case "Dropped_The_Bomb":
				irc.SendToChannel(irc.msg.destination," %s dropped the bomb.", player)
			case "Planted_The_Bomb": 
				irc.SendToChannel(irc.msg.destination," %s planted the bomb.", player)
			case "Got_The_Bomb":
				irc.SendToChannel(irc.msg.destination," %s picked up the bomb.", player)
			case "Defused_The_Bomb":
				irc.SendToChannel(irc.msg.destination, "%s defused the bomb.", player)
			case "Round_Start":
				cs.sm.ResetRoundPlayersLeft()
			case "Round_End":
				cs.rc.WriteData("say			CT Score (%d)  			T Score (%d)		", cs.sm.GetCTScore(), cs.sm.GetTScore())
				irc.SendToChannel(irc.msg.destination, "			CT Score (%d)  			T Score (%d)		", cs.sm.GetCTScore(), cs.sm.GetTScore())
				irc.SendToChannel(irc.msg.destination, "******************** ROUND ENDED ********************")
				irc.SendToChannel(irc.msg.destination, "******************** ROUND STARTED ******************")
		}
		return;
	} else if csBuffer[6] == "triggered" && cs.log {
		event := csBuffer[7][1:len(csBuffer[7])-1];

		switch event {
			case "SFUI_Notice_Target_Bombed":
				cs.sm.SetTScore(cs.sm.GetTScore()+1)
				irc.SendToChannel(irc.msg.destination, "*** Target bombed successfully, the Terrorists win! ***")
			case "SFUI_Notice_Terrorists_Win":
				cs.sm.SetTScore(cs.sm.GetTScore()+1)
				irc.SendToChannel(irc.msg.destination, "******* All CT's eliminated, the Terrorists win! *******")
			case "SFUI_Notice_Bomb_Defused":
				cs.sm.SetCTScore(cs.sm.GetCTScore()+1)
				irc.SendToChannel(irc.msg.destination, "******* Bomb defused, the Counter-Terrorists win! ******")
			case "SFUI_Notice_CTs_Win":
				cs.sm.SetCTScore(cs.sm.GetCTScore()+1)
				irc.SendToChannel(irc.msg.destination, "*** All Terrorists eliminated, the Counter-Terrorists win! ***\n")
		}

		if cs.sm.FirstHalfStarted() {
			if cs.sm.GetCTScore() + cs.sm.GetTScore() == 15 {
				irc.SendToChannel(irc.msg.destination, "*** The first half has been completed.")
				cs.rc.WriteData("say The first half has been completed.")
				cs.rc.WriteData("mp_swapteams")
				cs.sm.SetFirstHalfT(cs.sm.GetTScore())
				cs.sm.SetFirstHalfCT(cs.sm.GetCTScore())
				cs.sm.SetTScore(0)
				cs.sm.SetCTScore(0)
			}
		}
		if cs.sm.SecondHalfStarted() {
			if cs.sm.GetCTScore() + cs.sm.GetFirstHalfT() == 16 {
				irc.SendToChannel(irc.msg.destination, "MATCH COMPLETED SUCCESSFULLY. The score was %d - %d", cs.sm.GetCTScore() + cs.sm.GetFirstHalfT(), cs.sm.GetTScore() + cs.sm.GetFirstHalfCT())
				cs.rc.WriteData("say MATCH COMPLETED SUCCESSFULLY. The Score was %d - %d", cs.sm.GetCTScore() + cs.sm.GetFirstHalfT(), cs.sm.GetTScore() + cs.sm.GetFirstHalfCT())
			} else if cs.sm.GetCTScore() + cs.sm.GetFirstHalfT() == 15 {
				irc.SendToChannel(irc.msg.destination, "MATCH COMPLETED SUCCESSFULLY. The match was a draw.")
				cs.rc.WriteData("say MATCH COMPLETED SUCCESSFULLY. The match was a draw.")
			} else if cs.sm.GetTScore() + cs.sm.GetFirstHalfCT()  == 16 {
				irc.SendToChannel(irc.msg.destination, "MATCH COMPLETED SUCCESSFULLY. The score was %d - %d", cs.sm.GetTScore() + cs.sm.GetFirstHalfCT(), cs.sm.GetCTScore() + cs.sm.GetFirstHalfT())
				cs.rc.WriteData("say MATCH COMPLETED SUCCESSFULLY. The Score was %d - %d", cs.sm.GetTScore() + cs.sm.GetFirstHalfCT(), cs.sm.GetCTScore() + cs.sm.GetFirstHalfT())
			}

			if cs.sm.MatchCompleted() {
				pug, _ := GetPugByChannel(cs.ircChannel)
				pug.EndPug()
				DeletePug(pug.GetPugID())
				cs.sm.Reset()
				cs.rc.WriteData("_restart") // kick all clients and set pw to a temp one
				cs.rc.WriteData("sv_password %s", pug.GenerateRandomPassword("temp"))
				irc.SendToChannel(irc.msg.destination, "The PUG has completed, type !pug <map> to start a new one!")
			}
		}
	} else if (csBuffer[5] == "say") {
		player,_,steamID,_ := GetPlayerInfo(csBuffer[4])
		fmt.Printf("Player %s said %s\n", player, strings.Join(csBuffer[6:], " "))
		message := strings.Join(csBuffer[6:], " ")
		message = message[1:len(message)-1]
		msg := strings.Split(message, " ")

		if (len(cs.authSteamID) == 0) {
			if (msg[0] == "!login" && len(msg) > 1) {
				password := msg[1];
				fmt.Printf("IN-GAME AUTH request: comparing '%s' to '%s'\n", password, cs.pugAdminPassword)
				if (password == cs.pugAdminPassword) {
					cs.rc.WriteData("say PUG admin rights has been granted to %s", player)
					irc.SendToChannel(irc.msg.destination, "PUG admin rights has been granted to %s", player)
					cs.authSteamID = steamID
					return;
				}
			} else {
				// bot doesn't allow any unauthenticated message handling
				return
			}
		} else {
			if (cs.authSteamID != steamID) {
				fmt.Println("Invalid auth attempt.")
				return
			}
		}
		if (msg[0] == "!lo3") {
			if !cs.sm.FirstHalfStarted() && !cs.sm.SecondHalfStarted() {
				cs.sm.ResetRoundCounter()
				cs.sm.SetFirstHalfStarted(true)
			} else if cs.sm.FirstHalfStarted() && cs.sm.GetFirstHalfT() + cs.sm.GetFirstHalfCT() < 15 {
				cs.rc.WriteData("say First half has already commenced. If you wish to cancel the first half, please type !cancelhalf.")
				return;
			} else if cs.sm.FirstHalfStarted() && cs.sm.GetFirstHalfT() + cs.sm.GetFirstHalfCT() == 15 {
				cs.sm.SetSecondHalfStarted(true)
				cs.rc.WriteData("say The second half has begun!")
				irc.SendToChannel(irc.msg.destination, "The second half has begun!")
			}
			if (!cs.log) {
				cs.log = true
				fmt.Println("Game event relaying enabled.")
			}
			cs.rc.WriteData("say Going Live on 1 restart..")
			cs.rc.WriteData("mp_restartgame 3")
			cs.rc.WriteData("say LIVE! LIVE! LIVE! Good luck and have fun")
			irc.SendToChannel(irc.msg.destination, "*** MATCH HAS GONE LIVE.")
			return;
		} else if (msg[0] == "!request") {
			cs.rc.WriteData("say Requesting for players on IRC.")
			irc.SendToChannel(irc.msg.destination, "Need player! To join, use the connect string: connect %s; password %s", cs.serverIP, cs.serverPassword)
			return;
		} else if (msg[0] == "!restart") {
			if cs.sm.FirstHalfStarted() || cs.sm.SecondHalfStarted() {
				cs.rc.WriteData("say You are unable to restart the round once the game has gone live.")
				return
			}
			cs.rc.WriteData("mp_restartgame 1")
			return;
		} else if (msg[0] == "!cancelhalf") {
			if cs.sm.FirstHalfStarted() && !cs.sm.SecondHalfStarted() {
				cs.sm.ResetRoundCounter()
				cs.sm.SetFirstHalfStarted(false)
				cs.log = false
				cs.rc.WriteData("say First half has been cancelled. Please type !lo3 once all players are ready.")
				irc.SendToChannel(irc.msg.destination, "*** First half has been cancelled.")
				return
			} else if cs.sm.FirstHalfStarted() && cs.sm.SecondHalfStarted() {
				cs.sm.ResetRoundCounter();
				cs.sm.SetSecondHalfStarted(false)
				cs.log = false
				cs.rc.WriteData("say Second half has been cancelled. Please type !lo3 once all players are ready.")
				irc.SendToChannel(irc.msg.destination, "*** Second half has been cancelled.")
				return
			}
		} else if (msg[0] == "!map" && len(msg) > 1) {
			if cs.sm.FirstHalfStarted() || cs.sm.SecondHalfStarted() {
				cs.rc.WriteData("say You are unable to change the map once the game has gone live.")
				return
			}

			mapName := msg[1];
			if !IsValidMap(mapName) {
				cs.rc.WriteData("Invalid map selection. Please select a map from the following: %s ", GetValidMaps())
				return
			}

			cs.rc.WriteData("say Changing map to '%s'.", mapName)
			cs.rc.WriteData("changelevel %s", mapName)
			irc.SendToChannel(irc.msg.destination, "PUG admin changed level to %s", mapName)
			return;
		} else if (msg[0] == "!irc") {
			if (len(msg) > 1) {
				s := strings.Join(msg[1:], " ")
				cs.rc.WriteData("say Sending message to IRC: %s.", s)
				irc.SendToChannel(irc.msg.destination, "[CS]: %s", s)
				return;
			}
		}
	}
	if len(csBuffer) >= 14 && cs.log {
		if (csBuffer[8] == "killed") {
			player1,_,_,team1 := GetPlayerInfo(csBuffer[4])
			player2,_,_,team2 := GetPlayerInfo(csBuffer[9])
			weapon := csBuffer[14][1:len(csBuffer[14])-1];
			headshot := ""

			if (len(csBuffer) == 16) {
				headshot = "(headshot)"
			}

			if team1 == "TERRORIST" && team2 == "CT" {
				cs.sm.SetCTsLeft(cs.sm.GetCTsLeft()-1)
				irc.SendToChannel("irc.msg.destination, %s (T) killed %s (CT) with %s %s [%d/5 left]\n", player1, player2, weapon, headshot, cs.sm.GetCTsLeft())
			} else if team1 == "CT" && team2 == "TERRORIST" {
				cs.sm.SetTsLeft(cs.sm.GetTsLeft()-1)
				irc.SendToChannel("irc.msg.destination, %s (CT) killed %s (T) with %s %s [%d/5 left]\n", player1, player2, weapon, headshot, cs.sm.GetTsLeft())	
			} else if team1 == "TERRORIST" && team2 == "TERRORIST" {
				cs.sm.SetTsLeft(cs.sm.GetTsLeft()-1)
				irc.SendToChannel(irc.msg.destination, "%s (T) killed %s (T) with %s %s [%d/5 left]\n", player1, player2, weapon, headshot, cs.sm.GetTsLeft())	
			} else if team1 == "CT" && team2 == "CT" {
				cs.sm.SetCTsLeft(cs.sm.GetCTsLeft()-1)
				irc.SendToChannel(irc.msg.destination, "%s (CT) killed %s (CT) with %s %s [%d/5 left]\n", player1, player2, weapon, headshot, cs.sm.GetCTsLeft())	
			}
		}
	}
}

