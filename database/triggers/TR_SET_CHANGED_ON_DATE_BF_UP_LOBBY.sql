CREATE TRIGGER IF NOT EXISTS TR_SET_CHANGED_ON_DATE_BF_UP_LOBBY
    BEFORE UPDATE
    ON LOBBY
    FOR EACH ROW
    SET NEW.CHANGED_ON_DATE = CURRENT_TIMESTAMP();