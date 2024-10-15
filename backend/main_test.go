package main

import (
	"bytes"
	"encoding/json"
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

	requestBody, err := json.Marshal(game.KeyPress{Key: key})
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
	var dummyInstance *game.Instance

	var dummyLevelTime game.LevelTime = game.LevelTime{
		0,0,
	}

	newInstance := game.NewInstanceWithLevels([]game.CompletableLevel{
		&DummyLevel{game.Level{LevelName: "dummy1", LevelTime: dummyLevelTime}},
		&DummyLevel{game.Level{LevelName: "dummy2", LevelTime: dummyLevelTime}},
	})

	dummyInstance = &newInstance
	handler := http.HandlerFunc(dummyInstance.HandleKeyPress)

	initialLevel := dummyInstance.GetCurrentLevel()

	// Send key press to immediately skip over level
	rr := sendDummyKeyPress(t, handler, "l")

	levelAfterKeyPress := dummyInstance.GetCurrentLevel()
	// Then assert that the level has changed
	if initialLevel == levelAfterKeyPress{ 
		t.Error("Level did not progress after level was completed")
	}

	finishedResponse := struct{
		IsFinished bool `json:"finished"`
	}{}

	json.NewDecoder(rr).Decode(&finishedResponse)

	if (!finishedResponse.IsFinished){
		t.Error("JSON response did not indicate that the level had finished.")
	}
}

func TestGetLevelHandler(t *testing.T) {
	rr := httptest.NewRecorder()
	gameInstance := game.NewInstance()

	// Call the handler, passing in the ResponseRecorder and request.
	handler := http.HandlerFunc(gameInstance.HandleKeyPress)
	req := httptest.NewRequest(http.MethodPost, "/level", nil)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Level endpoint responded with %d", status)
	}

}

func TestInputChangesCursorPosition(t *testing.T) {
	// Send keypress j to game instance
	gameInstance := game.NewInstance()
	// Call the handler, passing in the ResponseRecorder and request.
	handler := http.HandlerFunc(gameInstance.HandleKeyPress)

	initialGameState := sendDummyKeyPress(t, handler, "").String()
	gameStateAfterKeypress := sendDummyKeyPress(t, handler, "l").String()

	if initialGameState == gameStateAfterKeypress {
		t.Error("Game state did not change following the key press l.")
	}
}

func TestAuthKey(t *testing.T){
	// Test simply asserts that the auth key is extracted and the server does not respond
	// with an error
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