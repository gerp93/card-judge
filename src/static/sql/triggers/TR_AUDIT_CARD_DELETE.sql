CREATE
OR REPLACE TRIGGER TR_AUDIT_CARD_DELETE
BEFORE DELETE ON CARD
FOR EACH ROW
BEGIN
    DECLARE VAR_IS_WILD_CARD BOOLEAN DEFAULT (
            SELECT
                IS_HIDDEN
            FROM DECK
            WHERE ID = OLD.DECK_ID
        );

    IF NOT VAR_IS_WILD_CARD THEN
        INSERT INTO AUDIT_CARD(
            AUDIT_TYPE,
            CARD_ID,
            DECK_ID,
            CATEGORY,
            TEXT,
            YOUTUBE,
            IMAGE
        )
        VALUES (
            'DELETE',
            OLD.ID,
            OLD.DECK_ID,
            OLD.CATEGORY,
            OLD.TEXT,
            OLD.YOUTUBE,
            OLD.IMAGE
        );
    END
    IF;
END;