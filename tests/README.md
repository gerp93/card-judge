# Tests

Development and testing utilities for the Card Judge project.

## Theme Validator

Automated screenshot tool for visual regression testing of all theme variations across all pages.

**Location:** `tests/theme-validator/`

### Quick Start

From the repository root:

**PowerShell / Command Prompt (Windows):**
```powershell
cd tests/theme-validator
go run .
```

**Bash / Shell (macOS / Linux):**
```bash
cd tests/theme-validator
go run .
```

That's it! The tool will automatically:
- Kill any existing server on port 2016
- Drop and recreate the test database (`card_judge_test`) from SQL scripts
- Seed test data (test users, decks with cards, lobbies, etc.)
- Start the server with the test database
- Prompt you to select which pages to capture
- Log in as the test user (Test1/password)
- Capture screenshots across all theme variations for every page (including variations showing modals/alerts)
- Generate PDF reports from the screenshots
- Save screenshots to `./screenshots/` and PDFs to `./theme-reports/`
- Stop the server when done

### What Gets Captured

The validator can screenshot all pages across all 22 theme variations. You're prompted to select which pages to capture:

**Available Pages:**
- home, about, account
- stats (with sub-pages: leaderboard, users, cards)
- users, lobbies, decks
- deck-view, lobby-game

**Output Structure:**
```
screenshots/
├── YYYYMMDD_HHMMSS/
│   ├── home/
│   │   ├── dark-theme.png
│   │   ├── hawkeye-theme.png
│   │   └── ... (22 themes total)
│   ├── about/
│   ├── account/
│   └── ...

theme-reports/
├── YYYYMMDD_HHMMSS/
│   ├── dark-theme.pdf
│   ├── hawkeye-theme.pdf
│   └── ... (one PDF per theme)
```

### Requirements

- MySQL or MariaDB installed and running locally
- Chrome, Edge, or Opera browser installed
- Go 1.24+ (for running the validator)

### Browser Detection

The validator automatically detects which Chromium-based browser is available on your system:

**Windows:** Checks for Edge → Chrome → Opera (both 32-bit and 64-bit Program Files locations)
**macOS:** Checks for Microsoft Edge → Google Chrome → Opera
**Linux:** Uses system PATH to find `microsoft-edge`, `google-chrome`, `chromium`, `chromium-browser`, or `opera`

If a browser is found, you'll see a log message like:
```
Using browser: C:\Program Files\Google\Chrome\Application\chrome.exe
```

If no browser paths are found, it will fall back to chromedp's default detection mechanism. This ensures the tool works across different platforms and browser installations without requiring manual configuration.

### How It Works

1. **Database Setup** — Builds `card_judge_test` database from scratch using SQL scripts in `src/static/sql/`
   - Reads the production `setup.sql` template
   - Adapts it in-memory by renaming `CARD_JUDGE` → `card_judge_test`
   - Executes all SQL files in the correct dependency order (shared with `src/main.go`)
2. **Test Data** — Seeds minimal test data (users, decks, cards, lobbies) needed for screenshots
3. **Server Management** — Automatically starts/stops the server with test database
4. **Screenshots** — Uses chromedp (headless browser) to capture all page variations
5. **Theme Testing** — Applies each of 22 themes and captures the result
6. **PDF Generation** — Creates PDF reports for each theme with all captured pages

### Architecture

**`tests/setup/`** — Shared test infrastructure package
- `database.go` — Test database creation and seeding
- `server.go` — Server lifecycle management
- `process.go` — Cross-platform process management
- `sql.go` — SQL file list (synchronized with `src/static/static.go`)
- `testdata.go` — Test data seeding (users, decks, lobbies, stats)

**`tests/theme-validator/`** — Theme screenshot application
- `theme-validator.go` — Main orchestrator
- `themes.go` — Dynamic theme loading from `src/static/css/colors.css`
- `pages.go` — Page configuration definitions
- `login.go` — Login automation
- `theme-report.go` — PDF generation from screenshots

### Cross-Platform

Works on Windows, macOS, and Linux with consistent behavior across all platforms.
- Windows: Uses `taskkill` for process management
- Linux/macOS: Uses standard signal handling
