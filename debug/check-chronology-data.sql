-- Diagnostic script for Chronology game issues
-- Run this to check if your game has data

-- Check if any Chronology games exist
SELECT 'Chronology Games' AS Check_Type, COUNT(*) AS Count FROM CHRONOLOGY_GAME;

-- Check draw pile cards
SELECT 'Draw Pile Cards' AS Check_Type, COUNT(*) AS Count FROM CHRONOLOGY_DRAW_PILE;

-- Check how many draw pile cards have years
SELECT 'Cards With Years' AS Check_Type, COUNT(*) AS Count 
FROM CHRONOLOGY_DRAW_PILE 
WHERE CARD_YEAR > 0;

-- Check cards without years
SELECT 'Cards Without Years' AS Check_Type, COUNT(*) AS Count 
FROM CHRONOLOGY_DRAW_PILE 
WHERE CARD_YEAR = 0;

-- Check current cards
SELECT 'Current Cards' AS Check_Type, COUNT(*) AS Count FROM CHRONOLOGY_CURRENT_CARD;

-- Show details of a specific game (change the game status condition as needed)
SELECT 
    'Game Details' AS Info,
    CG.ID AS GameId,
    L.NAME AS LobbyName,
    CG.GAME_STATUS AS Status,
    CG.CARDS_TO_WIN AS CardsToWin,
    (SELECT COUNT(*) FROM CHRONOLOGY_DRAW_PILE WHERE CHRONOLOGY_GAME_ID = CG.ID) AS DrawPileSize,
    (SELECT COUNT(*) FROM CHRONOLOGY_DRAW_PILE WHERE CHRONOLOGY_GAME_ID = CG.ID AND CARD_YEAR > 0) AS CardsWithYears,
    (SELECT COUNT(*) FROM CHRONOLOGY_DRAW_PILE WHERE CHRONOLOGY_GAME_ID = CG.ID AND DRAWN = 0) AS UndrawnCards
FROM CHRONOLOGY_GAME CG
INNER JOIN LOBBY L ON L.ID = CG.LOBBY_ID;

-- Show sample cards from draw pile (to check if years are being extracted)
SELECT 
    DP.ID,
    C.TEXT AS CardText,
    DP.CARD_YEAR AS ExtractedYear,
    DP.DRAWN AS IsDrawn
FROM CHRONOLOGY_DRAW_PILE DP
INNER JOIN CARD C ON C.ID = DP.CARD_ID
LIMIT 10;

-- Show available decks with PROMPT cards
SELECT 
    D.NAME AS DeckName,
    COUNT(*) AS PromptCardCount,
    GROUP_CONCAT(SUBSTRING(C.TEXT, 1, 50) SEPARATOR ' | ') AS SampleCards
FROM DECK D
INNER JOIN CARD C ON C.DECK_ID = D.ID
WHERE C.CATEGORY = 'PROMPT'
GROUP BY D.ID, D.NAME;
