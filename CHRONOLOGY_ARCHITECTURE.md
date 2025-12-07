# Chronology Game Module - Architecture Plan

## Overview
Add a new **Chronology** game mode to the Card Judge app while maintaining complete separation from the existing Cards Against Humanity (CAH) implementation. The goal is to reuse generic infrastructure (users, decks, cards, permissions, lobbies, WebSocket) without modifying existing CAH-specific code.

## Key Principles
1. **Minimal modifications to existing code** - No changes to CAH logic, stats, or database tables
2. **Complete game separation** - Chronology has its own API routes, handlers, and UI
3. **Shared infrastructure** - Reuse user auth, deck/card system, permissions, lobbies, WebSocket
4. **Distinct lobby types** - Lobby table gets a `game_type` column to distinguish CAH vs Chronology lobbies
5. **Independent game state** - Chronology maintains its own game state separate from CAH player/response tables

---

## 1. Database Layer Changes

### Minimal Changes (Backward Compatible)
```sql
-- Add game_type column to LOBBY table (default: 'cah' for existing lobbies)
ALTER TABLE LOBBY ADD COLUMN game_type ENUM('cah', 'chronology') DEFAULT 'cah';

-- Create new Chronology-specific tables
CREATE TABLE CHRONOLOGY_GAME (
    id CHAR(36) PRIMARY KEY,
    lobby_id CHAR(36) NOT NULL UNIQUE,
    created_on_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    current_year INT NULL,           -- Year being evaluated
    current_player_id CHAR(36) NULL, -- Player whose turn it is
    timeline_state JSON,             -- Chronology timeline state (array of placed events)
    game_status ENUM('setup', 'active', 'finished') DEFAULT 'setup',
    winner_id CHAR(36) NULL,
    FOREIGN KEY (lobby_id) REFERENCES LOBBY(id) ON DELETE CASCADE,
    FOREIGN KEY (current_player_id) REFERENCES PLAYER(id),
    FOREIGN KEY (winner_id) REFERENCES USER(id)
);

-- Chronology player state (per-player, per-game data)
CREATE TABLE CHRONOLOGY_PLAYER_STATE (
    id CHAR(36) PRIMARY KEY,
    chronology_game_id CHAR(36) NOT NULL,
    player_id CHAR(36) NOT NULL,
    hand_card_ids JSON,              -- Array of card IDs in player's hand
    score INT DEFAULT 0,
    eliminated BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (chronology_game_id) REFERENCES CHRONOLOGY_GAME(id) ON DELETE CASCADE,
    FOREIGN KEY (player_id) REFERENCES PLAYER(id) ON DELETE CASCADE,
    UNIQUE KEY unique_game_player (chronology_game_id, player_id)
);

-- Timeline card placements (for audit/replay)
CREATE TABLE CHRONOLOGY_TIMELINE_PLACEMENT (
    id CHAR(36) PRIMARY KEY,
    chronology_game_id CHAR(36) NOT NULL,
    card_id CHAR(36) NOT NULL,
    player_id CHAR(36) NOT NULL,
    correct_position INT,            -- Whether placement was correct (0 or 1)
    round_number INT,
    placed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (chronology_game_id) REFERENCES CHRONOLOGY_GAME(id) ON DELETE CASCADE,
    FOREIGN KEY (card_id) REFERENCES CARD(id),
    FOREIGN KEY (player_id) REFERENCES PLAYER(id)
);
```

**Why this approach:**
- `game_type` on LOBBY allows filtering lobbies by game type
- New tables are completely isolated - no modifications to existing CAH tables
- Uses same PLAYER table (reuses existing player/lobby relationships)
- Uses same CARD table (reuses deck card associations)
- Uses same USER/DECK/access control mechanisms

### No Changes Needed To:
- USER table (authentication, permissions)
- DECK table (deck definitions, ownership)
- CARD table (card content, types)
- LOBBY table (except adding `game_type` column)
- PLAYER table (already generic - can be reused)
- USER_ACCESS_DECK, USER_ACCESS_LOBBY (permission system)
- All CAH-specific tables (RESPONSE, HAND, JUDGE, WIN, KICK, etc.)

---

## 2. File Structure

