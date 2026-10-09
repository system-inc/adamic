// Package changelog writes a release's section of CHANGELOG.md from what changed in the rule sets cohere
// carries, and says the smallest version that change allows.
//
// The versioning policy Kirk approved (#ed2dp27): cohere is semver from 1.0.0. Turning a rule on in a set
// is a minor version, because a project that passed before can fail after it, which is a change a person
// opts into by taking a minor. A major version is for breaking the configuration format or the command
// line. The sets are the part of a release a diff can read exactly, so their entries are generated; what
// changed in cohere itself is written by a person under the same heading.
package changelog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
)

// SetDirectory is where the sets live in the module, and what their names are relative to.
const SetDirectory = "internal/lint/configuration/sets"

// Bump is how much a version must move.
type Bump int

const (
	// Patch is a release whose sets changed nothing that can fail a project that passed.
	Patch Bump = iota
	// Minor is a release that can fail a project that passed: a rule turned on, made stricter, or given
	// other options.
	Minor
	// Major is a release that breaks a configuration: a set removed, so an `extends` naming it no longer
	// loads.
	Major
)

func (bump Bump) String() string {
	return [...]string{"patch", "minor", "major"}[bump]
}

// SetChanges is what changed in one set between two releases.
type SetChanges struct {
	// Added is a set that did not exist before, Removed one that no longer exists.
	Added   bool
	Removed bool

	// TurnedOn are rules that now report, with their severity; TurnedOff are rules that no longer do.
	TurnedOn  []string
	TurnedOff []string
	// Severity are rules that report at another severity, as "rule: before -> after".
	Severity []string
	// Options are rules that report at the same severity with other options.
	Options []string
	// Settings are the set's other keys that changed: format, ignorePatterns, overrides and the rest.
	Settings []string

	// Bump is the smallest version these changes allow.
	Bump Bump
}

// Empty reports a set that did not change.
func (changes SetChanges) Empty() bool {
	return !changes.Added && !changes.Removed && len(changes.TurnedOn)+len(changes.TurnedOff)+len(changes.Severity)+
		len(changes.Options)+len(changes.Settings) == 0
}

// severity ranks a rule entry's severity: 0 off, 1 warn, 2 error, the way ESLint does, from a string, a
// number, or the first element of a list.
func severity(entry any) (int, error) {
	if list, ok := entry.([]any); ok {
		if len(list) == 0 {
			return 0, fmt.Errorf("an empty rule entry")
		}
		entry = list[0]
	}
	switch value := entry.(type) {
	case string:
		switch value {
		case "off":
			return 0, nil
		case "warn":
			return 1, nil
		case "error":
			return 2, nil
		}
	case float64:
		if value == 0 || value == 1 || value == 2 {
			return int(value), nil
		}
	}
	return 0, fmt.Errorf("%v is not a severity", entry)
}

// options is a rule entry's options, everything after its severity.
func options(entry any) []any {
	if list, ok := entry.([]any); ok && len(list) > 1 {
		return list[1:]
	}
	return nil
}

var severityNames = []string{"off", "warn", "error"}

