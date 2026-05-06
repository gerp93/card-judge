DROP EVENT IF EXISTS EVT_CLEAN_BAD_PROMPT_CARDS;

CREATE EVENT EVT_CLEAN_BAD_PROMPT_CARDS ON SCHEDULE EVERY 1 DAY
DO
    BEGIN
        DECLARE VAR_SKIP_COUNT INT;
        SET VAR_SKIP_COUNT = 10;
        CREATE TEMPORARY TABLE BAD_PROMPT_CARDS(CARD_ID CHAR(36));

        INSERT INTO BAD_PROMPT_CARDS(CARD_ID)
        SELECT
            CARD_ID
        FROM (
                SELECT
                    LS.CARD_ID,
                    COUNT(*) AS SKIP_COUNT
                FROM LOG_SKIP AS LS
                    INNER JOIN CARD AS C ON C.ID = LS.CARD_ID
                    LEFT JOIN (
                        SELECT
                            JUDGE_CARD_ID AS CARD_ID,
                            CREATED_ON_DATE AS LAST_PLAYED_DATE
                        FROM (
                                SELECT
                                    JUDGE_CARD_ID,
                                    CREATED_ON_DATE,
                                    RANK() OVER (
                                        PARTITION BY JUDGE_CARD_ID
                                        ORDER BY CREATED_ON_DATE DESC
                                    ) AS PLAY_ORDER
                                FROM LOG_RESPONSE_CARD
                            ) AS CARDSPLAYED
                        WHERE PLAY_ORDER = 1
                    ) AS LASTPLAYED ON LASTPLAYED.CARD_ID = LS.CARD_ID
                -- NEVER PLAYED
                WHERE LASTPLAYED.LAST_PLAYED_DATE IS NULL
                    -- SKIPS SINCE LAST PLAYED
                    OR LASTPLAYED.LAST_PLAYED_DATE < LS.CREATED_ON_DATE
                GROUP BY LS.CARD_ID
            ) AS BADCARDS
        WHERE SKIP_COUNT > VAR_SKIP_COUNT;

        INSERT INTO REVIEW_CARD(
            CARD_ID,
            DECK_ID,
            CATEGORY,
            TEXT,
            YOUTUBE,
            IMAGE
        )
        SELECT
            C.ID AS CARD_ID,
            C.DECK_ID,
            C.CATEGORY,
            C.TEXT,
            C.YOUTUBE,
            C.IMAGE
        FROM CARD AS C
            INNER JOIN BAD_PROMPT_CARDS AS B ON B.CARD_ID = C.ID;

        DELETE C
        FROM CARD AS C
            INNER JOIN BAD_PROMPT_CARDS AS B ON B.CARD_ID = C.ID;

        DROP TEMPORARY TABLE BAD_PROMPT_CARDS;
    END;