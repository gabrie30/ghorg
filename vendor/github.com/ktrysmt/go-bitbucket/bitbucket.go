package bitbucket

type users interface {
	Get(username string) (interface{}, error)
	Followers(username string) (interface{}, error)
	Following(username string) (interface{}, error)
	Repositories(username string) (interface{}, error)
}

type user interface {
	Profile() (*User, error)
	Emails() (interface{}, error)
}

type pullrequests interface {
	Create(opt PullRequestsOptions) (interface{}, error)
	Update(opt PullRequestsOptions) (interface{}, error)
	List(opt PullRequestsOptions) (interface{}, error)
	Get(opt PullRequestsOptions) (interface{}, error)
	Activities(opt PullRequestsOptions) (interface{}, error)
	Activity(opt PullRequestsOptions) (interface{}, error)
	Commits(opt PullRequestsOptions) (interface{}, error)
	Patch(opt PullRequestsOptions) (interface{}, error)
	Diff(opt PullRequestsOptions) (interface{}, error)
	Merge(opt PullRequestsOptions) (interface{}, error)
	Decline(opt PullRequestsOptions) (interface{}, error)
}

type repository interface {
	Get(opt RepositoryOptions) (*Repository, error)
	Create(opt RepositoryOptions) (*Repository, error)
	Delete(opt RepositoryOptions) (interface{}, error)
	ListWatchers(opt RepositoryOptions) (interface{}, error)
	ListForks(opt RepositoryOptions) (interface{}, error)
	UpdatePipelineConfig(opt RepositoryPipelineOptions) (*Pipeline, error)
	AddPipelineVariable(opt RepositoryPipelineVariableOptions) (*PipelineVariable, error)
	AddPipelineKeyPair(opt RepositoryPipelineKeyPairOptions) (*PipelineKeyPair, error)
	UpdatePipelineBuildNumber(opt RepositoryPipelineBuildNumberOptions) (*PipelineBuildNumber, error)
	ListFiles(opt RepositoryFilesOptions) (*[]RepositoryFile, error)
	GetFileBlob(opt RepositoryBlobOptions) (*RepositoryBlob, error)
	ListBranches(opt RepositoryBranchOptions) (*RepositoryBranches, error)
	BranchingModel(opt RepositoryBranchingModelOptions) (*BranchingModel, error)
}

type repositories interface {
	ListForAccount(opt RepositoriesOptions) (interface{}, error)
	ListForTeam(opt RepositoriesOptions) (interface{}, error)
	ListPublic() (interface{}, error)
}

type commits interface {
	GetCommits(opt CommitsOptions) (interface{}, error)
	GetCommit(opt CommitsOptions) (interface{}, error)
	GetCommitComments(opt CommitsOptions) (interface{}, error)
	GetCommitComment(opt CommitsOptions) (interface{}, error)
	GetCommitStatus(opt CommitsOptions) (interface{}, error)
	GiveApprove(opt CommitsOptions) (interface{}, error)
	RemoveApprove(opt CommitsOptions) (interface{}, error)
	CreateCommitStatus(cmo CommitsOptions, cso CommitStatusOptions) (interface{}, error)
}

type branchrestrictions interface {
	Gets(opt BranchRestrictionsOptions) (interface{}, error)
	Get(opt BranchRestrictionsOptions) (interface{}, error)
	Create(opt BranchRestrictionsOptions) (interface{}, error)
	Update(opt BranchRestrictionsOptions) (interface{}, error)
	Delete(opt BranchRestrictionsOptions) (interface{}, error)
}

type diff interface {
	GetDiff(opt DiffOptions) (interface{}, error)
	GetPatch(opt DiffOptions) (interface{}, error)
}

type webhooks interface {
	Gets(opt WebhooksOptions) (interface{}, error)
	Get(opt WebhooksOptions) (interface{}, error)
	Create(opt WebhooksOptions) (interface{}, error)
	Update(opt WebhooksOptions) (interface{}, error)
	Delete(opt WebhooksOptions) (interface{}, error)
}

type teams interface {
	List(role string) (interface{}, error) // [WIP?] role=[admin|contributor|member]
	Profile(teamname string) (interface{}, error)
	Members(teamname string) (interface{}, error)
	Followers(teamname string) (interface{}, error)
	Following(teamname string) (interface{}, error)
	Repositories(teamname string) (interface{}, error)
	Projects(teamname string) (interface{}, error)
}

