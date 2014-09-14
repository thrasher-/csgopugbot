package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"fmt"
	"strings"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	SERVERDATA_AUTH = 3
	SERVERDATA_EXECCOMMAND = 2
	SERVERDATA_AUTH_RESPONSE = 0
	SERVERDATA_RESPONSE_VALUE = 2
)

const readBufferSize = 4110

// RCON Protocol specs can be found here: https://developer.valvesoftware.com/wiki/Source_RCON_Protocol
// Thanks to https://github.com/james4k/ (james4k) for some of the RCON functions

type CS struct {
	listenAddress, rconPassword, pugPassword, authSteamID, localIP, externalIP, csServer string
	SrvSocket *net.UDPConn
	socket net.Conn
	rc RemoteConsole
	rconConnected, ProtocolDebug bool
}

type RemoteConsole struct {
	conn      net.Conn
	readbuf   []byte
	readmu    sync.Mutex
	reqid     int32
	queuedbuf []byte
}

var (
	ErrAuthFailed          = errors.New("rcon: authentication failed")
	ErrInvalidAuthResponse = errors.New("rcon: invalid response type during auth")
	ErrUnexpectedFormat    = errors.New("rcon: unexpected response format")
	ErrResponseTooLong     = errors.New("rcon: response too long")
)

func (r *RemoteConsole) WriteData(data string, v ...interface{}) (requestId int, err error){
	buffer := fmt.Sprintf(data, v...)
	fmt.Printf("Sent(RCON): %s\n", buffer)
	return r.writeCmd(SERVERDATA_EXECCOMMAND, buffer)
}

func (r *RemoteConsole) Read() (response string, requestId int, err error) {
	var respType int
	var respBytes []byte
	respType, requestId, respBytes, err = r.readResponse(2 * time.Minute)
	if err != nil || respType != SERVERDATA_RESPONSE_VALUE {
		response = ""
		requestId = 0
	} else {
		response = string(respBytes)
	}
	return
}

func (r *RemoteConsole) Close() error {
	return r.conn.Close()
}

func newRequestId(id int32) int32 {
	if id&0x0fffffff != id {
		return int32((time.Now().UnixNano() / 100000) % 100000)
	}
	return id + 1
}

func (r *RemoteConsole) writeCmd(cmdType int32, str string) (int, error) {
	buffer := bytes.NewBuffer(make([]byte, 0, 14+len(str)))
	reqid := atomic.LoadInt32(&r.reqid)
	reqid = newRequestId(reqid)
	atomic.StoreInt32(&r.reqid, reqid)

	binary.Write(buffer, binary.LittleEndian, int32(10+len(str)))
	binary.Write(buffer, binary.LittleEndian, int32(reqid))
	binary.Write(buffer, binary.LittleEndian, int32(cmdType))
	buffer.WriteString(str)
	binary.Write(buffer, binary.LittleEndian, byte(0))
	binary.Write(buffer, binary.LittleEndian, byte(0))

	r.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err := r.conn.Write(buffer.Bytes())
	return int(reqid), err
}

func (r *RemoteConsole) readResponse(timeout time.Duration) (int, int, []byte, error) {
	r.readmu.Lock()
	defer r.readmu.Unlock()

	r.conn.SetReadDeadline(time.Now().Add(timeout))
	var size int
	var err error
	if r.queuedbuf != nil {
		copy(r.readbuf, r.queuedbuf)
		size = len(r.queuedbuf)
		r.queuedbuf = nil
	} else {
		size, err = r.conn.Read(r.readbuf)
		if err != nil {
			return 0, 0, nil, err
		}
	}
	if size < 4 {
		// need the 4 byte packet size...
		s, err := r.conn.Read(r.readbuf[size:])
		if err != nil {
			return 0, 0, nil, err
		}
		size += s
	}

	var dataSize32 int32
	b := bytes.NewBuffer(r.readbuf[:size])
	binary.Read(b, binary.LittleEndian, &dataSize32)
	if dataSize32 < 10 {
		return 0, 0, nil, ErrUnexpectedFormat
	}

	totalSize := size
	dataSize := int(dataSize32)
	if dataSize > 4106 {
		return 0, 0, nil, ErrResponseTooLong
	}

	for dataSize+4 > totalSize {
		size, err := r.conn.Read(r.readbuf[totalSize:])
		if err != nil {
			return 0, 0, nil, err
		}
		totalSize += size
	}

	data := r.readbuf[4 : 4+dataSize]
	if totalSize > dataSize+4 {
		// start of the next buffer was at the end of this packet.
		// save it for the next read.
		r.queuedbuf = r.readbuf[4+dataSize : totalSize]
	}

	return r.readResponseData(data)
}

