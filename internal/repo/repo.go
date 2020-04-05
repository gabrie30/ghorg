// Package repo holds data for a repo
package repo

// Data represents an SCM repo
type Data struct {
	Name     string
	Path     string
	URL      string
	CloneURL string
}
