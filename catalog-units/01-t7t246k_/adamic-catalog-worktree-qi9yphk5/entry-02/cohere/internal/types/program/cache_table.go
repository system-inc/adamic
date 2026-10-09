package program

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/system-inc/cohere/internal/replace"
)

// The cache table: everything cohere keeps between runs for one project root, in one directory.
//
// It holds the run cache's recorded runs, one per invocation, the findings cache's per-file entries, the
// shapes, the format record and the types section. Each section keeps its own key and its own proof, and
// each lives in a file of its own with its own header, so a run reads and writes only the sections it
// uses (#45ekc65).
//
// # Why a file per section
//
// They shared one file until 2026-10-04, and replacing any section re-encoded all of them. Measured on
// ahra's 16 MB table: the format record encodes alone in 0.5ms, and saving it cost a 15ms read and a 21ms
// write of everything else. Three quarters of the bytes were the recorded runs, about 38,000 inputs per
// invocation, which only a replay reads, and only its own. Since every section carries its own proof, no
// section depends on another being written in the same instant, so the files need no atomicity across
// them: each is written whole and renamed into place, and the table's lock (command/cohere) keeps a reader
// from seeing one writer's run beside another's findings.
//
// # Why gob
//
// The two JSON files were 7.9 MB and 1.45 MB on ahra, decoding in 14ms and 3ms and encoding in 7ms and
// 25ms. Measured on #6tfmk0b: decode was about 6% of a replay, so speed was not the reason. The reason is
// that gob is written by reflection over the structs themselves. The lint cache's history is an encoder
// that forgot a field, shipped four times across a hand-written binary layout and a JSON one, and a
// format nobody writes by hand cannot forget a field.
//
// # The one way gob can lie, and the guard
//
// gob matches fields by name and skips what it does not recognize, in both directions. A field renamed,
// added or removed decodes without error into a zero value, which is the silent-default failure this
// format was chosen to escape, arriving by another door. So every file's header carries cacheTableVersion,
// and TestCacheTableShapeIsPinnedToItsVersion fails whenever the shape of what is encoded changes and the
// version does not. A changed shape always arrives as a discard, never as a partial read.
//
// # Discard, never repair
//
// Each file's header names the format, the cohere commit, the compiler commit, the Go toolchain, the
// platform that wrote it, and the section it holds. Any of them differing, a decode error, or a byte left
// over after the body is a discard of that file alone: the run reads the section as empty, says so once on
// stderr, and its next write of the section replaces the file. A discard costs one ordinary run. A repaired
// cache that is wrong costs a green run over a broken tree. The single file the table used to be,
// table.gob, is removed on sight and never read.
//
// One exception, and it is narrow. A format record written by another cohere commit, with the same format,
// compiler, toolchain and platform, is kept, while every other file from that commit is dropped. The format
// record is keyed by the formatter's own identity (FormatSection.Key), which moves only when the formatter
// can print differently, so a cohere commit that never touched the formatter need not reformat every file.
// The encoding is still guaranteed by the format version, so nothing here reads a field from the wrong
// place.
type CacheTable struct {
	// Runs is one recorded run per invocation, keyed by CacheTableInvocation. A bare run and `--no-fix`
	// print different reports under different keys, and kept apart they never overwrite each other.
	Runs map[string]*RunCache

	// Findings is the per-file findings cache, nil until a run has walked.
	Findings *LintCache

	// Signatures is each project file's shape, recorded so the next run computes only the files whose
	// bytes changed. Nil until a run has needed shapes.
	Signatures map[string]SignatureEntry

	// Formatted is the format record: which bytes cohere has already seen formatted, so a run formats
	// what changed since cohere last looked. Nil until a run has formatted. Its meaning is the format
	// phase's (command/cohere/format_record.go); this file only keeps it.
	Formatted *FormatSection

	// Types is each project file's semantic diagnostics under its shape fingerprint, for the types phase to
	// replay. Nil until a run has recorded them. See TypesSection.
	Types *TypesSection
}

// FormatSection is the format record's section of the table. Key names the binary and root its entries
// were recorded under; a section under any other key says nothing, since a rebuilt formatter may print
// differently.
type FormatSection struct {
	Key     string
	Entries map[string]FormatEntry
}

