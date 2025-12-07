-- Add game_type column to LOBBY table if it doesn't exist
-- Default is 'cah' for backwards compatibility with existing lobbies

ALTER TABLE LOBBY 
ADD COLUMN IF NOT EXISTS GAME_TYPE ENUM('cah', 'chronology') NOT NULL DEFAULT 'cah' AFTER NAME;
