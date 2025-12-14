-- Find the Chronology deck and check its cards
SELECT 
    D.ID,
    D.NAME AS DeckName,
    COUNT(C.ID) AS TotalCards,
    SUM(CASE WHEN C.CATEGORY = 'PROMPT' THEN 1 ELSE 0 END) AS PromptCards,
    SUM(CASE WHEN C.CATEGORY = 'RESPONSE' THEN 1 ELSE 0 END) AS ResponseCards
FROM DECK D
LEFT JOIN CARD C ON C.DECK_ID = D.ID
WHERE D.NAME LIKE '%chronology%'
GROUP BY D.ID, D.NAME;

-- Show all cards from Chronology deck
SELECT 
    C.CATEGORY,
    C.TEXT,
    CASE 
        WHEN C.TEXT REGEXP '\\b[12][0-9]{3}\\b' THEN 'HAS YEAR'
        ELSE 'NO YEAR'
    END AS HasYear
FROM DECK D
INNER JOIN CARD C ON C.DECK_ID = D.ID
WHERE D.NAME LIKE '%chronology%';

-- Check what's in the draw pile for the current game
SELECT 
    'Current Game Draw Pile' AS Info,
    COUNT(*) AS TotalCards,
    SUM(CASE WHEN CARD_YEAR > 0 THEN 1 ELSE 0 END) AS CardsWithYears
FROM CHRONOLOGY_DRAW_PILE;

-- Show actual cards in draw pile
SELECT 
    C.TEXT AS CardText,
    DP.CARD_YEAR AS ExtractedYear
FROM CHRONOLOGY_DRAW_PILE DP
INNER JOIN CARD C ON C.ID = DP.CARD_ID
LIMIT 10;
