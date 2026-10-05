CREATE
OR REPLACE TRIGGER TR_AUDIT_CARD_UPDATE
BEFORE UPDATE ON CARD
FOR EACH ROW
BEGIN
    -- Only real deck cards are audited; per-lobby wild cards (LOBBY_ID set) are
    -- ephemeral gameplay artifacts and are not part of the deck audit trail.
    IF OLD.LOBBY_ID IS NULL THEN
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
            'UPDATE',
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
