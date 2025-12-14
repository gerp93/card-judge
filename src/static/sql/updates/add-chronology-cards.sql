-- Add sample historical event cards to Chronology 1 deck
-- These cards have years that can be parsed for the timeline game

SET @deck_id = '88026803-d22a-11f0-b4d2-60cf84649547';

INSERT INTO CARD (ID, DECK_ID, CATEGORY, TEXT, CREATED_ON_DATE, CHANGED_ON_DATE) VALUES
(UUID(), @deck_id, 'PROMPT', '1776 - American Declaration of Independence signed', NOW(), NOW()),
(UUID(), @deck_id, 'PROMPT', '1492 - Columbus reaches the Americas', NOW(), NOW()),
(UUID(), @deck_id, 'PROMPT', '1969 - First humans land on the Moon', NOW(), NOW()),
(UUID(), @deck_id, 'PROMPT', '1945 - World War II ends', NOW(), NOW()),
(UUID(), @deck_id, 'PROMPT', '1914 - World War I begins', NOW(), NOW()),
(UUID(), @deck_id, 'PROMPT', '1865 - American Civil War ends', NOW(), NOW()),
(UUID(), @deck_id, 'PROMPT', '1989 - Fall of the Berlin Wall', NOW(), NOW()),
(UUID(), @deck_id, 'PROMPT', '2001 - September 11 terrorist attacks', NOW(), NOW()),
(UUID(), @deck_id, 'PROMPT', '1991 - World Wide Web becomes publicly available', NOW(), NOW()),
(UUID(), @deck_id, 'PROMPT', '1903 - Wright brothers first powered flight', NOW(), NOW()),
(UUID(), @deck_id, 'PROMPT', '1939 - World War II begins', NOW(), NOW()),
(UUID(), @deck_id, 'PROMPT', '1066 - Battle of Hastings', NOW(), NOW()),
(UUID(), @deck_id, 'PROMPT', '1215 - Magna Carta signed', NOW(), NOW()),
(UUID(), @deck_id, 'PROMPT', '1517 - Protestant Reformation begins', NOW(), NOW()),
(UUID(), @deck_id, 'PROMPT', '1789 - French Revolution begins', NOW(), NOW()),
(UUID(), @deck_id, 'PROMPT', '1861 - American Civil War begins', NOW(), NOW()),
(UUID(), @deck_id, 'PROMPT', '1929 - Stock market crash, Great Depression begins', NOW(), NOW()),
(UUID(), @deck_id, 'PROMPT', '1963 - JFK assassination', NOW(), NOW()),
(UUID(), @deck_id, 'PROMPT', '2007 - First iPhone released', NOW(), NOW()),
(UUID(), @deck_id, 'PROMPT', '1981 - First Space Shuttle launch', NOW(), NOW()),
(UUID(), @deck_id, 'PROMPT', '1955 - Rosa Parks refuses to give up bus seat', NOW(), NOW()),
(UUID(), @deck_id, 'PROMPT', '1620 - Pilgrims arrive at Plymouth Rock', NOW(), NOW()),
(UUID(), @deck_id, 'PROMPT', '1804 - Napoleon crowned Emperor of France', NOW(), NOW()),
(UUID(), @deck_id, 'PROMPT', '1918 - Spanish Flu pandemic begins', NOW(), NOW()),
(UUID(), @deck_id, 'PROMPT', '2020 - COVID-19 pandemic declared', NOW(), NOW());

SELECT CONCAT('Added ', COUNT(*), ' historical event cards to Chronology 1 deck') AS Result
FROM CARD 
WHERE DECK_ID = @deck_id;
