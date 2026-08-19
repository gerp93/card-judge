# CLAUDE.md — card-judge

Guidance for working in this repository. This file is a **style guide first, an
architecture map second**. It documents the conventions already in use so that
changes match the existing codebase. Match the surrounding code; do not
introduce new styles, formatters, or abstractions.

> Front-end conventions (themes, HTMX, CSS, templates) live in
> `.github/copilot-instructions.md` and still apply — this file does not repeat
> them, it points at them. Read both.

---

## What this is

An "Apples-to-Apples"/Cards-Against-Humanity–style party game. Players join a
lobby, a rotating **judge** plays a PROMPT card, everyone else plays RESPONSE
cards, the judge picks a winner. On top of that sits a bespoke **credits +
specials + perks** economy (wild/surprise/steal/find cards, betting, gambling,
handicap, streaks).

Stack: **Go (stdlib `net/http`) + HTMX + `gorilla/websocket` + MariaDB.** No web
framework, no ORM, no build step for the front-end.

## Layout

The repo root is a thin wrapper. **All application code lives under `src/`,
which is the Go module root** (`module github.com/grantfbarnes/card-judge`, Go
1.22.5). The reusable platform lives in the separate
**`github.com/gerp93/gameshell-framework`** module (auth, page middleware,
user/lobby-shell/player-base data layer, websocket hub, framework schema) —
this repo holds only the game.

```
src/
  main.go            entry point: registers the Game impl + framework params,
                     DB connect, framework schema then game schema, ALL route
                     wiring, server
  go.mod             module + framework dependency (pinned version)
  game/              hooks.go — CardJudge implements gameshell.Game
  api/               game HTTP handlers, grouped by domain
    pages/           full-page renderers (package apiPages)
    user/ deck/ card/ lobby/ access/ stats/   (packages apiUser, apiDeck, ...)
  database/          game data-access: one file per domain, hand-written SQL
    database.go      unexported query()/execute() delegating to the framework
  static/            embedded assets (//go:embed)
    static.go        embed.FS + SQLFiles (ORDERED game schema manifest, runs
                     AFTER the framework schema)
    sql/             game tables/ views/ functions/ procedures/ events/ triggers/
    html/            pages/ (base.html + body/*) and components/ (HTMX fragments)
    css/ js/ images/
```

There is intentionally **no `cmd/`, `internal/`, or `pkg/`** — flat top-level
packages under `src/`. Keep it that way. Handlers that need framework data
functions import them as `gsDatabase "github.com/gerp93/gameshell-framework/database"`
alongside the game `database` package.

## The most important architectural fact

**Most game logic lives in MariaDB**, not Go. Stored procedures (`SP_*`),
functions (`FN_*`), triggers (`TR_*`), and views (`V_*`) under `src/static/sql/`
implement the rules. The Go `database/*.go` functions are mostly thin wrappers
that `CALL SP_...` or `SELECT FN_...`. When you change game behavior, you are
usually editing SQL, not Go.

Schema is applied by iterating `static.SQLFiles` (in `src/static/static.go`) on
every server start via `database.RunFile`. **Order matters and is manual** —
when you add a SQL object, add it to `SQLFiles` in dependency order
(settings → tables → views → functions → procedures → events → triggers).
`setup.sql` (which drops/creates the database) is **not** in that list; it is run
by hand once (see `src/static/sql/README.md`).

## Go conventions (match these exactly)

- **Package naming:** subpackages under `api/` are named `api<Thing>` even though
  the directory is lowercase — package `apiUser` in `api/user/`, `apiPages` in
  `api/pages/`, `apiLobby` in `api/lobby/`. Top-level packages (`database`,
  `auth`, `websocket`, `static`, `api`) match their directory. `gofmt`/tabs.
- **Handlers** have the shape `func Name(w http.ResponseWriter, r *http.Request)`
  and are wired in `main.go` with Go 1.22 method+pattern routes
  (`http.Handle("POST /api/...", api.MiddlewareForAPIs(http.HandlerFunc(...)))`).
- **Form/param parsing** uses the range-switch idiom, not a decode library:
  ```go
  for key, val := range r.Form {
      switch key {
      case "name":
          name = val[0]
      }
  }
  ```
- **Responses are plain text**, written directly — no JSON envelope:
  ```go
  w.WriteHeader(http.StatusBadRequest)
  _, _ = w.Write([]byte("No name found."))
  ```
  Messages are human-readable sentences, capitalized, ending with a period. The
  `_, _ =` discard on `Write` is deliberate and consistent — keep it.
- **DB layer:** raw SQL strings passed to the unexported `query`/`execute`
  helpers in `database.go`. Multi-line SQL uses backtick literals; one-liners use
  double quotes. Read results row-by-row with `defer rows.Close()` then
  `rows.Scan(...)`. On scan error the pattern is
  `log.Println(err); return ..., errors.New("failed to scan row in query results")`.
  Structs mirror table columns (PascalCase fields, `sql.Null*` for nullables).
  No ORM, no query builder — do not introduce one.
- **IDs** are `uuid.UUID` (`github.com/google/uuid`), generated with
  `uuid.NewUUID()` in Go or `UUID()` in SQL.
- **Config** is environment variables via `os.Getenv`, all prefixed
  `CARD_JUDGE_` (`_SQL_HOST/_SQL_DATABASE/_SQL_USER/_SQL_PASSWORD`, `_PORT`,
  `_LOG_FILE`, `_CERT_FILE`, `_KEY_FILE`). No config files or libraries.
- **Comments are sparse.** Handlers and DB functions carry none; the
  `websocket/` files keep their gorilla-example doc comments. Match the density
  of the file you are editing.

## SQL conventions (match these exactly)

