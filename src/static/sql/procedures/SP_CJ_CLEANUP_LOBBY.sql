CREATE
OR REPLACE PROCEDURE SP_CJ_CLEANUP_LOBBY(IN VAR_LOBBY_ID UUID)
BEGIN
    -- Remove this lobby's ephemeral wild cards. (The LOBBY -> CARD FK cascade is
    -- a backstop for when the lobby row itself is deleted; this makes cleanup
    -- explicit at room-empty time, before the base LOBBY row is removed.)
    DELETE
    FROM CARD
    WHERE LOBBY_ID = VAR_LOBBY_ID;
END;
