package configuration

import (
	"encoding/binary"
	"encoding/json"
	"path/filepath"
	"strings"
	"sync"
)

// Resolved is one file's answer: whether it is linted at all, and which rules apply to it.
//
// Shared, not owned: files that resolve the same way get the same Rules map, so nothing may write into
// one (see resolutionMemo).
type Resolved struct {
	// Ignored is true when an ignorePattern excluded the file entirely.
	Ignored bool

	// IgnoredBy is the pattern that excluded it, so a report can say which one rather than only that
	// something did.
	IgnoredBy string

	// Rules is the effective setting per rule name after overrides are applied.
	Rules map[string]RuleSetting

	// identity is the memo entry this resolution came from, nil for one built by hand.
	identity *resolutionIdentity
}

// resolutionIdentity names one distinct resolution of one Config. A field of its own, because pointers to
// zero-size values need not be distinct.
type resolutionIdentity struct {
	match string
}

// Identity is the same for every file of one Config that resolved the same way, and different otherwise,
// so a caller can derive what depends only on the resolution once per resolution rather than once per
// file: which rules apply, and their decoded options. Nil for a resolution with no identity, an ignored
// file or one built by hand, which a caller treats as unique.
func (r Resolved) Identity() any {
	if r.identity == nil {
		return nil
	}
	return r.identity
}

// resolutionMemo is a Config's resolutions, by which of its overrides a path matched.
//
// A resolution depends on nothing else: the base rules, then each matching override in order. So two
// paths matching the same overrides resolve identically, and copying the base map for each was the
// largest cost of resolving: 248 MB allocated on a cold ahra run, one copy of about 500 settings per file
// (#9jpmqm9). A tree has a few dozen distinct override matches across thousands of files.
//
// Per Config, which under zero config means per house variant: the whole stack's Config dispatches to a
// variant before anything is matched, so a React file and a Node file matching the same overrides still
// resolve through different variants (see Config.Resolve).
type resolutionMemo struct {
	mutex   sync.Mutex
	byMatch map[string]Resolved
}

// configGlobs is a Config's patterns compiled once (#hsd2dfb): ignore is IgnorePatterns by index, and
// overrides holds each override's Files by the same indices as Overrides.
type configGlobs struct {
	ignore    []glob
	overrides [][]glob
}

// globs returns the Config's compiled patterns, compiling them on first use. Two goroutines racing to
// compile build the same answer, and the first one stored wins.
func (c *Config) globs() *configGlobs {
	if compiled := c.compiledGlobs.Load(); compiled != nil {
		return compiled
	}
	compiled := &configGlobs{
		ignore:    make([]glob, len(c.IgnorePatterns)),
		overrides: make([][]glob, len(c.Overrides)),
	}
	for index, pattern := range c.IgnorePatterns {
		compiled.ignore[index] = compileGlob(pattern)
	}
	for index, override := range c.Overrides {
		compiled.overrides[index] = make([]glob, len(override.Files))
		for fileIndex, pattern := range override.Files {
			compiled.overrides[index][fileIndex] = compileGlob(pattern)
		}
	}
	c.compiledGlobs.CompareAndSwap(nil, compiled)
	return c.compiledGlobs.Load()
}

// resolutions returns the Config's memo, creating it on first use.
func (c *Config) resolutions() *resolutionMemo {
	if memo := c.resolutionMemo.Load(); memo != nil {
		return memo
	}
	c.resolutionMemo.CompareAndSwap(nil, &resolutionMemo{byMatch: map[string]Resolved{}})
	return c.resolutionMemo.Load()
}

// Enabled reports whether a rule runs on this file.
func (r Resolved) Enabled(ruleName string) bool {
	status, _ := r.StatusOf(ruleName)
	return status == StatusEnabled
}

// Status is why a rule did or did not run on a file.
type Status int

const (
	// StatusEnabled means the rule runs.
	StatusEnabled Status = iota

	// StatusScopedOff means the config configures this rule and turned it off, whether at the base
	// level or through an override. Someone decided this.
	StatusScopedOff

	// StatusUnconfigured means the config never mentions the rule at all. Nobody decided anything;
	// the rule was added to the registry and no one has said whether it should run.
	//
	// This is a different fact from StatusScopedOff and reporting them as one is a lie. A rule
	// nobody has configured is a rule waiting on a decision, and it should read that way rather
	// than as a deliberate exclusion.
	StatusUnconfigured

	// StatusFileIgnored means no rule runs on this file.
	StatusFileIgnored
)

