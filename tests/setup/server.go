package setup

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/grantfbarnes/card-judge/tests/util"
)

// ServerManager handles starting and stopping the test server
type ServerManager struct {
	cmd            *exec.Cmd
	alreadyRunning bool
	baseURL        string
}

// NewServerManager creates a new server manager
func NewServerManager(baseURL string) *ServerManager {
	return &ServerManager{
		baseURL: baseURL,
	}
}

// Start starts the server if not already running
func (sm *ServerManager) Start() error {
	// Check if server is already running
	_, err := http.Get(sm.baseURL)
	sm.alreadyRunning = err == nil

	if sm.alreadyRunning {
		log.Println("⚠️  WARNING: Server is already running!")
		log.Printf("⚠️  Make sure it's using the test database (%s)\n", util.TestDatabaseName)
		log.Println("⚠️  If not, stop the server and run this script again.")
		fmt.Print("\nContinue anyway? (y/n): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			return fmt.Errorf("aborted by user")
		}
		return nil
	}

	log.Println("Server not running. Starting server with test database...")

	// Start server with test database
	sm.cmd = exec.Command("go", "run", "main.go")
	sm.cmd.Dir = "../../src"
	sm.cmd.Env = append(os.Environ(),
		fmt.Sprintf("CARD_JUDGE_SQL_DATABASE=%s", util.TestDatabaseName),
	)

	log.Printf("Starting server with: CARD_JUDGE_SQL_DATABASE=%s\n", util.TestDatabaseName)

	// Start the server in background
	if err := sm.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	log.Println("Waiting for server to be ready...")
	// Wait for server to be ready
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		time.Sleep(1 * time.Second)
		resp, err := http.Get(sm.baseURL)
		if err == nil {
			resp.Body.Close()
			log.Println("Server is ready!")
			return nil
		}
		if i == maxAttempts-1 {
			sm.Stop()
			return fmt.Errorf("server failed to start within 30 seconds")
		}
	}

	return nil
}

// Stop stops the server if it was started by this manager
func (sm *ServerManager) Stop() {
	if !sm.alreadyRunning && sm.cmd != nil && sm.cmd.Process != nil {
		log.Println("Stopping server...")
		KillServerOnPort(util.DefaultPort)
		time.Sleep(500 * time.Millisecond)
	}
}
