package game

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"vim-zombies/Utilities"
	"github.com/neovim/go-client/nvim"
)

type KeyPress struct {
	Key string `json:"key" vim:"key"`
	// Open to add additional fields such as time stamp
}

var errorCursor Cursor = Cursor{
	-1, -1,
}

type Instance struct {
	Vim          *nvim.Nvim
	window       nvim.Window
	cursor       Cursor
	levels       []CompletableLevel
	currentLevel int
	Cleanup      func()
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
		Vim:          vim,
		window:       windows[0],
		cursor:       Cursor{}, // Cursor will be set itself
		currentLevel: 0,        // Start on the first level
		Cleanup:      cleanup,
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
	if slices.Contains(prohibtedInputs, key) || slices.Contains(prohibtedInputs, strings.ToUpper(key)){
		// Optionally return to user that this input is not allowed
		log.Print("Prohibited input received.")
	} else {
		vi.Vim.Input(key)
	}
}

func (vi *Instance) HandleKeyPress(writer http.ResponseWriter, request *http.Request) {

	var keypress KeyPress
	lvl := vi.GetCurrentLevel()
	// Decode keypress from json
	err := json.NewDecoder(request.Body).Decode(&keypress)
	key := keypress.Key

	if err != nil {
		log.Print("Error decoding message request body:")
	}

	log.Printf("Got keypress %s.", key)

	vi.makeKeyPressIfValid(key)

	err = vi.updateCursorPosition()

	if err != nil {
		// Can handle this case if needed, but for now this just means incomplete input sequence
		// Current position
		log.Print("Key press only forms a partial input sequence.")
	} else {
		// Update game state with
		lvl.CursorCallback(vi.cursor)
	}

	var responseBestTime int64 = 0

	finished := lvl.IsFinished()
	if finished {
		log.Print("Finished level")
		vi.ProgressLevel()
		responseBestTime = lvl.GetBestTime()
	}

	response := struct {
		Cursor     Cursor  `json:"cursor"`
		IsFinished bool    `json:"finished"`
		BestTime   float64 `json:"bestTime"`
	}{
		vi.cursor,
		finished,
		float64(float64(responseBestTime) / 1000.0),
	}
	err = json.NewEncoder(writer).Encode(response)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	} else {

		writer.WriteHeader(http.StatusOK)
	}

}

func (vi *Instance) initFromLevel() {
	vim := vi.Vim
	lvl := vi.GetCurrentLevel()

	buffer, _ := vi.Vim.CreateBuffer(false, false)

	lvl.FillTextBlanks()

	if err := vim.SetBufferLines(buffer, 0, -1, true, lvl.GetText()); err != nil {
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

func (vi *Instance) GetLevel(writer http.ResponseWriter, request *http.Request) {
	log.Print("Got request for current level.")
	var stringLevel [][]string = ConvertBytesToStrings(vi.GetCurrentLevel().GetText())
	response := struct {
		Level  [][]string `json:"level"`
		Cursor Cursor     `json:"cursor"`
	}{
		stringLevel,
		vi.cursor,
	}
	// When returning the level, also return the current cursor position
	err := json.NewEncoder(writer).Encode(response)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	} else {

		writer.WriteHeader(http.StatusOK)
	}

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

	return levels
}

func (vi *Instance) GetCurrentLevel() CompletableLevel{
	return vi.levels[vi.currentLevel]

}

// Game levels wrap around
func (vi *Instance) ProgressLevel() {
	// Reset just completed level so that it can be used again
	var lvl CompletableLevel = vi.GetCurrentLevel()
	lvl.finishLevel()

	vi.currentLevel = (vi.currentLevel + 1) % (len(vi.levels)) // Need to actually update the buffer once the level has been completed
	vi.initFromLevel()                                         // Actually populate the buffer with the new level.
}

func (vi *Instance) ResetLevel(writer http.ResponseWriter, request *http.Request) {
	log.Print("Got request to reset current level")
	var lvl CompletableLevel = vi.GetCurrentLevel()
	lvl.resetLevel()
	vi.initFromLevel()
	writer.WriteHeader(http.StatusOK)
}