// StatusOf reports why a rule runs or does not, and its setting when it has one.
func (r Resolved) StatusOf(ruleName string) (Status, RuleSetting) {
	if r.Ignored {
		return StatusFileIgnored, RuleSetting{}
	}
	setting, configured := r.settingFor(ruleName)
	if !configured {
		// A rule the config never mentions does not run. Defaulting the other way would mean adding
		// a rule to the registry silently turns it on across the whole tree, which is a decision
		// that belongs in the config rather than in a Go file.
		return StatusUnconfigured, RuleSetting{}
	}
	if setting.Severity == SeverityOff {
		return StatusScopedOff, setting
	}
	return StatusEnabled, setting
}

// RawOptionsFor returns every option element a rule was configured with, as JSON, or nil when it
// has none.
//
// Every element, not the first. This is ESLint's `context.options`, and it used to be `tuple[1]`
// alone, which is how `["error", {"object": true}, {"enforceForRenamedProperties": true}]` loaded
// clean, ran, and reported nothing on source the second element exists to flag. The list goes to
// `OptionsRegistry.Decode` whole, which is the one place that knows how many elements each rule
// takes and refuses the rest by name.
//
// Raw rather than decoded, because only the rule's own package knows the struct its options should
// become. The registry pairs each rule with a decoder; this returns the bytes that decoder reads.
// Handing a rule this JSON directly would fail its type assertion and make it decline every file,
// which is the inert-rule defect wearing a different hat.
func (r Resolved) RawOptionsFor(ruleName string) []json.RawMessage {
	setting, configured := r.settingFor(ruleName)
	if !configured {
		return nil
	}
	return setting.Options
}

// settingFor looks a rule up by its registry name, reconciling the plugin prefix.
//
// The config writes `nexus/consistency-no-enum`; the registry writes `consistency-no-enum`. The
// prefix names which plugin supplied a rule, which mattered when rules were loaded and means
// nothing now that they are compiled in.
//
// This has to be resolved somewhere, and getting it wrong is not subtle: matching on the exact name
// alone made every rule unconfigured, therefore disabled, and cohere printed 0 findings over 3,408
// files with exit 0. The coverage line is what caught it, reporting every rule scoped off for every
// file. A tool without that line would have shipped a green run that checked nothing.
//
// The suffix must fall on a `/` boundary. Plain suffix matching would let a config entry for
// `no-enum` silence `consistency-no-enum`, which is a rule nobody named.
func (r Resolved) settingFor(ruleName string) (RuleSetting, bool) {
	if setting, configured := r.Rules[ruleName]; configured {
		return setting, true
	}

	// The suffix must fall on a `/` boundary, or a config entry for `no-enum` would silence
	// `consistency-no-enum`.
	//
	// Collected rather than returned on first match, which is the whole repair. The old loop
	// returned the first entry the range yielded, and Go randomises map iteration order, so a bare
	// registry name reachable from two prefixed config keys resolved differently between runs of one
	// binary over one config. Measured before this guard, with `@typescript-eslint/no-shadow` at
	// error and `nexus/no-shadow` at off: the bare name resolved enabled 157 times and disabled 43
	// times across 200 resolutions.
	//
	// That is the worst shape a configuration defect can take. A surprising result invites a re-run,
	// and a re-run here manufactures a second opinion rather than a confirmation.
	var matched RuleSetting
	found := false
	ambiguous := false
	for configuredName, setting := range r.Rules {
		if configuredName == ruleName || !KeyReachesRule(configuredName, ruleName) {
			continue
		}
		if found {
			ambiguous = true
			continue
		}
		matched = setting
		found = true
	}

	// Two prefixed entries reach one bare registry name and the config has not said which it means.
	// Reporting unconfigured is the honest answer: it routes to the caller that already knows how to
	// say "nobody has decided about this rule", rather than picking one at random and looking sure.
	if ambiguous {
		return RuleSetting{}, false
	}
	return matched, found
}