// DiffSet compares one set's text before and after. Either may be nil, for a set added or removed.
func DiffSet(before []byte, after []byte) (SetChanges, error) {
	changes := SetChanges{Added: before == nil && after != nil, Removed: before != nil && after == nil}
	old, err := decodeSet(before)
	if err != nil {
		return SetChanges{}, fmt.Errorf("reading the set before: %w", err)
	}
	updated, err := decodeSet(after)
	if err != nil {
		return SetChanges{}, fmt.Errorf("reading the set after: %w", err)
	}

	oldRules, _ := old["rules"].(map[string]any)
	newRules, _ := updated["rules"].(map[string]any)
	for _, name := range slices.Sorted(maps.Keys(unionKeys(oldRules, newRules))) {
		was, wasErr := severityOf(oldRules, name)
		now, nowErr := severityOf(newRules, name)
		if wasErr != nil || nowErr != nil {
			return SetChanges{}, fmt.Errorf("rule %s: %v", name, firstError(wasErr, nowErr))
		}
		switch {
		case was == 0 && now > 0:
			changes.TurnedOn = append(changes.TurnedOn, fmt.Sprintf("%s (%s)", name, severityNames[now]))
		case was > 0 && now == 0:
			changes.TurnedOff = append(changes.TurnedOff, name)
		case was != now:
			changes.Severity = append(changes.Severity, fmt.Sprintf("%s: %s -> %s", name, severityNames[was], severityNames[now]))
			if now > was {
				changes.Bump = Minor
			}
		case now > 0 && !reflect.DeepEqual(options(oldRules[name]), options(newRules[name])):
			changes.Options = append(changes.Options, name)
		}
	}
	if len(changes.TurnedOn) > 0 || len(changes.Options) > 0 {
		changes.Bump = Minor
	}

	delete(old, "rules")
	delete(updated, "rules")
	for _, key := range slices.Sorted(maps.Keys(unionKeys(old, updated))) {
		if !reflect.DeepEqual(old[key], updated[key]) {
			changes.Settings = append(changes.Settings, key)
		}
	}
	// A setting cannot be judged from its name: an ignore pattern removed or an override added can each
	// fail a project that passed. So any change to one counts as one that can.
	if len(changes.Settings) > 0 || changes.Added {
		changes.Bump = max(changes.Bump, Minor)
	}
	if changes.Removed {
		changes.Bump = Major
	}
	return changes, nil
}

func decodeSet(text []byte) (map[string]any, error) {
	decoded := map[string]any{}
	if text == nil {
		return decoded, nil
	}
	if err := json.Unmarshal(text, &decoded); err != nil {
		return nil, err
	}
	return decoded, nil
}

func severityOf(rules map[string]any, name string) (int, error) {
	entry, present := rules[name]
	if !present {
		return 0, nil
	}
	return severity(entry)
}

func unionKeys(left map[string]any, right map[string]any) map[string]bool {
	keys := map[string]bool{}
	for key := range left {
		keys[key] = true
	}
	for key := range right {
		keys[key] = true
	}
	return keys
}

func firstError(errors ...error) error {
	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}

// Diff compares every set between two trees, each as set name to its text, and returns the sets that
// changed and the smallest version the whole change allows.
func Diff(before map[string][]byte, after map[string][]byte) (map[string]SetChanges, Bump, error) {
	changed := map[string]SetChanges{}
	bump := Patch
	names := map[string]bool{}
	for name := range before {
		names[name] = true
	}
	for name := range after {
		names[name] = true
	}
	for name := range names {
		changes, err := DiffSet(before[name], after[name])
		if err != nil {
			return nil, Patch, fmt.Errorf("cohere:%s: %w", name, err)
		}
		if changes.Empty() {
			continue
		}
		changed[name] = changes
		bump = max(bump, changes.Bump)
	}
	return changed, bump, nil
}

// SetsAt reads every set as it is in a commit of the repository at moduleDirectory, as name to text. An
// empty ref reads the working tree.
func SetsAt(moduleDirectory string, ref string) (map[string][]byte, error) {
	sets := map[string][]byte{}
	if ref == "" {
		root := filepath.Join(moduleDirectory, SetDirectory)
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || filepath.Ext(path) != ".json" {
				return err
			}
			contents, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			relative, _ := filepath.Rel(root, path)
			sets[strings.TrimSuffix(filepath.ToSlash(relative), ".json")] = contents
			return nil
		})
		return sets, err
	}

	listing, err := git(moduleDirectory, "ls-tree", "-r", "--name-only", ref, "--", SetDirectory)
	if err != nil {
		return nil, err
	}
	for _, path := range strings.Fields(string(listing)) {
		if filepath.Ext(path) != ".json" {
			continue
		}
		contents, err := git(moduleDirectory, "show", ref+":"+path)
		if err != nil {
			return nil, err
		}
		sets[strings.TrimSuffix(strings.TrimPrefix(path, SetDirectory+"/"), ".json")] = contents
	}
	return sets, nil
}

