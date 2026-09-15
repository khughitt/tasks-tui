package tasksctl

type ClaimInfo struct {
	Owner    string `json:"owner"`
	Session  string `json:"session"`
	Host     string `json:"host"`
	PID      *int   `json:"pid"`
	Worktree string `json:"worktree"`
	Started  string `json:"started"`
	Seen     string `json:"seen"`
	Live     bool   `json:"live"`
}
type ParkInfo struct {
	At        string  `json:"at"`
	NextStep  string  `json:"next_step"`
	WaitingOn string  `json:"waiting_on"`
	Reason    *string `json:"reason"`
	Needs     *string `json:"needs"`
	Minutes   *int    `json:"minutes"`
	Session   string  `json:"session"`
	Owner     string  `json:"owner"`
	Host      string  `json:"host"`
	Worktree  string  `json:"worktree"`
}
type Escalation struct {
	Level   string `json:"level"`
	At      string `json:"at"`
	Session string `json:"session"`
}
type Periodic struct {
	Every    string  `json:"every"`
	LastDone *string `json:"last_done"`
	Due      *string `json:"due"`
	DueNow   bool    `json:"due_now"`
}

type Row struct {
	ID                  string     `json:"id"`
	Title               string     `json:"title"`
	Status              string     `json:"status"`
	Priority            int        `json:"priority"`
	Size                *string    `json:"size"`
	Complexity          *string    `json:"complexity"`
	Process             *string    `json:"process"`
	Owner               *string    `json:"owner"`
	Updated             string     `json:"updated"`
	Tags                []string   `json:"tags"`
	Parent              *string    `json:"parent"`
	ChildCount          int        `json:"child_count"`
	OpenDescendantCount int        `json:"open_descendant_count"`
	Claim               *ClaimInfo `json:"claim"`
	Park                *ParkInfo  `json:"park"`
	Periodic            *Periodic  `json:"periodic"`
}

func (r Row) Prefix() string        { return prefixOf(r.ID) }
func (r Row) LiveClaim() *ClaimInfo { return liveClaim(r.Claim) }

type ParkedRow struct {
	ID                  string      `json:"id"`
	Title               string      `json:"title"`
	Status              *string     `json:"status"`
	Priority            *int        `json:"priority"`
	Size                *string     `json:"size"`
	Complexity          *string     `json:"complexity"`
	Process             *string     `json:"process"`
	Owner               *string     `json:"owner"`
	Updated             *string     `json:"updated"`
	Tags                []string    `json:"tags"`
	Parent              *string     `json:"parent"`
	ChildCount          *int        `json:"child_count"`
	OpenDescendantCount *int        `json:"open_descendant_count"`
	Claim               *ClaimInfo  `json:"claim"`
	Park                *ParkInfo   `json:"park"`
	Escalation          *Escalation `json:"escalation"`
	Phase               *string     `json:"phase"`
}

func (p ParkedRow) Prefix() string        { return prefixOf(p.ID) }
func (p ParkedRow) Unresolved() bool      { return p.Status == nil }
func (p ParkedRow) LiveClaim() *ClaimInfo { return liveClaim(p.Claim) }
func liveClaim(c *ClaimInfo) *ClaimInfo {
	if c != nil && c.Live {
		return c
	}
	return nil
}
func prefixOf(id string) string {
	for i := len(id) - 1; i >= 0; i-- {
		if id[i] == '-' {
			return id[:i]
		}
	}
	return id
}

type Counts struct {
	Idea    int `json:"idea"`
	Todo    int `json:"todo"`
	Doing   int `json:"doing"`
	Blocked int `json:"blocked"`
	Shelved int `json:"shelved"`
	Done    int `json:"done"`
	Dropped int `json:"dropped"`
}
type Project struct {
	Prefix       string  `json:"prefix"`
	Root         string  `json:"root"`
	Reachable    bool    `json:"reachable"`
	Counts       *Counts `json:"counts"`
	Total        *int    `json:"total"`
	LastActivity *string `json:"last_activity"`
}
type Note struct {
	At   string `json:"at"`
	By   string `json:"by"`
	Text string `json:"text"`
}
type Task struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Status     string   `json:"status"`
	Priority   int      `json:"priority"`
	Size       *string  `json:"size"`
	Complexity *string  `json:"complexity"`
	Process    *string  `json:"process"`
	Parallel   bool     `json:"parallel"`
	Every      *string  `json:"every"`
	Owner      *string  `json:"owner"`
	Created    string   `json:"created"`
	Updated    string   `json:"updated"`
	Started    *string  `json:"started"`
	Completed  *string  `json:"completed"`
	LastDone   *string  `json:"last_done"`
	Depends    []string `json:"depends"`
	Parent     *string  `json:"parent"`
	Tags       []string `json:"tags"`
	Source     *string  `json:"source"`
	Model      *string  `json:"model"`
	Agent      *string  `json:"agent"`
	Spec       *string  `json:"spec"`
	Plan       *string  `json:"plan"`
	Step       *string  `json:"step"`
	Body       string   `json:"body"`
	Notes      []Note   `json:"notes"`
}

func (t Task) Prefix() string { return prefixOf(t.ID) }

type Relation struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}
type Dependency struct {
	ID       string  `json:"id"`
	Title    *string `json:"title"`
	Status   *string `json:"status"`
	Resolved bool    `json:"resolved"`
}
type ProjectsResult struct {
	Projects []Project `json:"projects"`
	Warnings []string  `json:"warnings"`
}
type PrimeResult struct {
	Prefix   string      `json:"prefix"`
	Counts   Counts      `json:"counts"`
	Ready    []Row       `json:"ready"`
	Doing    []Row       `json:"doing"`
	Parked   []ParkedRow `json:"parked"`
	Warnings []string    `json:"warnings"`
}
type RowsResult struct {
	Tasks    []Row    `json:"tasks"`
	Warnings []string `json:"warnings"`
}
type ParkedResult struct {
	Tasks    []ParkedRow `json:"tasks"`
	Warnings []string    `json:"warnings"`
}
type ShowResult struct {
	Task       Task         `json:"task"`
	SpecPath   *string      `json:"spec_path"`
	PlanPath   *string      `json:"plan_path"`
	StepFound  *bool        `json:"step_found"`
	DependsOn  []Dependency `json:"depends_on"`
	Parent     *Relation    `json:"parent"`
	Children   []Relation   `json:"children"`
	Claim      *ClaimInfo   `json:"claim"`
	Park       *ParkInfo    `json:"park"`
	Escalation *Escalation  `json:"escalation"`
	Periodic   *Periodic    `json:"periodic"`
	Warnings   []string     `json:"warnings"`
}

func (s ShowResult) LiveClaim() *ClaimInfo { return liveClaim(s.Claim) }

type RootResult struct {
	Prefix   string   `json:"prefix"`
	Root     string   `json:"root"`
	Warnings []string `json:"warnings"`
}
type Tag struct {
	Tag      string         `json:"tag"`
	Meaning  *string        `json:"meaning"`
	Count    int            `json:"count"`
	Projects map[string]int `json:"projects"`
}
type TagsResult struct {
	Tags     []Tag    `json:"tags"`
	Warnings []string `json:"warnings"`
}
type AddResult struct {
	ID       string   `json:"id"`
	Action   string   `json:"action"`
	Warnings []string `json:"warnings"`
}
type WriteResult struct {
	ID       string   `json:"id"`
	Warnings []string `json:"warnings"`
}
type ListOpts struct {
	Status      []string
	SortUpdated bool
}
type ParkSpec struct {
	NextStep      string
	WaitingOnUser bool
	Reason        string
}
