package docsdata

import (
	"fmt"
	"regexp"
	"strings"
)

// ChangelogEntry is one release's section of CHANGELOG.md.
type ChangelogEntry struct {
	Version string `json:"version"`
	// Date is the release date the heading gives, empty while the release is unreleased.
	Date       string `json:"date,omitempty"`
	Unreleased bool   `json:"unreleased,omitempty"`
	// Anchor is the id GitHub gives the heading, so a link into the file and into the site agree.
	Anchor string `json:"anchor"`
	// Body is the section's Markdown, everything under the heading up to the next release.
	Body string `json:"body"`
}

// releaseHeading matches a release's heading as internal/release/changelog renders it: `## 1.0.0 (date)`
// or `## 1.0.0 (unreleased)`.
var releaseHeading = regexp.MustCompile(`^## (\d+\.\d+\.\d+\S*) \(([^)]+)\)$`)

// parseChangelog splits CHANGELOG.md into its releases, newest first, as the file orders them. A second-
// level heading that is not a release refuses the parse, so a reshaped file fails here rather than
// publishing a release with another's notes.
func parseChangelog(changelog []byte) ([]ChangelogEntry, error) {
	entries := []ChangelogEntry{}
	var body []string
	flush := func() {
		if len(entries) > 0 {
			entries[len(entries)-1].Body = strings.TrimSpace(strings.Join(body, "\n")) + "\n"
		}
		body = nil
	}
	for _, line := range strings.Split(string(changelog), "\n") {
		if !strings.HasPrefix(line, "## ") {
			if len(entries) > 0 {
				body = append(body, line)
			}
			continue
		}
		match := releaseHeading.FindStringSubmatch(line)
		if match == nil {
			return nil, fmt.Errorf("docsdata: CHANGELOG.md heading %q is not a release (## <version> (<date>|unreleased))", line)
		}
		flush()
		entry := ChangelogEntry{Version: match[1], Anchor: headingAnchor(strings.TrimPrefix(line, "## "))}
		if match[2] == "unreleased" {
			entry.Unreleased = true
		} else {
			entry.Date = match[2]
		}
		entries = append(entries, entry)
	}
	flush()
	if len(entries) == 0 {
		return nil, fmt.Errorf("docsdata: CHANGELOG.md has no release")
	}
	return entries, nil
}

// headingAnchor is the id GitHub gives a heading: lowercased, punctuation other than hyphens dropped,
// spaces as hyphens. `1.0.0 (unreleased)` is `100-unreleased`.
func headingAnchor(heading string) string {
	var anchor strings.Builder
	for _, character := range strings.ToLower(heading) {
		switch {
		case character == ' ':
			anchor.WriteRune('-')
		case character == '-' || character == '_',
			character >= 'a' && character <= 'z',
			character >= '0' && character <= '9':
			anchor.WriteRune(character)
		}
	}
	return anchor.String()
}