func git(directory string, arguments ...string) ([]byte, error) {
	command := exec.Command("git", append([]string{"-C", directory}, arguments...)...)
	var standardError bytes.Buffer
	command.Stderr = &standardError
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", strings.Join(arguments, " "), err, strings.TrimSpace(standardError.String()))
	}
	return output, nil
}

// Render writes a release's section: its heading, then one subsection per set that changed, in name
// order, with every change named. The section for the first release, which has nothing to compare
// with, lists each set and how many rules it turns on.
func Render(version string, date string, changed map[string]SetChanges, first map[string][]byte) string {
	var section strings.Builder
	fmt.Fprintf(&section, "## %s (%s)\n\n", version, date)
	if first != nil {
		section.WriteString("The first release. It carries these rule sets:\n\n")
		for _, name := range slices.Sorted(maps.Keys(first)) {
			decoded, _ := decodeSet(first[name])
			rules, _ := decoded["rules"].(map[string]any)
			on := 0
			for ruleName := range rules {
				if level, err := severityOf(rules, ruleName); err == nil && level > 0 {
					on++
				}
			}
			fmt.Fprintf(&section, "- `cohere:%s`, %d rules on\n", name, on)
		}
		return section.String()
	}
	if len(changed) == 0 {
		section.WriteString("No rule set changed.\n")
		return section.String()
	}
	for _, name := range slices.Sorted(maps.Keys(changed)) {
		changes := changed[name]
		fmt.Fprintf(&section, "### `cohere:%s`\n\n", name)
		switch {
		case changes.Added:
			section.WriteString("New in this release.\n\n")
		case changes.Removed:
			section.WriteString("Removed in this release.\n\n")
		}
		list := func(label string, items []string) {
			for _, item := range items {
				fmt.Fprintf(&section, "- %s `%s`\n", label, item)
			}
		}
		list("On:", changes.TurnedOn)
		list("Off:", changes.TurnedOff)
		list("Severity:", changes.Severity)
		list("Options changed:", changes.Options)
		list("Setting changed:", changes.Settings)
		section.WriteString("\n")
	}
	return strings.TrimRight(section.String(), "\n") + "\n"
}

// RequireBump refuses a version that moves less than the change allows. Versions are compared as plain
// major.minor.patch. A prerelease, which release.Build allows, has no section of its own: the section is
// written for the release it leads to.
func RequireBump(previous string, next string, bump Bump) error {
	from, err := parseVersion(previous)
	if err != nil {
		return err
	}
	to, err := parseVersion(next)
	if err != nil {
		return err
	}
	switch {
	case to[0] != from[0]:
		if to[0] < from[0] {
			return fmt.Errorf("%s comes before %s", next, previous)
		}
		return nil
	case bump == Major:
		return fmt.Errorf("%s keeps the major version of %s, and a rule set was removed, so a configuration that extends it no longer loads, which needs a major version", next, previous)
	case to[1] != from[1]:
		if to[1] < from[1] {
			return fmt.Errorf("%s comes before %s", next, previous)
		}
		return nil
	case to[2] <= from[2]:
		return fmt.Errorf("%s does not come after %s", next, previous)
	case bump == Minor:
		return fmt.Errorf("%s is a patch after %s, and a rule set changed in a way that can fail a project that passed, which needs at least a minor version", next, previous)
	}
	return nil
}

func parseVersion(version string) ([3]int, error) {
	var parsed [3]int
	if _, err := fmt.Sscanf(version, "%d.%d.%d", &parsed[0], &parsed[1], &parsed[2]); err != nil ||
		fmt.Sprintf("%d.%d.%d", parsed[0], parsed[1], parsed[2]) != version {
		return parsed, fmt.Errorf("%q is not a major.minor.patch version", version)
	}
	return parsed, nil
}
