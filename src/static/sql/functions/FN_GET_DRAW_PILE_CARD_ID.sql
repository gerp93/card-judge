DROP FUNCTION IF EXISTS FN_GET_DRAW_PILE_CARD_ID;

CREATE FUNCTION FN_GET_DRAW_PILE_CARD_ID(
    IN VAR_CATEGORY ENUM('PROMPT', 'RESPONSE'),
    IN VAR_LOBBY_ID UUID
)
RETURNS UUID
BEGIN
    DECLARE VAR_LOBBY_DRAW_PRIORITY ENUM('RANDOM', 'PLAYCOUNT') DEFAULT (
            SELECT
                DRAW_PRIORITY
            FROM LOBBY
            WHERE ID = VAR_LOBBY_ID
        );

    IF VAR_LOBBY_DRAW_PRIORITY = 'PLAYCOUNT' THEN
        IF VAR_CATEGORY = 'PROMPT' THEN
        RETURN (
            SELECT
                C.ID
            FROM DRAW_PILE AS DP
                INNER JOIN CARD AS C ON C.ID = DP.CARD_ID
                LEFT JOIN (
                    SELECT
                        JUDGE_CARD_ID AS CARD_ID,
                        COUNT(*) AS PLAY_COUNT
                    FROM LOG_RESPONSE_CARD
                    GROUP BY JUDGE_CARD_ID,
                        ROUND_ID
                ) AS CP ON CP.CARD_ID = C.ID
                LEFT JOIN (
                    SELECT
                        CARD_ID,
                        COUNT(*) AS SKIP_COUNT
                    FROM LOG_SKIP
                    GROUP BY CARD_ID
                ) AS CS ON CS.CARD_ID = C.ID
            WHERE C.CATEGORY = VAR_CATEGORY
                AND DP.LOBBY_ID = VAR_LOBBY_ID
            ORDER BY CP.PLAY_COUNT,
                CS.SKIP_COUNT,
                RAND()
            LIMIT 1
        );
    END
    IF;

    RETURN (
        SELECT
            C.ID
        FROM DRAW_PILE AS DP
            INNER JOIN CARD AS C ON C.ID = DP.CARD_ID
            LEFT JOIN (
                SELECT
                    PLAYER_CARD_ID AS CARD_ID,
                    COUNT(*) AS PLAY_COUNT
                FROM LOG_RESPONSE_CARD
                GROUP BY PLAYER_CARD_ID
            ) AS CP ON CP.CARD_ID = C.ID
            LEFT JOIN (
                SELECT
                    CARD_ID,
                    COUNT(*) AS DISCARD_COUNT
                FROM LOG_DISCARD
                GROUP BY CARD_ID
            ) AS CD ON CD.CARD_ID = C.ID
        WHERE C.CATEGORY = VAR_CATEGORY
            AND DP.LOBBY_ID = VAR_LOBBY_ID
        ORDER BY CP.PLAY_COUNT,
            CD.DISCARD_COUNT,
            RAND()
        LIMIT 1
    );
END
IF;

RETURN (
    SELECT
        C.ID
    FROM DRAW_PILE AS DP
        INNER JOIN CARD AS C ON C.ID = DP.CARD_ID
    WHERE C.CATEGORY = VAR_CATEGORY
        AND DP.LOBBY_ID = VAR_LOBBY_ID
    ORDER BY RAND()
    LIMIT 1
);
END;