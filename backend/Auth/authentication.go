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
	lastInteractionUnixMS int64
}

type AuthenticatedUsersMutex struct{
	authenticatedUsers map[string]*Session
	sync.Mutex
}

var usersMapMutex AuthenticatedUsersMutex;
func init(){
	usersMapMutex = AuthenticatedUsersMutex{
		authenticatedUsers: make(map[string]*Session),
	}
	go checkForExpiredSessions()
}

func checkForExpiredSessions(){
	for{
		usersMapMutex.Lock()
		// Check all concurrent sessions to see if any have expired
		for id, session := range(usersMapMutex.authenticatedUsers){
			if time.Now().UnixMilli() >= session.lastInteractionUnixMS + 1000*TIMEOUT_AFTER_S{
				// delete this session
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
	vi.WriteInstanceResponseToWriter(writer)
}
func ResetLevelWrapper(writer http.ResponseWriter, request *http.Request){
	session, err:= AuthenticateUser(&HTTPWrapper{writer, request})
	vi := session.vi
	if err != nil{
		log.Print(err.Error())
	}else{
		vi.ResetLevel(writer, request)
	}
	vi.WriteInstanceResponseToWriter(writer)
}
func HandleKeyPressWrapper(writer http.ResponseWriter, request *http.Request){
	session, err:= AuthenticateUser(&HTTPWrapper{writer, request})
	vi := session.vi
	if err != nil{
		log.Print(err.Error())
	}else{
		vi.HandleKeyPress(request)
	}
	vi.WriteInstanceResponseToWriter(writer)
}

type HTTPWrapper struct{
	writer http.ResponseWriter
	request *http.Request
}
func AuthenticateUser(httpWrapper *HTTPWrapper) (*Session, error){

	req := struct{
		Unique_id string `json:"auth_key"`
	}{}

	request := httpWrapper.request

	bodyBytes, _:= io.ReadAll(request.Body)
	request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	err := json.NewDecoder(request.Body).Decode(&req)
	
	if err != nil{
		return &Session{},errors.New("Failed to decode auth_token from JSON.")
	}



	// Enter concurrency section
	usersMapMutex.Lock()
	_, ok:= usersMapMutex.authenticatedUsers[req.Unique_id]

	if !ok{
		// Add user to authenticated users
		newInstance := game.NewInstance();
		newSession := Session{
			vi: &newInstance,
		}
		usersMapMutex.authenticatedUsers[req.Unique_id] = &newSession
	}

	userSession := usersMapMutex.authenticatedUsers[req.Unique_id]
	usersMapMutex.Unlock()
	// Exit concurrency section

	userSession.lastInteractionUnixMS = time.Now().UnixMilli();
	userSession.vi.InstanceResponse["shouldReload"] = !ok
	// Reload iff new user was just created, this forces a frontend reload which involves
	// fetching the level and cursor again in the event that the connection was timed out.

	request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	

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