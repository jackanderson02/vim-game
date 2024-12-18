import CharacterCell from "./CharacterCell";
import { Arrow90degRight } from "react-bootstrap-icons";
import "../css/grid.css"
import { useContext, useEffect, useState } from "react";
import { Button } from "react-bootstrap";
import UserContext from "../context/UserContext";


const MAX_KEYS_DISPLAYED = 15

const CharacterGrid = ({ gridData, cellColours, fetchLevel}) => {
  console.log(cellColours)

  const fingerprint = useContext(UserContext)

  const fetchLevelAndCursor = async (loadLevelAfter) => {
    const cursor = await (fetchLevel(loadLevelAfter))
    console.log("fetching level and cursor")
    setCursor(cursor)
    // setCursor(cursor)
  }

  const requestResetLevel = async () => {
      try{
        const response = await fetch('http://localhost:8080/resetLevel', {
          method: 'POST',
          headers: {'Content-type': 'application/json'},
          body: JSON.stringify({auth_key: fingerprint})
        });
        if (!response.ok){
          throw new Error("Failed to fetch data.")
        }
        await fetchLevelAndCursor()
        

      } catch (error){
        console.error("Error resetting level", error)
      }
    }


  const [keysPressed, setKeysPressed] = useState("")
  const [cursor, setCursor] = useState({Row:0, Column: 0})
  const [bestTime, setBestTime] = useState(0)

  const [displayTextToUser, setDisplayTextToUser] = useState("")

  const vimifiedMappings = {
    'Escape': "<ESC>",
    'Backspace': "<BS>",
    ' ': '<Space>',
    'Tab': "<Tab>",
    // Can add more key mappings later
  }

  // Avoid Ctrl and Shift being sent additionally as separate key presses
  const avoidRepeated = ['Control', 'Shift']

  const sleep  = function sleep(milliseconds) {
    const start = Date.now();
    while (Date.now() - start < milliseconds) {
        // Busy-wait loop
    }
  }

  const vimifyKeyEvent = (event) => {
    console.log(event)

    // Check for special characters here, can choose what we want to support for now
    const {key, code, metaKey, shiftKey, altKey, ctrlKey} = event;
    let vimified 

    if(ctrlKey){
      console.log("Ctrl")
      vimified = `<C-${key}>`
    }else if(shiftKey){
      console.log("Shift")
      vimified = `<S-${key}>`
    }else if(altKey){
      console.log("Alt")
      vimified = `<A-${key}`
    }

    if(vimified){
      return vimified
    }

    const keyCodeStr = code + ""
    if(keyCodeStr.includes("Key") || keyCodeStr.includes("Digit")){
      return key
    }

    // Otherwise, look up in key mappings
    if (key in vimifiedMappings){
      return vimifiedMappings[key]
    }

    if (avoidRepeated.includes(key)){
      console.log("Control key")
    }else{
      // Return raw key
      return key
    }


  }

  const handleKeyDown = async (event) => {

    let vimKeyEvent = vimifyKeyEvent(event)
    // Update keys on screen
    if(vimKeyEvent){
      setKeysPressed((keysPressed+vimKeyEvent+" ").slice(-MAX_KEYS_DISPLAYED))
      // Send the key bind to neovim
      try{

        const response = await fetch('http://localhost:8080/keyPress', {
          method: 'POST',
          body: JSON.stringify({key: vimKeyEvent, auth_key: fingerprint})
        });
        if (!response.ok){
          throw new Error("Failed to fetch data.")
        }

        const responseJSON= await response.json()
        console.log(responseJSON)

        const responseCursor = responseJSON.cursor
        const finished = responseJSON.finished
        const responseBestTime = responseJSON.bestTime
        const shouldReloadResponse = responseJSON.shouldReloadResponse
        const shouldReload = shouldReloadResponse.shouldReload

        console.log("Response cursor" + responseCursor.Row + "," + responseCursor.Column)
        setCursor(responseCursor)
        setDisplayTextToUser("")

        if(finished){
          setBestTime(responseBestTime)
          setTimeout(() => {fetchLevelAndCursor()}, 500)
          // setBestTime(0)
        }else if(shouldReloadResponse && shouldReload){
          // Display the shouldReloadResponse text
          setDisplayTextToUser(shouldReloadResponse.reloadText)
          setTimeout(() => {fetchLevelAndCursor()}, 250)
        }
        
      } catch (error){
        console.error("Error fetching levels", error)
      }
    }

  }

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);

    return () => {
      window.removeEventListener('keydown', handleKeyDown)
    }
  })

  return (
    <>
    {bestTime ? (<h1>Current best time: {bestTime}</h1>) : <div></div>}
    {/* // {bestTime && (<h1>{bestTime}</h1>)} */}
    <div className="grid-container">
        {gridData ? gridData.map((row, rowIndex) => (
          <div key={rowIndex} className="grid-row">
            {row.map((char, colIndex) => (
              <CharacterCell
                key={colIndex}
                char={char}
                cellColour={cellColours? cellColours[rowIndex][colIndex]: null}
                isCursor={rowIndex === cursor.Row && colIndex === cursor.Column}
              />
            ))}
          </div>
        )): <div></div>}
      <h2>{keysPressed}</h2>
      {gridData ? (<Button variant="warning" style={{position: "relative", float:"right"}} onClick={() => requestResetLevel()}><h3> Reset level</h3></Button>): <div></div>}
      {displayTextToUser? (<h3 style={{position: "relative", float:"left"}}> {displayTextToUser} </h3>): <div></div>}
      
      </div>
    </>
  );
};

export default CharacterGrid