package game

import (
	"vim-zombies/Utilities"
	"time"
	"math"
)


type Cursor struct {
	Row int // 1 indexed `json: "row" vim:"row"`
	Column int // 0 indexed `json: "column" vim:"column"`
}

type LevelTime struct{
	StartMS int64 
	BestTimeMS int64
}

type Level struct{
	LevelName string
	text [][]byte
	bufferImmutable bool
	levelState [][]bool
	initialLevelState [][]bool
	LevelTime *LevelTime
}

type NavigateLevel struct{
	Level
}

type CompletableLevel interface{
	FillTextBlanks()
	GetText() [][]byte
	GetBestTime() int64
	IsFinished() bool
	GetProhibtedInputs() []string
	IsBufferImmutable() bool
	CursorCallback(Cursor)
	startLevel()
	finishLevel() int64
	resetLevel()
}

func FalseLvlStateFromText(text [][]byte) [][]bool{
	LevelState:= make([][]bool, len(text))
	for i := range LevelState{
		LevelState[i] = make([]bool, len(text[i]))
	}

	return LevelState
}

func (navLvl *NavigateLevel) CursorCallBack(cursorPosition Cursor){
	navLvl.levelState[cursorPosition.Row][cursorPosition.Column] = true
}

func (navLvl *NavigateLevel) IsFinished() bool{
	finished := true 
	stringBuffer := ConvertBytesToStrings(navLvl.text)
	for i, row:= range(stringBuffer){
		for j, _ := range(row){
			if(stringBuffer[i][j] != " "){
				finished = finished && navLvl.levelState[i][j]
			}
		}
	}

	return finished
}

func (navLvl *NavigateLevel) GetProhibtedInputs() []string{
	return []string{"W", "<S-W>", "W", "<S-B>"} 
}

func NewNavigateLevel(name string, initalText [][]byte, bufferImmutable bool) NavigateLevel{
	// Creates a new level whose win condition is to navgiate to all non-space characters
	// in the given initialText buffer

	// Construct LevelState to be false everywhere initially
	LevelState := FalseLvlStateFromText(initalText)
	
	var copyInitialLevelState = make([][]bool, len(LevelState))
	util.Copy2DArrayBool(copyInitialLevelState, LevelState)
	return NavigateLevel{Level{
		LevelName: name,
		text: initalText,
		bufferImmutable: bufferImmutable,
		levelState: LevelState,
		initialLevelState: copyInitialLevelState,
		LevelTime: &LevelTime{
			BestTimeMS: math.MaxInt64,
		},
	}}
}

func (lvl *Level) GetBestTime() int64{
	return lvl.LevelTime.BestTimeMS
}

func (lvl *Level) GetProhibtedInputs() []string{
	return []string{}
}

func (lvl *Level) CursorCallback(_ Cursor) {
	return
}

func (lvl *Level) IsBufferImmutable() bool{
	return lvl.bufferImmutable
}

func (lvl *Level) GetText() [][]byte{
	return lvl.text
}

func (lvl *Level) FillTextBlanks() {
	txt := lvl.text
	var max_len int = 0
	for _,v := range(txt){
		if max_len < len(v){
			max_len = len(v)
		}
	}
	// Then Fill in blanks
	txt_copy := make([][]byte, len(txt))
	for i, v := range(txt){
		txt_copy[i] = make([]byte, max_len)
		copy(txt_copy[i], v)
		// Need to then append max_len - len(v) empty slots
		for j:= 0; j <(max_len-len(v)); j++{
			txt_copy[i][len(v) + j -1] = byte(' ')
		}
	}
	copy(lvl.text, txt_copy)
	lvl.text = txt_copy
}
func ConvertBytesToStrings(byteArray [][]byte) [][]string{
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
	lvl.LevelTime.StartMS = time.Now().UnixMilli()
}

func (lvl *Level) finishLevel() int64{

	completionTime := (time.Now().UnixMilli() - lvl.LevelTime.StartMS)
	if completionTime < lvl.LevelTime.BestTimeMS {
		lvl.LevelTime.BestTimeMS = completionTime
	}

	return completionTime

}

func (lvl *Level) resetLevel(){
	util.Copy2DArrayBool(lvl.levelState, lvl.initialLevelState)
}