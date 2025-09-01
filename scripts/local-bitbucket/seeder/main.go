package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ktrysmt/go-bitbucket"
)

type Repository struct {
	Name                 string `json:"name"`
	InitializeWithReadme bool   `json:"initialize_with_readme"` // Note: Now creates .gitignore instead
	Description          string `json:"description,omitempty"`
	IsPrivate            bool   `json:"is_private"`
}

type Project struct {
	Name         string       `json:"name"`
	Key          string       `json:"key"`
	Description  string       `json:"description"`
	Repositories []Repository `json:"repositories,omitempty"`
}

type Workspace struct {
	Name        string    `json:"name"`
	Key         string    `json:"key"`
	Description string    `json:"description"`
	Projects    []Project `json:"projects,omitempty"`
}

type User struct {
	Username     string       `json:"username"`
	Email        string       `json:"email"`
	Password     string       `json:"password"`
	DisplayName  string       `json:"display_name"`
	Repositories []Repository `json:"repositories,omitempty"`
}

type AdminUser struct {
	Username     string       `json:"username"`
	Password     string       `json:"password"`
	Email        string       `json:"email"`
	DisplayName  string       `json:"display_name"`
	Repositories []Repository `json:"repositories"`
}

type SeedData struct {
	Workspaces []Workspace `json:"workspaces"`
	Users      []User      `json:"users"`
	AdminUser  AdminUser   `json:"admin_user"`
}

type BitbucketSeeder struct {
	client     *bitbucket.Client
	seedData   *SeedData
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

func NewBitbucketSeeder(username, password, baseURL string) (*BitbucketSeeder, error) {
	// Create Bitbucket client with basic authentication
	log.Printf("Creating Bitbucket client with basic authentication for user: %s", username)

	client := bitbucket.NewBasicAuth(username, password)

	// Set the base URL for on-premise Bitbucket Server
	if baseURL != "" {
		parsedURL, err := url.Parse(baseURL)
		if err != nil {
			return nil, fmt.Errorf("invalid base URL: %w", err)
		}
		client.SetApiBaseURL(*parsedURL)
	}

	// Create HTTP client for Bitbucket Server API calls
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &BitbucketSeeder{
		client:     client,
		baseURL:    baseURL,
		username:   username,
		password:   password,
		httpClient: httpClient,
	}, nil
}

func (b *BitbucketSeeder) LoadSeedData(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read seed config: %w", err)
	}

	b.seedData = &SeedData{}
	if err := json.Unmarshal(data, b.seedData); err != nil {
		return fmt.Errorf("failed to parse seed config: %w", err)
	}

	return nil
}

func (b *BitbucketSeeder) CreateWorkspaces() error {
	log.Println("Creating workspaces...")

	for _, workspace := range b.seedData.Workspaces {
		if err := b.createWorkspace(&workspace); err != nil {
			return fmt.Errorf("failed to create workspace %s: %w", workspace.Name, err)
		}
	}
	return nil
}

func (b *BitbucketSeeder) createWorkspace(workspace *Workspace) error {
	log.Printf("Creating workspace: %s", workspace.Name)

	// Note: In Bitbucket Server, workspaces might be created differently
	// For now, we'll assume the workspace exists and focus on creating projects and repositories
	log.Printf("Assuming workspace exists: %s (Key: %s)", workspace.Name, workspace.Key)

	// Create projects in this workspace
	for _, project := range workspace.Projects {
		if err := b.createProject(&project, workspace.Key); err != nil {
			return fmt.Errorf("failed to create project %s in workspace %s: %w", project.Name, workspace.Name, err)
		}
	}

	return nil
}

func (b *BitbucketSeeder) createProject(project *Project, workspaceKey string) error {
	log.Printf("Creating project: %s (Key: %s)", project.Name, project.Key)

	// Create project using Bitbucket Server REST API
	if err := b.createBitbucketServerProject(project); err != nil {
		return fmt.Errorf("failed to create project %s: %w", project.Name, err)
	}

	// Create repositories in this project
	for _, repo := range project.Repositories {
		if err := b.createBitbucketServerRepository(&repo, project.Key); err != nil {
			return fmt.Errorf("failed to create repository %s in project %s: %w", repo.Name, project.Name, err)
		}
	}

	return nil
}

