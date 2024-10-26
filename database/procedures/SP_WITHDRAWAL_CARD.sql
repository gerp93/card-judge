CREATE PROCEDURE IF NOT EXISTS SP_WITHDRAWAL_CARD(
    IN VAR_PLAYER_ID UUID,
    IN VAR_CARD_ID UUID
)
BEGIN
    DECLARE VAR_LOBBY_ID UUID;
    DECLARE VAR_PLAYER_USER_ID UUID;
    DECLARE VAR_JUDGE_USER_ID UUID;

    SELECT LOBBY_ID,
           USER_ID
    INTO
        VAR_LOBBY_ID,
        VAR_PLAYER_USER_ID
    FROM PLAYER
    WHERE ID = VAR_PLAYER_ID;

    SELECT P.USER_ID
    INTO VAR_JUDGE_USER_ID
    FROM JUDGE AS J
             INNER JOIN PLAYER AS P ON P.ID = J.PLAYER_ID
    WHERE J.LOBBY_ID = VAR_LOBBY_ID;

    INSERT INTO HAND (PLAYER_ID, CARD_ID)
    VALUES (VAR_PLAYER_ID, VAR_CARD_ID);

    DELETE
    FROM LOG_PLAY
    WHERE PLAYER_USER_ID = VAR_PLAYER_USER_ID
      AND JUDGE_USER_ID = VAR_JUDGE_USER_ID
      AND CARD_ID = VAR_CARD_ID;

    DELETE
    FROM BOARD
    WHERE LOBBY_ID = VAR_LOBBY_ID
      AND PLAYER_ID = VAR_PLAYER_ID
      AND CARD_ID = VAR_CARD_ID;
END;