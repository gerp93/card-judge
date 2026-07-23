-- Makes CARD.DECK_ID nullable so wild cards (keyed by LOBBY_ID instead) can
-- have no deck. Idempotent — re-running simply re-asserts the column type.
ALTER TABLE CARD MODIFY DECK_ID UUID NULL;
