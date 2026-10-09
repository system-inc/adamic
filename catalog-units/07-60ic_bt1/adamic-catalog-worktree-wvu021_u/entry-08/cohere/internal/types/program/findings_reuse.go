package program

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/suppression"
)

// FindingsReuse is one run's use of the findings cache: what the last run recorded, and what this
// one records for the next.
//
// The walk consults it per file. A file whose bytes, key and applied cacheable rules all match an
// entry is walked with only its uncacheable rules, and the cacheable rules' findings and coverage
// are taken from the entry. Every other file is walked in full and, if eligible, recorded.
//
// Coverage is replayed along with findings, by ruling: coverage recorded over inputs proven
// unchanged is still true of the tree, the argument the run cache settled. A replay is never
// silent about itself either: Replayed counts the files served from cache, and the lint line says
// how many, so a walked verdict and a remembered one can always be told apart.
type FindingsReuse struct {
	key      [sha256.Size]byte
	previous *LintCache

	// anchor is the project root the design system's paths are anchored at. See PathAnchor.
	anchor PathAnchor

	// designFingerprint is the previous run's design system when every path it read still holds, so its
	// design-system rules' findings may replay; zero when it does not, or there was none.
	designFingerprint [sha256.Size]byte

	mutex    sync.Mutex
	next     map[string]LintCacheEntry
	replayed atomic.Int64

	// missingEntries, changedEntries and unmeasuredEntries count the files looked up and not replayed, by why:
	// no entry for the file, an entry at other bytes or under another rule list, or one taken by a run that
	// did not measure readiness. See Misses.
	missingEntries    atomic.Int64
	changedEntries    atomic.Int64
	unmeasuredEntries atomic.Int64

	// designReads and designLoaded are what this run's design system read, gathered from each walk. See
	// NoteDesignSystem.
	designReads  []rule.FileRead
	designLoaded bool

	// readiness is whether this run measures Adamic readiness. See MeasureReadiness.
	readiness bool
}

// MeasureReadiness says this run measures Adamic readiness, so an entry recorded by a run that did not is a
// miss: its file was never measured, and replaying it would read as measured and clean. Called before any
// worker runs.
func (r *FindingsReuse) MeasureReadiness() {
	if r != nil {
		r.readiness = true
	}
}

// NewFindingsReuse prepares a run's reuse. A previous cache under a different key is dropped at
// once rather than consulted per file: nothing in it can be valid. anchor is the project root, which the
// design system's paths are named relative to; the zero anchor names them as they are.
//
// The previous cache's lookup index is built here, before any worker runs. Lookup would otherwise
// build it on first use, and sixteen workers missing at once would race to build it.
func NewFindingsReuse(key [sha256.Size]byte, previous *LintCache, anchor PathAnchor) *FindingsReuse {
	if previous != nil && previous.Key != key {
		previous = nil
	}
	reuse := &FindingsReuse{key: key, previous: previous, anchor: anchor, next: map[string]LintCacheEntry{}}
	if previous != nil {
		previous.ensureIndex()
		// Checked once, here, by re-hashing what the design system read last time: a few stats and small reads,
		// against building the design system to find out.
		if previous.DesignSystem.stillHolds(anchor) {
			reuse.designFingerprint = previous.DesignSystem.Fingerprint
		}
	}
	return reuse
}