// KeyReachesRule reports whether a config key configures a registered rule: the key is the rule's
// name, or the rule's name qualified with a plugin prefix on a `/` boundary.
//
// One direction only. `nexus/consistency-no-enum` reaches a rule registered as
// `consistency-no-enum`, but a bare `no-unused-vars` does not reach a rule registered as
// `@typescript-eslint/no-unused-vars`. It is exported so the orphaned-key report asks the same
// question: that report once tested the reverse direction, so after the 2026-10-02 rename it called
// phi api's bare keys resolving while this function left all three rules unconfigured, and 152
// findings disappeared with no warning.
//
// It is asked for every configured key against every rule not named exactly, for every file, so it
// allocates nothing: building "/"+ruleName for each comparison was about 0.2s of CPU in a cold ahra run
// (#zqsdzbq).
func KeyReachesRule(key string, ruleName string) bool {
	if key == ruleName {
		return true
	}
	boundary := len(key) - len(ruleName) - 1
	return boundary >= 0 && key[boundary] == '/' && key[boundary+1:] == ruleName
}

// Resolve computes the effective configuration for one file path.
//
// Precedence is base rules first, then each matching override in order, later blocks winning. That
// is how both ESLint and oxlint resolve, and matching them exactly is the whole point: cohere cannot
// be diffed honestly against the gate it replaces while the two disagree about which rules were
// even supposed to run.
//
// Under zero config the file resolves through the variant of the house stack it gets (see house.go).
// That dispatch comes first, before the memo, so each variant memoizes its own resolutions.
//
// Resolutions are memoized by which overrides a path matched (see resolutionMemo), so the result is
// shared with every other path that matched the same ones.
func (c *Config) Resolve(path string) Resolved {
	if c.house != nil {
		return c.house.variantFor(path).Resolve(path)
	}
	// Split once, and asked of every compiled pattern.
	segments := splitPath(c.relativePath(path))
	globs := c.globs()

	for index, pattern := range globs.ignore {
		if pattern.matches(segments) {
			return Resolved{Ignored: true, IgnoredBy: c.IgnorePatterns[index]}
		}
	}

	// The match, as the indices of the overrides that apply, in order. On the stack for the common case
	// of a few dozen overrides, and looked up without converting it to a string, so a path whose match is
	// already memoized allocates nothing here.
	var matchStorage [64]byte
	match := matchStorage[:0]
	for index, files := range globs.overrides {
		if matchesAny(files, segments) {
			match = binary.AppendUvarint(match, uint64(index))
		}
	}

	memo := c.resolutions()
	memo.mutex.Lock()
	defer memo.mutex.Unlock()
	if resolved, found := memo.byMatch[string(match)]; found {
		return resolved
	}
	resolved := c.resolveMatch(match)
	memo.byMatch[string(match)] = resolved
	return resolved
}

// resolveMatch builds the resolution for one match: the base rules, then each override the match names,
// in order.
func (c *Config) resolveMatch(match []byte) Resolved {
	// The base map is copied rather than shared, because an override writes into the result and a
	// shared map would leak one file's overrides into every later file.
	effective := make(map[string]RuleSetting, len(c.Rules))
	for name, setting := range c.Rules {
		effective[name] = setting
	}

	for remaining := match; len(remaining) > 0; {
		index, width := binary.Uvarint(remaining)
		remaining = remaining[width:]
		override := c.Overrides[index]
		for name, setting := range override.Rules {
			// A bare severity re-states the rule's level and keeps its options, as ESLint does and as a
			// bare severity across `extends` already does. Dropping them changed what a rule checked on
			// every file the override matched (#na0hgjz).
			if setting.Options == nil {
				if inForce, configured := effective[name]; configured {
					setting.Options = inForce.Options
				}
			}
			effective[name] = setting
		}
	}

	return Resolved{Rules: effective, identity: &resolutionIdentity{match: string(match)}}
}

// relativePath expresses a path the way the config's patterns are written: relative to the config
// file, slash-separated, with no leading dot.
//
// Getting this wrong is silent in the dangerous direction. An absolute path never matches
// `modules/**`, so every override quietly stops applying and the rules they scoped off come back —
// which is exactly the 336 findings this package exists to remove.
func (c *Config) relativePath(path string) string {
	normalized := filepath.ToSlash(path)

	if c.Root != "" && filepath.IsAbs(path) {
		if relative, err := filepath.Rel(c.Root, path); err == nil {
			normalized = filepath.ToSlash(relative)
		}
	}

	return strings.TrimPrefix(normalized, "./")
}
