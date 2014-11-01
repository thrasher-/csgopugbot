package main

import "fmt"

type Player struct {
	steamID, username, team string
	kills, deaths, assists int
}

type ScoreManager struct {
	firstHalfStarted, secondHalfStarted, matchCompleted bool
	firstHalfT, firstHalfCT int
	players []Player
	CTScore, TScore int
	CTsLeft, TsLeft int
}

func (sm *ScoreManager) AddPlayer(steamID, username string) {
	if sm.DoesPlayerExist(steamID, username) {
		return
	}

	player := Player{}
	player.steamID = steamID
	player.username = username
	player.team = ""
	player.kills = 0
	player.deaths = 0
	player.assists = 0

	sm.players = append(sm.players, player)
	fmt.Println("Added player: %s %s", steamID, username)
}

func (sm *ScoreManager) RemovePlayer(steamID, username string) {
	if !sm.DoesPlayerExist(steamID, username) {
		return
	}

	for i := range sm.players {
		if sm.players[i].steamID == steamID && sm.players[i].username == username && steamID != "BOT" {
			sm.players = append(sm.players[:i], sm.players[i+1:]...)
			return
		} else if sm.players[i].steamID == "BOT" && sm.players[i].username == username {
			sm.players = append(sm.players[:i], sm.players[i+1:]...)
			return
		}
	}
	return
}

func (sm *ScoreManager) DoesPlayerExist(steamID, username string) (bool) {
	for i := range sm.players {
		if sm.players[i].steamID == steamID && sm.players[i].username == username && steamID != "BOT" {
			return true
		} else if sm.players[i].steamID == "BOT" && sm.players[i].username == username {
			return true
		}
	}
	return false
}

func (sm *ScoreManager) FirstHalfStarted() (bool) {
	return sm.firstHalfStarted;
}

func (sm *ScoreManager) SetFirstHalfStarted(started bool) {
	sm.firstHalfStarted = started
}

func (sm *ScoreManager) SecondHalfStarted() (bool) {
	return sm.secondHalfStarted
}

func (sm *ScoreManager) SetSecondHalfStarted(started bool) {
	sm.secondHalfStarted = started
}

func (sm *ScoreManager) MatchCompleted() (bool) {
	return sm.matchCompleted
}

func (sm *ScoreManager) SetMatchCompleted(completed bool) {
	sm.matchCompleted = completed
}

func (sm *ScoreManager) SetCTScore(i int) {
	sm.CTScore = i
}

func (sm *ScoreManager) SetTScore(i int) {
	sm.TScore = i
}

func (sm *ScoreManager) GetCTScore() (int) {
	return sm.CTScore
}

func (sm *ScoreManager) GetTScore() (int) {
	return sm.TScore
}

func (sm *ScoreManager) SetCTsLeft(i int) {
	sm.CTsLeft = i
}

func (sm *ScoreManager) SetTsLeft(i int) {
	sm.TsLeft = i
}

func (sm *ScoreManager) GetCTsLeft() (int) {
	return sm.CTsLeft
}

func (sm *ScoreManager) GetTsLeft() (int) {
	return sm.TsLeft
}

func (sm *ScoreManager) GetFirstHalfT() (int) {
	return sm.firstHalfT
}

func (sm *ScoreManager) SetFirstHalfT(firstHalfT int) {
	sm.firstHalfT = firstHalfT;
}

func (sm *ScoreManager) GetFirstHalfCT() (int) {
	return sm.firstHalfCT
}

func (sm *ScoreManager) SetFirstHalfCT(firstHalfCT int) {
	sm.firstHalfCT = firstHalfCT;
}

func (sm *ScoreManager) ResetRoundPlayersLeft() {
	sm.CTsLeft = 5
	sm.TsLeft = 5
}

func (sm *ScoreManager) ResetRoundCounter() {
	sm.CTsLeft = 5
	sm.TsLeft = 5

	sm.CTScore = 0
	sm.TScore = 0
}

func (sm *ScoreManager) Reset() {
	sm.CTsLeft = 5
	sm.TsLeft = 5

	sm.CTScore = 0
	sm.TScore = 0

	sm.firstHalfStarted = false
	sm.secondHalfStarted = false
	sm.matchCompleted = false

	sm.firstHalfT = 0
	sm.firstHalfCT = 0
	sm.players = nil
}