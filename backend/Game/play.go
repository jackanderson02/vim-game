package game

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	util "vim-zombies/Utilities"

	"github.com/neovim/go-client/nvim"
)

type ShouldReloadResponse struct {
	ShouldReload bool   `json:"shouldReload"`
	ReloadText   string `json:"reloadText"`
}

func NewDoNotReloadResponse() ShouldReloadResponse {
	return ShouldReloadResponse{
		ShouldReload: false,
	}
}

func NewReloadResponse(reloadText string) ShouldReloadResponse {
	return ShouldReloadResponse{
		ShouldReload: true,
		ReloadText:   reloadText,
	}
}

var errorCursor Cursor = Cursor{
	-1, -1,
}

type Instance struct {
	Vim              *nvim.Nvim
	window           nvim.Window
	cursor           Cursor
	levels           []CompletableLevel
	InstanceResponse map[string]interface{}
	InstanceRequest  map[string]interface{}
	currentLevel     int
	Cleanup          func()
}

func baselineInstance() Instance {
	log.Println("Initializing Neovim...")
	vim, cleanup, err := util.ConnectWhenReady()
	if err != nil {
		log.Fatalf("Failed to initialize Neovim: %v", err)
	}

	windows, err := vim.Windows()
	if err != nil {
		log.Fatalf("Failed to get windows %v", err)
	}
	if len(windows) == 0 {
		log.Fatal("No windows found.")
	}

	vi := Instance{
		Vim:              vim,
		window:           windows[0],
		currentLevel:     0, // Start on the first level
		Cleanup:          cleanup,
		InstanceResponse: make(map[string]interface{}),
	}

	return vi

}

func NewInstance() Instance {
	vi := baselineInstance()
	vi.levels = initLevels()
	vi.initFromLevel()
	return vi
}
func NewInstanceWithLevels(levels []CompletableLevel) Instance {
	vi := baselineInstance()
	vi.levels = levels
	return vi
}

func NewInstanceWithoutLevels() Instance {
	vi := baselineInstance()
	return vi
}

func (vi *Instance) SetLevels(levels []CompletableLevel) {
	vi.levels = levels
}

func (vi *Instance) updateCursorPosition() error {
	vim := vi.Vim
	mode, _ := vim.Mode()
	if mode.Blocking {
		return fmt.Errorf("nvim is currently blocked because it is waiting additional user input.")
	}

	cursor, err := vim.WindowCursor(vi.window)
	if err != nil {
		return fmt.Errorf("failed to get cursor position: %v", err)
	}

	// IMPORTANT: convert 1 indexed rows to 0 indexed for compatibility with almost
	// every high level language I know of, apart from MATLAB :skull:
	vi.cursor = Cursor{
		cursor[0] - 1, cursor[1],
	}
	log.Printf("Settings cursor to %d, %d", cursor[0]-1, cursor[1])

	return nil
}

func getLastErrorMessage(vim *nvim.Nvim) (string, error) {
	var errmsg string
	if err := vim.Eval("v:errmsg", &errmsg); err != nil {
		return "", fmt.Errorf("failed to get v:errmsg: %v", err)
	}
	return errmsg, nil
}

func (vi *Instance) makeKeyPressIfValid(key string) {
	prohibtedInputs := vi.GetCurrentLevel().GetProhibtedInputs()
	if slices.Contains(prohibtedInputs, key) || slices.Contains(prohibtedInputs, strings.ToUpper(key)) {
		// Optionally return to user that this input is not allowed
		log.Print("Prohibited input received.")
	} else {
		vi.Vim.Input(key)
	}
}

