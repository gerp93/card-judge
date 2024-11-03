CREATE TABLE IF NOT EXISTS CARD
(
    ID              UUID                       NOT NULL DEFAULT UUID(),
    CREATED_ON_DATE DATETIME                   NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    CHANGED_ON_DATE DATETIME                   NOT NULL DEFAULT CURRENT_TIMESTAMP(),

    DECK_ID         UUID                       NOT NULL,
    CATEGORY        ENUM ('PROMPT','RESPONSE') NOT NULL DEFAULT 'PROMPT',
    TEXT            VARCHAR(510)               NOT NULL,
    IMAGE           BLOB                       NULL,

    PRIMARY KEY (ID),
    FOREIGN KEY (DECK_ID) REFERENCES DECK (ID) ON DELETE CASCADE,
    CONSTRAINT DECK_TEXT_UNIQUE UNIQUE (DECK_ID, TEXT)
);