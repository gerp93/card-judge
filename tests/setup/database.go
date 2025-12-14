package setup

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/grantfbarnes/card-judge/tests/util"
)

// SetupTestDatabase drops and recreates the test database with clean data from SQL files
func SetupTestDatabase() error {
	log.Println("Setting up test database from SQL files...")

	// Get database config
	config := getDatabaseConfig()

	// Connect to MySQL
	db, err := connectToMySQL(config)
	if err != nil {
		return err
	}
	defer db.Close()

	// Read production setup.sql and adapt it for test database
	// This ensures test DB is created with same config as production
	setupSQL, err := os.ReadFile(getSQLBasePath() + "/setup.sql")
	if err != nil {
		return fmt.Errorf("failed to read setup.sql: %w", err)
	}

	// Adapt for test database (in-memory only, no file modification)
	testSetupSQL := strings.ReplaceAll(string(setupSQL), util.ProductionDatabaseName, util.TestDatabaseName)

	// Split and execute each statement
	statements := strings.Split(testSetupSQL, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute setup statement: %w", err)
		}
	}
	log.Println("✓ Test database created using setup.sql template")

	// Connect to test database
	testDB, err := connectToTestDatabase(config)
	if err != nil {
		return err
	}
	defer testDB.Close()

	// Execute remaining SQL files (settings, tables, functions, etc.)
	if err := executeSQLFiles(testDB); err != nil {
		return err
	}

	// Seed test data
	if err := SeedTestData(testDB); err != nil {
		return fmt.Errorf("failed to seed test data: %w", err)
	}

	log.Println("\n✅ Test database setup complete!")
	return nil
}

type dbConfig struct {
	user     string
	password string
	host     string
	port     string
}

func getDatabaseConfig() dbConfig {
	user := os.Getenv("CARD_JUDGE_SQL_USER")
	if user == "" {
		user = "root"
	}
	password := os.Getenv("CARD_JUDGE_SQL_PASSWORD")
	host := os.Getenv("CARD_JUDGE_SQL_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("CARD_JUDGE_SQL_PORT")
	if port == "" {
		port = "3306"
	}

	return dbConfig{user, password, host, port}
}

func connectToMySQL(config dbConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", config.user, config.password, config.host, config.port)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping MySQL: %w", err)
	}
	log.Println("✓ Connected to MySQL")

	return db, nil
}

func connectToTestDatabase(config dbConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?multiStatements=true", config.user, config.password, config.host, config.port, util.TestDatabaseName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}
	return db, nil
}

func executeSQLFiles(db *sql.DB) error {
	sqlFiles := getSQLFileList()
	sqlBasePath := getSQLBasePath()

	log.Printf("✓ Executing %d SQL files...\n", len(sqlFiles))
	for i, sqlFile := range sqlFiles {
		filePath := fmt.Sprintf("%s/%s", sqlBasePath, sqlFile)
		if err := executeSQLFile(db, filePath); err != nil {
			return fmt.Errorf("failed to execute %s: %w", sqlFile, err)
		}
		if (i+1)%10 == 0 {
			log.Printf("  Executed %d/%d files...", i+1, len(sqlFiles))
		}
	}
	log.Println("✓ All SQL files executed successfully")

	return nil
}

// executeSQLFile reads and executes a SQL file
func executeSQLFile(db *sql.DB, filePath string) error {
	sqlBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	sqlContent := string(sqlBytes)

	// Execute the SQL (MySQL driver handles multi-statement execution via multiStatements=true in DSN)
	_, err = db.Exec(sqlContent)
	if err != nil {
		return fmt.Errorf("failed to execute SQL: %w", err)
	}

	return nil
}