func (b *BitbucketSeeder) createProjectRepository(repo *Repository, workspaceKey, projectKey string) error {
	log.Printf("Creating project repository: %s in project %s/%s", repo.Name, workspaceKey, projectKey)

	repoOptions := &bitbucket.RepositoryOptions{
		Owner:       workspaceKey,
		RepoSlug:    repo.Name,
		Description: repo.Description,
		IsPrivate:   "false", // Convert boolean to string
		HasWiki:     "false",
		HasIssues:   "false",
	}

	if repo.IsPrivate {
		repoOptions.IsPrivate = "true"
	}

	createdRepo, err := b.client.Repositories.Repository.Create(repoOptions)
	if err != nil {
		return fmt.Errorf("failed to create project repository: %w", err)
	}

	log.Printf("Created project repository: %s (Full name: %s)", createdRepo.Name, createdRepo.Full_name)

	// ALWAYS initialize repositories with .gitignore (needed for cloning)
	// Empty repositories cannot be cloned and will cause HTTP 500 errors
	log.Printf("Initializing repository %s with git content...", repo.Name)
	if err := b.initializeBitbucketServerRepositoryWithGitignore(projectKey, repo.Name); err != nil {
		return fmt.Errorf("CRITICAL: Failed to initialize repository %s with git content: %w", repo.Name, err)
	}
	log.Printf("✅ Repository %s successfully initialized with git content", repo.Name)

	return nil
}

func (b *BitbucketSeeder) CreateUsers() error {
	log.Println("Creating users...")

	for _, user := range b.seedData.Users {
		if err := b.createUser(&user); err != nil {
			return fmt.Errorf("failed to create user %s: %w", user.Username, err)
		}
	}
	return nil
}

func (b *BitbucketSeeder) createUser(user *User) error {
	log.Printf("Creating user: %s", user.Username)

	// In Bitbucket Server, create a personal project for this user
	// Generate a safe project key from username
	projectKey := strings.ToUpper(strings.ReplaceAll(user.Username, "-", ""))
	if len(projectKey) > 10 {
		projectKey = projectKey[:10] // Bitbucket project keys have length limits
	}
	projectKey += "PERS" // Add suffix to avoid conflicts

	personalProject := &Project{
		Name:        fmt.Sprintf("%s Personal", user.DisplayName),
		Key:         projectKey,
		Description: fmt.Sprintf("Personal project for %s", user.DisplayName),
	}

	// Create the personal project
	if err := b.createBitbucketServerProject(personalProject); err != nil {
		log.Printf("Warning: Failed to create personal project for %s: %v", user.Username, err)
		log.Printf("Skipping user %s repositories (project creation required)", user.Username)
		return nil // Don't fail the whole process
	}

	// Create repositories in the personal project
	for _, repo := range user.Repositories {
		if err := b.createBitbucketServerRepository(&repo, personalProject.Key); err != nil {
			return fmt.Errorf("failed to create repository %s for user %s: %w", repo.Name, user.Username, err)
		}
	}

	return nil
}

func (b *BitbucketSeeder) CreateAdminUserRepositories() error {
	log.Println("Creating admin user repositories...")

	// Create admin personal project
	adminProject := &Project{
		Name:        "Admin Personal",
		Key:         "ADMIN",
		Description: "Personal project for admin user",
	}

	// Create the admin project
	if err := b.createBitbucketServerProject(adminProject); err != nil {
		log.Printf("Warning: Failed to create admin project: %v", err)
		log.Printf("Skipping admin repositories (project creation required)")
		return nil // Don't fail the whole process
	}

	// Create repositories in the admin project
	for _, repo := range b.seedData.AdminUser.Repositories {
		if err := b.createBitbucketServerRepository(&repo, adminProject.Key); err != nil {
			return fmt.Errorf("failed to create admin repository %s: %w", repo.Name, err)
		}
	}
	return nil
}

