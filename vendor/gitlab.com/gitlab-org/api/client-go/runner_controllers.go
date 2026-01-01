package gitlab

import (
	"net/http"
	"time"
)

type (
	// RunnerControllersServiceInterface handles communication with the runner
	// controller related methods of the GitLab API. This is an admin-only endpoint.
	//
	// GitLab API docs: Documentation not yet available, see
	// https://gitlab.com/gitlab-org/gitlab/-/issues/581275
	RunnerControllersServiceInterface interface {
		// ListRunnerControllers gets a list of runner controllers. This is an
		// admin-only endpoint.
		//
		// GitLab API docs: Documentation not yet available, see
		// https://gitlab.com/gitlab-org/gitlab/-/issues/581275
		ListRunnerControllers(opt *ListRunnerControllersOptions, options ...RequestOptionFunc) ([]*RunnerController, *Response, error)
		// GetRunnerController gets a single runner controller. This is an admin-only
		// endpoint.
		//
		// GitLab API docs: Documentation not yet available, see
		// https://gitlab.com/gitlab-org/gitlab/-/issues/581275
		GetRunnerController(rid int64, options ...RequestOptionFunc) (*RunnerController, *Response, error)
		// CreateRunnerController creates a new runner controller. This is an
		// admin-only endpoint.
		//
		// GitLab API docs: Documentation not yet available, see
		// https://gitlab.com/gitlab-org/gitlab/-/issues/581275
		CreateRunnerController(opt *CreateRunnerControllerOptions, options ...RequestOptionFunc) (*RunnerController, *Response, error)
		// UpdateRunnerController updates a runner controller. This is an admin-only
		// endpoint.
		//
		// GitLab API docs: Documentation not yet available, see
		// https://gitlab.com/gitlab-org/gitlab/-/issues/581275
		UpdateRunnerController(rid int64, opt *UpdateRunnerControllerOptions, options ...RequestOptionFunc) (*RunnerController, *Response, error)
		// DeleteRunnerController deletes a runner controller. This is an admin-only
		// endpoint.
		//
		// GitLab API docs: Documentation not yet available, see
		// https://gitlab.com/gitlab-org/gitlab/-/issues/581275
		DeleteRunnerController(rid int64, options ...RequestOptionFunc) (*Response, error)
	}

	// RunnerControllersService handles communication with the runner controller
	// related methods of the GitLab API. This is an admin-only endpoint.
	//
	// GitLab API docs: Documentation not yet available, see
	// https://gitlab.com/gitlab-org/gitlab/-/issues/581275
	RunnerControllersService struct {
		client *Client
	}
)

var _ RunnerControllersServiceInterface = (*RunnerControllersService)(nil)

// RunnerController represents a GitLab runner controller.
type RunnerController struct {
	ID          int64      `json:"id"`
	Description string     `json:"description"`
	Enabled     bool       `json:"enabled"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

// ListRunnerControllersOptions represents the available
// ListRunnerControllers() options.
//
// GitLab API docs: Documentation not yet available, see
// https://gitlab.com/gitlab-org/gitlab/-/issues/581275
type ListRunnerControllersOptions struct {
	ListOptions
}

func (s *RunnerControllersService) ListRunnerControllers(opt *ListRunnerControllersOptions, options ...RequestOptionFunc) ([]*RunnerController, *Response, error) {
	return do[[]*RunnerController](s.client,
		withPath("runner_controllers"),
		withAPIOpts(opt),
		withRequestOpts(options...),
	)
}

func (s *RunnerControllersService) GetRunnerController(rid int64, options ...RequestOptionFunc) (*RunnerController, *Response, error) {
	return do[*RunnerController](s.client,
		withPath("runner_controllers/%d", rid),
		withRequestOpts(options...),
	)
}

// CreateRunnerControllerOptions represents the available
// CreateRunnerController() options.
//
// GitLab API docs: Documentation not yet available, see
// https://gitlab.com/gitlab-org/gitlab/-/issues/581275
type CreateRunnerControllerOptions struct {
	Description *string `url:"description,omitempty" json:"description,omitempty"`
	Enabled     *bool   `url:"enabled,omitempty" json:"enabled,omitempty"`
}

func (s *RunnerControllersService) CreateRunnerController(opt *CreateRunnerControllerOptions, options ...RequestOptionFunc) (*RunnerController, *Response, error) {
	return do[*RunnerController](s.client,
		withMethod(http.MethodPost),
		withPath("runner_controllers"),
		withAPIOpts(opt),
		withRequestOpts(options...),
	)
}

// UpdateRunnerControllerOptions represents the available
// UpdateRunnerController() options.
//
// GitLab API docs: Documentation not yet available, see
// https://gitlab.com/gitlab-org/gitlab/-/issues/581275
type UpdateRunnerControllerOptions struct {
	Description *string `url:"description,omitempty" json:"description,omitempty"`
	Enabled     *bool   `url:"enabled,omitempty" json:"enabled,omitempty"`
}

func (s *RunnerControllersService) UpdateRunnerController(rid int64, opt *UpdateRunnerControllerOptions, options ...RequestOptionFunc) (*RunnerController, *Response, error) {
	return do[*RunnerController](s.client,
		withMethod(http.MethodPut),
		withPath("runner_controllers/%d", rid),
		withAPIOpts(opt),
		withRequestOpts(options...),
	)
}

func (s *RunnerControllersService) DeleteRunnerController(rid int64, options ...RequestOptionFunc) (*Response, error) {
	_, resp, err := do[none](s.client,
		withMethod(http.MethodDelete),
		withPath("runner_controllers/%d", rid),
		withRequestOpts(options...),
	)
	return resp, err
}