func (vi *Instance) HandleKeyPress() {

	// var keypress KeyPress
	lvl := vi.GetCurrentLevel()
	// Decode keypress from json
	rawKey, hasKey := vi.InstanceRequest["key"]
	if !hasKey {
		log.Print("No keypress was provided in previous request.")
	}
	key, isString := rawKey.(string)
	if !isString {
		log.Fatalf("Could not convert given keypress to a string.")
	}
	log.Printf("Got keypress %s.", key)

	vi.makeKeyPressIfValid(key)

	err := vi.updateCursorPosition()

	if err != nil {
		// Can handle this case if needed, but for now this just means incomplete input sequence
		// Current position
		log.Print("Key press only forms a partial input sequence.")
	} else {
		// Update game state with
		lvl.CursorCallback(vi.cursor)
	}

	var responseBestTime int64 = 0

	gameStatus := lvl.UpdateLevelState()
	finished := gameStatus == FINISHED
	levelOver := gameStatus == OVER
	vi.InstanceResponse["cursor"] = vi.cursor
	if finished {
		log.Print("Finished level")
		vi.ProgressLevel()
		responseBestTime = lvl.GetBestTime()
	} else if levelOver {
		vi.ResetLevel()
		vi.InstanceResponse["shouldReloadResponse"] = NewReloadResponse("Level failed... resetting level")
	}

	log.Print(levelOver)
	log.Printf("Best time %d", responseBestTime)

	// Update the response map
	vi.InstanceResponse["finished"] = finished
	vi.InstanceResponse["bestTime"] = float64(float64(responseBestTime) / 1000.0)

}

func (vi *Instance) WriteInstanceResponseToWriter(writer http.ResponseWriter) {
	log.Print("Writing instance response\n")
	log.Println(vi.InstanceResponse)
	json.NewEncoder(writer).Encode(vi.InstanceResponse)
	vi.ClearResponseRequest()
}

func (vi *Instance) initFromLevel() {
	vim := vi.Vim
	lvl := vi.GetCurrentLevel()

	buffer, _ := vi.Vim.CreateBuffer(false, false)

	lvl.FillTextBlanks()

	if err := vim.SetBufferLines(buffer, 0, -1, true, lvl.GetLevelText()); err != nil {
		log.Fatalf("Failed to set buffer lines: %v", err)
	}

	if err := vim.SetCurrentBuffer(buffer); err != nil {
		log.Fatalf("Failed to set current buffer: %v", err)
	}
	if lvl.IsBufferImmutable() {
		vi.Vim.SetBufferOption(buffer, "modifiable", false)
	}
	// Cursor position will reset when new buffer loaded
	vi.updateCursorPosition()

	// Start the timer for this level
	lvl.startLevel()

}
func (vi *Instance) ClearResponseRequest() {
	vi.InstanceResponse = make(map[string]interface{})
	vi.InstanceRequest = make(map[string]interface{})
}

