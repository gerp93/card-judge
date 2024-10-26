CREATE TABLE IF NOT EXISTS LOG_DRAW
(
    ID               UUID                             NOT NULL DEFAULT UUID(),
    CREATED_ON_DATE  DATETIME                         NOT NULL DEFAULT CURRENT_TIMESTAMP(),

    LOBBY_ID         UUID                             NULL,
    PLAYER_USER_ID   UUID                             NULL,
    CARD_ID          UUID                             NULL,
    SPECIAL_CATEGORY ENUM ('STEAL','SURPRISE','WILD') NULL     DEFAULT NULL,

    PRIMARY KEY (ID),
    FOREIGN KEY (PLAYER_USER_ID) REFERENCES USER (ID) ON DELETE SET NULL,
    FOREIGN KEY (CARD_ID) REFERENCES CARD (ID) ON DELETE SET NULL
);