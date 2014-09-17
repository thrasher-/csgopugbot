package main

type PlayerInfo struct {
	steamID string
	kills, deaths, assists int
}

type ScoreManager struct {
	players PlayerInfo
	CTScore, TScore int
	CTsLeft, TsLeft int
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

func (sm *ScoreManager) ResetRoundStats() {
	sm.CTsLeft = 5
	sm.TsLeft = 5
}

func (sm *ScoreManager) StartScoreManager() {
	sm.CTsLeft = 5
	sm.TsLeft = 5

	sm.CTScore = 0
	sm.TScore = 0
}