```
src/
├── api/
│   ├── cah/                           # NEW: CAH-specific package (optional refactor)
│   │   ├── lobby.go                   # Move CAH lobby logic here
│   │   └── ... CAH handlers
│   │
│   ├── chronology/                    # NEW: Chronology game logic
│   │   ├── chronology.go              # Main Chronology API handlers
│   │   ├── game.go                    # Game creation, initialization
│   │   ├── play.go                    # Card placement, turn logic
│   │   ├── state.go                   # Get current game state
│   │   └── timeline.go                # Timeline management, scoring
│   │
│   ├── lobby/                         # EXISTING: Generic lobby (no changes)
│   │   └── ... remains as-is
│   │
│   ├── pages/                         # EXISTING: Generic pages
│   │   ├── pages.go                   # Add Chronology page routes
│   │   └── ... mostly unchanged
│   │
│   ├── card/                          # EXISTING: Generic card operations
│   ├── deck/                          # EXISTING: Generic deck operations
│   ├── user/                          # EXISTING: Generic user operations
│   └── ... other generic handlers
│
├── database/
│   ├── database.go                    # EXISTING: Connection (no changes)
│   ├── card.go                        # EXISTING: Generic card queries (no changes)
│   ├── deck.go                        # EXISTING: Generic deck queries (no changes)
│   ├── user.go                        # EXISTING: Generic user queries (no changes)
│   ├── lobby.go                       # EXISTING: Generic lobby queries (no changes)
│   │
│   ├── chronology.go                  # NEW: Chronology-specific queries
│   │   ├── CreateChronologyGame()
│   │   ├── GetChronologyGame()
│   │   ├── UpdateTimelineState()
│   │   ├── PlaceCard()
│   │   ├── GetPlayerHand()
│   │   ├── SetPlayerScore()
│   │   └── ... all Chronology DB ops
│   │
│   └── ... other tables remain unchanged
│
├── static/
│   ├── html/
│   │   ├── pages/body/
│   │   │   ├── ... existing CAH pages
│   │   │   ├── chronology.html        # NEW: Chronology lobby page
│   │   │   ├── chronology-game.html   # NEW: Chronology game interface
│   │   │   └── chronology-lobbies.html # NEW: List Chronology lobbies
│   │   │
│   │   └── components/
│   │       ├── chronology/            # NEW: Chronology-specific components
│   │       │   ├── timeline.html      # Timeline display
│   │       │   ├── card-placement.html # Card placement UI
│   │       │   ├── player-hand.html   # Player hand for Chronology
│   │       │   └── game-info.html     # Game state display
│   │       │
│   │       └── ... existing components unchanged
│   │
│   ├── css/
│   │   ├── chronology.css             # NEW: Chronology-specific styles
│   │   └── ... existing CSS unchanged
│   │
│   ├── js/
│   │   ├── chronology.js              # NEW: Chronology game logic
│   │   └── ... existing JS unchanged
│   │
│   └── sql/
│       ├── tables/
│       │   ├── CHRONOLOGY_GAME.sql    # NEW: Chronology game table
│       │   ├── CHRONOLOGY_PLAYER_STATE.sql # NEW
│       │   └── CHRONOLOGY_TIMELINE_PLACEMENT.sql # NEW
│       │
│       ├── procedures/
│       │   ├── SP_CHRONOLOGY_CREATE_GAME.sql        # NEW
│       │   ├── SP_CHRONOLOGY_PLACE_CARD.sql         # NEW
│       │   ├── SP_CHRONOLOGY_NEXT_TURN.sql          # NEW
│       │   └── ... other Chronology procedures
│       │
│       └── ... existing procedures unchanged
│
└── websocket/
    ├── hub.go                         # EXISTING: Generic WebSocket hub (no changes)
    └── client.go                      # EXISTING: Generic client handler (no changes)
```

---

## 3. Reuse Strategy

### Completely Reusable Components
✅ **User Authentication & Authorization**
- Existing middleware: `api.MiddlewareForPages()`, `api.MiddlewareForAPIs()`
- Permission checks: `database.UserHasDeckAccess()`, `database.UserHasLobbyAccess()`
- User session management (already in auth package)

✅ **Deck & Card System**
- `database.GetReadableDecks()` - Get decks user can access
- `database.GetDeck()` - Get deck details
- `database.SearchCardsInDeck()` - Find cards in a deck
- Card types, CARD table structure (works for any card type)