// FormatEntry is one file's format status: Sum is the SHA-256 of a text the formatter left unchanged
// under the options Options fingerprints, and Size and ModifiedNanoseconds are the file's signature
// when its bytes last matched Sum, so an unchanged file is a stat rather than a read.
type FormatEntry struct {
	Sum                 string
	Options             string
	Size                int64
	ModifiedNanoseconds int64
}

// CacheTableIdentity is what built the binary that writes or reads a table. It goes in the header, so a
// table written by any other build is discarded before a byte of its body is trusted.
type CacheTableIdentity struct {
	SelfCommit     string
	CompilerCommit string
	GoToolchain    string
	Platform       string
}

// cacheTableVersion is the format, and it moves whenever the encoded shape or its meaning does. The shape
// half is enforced by TestCacheTableShapeIsPinnedToItsVersion; the meaning half is the reason each
// section also keeps its own version.
//
// 15: findings entries carry whether their file has directives, and the findings those directives withheld.
//
// 14: findings entries carry derived rules and their fingerprint (#kdee854).
//
// 13: a signature entry carries Abandoned, set on a file whose emit outran the post-verdict graduation, so it is
// never emitted again while its bytes are the same (#9knyr86).
//
// 12: paths below the project root are stored relative to it, and every fingerprint hashes them that way, so
// every spelling of the root reads one table; recorded runs and the findings section carry their key's parts,
// so a miss can say which moved (#547dhjz).
//
// 11: findings entries carry their Adamic readiness record.
//
// 10: recorded runs carry the summary a replay renders its footer from.
//
// 9: run inputs carry ExistenceOnly, for a directory the run only asked whether it exists.
//
// 8: one file per section and per recorded run, each with its own header, which names its section.
//
// 7: findings entries carry their rules' notes.
//
// 6: the types section carries its compiler-options key.
//
// 5: the table holds the types phase's section.
//
// 4: run-cache inputs carry their change time and inode.
//
// 3: findings entries carry design-system rules and their fingerprint, and the findings section carries the
// design system's key.
//
// 2: findings entries carry shape-keyed rules and their fingerprint, and the table holds Signatures.
const cacheTableVersion = 15

// cacheTableMagic opens every file of the table, so a file that is not one is refused on its first field.
const cacheTableMagic = "cohere cache table"

// ErrCacheTableUnreadable means one or more of the table's files are not ones this build may use. Never a
// reason to fail a run: the answer is to run without those sections and write them anew.
var ErrCacheTableUnreadable = errors.New("cache table discarded")

// ErrCacheTablePartlyKept means the table's files were written by another cohere commit: its format record
// is kept, for its own key to decide, and every other file is dropped. Reported like a discard, since most
// of what the table held is gone.
var ErrCacheTablePartlyKept = errors.New("cache table written by another cohere commit")

// cacheTableHeader opens every file of the table. Section names what the body holds, so a file renamed
// into another section's place is refused rather than decoded as it.
type cacheTableHeader struct {
	Magic    string
	Version  int
	Identity CacheTableIdentity
	Section  string
}

// The table's sections, by the name each file's header gives and the file it lives in. A recorded run's
// file is named for its invocation (cacheTableRunFile).
const (
	cacheTableRunSection        = "run"
	cacheTableFindingsSection   = "findings"
	cacheTableSignaturesSection = "signatures"
	cacheTableFormatSection     = "format"
	cacheTableTypesSection      = "types"
)

// legacyCacheTableFile is the single file the table was until format 8, removed on sight.
const legacyCacheTableFile = "table.gob"

// cacheTableRunBody is a recorded run's file. Invocation is checked against the one the file is read for,
// so a file never answers for an invocation it was not recorded under.
type cacheTableRunBody struct {
	Invocation string
	Run        *RunCache
}

// cacheTableBodies is every body a file of the table can hold, one field per section, for the shape pin
// alone: each file holds exactly one of them.
type cacheTableBodies struct {
	Run        cacheTableRunBody
	Findings   *lintCacheWire
	Signatures map[string]SignatureEntry
	Formatted  *FormatSection
	Types      *TypesSection
}

// CacheTableSections names the parts of a table a read loads or a write replaces. Runs names invocations;
// AllRuns is every recorded run, which only --cache-dump reads.
type CacheTableSections struct {
	Runs       []string
	AllRuns    bool
	Findings   bool
	Signatures bool
	Formatted  bool
	Types      bool
}

