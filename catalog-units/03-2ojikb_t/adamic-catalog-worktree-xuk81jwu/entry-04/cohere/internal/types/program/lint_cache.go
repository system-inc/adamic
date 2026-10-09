package program

import (
	"crypto/sha256"
	"sort"

	"github.com/system-inc/cohere/internal/lint/rule"
)

// The lint findings cache: what each rule reported about one file, replayed when neither the
// file nor the rule set has changed.
//
// This is the cheapest of the three caches to make correct, and the reason is worth stating
// because it does not hold for the other two. A pure rule's findings are a function of exactly
// one file's bytes. Same bytes, same rule, same findings, with nothing else in the world able
// to change the answer. So the key is a content hash and there is no invalidation graph.
//
// The word doing the work there is "pure". A rule that reads other files is not, and caching
// it under one file's hash serves a stale result forever when the OTHER file changes. Two live
// rules are cross-file today, and a third case is latent: the upstream adapter hands the whole
// program to every rule it wraps and cannot inspect what it wrapped, so it can never truthfully
// declare purity for a body it does not own.
//
// That is what CacheableRules exists to decide, and why it defaults to excluding rather than
// including. The asymmetry is the same one the codebase already chose for the type checker:
// under-declaring serves a stale finding silently and forever, over-declaring costs a cache
// miss. One of those is a correctness failure and the other is a performance one.
//
// It is wired as the run cache's second layer: when an eligible run's inputs are not all unchanged, files
// whose bytes and configuration are unchanged are walked with only their uncacheable rules, and the
// cacheable rules' findings and coverage are replayed. See FindingsReuse.
//
// It was measured and left unwired once, and the reversal is worth keeping because the measurement
// method is the same and only the tree changed. On 2026-08-25, with 212 rules, walking with only the
// uncacheable ones saved about 9%: the walk visits every node whatever the rule set, so with light
// rules there was little left to skip. On 2026-10-02, with 464 rules, the same experiment on ahra's real
// config, each walk on a freshly built graph:
//
//	walk, all 464 rules                     about 2.56s
//	walk, the 45 ReadsProgram rules only    about 1.37s
//
// The second line was the ceiling for a fully cached run, since those rules always ran. #b8k3bp6
// split ReadsProgram into ProgramReads by kind, and only the 9 rules reading other files still always
// run: about 0.23s on 2026-10-02, best of 3 on fresh graphs, against 1.35s for the 45 the same day.
// The 1.37s was first published here as 0.4s, measured on a graph whose earlier walks had already
// paid for type checking: the checker computes types lazily and whichever walk asks first pays, so a
// later walk looks cheap for a reason that has nothing to do with its rules. Every walk is now on a fresh graph. A real
// one-file-changed run on ahra saved about 1.1s of the fix phase, close to that ceiling.
//
// The node count is identical across rule sets, so the walk itself still cannot be skipped; what grew
// is the work rules do at each node. Type-aware rules are cached on a type fingerprint as well as the
// file's bytes; see CacheClasses and Graph.TypeFingerprints.
//
// Two kinds of file are never served from cache. A file with any suppression directive, because a
// directive's accounting spans every rule in the file, cached and walked alike. A file with any
// finding that carries a fix or a suggestion, because a cached finding stores neither and the fix
// phase needs the edit itself.
type LintCache struct {
	// Version is the format. A different version is a miss, never a best-effort read.
	Version int

	// Key covers everything a cacheable rule's answer depends on beyond the file's own bytes: the
	// binary, the lint config, the tsconfig chain, the project root and the rule set in order. Any
	// change to one invalidates every entry at once, which is what makes it a miss rather than a
	// stale pass. It was called RuleSetHash when the rule set was all it covered.
	Key [sha256.Size]byte

	// KeyParts is Key's parts, so a run under another key can say which part moved. Zero when the run that
	// recorded it kept none. See RunCacheKeyParts.
	KeyParts RunCacheKeyParts

	// Entries is one record per cached file.
	Entries []LintCacheEntry

	// DesignSystem is the design system the design-system rules' findings were produced under, nil when none
	// was recorded. See DesignSystemKey.
	DesignSystem *DesignSystemKey

	// index maps path to position in Entries, built lazily on the first Lookup. Not serialized:
	// it is derived from Entries and rebuilding it costs less than storing it.
	index map[string]int
}

