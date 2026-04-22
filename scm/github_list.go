package scm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/google/go-github/v72/github"
)

const (
	// maxGithubRepoListConcurrency caps GHORG_GITHUB_REPO_LIST_CONCURRENCY when set (typo guard).
	maxGithubRepoListConcurrency = 1024
	maxGithubRepoListRetries     = 6
	// githubRateLimitLogMinWait: detailed rate-limit logs only if backoff exceeds this.
	githubRateLimitLogMinWait = 30 * time.Second
)

// githubListWorkerCount returns how many concurrent workers to use for
// paginated GitHub repo listing (pages beyond the first). If
// GHORG_GITHUB_REPO_LIST_CONCURRENCY is unset or empty, all extra pages are
// listed in parallel (historical ghorg behavior). If set to a positive
// integer, at most that many list requests run at once.
func githubListWorkerCount(extraPages int) int {
	if extraPages < 1 {
		return 1
	}
	s := strings.TrimSpace(os.Getenv("GHORG_GITHUB_REPO_LIST_CONCURRENCY"))
	if s == "" {
		return extraPages
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return extraPages
	}
	if n > maxGithubRepoListConcurrency {
		n = maxGithubRepoListConcurrency
	}
	if n > extraPages {
		return extraPages
	}
	return n
}

func isGitHubListRateLimitError(err error) bool {
	var abuse *github.AbuseRateLimitError
	var primary *github.RateLimitError
	return errors.As(err, &abuse) || errors.As(err, &primary)
}

func formatGithubRateLimitWait(d time.Duration) string {
	d = d.Round(time.Second)
	if d < time.Minute {
		s := int(d.Seconds())
		if s <= 1 {
			return "about 1 second"
		}
		return fmt.Sprintf("about %d seconds", s)
	}
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	if s == 0 {
		return fmt.Sprintf("about %d minute(s)", m)
	}
	return fmt.Sprintf("about %d minute(s) and %d seconds", m, s)
}

func truncateForLog(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

// logGithubRateLimitWait tells the user why ghorg is pausing and for how long.
func logGithubRateLimitWait(attempt int, err error, wait time.Duration, waitCappedToMax bool) {
	nextAttempt := attempt + 2
	waitHuman := formatGithubRateLimitWait(wait)
	detail := truncateForLog(strings.TrimSpace(err.Error()), 400)

	var abuse *github.AbuseRateLimitError
	var primary *github.RateLimitError
	switch {
	case errors.As(err, &abuse):
		colorlog.PrintInfo("\nGhorg: GitHub rate limited the request used to list your repositories (REST API pagination — this is separate from GHORG_CONCURRENCY, which only affects git clones).")
		colorlog.PrintInfo("Reason: secondary rate limit (GitHub anti-abuse / scraping protection). Too many list requests happened too quickly for this token or IP.")
		colorlog.PrintInfo(fmt.Sprintf("Action: waiting %s, then retrying (next list request will be attempt %d of %d).", waitHuman, nextAttempt, maxGithubRepoListRetries))
		if waitCappedToMax {
			colorlog.PrintInfo("Note: this wait is capped at 5 minutes per pause; if GitHub still returns a rate limit, ghorg will log again and wait before the following retry.")
		}
		colorlog.PrintInfo("How to reduce this: set a lower GHORG_GITHUB_REPO_LIST_CONCURRENCY (try 4 or 1 for fully sequential listing), or use flag --github-repo-list-concurrency=4. If it persists, wait a few minutes before running ghorg again.")
		colorlog.PrintInfo("If several identical messages appear together, multiple list requests hit the limit in parallel—lowering GHORG_GITHUB_REPO_LIST_CONCURRENCY reduces that.")
		colorlog.PrintInfo("Reference: https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api")
		if detail != "" {
			colorlog.PrintInfo("GitHub response: " + detail)
		}
	case errors.As(err, &primary):
		reset := primary.Rate.Reset.UTC().Format(time.RFC3339)
		colorlog.PrintInfo("Ghorg: GitHub rate limited the request used to list your repositories (REST API pagination — this is separate from GHORG_CONCURRENCY, which only affects git clones).")
		colorlog.PrintInfo("Reason: primary REST API rate limit for this token (hourly request budget). Listing paused until the limit window resets.")
		if waitCappedToMax {
			colorlog.PrintInfo(fmt.Sprintf("Action: waiting %s before the next list request (attempt %d of %d). This pause is capped at 5 minutes per wait; GitHub indicates your token resets around %s UTC — if the next request is still limited, ghorg will wait again and log it.", waitHuman, nextAttempt, maxGithubRepoListRetries, reset))
		} else {
			colorlog.PrintInfo(fmt.Sprintf("Action: waiting %s (aligned with GitHub reset ~%s UTC), then retrying (next list request will be attempt %d of %d).", waitHuman, reset, nextAttempt, maxGithubRepoListRetries))
		}
		colorlog.PrintInfo("How to reduce this: wait until after the reset time, avoid running other heavy GitHub API clients on the same token at the same time, and/or lower GHORG_GITHUB_REPO_LIST_CONCURRENCY to fetch pages less aggressively.")
		colorlog.PrintInfo("Reference: https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api")
		if detail != "" {
			colorlog.PrintInfo("GitHub response: " + detail)
		}
	default:
		colorlog.PrintInfo("Ghorg: GitHub rate limited repository listing. Waiting " + waitHuman + " before retry.")
		if waitCappedToMax {
			colorlog.PrintInfo("Note: this wait is capped at 5 minutes per pause; ghorg may log and wait again if limits continue.")
		}
		if detail != "" {
			colorlog.PrintInfo("GitHub response: " + detail)
		}
	}
}

func sleepForGitHubRateLimit(ctx context.Context, attempt int, err error) error {
	var wait time.Duration
	var abuse *github.AbuseRateLimitError
	var primary *github.RateLimitError
	switch {
	case errors.As(err, &abuse):
		if abuse.RetryAfter != nil && *abuse.RetryAfter > 0 {
			wait = *abuse.RetryAfter
		} else {
			wait = time.Duration(30*(attempt+1)) * time.Second
		}
	case errors.As(err, &primary):
		wait = time.Until(primary.Rate.Reset.Time) + 2*time.Second
		if wait < time.Second {
			wait = time.Second
		}
	}
	if wait <= 0 {
		wait = time.Second
	}
	const maxWait = 5 * time.Minute
	waitCapped := false
	if wait > maxWait {
		wait = maxWait
		waitCapped = true
	}
	if wait > githubRateLimitLogMinWait {
		logGithubRateLimitWait(attempt, err, wait, waitCapped)
	}
	t := time.NewTimer(wait)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// fetchGitHubRepoPageWithRetry calls fetch until success or a non–rate-limit
// error, or retries are exhausted.
func fetchGitHubRepoPageWithRetry(ctx context.Context, fetch func(context.Context) ([]*github.Repository, error)) ([]*github.Repository, error) {
	for attempt := 0; attempt < maxGithubRepoListRetries; attempt++ {
		repos, err := fetch(ctx)
		if err == nil {
			return repos, nil
		}
		if !isGitHubListRateLimitError(err) || attempt == maxGithubRepoListRetries-1 {
			return nil, err
		}
		if err := sleepForGitHubRateLimit(ctx, attempt, err); err != nil {
			return nil, err
		}
	}
	panic("unreachable: github list retries")
}
