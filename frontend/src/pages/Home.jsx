import { useEffect } from "react";
import CharacterGrid from "../components/CharacterGrid";
import { useState } from "react";

const Home = () => {
  const [level, setLevel] = useState([])
  const fetchLevel = async () => {
    try{
      // const response = await fetch('host.docker.internal:8080/level', {
      const response = await fetch('http://localhost:8080/level', {
        method: 'GET',
      });
      if (!response.ok){
        throw new Error("Failed to fetch data.")
      }

      const levelJSON = await response.json()
      console.log(levelJSON)

      setLevel(levelJSON.level)

      return levelJSON.cursor


    } catch (error){
      console.error("Error fetching levels", error)
      return null 
    }
  }
  useEffect(() => {
    fetchLevel()
  }, [])
  // const gridData = [["a"]]


  return (
    <CharacterGrid gridData={level} fetchLevel={fetchLevel} />
  );

};

export default Home;
