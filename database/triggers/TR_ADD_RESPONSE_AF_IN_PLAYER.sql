CREATE TRIGGER IF NOT EXISTS TR_ADD_RESPONSE_AF_IN_PLAYER
    AFTER INSERT
    ON PLAYER
    FOR EACH ROW
BEGIN
    DECLARE VAR_CREATED_COUNT INT DEFAULT 0;
    DECLARE VAR_RESPONSE_COUNT INT;

    SELECT RESPONSE_COUNT
    INTO VAR_RESPONSE_COUNT
    FROM JUDGE
    WHERE LOBBY_ID = NEW.LOBBY_ID;

    WHILE VAR_CREATED_COUNT < VAR_RESPONSE_COUNT
        DO
            INSERT INTO RESPONSE (PLAYER_ID) VALUE (NEW.ID);
            SET VAR_CREATED_COUNT = VAR_CREATED_COUNT + 1;
        END WHILE;
END;