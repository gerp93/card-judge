DROP EVENT IF EXISTS EVT_CLEAN_BAD_RESPONSE_CARDS;

CREATE EVENT EVT_CLEAN_BAD_RESPONSE_CARDS ON SCHEDULE EVERY 1 DAY
DO
    BEGIN
        DECLARE VAR_DISCARD_COUNT INT;
        SET VAR_DISCARD_COUNT = 10;
        CREATE TEMPORARY TABLE BAD_RESPONSE_CARDS(CARD_ID CHAR(36));

        INSERT INTO BAD_RESPONSE_CARDS(CARD_ID)
        SELECT
            CARD_ID
        FROM (
                SELECT
                    LD.CARD_ID,
                    COUNT(*) AS DISCARD_COUNT
                FROM LOG_DISCARD AS LD
                    INNER JOIN CARD AS C ON C.ID = LD.CARD_ID
                    LEFT JOIN (
                        SELECT
                            PLAYER_CARD_ID AS CARD_ID,
                            CREATED_ON_DATE AS LAST_PLAYED_DATE
                        FROM (
                                SELECT
                                    PLAYER_CARD_ID,
                                    CREATED_ON_DATE,
                                    RANK() OVER (
                                        PARTITION BY PLAYER_CARD_ID
                                        ORDER BY CREATED_ON_DATE DESC
                                    ) AS PLAY_ORDER
                                FROM LOG_RESPONSE_CARD
                            ) AS CARDSPLAYED
                        WHERE PLAY_ORDER = 1
                    ) AS LASTPLAYED ON LASTPLAYED.CARD_ID = LD.CARD_ID
                -- NEVER PLAYED
                WHERE LASTPLAYED.LAST_PLAYED_DATE IS NULL
                    -- DISCARDS SINCE LAST PLAYED
                    OR LASTPLAYED.LAST_PLAYED_DATE < LD.CREATED_ON_DATE
                GROUP BY LD.CARD_ID
            ) AS BADCARDS
        WHERE DISCARD_COUNT > VAR_DISCARD_COUNT;

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
            INNER JOIN BAD_RESPONSE_CARDS AS B ON B.CARD_ID = C.ID;

        DELETE C
        FROM CARD AS C
            INNER JOIN BAD_RESPONSE_CARDS AS B ON B.CARD_ID = C.ID;

        DROP TEMPORARY TABLE BAD_RESPONSE_CARDS;
    END;