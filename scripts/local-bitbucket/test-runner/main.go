package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

type TestScenario struct {
	Name                  string   `json:"name"`
	Description           string   `json:"description"`
	Command               string   `json:"command"`
	RunTwice              bool     `json:"run_twice"`
	SetupCommands         []string `json:"setup_commands,omitempty"`
	VerifyCommands        []string `json:"verify_commands,omitempty"`
	ExpectedStructure     []string `json:"expected_structure"`
	Disabled              bool     `json:"disabled,omitempty"`
	SkipTokenVerification bool     `json:"skip_token_verification,omitempty"`
}

type TestConfig struct {
	TestScenarios []TestScenario `json:"test_scenarios"`
}

type TestContext struct {
	BaseURL  string
	Token    string
	GhorgDir string
}

type TestRunner struct {
	config  *TestConfig
	context *TestContext
}

func NewTestRunner(configPath string, context *TestContext) (*TestRunner, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read test config: %w", err)
	}

	config := &TestConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse test config: %w", err)
	}

	return &TestRunner{
		config:  config,
		context: context,
	}, nil
}

func (tr *TestRunner) RunAllTests() error {
	log.Printf("Starting integration tests with %d scenarios...", len(tr.config.TestScenarios))

	// Ensure the ghorg directory exists
	if err := tr.ensureGhorgDirectoryExists(); err != nil {
		return fmt.Errorf("failed to create ghorg directory: %w", err)
	}

	// Clean up any existing test directories
	if err := tr.cleanupTestDirectories(); err != nil {
		log.Printf("Warning: Failed to clean up test directories: %v", err)
	}

	passed := 0
	failed := 0
	skipped := 0

	for i, scenario := range tr.config.TestScenarios {
		log.Printf("\n=== Running Test %d/%d: %s ===", i+1, len(tr.config.TestScenarios), scenario.Name)
		log.Printf("Description: %s", scenario.Description)

		if scenario.Disabled {
			log.Printf("⏭️ SKIPPED: %s (test is disabled)", scenario.Name)
			skipped++
			continue
		}

		// Extended pre-test server health check with more retries
		log.Printf("🔍 Checking server health before test...")
		if err := tr.waitForServerRecovery(10, 5*time.Second); err != nil {
			log.Printf("⚠️  Server health check failed before test: %v", err)
			log.Printf("❌ FAILED: %s - server not healthy", scenario.Name)
			failed++

			// Wait longer before continuing after server health failure
			if i < len(tr.config.TestScenarios)-1 {
						log.Printf("⏳ Skipping server recovery wait for faster testing...")
		// time.Sleep removed
			}
			continue
		}

		// Additional delay before starting test to ensure server stability
			log.Printf("⏳ Starting test immediately...")
	time.Sleep(1 * time.Second) // Minimal coordination delay

		// Execute the test
		if err := tr.runTest(&scenario); err != nil {
			log.Printf("❌ FAILED: %s - %v", scenario.Name, err)
			failed++

			// Exponential backoff after failed tests
			waitTime := 15 + (failed * 5) // Start at 15s, add 5s per failure
			if waitTime > 60 {
				waitTime = 60 // Cap at 60 seconds
			}

			if i < len(tr.config.TestScenarios)-1 {
				log.Printf("⏳ Waiting %d seconds after failed test for server recovery (attempt %d)...", waitTime, failed)
				// time.Sleep removed for faster execution

				// Additional health check after failure to ensure recovery
				log.Printf("🔍 Verifying server recovery after failure...")
				if err := tr.waitForServerRecovery(15, 5*time.Second); err != nil {
					log.Printf("⚠️  Server failed to recover properly: %v", err)
					log.Printf("⏳ Additional 30 second recovery wait...")
					// time.Sleep removed for faster execution
				}
			}
		} else {
			log.Printf("✅ PASSED: %s", scenario.Name)
			passed++

			// Longer delay between successful tests to prevent server overload
			if i < len(tr.config.TestScenarios)-1 {
				log.Printf("⏳ Waiting 8 seconds before next test...")
				// time.Sleep removed for faster execution
			}
		}
	}

	log.Printf("\n=== Test Results ===")
	log.Printf("Passed: %d", passed)
	log.Printf("Failed: %d", failed)
	log.Printf("Skipped: %d", skipped)
	log.Printf("Total: %d", len(tr.config.TestScenarios))

	if failed > 0 {
		return fmt.Errorf("integration tests failed: %d passed, %d failed, %d skipped", passed, failed, skipped)
	}

	return nil
}

