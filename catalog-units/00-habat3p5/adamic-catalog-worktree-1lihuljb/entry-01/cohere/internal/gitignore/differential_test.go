package gitignore

import (
	"bytes"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// The differential: on a real tree, every path's verdict and deciding rule equals git check-ignore's.
//
// git runs here and nowhere else, as the oracle. It reads the trees named by COHERE_GITIGNORE_TREES
// (colon separated), and every repository nested in them is compared as a repository of its own, the way
// a walk meets it. git runs under sandbox-exec, with no network and no writes outside a temporary
// directory, and the sandbox is shown to hold before git reads anything.
func TestTheMatcherAgreesWithGitOnRealTrees(t *testing.T) {
	t.Parallel()
	trees := os.Getenv("COHERE_GITIGNORE_TREES")
	if trees == "" {
		t.Skip("COHERE_GITIGNORE_TREES names no trees to compare")
	}
	oracle := newOracle(t)
	for _, tree := range strings.Split(trees, ":") {
		pending := []string{tree}
		for len(pending) > 0 {
			root := pending[0]
			pending = pending[1:]
			comparison := compareWithGit(t, oracle, root)
			pending = append(pending, comparison.nested...)
			t.Logf("%s: %d paths compared (%d ignored, %d re-included by a negation, %d decided by no rule), %d ignore files read, %d nested repositories, %d differences",
				root, comparison.compared, comparison.ignored, comparison.reincluded, comparison.undecided,
				comparison.ignoreFiles, len(comparison.nested), len(comparison.differences))
			for index, difference := range comparison.differences {
				if index == 20 {
					t.Errorf("%s: %d more differences", root, len(comparison.differences)-20)
					break
				}
				t.Errorf("%s: %s", root, difference)
			}
			if comparison.compared == 0 {
				t.Errorf("%s: no path was compared, so agreeing proves nothing", root)
			}
		}
	}
}

// The same differential over generated trees: random names, random ignore files at random depths, and
// patterns drawn from every construct gitignore(5) describes, so the matcher meets combinations no real
// tree holds. COHERE_GITIGNORE_GENERATED is how many trees, COHERE_GITIGNORE_SEED fixes the sequence.
func TestTheMatcherAgreesWithGitOnGeneratedTrees(t *testing.T) {
	t.Parallel()
	count, _ := strconv.Atoi(os.Getenv("COHERE_GITIGNORE_GENERATED"))
	if count <= 0 {
		t.Skip("COHERE_GITIGNORE_GENERATED asks for no generated trees")
	}
	seed, err := strconv.ParseInt(os.Getenv("COHERE_GITIGNORE_SEED"), 10, 64)
	if err != nil {
		seed = rand.Int63()
	}
	t.Logf("seed %d", seed)
	random := rand.New(rand.NewSource(seed))
	oracle := newOracle(t)

	compared, ignored, reincluded, failing := 0, 0, 0, 0
	for tree := 0; tree < count; tree++ {
		root := filepath.Join(oracle.scratch, fmt.Sprintf("generated-%d", tree))
		exclude := generateTree(t, random, root)
		oracle.run(t, root, nil, "init", "--quiet")
		if exclude != "" {
			writeTree(t, root, map[string]string{ExcludeFile: exclude})
		}
		comparison := compareWithGit(t, oracle, root)
		compared += comparison.compared
		ignored += comparison.ignored
		reincluded += comparison.reincluded
		if len(comparison.differences) > 0 {
			failing++
			if failing <= 5 {
				t.Errorf("generated tree %d (%s), %d differences, first: %s", tree, root, len(comparison.differences), comparison.differences[0])
			}
		}
	}
	t.Logf("%d generated trees, %d paths compared (%d ignored, %d re-included by a negation), %d trees with a difference",
		count, compared, ignored, reincluded, failing)
	// Trees whose rules decided nothing would agree with git without testing the matcher.
	if compared == 0 || ignored == 0 || reincluded == 0 {
		t.Errorf("the generated trees exercised too little: %d compared, %d ignored, %d re-included", compared, ignored, reincluded)
	}
}

// oracle runs git sandboxed.
type oracle struct {
	scratch      string
	profile      string
	emptyExclude string
}

func newOracle(t *testing.T) *oracle {
	t.Helper()
	if runtime.GOOS != "darwin" {
		t.Skip("the oracle sandbox is sandbox-exec, which is macOS only")
	}
	scratch, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	profile := filepath.Join(scratch, "oracle.sb")
	rules := fmt.Sprintf(`(version 1)
(allow default)
(deny network*)
(deny file-write*)
(allow file-write* (subpath %q) (literal "/dev/null") (regex #"^/dev/fd/") (literal "/dev/stdout") (literal "/dev/stderr"))
`, scratch)
	if err := os.WriteFile(profile, []byte(rules), 0o644); err != nil {
		t.Fatal(err)
	}
	emptyExclude := filepath.Join(scratch, "empty-excludes")
	if err := os.WriteFile(emptyExclude, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	result := &oracle{scratch: scratch, profile: profile, emptyExclude: emptyExclude}

	// The sandbox must hold before git reads a real tree: a write outside the scratch directory and a
	// network connection both fail, and a write inside succeeds.
	outside := filepath.Join(os.Getenv("HOME"), fmt.Sprintf(".cohere-sandbox-probe-%d", os.Getpid()))
	if err := exec.Command("sandbox-exec", "-f", profile, "/usr/bin/touch", outside).Run(); err == nil {
		_ = os.Remove(outside)
		t.Fatal("the sandbox let a write outside its scratch directory through")
	}
	if err := exec.Command("sandbox-exec", "-f", profile, "/usr/bin/touch", filepath.Join(scratch, "inside")).Run(); err != nil {
		t.Fatalf("the sandbox refused a write inside its scratch directory: %v", err)
	}
	if err := exec.Command("sandbox-exec", "-f", profile, "/usr/bin/nc", "-z", "-G", "2", "1.1.1.1", "443").Run(); err == nil {
		t.Fatal("the sandbox let a network connection through")
	}
	return result
}

// run runs git in repository, sandboxed, with no configuration but the repository's own, no global
// excludes file, and case sensitive matching.
func (oracle *oracle) run(t *testing.T, repository string, stdin []byte, arguments ...string) []byte {
	t.Helper()
	command := exec.Command("sandbox-exec", append([]string{"-f", oracle.profile, "git",
		"-c", "core.excludesFile=" + oracle.emptyExclude, "-c", "core.ignorecase=false"}, arguments...)...)
	command.Dir = repository
	command.Env = []string{
		"PATH=/usr/bin:/bin", "HOME=" + oracle.scratch, "XDG_CONFIG_HOME=" + oracle.scratch,
		"GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_OPTIONAL_LOCKS=0",
	}
	command.Stdin = bytes.NewReader(stdin)
	var standardError bytes.Buffer
	command.Stderr = &standardError
	output, err := command.Output()
	// check-ignore exits 1 when nothing is ignored, which is an answer, not a failure.
	if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 && arguments[0] == "check-ignore" {
		err = nil
	}
	if err != nil {
		t.Fatalf("git %s in %s: %v\n%s", strings.Join(arguments, " "), repository, err, standardError.String())
	}
	return output
}

// verdict is one path's answer.
type verdict struct {
	ignored bool
	source  Source
}

type comparison struct {
	compared, ignored, reincluded, undecided, ignoreFiles int
	nested                                                []string
	differences                                           []string
}

// compareWithGit decides every path in the repository at root with the matcher and with git, and lists
// where they differ. Nested repositories are not entered; they are returned for their own comparison.
func compareWithGit(t *testing.T, oracle *oracle, root string) comparison {
	t.Helper()
	var result comparison
	matcher, err := New(root)
	if err != nil {
		t.Fatalf("%s: %v", root, err)
	}

	ours := map[string]verdict{}
	var paths []string
	var walk func(scope *Matcher, directory string)
	walk = func(scope *Matcher, directory string) {
		entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(directory)))
		if err != nil {
			t.Fatalf("%s: %v", root, err)
		}
		for _, entry := range entries {
			name := entry.Name()
			relative := joinRelative(directory, name)
			if name == ".git" {
				continue
			}
			if name == IgnoreFileName {
				result.ignoreFiles++
			}
			isDirectory := entry.Type()&fs.ModeType == fs.ModeDir
			if isDirectory {
				if _, err := os.Lstat(filepath.Join(root, filepath.FromSlash(relative), ".git")); err == nil {
					result.nested = append(result.nested, filepath.Join(root, filepath.FromSlash(relative)))
					continue
				}
			}
			ignored, source := scope.Ignored(relative, isDirectory)
			ours[relative] = verdict{ignored, source}
			paths = append(paths, relative)
			if isDirectory {
				// Below an excluded directory every path takes the same answer, from the rule that excluded
				// the directory. Its own entries are compared, which shows the answer is inherited; going
				// deeper (node_modules, a build cache) would compare that one fact many thousands of times.
				if excluded, _ := scope.Excluded(); excluded {
					continue
				}
				child, err := scope.Enter(relative)
				if err != nil {
					t.Fatalf("%s: %v", root, err)
				}
				walk(child, relative)
			}
		}
	}
	walk(matcher, "")

	var input bytes.Buffer
	for _, relative := range paths {
		input.WriteString(relative)
		input.WriteByte(0)
	}
	output := oracle.run(t, root, input.Bytes(), "check-ignore", "--no-index", "--verbose", "--non-matching", "-z", "--stdin")
	fields := bytes.Split(bytes.TrimSuffix(output, []byte{0}), []byte{0})
	if len(output) > 0 && len(fields)%4 != 0 {
		t.Fatalf("%s: git printed %d fields, not a multiple of four", root, len(fields))
	}
	theirs := map[string]verdict{}
	for index := 0; index+3 < len(fields); index += 4 {
		file, line, pattern, relative := string(fields[index]), string(fields[index+1]), string(fields[index+2]), string(fields[index+3])
		if file == "" {
			theirs[relative] = verdict{}
			continue
		}
		lineNumber, err := strconv.Atoi(line)
		if err != nil {
			t.Fatalf("%s: git printed line %q for %s", root, line, relative)
		}
		theirs[relative] = verdict{ignored: !strings.HasPrefix(pattern, "!"), source: Source{File: file, Line: lineNumber, Pattern: pattern}}
	}

	sort.Strings(paths)
	for _, relative := range paths {
		mine := ours[relative]
		gits, answered := theirs[relative]
		if !answered {
			result.differences = append(result.differences, fmt.Sprintf("%s: git gave no answer", relative))
			continue
		}
		result.compared++
		switch {
		case mine.ignored:
			result.ignored++
		case !mine.source.IsZero():
			result.reincluded++
		default:
			result.undecided++
		}
		if mine != gits {
			result.differences = append(result.differences, fmt.Sprintf("%s: ours %v by %q, git %v by %q",
				relative, mine.ignored, mine.source, gits.ignored, gits.source))
		}
	}
	if len(theirs) != len(paths) {
		result.differences = append(result.differences, fmt.Sprintf("git answered for %d paths, %d were asked", len(theirs), len(paths)))
	}
	return result
}

