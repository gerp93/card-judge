CREATE FUNCTION IF NOT EXISTS FN_GET_LOBBY_JUDGE_PLAYER_ID(
    IN VAR_LOBBY_ID UUID
) RETURNS UUID
BEGIN
    RETURN (SELECT P.ID
            FROM PLAYER AS P
                     INNER JOIN JUDGE AS J ON J.PLAYER_ID = P.ID
            WHERE P.LOBBY_ID = VAR_LOBBY_ID
              AND P.IS_ACTIVE = 1);
END;