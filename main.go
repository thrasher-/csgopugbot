package main

import "fmt"
import "net"
import "time"
import "strings"
import "strconv"
import "bufio"

//testing commit

type Config struct {
	IRCServer, Nickname, Username string
}

type IRC struct {
	server, nickname, username string
	testconn net.Conn
	connected bool
}

func (i *IRC) WriteData(data string) {
	_, err := i.testconn.Write([]byte(data));

	if err != nil {
		fmt.Println("Error, unable to send data.")
		i.connected = false;
	}

	fmt.Printf("Sending %s: ", data)
}

func (i *IRC) CloseConnection() {
	i.testconn.Close();
}

func (i *IRC) RecvData() {
	buffer := bufio.NewReaderSize(i.testconn, 512)
	reply, err := buffer.ReadString('\n')
		
	if err != nil {
		fmt.Println("Error, unable to receive data.")
		i.connected = false;
	}

	reply = reply[:len(reply)-2]
	fmt.Printf("Received: %s\n", reply)
}

func (i *IRC) ConnectToServer() bool { 
	ports := strings.Split(i.server, ":")[1]
	port, err := strconv.Atoi(ports)

	if err != nil {
		return false;
	}
	if !((port >= 0) && (port <= 65535)) {
		return false;
	}

	fmt.Printf("Attempting connection to %s..\n", i.server)
	i.testconn, err = net.Dial("tcp", i.server)

	if err != nil {
		fmt.Println("Error, unable to connect.")
		return false
	}
	return true
}

func main() {
	cfg := Config{"irc.freenode.net:6667", "PugBotTest", "0 0 0 0"}
	irc := IRC{cfg.IRCServer, cfg.Nickname, cfg.Username, nil, false}

	for {
		if (!irc.connected) {
			if (irc.ConnectToServer()) {
				fmt.Println("Connected!");
				buffer := fmt.Sprintf("NICK %s\r\nUSER %s\r\n", cfg.Nickname, cfg.Username);
				irc.WriteData(buffer)
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