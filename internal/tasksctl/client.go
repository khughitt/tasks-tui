package tasksctl

import "context"

type Client struct{ R Runner }

func (c *Client) Tags(ctx context.Context, prefix string) (res TagsResult, err error) {
	raw, err := c.R.Run(ctx, "", "tags", "--project", prefix)
	if err != nil {
		return res, err
	}
	return res, decodeInto(raw, &res, "!tags", "!warnings")
}

func (c *Client) Projects(ctx context.Context) (res ProjectsResult, err error) {
	raw, err := c.R.Run(ctx, "", "projects")
	if err != nil {
		return res, err
	}
	return res, decodeInto(raw, &res, "!projects", "!warnings")
}
func (c *Client) Prime(ctx context.Context, prefix string) (res PrimeResult, err error) {
	raw, err := c.R.Run(ctx, "", "prime", "--project", prefix)
	if err != nil {
		return res, err
	}
	return res, decodeInto(raw, &res, "!prefix", "!counts", "!ready", "!doing", "!parked", "!warnings")
}
func (c *Client) Ready(ctx context.Context, prefix string) (RowsResult, error) {
	return c.rows(ctx, "ready", "--project", prefix)
}
func (c *Client) List(ctx context.Context, prefix string, opts ListOpts) (RowsResult, error) {
	args := []string{"list", "--project", prefix}
	for _, status := range opts.Status {
		args = append(args, "--status", status)
	}
	if opts.SortUpdated {
		args = append(args, "--sort", "updated")
	}
	return c.rows(ctx, args...)
}
func (c *Client) DoingAll(ctx context.Context) (RowsResult, error) {
	return c.rows(ctx, "list", "--all-projects", "--status", "doing")
}
func (c *Client) rows(ctx context.Context, args ...string) (res RowsResult, err error) {
	raw, err := c.R.Run(ctx, "", args...)
	if err != nil {
		return res, err
	}
	return res, decodeInto(raw, &res, "!tasks", "!warnings")
}
func (c *Client) Parked(ctx context.Context, prefix string) (res ParkedResult, err error) {
	args := []string{"list", "--all-projects", "--parked"}
	if prefix != "" {
		args = []string{"list", "--project", prefix, "--parked"}
	}
	raw, err := c.R.Run(ctx, "", args...)
	if err != nil {
		return res, err
	}
	return res, decodeInto(raw, &res, "!tasks", "!warnings")
}
func (c *Client) Show(ctx context.Context, dir, id string) (res ShowResult, err error) {
	raw, err := c.R.Run(ctx, dir, "show", id)
	if err != nil {
		return res, err
	}
	return res, decodeInto(raw, &res, "!task", "spec_path", "plan_path", "step_found", "!depends_on", "parent", "!children", "claim", "park", "escalation", "periodic", "!warnings")
}
func (c *Client) Root(ctx context.Context, id string) (res RootResult, err error) {
	raw, err := c.R.Run(ctx, "", "root", id)
	if err != nil {
		return res, err
	}
	return res, decodeInto(raw, &res, "!prefix", "!root", "!warnings")
}

// Add runs the argv quickadd built (spec §6.1). --project is inside args; no -C.
func (c *Client) Add(ctx context.Context, args []string) (res AddResult, err error) {
	raw, err := c.R.Run(ctx, "", args...)
	if err != nil {
		return res, err
	}
	return res, decodeInto(raw, &res, "!id", "!action", "!warnings")
}

func (c *Client) Start(ctx context.Context, dir, id string, force bool) (WriteResult, error) {
	args := []string{"start"}
	if force {
		args = append(args, "--force")
	}
	return c.write(ctx, dir, append(args, id)...)
}

func (c *Client) Park(ctx context.Context, dir, id string, p ParkSpec) (WriteResult, error) {
	args := []string{"park", id, p.NextStep}
	if p.WaitingOnUser {
		args = append(args, "--waiting-on", "user")
	}
	if p.Reason != "" {
		args = append(args, "--reason", p.Reason)
	}
	return c.write(ctx, dir, args...)
}

func (c *Client) Done(ctx context.Context, dir, id, message string) (WriteResult, error) {
	return c.write(ctx, dir, withMessage([]string{"done", id}, message)...)
}

func (c *Client) Drop(ctx context.Context, dir, id, message string) (WriteResult, error) {
	return c.write(ctx, dir, withMessage([]string{"drop", id}, message)...)
}

func withMessage(args []string, message string) []string {
	if message != "" {
		return append(args, message)
	}
	return args
}

func (c *Client) write(ctx context.Context, dir string, args ...string) (res WriteResult, err error) {
	raw, err := c.R.Run(ctx, dir, args...)
	if err != nil {
		return res, err
	}
	return res, decodeInto(raw, &res, "!id", "!warnings")
}