// EveryCacheTableSection is the whole table.
var EveryCacheTableSection = CacheTableSections{AllRuns: true, Findings: true, Signatures: true, Formatted: true, Types: true}

// cacheTableRunFile is the file a recorded run lives in: a hash of its invocation, since an invocation is
// arguments joined by NUL and no file name can hold it.
func cacheTableRunFile(invocation string) string {
	sum := sha256.Sum256([]byte(invocation))
	return fmt.Sprintf("run-%x.gob", sum[:8])
}

// cacheTableFile is the file a section other than a run lives in.
func cacheTableFile(section string) string {
	return section + ".gob"
}

// lintCacheWire is the findings section as encoded. Rule lists are stored once and referenced by index:
// on ahra 3,605 entries share 4 distinct rule lists and 24 listening lists, and writing each out in full
// once made the file 53 MB.
type lintCacheWire struct {
	Version  int
	Key      [sha256.Size]byte
	KeyParts RunCacheKeyParts
	Lists    [][]string
	Entries  []lintCacheWireEntry

	DesignSystem *DesignSystemKey
}

type lintCacheWireEntry struct {
	Path            string
	ContentHash     [sha256.Size]byte
	Rules           int
	TypedRules      int
	TypeFingerprint [sha256.Size]byte
	Listening       int
	VisitedNodes    int
	Findings        []LintCacheFinding

	ShapedRules      int
	ShapeFingerprint [sha256.Size]byte

	DesignRules       int
	DesignFingerprint [sha256.Size]byte

	DerivedRules       int
	DerivedFingerprint [sha256.Size]byte

	Notes []lintCacheWireNote

	Directives bool
	Withheld   []LintCacheWithheld

	// Adamic is the entry's readiness record as it is, nil and empty kept apart: gob sends a pointer to an empty
	// struct and leaves a nil one out.
	Adamic *AdamicRecord
}

// lintCacheWireNote is one rule's count of one note key in one file. Notes are a sorted slice on the wire
// rather than the entry's map, so a table's bytes do not depend on map order.
type lintCacheWireNote struct {
	Rule  string
	Key   string
	Count int
}

// wireNotes flattens an entry's notes, sorted by rule and then key.
func wireNotes(notes RuleNotes) []lintCacheWireNote {
	var flat []lintCacheWireNote
	for ruleName, counts := range notes {
		for key, count := range counts {
			flat = append(flat, lintCacheWireNote{Rule: ruleName, Key: key, Count: count})
		}
	}
	sort.Slice(flat, func(first, second int) bool {
		if flat[first].Rule != flat[second].Rule {
			return flat[first].Rule < flat[second].Rule
		}
		return flat[first].Key < flat[second].Key
	})
	return flat
}

// ruleNotes is wireNotes undone, nil when there are none.
func ruleNotes(flat []lintCacheWireNote) RuleNotes {
	if len(flat) == 0 {
		return nil
	}
	notes := RuleNotes{}
	for _, note := range flat {
		if notes[note.Rule] == nil {
			notes[note.Rule] = map[string]int{}
		}
		notes[note.Rule][note.Key] = note.Count
	}
	return notes
}

// NewCacheTable is an empty table, what a first run and every discard start from.
func NewCacheTable() *CacheTable {
	return &CacheTable{Runs: map[string]*RunCache{}}
}

// CacheTableInvocation names an invocation within a table.
func CacheTableInvocation(arguments []string) string {
	return strings.Join(arguments, "\x00")
}

// encodeCacheTableFile encodes one file of the table: its header, then the section's body.
func encodeCacheTableFile(section string, body any, identity CacheTableIdentity) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(cacheTableHeader{Magic: cacheTableMagic, Version: cacheTableVersion, Identity: identity, Section: section}); err != nil {
		return nil, fmt.Errorf("encoding the %s file's header: %w", section, err)
	}
	if err := encoder.Encode(body); err != nil {
		return nil, fmt.Errorf("encoding the %s file: %w", section, err)
	}
	return buffer.Bytes(), nil
}

// errWrittenByAnotherCommit is a file whose header differs from this build's only in the cohere commit, the
// one difference the format record survives.
var errWrittenByAnotherCommit = errors.New("written by another cohere commit")