func (tr *TestRunner) ensureGhorgDirectoryExists() error {
	if err := os.MkdirAll(tr.context.GhorgDir, 0755); err != nil {
		return fmt.Errorf("failed to create ghorg directory %s: %w", tr.context.GhorgDir, err)
	}
	return nil
}

func (tr *TestRunner) cleanupTestDirectories() error {
	log.Printf("Cleaning up test directories in %s...", tr.context.GhorgDir)

	// Remove all directories that start with "local-bitbucket-"
	entries, err := os.ReadDir(tr.context.GhorgDir)
	if err != nil {
		return fmt.Errorf("failed to read ghorg directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "local-bitbucket-") {
			dirPath := filepath.Join(tr.context.GhorgDir, entry.Name())
			log.Printf("Removing test directory: %s", dirPath)
			if err := os.RemoveAll(dirPath); err != nil {
				log.Printf("Warning: Failed to remove directory %s: %v", dirPath, err)
			}
		}
	}

	// Also clean up any test directories in /tmp
	tmpEntries, err := os.ReadDir("/tmp")
	if err == nil {
		for _, entry := range tmpEntries {
			if entry.IsDir() && strings.HasPrefix(entry.Name(), "bitbucket-") {
				dirPath := filepath.Join("/tmp", entry.Name())
				log.Printf("Removing tmp test directory: %s", dirPath)
				if err := os.RemoveAll(dirPath); err != nil {
					log.Printf("Warning: Failed to remove tmp directory %s: %v", dirPath, err)
				}
			}
		}
	}

	return nil
}

// checkServerHealth performs a comprehensive health check on the Bitbucket Server
func (tr *TestRunner) checkServerHealth() error {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	// First check: Basic status endpoint
	statusResp, err := client.Get(tr.context.BaseURL + "/status")
	if err != nil {
		return fmt.Errorf("server status check failed: %w", err)
	}
	defer statusResp.Body.Close()

	if statusResp.StatusCode != http.StatusOK {
		return fmt.Errorf("server status check failed with status %d", statusResp.StatusCode)
	}

	// Second check: Try to access the projects API with authentication
	req, err := http.NewRequest("GET", tr.context.BaseURL+"/rest/api/1.0/projects", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}
	req.SetBasicAuth("admin", tr.context.Token)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("server API health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server API health check failed with status %d", resp.StatusCode)
	}

	// Third check: Verify we can access a specific project (LBP1)
	projectReq, err := http.NewRequest("GET", tr.context.BaseURL+"/rest/api/1.0/projects/LBP1", nil)
	if err != nil {
		return fmt.Errorf("failed to create project health check request: %w", err)
	}
	projectReq.SetBasicAuth("admin", tr.context.Token)

	projectResp, err := client.Do(projectReq)
	if err != nil {
		return fmt.Errorf("project health check failed: %w", err)
	}
	defer projectResp.Body.Close()

	if projectResp.StatusCode != http.StatusOK {
		return fmt.Errorf("project health check failed with status %d", projectResp.StatusCode)
	}

	return nil
}

// waitForServerRecovery waits for the server to be healthy with retries
func (tr *TestRunner) waitForServerRecovery(retries int, delayBetweenRetries time.Duration) error {
	for i := 0; i < retries; i++ {
		if err := tr.checkServerHealth(); err == nil {
			if i > 0 {
				log.Printf("✅ Server recovered after %d attempts", i+1)
			}
			return nil
		}

		if i < retries-1 {
			log.Printf("⏳ Server not ready (attempt %d/%d), waiting %v...", i+1, retries, delayBetweenRetries)
			time.Sleep(delayBetweenRetries)
		}
	}

	return fmt.Errorf("server failed to recover after %d attempts", retries)
}

