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
func (c *Client) Pending(ctx context.Context, prefix string) (res PendingResult, err error) {
	args := []string{"list", "--all-projects", "--pending"}
	if prefix != "" {
		args = []string{"list", "--project", prefix, "--pending"}
	}
	raw, err := c.R.Run(ctx, "", args...)
	if err != nil {
		return res, err
	}
	return res, decodeInto(raw, &res, "!observed_at", "!coverage", "!items", "!warnings")
}
func (c *Client) DecisionAnswer(ctx context.Context, dir, taskID, decisionID string, spec DecisionAnswerSpec) (res DecisionOut, err error) {
	var flag string
	switch spec.Outcome {
	case "accept":
		flag = "--accept"
	case "decline":
		flag = "--decline"
	case "reply":
		flag = "--reply"
	default:
		return res, &Error{Kind: "validation", Detail: "invalid decision answer outcome: " + spec.Outcome}
	}
	args := []string{"decision", "answer", taskID, decisionID, flag, spec.Text}
	if spec.SeenScope != "" {
		args = append(args, "--seen-scope", spec.SeenScope)
	}
	raw, err := c.R.Run(ctx, dir, args...)
	if err != nil {
		return res, err
	}
	return res, decodeInto(raw, &res, "!id", "!action", "!replayed", "!decision", "!warnings")
}
func (c *Client) DecisionWithdraw(ctx context.Context, dir, taskID, decisionID, reason string) (res DecisionOut, err error) {
	raw, err := c.R.Run(ctx, dir, "decision", "withdraw", taskID, decisionID, "--reason", reason)
	if err != nil {
		return res, err
	}
	return res, decodeInto(raw, &res, "!id", "!action", "!replayed", "!decision", "!warnings")
}
func (c *Client) Show(ctx context.Context, dir, id string) (res ShowResult, err error) {
	raw, err := c.R.Run(ctx, dir, "show", id)
	if err != nil {
		return res, err
	}
	return res, decodeInto(raw, &res, "!task", "!warnings")
}
func (c *Client) Root(ctx context.Context, id string) (res RootResult, err error) {
	raw, err := c.R.Run(ctx, "", "root", id)
	if err != nil {
		return res, err
	}
	return res, decodeInto(raw, &res, "!prefix", "!root", "!warnings")
}

type TaskFields struct {
	Title string
	Body  string
	Tags  []string
}

func (c *Client) AddTask(ctx context.Context, prefix string, fields TaskFields) (res AddResult, err error) {
	args := []string{"add", fields.Title, "--project", prefix}
	if fields.Body != "" {
		args = append(args, "-b", fields.Body)
	}
	for _, tag := range fields.Tags {
		args = append(args, "--tag", tag)
	}
	raw, err := c.R.Run(ctx, "", args...)
	if err != nil {
		return res, err
	}
	return res, decodeInto(raw, &res, "!id", "!action", "!warnings")
}

func (c *Client) Edit(ctx context.Context, dir, id string, fields TaskFields) (WriteResult, error) {
	args := []string{"edit", id, "--title", fields.Title, "-b", fields.Body, "--no-tags"}
	for _, tag := range fields.Tags {
		args = append(args, "--tag", tag)
	}
	return c.write(ctx, dir, args...)
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
	if p.DecisionID != "" {
		args = append(args, "--decision", p.DecisionID)
	}
	if p.DecisionKind != "" {
		args = append(args, "--decision-kind", p.DecisionKind)
	}
	if p.DecisionOrigin != "" {
		args = append(args, "--decision-origin", p.DecisionOrigin)
	}
	if p.DecisionText != "" {
		args = append(args, "--decision-text", p.DecisionText)
	}
	if p.ArtifactRepo != "" {
		args = append(args, "--artifact-repo", p.ArtifactRepo)
	}
	if p.ArtifactPath != "" {
		args = append(args, "--artifact-path", p.ArtifactPath)
	}
	if p.ArtifactCommit != "" {
		args = append(args, "--artifact-commit", p.ArtifactCommit)
	}
	if p.ArtifactSHA256 != "" {
		args = append(args, "--artifact-sha256", p.ArtifactSHA256)
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