// LintCacheEntry is one file's findings, keyed by what the file contained.
type LintCacheEntry struct {
	// Path identifies the file. Kept for debugging and for eviction; the hash is what decides
	// a hit.
	Path string

	// ContentHash is the hash of the file's bytes when these findings were produced.
	ContentHash [sha256.Size]byte

	// Rules is the cacheable rules the configuration applied to this file, in order. A hit requires
	// the same list now, so an override that changes which rules reach this file is a miss even under
	// an unchanged key.
	Rules []string

	// TypedRules is the type-aware rules applied to this file, in order. Their findings depend on the
	// types the file can see as well as its bytes, so they replay only while TypeFingerprint matches
	// too. See Graph.TypeFingerprints.
	TypedRules []string

	// TypeFingerprint is the file's type fingerprint when TypedRules' findings were produced.
	TypeFingerprint [sha256.Size]byte

	// ShapedRules is the type-aware rules applied to this file that declare rule.TypeReachShapes. Their
	// findings replay while ShapeFingerprint matches, which an edit inside an imported function's body does
	// not move. See Graph.SignatureFingerprints.
	ShapedRules []string

	// ShapeFingerprint is the file's shape fingerprint when ShapedRules' findings were produced.
	ShapeFingerprint [sha256.Size]byte

	// DesignRules is the rules applied to this file that read the design system (rule.ReadsDesignSystem). Their
	// findings replay while the design system is the one DesignFingerprint names. See DesignSystemKey.
	DesignRules []string

	// DesignFingerprint is the design system's fingerprint when DesignRules' findings were produced, zero when
	// they cannot be replayed under any.
	DesignFingerprint [sha256.Size]byte

	// DerivedRules is the rules applied to this file that declare rule.ProgramFingerprint. Their findings replay
	// while DerivedFingerprint matches.
	DerivedRules []string

	// DerivedFingerprint hashes each DerivedRules rule's name and program fingerprint, and the file's type
	// fingerprint when any of them reads types, as they were when DerivedRules' findings were produced. Zero
	// when one had no fingerprint that run, which no lookup matches.
	DerivedFingerprint [sha256.Size]byte

	// Listening is the subset of Rules and TypedRules that registered a listener on this file. A replay
	// counts them as listening, which is what keeps the coverage line identical to a walked run's.
	Listening []string

	// VisitedNodes is how many nodes the full walk of this file visited. The walk counts nodes only
	// when some rule listens, so a cached file whose remaining rules listen to nothing would otherwise
	// contribute zero and change the coverage line.
	VisitedNodes int

	// Notes is what the cacheable rules noted through rule.Context.Note, per rule and per key, nil when none
	// did. A replay restores the notes of the rules it replays, so a warm run counts them as a walked one does.
	Notes RuleNotes

	// Findings is what the cacheable rules reported. Empty is a real answer: it means the rules ran
	// and found nothing, which is exactly the case worth caching since most files are clean.
	Findings []LintCacheFinding

	// Directives is whether the file has a suppression directive, or names a rule in a disable or enable comment.
	// A replay of such a file still reads its comments, so its directives are tallied and its unknown rule names
	// reported as a walk's are.
	Directives bool

	// Withheld is the cacheable rules' findings a directive silenced, each as its rule and the directive's index
	// (suppression.Index.SuppressedBy). A replay marks the directive applied for each one whose rule it replays,
	// so a directive is used or unused exactly as a walk finds it. The bytes are the entry's, so the indices name
	// the same directives (#kdee854).
	Withheld []LintCacheWithheld

	// Adamic is what the cacheable cohere:adamic rules measured on this file for Adamic readiness, nil when
	// the run that produced the entry did not measure readiness. Nil is not "ready": a run that measures
	// readiness misses on it (FindingsReuse.lookup), since replaying it would read every unmeasured file as
	// having no findings, which is the silent clean this cache must never produce (#drbrp8c).
	Adamic *AdamicRecord
}

// LintCacheWithheld is one cacheable rule's finding a directive silenced: the rule, and the directive's index among
// the file's directives.
type LintCacheWithheld struct {
	RuleName  string
	Directive int32
}

// AdamicRecord is one file's Adamic readiness measurement, kept per rule so an entry whose type-aware rules
// ran again can be merged rule by rule (refreshClasses), as its findings are.
//
// Measure-only findings are counts, never LintCacheFindings: they cannot be printed, cannot fail a run and
// carry no fixes, so nothing about them needs replaying beyond how many there were.
type AdamicRecord struct {
	// Counts is every cacheable cohere:adamic rule that ran on the file with how many findings it had at the
	// set's own options, before suppression, sorted by rule. A rule that found nothing has a zero pair, so a
	// rule that ran clean and one that did not run can be told apart.
	Counts []AdamicCount

	// Skipped is the cacheable cohere:adamic rules that skipped the file, sorted. A skip leaves the file
	// unmeasured, not ready.
	Skipped []string
}

