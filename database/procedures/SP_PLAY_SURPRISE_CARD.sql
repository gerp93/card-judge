CREATE PROCEDURE IF NOT EXISTS SP_PLAY_SURPRISE_CARD(
    IN VAR_PLAYER_ID UUID
)
BEGIN
    DECLARE VAR_LOBBY_ID UUID;
    DECLARE VAR_PLAYER_USER_ID UUID;
    DECLARE VAR_CARD_ID UUID;

    SELECT LOBBY_ID,
           USER_ID
    INTO
        VAR_LOBBY_ID,
        VAR_PLAYER_USER_ID
    FROM PLAYER
    WHERE ID = VAR_PLAYER_ID;

    SELECT DP.CARD_ID
    INTO VAR_CARD_ID
    FROM DRAW_PILE AS DP
             INNER JOIN CARD AS C ON C.ID = DP.CARD_ID
    WHERE DP.LOBBY_ID = VAR_LOBBY_ID
      AND C.CATEGORY = 'RESPONSE'
    ORDER BY RAND()
    LIMIT 1;

    DELETE
    FROM DRAW_PILE
    WHERE CARD_ID = VAR_CARD_ID;

    UPDATE PLAYER
    SET CREDITS_SPENT = CREDITS_SPENT + 1
    WHERE ID = VAR_PLAYER_ID;

    INSERT INTO LOG_DRAW (LOBBY_ID, PLAYER_USER_ID, CARD_ID, SPECIAL_CATEGORY)
    VALUES (VAR_LOBBY_ID, VAR_PLAYER_USER_ID, VAR_CARD_ID, 'SURPRISE');

    CALL SP_PLAY_CARD(VAR_PLAYER_ID, VAR_CARD_ID, 'SURPRISE');
END;