// generateTree writes a random tree with random ignore files under root, and returns patterns for its
// info/exclude, written once git has made the repository, or "" for none.
func generateTree(t *testing.T, random *rand.Rand, root string) string {
	t.Helper()
	names := []string{"a", "b", "ab", "abc", "x.log", "y.tmp", "build", "out", "node_modules", "keep",
		"#hash", "!bang", "sp ace", "trail ", "star*", "q?", "br[ack]et", "back\\slash", "deep", "Dot.go", ".hidden"}
	var directories = []string{""}
	files := map[string]string{}
	for count := 0; count < 12+random.Intn(20); count++ {
		parent := directories[random.Intn(len(directories))]
		name := names[random.Intn(len(names))]
		if random.Intn(3) == 0 && strings.Count(parent, "/") < 4 {
			directory := joinRelative(parent, name)
			files[directory+"/"] = ""
			directories = append(directories, directory)
			continue
		}
		files[joinRelative(parent, name)] = ""
	}
	for _, directory := range directories {
		if random.Intn(2) == 0 {
			files[joinRelative(directory, IgnoreFileName)] = generatePatterns(random, names)
		}
	}
	writeTreeTolerant(t, root, files)
	if random.Intn(2) == 0 {
		return generatePatterns(random, names)
	}
	return ""
}

