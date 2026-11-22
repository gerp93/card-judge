CREATE
OR REPLACE TRIGGER TR_AUDIT_DECK_DELETE
BEFORE DELETE ON DECK
FOR EACH ROW
BEGIN
    DECLARE VAR_IS_WILD_DECK BOOLEAN DEFAULT (
            SELECT
                IS_LOBBY_WILD_DECK
            FROM DECK
            WHERE ID = OLD.ID
        );

    IF NOT VAR_IS_WILD_DECK THEN
        INSERT INTO AUDIT_CARD(
            AUDIT_TYPE,
            CARD_ID,
            DECK_ID,
            CATEGORY,
            TEXT,
            YOUTUBE,
            IMAGE
        )
        SELECT
            'DELETE',
            ID,
            DECK_ID,
            CATEGORY,
            TEXT,
            YOUTUBE,
            IMAGE
        FROM CARD
        WHERE DECK_ID = OLD.ID;

        INSERT INTO AUDIT_DECK(
            AUDIT_TYPE,
            DECK_ID,
            NAME,
            PASSWORD_HASH,
            IS_PUBLIC_READONLY,
            IS_LOBBY_WILD_DECK
        )
        VALUES (
            'DELETE',
            OLD.ID,
            OLD.NAME,
            OLD.PASSWORD_HASH,
            OLD.IS_PUBLIC_READONLY,
            OLD.IS_LOBBY_WILD_DECK
        );
    END IF;
END;
