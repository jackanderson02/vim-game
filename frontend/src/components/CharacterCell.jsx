const CharacterCell = ({ char, cellColour, isCursor, cellWidth, cellHeight }) => {
  const backgroundColor = cellColour ? cellColour:  'transparent'

  return (
    <span
      className={`cell ${isCursor ? 'cursor' : ''}`}
      style={{ width: cellWidth, height: cellHeight, backgroundColor}}
    >
      {char}
    </span>
  );
};

export default CharacterCell