func (tr *TestRunner) runTest(scenario *TestScenario) error {
	// Run setup commands if any
	for _, setupCmd := range scenario.SetupCommands {
		log.Printf("Running setup command: %s", setupCmd)
		if err := tr.executeCommand(setupCmd); err != nil {
			return fmt.Errorf("setup command failed: %w", err)
		}
	}

	// Render the command template with context variables
	command, err := tr.renderCommand(scenario.Command)
	if err != nil {
		return fmt.Errorf("failed to render command: %w", err)
	}

	log.Printf("Executing: %s", command)

	// Run the command
	if err := tr.executeCommand(command); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	// Run the command twice if specified
	if scenario.RunTwice {
		log.Printf("Running command again (run_twice=true)...")
		if err := tr.executeCommand(command); err != nil {
			return fmt.Errorf("second command execution failed: %w", err)
		}
	}

	// Verify the expected structure
	if len(scenario.ExpectedStructure) > 0 {
		if err := tr.verifyExpectedStructure(scenario.ExpectedStructure); err != nil {
			return fmt.Errorf("structure verification failed: %w", err)
		}
	}

	// Run verification commands if any
	for _, verifyCmd := range scenario.VerifyCommands {
		log.Printf("Running verification command: %s", verifyCmd)
		if err := tr.executeCommand(verifyCmd); err != nil {
			return fmt.Errorf("verification command failed: %w", err)
		}
	}

	return nil
}

func (tr *TestRunner) renderCommand(commandTemplate string) (string, error) {
	tmpl, err := template.New("command").Parse(commandTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse command template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, tr.context); err != nil {
		return "", fmt.Errorf("failed to execute command template: %w", err)
	}

	return buf.String(), nil
}

func (tr *TestRunner) executeCommand(command string) error {
	// Change to the ghorg directory
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if err := os.Chdir(tr.context.GhorgDir); err != nil {
		return fmt.Errorf("failed to change to ghorg directory: %w", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			log.Printf("Warning: Failed to change back to original directory: %v", err)
		}
	}()

	// Split the command into parts
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	// Execute the command
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set environment variables for Bitbucket Server testing
	env := os.Environ()
	env = append(env, "GHORG_INSECURE_BITBUCKET_CLIENT=true")
	env = append(env, "GHORG_CLONE_PROTOCOL=https")
	cmd.Env = env

	return cmd.Run()
}

func (tr *TestRunner) verifyExpectedStructure(expectedPaths []string) error {
	log.Printf("Verifying expected structure (%d paths)...", len(expectedPaths))

	missing := []string{}
	for _, expectedPath := range expectedPaths {
		// Handle absolute paths and relative paths
		var fullPath string
		if filepath.IsAbs(expectedPath) {
			fullPath = expectedPath
		} else {
			fullPath = filepath.Join(tr.context.GhorgDir, expectedPath)
		}

		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			missing = append(missing, expectedPath)
			log.Printf("❌ Missing: %s (checked: %s)", expectedPath, fullPath)
		} else {
			log.Printf("✅ Found: %s", expectedPath)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing expected paths: %s", strings.Join(missing, ", "))
	}

	log.Printf("All expected paths verified successfully!")
	return nil
}

func main() {
	var (
		token      = flag.String("token", "", "Bitbucket Server API token or password")
		baseURL    = flag.String("base-url", "http://bitbucket.example.com:7990", "Bitbucket Server base URL")
		ghorgDir   = flag.String("ghorg-dir", "", "Directory where ghorg will clone repositories")
		configPath = flag.String("config", "configs/test-scenarios.json", "Path to test scenarios configuration file")
	)
	flag.Parse()

	if *token == "" {
		log.Fatal("Token is required")
	}

	if *ghorgDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
		*ghorgDir = filepath.Join(homeDir, "ghorg")
	}

	context := &TestContext{
		BaseURL:  *baseURL,
		Token:    *token,
		GhorgDir: *ghorgDir,
	}

	runner, err := NewTestRunner(*configPath, context)
	if err != nil {
		log.Fatalf("Failed to create test runner: %v", err)
	}

	if err := runner.RunAllTests(); err != nil {
		log.Fatalf("Integration tests failed: %v", err)
	}

	log.Println("All integration tests passed successfully!")
}
