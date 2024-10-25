import { useEffect } from "react";
import CharacterGrid from "../components/CharacterGrid";
import { useState } from "react";
import { useContext } from "react";
import UserContext from "../context/UserContext";

const Home = () => {
  const [level, setLevel] = useState(null)
  const [cellColours, setCellColours] = useState(null)
  const fingerprint = useContext(UserContext)

  const fetchLevel = async () => {
    try{
      // const response = await fetch('host.docker.internal:8080/level', {
      const response = await fetch('http://localhost:8080/level', {
        method: 'POST',
        headers: {'Content-type': 'application/json'},
        body: JSON.stringify({auth_key: fingerprint})
      });
      if (!response.ok){
        throw new Error("Failed to fetch data.")
      }

      const levelJSON = await response.json()
      console.log(levelJSON)

      setLevel(levelJSON.level)
      setCellColours(levelJSON.cellColours)

      return levelJSON.cursor

    } catch (error){
      console.error("Error fetching levels", error)
      return null 
    }
  }
  useEffect(() => {
    fetchLevel()
  }, [])


  return (
    <CharacterGrid gridData={level} cellColours={cellColours} fetchLevel={fetchLevel} />
  );

};

export default Home;