// AdamicCount is one rule's measure-only finding count on one file.
type AdamicCount struct {
	Rule     string
	Findings int32
}

// adamicOf is the part of a record that the named rules produced, nil when the record is nil. A record that
// keeps nothing is empty, not nil: the file was measured and those rules simply do not apply.
func adamicOf(record *AdamicRecord, names map[string]bool) *AdamicRecord {
	if record == nil {
		return nil
	}
	kept := &AdamicRecord{}
	for _, count := range record.Counts {
		if names[count.Rule] {
			kept.Counts = append(kept.Counts, count)
		}
	}
	for _, name := range record.Skipped {
		if names[name] {
			kept.Skipped = append(kept.Skipped, name)
		}
	}
	return kept
}

// mergeAdamic is two records of disjoint rules as one, sorted, nil when either is nil: a merge with an
// unmeasured half is unmeasured.
func mergeAdamic(first *AdamicRecord, second *AdamicRecord) *AdamicRecord {
	if first == nil || second == nil {
		return nil
	}
	merged := &AdamicRecord{
		Counts:  append(append([]AdamicCount{}, first.Counts...), second.Counts...),
		Skipped: append(append([]string{}, first.Skipped...), second.Skipped...),
	}
	sort.Slice(merged.Counts, func(left, right int) bool { return merged.Counts[left].Rule < merged.Counts[right].Rule })
	sort.Strings(merged.Skipped)
	if len(merged.Counts) == 0 {
		merged.Counts = nil
	}
	if len(merged.Skipped) == 0 {
		merged.Skipped = nil
	}
	return merged
}

// LintCacheFinding is one finding, flattened.
//
// rule.Diagnostic carries a *ast.SourceFile belonging to a program that no longer exists by
// the time this is read back, so the file is implied by the entry rather than stored per
// finding.
//
// MessageDescription holds the rendered text. An earlier version of this comment said text was
// "derived from the id when a finding is rendered" and therefore safe to omit. That was untrue,
// and the field was absent because of it. Nothing in this tree derives text from an id: there is
// no id-to-text table, and `printRuleDiagnostic` prints `Message.Description` directly. Measured
// 2026-08-25 over `internal/rules`: 340 message ids, of which 11 build their description with
// `fmt.Sprintf` from runtime values, so for those one id maps to many strings and no table could
// exist. A cached finding replaying only the id would print an empty or wrong sentence beside a
// correct file, range and rule — the right metadata carrying the wrong message, which is the one
// failure here that looks exactly like success.
//
// The text is stored rather than the 11 being refused because refusing them buys bytes with a
// standing per-rule obligation, and this format already depends on `ProgramReads` carrying one.
// The size scales the right way: the artifact grows with a dirty tree, which is the run where
// this cache saves least, because changed files are recomputed anyway.
//
// Message ids are not unique on their own, which is a second reason the id could never have been
// the key. `unexpected` carries five distinct descriptions across six core rules, colliding across
// rules and never within one, so the honest identity of a message is the pair (RuleName,
// MessageId). Lookups key on the pair even though the text is now stored, so nobody later assumes
// ids are unique.
//
// Fixes and suggestions are counted rather than stored, and that omission IS sound: a fix is a
// replacement over a byte range, and replaying one computed against different bytes would corrupt
// the file it claims to repair. A cached finding is a report, never an edit.
type LintCacheFinding struct {
	RuleName string
	Start    int32
	End      int32

	// MessageId identifies the message within its rule. Not unique across rules; see above.
	MessageId string

	// MessageDescription is the rendered sentence, stored because nothing can rebuild it.
	MessageDescription string

	FixCount        int32
	SuggestionCount int32
}

// MessageKey is the honest identity of a message: the pair, never the id alone.
func (f LintCacheFinding) MessageKey() (string, string) {
	return f.RuleName, f.MessageId
}

// HashContent is the key a cached entry is stored under.
func HashContent(content string) [sha256.Size]byte {
	return sha256.Sum256([]byte(content))
}

