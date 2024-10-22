package game

import (
	"log"
	"math"
	"time"
	"vim-zombies/Utilities"
)

type Cursor struct {
	Row    int // 1 indexed `json: "row" vim:"row"`
	Column int // 0 indexed `json: "column" vim:"column"`
}

type LevelTime struct {
	StartMS    int64
	BestTimeMS int64
}

type Level struct {
	LevelName         string
	text              [][]byte
	bufferImmutable   bool
	levelState        [][]bool
	initialLevelState [][]bool
	LevelTime         *LevelTime
}

type NavigateLevel struct {
	Level
}

type CompletableLevel interface {
	FillTextBlanks()
	GetText() [][]byte
	GetBestTime() int64
	IsFinished() bool
	GetProhibtedInputs() []string
	IsBufferImmutable() bool
	CursorCallback(Cursor)
	startLevel()
	finishLevel() 
	resetLevel()
}

func FalseLvlStateFromText(text [][]byte) [][]bool {
	LevelState := make([][]bool, len(text))
	for i := range LevelState {
		LevelState[i] = make([]bool, len(text[i]))
	}

	return LevelState
}

func (lvl *Level) GetBestTime() int64 {
	return lvl.LevelTime.BestTimeMS
}

func (lvl *Level) GetProhibtedInputs() []string {
	return []string{}
}

func (lvl *Level) CursorCallback(_ Cursor) {
}

func (lvl *Level) IsBufferImmutable() bool {
	return lvl.bufferImmutable
}

func (lvl *Level) GetText() [][]byte {
	return lvl.text
}

func (lvl *Level) FillTextBlanks() {
	txt := lvl.text
	var max_len int = 0
	for _, v := range txt {
		if max_len < len(v) {
			max_len = len(v)
		}
	}
	// Then Fill in blanks
	txt_copy := make([][]byte, len(txt))
	for i, v := range txt {
		txt_copy[i] = make([]byte, max_len)
		copy(txt_copy[i], v)
		// Need to then append max_len - len(v) empty slots
		for j := 0; j < (max_len - len(v)); j++ {
			txt_copy[i][len(v)+j-1] = byte(' ')
		}
	}
	copy(lvl.text, txt_copy)
	lvl.text = txt_copy
}
func ConvertBytesToStrings(byteArray [][]byte) [][]string {
	stringArray := make([][]string, len(byteArray))
	for i, row := range byteArray {
		stringArray[i] = make([]string, len(row))
		for j, b := range row {
			stringArray[i][j] = string(b)
		}
	}

	return stringArray
}

func (lvl *Level) startLevel() {

	lvl.LevelTime = &LevelTime{
		StartMS:    time.Now().UnixMilli(),
		BestTimeMS: math.MaxInt64,
	}
}

func (lvl *Level) finishLevel() {

	completionTime := (time.Now().UnixMilli() - lvl.LevelTime.StartMS)
	if completionTime < lvl.LevelTime.BestTimeMS {
		lvl.LevelTime.BestTimeMS = completionTime
	}
	lvl.resetLevel() // Resetting the level so it can be played again
	log.Print(completionTime)

}

func (lvl *Level) resetLevel() {
	util.Copy2DArray(lvl.levelState, lvl.initialLevelState)
}