type RepositoriesOptions struct {
	Owner string `json:"owner"`
	Role  string `json:"role"` // role=[owner|admin|contributor|member]
}

type RepositoryOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	Scm      string `json:"scm"`
	//	Name        string `json:"name"`
	IsPrivate   string `json:"is_private"`
	Description string `json:"description"`
	ForkPolicy  string `json:"fork_policy"`
	Language    string `json:"language"`
	HasIssues   string `json:"has_issues"`
	HasWiki     string `json:"has_wiki"`
	Project     string `json:"project"`
}

type RepositoryFilesOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	Ref      string `json:"ref"`
	Path     string `json:"path"`
}

type RepositoryBlobOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	Ref      string `json:"ref"`
	Path     string `json:"path"`
}

type RepositoryBranchOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	Query    string `json:"q"`
	Sort     string `json:"sort"`
	PageNum  int    `json:"page"`
	Pagelen  int    `json:"pagelen"`
	MaxDepth int    `json:"max_depth"`
}

type RepositoryTagOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	Query    string `json:"q"`
	Sort     string `json:"sort"`
	PageNum  int    `json:"page"`
	Pagelen  int    `json:"pagelen"`
	MaxDepth int    `json:"max_depth"`
}

type PullRequestsOptions struct {
	ID                string   `json:"id"`
	CommentID         string   `json:"comment_id"`
	Owner             string   `json:"owner"`
	RepoSlug          string   `json:"repo_slug"`
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	CloseSourceBranch bool     `json:"close_source_branch"`
	SourceBranch      string   `json:"source_branch"`
	SourceRepository  string   `json:"source_repository"`
	DestinationBranch string   `json:"destination_branch"`
	DestinationCommit string   `json:"destination_repository"`
	Message           string   `json:"message"`
	Reviewers         []string `json:"reviewers"`
	States            []string `json:"states"`
	Query             string   `json:"query"`
	Sort              string   `json:"sort"`
}

type CommitsOptions struct {
	Owner       string `json:"owner"`
	RepoSlug    string `json:"repo_slug"`
	Revision    string `json:"revision"`
	Branchortag string `json:"branchortag"`
	Include     string `json:"include"`
	Exclude     string `json:"exclude"`
	CommentID   string `json:"comment_id"`
}

type CommitStatusOptions struct {
	Key         string `json:"key"`
	Url         string `json:"url"`
	State       string `json:"state"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type BranchRestrictionsOptions struct {
	Owner    string            `json:"owner"`
	RepoSlug string            `json:"repo_slug"`
	ID       string            `json:"id"`
	Groups   map[string]string `json:"groups"`
	Pattern  string            `json:"pattern"`
	Users    []string          `json:"users"`
	Kind     string            `json:"kind"`
	FullSlug string            `json:"full_slug"`
	Name     string            `json:"name"`
	Value    interface{}       `json:"value"`
}

type DiffOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	Spec     string `json:"spec"`
}

type WebhooksOptions struct {
	Owner       string   `json:"owner"`
	RepoSlug    string   `json:"repo_slug"`
	Uuid        string   `json:"uuid"`
	Description string   `json:"description"`
	Url         string   `json:"url"`
	Active      bool     `json:"active"`
	Events      []string `json:"events"` // EX) {'repo:push','issue:created',..} REF) https://goo.gl/VTj93b
}

type RepositoryPipelineOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	Enabled  bool   `json:"has_pipelines"`
}

type RepositoryPipelineVariableOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	Uuid     string `json:"uuid"`
	Key      string `json:"key"`
	Value    string `json:"value"`
	Secured  bool   `json:"secured"`
}

type RepositoryPipelineKeyPairOptions struct {
	Owner      string `json:"owner"`
	RepoSlug   string `json:"repo_slug"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

type RepositoryPipelineBuildNumberOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	Next     int    `json:"next"`
}

type RepositoryBranchingModelOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
}

type DownloadsOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	FilePath string `json:"filepath"`
	FileName string `json:"filename"`
}

type PageRes struct {
	Page     int32 `json:"page"`
	PageLen  int32 `json:"pagelen"`
	MaxDepth int32 `json:"max_depth"`
	Size     int32 `json:"size"`
}
