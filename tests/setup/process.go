package setup

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// KillServerOnPort kills any process listening on the specified port
func KillServerOnPort(port int) error {
	log.Printf("Attempting to kill any server on port %d...\n", port)

	switch runtime.GOOS {
	case "windows":
		return killServerWindows(port)
	case "linux":
		return killServerLinux(port)
	case "darwin":
		return killServerMacOS(port)
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func killServerWindows(port int) error {
	// Use netstat to find the PID listening on the port
	cmd := exec.Command("netstat", "-ano")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Warning: Failed to run netstat: %v\n", err)
		return nil
	}

	// Parse netstat output to find PID on our port
	lines := strings.Split(string(output), "\n")
	portStr := fmt.Sprintf(":%d", port)
	var pidToKill string

	for _, line := range lines {
		if strings.Contains(line, portStr) && strings.Contains(line, "LISTENING") {
			// Extract PID from the end of the line
			fields := strings.Fields(line)
			if len(fields) > 0 {
				pidToKill = fields[len(fields)-1]
				break
			}
		}
	}

	if pidToKill != "" {
		log.Printf("Found process %s on port %d, killing...\n", pidToKill, port)
		killCmd := exec.Command("taskkill", "/F", "/PID", pidToKill)
		if err := killCmd.Run(); err != nil {
			log.Printf("Warning: Failed to kill process: %v\n", err)
		} else {
			log.Printf("✓ Killed process %s\n", pidToKill)
		}
	} else {
		log.Printf("No process found on port %d\n", port)
	}

	time.Sleep(500 * time.Millisecond)
	return nil
}

func killServerLinux(port int) error {
	// Use fuser to find and kill process on port
	cmd := exec.Command("fuser", "-k", fmt.Sprintf("%d/tcp", port))
	if err := cmd.Run(); err != nil {
		// fuser returns error if no process found, which is fine
		log.Printf("No process found on port %d (or fuser not available)\n", port)
	}
	time.Sleep(500 * time.Millisecond)
	return nil
}

func killServerMacOS(port int) error {
	// Use lsof to find and kill process on port
	cmd := exec.Command("lsof", "-ti", fmt.Sprintf(":%d", port))
	output, err := cmd.Output()
	if err != nil {
		log.Printf("No process found on port %d\n", port)
		return nil
	}

	pid := strings.TrimSpace(string(output))
	if pid != "" {
		killCmd := exec.Command("kill", "-9", pid)
		if err := killCmd.Run(); err != nil {
			return fmt.Errorf("failed to kill process %s: %w", pid, err)
		}
		log.Printf("Killed process %s on port %d\n", pid, port)
	}

	time.Sleep(500 * time.Millisecond)
	return nil
}
