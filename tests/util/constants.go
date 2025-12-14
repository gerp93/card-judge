package util

// Test constants
const (
	DefaultPort            = 2016
	DefaultHost            = "localhost"
	TestDatabaseName       = "card_judge_test"
	ProductionDatabaseName = "CARD_JUDGE"
)

// Test user credentials
const (
	TestUsername  = "Test1"
	TestPassword  = "password"
	TestUser1Name = "Test1"
	TestUser2Name = "Test2"
	TestUser3Name = "Test3"
	TestUser4Name = "Test4"
)

// Test data UUIDs
const (
	TestUser1ID            = "496aa604-d4c2-11f0-a722-60cf84649547"
	TestUser2ID            = "496aa604-d4c2-11f0-a722-60cf84649548"
	TestUser3ID            = "496aa604-d4c2-11f0-a722-60cf84649549"
	TestUser4ID            = "496aa604-d4c2-11f0-a722-60cf84649550"
	TestLobby1ID           = "11111111-1111-1111-1111-111111111111"
	TestLobby2ID           = "66666666-6666-6666-6666-666666666666"
	TestCardID             = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	TestDeckIDBase         = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa"
	TestPlayer1IDBase      = "22222222-2222-2222-2222-2222222222"
	TestPlayer1Lobby1ID    = "22222222-2222-2222-2222-222222222200"
	TestPlayer2IDBase      = "33333333-3333-3333-3333-3333333333"
	TestPlayer3IDBase      = "44444444-4444-4444-4444-4444444444"
	TestPlayer4IDBase      = "55555555-5555-5555-5555-5555555555"
	TestResponseCardIDBase = "cccccccc-cccc-cccc-cccc-cccccccccc"
	TestResponseIDBase     = "dddddddd-dddd-dddd-dddd-dddddddddd"
)

// Test data counts
const (
	TestDecksCount           = 11 // 11+ triggers pagination
	TestResponseCardsCount   = 10 // 1 prompt + 10 responses
	TestResponsesCount       = 5  // For stat tracking
	TestResponseLogsCount    = 3  // Response card logs
	TestDiscardsPerCardCount = 12 // For review cards
	TestSkipCardsCount       = 2  // Skip logs
)

// Test data content
const (
	TestPromptCardText = "This is a test prompt card with _."
	TestLobby1Message  = "A lobby for screenshot testing"
)

// Directories
const (
	SQLDir                  = "src/static/sql"
	ColorsCSSPath           = "src/static/css/colors.css"
	ScreenshotDir           = "screenshots"
	ThemeReportDir          = "theme-reports"
	AccessibilityReportFile = "accessibility-report.txt"
)

// Accessibility thresholds
const (
	WCAGLevelAAA               = "AAA"
	WCAGLevelAA                = "AA"
	WCAGLevelA                 = "A"
	MaxViolationsForAAA        = 0
	MaxViolationsForAA         = 2
	ContrastViolationThreshold = 0
)