func (vi *Instance) GetLevel() {
	log.Print("Got request for current level.")
	// var stringLevel [][]string = ConvertBytesToStrings(vi.GetCurrentLevel().GetLevelText())
	// When returning the level, also return the current cursor position
	// vi.InstanceResponse["level"] = stringLevel
	// vi.InstanceResponse["cursor"] = vi.cursor
	currentLevelInformation:= vi.GetCurrentLevel().GetLevelInformation()

	for key, value := range currentLevelInformation{
		vi.InstanceResponse[key] = value
	}
	vi.InstanceResponse["cursor"] = vi.cursor

}
func initLevels() []CompletableLevel {

	var levels []CompletableLevel

	// Populate hard coded levels

	level1 := NewNavigateLevel("Level 1", [][]byte{
		{' ', '{', ' ', ' ', '(', ' ', ' ', ')', ' ', ' '},  // Curly braces and parentheses with more gaps
		{' ', ' ', 'f', ' ', ' ', 'i', ' ', ' ', 'v', ' '},  // Go keywords (func, if, var) with gaps
		{' ', 'r', ' ', ' ', '=', ' ', ' ', 't', ' ', ' '},  // Assignment and boolean operators with gaps
		{' ', ' ', 'p', ' ', ' ', 'n', ' ', ' ', '[', ' '},  // Print, newline, and array symbols with gaps
		{' ', '&', ' ', ' ', '*', ' ', ' ', '/', ' ', '\\'}, // Pointers and operators with more gaps
		{' ', ' ', '<', ' ', ' ', '>', ' ', ' ', ';', ' '},  // Comparison operators and semicolon
		{' ', '!', ' ', ' ', '&', ' ', ' ', '|', ' ', ' '},  // Logical operators
		{' ', ' ', ':', ' ', ' ', ',', ' ', ' ', '_', ' '},  // Colon, comma, and underscore
		{' ', '%', ' ', ' ', '[', ' ', ' ', ']', ' ', ' '},
	}, true)

	levels = append(levels, &level1)

	level2 := NewNavigateLevel("Level 2", [][]byte{
		{' ', '{', ' ', '(', ' ', ')', ' ', '}', ' ', '['},   // Curly braces, parentheses, brackets
		{'+', ' ', '-', ' ', '*', ' ', '/', ' ', '%', ' '},   // Arithmetic operators
		{' ', '<', ' ', '>', ' ', '=', ' ', '!', ' ', '&'},   // Comparison and logical operators
		{'$', ' ', '@', ' ', '#', ' ', '^', ' ', '_', ' '},   // Variable identifiers, bitwise operators
		{'?', ' ', ':', ' ', '.', ' ', ',', ' ', ';', ' '},   // Conditional, punctuation
		{' ', '|', ' ', '\\', ' ', '`', ' ', '~', ' ', '\''}, // Bitwise OR, escape, quotes
		{'=', ' ', '=', ' ', '!', ' ', '<', ' ', '>', ' '},   // Equality and relational operators
		{' ', '&', ' ', '|', ' ', '>', ' ', '>', ' ', '/'},   // Logical operators, lambda, comments
		{'"', ' ', '<', ' ', '>', ' ', '*', ' ', '&', ' '},   // Shift operators, exponentiation, logical AND
	}, true)

	levels = append(levels, &level2)

	level3 := NewStaticAvoidanceLevel("Level 3", [][]byte{
		{' ', '{', ' ', 'X', ' ', 'X', ' ', '}', 'X', '['},   // Curly braces, parentheses, brackets
		{'+', ' ', '-', 'X', '*', ' ', '/', ' ', 'X', ' '},   // Arithmetic operators
		{' ', '<', ' ', '>', 'X', 'X', ' ', '!', ' ', '&'},   // Comparison and logical operators
		{'$', ' ', '@', ' ', '#', ' ', '^', 'X', '_', ' '},   // Variable identifiers, bitwise operators
		{'?', ' ', 'X', ' ', '.', ' ', ',', ' ', ';', ' '},   // Conditional, punctuation
		{' ', '|', 'X', '\\', ' ', '`', ' ', '~', ' ', '\''}, // Bitwise OR, escape, quotes
		{'=', ' ', '=', ' ', 'X', ' ', '<', 'X', '>', ' '},   // Equality and relational operators
		{'X', '&', ' ', '|', ' ', 'X', ' ', '>', ' ', '/'},   // Logical operators, lambda, comments
		{'"', ' ', '<', ' ', '>', 'X', '*', ' ', 'X', ' '},   // Shift operators, exponentiation, logical AND
	}, true, []string{"X"})

	levels = append(levels, &level3)
	return levels
}

func (vi *Instance) GetCurrentLevel() CompletableLevel {
	return vi.levels[vi.currentLevel]
}



func (vi *Instance) ProgressLevel() {
	vi.GetCurrentLevel().finishLevel()                         // finish the current level to update the times
	vi.currentLevel = (vi.currentLevel + 1) % (len(vi.levels)) // Update to point to the next level; game levels wrap around
	vi.initFromLevel()
}

func (vi *Instance) ResetLevel() {
	log.Print("Got request to reset current level")
	var lvl CompletableLevel = vi.GetCurrentLevel()
	lvl.resetLevel()
	vi.initFromLevel()
}