// decodeCacheTableFile decodes one file of the table into body, a pointer to the section's type. Anything but
// a file of this section, written by this build, with nothing after its body, is an error: a file from
// another cohere commit that differs in nothing else wraps errWrittenByAnotherCommit, and body is still
// filled, for the one section that may keep it.
func decodeCacheTableFile(contents []byte, section string, identity CacheTableIdentity, body any) error {
	reader := bytes.NewReader(contents)
	decoder := gob.NewDecoder(reader)

	var header cacheTableHeader
	if err := decoder.Decode(&header); err != nil {
		return fmt.Errorf("unreadable header: %v", err)
	}
	if header.Magic != cacheTableMagic {
		return errors.New("not a cache table file")
	}
	if header.Version != cacheTableVersion {
		return fmt.Errorf("format version %d, this build reads %d", header.Version, cacheTableVersion)
	}
	if header.Section != section {
		return fmt.Errorf("holds the %s section, read for %s", header.Section, section)
	}
	var otherCommit error
	if header.Identity != identity {
		other := header.Identity
		other.SelfCommit = identity.SelfCommit
		if other != identity {
			return fmt.Errorf("written by %s, this is %s", header.Identity, identity)
		}
		otherCommit = fmt.Errorf("%w (%s, this is %s)", errWrittenByAnotherCommit, orUnknown(header.Identity.SelfCommit), orUnknown(identity.SelfCommit))
	}

	if err := decoder.Decode(body); err != nil {
		return fmt.Errorf("unreadable body: %v", err)
	}
	// gob stops at the end of the value it was asked for, so bytes after it would otherwise go unread. A
	// file longer than its contents is a file something else wrote into.
	if reader.Len() != 0 {
		return fmt.Errorf("%d bytes after the body", reader.Len())
	}
	return otherCommit
}

func (identity CacheTableIdentity) String() string {
	return fmt.Sprintf("cohere %s, compiler %s, %s, %s",
		orUnknown(identity.SelfCommit), orUnknown(identity.CompilerCommit), identity.GoToolchain, identity.Platform)
}

func orUnknown(value string) string {
	if value == "" {
		return "unknown"
	}
	return value
}

// wire interns the findings cache's rule lists. It is stamped with this build's lintCacheVersion: a cache in
// memory always means what this build means.
func (c *LintCache) wire() *lintCacheWire {
	wire := &lintCacheWire{Version: lintCacheVersion, Key: c.Key, KeyParts: c.KeyParts, DesignSystem: c.DesignSystem}
	// Interned by a cheap key, the length and the two ends, and confirmed by comparing names, never by joining
	// a list into one string: ahra's entries hold about 18,000 lists of up to ~250 names, and joining each cost
	// about 25ms of every recording run to find 31 distinct ones (#a66sfmh).
	type listKey struct {
		length      int
		first, last string
	}
	positions := map[listKey][]int{}
	intern := func(list []string) int {
		key := listKey{length: len(list)}
		if len(list) > 0 {
			key.first, key.last = list[0], list[len(list)-1]
		}
		for _, position := range positions[key] {
			if equalStrings(wire.Lists[position], list) {
				return position
			}
		}
		positions[key] = append(positions[key], len(wire.Lists))
		wire.Lists = append(wire.Lists, append([]string{}, list...))
		return len(wire.Lists) - 1
	}
	for _, entry := range c.Entries {
		wire.Entries = append(wire.Entries, lintCacheWireEntry{
			Path:            entry.Path,
			ContentHash:     entry.ContentHash,
			Rules:           intern(entry.Rules),
			TypedRules:      intern(entry.TypedRules),
			TypeFingerprint: entry.TypeFingerprint,
			Listening:       intern(entry.Listening),
			VisitedNodes:    entry.VisitedNodes,
			Findings:        entry.Findings,

			ShapedRules:      intern(entry.ShapedRules),
			ShapeFingerprint: entry.ShapeFingerprint,

			DesignRules:       intern(entry.DesignRules),
			DesignFingerprint: entry.DesignFingerprint,

			DerivedRules:       intern(entry.DerivedRules),
			DerivedFingerprint: entry.DerivedFingerprint,

			Notes: wireNotes(entry.Notes),

			Directives: entry.Directives,
			Withheld:   entry.Withheld,

			Adamic: entry.Adamic,
		})
	}
	return wire
}

