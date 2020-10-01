package base

type Client interface {
	GetUserRepos(targetUsername string) ([]Repo, error)
	GetOrgRepos(targetOrg string) ([]Repo, error)

	GetType() string
}
