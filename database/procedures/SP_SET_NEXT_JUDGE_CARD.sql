CREATE PROCEDURE IF NOT EXISTS SP_SET_NEXT_JUDGE_CARD(
    IN VAR_LOBBY_ID UUID
)
BEGIN
    DECLARE VAR_NEW_CARD_ID UUID;
    DECLARE VAR_NEW_CARD_TEXT VARCHAR(510);
    DECLARE VAR_BLANK_COUNT INT;

    SELECT C.ID, C.TEXT
    INTO VAR_NEW_CARD_ID, VAR_NEW_CARD_TEXT
    FROM DRAW_PILE AS DP
             INNER JOIN CARD AS C ON C.ID = DP.CARD_ID
    WHERE C.CATEGORY = 'PROMPT'
      AND DP.LOBBY_ID = VAR_LOBBY_ID
    ORDER BY RAND()
    LIMIT 1;

    SELECT ROUND((LENGTH(VAR_NEW_CARD_TEXT) - LENGTH(REPLACE(VAR_NEW_CARD_TEXT, '_____', ''))) / LENGTH('_____'))
    INTO VAR_BLANK_COUNT;

    IF VAR_BLANK_COUNT < 1 THEN
        SET VAR_BLANK_COUNT = 1;
    END IF;

    UPDATE JUDGE
    SET CARD_ID     = VAR_NEW_CARD_ID,
        BLANK_COUNT = COALESCE(VAR_BLANK_COUNT, 1)
    WHERE LOBBY_ID = VAR_LOBBY_ID;

    DELETE
    FROM DRAW_PILE
    WHERE LOBBY_ID = VAR_LOBBY_ID
      AND CARD_ID = VAR_NEW_CARD_ID;
END;