✅ **Lobby Infrastructure**
- `database.Lobby` struct (add `game_type` field)
- `database.Player` struct (generic player representation)
- `database.AddUserToLobby()` - Add user to lobby
- `database.GetLobby()` - Fetch lobby state
- WebSocket hub for real-time updates (generic per-lobby messaging)

✅ **Access Control**
- `database.UserHasDeckAccess()` - Check deck permissions
- `database.UserHasLobbyAccess()` - Check lobby access (will extend for Chronology)
- PASSWORD_HASH mechanism (reuse for lobby passwords if needed)

✅ **WebSocket/Real-time Updates**
- `websocket.Hub` - Generic lobby message broadcasting
- `websocket.Client` - Already handles per-lobby updates
- No changes needed; just use for Chronology game state updates

### Components Requiring Extension (Not Modification)
🔄 **Lobby Creation & List Pages**
- Extend `apiPages.Lobbies()` to show Chronology and CAH lobbies separately (or add filter)
- Add new `apiPages.ChronologyLobbies()` page
- Lobby creation form needs game type selector

🔄 **Lobby Access Check**
- Extend `database.UserHasLobbyAccess()` logic (add game_type check if needed)
- OR create `database.UserHasChronologyLobbyAccess()` wrapper

### Game-Specific New Code
❌ **Chronology-Specific** (must be new)
- Card placement logic (validate position on timeline)
- Timeline state management
- Turn rotation (different from CAH judge rotation)
- Scoring system (place card correctly = points)
- Chronology handlers: `api/chronology/`
- Chronology database layer: `database/chronology.go`
- Chronology templates and styles

---

## 4. API Routes

### New Routes for Chronology
```go
// Pages
GET  /chronology                           // Chronology lobbies list
GET  /chronology/lobbies                   // Same as above (or dashboard)
GET  /chronology-lobby/{lobbyId}           // Chronology game page
GET  /chronology-lobby/{lobbyId}/access    // Chronology access gate

// Chronology API
POST /api/chronology/create                // Create Chronology game
GET  /api/chronology/{lobbyId}/state       // Get current game state
POST /api/chronology/{lobbyId}/place-card  // Place a card on timeline
POST /api/chronology/{lobbyId}/next-turn   // Advance to next turn/player
GET  /api/chronology/{lobbyId}/timeline    // Get timeline HTML
GET  /api/chronology/{lobbyId}/hand        // Get player hand HTML
POST /api/chronology/{lobbyId}/end-game    // End game, determine winner
```

### Reused Routes
```go
// Generic (work for all game types via middleware)
POST /api/user/login                       // User authentication
POST /api/access/lobby/{lobbyId}           // Check lobby password
POST /api/access/deck/{deckId}             // Check deck password

// Deck/Card management (works across all lobbies)
GET  /api/deck/{deckId}/cards              // Get cards in deck (generic)
POST /api/card/find                        // Search cards (generic)

// WebSocket
GET  /ws/lobby/{lobbyId}                   // Works for any lobby type (Chronology or CAH)
```

---

## 5. Main.go Changes

```go
// In main.go

// SQL files to load
sqlFiles := []string{
    // ... existing CAH tables and procedures
    
    // NEW: Chronology tables and procedures
    "sql/tables/CHRONOLOGY_GAME.sql",
    "sql/tables/CHRONOLOGY_PLAYER_STATE.sql",
    "sql/tables/CHRONOLOGY_TIMELINE_PLACEMENT.sql",
    
    "sql/procedures/SP_CHRONOLOGY_CREATE_GAME.sql",
    "sql/procedures/SP_CHRONOLOGY_PLACE_CARD.sql",
    "sql/procedures/SP_CHRONOLOGY_NEXT_TURN.sql",
    // ... more Chronology procedures
}

// HTTP routes
import apiChronology "github.com/grantfbarnes/card-judge/api/chronology"

// Chronology pages
http.Handle("GET /chronology/lobbies", api.MiddlewareForPages(http.HandlerFunc(apiPages.ChronologyLobbies)))
http.Handle("GET /chronology-lobby/{lobbyId}", api.MiddlewareForPages(http.HandlerFunc(apiPages.ChronologyLobby)))
http.Handle("GET /chronology-lobby/{lobbyId}/access", api.MiddlewareForPages(http.HandlerFunc(apiPages.ChronologyLobbyAccess)))

// Chronology APIs
http.Handle("POST /api/chronology/create", api.MiddlewareForAPIs(http.HandlerFunc(apiChronology.Create)))
http.Handle("GET /api/chronology/{lobbyId}/state", api.MiddlewareForAPIs(http.HandlerFunc(apiChronology.GetState)))
http.Handle("POST /api/chronology/{lobbyId}/place-card", api.MiddlewareForAPIs(http.HandlerFunc(apiChronology.PlaceCard)))
// ... more routes
```

