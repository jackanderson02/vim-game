package game

import (
	"vim-zombies/Utilities"
)


type Cursor struct {
	Row int // 1 indexed `json: "row" vim:"row"`
	Column int // 0 indexed `json: "column" vim:"column"`
}


type Level struct{
	LevelName string
	Text [][]byte
	BufferImmutable bool
	isFinished (func() bool)
	LevelState [][]bool
	InitialLevelState [][]bool
	CursorCallback (func(Cursor))
	ProhibitedInputs []string
}


func NewNavigateLevel(name string, initalText [][]byte, bufferImmutable bool) Level{
	// Creates a new level whose win condition is to navgiate to all non-space characters
	// in the given initialText buffer

	// Construct LevelState to be false everywhere initially

	LevelState:= make([][]bool, len(initalText))
	for i := range LevelState{
		LevelState[i] = make([]bool, len(initalText[i]))
	}
	
	// Go through each position and check whether each non empty byte has been visited
	var isFinishedNavigate = func() bool{
		finished := true 
		stringBuffer := ConvertBytesToStrings(initalText)
		for i, row:= range(stringBuffer){
			for j, _ := range(row){
				// string buffer is a character we want to navigate to and they have been there
				// log.Print(stringBuffer[i][j] != " ")
				if(stringBuffer[i][j] != " "){
					finished = finished && LevelState[i][j]
				}
			}
		}

		return finished
	}

	// Given the new cursor position, set that positon in the LevelState array to true
	var CursorCallback = func(cursorPosition Cursor) {
		LevelState[cursorPosition.Row][cursorPosition.Column] = true
		// log.Print(LevelState)
	}

	var copyInitialLevelState = make([][]bool, len(LevelState))
	util.Copy2DArrayBool(copyInitialLevelState, LevelState)
	return Level{
		LevelName: name,
		Text: initalText,
		BufferImmutable: bufferImmutable,
		isFinished: isFinishedNavigate,
		LevelState: LevelState,
		InitialLevelState: copyInitialLevelState,
		CursorCallback: CursorCallback,
		ProhibitedInputs: []string{"W", "<S-W>", "W", "<S-B>"},
	}
}
func (lvl *Level) HasWonLevel() bool{
	return lvl.isFinished()
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

func (lvl *Level) resetLevel(){
	util.Copy2DArrayBool(lvl.LevelState, lvl.InitialLevelState)
}