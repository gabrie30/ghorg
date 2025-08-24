package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type Snippet struct {
	Title       string `json:"title"`
	FileName    string `json:"file_name"`
	Content     string `json:"content"`
	Description string `json:"description,omitempty"`
	Visibility  string `json:"visibility"`
}

type Repository struct {
	Name                 string    `json:"name"`
	InitializeWithReadme bool      `json:"initialize_with_readme"`
	Snippets             []Snippet `json:"snippets,omitempty"`
}

type Group struct {
	Name         string       `json:"name"`
	Path         string       `json:"path"`
	Description  string       `json:"description"`
	Repositories []Repository `json:"repositories,omitempty"`
	Subgroups    []Group      `json:"subgroups,omitempty"`
}

type User struct {
	Username     string       `json:"username"`
	Email        string       `json:"email"`
	Password     string       `json:"password"`
	Name         string       `json:"name"`
	Repositories []Repository `json:"repositories,omitempty"`
}

type RootUser struct {
	Repositories []Repository `json:"repositories"`
}

type SeedData struct {
	Groups       []Group   `json:"groups"`
	Users        []User    `json:"users"`
	RootUser     RootUser  `json:"root_user"`
	RootSnippets []Snippet `json:"root_snippets"`
}

type GitLabSeeder struct {
	client   *gitlab.Client
	seedData *SeedData
	baseURL  string
}

