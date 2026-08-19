-- Adds the CARD.LOBBY_ID -> LOBBY foreign key (ON DELETE CASCADE) on
-- pre-existing databases. The constraint is named FK_CARD_LOBBY to match the
-- name in CARD.sql, so IF NOT EXISTS makes this a no-op on fresh databases
-- (where CREATE TABLE already added it) rather than creating a duplicate.
ALTER TABLE CARD
    ADD FOREIGN KEY IF NOT EXISTS FK_CARD_LOBBY (LOBBY_ID) REFERENCES LOBBY(ID) ON DELETE CASCADE;