- **Uppercase everything** — keywords AND identifiers (table/column/proc names).
- **One database object per file**, named after the object, using prefixes:
  `SP_` procedure, `FN_` function, `TR_` trigger, `EVT_` event, `V_` view,
  `AUDIT_`/`LOG_` history tables.
- Tables use `CREATE TABLE IF NOT EXISTS`; procedures/functions/views/triggers/
  events use `CREATE OR REPLACE` so re-running the manifest is idempotent.
- Procedure/function local variables are prefixed `VAR_`
  (e.g. `DECLARE VAR_LOBBY_ID UUID DEFAULT ...`).
- **Format with the repo's formatter**, not by hand:
  `src/static/sql/sqlfmt.sh` runs `sqlfmt --newlines --upper --spaces 4
  --comment-pre-space` over every `*.sql`. Run it after editing SQL.
- After adding/removing a SQL file, update `SQLFiles` in `src/static/static.go`.

## Real-time (websocket) pattern

Messages over the socket are **short control strings, not structured payloads**.
The server broadcasts a hint (e.g. `refresh`, `refresh-player-hand`,
`refresh-lobby-game-board`, `timer;;...`, `alert;;...`) and the browser
(`src/static/js/lobby.js`) reacts by re-fetching the relevant HTML fragment via
`htmx.ajax` from a `/api/lobby/{lobbyId}/html/...` route. HTML is never pushed
over the socket. Handlers trigger updates with `websocket.LobbyBroadcast(...)`
and `websocket.PlayerBroadcast(...)` (`src/websocket/hub.go`). Keep this
"broadcast a hint → client re-fetches fragment" model.

One hub exists per lobby (`lobbyHubs` map). Presence is `PLAYER.IS_ACTIVE`, flipped
on websocket connect (`AddUserToLobby` → `SP_SET_PLAYER_ACTIVE`) and disconnect
(`hub.unregisterClient` → `SetPlayerInactive` → `SP_SET_PLAYER_INACTIVE`).

## Build / run / verify

- Build: `cd src && go build ./...` (release builds `go build -o card-judge`).
- Run: needs a MariaDB reachable via the `CARD_JUDGE_SQL_*` env vars; create the
  DB once with `src/static/sql/setup.sql`, then the server applies the rest of
  the schema on startup. Serves on `:2016` (or `CARD_JUDGE_PORT`).
- Docker: root `Dockerfile` builds and runs the binary.
- Versioning: git tag `vMAJOR.MINOR.PATCH` triggers `.github/workflows/release.yml`;
  bump with `version_bump.sh {major|minor|patch}`.
- There is no automated test suite; **verify changes by running the app and
  playing through the affected flow** (create lobby, join, play a round,
  including a wild card and a surprise card, judge a winner).

## Known quirks (preserve unless explicitly changing)

- The full SQL schema re-runs on every startup (idempotent by design).
- The lobby is **deleted when its last websocket client disconnects**
  (`websocket/hub.go`).
- The auth signing secret is process-random (`auth/cookie.go`), so sessions do
  not survive a restart and cannot be shared across instances.

---

## The "Gameshell Framework" split (in progress)

This repo is being split so the reusable platform becomes a separately
versioned Go module, `github.com/gerp93/gameshell-framework`, that many games
depend on. The split is a **move, not a rewrite** — all conventions above are
preserved. Landed so far (see PR #9):

- **Slim base tables:** `LOBBY(ID, CREATED_ON_DATE, NAME, MESSAGE,
  PASSWORD_HASH)` and `PLAYER(ID, CREATED_ON_DATE, LOBBY_ID, USER_ID,
  JOIN_ORDER, IS_ACTIVE)` hold only platform columns. card-judge's game
  columns live in **1:1 extension tables** (`CJ_LOBBY_SETTINGS`,
  `CJ_PLAYER_STATE`) joined by FK; game `SP_*`/`FN_*` read/write those.
- **`Game` interface with lifecycle hooks** (`src/gameshell/gameshell.go`):
  `OnRoomCreated`, `OnPlayerJoined`, `OnRoomEmpty`. card-judge implements it in
  `src/game/hooks.go` as thin wrappers over `SP_CJ_INIT_LOBBY`,
  `SP_CJ_INIT_PLAYER`, `SP_CJ_CLEANUP_LOBBY`. These replaced the game
  bootstrap/cleanup triggers (`TR_LOBBY_AFTER_INSERT`, `TR_PLAYER_AFTER_INSERT`,
  `TR_LOBBY_AFTER_DELETE`) — do not reintroduce game logic in triggers on the
  base tables. `main.go` registers the game via `gameshell.Register`.
- **Declarative page policy:** required-login/admin paths are set by the app
  at startup (`api.SetPagePolicy` in `main.go`), and the brand + cookie prefix
  are parameters (`api.SetBrandName`, `auth.SetCookiePrefix`).
- **Dependency direction is one-way:** only `main.go` imports `game`;
  `gameshell` imports nothing from the app. The framework must never import a
  game.
- Framework schema loads **before** game schema (extension-table FKs depend on
  it).
- **Upgrade caveat:** the startup manifest only creates/replaces objects, so
  removing a SQL object from `SQLFiles` does not drop it from an existing
  database — dropped objects (e.g. the replaced triggers) need a manual `DROP`
  on already-provisioned databases, or a fresh `setup.sql`.

The physical extraction is DONE: the framework code lives in the
`gameshell-framework` repo and card-judge consumes it as a module dependency.
Until the framework repo is published and tagged, `go.mod` carries a
temporary `replace` directive pointing at a sibling checkout — remove it and
pin the tagged version once `github.com/gerp93/gameshell-framework` exists.
