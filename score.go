package main

import "fmt"

type Player struct {
	steamID, username, team string
	kills, deaths, assists, bombPlanted, bombDropped, bombPickedUp, targetBombed, bombDefused, bombDefuseAttemptWithKit, bombDefuseAttemptWithoutKit int
}

const (
	BOMB_PLANTED = iota
	BOMB_DROPPED
	BOMB_PICKED_UP
	TARGET_BOMBED
	BOMB_DEFUSED
	BOMB_DEFUSE_ATTEMPTED_WITH_KIT
	BOMB_DEFUSE_ATTEMPTED_WITHOUT_KIT
)

type ScoreManager struct {
	firstHalfStarted, secondHalfStarted, matchCompleted bool
	firstHalfT, firstHalfCT int
	players []Player
	playersStatsFirstHalf []Player
	playersStatsSecondHalf []Player
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
	player.bombDefused = 0
	player.bombPlanted = 0
	player.targetBombed = 0
	sm.players = append(sm.players, player)
	fmt.Printf("Added player: %s %s\n", steamID, username)
}

func (sm *ScoreManager) RemovePlayer(steamID, username string) {
	if !sm.DoesPlayerExist(steamID, username) {
		return
	}

	for i := range sm.players {
		if sm.players[i].steamID == steamID && steamID != "BOT" || sm.players[i].steamID == "BOT" && sm.players[i].username == username{
			sm.players = append(sm.players[:i], sm.players[i+1:]...)
			fmt.Printf("Removed player: %s %s\n", steamID, username)
			return
		}
	}
	return
}

func (sm *ScoreManager) DoesPlayerExist(steamID, username string) (bool) {
	for i := range sm.players {
		if sm.players[i].steamID == steamID && steamID != "BOT" || sm.players[i].steamID == "BOT" && sm.players[i].username == username {
			return true
		}
	}
	return false
}

func (sm *ScoreManager) AddEventStats(eventType int, steamID, username string) {
	for i := range sm.players {
		if sm.players[i].steamID == steamID && steamID != "BOT" || sm.players[i].steamID == "BOT" && sm.players[i].username == username {
			switch eventType {
				case BOMB_PLANTED:
					sm.players[i].bombPlanted += 1
				case BOMB_DROPPED:
					sm.players[i].bombDropped += 1
				case BOMB_PICKED_UP:
					sm.players[i].bombPickedUp += 1
				case TARGET_BOMBED:
					sm.players[i].targetBombed += 1
				case BOMB_DEFUSED:
					sm.players[i].bombDefused += 1
				case BOMB_DEFUSE_ATTEMPTED_WITH_KIT:
					sm.players[i].bombDefuseAttemptWithKit += 1
				case BOMB_DEFUSE_ATTEMPTED_WITHOUT_KIT:
					sm.players[i].bombDefuseAttemptWithoutKit += 1
			}
		} 
	}
}

func (sm *ScoreManager) AddKillAndDeathStats(p1steamID, p1username, p2steamID, p2username string) {
	for i := range sm.players {
		if sm.players[i].steamID == p1steamID && p1steamID != "BOT" || sm.players[i].steamID == "BOT" && sm.players[i].username == p1username {
			sm.players[i].kills += 1
		}
		if sm.players[i].steamID == p2steamID && p2steamID != "BOT" || sm.players[i].steamID == "BOT" && sm.players[i].username == p2username {
			sm.players[i].deaths += 1
		}
	}
}

func (sm *ScoreManager) ResetPlayerStats() {
	for i := range sm.players {
		sm.players[i].kills = 0
		sm.players[i].deaths = 0
		sm.players[i].assists = 0
		sm.players[i].bombPlanted = 0
		sm.players[i].bombDropped = 0
		sm.players[i].bombPickedUp = 0
		sm.players[i].targetBombed = 0
		sm.players[i].bombDefused = 0
		sm.players[i].bombDefuseAttemptWithKit = 0
		sm.players[i].bombDefuseAttemptWithoutKit = 0
	}
}

func (sm *ScoreManager) EnumerateStats() {
	for i := range sm.players {
		fmt.Printf("==============================================================\n")
		fmt.Printf("Player %s (%s)\n", sm.players[i].username, sm.players[i].steamID)
		fmt.Printf("Kills: %d\n", sm.players[i].kills)
		fmt.Printf("Deaths: %d\n", sm.players[i].deaths)
		fmt.Printf("Assists: %d\n", sm.players[i].assists)
		fmt.Printf("Bombs planted: %d\n", sm.players[i].bombPlanted)
		fmt.Printf("Bombs dropped: %d\n", sm.players[i].bombDropped)
		fmt.Printf("Bombs picked up: %d\n", sm.players[i].bombPickedUp)
		fmt.Printf("Target bombed: %d\n", sm.players[i].targetBombed)
		fmt.Printf("Bombs defused: %d\n", sm.players[i].bombDefused)
		fmt.Printf("Bombs defuse attempts with kit %d\n", sm.players[i].bombDefuseAttemptWithKit)
		fmt.Printf("Bombs defuse attempts without kit %d\n", sm.players[i].bombDefuseAttemptWithoutKit)
	}
}

func (sm *ScoreManager) PreservePlayerStatsFirstHalf() {
	copy(sm.playersStatsFirstHalf, sm.players)
}

func (sm *ScoreManager) SaveMatchData() {
	//todo
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
	copy(sm.playersStatsSecondHalf, sm.players)
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

	sm.players = nil
	sm.playersStatsFirstHalf = nil
	sm.playersStatsSecondHalf = nil

	sm.firstHalfStarted = false
	sm.secondHalfStarted = false
	sm.matchCompleted = false

	sm.firstHalfT = 0
	sm.firstHalfCT = 0
}