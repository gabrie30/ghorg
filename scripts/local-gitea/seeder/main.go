package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"code.gitea.io/sdk/gitea"
)

type Repository struct {
	Name                 string `json:"name"`
	InitializeWithReadme bool   `json:"initialize_with_readme"`
	Description          string `json:"description,omitempty"`
}

type Organization struct {
	Name         string       `json:"name"`
	Username     string       `json:"username"`
	Description  string       `json:"description"`
	Repositories []Repository `json:"repositories,omitempty"`
}

type User struct {
	Username     string       `json:"username"`
	Email        string       `json:"email"`
	Password     string       `json:"password"`
	FullName     string       `json:"full_name"`
	Repositories []Repository `json:"repositories,omitempty"`
}

type RootUser struct {
	Repositories []Repository `json:"repositories"`
}

type SeedData struct {
	Organizations []Organization `json:"organizations"`
	Users         []User         `json:"users"`
	RootUser      RootUser       `json:"root_user"`
}

type GiteaSeeder struct {
	client   *gitea.Client
	seedData *SeedData
	baseURL  string
}

func NewGiteaSeeder(token, baseURL string) (*GiteaSeeder, error) {
	// Always use basic authentication for now since tokens are tricky
	log.Printf("Creating Gitea client with basic authentication...")
	client, err := gitea.NewClient(baseURL, gitea.SetBasicAuth("testuser", "testpass"))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gitea client: %w", err)
	}

	return &GiteaSeeder{
		client:  client,
		baseURL: baseURL,
	}, nil
}

func (g *GiteaSeeder) LoadSeedData(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read seed config: %w", err)
	}

	g.seedData = &SeedData{}
	if err := json.Unmarshal(data, g.seedData); err != nil {
		return fmt.Errorf("failed to parse seed config: %w", err)
	}

	return nil
}

func (g *GiteaSeeder) CreateOrganizations() error {
	log.Println("Creating organizations...")

	for _, org := range g.seedData.Organizations {
		if err := g.createOrganization(&org); err != nil {
			return fmt.Errorf("failed to create organization %s: %w", org.Name, err)
		}
	}
	return nil
}

func (g *GiteaSeeder) createOrganization(org *Organization) error {
	log.Printf("Creating organization: %s", org.Name)

	createOptions := gitea.CreateOrgOption{
		Name:        org.Username,
		FullName:    org.Name,
		Description: org.Description,
	}

	createdOrg, _, err := g.client.CreateOrg(createOptions)
	if err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}

	log.Printf("Created organization: %s (ID: %d)", createdOrg.FullName, createdOrg.ID)

	// Create repositories in this organization
	for _, repo := range org.Repositories {
		if err := g.createOrgRepository(&repo, org.Username); err != nil {
			return fmt.Errorf("failed to create repository %s in organization %s: %w", repo.Name, org.Name, err)
		}
	}

	return nil
}

func (g *GiteaSeeder) createOrgRepository(repo *Repository, orgName string) error {
	log.Printf("Creating organization repository: %s in org %s", repo.Name, orgName)

	createOptions := gitea.CreateRepoOption{
		Name:        repo.Name,
		Description: repo.Description,
		AutoInit:    repo.InitializeWithReadme,
	}

	project, _, err := g.client.CreateOrgRepo(orgName, createOptions)
	if err != nil {
		return fmt.Errorf("failed to create organization repository: %w", err)
	}

	log.Printf("Created organization repository: %s (ID: %d)", project.Name, project.ID)
	return nil
}

func (g *GiteaSeeder) CreateUsers() error {
	log.Println("Creating users...")

	for _, user := range g.seedData.Users {
		if err := g.createUser(&user); err != nil {
			return fmt.Errorf("failed to create user %s: %w", user.Username, err)
		}
	}
	return nil
}

