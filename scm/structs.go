package scm

// Repo represents an SCM repo, should probably be renamed to "cloneable" since we clone wikis and snippets with this
type Repo struct {
	// The ID of the repo that is assigned via the SCM provider. This is used for example with gitlab snippets on cloud gropus where we need to know the repo id to look up all he snippets it has.
	ID string
	// Name is the name of the repo https://www.github.com/gabrie30/ghorg.git the Name would be ghorg
	Name string
	// HostPath is the path on the users machine that the repo will be cloned to. Its used in all the git commands to locate the directory of the repo. HostPath is updated for wikis and snippets because the folder for the clone is appended with .wiki and .snippet
	HostPath string
	// Path where the repo is located within the scm provider. Its mostly used with gitlab repos when the directory structure is preserved. In this case the path becomes where to locate the repo in relation to gitlab.com/group/group/group/repo.git => /group/group/group/repo
	Path string
	// URL is the web address of the repo
	URL string
	// CloneURL is the url for cloning the repo, will be different for ssh vs http clones and will have the .git extention
	CloneURL string
	// CloneBranch the branch to clone. This will be the default branch if not specified. It will always be main for snippets.
	CloneBranch string
	// IsWiki is set to true when the data is for a wiki page
	IsWiki bool
	// IsGitLabSnippet is set to true when the data is for a gitlab snippet
	IsGitLabSnippet bool
	// IsGitLabRootLevelSnippet is set to true when a snippet was not created for a repo
	IsGitLabRootLevelSnippet bool
	// GitLabSnippetInfo provides additional information when the thing we are cloning is a gitlab snippet
	GitLabSnippetInfo GitLabSnippet
	Commits           RepoCommits
}

type RepoCommits struct {
	CountPrePull  int
	CountPostPull int
	CountDiff     int
}

type GitLabSnippet struct {
	// GitLab ID of the snippet
	ID string
	// Title of the snippet
	Title string
	// URL of the repo that snippet was made on
	URLOfRepo string
	// Name of the repo that the snippet was made on
	NameOfRepo string
}