// writeTreeTolerant writes a generated tree, skipping a file whose name a directory already took.
func writeTreeTolerant(t *testing.T, root string, files map[string]string) {
	t.Helper()
	keys := make([]string, 0, len(files))
	for name := range files {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for _, name := range keys {
		full := filepath.Join(root, filepath.FromSlash(name))
		if strings.HasSuffix(name, "/") {
			_ = os.MkdirAll(full, 0o755)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			continue
		}
		if info, err := os.Stat(full); err == nil && info.IsDir() {
			continue
		}
		if err := os.WriteFile(full, []byte(files[name]), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// generatePatterns writes an ignore file's worth of lines from every construct: literals, `*`, `?`, sets,
// `**` in each position, anchoring, trailing slashes, negation, escapes, comments, blank lines and
// trailing spaces.
func generatePatterns(random *rand.Rand, names []string) string {
	pieces := []string{"*", "?", "**", "[ab]", "[!a]", "[a-c]", "[[:alpha:]]", "\\*", "\\?", "*.log", "*.tmp"}
	var lines []string
	for count := 0; count < 1+random.Intn(8); count++ {
		var segments []string
		for segment := 0; segment < 1+random.Intn(3); segment++ {
			if random.Intn(2) == 0 {
				segments = append(segments, pieces[random.Intn(len(pieces))])
			} else {
				segments = append(segments, escapeName(names[random.Intn(len(names))]))
			}
		}
		line := strings.Join(segments, "/")
		switch random.Intn(10) {
		case 0:
			line = "/" + line
		case 1:
			line += "/"
		case 2:
			line = "!" + line
		case 3:
			line = "# " + line
		case 4:
			line += "   "
		case 5:
			line += "\\ "
		case 6:
			line = ""
		case 7:
			line += "\r"
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n") + "\n"
}

// escapeName writes a file name as a pattern that matches it literally. Live glob characters come from
// the pieces generatePatterns mixes in.
func escapeName(name string) string {
	var escaped strings.Builder
	for index := 0; index < len(name); index++ {
		character := name[index]
		if strings.IndexByte(`*?[]\!#`, character) >= 0 {
			escaped.WriteByte('\\')
		}
		escaped.WriteByte(character)
	}
	return escaped.String()
}
