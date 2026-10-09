// Package gitignore decides which paths a repository's ignore files exclude, the way git decides it,
// without running git.
//
// It reads what gitignore(5) names for a working tree: every `.gitignore` from the repository root down,
// each scoped to its own directory, and `.git/info/exclude` when `.git` is a directory. The user's global
// core.excludesFile is not read: it is configuration on one machine, not the repository's, so a check that
// read it would pass on one machine and fail on another. Matching is case sensitive, as git's is unless
// core.ignorecase is set, which is likewise a property of a clone rather than of the repository.
//
// The semantics are gitignore(5)'s:
//
//   - Precedence, highest first: a `.gitignore` in the path's own directory, then each parent's up to the
//     root, then info/exclude. Within one file the last matching line decides, and a `!` line re-includes.
//   - A pattern with a `/` at its start or middle is anchored to its file's directory. One without matches
//     a base name at any depth below it. A trailing `/` matches directories only.
//   - Nothing below an excluded directory can be re-included: git never looks inside one, so the rule
//     that excluded the directory decides every path under it.
//   - Trailing spaces are dropped unless escaped, `#` starts a comment, and a backslash escapes `#`, `!`
//     and any other byte. A file's byte order mark and each line's carriage return are not part of a
//     pattern.
//
// What git would read differently from its documentation is refused by name rather than followed: a
// `.gitignore` that is a symbolic link (git does not follow one inside a working tree), one larger than
// 100 MiB (git skips it with a warning), and a line holding a NUL byte. Each is an error naming the file,
// because skipping one would leave its patterns unapplied without a word.
//
// The matcher only answers. Walking, and pruning what it excludes, belong to the caller: it enters each
// directory it descends into with Enter, and asks Ignored about each entry it finds there.
package gitignore

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// IgnoreFileName is the per-directory ignore file.
const IgnoreFileName = ".gitignore"

// ExcludeFile is the repository's own exclude file, relative to its root.
const ExcludeFile = ".git/info/exclude"

// maximumFileSize is the largest ignore file read. Git skips one past 100 MiB with a warning; here it is
// refused, since skipping it would leave every pattern in it unapplied without a word.
const maximumFileSize = 100 << 20

// ErrNestedRepository is a refusal to enter a directory that is a repository of its own: its ignore
// files govern it, through its own New, and the parent's patterns do not.
var ErrNestedRepository = errors.New("is a repository of its own")

// Source is the rule that decided a path: the ignore file, relative to the repository root, the line in
// it, and the pattern as git prints it, `!` and a trailing `/` included. The zero Source means no rule
// matched.
type Source struct {
	File    string
	Line    int
	Pattern string
}

// IsZero reports whether no rule decided.
func (source Source) IsZero() bool { return source.File == "" }

// String is git check-ignore's verbose form, `file:line:pattern`.
func (source Source) String() string {
	return fmt.Sprintf("%s:%d:%s", source.File, source.Line, source.Pattern)
}

// rule is one compiled pattern line.
type rule struct {
	source        Source
	negated       bool
	directoryOnly bool

	// anchored is a pattern matched against the path below base rather than against a base name. base is
	// the directory of the file the rule came from, relative to the root, "" at the root.
	anchored bool
	base     string
	glob     glob
}

// Matcher is one directory's view of a repository's ignore rules: those of every ignore file from the
// root down to that directory. It is never modified once made, so a walk may keep one per directory and
// share them freely.
type Matcher struct {
	root      string
	directory string

	// files are the rule lists of each `.gitignore` from the root down, root first. exclude is
	// info/exclude's.
	files   [][]rule
	exclude []rule

	// excludedBy is the rule that excluded this directory or one above it. It decides every path below.
	excludedBy *rule

	// listing, when set, answers Enter's questions about a `.git` or a `.gitignore` a directory does not
	// hold, from a listing already read. See WithListing.
	listing Listing
}

// Listing is a directory's entries as os.ReadDir returned them, and whether the directory was listed, for an
// absolute path. A directory not listed is not empty: it is read from the disk.
type Listing func(directory string) ([]os.DirEntry, bool)

// WithListing returns the matcher, and every matcher Enter makes from it, answering from listing where a
// listed directory shows it holds no `.git` or no `.gitignore`, as an lstat of either would then fail. One
// the listing names is still asked of the disk, so what is read is read as before. The format walk passes
// discovery's listings (#g3046x5), which spares two lstats per directory entered.
func (matcher *Matcher) WithListing(listing Listing) *Matcher {
	listed := *matcher
	listed.listing = listing
	return &listed
}

// listedAbsent reports whether a listing of directory, an absolute path, shows it holds no entry called name.
func (matcher *Matcher) listedAbsent(directory string, name string) bool {
	if matcher.listing == nil {
		return false
	}
	entries, listed := matcher.listing(directory)
	if !listed {
		return false
	}
	for _, entry := range entries {
		if entry.Name() == name {
			return false
		}
	}
	return true
}

