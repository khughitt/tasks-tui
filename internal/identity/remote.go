// Package identity reproduces familiar's project identity: the same key, pins, hash,
// and hue table, so a project's accent in the TUI matches its pet and terminal tint
// (spec §8). Nothing here imports familiar; the algorithm is copied and tested against
// familiar's own vectors.
package identity

import (
	"regexp"
	"strings"
)

var (
	schemeRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*://`)
	userRe   = regexp.MustCompile(`^[^/@]+@`)
	portRe   = regexp.MustCompile(`:(\d+)/`)
)

// NormalizeRemote reduces every URL form of one repository to host/owner/name,
// lowercased, as familiar's normalizeRemote does. ok is false when the input is not a
// remote (empty, a local path, no host).
func NormalizeRemote(url string) (string, bool) {
	rest := strings.TrimSpace(url)
	if rest == "" {
		return "", false
	}
	hadScheme := schemeRe.MatchString(rest)
	rest = schemeRe.ReplaceAllString(rest, "")
	rest = userRe.ReplaceAllString(rest, "")
	if hadScheme {
		// `:<digits>/` is a port only in URL form; in scp form the text after the colon is a path.
		rest = portRe.ReplaceAllString(rest, "/")
	}
	rest = strings.Replace(rest, ":", "/", 1)
	rest = strings.TrimSuffix(rest, ".git")
	rest = strings.TrimRight(rest, "/")

	var parts []string
	for _, p := range strings.Split(rest, "/") {
		if p != "" {
			parts = append(parts, p)
		}
	}
	if len(parts) < 2 || !strings.Contains(parts[0], ".") {
		return "", false
	}
	return strings.ToLower(strings.Join(parts, "/")), true
}
