package scm

// Repo represents an SCM repo
type Repo struct {
	Name        string
	HostPath    string
	Path        string
	URL         string
	CloneURL    string
	CloneBranch string
}
