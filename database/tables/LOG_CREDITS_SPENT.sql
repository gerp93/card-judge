CREATE TABLE IF NOT EXISTS LOG_CREDITS_SPENT
(
    ID               UUID     NOT NULL DEFAULT UUID(),
    CREATED_ON_DATE  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP(),

    LOBBY_ID         UUID     NOT NULL,
    USER_ID          UUID     NOT NULL,
    AMOUNT           INT      NOT NULL,
    CATEGORY         ENUM (
        'WINNING-STREAK',
        'LOSING-STREAK',
        'PURCHASE',
        'GAMBLE',
        'GAMBLE-WIN',
        'BET',
        'BET-WIN',
        'EXTRA-RESPONSE',
        'STEAL',
        'STEAL-VICTIM',
        'SURPRISE',
        'FIND',
        'WILD'
        )                     NOT NULL,

    PRIMARY KEY (ID)
);