// New reads a repository's root `.gitignore` and, when `.git` is a directory, its info/exclude.
func New(repositoryRoot string) (*Matcher, error) {
	root, err := filepath.Abs(repositoryRoot)
	if err != nil {
		return nil, err
	}
	matcher := &Matcher{root: root}

	rules, err := readTreeFile(root, "")
	if err != nil {
		return nil, err
	}
	matcher.files = append(matcher.files, rules)

	if info, statError := os.Stat(filepath.Join(root, ".git")); statError == nil && info.IsDir() {
		// info/exclude lives outside the tree, so a symbolic link to it is followed, as git follows it.
		excludePath := filepath.Join(root, filepath.FromSlash(ExcludeFile))
		contents, readError := readIgnoreFile(excludePath, true)
		if readError != nil {
			return nil, readError
		}
		matcher.exclude, err = parseRules(contents, ExcludeFile, "")
		if err != nil {
			return nil, err
		}
	}
	return matcher, nil
}

// Root is the repository root the matcher reads, as an absolute path.
func (matcher *Matcher) Root() string { return matcher.root }

// Directory is the directory this matcher answers for, relative to the root, "" for the root itself.
func (matcher *Matcher) Directory() string { return matcher.directory }

// Excluded reports whether this directory, or one above it, is excluded, and by which rule. A walk that
// entered one anyway gets the same answer for everything inside.
func (matcher *Matcher) Excluded() (bool, Source) {
	if matcher.excludedBy == nil {
		return false, Source{}
	}
	return true, matcher.excludedBy.source
}

// Enter returns the matcher for relativeDirectory, which is this matcher's directory or one below it,
// slash separated and relative to the root.
//
// It reads each `.gitignore` between the two once, and stops reading at a directory that is itself
// excluded, as git does: the rule that excluded it then decides everything below. A directory holding a
// `.git` is refused with ErrNestedRepository, and an ignore file that cannot be read faithfully (a
// symbolic link, which git does not follow in a tree; one too large; one unreadable) is an error naming
// it, never skipped.
func (matcher *Matcher) Enter(relativeDirectory string) (*Matcher, error) {
	relativeDirectory = cleanRelative(relativeDirectory)
	if relativeDirectory == matcher.directory {
		return matcher, nil
	}
	remainder, below := strings.CutPrefix(relativeDirectory, matcher.directory)
	if matcher.directory != "" {
		remainder, below = strings.CutPrefix(remainder, "/")
	}
	if !below || remainder == "" {
		return nil, fmt.Errorf("gitignore: %s is not below %s", relativeDirectory, displayDirectory(matcher.directory))
	}

	entered := *matcher
	entered.files = entered.files[:len(entered.files):len(entered.files)]
	for _, name := range strings.Split(remainder, "/") {
		child := joinRelative(entered.directory, name)
		childDirectory := filepath.Join(matcher.root, filepath.FromSlash(child))
		if matcher.listedAbsent(childDirectory, ".git") {
			// Not there, as the lstat would have found.
		} else if _, err := os.Lstat(filepath.Join(childDirectory, ".git")); err == nil {
			return nil, fmt.Errorf("gitignore: %s %w", childDirectory, ErrNestedRepository)
		}
		entered.directory = child
		if entered.excludedBy != nil {
			continue
		}
		if decided := entered.decide(child, path.Base(child), true, len(entered.files)); decided != nil && !decided.negated {
			entered.excludedBy = decided
			continue
		}
		if matcher.listedAbsent(childDirectory, IgnoreFileName) {
			// No `.gitignore`, which readTreeFile reads as no rules.
			entered.files = append(entered.files, nil)
			continue
		}
		rules, err := readTreeFile(matcher.root, child)
		if err != nil {
			return nil, err
		}
		entered.files = append(entered.files, rules)
	}
	return &entered, nil
}

// Ignored reports whether relativePath, an entry of this matcher's directory given relative to the root,
// is excluded, and the rule that decided. A rule that re-included the path is reported with false.
//
// isDirectory says whether the path is a directory, which is what a pattern with a trailing `/` asks; a
// symbolic link to a directory is not one, as git does not follow it. Asking about a path outside this
// directory is a programming error and panics, because the answer would silently leave out the ignore
// files between the two.
func (matcher *Matcher) Ignored(relativePath string, isDirectory bool) (bool, Source) {
	relativePath = cleanRelative(relativePath)
	parent := path.Dir(relativePath)
	if parent == "." {
		parent = ""
	}
	if parent != matcher.directory {
		panic(fmt.Sprintf("gitignore: asked about %s through the matcher for %s; Enter its directory first",
			relativePath, displayDirectory(matcher.directory)))
	}
	if matcher.excludedBy != nil {
		return true, matcher.excludedBy.source
	}
	decided := matcher.decide(relativePath, path.Base(relativePath), isDirectory, len(matcher.files))
	if decided == nil {
		return false, Source{}
	}
	return !decided.negated, decided.source
}

