package bitbucket

type Diff struct {
	c *Client
}

func (d *Diff) GetDiff(do *DiffOptions) (interface{}, error) {
	urlStr := d.c.requestUrl("/repositories/%s/%s/diff/%s", do.Owner, do.RepoSlug, do.Spec)
	return d.c.executeRaw("GET", urlStr, "diff")
}

func (d *Diff) GetPatch(do *DiffOptions) (interface{}, error) {
	urlStr := d.c.requestUrl("/repositories/%s/%s/patch/%s", do.Owner, do.RepoSlug, do.Spec)
	return d.c.execute("GET", urlStr, "")
}