// NoteDesignSystem records what this run's design system read, from rule.DesignSystemReads after a walk. A run
// that walks more than once adds every walk's reads, and loaded is whether any walk loaded it at all.
func (r *FindingsReuse) NoteDesignSystem(reads []rule.FileRead, loaded bool) {
	if r == nil || !loaded {
		return
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.designLoaded = true
	r.designReads = append(r.designReads, reads...)
}

// Misses says why the files this run looked up and did not replay were not replayed, in words, empty when
// every one was or none was looked up. A run whose whole findings section was missing or under another key
// says so itself (the command does), since then no file was looked up in anything.
func (r *FindingsReuse) Misses() string {
	if r == nil {
		return ""
	}
	var reasons []string
	if count := r.missingEntries.Load(); count > 0 {
		reasons = append(reasons, fmt.Sprintf("%d with no entry", count))
	}
	if count := r.changedEntries.Load(); count > 0 {
		reasons = append(reasons, fmt.Sprintf("%d whose bytes or rules changed", count))
	}
	if count := r.unmeasuredEntries.Load(); count > 0 {
		reasons = append(reasons, fmt.Sprintf("%d taken without measuring readiness", count))
	}
	if len(reasons) == 0 {
		return ""
	}
	return "files not replayed: " + strings.Join(reasons, ", ")
}

// Replayed is how many files this run served from cache.
func (r *FindingsReuse) Replayed() int {
	if r == nil {
		return 0
	}
	return int(r.replayed.Load())
}

// Recorded is what this run leaves for the next: every file it served or walked eligibly, under
// this run's key. Files this run did not reach are dropped, so the cache never outgrows the tree.
//
// The design system is settled here, once the walks are done. Entries whose design-system rules ran this run
// carry a zero fingerprint until now; they take this run's. Entries that replayed theirs keep the previous one,
// unless this run loaded a design system that differs from it, and then every entry's design-system findings
// are dropped from replay, since which design system they saw can no longer be told.
func (r *FindingsReuse) Recorded() *LintCache {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	design, keepReplayed := r.settleDesignSystem()
	cache := &LintCache{Version: lintCacheVersion, Key: r.key, Entries: make([]LintCacheEntry, 0, len(r.next)), DesignSystem: design}
	for _, entry := range r.next {
		switch {
		case entry.DesignFingerprint == [sha256.Size]byte{} && design != nil:
			entry.DesignFingerprint = design.Fingerprint
		case !keepReplayed:
			entry.DesignFingerprint = [sha256.Size]byte{}
		}
		cache.Entries = append(cache.Entries, entry)
	}
	sort.Slice(cache.Entries, func(first, second int) bool { return cache.Entries[first].Path < cache.Entries[second].Path })
	return cache
}

// settleDesignSystem decides the design system this run's entries were produced under, and whether entries that
// replayed design-system findings under the previous one may keep them.
//
// A run that loaded the design system keys its reads, unless a path moved under it, and then nothing it
// produced is replayable. A run that never loaded it produced design-system findings that read no stylesheet,
// true under any; it carries the previous key when that still held, so the entries that replayed under it
// stay valid, and otherwise records the empty key.
func (r *FindingsReuse) settleDesignSystem() (key *DesignSystemKey, keepReplayed bool) {
	zero := [sha256.Size]byte{}
	if r.designLoaded {
		key, held := designSystemKeyFromReads(r.designReads, r.anchor)
		if !held {
			return nil, false
		}
		return key, r.designFingerprint == zero || r.designFingerprint == key.Fingerprint
	}
	if r.designFingerprint != zero {
		return r.previous.DesignSystem, true
	}
	return emptyDesignSystemKey(), true
}

// ShapeClasses splits CacheClasses' type-aware rules by what they can see of imported files: those that
// declare rule.TypeReachShapes, keyed on the shape fingerprint, and the rest, keyed on the type
// fingerprint.
func ShapeClasses(typeAware []rule.Rule) (shaped []rule.Rule, typed []rule.Rule) {
	for _, subject := range typeAware {
		if subject.TypeReach == rule.TypeReachShapes {
			shaped = append(shaped, subject)
		} else {
			typed = append(typed, subject)
		}
	}
	return shaped, typed
}

// cacheKeys is what a file's cached findings are looked up and recorded under: its bytes, the rules of
// each class that apply to it, and the two fingerprints the type-aware classes are keyed on.
type cacheKeys struct {
	contentHash      [sha256.Size]byte
	pure             []string
	typed            []string
	typeFingerprint  [sha256.Size]byte
	shaped           []string
	shapeFingerprint [sha256.Size]byte
	design           []string

	// derived is the rules declaring rule.ProgramFingerprint, and derivedFingerprint their key (derivedKey),
	// zero when one has no fingerprint this run.
	derived            []string
	derivedFingerprint [sha256.Size]byte
}

// lookup reports whether a file's pure rules can be replayed, and whether each keyed class can be too. All
// need the first, and each also needs its own rule list and its own fingerprint unchanged: an importer of an
// edited file has unchanged bytes and a changed type fingerprint, so its pure rules replay and its
// content-keyed rules run again, while its shape-keyed rules replay unless the edit changed a shape. The
// design-system rules replay while the design system still holds (NewFindingsReuse) and is the one the entry
// names; a file none of them applies to has nothing to replay or run.
//
// An entry whose withheld findings do not fit the file's directives as they read now, a directive index out of
// range, one that cannot silence the rule it is said to have silenced, or a rule the entry did not run, is not
// replayed: it is a miss, never repaired. text is the file's bytes, which a pure hit has already proven are the
// entry's.
func (r *FindingsReuse) lookup(path string, keys cacheKeys, text string) (entry LintCacheEntry, hits classHits) {
	entry, pureHit := r.previous.Lookup(path, keys.contentHash, r.key, keys.pure)
	if pureHit && !withheldFits(entry, text) {
		pureHit = false
	}
	switch {
	case !pureHit && r.previous.holds(path):
		r.changedEntries.Add(1)
	case !pureHit:
		r.missingEntries.Add(1)
	case r.readiness && entry.Adamic == nil:
		r.unmeasuredEntries.Add(1)
	}
	if !pureHit || r.readiness && entry.Adamic == nil {
		return LintCacheEntry{}, classHits{}
	}
	r.replayed.Add(1)
	r.keep(entry)
	return entry, classHits{
		pure:   true,
		typed:  equalStrings(entry.TypedRules, keys.typed) && entry.TypeFingerprint == keys.typeFingerprint,
		shaped: equalStrings(entry.ShapedRules, keys.shaped) && entry.ShapeFingerprint == keys.shapeFingerprint,
		design: equalStrings(entry.DesignRules, keys.design) && (len(keys.design) == 0 ||
			r.designFingerprint != [sha256.Size]byte{} && entry.DesignFingerprint == r.designFingerprint),
		derived: equalStrings(entry.DerivedRules, keys.derived) && (len(keys.derived) == 0 ||
			keys.derivedFingerprint != [sha256.Size]byte{} && entry.DerivedFingerprint == keys.derivedFingerprint),
	}
}

// withheldFits reports whether every finding an entry says a directive withheld names a rule the entry ran and a
// directive in text that could silence it.
func withheldFits(entry LintCacheEntry, text string) bool {
	if len(entry.Withheld) == 0 {
		return true
	}
	ran := make(map[string]bool, len(entry.Rules)+len(entry.TypedRules)+len(entry.ShapedRules)+len(entry.DesignRules)+
		len(entry.DerivedRules))
	for _, names := range [][]string{entry.Rules, entry.TypedRules, entry.ShapedRules, entry.DesignRules, entry.DerivedRules} {
		for _, name := range names {
			ran[name] = true
		}
	}
	directives := suppression.Build(text).Directives()
	for _, withheld := range entry.Withheld {
		if !ran[withheld.RuleName] || withheld.Directive < 0 || int(withheld.Directive) >= len(directives) ||
			!directives[withheld.Directive].Names(withheld.RuleName) {
			return false
		}
	}
	return true
}

// classHits is which of a file's classes replay: pure, the two type-aware ones, the design-system rules and the
// derived ones. The others need pure.
type classHits struct {
	pure, typed, shaped, design, derived bool
}

func (r *FindingsReuse) keep(entry LintCacheEntry) {
	r.mutex.Lock()
	r.next[entry.Path] = entry
	r.mutex.Unlock()
}

// ruleNames lists rules by name, in order.
func ruleNames(rules []rule.Rule) []string {
	names := make([]string, len(rules))
	for index, subject := range rules {
		names[index] = subject.Name
	}
	return names
}

// replayEntry turns a cached entry back into the walk's diagnostics and coverage for one file, and returns the
// replayed rules' notes.
//
// hits says which keyed classes are replayed as well as the pure one. A class that is not runs again on this
// file and counts its own coverage and notes, so replaying them here would count them twice.
func replayEntry(entry LintCacheEntry, hits classHits, sourceFile *ast.SourceFile, diagnostics *[]rule.Diagnostic,
	reporting map[string]int, offered map[string]int, listening map[string]int) RuleNotes {
	replays := make(map[string]bool, len(entry.Rules)+len(entry.TypedRules)+len(entry.ShapedRules)+len(entry.DesignRules)+
		len(entry.DerivedRules))
	for _, names := range [][]string{entry.Rules, classIf(hits.typed, entry.TypedRules), classIf(hits.shaped, entry.ShapedRules),
		classIf(hits.design, entry.DesignRules), classIf(hits.derived, entry.DerivedRules)} {
		for _, name := range names {
			replays[name] = true
			offered[name]++
		}
	}
	for _, name := range entry.Listening {
		if replays[name] {
			listening[name]++
		}
	}
	for _, finding := range entry.Findings {
		if !replays[finding.RuleName] {
			continue
		}
		*diagnostics = append(*diagnostics, rule.Diagnostic{
			RuleName:   finding.RuleName,
			Range:      core.NewTextRange(int(finding.Start), int(finding.End)),
			Message:    rule.Message{Id: finding.MessageId, Description: finding.MessageDescription},
			SourceFile: sourceFile,
		})
		reporting[finding.RuleName]++
	}
	return notesOf(entry.Notes, replays)
}

// replayedAdamic is the part of an entry's readiness record that its replayed classes produced, nil when the
// entry has none. The walk adds what the rules it ran measured.
func replayedAdamic(entry LintCacheEntry, hits classHits) *AdamicRecord {
	return adamicOf(entry.Adamic, ruleSet(entry.Rules, classIf(hits.typed, entry.TypedRules),
		classIf(hits.shaped, entry.ShapedRules), classIf(hits.design, entry.DesignRules), classIf(hits.derived, entry.DerivedRules)))
}

// ruleSet is every name in the lists, as a set.
func ruleSet(lists ...[]string) map[string]bool {
	names := map[string]bool{}
	for _, list := range lists {
		for _, name := range list {
			names[name] = true
		}
	}
	return names
}

// classIf is names when the class is included, and nothing otherwise.
func classIf(included bool, names []string) []string {
	if included {
		return names
	}
	return nil
}

// notesOf is the notes of the named rules, nil when none of them noted anything. The per-key counts are
// shared, not copied: nothing writes to a file's notes once its walk has returned them.
func notesOf(notes RuleNotes, names map[string]bool) RuleNotes {
	var kept RuleNotes
	for ruleName, counts := range notes {
		if !names[ruleName] {
			continue
		}
		if kept == nil {
			kept = RuleNotes{}
		}
		kept[ruleName] = counts
	}
	return kept
}

// mergeNotes is the notes of two disjoint sets of rules together, nil when neither has any.
func mergeNotes(first RuleNotes, second RuleNotes) RuleNotes {
	if len(first) == 0 {
		return second
	}
	if len(second) == 0 {
		return first
	}
	merged := make(RuleNotes, len(first)+len(second))
	for _, notes := range []RuleNotes{first, second} {
		for ruleName, counts := range notes {
			merged[ruleName] = counts
		}
	}
	return merged
}

// recordableEntry builds the entry a fully walked file would leave, or reports that it must not leave
// one.
//
// A file is not recorded when any cacheable rule's finding carries a fix or a suggestion, since a cached
// finding stores neither and the fix phase needs the edit itself. A file with directives is recorded with the
// findings they withheld from its cacheable rules, which a replay marks applied again (#kdee854): a directive's
// accounting spans every rule in the file, cached and walked alike, and those are what the cached part adds.
//
// fileAdamic is what the walk measured for readiness, nil when it measured nothing; only the cacheable rules'
// part is kept, since the rest run on every walk.
func recordableEntry(sourceFile *ast.SourceFile, keys cacheKeys, fileDiagnostics []rule.Diagnostic,
	fileListening map[string]int, fileNotes RuleNotes, fileAdamic *AdamicRecord, visited int, silenced suppressionTally) (LintCacheEntry, bool) {
	cacheable := append(append(append(append(append([]string{}, keys.pure...), keys.typed...), keys.shaped...), keys.design...),
		keys.derived...)
	isCacheable := make(map[string]bool, len(cacheable))
	for _, name := range cacheable {
		isCacheable[name] = true
	}
	entry := LintCacheEntry{
		Path:               sourceFile.FileName().AsString(),
		ContentHash:        keys.contentHash,
		Rules:              keys.pure,
		TypedRules:         keys.typed,
		TypeFingerprint:    keys.typeFingerprint,
		ShapedRules:        keys.shaped,
		ShapeFingerprint:   keys.shapeFingerprint,
		DesignRules:        keys.design,
		DerivedRules:       keys.derived,
		DerivedFingerprint: keys.derivedFingerprint,
		VisitedNodes:       visited,
		Notes:              notesOf(fileNotes, isCacheable),
		Directives:         silenced.directives,
		Withheld:           withheldBy(silenced.withheld, isCacheable),
		Adamic:             adamicOf(fileAdamic, isCacheable),
	}
	for _, name := range cacheable {
		if fileListening[name] > 0 {
			entry.Listening = append(entry.Listening, name)
		}
	}
	for _, diagnostic := range fileDiagnostics {
		if !isCacheable[diagnostic.RuleName] {
			continue
		}
		if len(diagnostic.Fixes) > 0 || len(diagnostic.Suggestions) > 0 {
			return LintCacheEntry{}, false
		}
		entry.Findings = append(entry.Findings, LintCacheFinding{
			RuleName:           diagnostic.RuleName,
			Start:              int32(diagnostic.Range.Pos()),
			End:                int32(diagnostic.Range.End()),
			MessageId:          diagnostic.Message.Id,
			MessageDescription: diagnostic.Message.Description,
		})
	}
	return entry, true
}

// withheldBy is the withheld findings of the named rules, nil when none.
func withheldBy(withheld []LintCacheWithheld, names map[string]bool) []LintCacheWithheld {
	var kept []LintCacheWithheld
	for _, finding := range withheld {
		if names[finding.RuleName] {
			kept = append(kept, finding)
		}
	}
	return kept
}

// refreshClasses is an entry whose pure part was replayed and one or more of whose keyed classes just ran
// again: the old findings, listening and notes of every class that replayed, and the fresh ones of every class
// that ran, under this run's fingerprints. Refused, like any recording, when a fresh finding carries a fix or a
// suggestion.
//
// A design-system class that ran takes a zero fingerprint, settled to this run's design system in Recorded.
//
// The readiness record merges the same way, rule by rule, and is nil when this walk measured nothing: an entry
// half measured is unmeasured.
func refreshClasses(old LintCacheEntry, keys cacheKeys, hits classHits, fileDiagnostics []rule.Diagnostic,
	fileListening map[string]int, fileNotes RuleNotes, fileAdamic *AdamicRecord, silenced suppressionTally) (LintCacheEntry, bool) {
	ran := map[string]bool{}
	for _, name := range append(append(append(append([]string{}, classIf(!hits.typed, keys.typed)...), classIf(!hits.shaped, keys.shaped)...),
		classIf(!hits.design, keys.design)...), classIf(!hits.derived, keys.derived)...) {
		ran[name] = true
	}
	kept := make(map[string]bool, len(old.Rules)+len(old.TypedRules)+len(old.ShapedRules)+len(old.DesignRules)+len(old.DerivedRules))
	for _, name := range append(append(append(append(append([]string{}, old.Rules...), classIf(hits.typed, old.TypedRules)...),
		classIf(hits.shaped, old.ShapedRules)...), classIf(hits.design, old.DesignRules)...), classIf(hits.derived, old.DerivedRules)...) {
		kept[name] = true
	}
	refreshed := LintCacheEntry{
		Path:               old.Path,
		ContentHash:        old.ContentHash,
		Rules:              old.Rules,
		TypedRules:         keys.typed,
		TypeFingerprint:    keys.typeFingerprint,
		ShapedRules:        keys.shaped,
		ShapeFingerprint:   keys.shapeFingerprint,
		DesignRules:        keys.design,
		DerivedRules:       keys.derived,
		DerivedFingerprint: keys.derivedFingerprint,
		VisitedNodes:       old.VisitedNodes,
		Notes:              mergeNotes(notesOf(old.Notes, kept), notesOf(fileNotes, ran)),
		Directives:         old.Directives,
		Withheld:           append(withheldBy(old.Withheld, kept), withheldBy(silenced.withheld, ran)...),
		Adamic:             mergeAdamic(adamicOf(old.Adamic, kept), adamicOf(fileAdamic, ran)),
	}
	if hits.design {
		refreshed.DesignFingerprint = old.DesignFingerprint
	}
	for _, name := range old.Listening {
		if kept[name] {
			refreshed.Listening = append(refreshed.Listening, name)
		}
	}
	for _, names := range [][]string{keys.typed, keys.shaped, keys.design, keys.derived} {
		for _, name := range names {
			if ran[name] && fileListening[name] > 0 {
				refreshed.Listening = append(refreshed.Listening, name)
			}
		}
	}
	for _, finding := range old.Findings {
		if kept[finding.RuleName] {
			refreshed.Findings = append(refreshed.Findings, finding)
		}
	}
	for _, diagnostic := range fileDiagnostics {
		if !ran[diagnostic.RuleName] {
			continue
		}
		if len(diagnostic.Fixes) > 0 || len(diagnostic.Suggestions) > 0 {
			return LintCacheEntry{}, false
		}
		refreshed.Findings = append(refreshed.Findings, LintCacheFinding{
			RuleName:           diagnostic.RuleName,
			Start:              int32(diagnostic.Range.Pos()),
			End:                int32(diagnostic.Range.End()),
			MessageId:          diagnostic.Message.Id,
			MessageDescription: diagnostic.Message.Description,
		})
	}
	return refreshed, true
}