func NewGitLabSeeder(token, baseURL string) (*GitLabSeeder, error) {
	client, err := gitlab.NewClient(token, gitlab.WithBaseURL(baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	return &GitLabSeeder{
		client:  client,
		baseURL: baseURL,
	}, nil
}

func (g *GitLabSeeder) LoadSeedData(configPath string) error {
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

func (g *GitLabSeeder) CreateGroups() error {
	log.Println("Creating groups...")

	for _, group := range g.seedData.Groups {
		if err := g.createGroup(&group, nil); err != nil {
			return fmt.Errorf("failed to create group %s: %w", group.Name, err)
		}
	}
	return nil
}

func (g *GitLabSeeder) createGroup(group *Group, parentID *int) error {
	log.Printf("Creating group: %s", group.Name)

	createOptions := &gitlab.CreateGroupOptions{
		Name:        &group.Name,
		Path:        &group.Path,
		Description: &group.Description,
	}

	if parentID != nil {
		createOptions.ParentID = parentID
	}

	createdGroup, _, err := g.client.Groups.CreateGroup(createOptions)
	if err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}

	log.Printf("Created group: %s (ID: %d)", createdGroup.Name, createdGroup.ID)

	// Create repositories in this group
	for _, repo := range group.Repositories {
		if err := g.createRepository(&repo, &createdGroup.ID, group.Path); err != nil {
			return fmt.Errorf("failed to create repository %s in group %s: %w", repo.Name, group.Name, err)
		}
	}

	// Create subgroups recursively
	for _, subgroup := range group.Subgroups {
		if err := g.createGroup(&subgroup, &createdGroup.ID); err != nil {
			return fmt.Errorf("failed to create subgroup %s: %w", subgroup.Name, err)
		}
	}

	return nil
}

func (g *GitLabSeeder) createRepository(repo *Repository, namespaceID *int, groupPath string) error {
	log.Printf("Creating repository: %s", repo.Name)

	createOptions := &gitlab.CreateProjectOptions{
		Name:                 &repo.Name,
		InitializeWithReadme: gitlab.Ptr(repo.InitializeWithReadme),
	}

	if namespaceID != nil {
		createOptions.NamespaceID = namespaceID
	}

	project, _, err := g.client.Projects.CreateProject(createOptions)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	log.Printf("Created repository: %s (ID: %d)", project.Name, project.ID)

	// Create snippets for this repository
	for _, snippet := range repo.Snippets {
		if err := g.createProjectSnippet(&snippet, project.ID, groupPath, repo.Name); err != nil {
			return fmt.Errorf("failed to create snippet for repository %s: %w", repo.Name, err)
		}
	}

	return nil
}

func (g *GitLabSeeder) createProjectSnippet(snippet *Snippet, projectID int, groupPath, repoName string) error {
	log.Printf("Creating project snippet: %s for project %d", snippet.Title, projectID)

	visibility := gitlab.PublicVisibility
	switch snippet.Visibility {
	case "private":
		visibility = gitlab.PrivateVisibility
	case "internal":
		visibility = gitlab.InternalVisibility
	}

	createOptions := &gitlab.CreateProjectSnippetOptions{
		Title:       &snippet.Title,
		FileName:    &snippet.FileName,
		Content:     &snippet.Content,
		Visibility:  &visibility,
		Description: &snippet.Description,
	}

	_, _, err := g.client.ProjectSnippets.CreateSnippet(projectID, createOptions)
	if err != nil {
		return fmt.Errorf("failed to create project snippet: %w", err)
	}

	log.Printf("Created project snippet: %s", snippet.Title)
	return nil
}

func (g *GitLabSeeder) CreateUsers() error {
	log.Println("Creating users...")

	for _, user := range g.seedData.Users {
		if err := g.createUser(&user); err != nil {
			return fmt.Errorf("failed to create user %s: %w", user.Username, err)
		}
	}
	return nil
}

func (g *GitLabSeeder) createUser(user *User) error {
	log.Printf("Creating user: %s", user.Username)

	createOptions := &gitlab.CreateUserOptions{
		Username: &user.Username,
		Email:    &user.Email,
		Password: &user.Password,
		Name:     &user.Name,
	}

	createdUser, _, err := g.client.Users.CreateUser(createOptions)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	log.Printf("Created user: %s (ID: %d)", createdUser.Username, createdUser.ID)

	// Create repositories for this user
	for _, repo := range user.Repositories {
		if err := g.createUserRepository(&repo, createdUser.ID, user.Username); err != nil {
			return fmt.Errorf("failed to create repository %s for user %s: %w", repo.Name, user.Username, err)
		}
	}

	return nil
}

func (g *GitLabSeeder) createUserRepository(repo *Repository, userID int, username string) error {
	log.Printf("Creating user repository: %s for user %s", repo.Name, username)

	// Create project for user using the correct API format
	createOptions := &gitlab.CreateProjectOptions{
		Name:                 &repo.Name,
		InitializeWithReadme: gitlab.Ptr(repo.InitializeWithReadme),
	}

	// We need to get the user's namespace first
	user, _, err := g.client.Users.GetUser(userID, gitlab.GetUsersOptions{})
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Find the user's personal namespace
	namespaces, _, err := g.client.Namespaces.ListNamespaces(&gitlab.ListNamespacesOptions{})
	if err != nil {
		return fmt.Errorf("failed to list namespaces: %w", err)
	}

	var userNamespaceID *int
	for _, ns := range namespaces {
		if ns.Kind == "user" && ns.Path == user.Username {
			userNamespaceID = &ns.ID
			break
		}
	}

	if userNamespaceID == nil {
		return fmt.Errorf("could not find user namespace for user %s", username)
	}

	createOptions.NamespaceID = userNamespaceID

	project, _, err := g.client.Projects.CreateProject(createOptions)
	if err != nil {
		return fmt.Errorf("failed to create user repository: %w", err)
	}

	log.Printf("Created user repository: %s (ID: %d)", project.Name, project.ID)

	// Create snippets for this repository
	for _, snippet := range repo.Snippets {
		if err := g.createProjectSnippet(&snippet, project.ID, username, repo.Name); err != nil {
			return fmt.Errorf("failed to create snippet for user repository %s: %w", repo.Name, err)
		}
	}

	return nil
}

func (g *GitLabSeeder) CreateRootUserRepositories() error {
	log.Println("Creating root user repositories...")

	for _, repo := range g.seedData.RootUser.Repositories {
		if err := g.createRepository(&repo, nil, "root"); err != nil {
			return fmt.Errorf("failed to create root repository %s: %w", repo.Name, err)
		}
	}
	return nil
}

func (g *GitLabSeeder) CreateRootSnippets() error {
	log.Println("Creating root-level snippets...")

	for _, snippet := range g.seedData.RootSnippets {
		if err := g.createRootSnippet(&snippet); err != nil {
			return fmt.Errorf("failed to create root snippet %s: %w", snippet.Title, err)
		}
	}
	return nil
}

func (g *GitLabSeeder) createRootSnippet(snippet *Snippet) error {
	log.Printf("Creating root snippet: %s", snippet.Title)

	visibility := gitlab.PublicVisibility
	switch snippet.Visibility {
	case "private":
		visibility = gitlab.PrivateVisibility
	case "internal":
		visibility = gitlab.InternalVisibility
	}

	createOptions := &gitlab.CreateSnippetOptions{
		Title:       &snippet.Title,
		FileName:    &snippet.FileName,
		Content:     &snippet.Content,
		Visibility:  &visibility,
		Description: &snippet.Description,
	}

	_, _, err := g.client.Snippets.CreateSnippet(createOptions)
	if err != nil {
		return fmt.Errorf("failed to create root snippet: %w", err)
	}

	log.Printf("Created root snippet: %s", snippet.Title)
	return nil
}

func (g *GitLabSeeder) SeedAll() error {
	log.Println("Starting GitLab seeding process...")

	if err := g.CreateGroups(); err != nil {
		return err
	}

	if err := g.CreateUsers(); err != nil {
		return err
	}

	if err := g.CreateRootUserRepositories(); err != nil {
		return err
	}

	if err := g.CreateRootSnippets(); err != nil {
		return err
	}

	log.Println("GitLab seeding completed successfully!")
	return nil
}

func main() {
	var (
		token      = flag.String("token", "", "GitLab API token")
		baseURL    = flag.String("base-url", "http://gitlab.example.com", "GitLab base URL")
		configPath = flag.String("config", "configs/seed-data.json", "Path to seed data configuration file")
	)
	flag.Parse()

	if *token == "" {
		log.Fatal("Token is required")
	}

	seeder, err := NewGitLabSeeder(*token, *baseURL)
	if err != nil {
		log.Fatalf("Failed to create seeder: %v", err)
	}

	if err := seeder.LoadSeedData(*configPath); err != nil {
		log.Fatalf("Failed to load seed data: %v", err)
	}

	if err := seeder.SeedAll(); err != nil {
		log.Fatalf("Failed to seed GitLab: %v", err)
	}
}