// cache resolves the interned lists. An index past its table is an error rather than an empty list: an
// index trusted past its table would hand an entry another file's rule list.
//
// A findings section of another version decodes to nil rather than failing the table. The runs beside it
// are proven by their own keys, and a findings cache is only ever a miss when it is absent.
func (wire *lintCacheWire) cache() (*LintCache, error) {
	if wire.Version != lintCacheVersion {
		return nil, nil
	}
	cache := &LintCache{Version: wire.Version, Key: wire.Key, KeyParts: wire.KeyParts, Entries: make([]LintCacheEntry, 0, len(wire.Entries)),
		DesignSystem: wire.DesignSystem}
	list := func(index int) ([]string, error) {
		if index < 0 || index >= len(wire.Lists) {
			return nil, fmt.Errorf("list %d of %d", index, len(wire.Lists))
		}
		return wire.Lists[index], nil
	}
	for position, stored := range wire.Entries {
		entry := LintCacheEntry{
			Path:            stored.Path,
			ContentHash:     stored.ContentHash,
			TypeFingerprint: stored.TypeFingerprint,
			VisitedNodes:    stored.VisitedNodes,
			Findings:        stored.Findings,

			ShapeFingerprint: stored.ShapeFingerprint,

			DesignFingerprint: stored.DesignFingerprint,

			DerivedFingerprint: stored.DerivedFingerprint,

			Notes: ruleNotes(stored.Notes),

			Directives: stored.Directives,
			Withheld:   stored.Withheld,

			Adamic: stored.Adamic,
		}
		var err error
		if entry.Rules, err = list(stored.Rules); err != nil {
			return nil, fmt.Errorf("entry %d rules: %v", position, err)
		}
		if entry.TypedRules, err = list(stored.TypedRules); err != nil {
			return nil, fmt.Errorf("entry %d typed rules: %v", position, err)
		}
		if entry.Listening, err = list(stored.Listening); err != nil {
			return nil, fmt.Errorf("entry %d listening: %v", position, err)
		}
		if entry.ShapedRules, err = list(stored.ShapedRules); err != nil {
			return nil, fmt.Errorf("entry %d shaped rules: %v", position, err)
		}
		if entry.DesignRules, err = list(stored.DesignRules); err != nil {
			return nil, fmt.Errorf("entry %d design-system rules: %v", position, err)
		}
		if entry.DerivedRules, err = list(stored.DerivedRules); err != nil {
			return nil, fmt.Errorf("entry %d derived rules: %v", position, err)
		}
		cache.Entries = append(cache.Entries, entry)
	}
	return cache, nil
}

// ReadCacheTable loads the named sections of the table in directory. It always returns a table: every
// section whose file is readable, the format record from another cohere commit, and nothing for the rest.
// Every file path in it is as anchor spells it: the table stores paths below the project root relative to it,
// so a run from another spelling of the root reads the same entries (#547dhjz; see PathAnchor).
// The error says what was not kept. Nothing on disk at all wraps os.ErrNotExist, which is a first run and
// nothing to report. Anything else wraps ErrCacheTableUnreadable or ErrCacheTablePartlyKept, which a
// caller reports, because a cache that is silently thrown away every run is a saving that quietly never
// appears.
//
// The single table.gob of earlier formats is removed here, unread.
func ReadCacheTable(directory string, identity CacheTableIdentity, sections CacheTableSections, anchor PathAnchor) (*CacheTable, error) {
	table, err := readCacheTable(directory, identity, sections)
	table.spell(anchor)
	return table, err
}

