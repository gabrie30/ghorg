package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

type TestScenario struct {
	Name                   string   `json:"name"`
	Description            string   `json:"description"`
	Command                string   `json:"command"`
	RunTwice               bool     `json:"run_twice"`
	SetupCommands          []string `json:"setup_commands,omitempty"`
	VerifyCommands         []string `json:"verify_commands,omitempty"`
	ExpectedStructure      []string `json:"expected_structure"`
	Disabled               bool     `json:"disabled,omitempty"`
	SkipTokenVerification  bool     `json:"skip_token_verification,omitempty"`
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

		if err := tr.runTest(&scenario); err != nil {
			log.Printf("❌ FAILED: %s - %v", scenario.Name, err)
			failed++
		} else {
			log.Printf("✅ PASSED: %s", scenario.Name)
			passed++
		}
	}

	log.Printf("\n=== Test Results ===")
	log.Printf("Passed: %d", passed)
	log.Printf("Failed: %d", failed)
	log.Printf("Skipped: %d", skipped)
	log.Printf("Total: %d", len(tr.config.TestScenarios))

	if failed > 0 {
		return fmt.Errorf("%d tests failed", failed)
	}

	log.Println("All integration tests passed successfully!")
	return nil
}

func (tr *TestRunner) runTest(scenario *TestScenario) error {
	// Execute setup commands if any
	for _, setupCmd := range scenario.SetupCommands {
		renderedCmd, err := tr.renderTemplate(setupCmd)
		if err != nil {
			return fmt.Errorf("failed to render setup command: %w", err)
		}

		log.Printf("Setup: %s", renderedCmd)
		if err := tr.executeCommand(renderedCmd); err != nil {
			return fmt.Errorf("setup command failed: %w", err)
		}
	}

	// Render the main command
	renderedCmd, err := tr.renderTemplate(scenario.Command)
	if err != nil {
		return fmt.Errorf("failed to render command: %w", err)
	}

	// Execute the command once
	log.Printf("Executing: %s", renderedCmd)
	if err := tr.executeCommand(renderedCmd); err != nil {
		return fmt.Errorf("first execution failed: %w", err)
	}

	// Execute the command twice if specified (for testing clone then pull)
	if scenario.RunTwice {
		log.Printf("Executing (second time): %s", renderedCmd)
		if err := tr.executeCommand(renderedCmd); err != nil {
			return fmt.Errorf("second execution failed: %w", err)
		}
	}

	// Verify the expected structure
	if err := tr.verifyExpectedStructure(scenario.ExpectedStructure); err != nil {
		return fmt.Errorf("structure verification failed: %w", err)
	}

	// Verify no tokens in git remotes by default (unless explicitly skipped)
	if len(scenario.ExpectedStructure) > 0 && !scenario.SkipTokenVerification {
		if err := tr.verifyNoTokensInRemotes(scenario.ExpectedStructure, tr.context.Token); err != nil {
			return fmt.Errorf("token verification failed: %w", err)
		}
	}

	// Execute verification commands if any
	for _, verifyCmd := range scenario.VerifyCommands {
		renderedCmd, err := tr.renderTemplate(verifyCmd)
		if err != nil {
			return fmt.Errorf("failed to render verify command: %w", err)
		}

		log.Printf("Verify: %s", renderedCmd)
		if err := tr.executeCommand(renderedCmd); err != nil {
			return fmt.Errorf("verification command failed: %w", err)
		}
	}

	return nil
}

func (tr *TestRunner) renderTemplate(tmplText string) (string, error) {
	tmpl, err := template.New("command").Parse(tmplText)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, tr.context); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (tr *TestRunner) executeCommand(command string) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = tr.context.GhorgDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %s\nOutput: %s", err, string(output))
	}

	return nil
}

func (tr *TestRunner) verifyExpectedStructure(expectedPaths []string) error {
	log.Printf("Verifying expected structure (%d paths)...", len(expectedPaths))

	for _, expectedPath := range expectedPaths {
		fullPath := filepath.Join(tr.context.GhorgDir, expectedPath)

		if _, err := os.Stat(fullPath); err != nil {
			if os.IsNotExist(err) {
				// Use ghorg ls to check what actually exists
				return fmt.Errorf("expected path does not exist: %s", expectedPath)
			}
			return fmt.Errorf("failed to check path %s: %w", expectedPath, err)
		}

		log.Printf("✓ Found: %s", expectedPath)
	}

	return nil
}