func (b *BitbucketSeeder) SeedAll() error {
	log.Println("Starting Bitbucket seeding process...")

	var errors []string

	if err := b.CreateWorkspaces(); err != nil {
		log.Printf("Failed to create workspaces: %v", err)
		errors = append(errors, fmt.Sprintf("workspaces: %v", err))
	}

	if err := b.CreateUsers(); err != nil {
		log.Printf("Failed to create users: %v", err)
		errors = append(errors, fmt.Sprintf("users: %v", err))
	}

	if err := b.CreateAdminUserRepositories(); err != nil {
		log.Printf("Failed to create admin repositories: %v", err)
		errors = append(errors, fmt.Sprintf("admin repositories: %v", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("seeding completed with errors: %s", strings.Join(errors, "; "))
	}

	log.Println("Bitbucket seeding completed successfully!")
	return nil
}

// createBitbucketServerProject creates a project using Bitbucket Server REST API
func (b *BitbucketSeeder) createBitbucketServerProject(project *Project) error {
	// Check if project already exists
	checkURL := fmt.Sprintf("%s/rest/api/1.0/projects/%s", b.baseURL, project.Key)
	req, err := http.NewRequest("GET", checkURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.SetBasicAuth(b.username, b.password)

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to check project existence: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.Printf("Project %s already exists, skipping creation", project.Key)
		return nil
	}

	// Create project payload
	payload := map[string]interface{}{
		"key":         project.Key,
		"name":        project.Name,
		"description": project.Description,
		"public":      true, // Make project public for testing
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal project data: %w", err)
	}

	// Create project
	createURL := fmt.Sprintf("%s/rest/api/1.0/projects", b.baseURL)
	req, err = http.NewRequest("POST", createURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.SetBasicAuth(b.username, b.password)
	req.Header.Set("Content-Type", "application/json")

	resp, err = b.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 201 {
		return fmt.Errorf("failed to create project %s: status %d, body: %s", project.Key, resp.StatusCode, string(body))
	}

	log.Printf("Successfully created project: %s (Key: %s)", project.Name, project.Key)
	return nil
}

// createBitbucketServerRepository creates a repository using Bitbucket Server REST API
func (b *BitbucketSeeder) createBitbucketServerRepository(repo *Repository, projectKey string) error {
	// Check if repository already exists
	checkURL := fmt.Sprintf("%s/rest/api/1.0/projects/%s/repos/%s", b.baseURL, projectKey, repo.Name)
	req, err := http.NewRequest("GET", checkURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.SetBasicAuth(b.username, b.password)

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to check repository existence: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.Printf("Repository %s already exists in project %s, skipping creation", repo.Name, projectKey)
		return nil
	}

	// Create repository payload
	payload := map[string]interface{}{
		"name":        repo.Name,
		"description": repo.Description,
		"public":      !repo.IsPrivate, // Convert private to public
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal repository data: %w", err)
	}

	// Create repository
	createURL := fmt.Sprintf("%s/rest/api/1.0/projects/%s/repos", b.baseURL, projectKey)
	req, err = http.NewRequest("POST", createURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.SetBasicAuth(b.username, b.password)
	req.Header.Set("Content-Type", "application/json")

	resp, err = b.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 201 {
		return fmt.Errorf("failed to create repository %s in project %s: status %d, body: %s", repo.Name, projectKey, resp.StatusCode, string(body))
	}

	log.Printf("Successfully created repository: %s in project %s", repo.Name, projectKey)

	// ALWAYS initialize repositories with git content (needed for cloning)
	// Empty repositories cannot be cloned and will cause HTTP 500 errors
	log.Printf("🔧 CRITICAL: Initializing repository %s with git content...", repo.Name)

	// Skip API attempt and go straight to git commands for reliability
	gitignoreContent := `# Logs
*.log

# Runtime data
pids
*.pid
*.seed

# Coverage directory used by tools like istanbul
coverage/

# Dependency directories
node_modules/

# Optional npm cache directory
.npm

# Optional REPL history
.node_repl_history

# OS generated files
.DS_Store
Thumbs.db
`

	if err := b.initializeRepositoryWithGitCommands(projectKey, repo.Name, gitignoreContent); err != nil {
		return fmt.Errorf("CRITICAL: Failed to initialize repository %s with git content: %w", repo.Name, err)
	}
	log.Printf("✅ Repository %s successfully initialized with git content", repo.Name)

	// Server delay removed for faster processing

	return nil
}

// initializeBitbucketServerRepositoryWithGitignore initializes a repository with a .gitignore file
func (b *BitbucketSeeder) initializeBitbucketServerRepositoryWithGitignore(projectKey, repoName string) error {
	// Create basic .gitignore content
	gitignoreContent := `# Logs
*.log

# Runtime data
pids
*.pid
*.seed

# Coverage directory used by tools like istanbul
coverage/

# Dependency directories
node_modules/

# Optional npm cache directory
.npm

# Optional REPL history
.node_repl_history

# OS generated files
.DS_Store
Thumbs.db
`

	// Try to create the .gitignore file using Bitbucket Server's file API
	// Use the edit endpoint which is more widely supported
	createURL := fmt.Sprintf("%s/rest/api/1.0/projects/%s/repos/%s/browse/.gitignore", b.baseURL, projectKey, repoName)

	// Create form data for file creation
	payload := map[string]interface{}{
		"content":        gitignoreContent,
		"message":        "Initial commit: Add .gitignore",
		"sourceCommitId": "", // Empty for new file
		"targetBranch":   "master",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal .gitignore data: %w", err)
	}

	req, err := http.NewRequest("PUT", createURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.SetBasicAuth(b.username, b.password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		log.Printf("Successfully initialized repository %s with .gitignore", repoName)
		return nil
	}

	// If the browse API doesn't work, initialize with git commands
	body, _ := io.ReadAll(resp.Body)
	log.Printf(".gitignore creation via API failed for %s (status %d): %s", repoName, resp.StatusCode, string(body))

	// CRITICAL FIX: Empty repositories cannot be cloned! Must have at least one commit.
	log.Printf("Initializing repository %s with git commands...", repoName)
	return b.initializeRepositoryWithGitCommands(projectKey, repoName, gitignoreContent)
}

// initializeRepositoryWithGitCommands initializes a repository using git commands
func (b *BitbucketSeeder) initializeRepositoryWithGitCommands(projectKey, repoName, content string) error {
	// Create temporary directory
	tempDir := fmt.Sprintf("/tmp/bitbucket-init-%s-%s", projectKey, repoName)
	defer os.RemoveAll(tempDir) // Clean up

	// Repository clone URL with credentials - fix hostname extraction
	baseHost := strings.TrimPrefix(strings.TrimPrefix(b.baseURL, "http://"), "https://")
	cloneURL := fmt.Sprintf("http://%s:%s@%s/scm/%s/%s.git",
		b.username, b.password, baseHost,
		strings.ToLower(projectKey), repoName)

	log.Printf("🔧 Initializing repository via git commands: %s", cloneURL)

	// CRITICAL: Wait a moment for Bitbucket to fully initialize the repository
	log.Printf("	⏳ Bitbucket Server repository initialization...")
	// Removed wait - repository should be ready immediately

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Initialize git repo locally first (don't try to clone empty repo)
	if err := runGitCommand(tempDir, "init"); err != nil {
		return fmt.Errorf("git init failed: %w", err)
	}

	// Set default branch to master immediately for Bitbucket Server compatibility
	if err := runGitCommand(tempDir, "checkout", "-b", "master"); err != nil {
		return fmt.Errorf("git checkout master failed: %w", err)
	}

	// Create .gitignore file
	gitignorePath := fmt.Sprintf("%s/.gitignore", tempDir)
	if err := os.WriteFile(gitignorePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}

	// Add and commit
	if err := runGitCommand(tempDir, "add", ".gitignore"); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	if err := runGitCommand(tempDir, "commit", "-m", "Initial commit: Add .gitignore"); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	// Add remote origin
	if err := runGitCommand(tempDir, "remote", "add", "origin", cloneURL); err != nil {
		return fmt.Errorf("git remote add failed: %w", err)
	}

	// Push with more specific options and retry logic
	log.Printf("🔄 Pushing initial commit to Bitbucket Server...")

	// Try push with retries (sometimes Bitbucket needs a moment)
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			log.Printf("⏳ Retry %d/%d: Waiting 2 seconds before push...", i, maxRetries-1)
			time.Sleep(2 * time.Second)
		}

		err := runGitCommand(tempDir, "push", "-u", "origin", "master")
		if err == nil {
			log.Printf("✅ Successfully pushed initial commit for repository %s", repoName)
			return nil
		}

		log.Printf("⚠️ Push attempt %d failed: %v", i+1, err)
		if i == maxRetries-1 {
			return fmt.Errorf("git push failed after %d attempts: %w", maxRetries, err)
		}
	}

	return nil
}

// runGitCommand executes a git command in the specified directory
func runGitCommand(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Bitbucket Seeder",
		"GIT_AUTHOR_EMAIL=seeder@bitbucket.local",
		"GIT_COMMITTER_NAME=Bitbucket Seeder",
		"GIT_COMMITTER_EMAIL=seeder@bitbucket.local")

	log.Printf("🔨 Running git command: git %s (in %s)", strings.Join(args, " "), dir)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ Git command failed: git %s", strings.Join(args, " "))
		log.Printf("❌ Error: %v", err)
		log.Printf("❌ Output: %s", string(output))
		return fmt.Errorf("command 'git %v' failed: %w, output: %s", args, err, string(output))
	}

	if len(output) > 0 {
		log.Printf("✅ Git command output: %s", string(output))
	}

	return nil
}

func main() {
	var (
		username   = flag.String("username", "admin", "Bitbucket admin username")
		password   = flag.String("password", "admin", "Bitbucket admin password")
		baseURL    = flag.String("base-url", "http://bitbucket.example.com:7990", "Bitbucket Server base URL")
		configPath = flag.String("config", "configs/seed-data.json", "Path to seed data configuration file")
	)
	flag.Parse()

	if *username == "" || *password == "" {
		log.Fatal("Username and password are required")
	}

	seeder, err := NewBitbucketSeeder(*username, *password, *baseURL)
	if err != nil {
		log.Fatalf("Failed to create seeder: %v", err)
	}

	if err := seeder.LoadSeedData(*configPath); err != nil {
		log.Fatalf("Failed to load seed data: %v", err)
	}

	if err := seeder.SeedAll(); err != nil {
		log.Printf("Seeding completed with some errors: %v", err)
		// Don't exit with error code, as partial success is still useful
	} else {
		log.Println("Seeding completed successfully!")
	}
}
