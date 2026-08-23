package scm

// Repo represents an SCM repo, should probably be renamed to "cloneable" since we clone wikis and snippets with this
type Repo struct {
	// The ID of the repo that is assigned via the SCM provider. This is used for example with gitlab snippets on cloud gropus where we need to know the repo id to look up all he snippets it has.
	ID string `json:"id"`
	// Name is the name of the repo https://www.github.com/gabrie30/ghorg.git the Name would be ghorg
	Name string `json:"name"`
	// HostPath is the path on the users machine that the repo will be cloned to. Its used in all the git commands to locate the directory of the repo. HostPath is updated for wikis and snippets because the folder for the clone is appended with .wiki and .snippet
	HostPath string `json:"host_path"`
	// Path where the repo is located within the scm provider. Its mostly used with gitlab repos when the directory structure is preserved. In this case the path becomes where to locate the repo in relation to gitlab.com/group/group/group/repo.git => /group/group/group/repo
	Path string `json:"path"`
	// URL is the web address of the repo
	URL string `json:"url"`
	// CloneURL is the url for cloning the repo, will be different for ssh vs http clones and will have the .git extention
	CloneURL string `json:"clone_url"`
	// CloneBranch the branch to clone. This will be the default branch if not specified. It will always be main for snippets.
	CloneBranch string `json:"clone_branch"`
	// IsWiki is set to true when the data is for a wiki page
	IsWiki bool `json:"is_wiki"`
	// IsGitLabSnippet is set to true when the data is for a gitlab snippet
	IsGitLabSnippet bool `json:"is_gitlab_snippet"`
	// IsGitLabRootLevelSnippet is set to true when a snippet was not created for a repo
	IsGitLabRootLevelSnippet bool `json:"is_gitlab_root_level_snippet"`
	// IsGitHubGist is set to true when the data is for a github gist
	IsGitHubGist bool `json:"is_github_gist"`
	// GitLabSnippetInfo provides additional information when the thing we are cloning is a gitlab snippet
	GitLabSnippetInfo GitLabSnippet `json:"gitlab_snippet_info"`
	Commits           RepoCommits   `json:"commits"`
}

type RepoCommits struct {
	CountPrePull  int `json:"count_pre_pull"`
	CountPostPull int `json:"count_post_pull"`
	CountDiff     int `json:"count_diff"`
}

type GitLabSnippet struct {
	// GitLab ID of the snippet
	ID string `json:"id"`
	// Title of the snippet
	Title string `json:"title"`
	// URL of the repo that snippet was made on
	URLOfRepo string `json:"url_of_repo"`
	// Name of the repo that the snippet was made on
	NameOfRepo string `json:"name_of_repo"`
}