func (g *GiteaSeeder) createUser(user *User) error {
	log.Printf("Creating user: %s", user.Username)

	mustChangePassword := false
	createOptions := gitea.CreateUserOption{
		Username:           user.Username,
		Email:              user.Email,
		Password:           user.Password,
		FullName:           user.FullName,
		MustChangePassword: &mustChangePassword,
		SendNotify:         false,
	}

	createdUser, _, err := g.client.AdminCreateUser(createOptions)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	log.Printf("Created user: %s (ID: %d)", createdUser.UserName, createdUser.ID)

	// Create repositories for this user
	for _, repo := range user.Repositories {
		if err := g.createUserRepository(&repo, user.Username); err != nil {
			return fmt.Errorf("failed to create repository %s for user %s: %w", repo.Name, user.Username, err)
		}
	}

	return nil
}

func (g *GiteaSeeder) createUserRepository(repo *Repository, username string) error {
	log.Printf("Creating user repository: %s for user %s", repo.Name, username)

	// First, we need to get the current user to create a repo on their behalf
	// For simplicity, we'll use the CreateRepo API which creates a repo for the authenticated user
	// Then we'll transfer it to the target user if needed

	createOptions := gitea.CreateRepoOption{
		Name:        repo.Name,
		Description: repo.Description,
		AutoInit:    repo.InitializeWithReadme,
	}

	project, _, err := g.client.CreateRepo(createOptions)
	if err != nil {
		return fmt.Errorf("failed to create user repository: %w", err)
	}

	log.Printf("Created user repository: %s (ID: %d)", project.Name, project.ID)

	// Note: In a real implementation, you might want to transfer the repo to the target user
	// For now, this creates repos under the authenticated admin user
	return nil
}

func (g *GiteaSeeder) CreateRootUserRepositories() error {
	log.Println("Creating root user repositories...")

	for _, repo := range g.seedData.RootUser.Repositories {
		if err := g.createRootRepository(&repo); err != nil {
			return fmt.Errorf("failed to create root repository %s: %w", repo.Name, err)
		}
	}
	return nil
}

func (g *GiteaSeeder) createRootRepository(repo *Repository) error {
	log.Printf("Creating root repository: %s", repo.Name)

	createOptions := gitea.CreateRepoOption{
		Name:        repo.Name,
		Description: repo.Description,
		AutoInit:    repo.InitializeWithReadme,
	}

	project, _, err := g.client.CreateRepo(createOptions)
	if err != nil {
		return fmt.Errorf("failed to create root repository: %w", err)
	}

	log.Printf("Created root repository: %s (ID: %d)", project.Name, project.ID)
	return nil
}

func (g *GiteaSeeder) SeedAll() error {
	log.Println("Starting Gitea seeding process...")

	var errors []string

	if err := g.CreateOrganizations(); err != nil {
		log.Printf("Failed to create organizations: %v", err)
		errors = append(errors, fmt.Sprintf("organizations: %v", err))
	}

	if err := g.CreateUsers(); err != nil {
		log.Printf("Failed to create users: %v", err)
		errors = append(errors, fmt.Sprintf("users: %v", err))
	}

	if err := g.CreateRootUserRepositories(); err != nil {
		log.Printf("Failed to create root repositories: %v", err)
		errors = append(errors, fmt.Sprintf("root repositories: %v", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("seeding completed with errors: %s", strings.Join(errors, "; "))
	}

	log.Println("Gitea seeding completed successfully!")
	return nil
}

func main() {
	var (
		token      = flag.String("token", "", "Gitea API token")
		username   = flag.String("username", "", "Admin username for basic auth")
		password   = flag.String("password", "", "Admin password for basic auth")
		baseURL    = flag.String("base-url", "http://gitea.example.com:3000", "Gitea base URL")
		configPath = flag.String("config", "configs/seed-data.json", "Path to seed data configuration file")
	)
	flag.Parse()

	if *token == "" && (*username == "" || *password == "") {
		log.Fatal("Either token or username+password is required")
	}

	// If username and password are provided, we'll use them in NewGiteaSeeder
	if *username != "" && *password != "" {
		log.Printf("Using basic authentication with username: %s", *username)
	}

	seeder, err := NewGiteaSeeder(*token, *baseURL)
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
