package game

import(
	"vim-zombies/Utilities"
)


func (navLvl *NavigateLevel) CursorCallback(cursorPosition Cursor){
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
	return []string{}
	// return []string{"W", "<S-W>", "B", "<S-B>"} 
}

func NewNavigateLevel(name string, initalText [][]byte, bufferImmutable bool) NavigateLevel{
	// Creates a new level whose win condition is to navgiate to all non-space characters
	// in the given initialText buffer

	// Construct LevelState to be false everywhere initially
	LevelState := FalseLvlStateFromText(initalText)
	
	var copyInitialLevelState = make([][]bool, len(LevelState))
	util.Copy2DArray(copyInitialLevelState, LevelState)
	return NavigateLevel{Level{
		LevelName: name,
		text: initalText,
		bufferImmutable: bufferImmutable,
		levelState: LevelState,
		initialLevelState: copyInitialLevelState,
	}}
}
