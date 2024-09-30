import { useEffect } from "react";
import CharacterGrid from "../components/CharacterGrid";

const Home = () => {
  useEffect(() => {}, []);

  const gridData = [
  ['H', 'e', 'l', 'l', 'o'],
  ['W', 'o', 'r', 'l', 'd']
];

  const cursorPosition = { row: 1, col: 3 }; // Example cursor position from backend API

  return (
    <CharacterGrid gridData={gridData} cursorPosition={cursorPosition} />
  );

};

export default Home;