---

## 6. Implementation Phases

### Phase 1: Foundation (Minimal CAH Changes)
1. Add `game_type` column to LOBBY table
2. Create Chronology database tables
3. Create `database/chronology.go` with basic queries
4. Create `api/chronology/` package with handlers

### Phase 2: Core Game Logic
1. Implement card placement logic
2. Implement timeline validation
3. Implement turn rotation
4. Implement scoring

### Phase 3: UI & Real-time
1. Create HTML templates for Chronology lobby
2. Create game interface template
3. Create CSS for Chronology
4. Implement WebSocket updates
5. Implement HTMX interactions

### Phase 4: Integration
1. Add Chronology lobby list to home page
2. Update generic lobby access checks for Chronology
3. Test permission system
4. Test WebSocket updates

### Phase 5: Polish
1. Stats/leaderboards (optional - initially skip per requirements)
2. Game replay/history
3. Settings/customization for Chronology lobbies

---

## 7. Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| **New `game_type` column on LOBBY** | Allows filtering/distinguishing lobby types without separate tables |
| **Separate Chronology tables** | Keeps CAH logic untouched; makes Chronology completely removable if needed |
| **Reuse PLAYER table** | Player concept is generic; no need to duplicate |
| **Reuse WebSocket hub** | Generic per-lobby messaging; no CAH-specific logic needed |
| **Separate API package** | Makes it clear which code is Chronology vs CAH |
| **No Chronology stats** | Keeps the codebase simpler; stats table remains CAH-only |
| **Extend, don't modify access queries** | Add wrappers/extensions instead of changing existing database functions |
| **Game state in JSON column** | Flexible for storing complex Chronology timeline structure without schema changes |

---

## 8. Testing Considerations

- **Database migration**: Test that adding `game_type` column doesn't break existing lobbies
- **Lobby creation**: Ensure CAH lobbies still work, can create Chronology lobbies
- **Permissions**: Verify deck/lobby access control works for Chronology
- **WebSocket**: Test real-time updates for Chronology lobbies
- **Lobby isolation**: Ensure Chronology and CAH lobbies don't interfere
- **Card selection**: Ensure card filtering works for different deck types

---

## 9. Example: Creating a Chronology Game

```go
// User navigates to /chronology/lobbies
// Clicks "Create New Game"
// Selects deck (same as CAH, uses database.GetReadableDecks())
// Sets optional lobby password
// Clicks "Start Game"

// Backend (api/chronology/game.go):
func Create(w http.ResponseWriter, r *http.Request) {
    userId := r.Context().Value("userId").(uuid.UUID)
    deckId := uuid.Parse(r.FormValue("deckId"))
    
    // Reuse: Create generic lobby
    lobby, err := database.CreateLobby(deckId, "Chronology Game")
    
    // NEW: Create Chronology-specific game
    game, err := database.CreateChronologyGame(lobby.Id, deckId)
    
    // Reuse: Add user to lobby
    playerId, err := database.AddUserToLobby(lobby.Id, userId)
    
    // NEW: Initialize player hand in Chronology
    err = database.InitializeChronologyPlayerHand(game.Id, playerId, deckId)
    
    // Redirect to game
    http.Redirect(w, r, fmt.Sprintf("/chronology-lobby/%s", lobby.Id), http.StatusSeeOther)
}
```

---

## Summary

**What changes in existing code:** Minimal
- Add `game_type` column to LOBBY table only
- Add routes to `main.go` for Chronology
- Extend some page handlers to show Chronology lobbies

**What's new:**
- Complete Chronology API package
- Chronology database layer
- Chronology SQL tables/procedures
- Chronology templates, CSS, JavaScript
- Chronology game logic (timelines, scoring, turns)

**What stays untouched:**
- CAH game logic
- User auth/permissions
- Deck/card system
- WebSocket infrastructure
- All CAH-specific code

This approach keeps the two games completely separate while sharing the generic infrastructure they both need.
