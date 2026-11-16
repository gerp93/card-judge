-- Adds CARD.LOBBY_ID for per-lobby wild cards on databases provisioned before
-- wild cards stopped being modeled as rows in a hidden DECK. Idempotent.
-- Must run before TR_AUDIT_CARD_* are (re)created, since those triggers
-- reference OLD.LOBBY_ID.
ALTER TABLE CARD ADD COLUMN IF NOT EXISTS LOBBY_ID UUID NULL;
