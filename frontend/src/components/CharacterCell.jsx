const CharacterCell = ({ char, isCursor, cellWidth, cellHeight }) => {
  return (
    <span
      className={`cell ${isCursor ? 'cursor' : ''}`}
      style={{ width: cellWidth, height: cellHeight }}
    >
      {char}
    </span>
  );
};

export default CharacterCell