func readCacheTable(directory string, identity CacheTableIdentity, sections CacheTableSections) (*CacheTable, error) {
	os.Remove(filepath.Join(directory, legacyCacheTableFile))
	table := NewCacheTable()
	found := false
	var dropped []string
	anotherCommit := ""
	onlyAnotherCommit := true
	read := func(name string, section string, body any) bool {
		contents, err := os.ReadFile(filepath.Join(directory, name))
		if errors.Is(err, os.ErrNotExist) {
			return false
		}
		found = true
		if err == nil {
			err = decodeCacheTableFile(contents, section, identity, body)
		}
		switch {
		case err == nil:
			return true
		case errors.Is(err, errWrittenByAnotherCommit):
			if section == cacheTableFormatSection {
				return true
			}
			anotherCommit = strings.TrimPrefix(err.Error(), errWrittenByAnotherCommit.Error()+" ")
			dropped = append(dropped, name)
		default:
			onlyAnotherCommit = false
			dropped = append(dropped, fmt.Sprintf("%s (%v)", name, err))
		}
		return false
	}

	invocations := sections.Runs
	if sections.AllRuns {
		invocations = nil
		names, _ := filepath.Glob(filepath.Join(directory, "run-*.gob"))
		for _, name := range names {
			var body cacheTableRunBody
			if read(filepath.Base(name), cacheTableRunSection, &body) && body.Run != nil {
				table.Runs[body.Invocation] = body.Run
			}
		}
	}
	for _, invocation := range invocations {
		var body cacheTableRunBody
		name := cacheTableRunFile(invocation)
		if !read(name, cacheTableRunSection, &body) || body.Run == nil {
			continue
		}
		if body.Invocation != invocation {
			onlyAnotherCommit = false
			dropped = append(dropped, fmt.Sprintf("%s (recorded for another invocation)", name))
			continue
		}
		table.Runs[invocation] = body.Run
	}
	if sections.Findings {
		var wire *lintCacheWire
		if read(cacheTableFile(cacheTableFindingsSection), cacheTableFindingsSection, &wire) && wire != nil {
			findings, err := wire.cache()
			if err != nil {
				onlyAnotherCommit = false
				dropped = append(dropped, fmt.Sprintf("%s (%v)", cacheTableFile(cacheTableFindingsSection), err))
			} else {
				table.Findings = findings
			}
		}
	}
	if sections.Signatures {
		var signatures map[string]SignatureEntry
		if read(cacheTableFile(cacheTableSignaturesSection), cacheTableSignaturesSection, &signatures) {
			table.Signatures = signatures
		}
	}
	if sections.Formatted {
		var formatted *FormatSection
		if read(cacheTableFile(cacheTableFormatSection), cacheTableFormatSection, &formatted) {
			table.Formatted = formatted
		}
	}
	if sections.Types {
		var types *TypesSection
		if read(cacheTableFile(cacheTableTypesSection), cacheTableTypesSection, &types) {
			table.Types = types
		}
	}

	switch {
	case len(dropped) == 0 && !found:
		return table, fmt.Errorf("no cache table in %s: %w", directory, os.ErrNotExist)
	case len(dropped) == 0:
		return table, nil
	case onlyAnotherCommit:
		return table, fmt.Errorf("%w %s: %s dropped, and its format record is kept for its own key to decide",
			ErrCacheTablePartlyKept, anotherCommit, strings.Join(dropped, ", "))
	default:
		return table, fmt.Errorf("%w: %s", ErrCacheTableUnreadable, strings.Join(dropped, ", "))
	}
}

// beforeCacheTableRename, when set, runs between writing a temporary and renaming it into place. Only a test
// sets it, to stop a writer at the one moment a kill could matter.
var beforeCacheTableRename func(temporaryName string)

// WriteCacheTable writes the named sections of table into directory, each to its own file, and leaves every
// other file as it is. A section named but nil in table is not written, so a run that has nothing to say
// about a section never erases what another run left there. Each file path is written anchored, below the
// project root as a path from it; table itself is left as it is. See ReadCacheTable.
//
// Each file is written atomically: a temporary in the directory, then a rename. Two runs can share a tree,
// and a reader must never see half a file. A run that returned to its caller early writes after the caller
// has moved on, where nothing stops a kill landing mid-write; the rename is what keeps the old file whole if
// one does (TestAWriterKilledBeforeItsRenameLeavesTheOldTable). Files are not atomic with each other, and
// need not be: each carries its own proof, and the table's lock orders whole writes against reads.
func WriteCacheTable(directory string, table *CacheTable, identity CacheTableIdentity, sections CacheTableSections, anchor PathAnchor) error {
	table = table.anchored(anchor)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", directory, err)
	}
	write := func(name string, section string, body any) error {
		encoded, err := encodeCacheTableFile(section, body, identity)
		if err != nil {
			return err
		}
		return writeCacheTableFile(filepath.Join(directory, name), encoded)
	}
	invocations := sections.Runs
	if sections.AllRuns {
		invocations = invocations[:0:0]
		for invocation := range table.Runs {
			invocations = append(invocations, invocation)
		}
		sort.Strings(invocations)
	}
	for _, invocation := range invocations {
		if run := table.Runs[invocation]; run != nil {
			if err := write(cacheTableRunFile(invocation), cacheTableRunSection, cacheTableRunBody{Invocation: invocation, Run: run}); err != nil {
				return err
			}
		}
	}
	if sections.Findings && table.Findings != nil {
		if err := write(cacheTableFile(cacheTableFindingsSection), cacheTableFindingsSection, table.Findings.wire()); err != nil {
			return err
		}
	}
	if sections.Signatures && table.Signatures != nil {
		if err := write(cacheTableFile(cacheTableSignaturesSection), cacheTableSignaturesSection, table.Signatures); err != nil {
			return err
		}
	}
	if sections.Formatted && table.Formatted != nil {
		if err := write(cacheTableFile(cacheTableFormatSection), cacheTableFormatSection, table.Formatted); err != nil {
			return err
		}
	}
	if sections.Types && table.Types != nil {
		if err := write(cacheTableFile(cacheTableTypesSection), cacheTableTypesSection, table.Types); err != nil {
			return err
		}
	}
	return nil
}