func (r *RemoteConsole) readResponseData(data []byte) (int, int, []byte, error) {
	var requestId, responseType int32
	var response []byte
	b := bytes.NewBuffer(data)
	binary.Read(b, binary.LittleEndian, &requestId)
	binary.Read(b, binary.LittleEndian, &responseType)
	response, err := b.ReadBytes(0x00)
	if err != nil && err != io.EOF {
		return 0, 0, nil, err
	}
	if err == nil {
		// if we didn't hit EOF, we have a null byte to remove
		response = response[:len(response)-1]
	}
	return int(responseType), int(requestId), response, nil
}


func (cs *CS) StartUDPServer() bool {
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
	cs.rc.conn, err = net.Dial("tcp", cs.csServer)
	
	if err != nil {
		fmt.Println("Unable to connect to RCON")
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
	cs.rconConnected = true
	return true
}

func (cs *CS) RecvData() ([]string, bool) {
	buffer := make([]byte, 1024)
	rlen, _, err := cs.SrvSocket.ReadFromUDP(buffer)

	if err != nil {
		fmt.Printf("Unable to read data from UDP socket. Error: %s\n", err)
		return nil, false
	}

	s := string(buffer)
	s = s[5:rlen-2]
	fmt.Printf("Received %d bytes: (%s)\n", rlen, s)
	return strings.Split(s, " "), true
}

func (irc *IRC) HandleCSBuffer(csBuffer []string, cs CS) {
	if (cs.ProtocolDebug) {
		fmt.Printf("csBuffer size: %d\n", len(csBuffer))
		for i, _ := range csBuffer {
			fmt.Printf("csBuffer[%d] = %s\n", i, csBuffer[i])
		}
	}
	if (csBuffer[5] == "entered") {
		player := csBuffer[4];
		irc.SendToChannel("%s has entered the game.", player)
	} else if (csBuffer[5] == "purchased") {
		player := csBuffer[4];
		item := csBuffer[6][1:len(csBuffer[6])-1];
		fmt.Printf("Player %s purchased %s\n", player, item)
		return;
	} else if (csBuffer[5] == "triggered") {
		player := csBuffer[4]
		event := csBuffer[6][1:len(csBuffer[6])-1];

		switch (event) {
			case "Begin_Bomb_Defuse_Without_Kit" : irc.SendToChannel("%s started bomb defuse without kit.", player) 
			case "Begin_Bomb_Defuse_With_Kit" : irc.SendToChannel("%s started bomb defuse with kit.", player)
			case "Dropped_The_Bomb" : irc.SendToChannel("%s dropped the bomb.", player)
			case "Planted_The_Bomb" : irc.SendToChannel("%s planted the bomb.", player)
			case "Got_The_Bomb" :  irc.SendToChannel("%s picked up the bomb.", player)
			case "Defused_The_Bomb" : irc.SendToChannel("%s defused the bomb.", player)
			case "Round_Start" : irc.SendToChannel("Round Started\n")
			case "Round_End" : irc.SendToChannel("Round ended\n")
		}
		return;
	} else if (csBuffer[5] == "say") {
		player := csBuffer[4];
		fmt.Printf("Player %s said %s\n", player, strings.Join(csBuffer[6:], " "))
		message := strings.Join(csBuffer[6:], " ")
		message = message[1:len(message)-1]
		msg := strings.Split(message, " ")

		if (len(cs.authSteamID) == 0) {
			if (msg[0] == "!login" && len(msg) > 1) {
				password := msg[1];
				fmt.Printf("IN-GAME AUTH request: comparing '%s' to '%s'", password, cs.pugPassword)
				if (password == cs.pugPassword) {
					cs.rc.WriteData("say PUG admin rights has been granted to %s", player)
					irc.SendToChannel("PUG admin rights has been granted to %s", player)
				}
			}
		}
		if (msg[0] == "!lo3") {
			cs.rc.WriteData("Going Live on 3 restarts..")
		} else if (msg[0] == "!request") {
			cs.rc.WriteData("say Requesting for players on IRC.")
			irc.SendToChannel("MATCH has gone live!")
			return;
		} else if (msg[0] == "!map" && len(msg) > 1) {
			mapName := msg[1];
			cs.rc.WriteData("say Changing map to '%s'.", mapName)
			cs.rc.WriteData("changelevel %s", mapName)
			irc.SendToChannel("PUG admin changed level to %s", mapName)
			return;
		} else if (msg[0] == "!irc") {
			if (len(msg) > 1) {
				s := strings.Join(msg[1:], " ")
				cs.rc.WriteData("say Sending message to IRC: %s.", s)
				irc.SendToChannel("[CS]: %s", s)
				return;
			}
		}
	}
	if (len(csBuffer) >= 14) {
		if (csBuffer[8] == "killed") {
			player1 := csBuffer[4]
			player2 := csBuffer[9];
			weapon := csBuffer[14][1:len(csBuffer[14])-1];

			if (len(csBuffer) == 16) {
				irc.SendToChannel("%s killed %s with %s (headshot)\n", player1, player2, weapon)
			} else {
				irc.SendToChannel("%s killed %s with %s\n", player1, player2, weapon)
			}
		}
	}
}

