import CharacterCell from "./CharacterCell";
import { Arrow90degRight } from "react-bootstrap-icons";
import "../css/grid.css"
import { useEffect, useState } from "react";


const MAX_KEYS_DISPLAYED = 15

const CharacterGrid = ({ gridData, fetchLevel}) => {

  const fetchLevelAndCursor = () => {
    setCursor(fetchLevel())
  }

  const [keysPressed, setKeysPressed] = useState("")
  const [cursor, setCursor] = useState({Row: 0, Column:0})
  // const [finished, setFinished] = useState(false)



  const vimifiedMappings = {
    'Escape': "<ESC>",
    'Backspace': "<BS>",
    ' ': '<Space>',
    'Tab': "<Tab>",
    // Can add more key mappings later
  }

  // Avoid Ctrl and Shift being sent additionally as separate key presses
  const avoidRepeated = ['Control', 'Shift']

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
          body: JSON.stringify({key: vimKeyEvent})
        });
        if (!response.ok){
          throw new Error("Failed to fetch data.")
        }
        const responseJSON= await response.json()
        const cursor = responseJSON.cursor


        const finished = responseJSON.finished
        if(finished){
          // Go off and fetch data gain
          console.log("Fetching next level")
          fetchLevelAndCursor()
        }
        // setFinished(responseJSON.finished)
        setCursor({Row: cursor.Row, Column: cursor.Column})

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
      <div className="grid-container">
        {gridData.map((row, rowIndex) => (
          <div key={rowIndex} className="grid-row">
            {row.map((char, colIndex) => (
              <CharacterCell
                key={colIndex}
                char={char}
                isCursor={rowIndex === cursor.Row && colIndex === cursor.Column}
              />
            ))}
          </div>
        ))}
      <h2>{keysPressed}</h2>
          {/* {finished && (<> <div style={{position: "relative", float:"right", }}>
            <h3>Next level</h3>
            <Arrow90degRight  style={{width:"50px", height:"50px"}}onClick={() => getNextLevel()}></Arrow90degRight>

             </div> </>  )} */}
      
      </div>
    </>
  );
};

export default CharacterGrid