// IgnoredPath is Ignored for a path anywhere below this matcher's directory: it enters the directories in
// between first, reading their ignore files, so the answer names a nested file's line when one decides.
// It is for a question about one path, such as explaining why a file was left out; a walk enters each
// directory itself and asks Ignored, which reads every ignore file once.
func (matcher *Matcher) IgnoredPath(relativePath string, isDirectory bool) (bool, Source, error) {
	relativePath = cleanRelative(relativePath)
	parent := path.Dir(relativePath)
	if parent == "." {
		parent = ""
	}
	scope, err := matcher.Enter(parent)
	if err != nil {
		return false, Source{}, err
	}
	ignored, source := scope.Ignored(relativePath, isDirectory)
	return ignored, source, nil
}

// Patterns is a list of ignore-file lines that is not a file in the tree, read with the same syntax and
// relative to the root it is applied at: a project setting that says what to skip in the words a
// `.gitignore` would use. Within the list the last matching line decides, as within one file.
type Patterns struct {
	rules []rule
}

// CompilePatterns reads lines as the lines of an ignore file named name, which is what each Source
// reports. A line holding a newline or a NUL byte cannot be an ignore-file line and is refused.
func CompilePatterns(lines []string, name string) (*Patterns, error) {
	patterns := &Patterns{}
	for index, line := range lines {
		if strings.ContainsAny(line, "\n\x00") {
			return nil, fmt.Errorf("gitignore: %s entry %d, %q, holds a newline or a NUL byte, which no ignore-file line can", name, index+1, line)
		}
		line = strings.TrimSuffix(line, "\r")
		if line == "" || line[0] == '#' {
			continue
		}
		patterns.rules = append(patterns.rules, compileRule(trimTrailingSpaces(line), name, index+1, ""))
	}
	return patterns, nil
}

// Ignored reports whether the list excludes relativePath, given relative to the root it applies at, and
// the line that decided. Unlike a Matcher it knows nothing of directories above the path: a walk applying
// it prunes each directory it excludes, so a path below one is never asked about.
func (patterns *Patterns) Ignored(relativePath string, isDirectory bool) (bool, Source) {
	relativePath = cleanRelative(relativePath)
	decided := lastMatching(patterns.rules, relativePath, path.Base(relativePath), isDirectory)
	if decided == nil {
		return false, Source{}
	}
	return !decided.negated, decided.source
}

// decide is the rule that decides relativePath among the first fileCount tree files and info/exclude,
// or nil: the deepest file first, the last line of each first, info/exclude last.
func (matcher *Matcher) decide(relativePath string, baseName string, isDirectory bool, fileCount int) *rule {
	for fileIndex := fileCount - 1; fileIndex >= 0; fileIndex-- {
		if decided := lastMatching(matcher.files[fileIndex], relativePath, baseName, isDirectory); decided != nil {
			return decided
		}
	}
	return lastMatching(matcher.exclude, relativePath, baseName, isDirectory)
}

func lastMatching(rules []rule, relativePath string, baseName string, isDirectory bool) *rule {
	for index := len(rules) - 1; index >= 0; index-- {
		candidate := &rules[index]
		if candidate.directoryOnly && !isDirectory {
			continue
		}
		if candidate.matches(relativePath, baseName) {
			return candidate
		}
	}
	return nil
}

func (candidate *rule) matches(relativePath string, baseName string) bool {
	if !candidate.anchored {
		return candidate.glob.matches(baseName, false)
	}
	below := relativePath
	if candidate.base != "" {
		var ok bool
		below, ok = strings.CutPrefix(relativePath, candidate.base+"/")
		if !ok {
			return false
		}
	}
	return candidate.glob.matches(below, true)
}

// readTreeFile reads the `.gitignore` of directory, relative to root, returning no rules when it has
// none.
func readTreeFile(root string, directory string) ([]rule, error) {
	relativeFile := joinRelative(directory, IgnoreFileName)
	contents, err := readIgnoreFile(filepath.Join(root, filepath.FromSlash(relativeFile)), false)
	if err != nil || contents == nil {
		return nil, err
	}
	return parseRules(contents, relativeFile, directory)
}

