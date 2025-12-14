-- Check which decks have PROMPT cards suitable for Chronology

-- Show all decks with PROMPT card counts
SELECT 
    D.ID,
    D.NAME AS DeckName,
    COUNT(C.ID) AS PromptCardCount,
    SUM(CASE WHEN C.TEXT REGEXP '\\b[12][0-9]{3}\\b' THEN 1 ELSE 0 END) AS CardsWithYears
FROM DECK D
LEFT JOIN CARD C ON C.DECK_ID = D.ID AND C.CATEGORY = 'PROMPT'
GROUP BY D.ID, D.NAME
HAVING PromptCardCount > 0
ORDER BY CardsWithYears DESC, PromptCardCount DESC;

-- Show sample PROMPT cards from each deck
SELECT 
    D.NAME AS DeckName,
    C.TEXT AS CardText,
    CASE 
        WHEN C.TEXT REGEXP '\\b[12][0-9]{3}\\b' THEN 'HAS YEAR'
        ELSE 'NO YEAR'
    END AS HasYear
FROM DECK D
INNER JOIN CARD C ON C.DECK_ID = D.ID
WHERE C.CATEGORY = 'PROMPT'
LIMIT 20;
