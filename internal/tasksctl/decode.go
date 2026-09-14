package tasksctl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

func requireKeys(raw []byte, keys ...string) error {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return &Error{Kind: "decode", Detail: fmt.Sprintf("not a JSON object: %v", err)}
	}
	var missing, null []string
	for _, key := range keys {
		nonNull := strings.HasPrefix(key, "!")
		key = strings.TrimPrefix(key, "!")
		value, ok := obj[key]
		if !ok {
			missing = append(missing, key)
		} else if nonNull && string(bytes.TrimSpace(value)) == "null" {
			null = append(null, key)
		}
	}
	var problems []string
	if len(missing) > 0 {
		problems = append(problems, "missing "+strings.Join(missing, ", "))
	}
	if len(null) > 0 {
		problems = append(problems, "null "+strings.Join(null, ", "))
	}
	if len(problems) > 0 {
		return &Error{Kind: "decode", Detail: "response " + strings.Join(problems, "; ")}
	}
	return nil
}
func decodeInto(raw []byte, into any, keys ...string) error {
	if err := requireKeys(raw, keys...); err != nil {
		return err
	}
	if err := json.Unmarshal(raw, into); err != nil {
		return &Error{Kind: "decode", Detail: err.Error()}
	}
	return nil
}

var (
	rowKeys    = []string{"!id", "!title", "!status", "!priority", "size", "complexity", "process", "owner", "!updated", "!tags", "parent", "!child_count", "!open_descendant_count", "claim", "park", "periodic"}
	parkedKeys = []string{"!id", "!title", "status", "priority", "size", "complexity", "process", "owner", "updated", "!tags", "parent", "child_count", "open_descendant_count", "claim", "park", "escalation", "phase"}
	taskKeys   = []string{"!id", "!title", "!status", "!priority", "size", "complexity", "process", "!parallel", "every", "owner", "!created", "!updated", "started", "completed", "last_done", "!depends", "parent", "!tags", "source", "model", "agent", "spec", "plan", "step", "!body", "!notes"}
	claimKeys  = []string{"!owner", "!session", "!host", "pid", "!worktree", "!started", "!seen", "!live"}
	parkKeys   = []string{"!at", "!next_step", "!waiting_on", "reason", "needs", "minutes", "!session", "!owner", "!host", "!worktree"}
	projKeys   = []string{"!prefix", "!root", "!reachable", "counts", "total", "last_activity"}
	countKeys  = []string{"!idea", "!todo", "!doing", "!blocked", "!shelved", "!done", "!dropped"}
	perKeys    = []string{"!every", "last_done", "due", "!due_now"}
	relKeys    = []string{"!id", "!title", "!status"}
	depKeys    = []string{"!id", "title", "status", "!resolved"}
	noteKeys   = []string{"!at", "!by", "!text"}
	escKeys    = []string{"!level", "!at", "!session"}
)

func (r *Row) UnmarshalJSON(raw []byte) error {
	type plain Row
	return decodeInto(raw, (*plain)(r), rowKeys...)
}
func (p *ParkedRow) UnmarshalJSON(raw []byte) error {
	type plain ParkedRow
	return decodeInto(raw, (*plain)(p), parkedKeys...)
}
func (t *Task) UnmarshalJSON(raw []byte) error {
	type plain Task
	return decodeInto(raw, (*plain)(t), taskKeys...)
}
func (c *ClaimInfo) UnmarshalJSON(raw []byte) error {
	type plain ClaimInfo
	return decodeInto(raw, (*plain)(c), claimKeys...)
}
func (p *ParkInfo) UnmarshalJSON(raw []byte) error {
	type plain ParkInfo
	return decodeInto(raw, (*plain)(p), parkKeys...)
}
func (p *Project) UnmarshalJSON(raw []byte) error {
	type plain Project
	return decodeInto(raw, (*plain)(p), projKeys...)
}
func (c *Counts) UnmarshalJSON(raw []byte) error {
	type plain Counts
	return decodeInto(raw, (*plain)(c), countKeys...)
}
func (p *Periodic) UnmarshalJSON(raw []byte) error {
	type plain Periodic
	return decodeInto(raw, (*plain)(p), perKeys...)
}
func (r *Relation) UnmarshalJSON(raw []byte) error {
	type plain Relation
	return decodeInto(raw, (*plain)(r), relKeys...)
}
func (d *Dependency) UnmarshalJSON(raw []byte) error {
	type plain Dependency
	return decodeInto(raw, (*plain)(d), depKeys...)
}
func (n *Note) UnmarshalJSON(raw []byte) error {
	type plain Note
	return decodeInto(raw, (*plain)(n), noteKeys...)
}
func (e *Escalation) UnmarshalJSON(raw []byte) error {
	type plain Escalation
	return decodeInto(raw, (*plain)(e), escKeys...)
}