// readIgnoreFile returns a file's contents, or nil when it does not exist. follow says whether a symbolic
// link is read through: git follows one for info/exclude and not for a `.gitignore` in the tree.
func readIgnoreFile(filePath string, follow bool) ([]byte, error) {
	info, err := os.Lstat(filePath)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("gitignore: reading %s: %w", filePath, err)
	}
	if info.Mode()&fs.ModeSymlink != 0 {
		if !follow {
			return nil, fmt.Errorf("gitignore: %s is a symbolic link, which git does not follow inside a working tree, so its patterns would not apply; replace it with the file it points to", filePath)
		}
		if info, err = os.Stat(filePath); errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		} else if err != nil {
			return nil, fmt.Errorf("gitignore: reading %s: %w", filePath, err)
		}
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("gitignore: %s is not a regular file, so it cannot be read as an ignore file", filePath)
	}
	if info.Size() > maximumFileSize {
		return nil, fmt.Errorf("gitignore: %s is larger than 100 MiB, which git skips with a warning, leaving its patterns unapplied; cohere refuses it rather than skip it", filePath)
	}
	contents, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("gitignore: reading %s: %w", filePath, err)
	}
	return contents, nil
}

// byteOrderMark opens a file written by an editor that marks its encoding.
var byteOrderMark = []byte{0xEF, 0xBB, 0xBF}

// parseRules compiles an ignore file's lines. file names it in every Source, and base is the directory its
// anchored patterns are relative to.
func parseRules(contents []byte, file string, base string) ([]rule, error) {
	contents = bytes.TrimPrefix(contents, byteOrderMark)
	var rules []rule
	for index, line := range strings.Split(string(contents), "\n") {
		line = strings.TrimSuffix(line, "\r")
		if line == "" || line[0] == '#' {
			continue
		}
		if strings.IndexByte(line, 0) >= 0 {
			return nil, fmt.Errorf("gitignore: %s:%d holds a NUL byte, which no pattern can contain", file, index+1)
		}
		line = trimTrailingSpaces(line)
		rules = append(rules, compileRule(line, file, index+1, base))
	}
	return rules, nil
}

// trimTrailingSpaces drops the spaces ending a line, except those a backslash escapes. Only spaces: a
// trailing tab is part of the pattern.
func trimTrailingSpaces(line string) string {
	// spacesStart is where the current run of unescaped spaces began, or -1 outside one.
	spacesStart := -1
	for index := 0; index < len(line); index++ {
		switch line[index] {
		case ' ':
			if spacesStart < 0 {
				spacesStart = index
			}
		case '\\':
			// The escaped byte is kept whatever it is, a space included.
			index++
			spacesStart = -1
		default:
			spacesStart = -1
		}
	}
	if spacesStart >= 0 {
		return line[:spacesStart]
	}
	return line
}

// compileRule reads one pattern line: an optional `!`, the pattern, and an optional trailing `/`.
func compileRule(line string, file string, lineNumber int, base string) rule {
	compiled := rule{base: base, source: Source{File: file, Line: lineNumber}}
	pattern := line
	if strings.HasPrefix(pattern, "!") {
		compiled.negated = true
		pattern = pattern[1:]
	}
	if strings.HasSuffix(pattern, "/") {
		compiled.directoryOnly = true
		pattern = pattern[:len(pattern)-1]
	}

	compiled.source.Pattern = pattern
	if compiled.negated {
		compiled.source.Pattern = "!" + compiled.source.Pattern
	}
	if compiled.directoryOnly {
		compiled.source.Pattern += "/"
	}

	compiled.anchored = strings.Contains(pattern, "/")
	if compiled.anchored {
		compiled.glob = compileGlob(strings.TrimPrefix(pattern, "/"), true)
	} else {
		compiled.glob = compileGlob(pattern, false)
	}
	return compiled
}

// cleanRelative normalizes a relative path to the slash form the matcher keys on, "" for the root. A walk
// passes paths already in that form, which are returned as they are without allocating.
func cleanRelative(relative string) string {
	if isCleanRelative(relative) {
		return relative
	}
	relative = path.Clean(filepath.ToSlash(relative))
	if relative == "." || relative == "/" {
		return ""
	}
	return strings.TrimPrefix(relative, "/")
}

// isCleanRelative reports whether a path is already what cleanRelative would make it: slash separated, with
// no empty, `.` or `..` segment and no leading or trailing slash.
func isCleanRelative(relative string) bool {
	if relative == "" {
		return true
	}
	if strings.ContainsRune(relative, '\\') && filepath.Separator == '\\' {
		return false
	}
	start := 0
	for index := 0; index <= len(relative); index++ {
		if index < len(relative) && relative[index] != '/' {
			continue
		}
		if segment := relative[start:index]; segment == "" || segment == "." || segment == ".." {
			return false
		}
		start = index + 1
	}
	return true
}

func joinRelative(directory string, name string) string {
	if directory == "" {
		return name
	}
	return directory + "/" + name
}

func displayDirectory(directory string) string {
	if directory == "" {
		return "the repository root"
	}
	return directory
}
