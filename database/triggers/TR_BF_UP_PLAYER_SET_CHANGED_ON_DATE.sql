CREATE TRIGGER IF NOT EXISTS TR_BF_UP_PLAYER_SET_CHANGED_ON_DATE
    BEFORE UPDATE
    ON PLAYER
    FOR EACH ROW
    SET NEW.CHANGED_ON_DATE = CURRENT_TIMESTAMP();