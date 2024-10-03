package game


import(
	"fmt"
	"slices"
	"strings"
	"log"
	"vim-zombies/Utilities"
	"encoding/json"
	"net/http"
	"github.com/neovim/go-client/nvim"
)

type KeyPress struct{
	Key string `json:"key" vim:"key"`
	// Open to add additional fields such as time stamp
}
	
var errorCursor Cursor = Cursor{
	-1, -1,
}

type Instance struct {
	Vim *nvim.Nvim
	window nvim.Window
	cursor Cursor
	levels       []Level
	currentLevel int
	Cleanup func()
}

func NewInstance() Instance{
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
		vim,
		windows[0],
		Cursor{}, // Cursor will be set itself
		initLevels(),
		0, // Start on the first level
		cleanup,
	};

	vi.initFromLevel()

	return vi

}
// Function to get the cursor position
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
		cursor[0] -1 , cursor[1],
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

func (vi *Instance) HandleKeyPress(writer http.ResponseWriter, request *http.Request) {

	var keypress KeyPress;
	lvl := vi.GetCurrentLevel()
	// Decode keypress from json
	err := json.NewDecoder(request.Body).Decode(&keypress)
	key := keypress.Key

	if err != nil {
		log.Print("Error decoding message request body:")
		return
	}

	log.Printf("Got keypress %s.", key)

	// Make the keypress only if key is not in probihited inputs
	if slices.Contains(vi.GetCurrentLevel().ProhibitedInputs, key) || 
	slices.Contains(vi.GetCurrentLevel().ProhibitedInputs, strings.ToUpper(key)){
		// Optionally return to user that this input is not allowed
		log.Print("Prohibited input received.")
	} else{
		vi.Vim.Input(key)
	}

	err = vi.updateCursorPosition()

	if err != nil{
		// Can handle this case if needed, but for now this just means incomplete input sequence
		// Current position
		log.Print("Key press only forms a partial input sequence.")
	} else {
		// Update game state with 
		lvl.CursorCallback(vi.cursor)
	}

	finished := lvl.HasWonLevel()
	if(finished){
		log.Print("Finished level")
		vi.ProgressLevel()
	}

	response := struct{
		Cursor Cursor `json:"cursor"`
		IsFinished bool `json:"finished"`
	}{
		vi.cursor,
		finished,
	}

	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(response)

}

func (vi *Instance) initFromLevel() {
	vim := vi.Vim
	lvl := vi.GetCurrentLevel()

	buffer, _ := vi.Vim.CreateBuffer(false, false)

	// Fill level horizontally
	// go over each row and find max length
	txt := lvl.Text
	var max_len int = 0
	for _,v := range(txt){
		if max_len < len(v){
			max_len = len(v)
		}
	}
	// Then Fill in blanks
	txt_copy := make([][]byte, len(txt))
	// copy(txt_copy, txt)
	for i, v := range(txt){
		txt_copy[i] = make([]byte, max_len)
		copy(txt_copy[i], v)
		// Need to then append max_len - len(v) empty slots
		for j:= 0; j <(max_len-len(v)); j++{
			txt_copy[i][len(v) + j -1] = byte(' ')
		}
	}
	copy(lvl.Text, txt_copy)
	lvl.Text = txt_copy

	if err := vim.SetBufferLines(buffer, 0, -1, true, lvl.Text); err != nil {
		log.Fatalf("Failed to set buffer lines: %v", err)
	}

	if err := vim.SetCurrentBuffer(buffer); err != nil {
		log.Fatalf("Failed to set current buffer: %v", err)
	}
	if lvl.BufferImmutable{
		vi.Vim.SetBufferOption(buffer, "modifiable", false)
	}
	// Cursor position will reset when new buffer loaded
	vi.updateCursorPosition()
	
}

func (vi *Instance) GetLevel (writer http.ResponseWriter, request *http.Request) {
	log.Print("Got request for current level.")
	var stringLevel [][]string = ConvertBytesToStrings(vi.GetCurrentLevel().Text)
	response := struct{
		Level [][]string `json:"level"`
		Cursor Cursor `json:"cursor"`
	}{
		stringLevel,
		vi.cursor,
	}
	// When returning the level, also return the current cursor position
	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(response)
	// data, _ := json.Marshal(&stringLevel)
	// writer.Write(data)
}


func initLevels() []Level{
	var levels []Level

	// Populate hard coded levels

	// First level is navigation level
	levels = append(levels, NewNavigateLevel("Level 1", [][]byte{
		{' ', '{', ' ', ' ', '(', ' ', ' ', ')', ' ', ' '},  // Curly braces and parentheses with more gaps
		{' ', ' ', 'f', ' ', ' ', 'i', ' ', ' ', 'v', ' '},  // Go keywords (func, if, var) with gaps
		{' ', 'r', ' ', ' ', '=', ' ', ' ', 't', ' ', ' '},  // Assignment and boolean operators with gaps
		{' ', ' ', 'p', ' ', ' ', 'n', ' ', ' ', '[', ' '},  // Print, newline, and array symbols with gaps
		{' ', '&', ' ', ' ', '*', ' ', ' ', '/', ' ', '\\'}, // Pointers and operators with more gaps
		{' ', ' ', '<', ' ', ' ', '>', ' ', ' ', ';', ' '},  // Comparison operators and semicolon
		{' ', '!', ' ', ' ', '&', ' ', ' ', '|', ' ', ' '},  // Logical operators
		{' ', ' ', ':', ' ', ' ', ',', ' ', ' ', '_', ' '},  // Colon, comma, and underscore
		{' ', '%', ' ', ' ', '[', ' ', ' ', ']', ' ', ' '},
	}, true),
	)

	levels = append(levels, NewNavigateLevel("Level 2", [][]byte{
		{' ', '{', ' ', '(', ' ', ')', ' ', '}', ' ', '['},   // Curly braces, parentheses, brackets
		{'+', ' ', '-', ' ', '*', ' ', '/', ' ', '%', ' '},   // Arithmetic operators
		{' ', '<', ' ', '>', ' ', '=', ' ', '!', ' ', '&'},   // Comparison and logical operators
		{'$', ' ', '@', ' ', '#', ' ', '^', ' ', '_', ' '},   // Variable identifiers, bitwise operators
		{'?', ' ', ':', ' ', '.', ' ', ',', ' ', ';', ' '},   // Conditional, punctuation
		{' ', '|', ' ', '\\', ' ', '`', ' ', '~', ' ', '\''}, // Bitwise OR, escape, quotes
		{'=', ' ', '=', ' ', '!', ' ', '<', ' ', '>', ' '},   // Equality and relational operators
		{' ', '&', ' ', '|', ' ', '>', ' ', '>', ' ', '/'},   // Logical operators, lambda, comments
		{'"', ' ', '<', ' ', '>', ' ', '*', ' ', '&', ' '},   // Shift operators, exponentiation, logical AND
	}, true),)

	return levels
}

func (vi *Instance) GetCurrentLevel() Level {
	return vi.levels[vi.currentLevel]
}

// Game levels wrap around
func (vi *Instance) ProgressLevel() {
	// Reset just completed level so that it can be used again
	var lvl Level = vi.GetCurrentLevel()
	var plevel *Level = &lvl
	plevel.resetLevel()
	vi.currentLevel = (vi.currentLevel + 1)%(len(vi.levels)) // Need to actually update the buffer once the level has been completed
	vi.initFromLevel() // Actually populate the buffer with the new level.
}