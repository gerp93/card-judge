DROP PROCEDURE IF EXISTS SP_SET_NEXT_JUDGE_CARD;

CREATE PROCEDURE SP_SET_NEXT_JUDGE_CARD(IN VAR_LOBBY_ID CHAR(36))
BEGIN
    DECLARE VAR_NEW_CARD_ID CHAR(36) DEFAULT FN_GET_DRAW_PILE_CARD_ID(
            'PROMPT',
            VAR_LOBBY_ID
        );

    DECLARE VAR_NEW_CARD_TEXT VARCHAR(510) DEFAULT (
            SELECT
                TEXT
            FROM CARD
            WHERE ID = VAR_NEW_CARD_ID
        );

    DECLARE VAR_BLANK_COUNT INT DEFAULT (
            SELECT
                ROUND(
                    (
                        LENGTH(VAR_NEW_CARD_TEXT) -
                        LENGTH(REPLACE(VAR_NEW_CARD_TEXT, '_____', ''))
                    ) / LENGTH('_____')
                )
        );

    IF VAR_BLANK_COUNT < 1 THEN
        SET VAR_BLANK_COUNT = 1;
    END
    IF;

    UPDATE JUDGE
    SET CARD_ID = VAR_NEW_CARD_ID,
        BLANK_COUNT = COALESCE(VAR_BLANK_COUNT, 1)
    WHERE LOBBY_ID = VAR_LOBBY_ID;

    DELETE
    FROM DRAW_PILE
    WHERE LOBBY_ID = VAR_LOBBY_ID
        AND CARD_ID = VAR_NEW_CARD_ID;
END;