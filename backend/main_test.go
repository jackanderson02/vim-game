package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"vim-zombies/Auth"
	"vim-zombies/Game"
)

type DummyLevel struct {
	game.Level
}

func (lvl DummyLevel) IsFinished() bool {
	return true
}

func sendDummyKeyPress(t *testing.T, handler http.HandlerFunc, key string) *bytes.Buffer {
	// Create a ResponseRecorder to record the response.
	rr := httptest.NewRecorder()

	requestBody, err := json.Marshal(map[string]interface{}{
		"auth_key": "imbecile",
		"key": key,
	})

	if err != nil {
		t.Fatal("Failed to marshal JSON data.")
	}
	req := httptest.NewRequest(http.MethodPost, "/keyPress", bytes.NewBuffer(requestBody))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Keypress endpoint responded with %d", status)
	}

	return rr.Body
}

func TestLevelProgression(t *testing.T) {

	// Allows us to skip a level

	var dummyLevelTime *game.LevelTime = &game.LevelTime{
		0,0,
	}


	makeNewInstance := func () game.Instance {
		return game.NewInstanceWithLevels([]game.CompletableLevel{
		&DummyLevel{game.Level{LevelName: "dummy1", LevelTime: dummyLevelTime}},
		&DummyLevel{game.Level{LevelName: "dummy2", LevelTime: dummyLevelTime}},
		})
	}

	// Question of how can we force an arbitary level to progress, how can we cheat?
	// dummyInstance = &newInstance
	auth := auth.NewAuthenticatedUsersMutexWithInstanceFunc(makeNewInstance)

	// Send key press to immediately skip over level
	keyPressHandler := http.HandlerFunc(auth.HandleKeyPressWrapper)
	resp := sendDummyKeyPress(t, keyPressHandler, "l")

	// Get the level after the keypress
	var keyPressResponse map[string]interface{}
	json.NewDecoder(resp).Decode(&keyPressResponse)
	// req L= httptest.NewRequest(http.MethodPost, "/level", bytes.NewBuffer(requestBody))
	// handler.ServeHTTP(rr, req)

	log.Print(keyPressResponse)

	if finished, ok := (keyPressResponse["finished"]).(bool); ok && !finished{
		t.Error("Response to key press did not indicate that the previous level has been completed")
	}else if !ok{
		t.Error("Response to key press did not return a valid boolean flag called finished to indicate level completion.")
	}
}

func TestGetLevelHandler(t *testing.T) {
	rr := httptest.NewRecorder()

	// Call the handler, passing in the ResponseRecorder and request.
	auth := auth.NewAuthenticatedUsersMutex()
	handler := http.HandlerFunc(auth.GetLevelWrapper)
	// Create dummy auth key
	requestBody, err := json.Marshal(map[string]interface{}{
		"auth_key": "idiot",
	})
	if err != nil {
		t.Fatal("Failed to marshal JSON data.")
	}
	req := httptest.NewRequest(http.MethodPost, "/level", bytes.NewBuffer(requestBody))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Level endpoint responded with %d", status)
	}

}


func TestInputChangesCursorPosition(t *testing.T) {
	// Send keypress j to game instance
	// Call the handler, passing in the ResponseRecorder and request.

	auth := auth.NewAuthenticatedUsersMutex()
	handler := http.HandlerFunc(auth.HandleKeyPressWrapper)

	initialGameState := sendDummyKeyPress(t, handler, "").String()
	gameStateAfterKeypress := sendDummyKeyPress(t, handler, "l").String()

	if initialGameState == gameStateAfterKeypress {
		t.Error("Game state did not change following the key press l.")
	}
}

func TestAuthKey(t *testing.T){
	// Test simply asserts that the auth key is extracted and the server does not respond
	// with an error
	auth := auth.NewAuthenticatedUsersMutex()
	handler := http.HandlerFunc(auth.GetLevelWrapper)
	rr := httptest.NewRecorder()

	authReq := struct{
		Unique_id string `json:"auth_key"`
	}{
		"test",
	}
	requestBody, err := json.Marshal(authReq)

	if err != nil {
		t.Fatal("Failed to marshal JSON data.")
	}

	req := httptest.NewRequest(http.MethodPost, "/level", bytes.NewBuffer(requestBody))

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK{
		t.Error("Got " + rr.Result().Status + " from server but did not expect error.")
	}
}