// anchored is a copy of the table with every file path it keys on anchored, for writing. A recorded run's
// output is rendered for the run that replays it by the command (see recordedOutput there).
func (table *CacheTable) anchored(anchor PathAnchor) *CacheTable {
	return table.mapPaths(anchor.Stable)
}

// spell turns every anchored path in the table, as read, into this run's spelling.
func (table *CacheTable) spell(anchor PathAnchor) {
	*table = *table.mapPaths(anchor.Spelled)
}

// mapPaths is a copy of the table with path applied to every file path a section keys on: each recorded run's
// inputs, each findings entry's, the shapes', the format record's and the types section's. Sections it does not
// hold stay nil.
func (table *CacheTable) mapPaths(path func(string) string) *CacheTable {
	mapped := *table
	if table.Runs != nil {
		mapped.Runs = make(map[string]*RunCache, len(table.Runs))
		for invocation, run := range table.Runs {
			if run == nil {
				mapped.Runs[invocation] = nil
				continue
			}
			copied := *run
			copied.Inputs = make([]RunCacheInput, len(run.Inputs))
			for index, input := range run.Inputs {
				input.Path = path(input.Path)
				copied.Inputs[index] = input
			}
			mapped.Runs[invocation] = &copied
		}
	}
	if table.Findings != nil {
		findings := *table.Findings
		findings.Entries = make([]LintCacheEntry, len(table.Findings.Entries))
		for index, entry := range table.Findings.Entries {
			entry.Path = path(entry.Path)
			findings.Entries[index] = entry
		}
		findings.index = nil
		mapped.Findings = &findings
	}
	if table.Signatures != nil {
		mapped.Signatures = make(map[string]SignatureEntry, len(table.Signatures))
		for name, entry := range table.Signatures {
			mapped.Signatures[path(name)] = entry
		}
	}
	if table.Formatted != nil {
		formatted := *table.Formatted
		if table.Formatted.Entries != nil {
			formatted.Entries = make(map[string]FormatEntry, len(table.Formatted.Entries))
			for name, entry := range table.Formatted.Entries {
				formatted.Entries[path(name)] = entry
			}
		}
		mapped.Formatted = &formatted
	}
	if table.Types != nil {
		types := *table.Types
		if table.Types.Entries != nil {
			types.Entries = make(map[string]TypesEntry, len(table.Types.Entries))
			for name, entry := range table.Types.Entries {
				types.Entries[path(name)] = entry
			}
		}
		mapped.Types = &types
	}
	return &mapped
}

// writeCacheTableFile writes one file through a temporary beside it and a rename.
func writeCacheTableFile(path string, encoded []byte) error {
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".cachetable-*")
	if err != nil {
		return fmt.Errorf("creating a temporary in %s: %w", directory, err)
	}
	temporaryName := temporary.Name()
	defer func() {
		if temporaryName != "" {
			os.Remove(temporaryName)
		}
	}()
	if _, err := temporary.Write(encoded); err != nil {
		temporary.Close()
		return fmt.Errorf("writing %s: %w", temporaryName, err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("closing %s: %w", temporaryName, err)
	}
	if beforeCacheTableRename != nil {
		beforeCacheTableRename(temporaryName)
	}
	// replace.File, which retries through Windows refusing the rename while another run reads the file.
	if err := replace.File(temporaryName, path); err != nil {
		return fmt.Errorf("renaming %s into place: %w", filepath.Base(path), err)
	}
	temporaryName = ""
	return nil
}