// HashRuleSet covers the names of every rule that ran, in order.
//
// Order is included deliberately, but not for the reason first written here. That reason was:
// "the order they ran in is the order the findings come back in". It is false. Findings come
// back in file-then-position order, which the walk imposes after its workers finish
// (sortDiagnostics in walk.go); rule order never decided it.
//
// The conclusion survives the correction on different grounds. A reordered rule set can change
// which findings EXIST rather than merely their sequence, because fixes applied by an earlier
// rule change the text a later rule reads, and suppression directives are consumed in the order
// rules run. So a cache keyed without rule order could replay a set the current configuration
// would not produce, which is what this hash prevents.
//
// For anyone wiring this cache: LintCacheEntry.Findings is an ordered slice, and replayed findings
// must reach the caller through the same sortDiagnostics the walk uses, so a warm run and a cold
// run return one sequence and cache state cannot decide it.
func HashRuleSet(ruleNames []string) [sha256.Size]byte {
	hash := sha256.New()
	for _, name := range ruleNames {
		hash.Write([]byte(name))
		hash.Write([]byte{0})
	}
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result
}

// lintCacheVersion is bumped whenever the format's meaning changes.
//
// 10: a file whose directives silenced something or went unused has an entry, which carries what its cacheable
// rules' findings the directives withheld. A version 9 cache has no such file, and none of its entries says
// whether its file has directives.
//
// 9: entries carry what the cacheable cohere:adamic rules measured for Adamic readiness. A version 8 entry has
// no record, and one read as nil-but-current would be indistinguishable from an entry recorded by a run that
// did not measure.
//
// 8: entries carry their cacheable rules' notes. A version 7 entry has none, and replaying it would count
// none where a walk counts some.
//
// 7: entries carry design-system rules and the design system they were produced under, and the cache carries
// that design system's read set.
//
// 6: entries carry shape-keyed type-aware rules and the shape fingerprint they were produced under.
//
// 5: entries carry type-aware rules and the type fingerprint they were produced under.
//
// 4: rule lists are stored once and referenced by index, and hashes are hex. Version 3 wrote each
// entry's applied and listening rules out in full; on ahra 3,605 entries shared 4 distinct rule lists
// and 24 listening lists, and the file was 53 MB and 113ms to decode, a tenth of what the cache saves.
// Versions 1 and 2 were a binary layout of offsets into an interned blob, replaced because adding
// fields to it was where its "an encoder forgot a field" bug had shipped four times.
//
// The encoding itself is no longer this version's business. The cache is the findings section of the
// cache table (cache_table.go), encoded with gob behind the table's header, so a change to the encoded
// shape moves cacheTableVersion. This one moves when what an entry means changes.
const lintCacheVersion = 10

// Lookup returns a file's cached entry, and whether the cache had a usable answer.
//
// The second return distinguishes "cached, no findings" from "not cached", and that distinction is
// the whole correctness of this cache: a clean file and an unknown file both produce no findings, and
// treating them the same is how a cache silently reports a clean tree.
//
// A hit needs the key, the path, the bytes and the applied rule list all to match. The index is built
// on first use, so a caller that loads a cache and discards it on a key mismatch never pays for it.
func (c *LintCache) Lookup(path string, contentHash [sha256.Size]byte, key [sha256.Size]byte, rules []string) (LintCacheEntry, bool) {
	if c == nil || c.Key != key {
		return LintCacheEntry{}, false
	}
	c.ensureIndex()
	position, found := c.index[path]
	if !found {
		return LintCacheEntry{}, false
	}
	entry := c.Entries[position]
	// The path matching is not enough. A file whose contents changed has a stale entry under the same
	// path, and returning it is precisely the silent failure this cache must not have.
	if entry.ContentHash != contentHash || !equalStrings(entry.Rules, rules) {
		return LintCacheEntry{}, false
	}
	return entry, true
}

// holds reports whether the cache has an entry for path at all, matching or not, so a miss can say whether the
// file was never recorded or recorded at other bytes. The index must already be built (ensureIndex).
func (c *LintCache) holds(path string) bool {
	if c == nil || c.index == nil {
		return false
	}
	_, found := c.index[path]
	return found
}

// ensureIndex builds the path index if it is missing. A caller about to share the cache across
// goroutines calls it first, since building it lazily inside concurrent lookups would race.
func (c *LintCache) ensureIndex() {
	if c.index != nil {
		return
	}
	c.index = make(map[string]int, len(c.Entries))
	for position := range c.Entries {
		c.index[c.Entries[position].Path] = position
	}
}

func equalStrings(first []string, second []string) bool {
	if len(first) != len(second) {
		return false
	}
	for index := range first {
		if first[index] != second[index] {
			return false
		}
	}
	return true
}

