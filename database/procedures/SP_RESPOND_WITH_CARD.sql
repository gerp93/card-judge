CREATE PROCEDURE IF NOT EXISTS SP_RESPOND_WITH_CARD(
    IN VAR_PLAYER_ID UUID,
    IN VAR_CARD_ID UUID,
    IN VAR_SPECIAL_CATEGORY ENUM (
        'STEAL',
        'SURPRISE',
        'FIND',
        'WILD'
        )
)
BEGIN
    DECLARE VAR_RESPONSE_CARD_ID UUID DEFAULT UUID();

    DECLARE VAR_LOBBY_ID UUID;
    DECLARE VAR_JUDGE_BLANK_COUNT INT;
    DECLARE VAR_RESPONSE_ID UUID;

    SELECT LOBBY_ID
    INTO VAR_LOBBY_ID
    FROM PLAYER
    WHERE ID = VAR_PLAYER_ID;

    SELECT BLANK_COUNT
    INTO VAR_JUDGE_BLANK_COUNT
    FROM JUDGE
    WHERE LOBBY_ID = VAR_LOBBY_ID;

    SELECT ID
    INTO VAR_RESPONSE_ID
    FROM (SELECT R.ID, R.CREATED_ON_DATE, COUNT(RC.ID) AS CARD_COUNT
          FROM RESPONSE AS R
                   LEFT JOIN RESPONSE_CARD AS RC ON RC.RESPONSE_ID = R.ID
          WHERE R.PLAYER_ID = VAR_PLAYER_ID
          GROUP BY R.ID) AS T
    WHERE CARD_COUNT < VAR_JUDGE_BLANK_COUNT
    ORDER BY CREATED_ON_DATE
    LIMIT 1;

    INSERT INTO RESPONSE_CARD (ID,
                               RESPONSE_ID,
                               CARD_ID,
                               SPECIAL_CATEGORY)
    VALUES (VAR_RESPONSE_CARD_ID,
            VAR_RESPONSE_ID,
            VAR_CARD_ID,
            VAR_SPECIAL_CATEGORY);

    INSERT INTO LOG_RESPONSE_CARD (LOBBY_ID,
                                   ROUND_ID,
                                   RESPONSE_ID,
                                   RESPONSE_CARD_ID,
                                   JUDGE_USER_ID,
                                   JUDGE_CARD_ID,
                                   PLAYER_USER_ID,
                                   PLAYER_CARD_ID,
                                   SPECIAL_CATEGORY)
    SELECT L.ID                AS LOBBY_ID,
           L.ROUND_ID          AS ROUND_ID,
           R.ID                AS RESPONSE_ID,
           RC.ID               AS RESPONSE_CARD_ID,
           JP.USER_ID          AS JUDGE_USER_ID,
           J.CARD_ID           AS JUDGE_CARD_ID,
           P.USER_ID           AS PLAYER_USER_ID,
           RC.CARD_ID          AS PLAYER_CARD_ID,
           RC.SPECIAL_CATEGORY AS SPECIAL_CATEGORY
    FROM RESPONSE_CARD AS RC
             INNER JOIN RESPONSE AS R ON R.ID = RC.RESPONSE_ID
             INNER JOIN PLAYER AS P ON P.ID = R.PLAYER_ID
             INNER JOIN LOBBY AS L ON L.ID = P.LOBBY_ID
             INNER JOIN JUDGE AS J ON J.LOBBY_ID = L.ID
             INNER JOIN PLAYER AS JP ON JP.ID = J.PLAYER_ID
    WHERE RC.ID = VAR_RESPONSE_CARD_ID;

    DELETE
    FROM HAND
    WHERE PLAYER_ID = VAR_PLAYER_ID
      AND CARD_ID = VAR_CARD_ID;

    CALL SP_DRAW_HAND(VAR_PLAYER_ID);
END;