// DumpCacheTable prints a table for a person: every run and every file entry, one line each, so a cache
// that looks wrong can be read rather than guessed at.
func DumpCacheTable(out io.Writer, directory string, table *CacheTable, identity CacheTableIdentity) {
	fmt.Fprintf(out, "cache table %s\n", directory)
	fmt.Fprintf(out, "  format %d, written by %s\n", cacheTableVersion, identity)

	invocations := make([]string, 0, len(table.Runs))
	for invocation := range table.Runs {
		invocations = append(invocations, invocation)
	}
	sort.Strings(invocations)
	fmt.Fprintf(out, "runs: %d\n", len(invocations))
	for _, invocation := range invocations {
		run := table.Runs[invocation]
		name := strings.ReplaceAll(invocation, "\x00", " ")
		if name == "" {
			name = "(bare)"
		}
		files, directories, absent := 0, 0, 0
		for _, input := range run.Inputs {
			switch {
			case !input.Exists:
				absent++
			case input.Directory:
				directories++
			default:
				files++
			}
		}
		fmt.Fprintf(out, "  %s: recorded %s, exit %d, key %.12s, %d inputs (%d files, %d directories, %d absent), %d bytes of output, %d of errors\n",
			name, time.Unix(0, run.RecordedUnixNanoseconds).Format(time.DateTime), run.ExitCode, run.Key,
			len(run.Inputs), files, directories, absent, len(run.Output), len(run.Errors))
		if changed := run.ChangedInput(); changed != nil {
			fmt.Fprintf(out, "    now: would not replay, %v\n", changed)
		} else {
			fmt.Fprintln(out, "    now: every input as recorded, so it replays if the binary, flags and directory match")
		}
	}

	fmt.Fprintf(out, "signatures: %d files\n", len(table.Signatures))

	if table.Formatted == nil {
		fmt.Fprintln(out, "formatted: none")
	} else {
		fmt.Fprintf(out, "formatted: %d files, key %.12s\n", len(table.Formatted.Entries), table.Formatted.Key)
	}

	if table.Findings == nil {
		fmt.Fprintln(out, "findings: none")
		return
	}
	findingCount := 0
	for _, entry := range table.Findings.Entries {
		findingCount += len(entry.Findings)
	}
	fmt.Fprintf(out, "findings: %d files, %d findings, key %x\n", len(table.Findings.Entries), findingCount, table.Findings.Key[:6])
	if design := table.Findings.DesignSystem; design == nil {
		fmt.Fprintln(out, "  design system: none recorded")
	} else {
		fmt.Fprintf(out, "  design system: %x, %d paths read\n", design.Fingerprint[:6], len(design.Reads))
		for _, read := range design.Reads {
			state := "absent"
			switch {
			case read.Directory:
				state = fmt.Sprintf("directory %x", read.Hash[:6])
			case read.Present:
				state = fmt.Sprintf("file %x", read.Hash[:6])
			}
			fmt.Fprintf(out, "    %s  %s\n", read.Path, state)
		}
	}
	for _, entry := range table.Findings.Entries {
		fmt.Fprintf(out, "  %s  content %x  types %x  shapes %x  design %x  %d pure + %d type-aware + %d shape-keyed + %d design-system rules, %d listening, %d nodes, %d findings\n",
			entry.Path, entry.ContentHash[:6], entry.TypeFingerprint[:6], entry.ShapeFingerprint[:6], entry.DesignFingerprint[:6], len(entry.Rules),
			len(entry.TypedRules), len(entry.ShapedRules), len(entry.DesignRules), len(entry.Listening), entry.VisitedNodes, len(entry.Findings))
		for _, finding := range entry.Findings {
			fmt.Fprintf(out, "    %d-%d %s/%s: %s\n", finding.Start, finding.End, finding.RuleName, finding.MessageId, finding.MessageDescription)
		}
		for _, withheld := range entry.Withheld {
			fmt.Fprintf(out, "    withheld by directive %d: %s\n", withheld.Directive, withheld.RuleName)
		}
		if entry.Adamic != nil {
			fmt.Fprintf(out, "    adamic: %d rules measured, %d skipped\n", len(entry.Adamic.Counts), len(entry.Adamic.Skipped))
			for _, count := range entry.Adamic.Counts {
				if count.Findings > 0 {
					fmt.Fprintf(out, "      %s: %d\n", count.Rule, count.Findings)
				}
			}
		}
	}
}
