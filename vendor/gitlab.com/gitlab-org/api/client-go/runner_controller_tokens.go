package gitlab

import (
	"net/http"
	"time"
)

type (
	// RunnerControllerTokensServiceInterface handles communication with the runner
	// controller token related methods of the GitLab API. This is an admin-only
	// endpoint.
	//
	// GitLab API docs: Documentation not yet available, see
	// https://gitlab.com/gitlab-org/gitlab/-/issues/581275
	RunnerControllerTokensServiceInterface interface {
		// ListRunnerControllerTokens gets a list of runner controller tokens. This is
		// an admin-only endpoint.
		//
		// GitLab API docs: Documentation not yet available, see
		// https://gitlab.com/gitlab-org/gitlab/-/issues/581275
		ListRunnerControllerTokens(rid int64, opt *ListRunnerControllerTokensOptions, options ...RequestOptionFunc) ([]*RunnerControllerToken, *Response, error)
		// GetRunnerControllerToken gets a single runner controller token. This is an
		// admin-only endpoint.
		//
		// GitLab API docs: Documentation not yet available, see
		// https://gitlab.com/gitlab-org/gitlab/-/issues/581275
		GetRunnerControllerToken(rid int64, tokenID int64, options ...RequestOptionFunc) (*RunnerControllerToken, *Response, error)
		// CreateRunnerControllerToken creates a new runner controller token. This is
		// an admin-only endpoint.
		//
		// GitLab API docs: Documentation not yet available, see
		// https://gitlab.com/gitlab-org/gitlab/-/issues/581275
		CreateRunnerControllerToken(rid int64, opt *CreateRunnerControllerTokenOptions, options ...RequestOptionFunc) (*RunnerControllerToken, *Response, error)
		// RevokeRunnerControllerToken revokes a runner controller token. This is an
		// admin-only endpoint.
		//
		// GitLab API docs: Documentation not yet available, see
		// https://gitlab.com/gitlab-org/gitlab/-/issues/581275
		RevokeRunnerControllerToken(rid int64, tokenID int64, options ...RequestOptionFunc) (*Response, error)
	}

	// RunnerControllerTokensService handles communication with the runner
	// controller token related methods of the GitLab API. This is an admin-only
	// endpoint.
	//
	// GitLab API docs: Documentation not yet available, see
	// https://gitlab.com/gitlab-org/gitlab/-/issues/581275
	RunnerControllerTokensService struct {
		client *Client
	}
)

var _ RunnerControllerTokensServiceInterface = (*RunnerControllerTokensService)(nil)

// RunnerControllerToken represents a GitLab runner controller token.
type RunnerControllerToken struct {
	ID          int64      `json:"id"`
	Description string     `json:"description"`
	Token       string     `json:"token,omitempty"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

// ListRunnerControllerTokensOptions represents the available
// ListRunnerControllerTokens() options.
//
// GitLab API docs: Documentation not yet available, see
// https://gitlab.com/gitlab-org/gitlab/-/issues/581275
type ListRunnerControllerTokensOptions struct {
	ListOptions
}

func (s *RunnerControllerTokensService) ListRunnerControllerTokens(rid int64, opt *ListRunnerControllerTokensOptions, options ...RequestOptionFunc) ([]*RunnerControllerToken, *Response, error) {
	return do[[]*RunnerControllerToken](s.client,
		withPath("runner_controllers/%d/tokens", rid),
		withAPIOpts(opt),
		withRequestOpts(options...),
	)
}

func (s *RunnerControllerTokensService) GetRunnerControllerToken(rid int64, tokenID int64, options ...RequestOptionFunc) (*RunnerControllerToken, *Response, error) {
	return do[*RunnerControllerToken](s.client,
		withPath("runner_controllers/%d/tokens/%d", rid, tokenID),
		withRequestOpts(options...),
	)
}

// CreateRunnerControllerTokenOptions represents the available
// CreateRunnerControllerToken() options.
//
// GitLab API docs: Documentation not yet available, see
// https://gitlab.com/gitlab-org/gitlab/-/issues/581275
type CreateRunnerControllerTokenOptions struct {
	Description *string `url:"description,omitempty" json:"description,omitempty"`
}

func (s *RunnerControllerTokensService) CreateRunnerControllerToken(rid int64, opt *CreateRunnerControllerTokenOptions, options ...RequestOptionFunc) (*RunnerControllerToken, *Response, error) {
	return do[*RunnerControllerToken](s.client,
		withMethod(http.MethodPost),
		withPath("runner_controllers/%d/tokens", rid),
		withAPIOpts(opt),
		withRequestOpts(options...),
	)
}

func (s *RunnerControllerTokensService) RevokeRunnerControllerToken(rid int64, tokenID int64, options ...RequestOptionFunc) (*Response, error) {
	_, resp, err := do[none](s.client,
		withMethod(http.MethodDelete),
		withPath("runner_controllers/%d/tokens/%d", rid, tokenID),
		withRequestOpts(options...),
	)
	return resp, err
}
