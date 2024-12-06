CREATE PROCEDURE IF NOT EXISTS SP_SET_WINNING_STREAK(
    IN VAR_WINNER_PLAYER_ID UUID
)
BEGIN
    DECLARE VAR_LOBBY_ID UUID;
    DECLARE VAR_JUDGE_PLAYER_ID UUID;
    DECLARE VAR_LOBBY_THRESHOLD INT;

    SELECT LOBBY_ID
    INTO VAR_LOBBY_ID
    FROM PLAYER
    WHERE ID = VAR_WINNER_PLAYER_ID;

    SELECT PLAYER_ID
    INTO VAR_JUDGE_PLAYER_ID
    FROM JUDGE
    WHERE LOBBY_ID = VAR_LOBBY_ID;

    SELECT WIN_STREAK_THRESHOLD
    INTO VAR_LOBBY_THRESHOLD
    FROM LOBBY
    WHERE ID = VAR_LOBBY_ID;

    -- INCREMENT WINNING STREAK OF WINNER
    UPDATE PLAYER
    SET WINNING_STREAK = WINNING_STREAK + 1
    WHERE ID = VAR_WINNER_PLAYER_ID;

    -- RESET WINNING STREAK OF LOSERS
    UPDATE PLAYER
    SET WINNING_STREAK = 0
    WHERE LOBBY_ID = VAR_LOBBY_ID
      AND IS_ACTIVE = 1
      AND ID <> VAR_WINNER_PLAYER_ID
      AND ID <> VAR_JUDGE_PLAYER_ID;

    -- SPEND CREDIT AND RESET STREAK AFTER BREAKING THRESHOLD
    UPDATE PLAYER
    SET CREDITS_SPENT = CREDITS_SPENT + 1,
        WINNING_STREAK = 0
    WHERE LOBBY_ID = VAR_LOBBY_ID
      AND IS_ACTIVE = 1
      AND WINNING_STREAK >= VAR_LOBBY_THRESHOLD;
END;