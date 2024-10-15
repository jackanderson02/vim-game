package auth 

import (
	"net/http"
	"sync"
	"vim-zombies/Game"
	"bytes"
	"io"
	"encoding/json"
	"errors"
	"log"
	"time"

)

const TIMEOUT_AFTER_S = 120 
type Session struct {
	vi *game.Instance
	lastInteraction time.Time
}

type AuthenticatedUsersMutex struct{
	authenticatedUsers map[string]Session
	sync.Mutex
}

var usersMapMutex AuthenticatedUsersMutex;
func init(){
	usersMapMutex = AuthenticatedUsersMutex{
		authenticatedUsers: make(map[string]Session),
	}
	go checkForExpiredSessions()
}

func checkForExpiredSessions(){
	for{
		usersMapMutex.Lock()
		// Check all concurrent sessions to see if any have expired
		for id, session := range(usersMapMutex.authenticatedUsers){
			if time.Now().After(session.lastInteraction.Add(time.Duration(time.Second * TIMEOUT_AFTER_S ))){
				// delete this session
				log.Printf("Removing session with id %s", id)
				session.vi.Cleanup()
				delete(usersMapMutex.authenticatedUsers, id)
			}
		}
		usersMapMutex.Unlock()
		time.Sleep(time.Second)
	}

}
func GetLevelWrapper(writer http.ResponseWriter, request *http.Request){
	session, err:= AuthenticateUser(&HTTPWrapper{writer, request})
	vi := session.vi
	if err != nil{
		log.Print(err.Error())
	}else{
		vi.GetLevel(writer, request)
	}
}
func ResetLevelWrapper(writer http.ResponseWriter, request *http.Request){
	session, err:= AuthenticateUser(&HTTPWrapper{writer, request})
	vi := session.vi
	if err != nil{
		log.Print(err.Error())
	}else{
		vi.ResetLevel(writer, request)
	}

}
func HandleKeyPressWrapper(writer http.ResponseWriter, request *http.Request){
	session, err:= AuthenticateUser(&HTTPWrapper{writer, request})
	vi := session.vi
	if err != nil{
		log.Print(err.Error())
	}else{
		vi.HandleKeyPress(writer, request)
	}
}

type HTTPWrapper struct{
	writer http.ResponseWriter
	request *http.Request
}
func AuthenticateUser(httpWrapper *HTTPWrapper) (Session, error){

	// TODO handle cookie expiration

	req := struct{
		Unique_id string `json:"auth_key"`
	}{}

	request := httpWrapper.request


	bodyBytes, _:= io.ReadAll(request.Body)
	request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))


	err := json.NewDecoder(request.Body).Decode(&req)
	
	if err != nil{
		return Session{},errors.New("Failed to decode auth_token from JSON.")
	}

	log.Printf("auth_key: %s",req.Unique_id) 


	// Enter concurrency section
	usersMapMutex.Lock()
	_, ok:= usersMapMutex.authenticatedUsers[req.Unique_id]

	if !ok{
		// Add user to authenticated users
		newInstance := game.NewInstance();
		newSession := Session{
			vi: &newInstance,
			lastInteraction: time.Now(),
		}
		usersMapMutex.authenticatedUsers[req.Unique_id] = newSession
	}

	userSession := usersMapMutex.authenticatedUsers[req.Unique_id]
	request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	
	usersMapMutex.Unlock()
	// Exit concurrency section

	return userSession, nil

}

func DoAllCleanups(){
	// TODO, cleanup all game instances
	usersMapMutex.Lock()
	for _, session:= range usersMapMutex.authenticatedUsers{
		session.vi.Cleanup();
	}
	usersMapMutex.Unlock()
}