// Store records one file's findings, replacing any earlier entry for the same path.
//
// Replace rather than append, because a path appearing twice makes Lookup's answer depend on which
// copy the index happened to keep. The index is invalidated rather than patched: Store is called
// once per file during a walk and Lookup runs afterward, so rebuilding once on the next lookup
// costs less than maintaining the map through every insert.
//
// Empty findings are stored deliberately. A clean file is the most valuable thing this cache holds,
// because most files are clean, and an entry saying "these bytes produce nothing" is a real answer.
// Skipping it would make a clean file indistinguishable from an unknown one, which is the exact
// ambiguity Lookup's second return exists to destroy.
func (c *LintCache) Store(entry LintCacheEntry) {
	if c == nil {
		return
	}
	if c.index != nil {
		if position, found := c.index[entry.Path]; found {
			c.Entries[position] = entry
			return
		}
	}
	for position := range c.Entries {
		if c.Entries[position].Path == entry.Path {
			c.Entries[position] = entry
			c.index = nil
			return
		}
	}
	c.Entries = append(c.Entries, entry)
	c.index = nil
}

// CacheClasses splits a rule set five ways, in order: rules whose findings depend only on the file's
// bytes, type-aware rules whose findings also depend on the types the file can see, design-system rules
// whose findings also depend on the stylesheets the design system read, derived rules whose findings also
// depend on a program-wide fingerprint they declare, and rules that may never be cached.
//
// ReadsOtherFiles is the one disqualification left, unless the rule declares rule.ProgramFingerprint. A rule
// reading the program's file list, a file by name, another file's imports, or the file system reaches
// something a per-file key cannot name, so its answer can change while every key it could have is unmoved.
// A rule that names it, as a hash of exactly the program-wide data its verdict reads, is derived: keyed on
// that hash beside its bytes, and on the type fingerprint too when it reads types (#kdee854). Without one it
// stays uncacheable, so the default is the sound one. The program's other reads are in
// the key already: compiler options through the tsconfig chain, the default library through the
// binary, and this file's own module resolution through the type fingerprint, whose edges come from
// it. So a rule reading module resolution is type-aware even with no checker (#b8k3bp6).
//
// NeedsTypeChecker used to disqualify too, on the reasoning that a type-aware rule's answer depends on
// what the file imports. It does, and that is now in its key: Graph.TypeFingerprints hashes a file's
// import closure and everything global, so an edit to anything the file's types come from changes
// the fingerprint and re-runs the rule. Measured on ahra before the change: 164 of the 172 rules this
// excluded were excluded for the type checker alone.
//
// ReadsDesignSystem is keyed on the design system's read set, everything its load asked the file system about,
// observed rather than listed (#pyhm2t2, #35nqkwc). A rule reading the design system and also the types is
// excluded rather than given a fifth class: none does, and a key over both would need both to hold.
//
// The asymmetry still holds. Including a rule wrongly serves stale findings silently and forever;
// excluding one wrongly costs a cache miss. A rule that reads other files and the checker both is
// excluded, since other files are the larger reach.
func CacheClasses(rules []rule.Rule) (pure []rule.Rule, typeAware []rule.Rule, design []rule.Rule, derived []rule.Rule,
	uncacheable []rule.Rule) {
	for _, subject := range rules {
		readsTypes := subject.NeedsTypeChecker || subject.ProgramReads&rule.ReadsModuleResolution != 0
		readsDesignSystem := subject.ProgramReads&rule.ReadsDesignSystem != 0
		readsOtherFiles := subject.ProgramReads&rule.ReadsOtherFiles != 0
		switch {
		case readsOtherFiles && subject.ProgramFingerprint != nil && !readsDesignSystem:
			derived = append(derived, subject)
		case readsOtherFiles, readsDesignSystem && readsTypes:
			uncacheable = append(uncacheable, subject)
		case readsDesignSystem:
			design = append(design, subject)
		case readsTypes:
			typeAware = append(typeAware, subject)
		default:
			pure = append(pure, subject)
		}
	}
	return pure, typeAware, design, derived, uncacheable
}

// CacheableRules splits a rule set into the rules whose findings depend on nothing but the file's
// bytes, and everything else. See CacheClasses for the three-way split the walk uses.
func CacheableRules(rules []rule.Rule) (cacheable []rule.Rule, uncacheable []rule.Rule) {
	pure, typeAware, design, derived, never := CacheClasses(rules)
	return pure, append(append(append(typeAware, design...), derived...), never...)
}