func (tr *TestRunner) verifyNoTokensInRemotes(expectedPaths []string, token string) error {
	log.Printf("Verifying no tokens in git remotes...")

	for _, expectedPath := range expectedPaths {
		fullPath := filepath.Join(tr.context.GhorgDir, expectedPath)

		// Check if this is a git repository directory
		if _, err := os.Stat(filepath.Join(fullPath, ".git")); err != nil {
			if os.IsNotExist(err) {
				// Not a git repository, skip
				continue
			}
			return fmt.Errorf("failed to check .git directory in %s: %w", expectedPath, err)
		}

		// Run git remote -v to get all remotes
		cmd := exec.Command("git", "remote", "-v")
		cmd.Dir = fullPath

		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to get git remotes for %s: %w\nOutput: %s", expectedPath, err, string(output))
		}

		// Check if the token appears in any remote URL
		remoteOutput := string(output)
		if strings.Contains(remoteOutput, token) {
			return fmt.Errorf("token found in git remote URLs for %s:\n%s", expectedPath, remoteOutput)
		}

		log.Printf("✓ No tokens found in remotes for: %s", expectedPath)
	}

	return nil
}

func (tr *TestRunner) ensureGhorgDirectoryExists() error {
	log.Printf("Ensuring ghorg directory exists: %s", tr.context.GhorgDir)

	// Check if directory already exists
	if _, err := os.Stat(tr.context.GhorgDir); err == nil {
		log.Printf("Ghorg directory already exists: %s", tr.context.GhorgDir)
		return nil
	}

	// Create the directory with appropriate permissions
	if err := os.MkdirAll(tr.context.GhorgDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", tr.context.GhorgDir, err)
	}

	log.Printf("Created ghorg directory: %s", tr.context.GhorgDir)
	return nil
}

func (tr *TestRunner) cleanupTestDirectories() error {
	log.Println("Cleaning up test directories...")

	// Delete all folders that start with local-gitlab-* in the ghorg directory
	matches, err := filepath.Glob(filepath.Join(tr.context.GhorgDir, "local-gitlab-*"))
	if err != nil {
		return err
	}

	for _, match := range matches {
		if err := os.RemoveAll(match); err != nil {
			log.Printf("Warning: Failed to remove %s: %v", match, err)
		} else {
			log.Printf("Removed: %s", match)
		}
	}

	// Also clean up gitlab.example.com directory if it exists
	gitlabDir := filepath.Join(tr.context.GhorgDir, "gitlab.example.com")
	if _, err := os.Stat(gitlabDir); err == nil {
		if err := os.RemoveAll(gitlabDir); err != nil {
			log.Printf("Warning: Failed to remove %s: %v", gitlabDir, err)
		} else {
			log.Printf("Removed: %s", gitlabDir)
		}
	}

	return nil
}

func (tr *TestRunner) RunSpecificTest(testName string) error {
	// Ensure the ghorg directory exists
	if err := tr.ensureGhorgDirectoryExists(); err != nil {
		return fmt.Errorf("failed to create ghorg directory: %w", err)
	}

	for _, scenario := range tr.config.TestScenarios {
		if scenario.Name == testName {
			if scenario.Disabled {
				return fmt.Errorf("test '%s' is disabled and cannot be run", testName)
			}
			log.Printf("Running specific test: %s", testName)
			return tr.runTest(&scenario)
		}
	}

	return fmt.Errorf("test not found: %s", testName)
}

func (tr *TestRunner) ListTests() {
	log.Printf("Available tests:")
	for i, scenario := range tr.config.TestScenarios {
		status := ""
		if scenario.Disabled {
			status = " (DISABLED)"
		}
		log.Printf("%d. %s - %s%s", i+1, scenario.Name, scenario.Description, status)
	}
}

func main() {
	var (
		configPath = flag.String("config", "configs/test-scenarios.json", "Path to test scenarios configuration file")
		baseURL    = flag.String("base-url", "http://gitlab.example.com", "GitLab base URL")
		token      = flag.String("token", "", "GitLab API token")
		ghorgDir   = flag.String("ghorg-dir", "", "Ghorg directory (default: $HOME/ghorg)")
		testName   = flag.String("test", "", "Run specific test by name")
		listTests  = flag.Bool("list", false, "List available tests")
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

	if *listTests {
		runner.ListTests()
		return
	}

	if *testName != "" {
		if err := runner.RunSpecificTest(*testName); err != nil {
			log.Fatalf("Test failed: %v", err)
		}
		return
	}

	if err := runner.RunAllTests(); err != nil {
		log.Fatalf("Integration tests failed: %v", err)
	}
}
