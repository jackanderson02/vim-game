import CharacterCell from "./CharacterCell";
import "../css/grid.css"
const CharacterGrid = ({ gridData, cursorPosition, cellWidth, cellHeight }) => {
  return (
    <div className="grid-container">
      {gridData.map((row, rowIndex) => (
        <div key={rowIndex} className="grid-row">
          {row.map((char, colIndex) => (
            <CharacterCell
              key={colIndex}
              char={char}
              isCursor={rowIndex === cursorPosition.row && colIndex === cursorPosition.col}
              cellWidth={cellWidth}
              cellHeight={cellHeight}
            />
          ))}
        </div>
      ))}
    </div>
  );
};

export default CharacterGrid;