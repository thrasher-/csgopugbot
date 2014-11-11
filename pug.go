package main

import (
	"log"
	"strings"
	"strconv"
	"math/rand"
	"time"
)

const MAX_PLAYERS = 10
var validMaps, teamName []string
var pugManager []*PUG
var teamNameCT, teamNameT string

type PUG struct {
	pugID int
	mapName string
	ircChannel string
	players[] string
	pugAdmin string
	pugStarted, pugActive bool
}

func SetAllowedMaps(maps []string)  {
	validMaps = maps;
}

func GetValidMaps() string {
	return strings.Join(validMaps, " ")
}

func IsValidMap(mapName string) bool {
	validMaps := strings.Join(validMaps, " ")
	if (strings.Contains(validMaps, mapName)) {
		return true
	} else {
		return false
	}
}

func GetPugCounter() (int) {
	return len(pugManager)
}

func NewPug(pug *PUG) {
	if (len(pugManager) == 0) {
		pug.pugID = 0
	} else {
		pug.pugID = len(pugManager)
	}
	pugManager = append(pugManager, pug)
}

func GetPugByPlayer(player string) (*PUG, bool) {
	for i := range pugManager {
		if pugManager[i].GetPlayerByName(player) {
			return pugManager[i], true
		}
	}
	return nil, false
}

func GetPugByChannel(channel string) (*PUG, bool) {
	for i := range pugManager {
		if pugManager[i].GetIRCChannel() == channel {
			return pugManager[i], true
		}
	}
	return nil, false
}

func DeletePug(pugID int) {
	for i := range pugManager {
		if (pugID == pugManager[i].GetPugID()) {
			pugManager = append(pugManager[:i], pugManager[i+1:]...)
			return
		}
	}
}

func (p *PUG) GetPlayerByName(player string) (bool) {
	if (!p.pugStarted) {
		return false
	}

	for i := range p.players {
		if (player == p.players[i]) {
			return true
		}
	}

	return false
}
func (p *PUG) PugStarted() bool {
	return p.pugStarted
}

func (p *PUG) SetPugActive(active bool) {
	p.pugActive = active
}

func (p *PUG) PugActive() bool {
	return p.pugActive
}

func (p *PUG) GenerateRandomPassword(word string) string {
	rand.Seed(time.Now().UTC().UnixNano())
	i := rand.Intn(999-100) + 100
	return word + strconv.Itoa(i)
}

func (p *PUG) JoinPug(player string) bool {
	if (!p.pugStarted || len(p.players) == MAX_PLAYERS) {
		return false
	}

	if (len(p.players) == 0) {
		p.pugAdmin = player;
	}

	players := strings.Join(p.players, " ")
	if (strings.Contains(players, player)) {
		return false
	} else {
		p.players = append(p.players, player)
		return true
	}
}

func (p *PUG) GetPlayerCount() int {
	return len(p.players)
}

func (p *PUG) AssignNewAdmin() {
	if (p.GetPlayerCount() == 1) {
		p.pugAdmin = p.players[0]
		return;
	}

	rand.Seed(time.Now().UTC().UnixNano())
	i := rand.Intn(p.GetPlayerCount()-0)
	p.pugAdmin = p.players[i]
}

func (p *PUG) RandomisePlayerList() {
	rand.Seed(time.Now().UnixNano())

	for i := range p.players {
		j := rand.Intn(i + 1)
		p.players[i], p.players[j] = p.players[j], p.players[i]
	}
}

func (p *PUG) UpdatePlayerNickname(oldNick string, newNick string) {
	for i := range p.players {
		if (oldNick == p.players[i]) {
			p.players[i] = newNick
			if (p.pugAdmin == oldNick) {
				p.pugAdmin = newNick
			}
			break;
		}
	}
}

func (p *PUG) LeavePug(player string) bool {
	if (!p.pugStarted) {
		return false
	}
	
	for i := range p.players {
		if (player == p.players[i]) {
			p.players = append(p.players[:i], p.players[i+1:]...)
			return true
		}
	}
	return false
}

func (p *PUG) GetPugID() (int) {
	return p.pugID
}

func (p *PUG) SetMap(mapName string) {
	p.mapName = mapName
}

func (p *PUG) GetMap() string {
	return p.mapName
}

func (p *PUG) GetAdmin() string {
	return p.pugAdmin
}

func (p *PUG) SetIRCChannel(channel string) {
	p.ircChannel = channel
}

func (p *PUG) GetIRCChannel() string {
	return p.ircChannel
}

func (p *PUG) GetPlayers() []string {
	return p.players
}

func (p *PUG) StartPug() {
	if (p.pugStarted) {
		return;
	}
	
	if (len(p.mapName) > 0) && IsValidMap(p.mapName) {
		log.Printf("Pug map is %s", p.mapName)
	} else {
		p.mapName = "de_dust2"
	}

	p.pugStarted = true
}

func (p *PUG) EndPug() {
	if (!p.pugStarted) {
		return
	}

	p.pugStarted = false
	p.pugActive = false
	p.mapName = ""
	p.players = nil
	p.ircChannel = ""
}

func SetTeamName(names []string) {
	teamName = names
	teamNameCT, teamNameT = teamName[rand.Intn(len(teamName))], teamName[rand.Intn(len(teamName))] 
	for teamNameCT == teamNameT {
		teamNameT = teamName[rand.Intn(len(teamName))]

	}	
	log.Println("The set teams are: ", teamNameCT, " & ", teamNameT)
}

func GetTeamNameCT() string {
	return teamNameCT
}

func GetTeamNameT() string {
	return teamNameT
}


