CREATE PROCEDURE IF NOT EXISTS SP_SET_NEXT_JUDGE_PLAYER(
    IN VAR_LOBBY_ID UUID
)
BEGIN
    DECLARE VAR_PLAYER_COUNT INT;
    DECLARE VAR_CURRENT_POSITION INT;
    DECLARE VAR_NEXT_POSITION INT;
    DECLARE VAR_NEXT_JUDGE_PLAYER_ID UUID;

    SELECT COUNT(*)
    INTO VAR_PLAYER_COUNT
    FROM PLAYER
    WHERE IS_ACTIVE = 1
      AND LOBBY_ID = VAR_LOBBY_ID;

    SELECT POSITION
    INTO VAR_CURRENT_POSITION
    FROM JUDGE
    WHERE LOBBY_ID = VAR_LOBBY_ID;

    SET VAR_NEXT_POSITION = VAR_CURRENT_POSITION + 1;
    IF VAR_NEXT_POSITION > VAR_PLAYER_COUNT THEN
        SET VAR_NEXT_POSITION = 1;
    END IF;

    SELECT PLAYER_ID
    INTO VAR_NEXT_JUDGE_PLAYER_ID
    FROM (SELECT ID                                                           AS PLAYER_ID,
                 RANK() OVER (PARTITION BY LOBBY_ID ORDER BY CREATED_ON_DATE) AS JOIN_ORDER
          FROM PLAYER
          WHERE IS_ACTIVE = 1
            AND LOBBY_ID = VAR_LOBBY_ID) AS T
    WHERE JOIN_ORDER = VAR_NEXT_POSITION;

    UPDATE JUDGE
    SET POSITION  = VAR_NEXT_POSITION,
        PLAYER_ID = VAR_NEXT_JUDGE_PLAYER_ID
    WHERE LOBBY_ID = VAR_LOBBY_ID;
END;