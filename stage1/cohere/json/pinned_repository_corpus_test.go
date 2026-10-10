package json

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
)

// Freeze repository JSON so selection need not rerun this package on every
// .json edit. These inputs move only when someone explicitly bumps this pin.
// The checkout that builds the corpus product must already hold this commit (a
// shallow one fetches it at checkout, git fetch --depth=1 origin <pin>); the
// corpus never fetches, and a missing pin is fatal.
const repositoryCorpusCommit = "2b4dbde172c1c2f23c065ecae8b906f8fcb72a20"

// SHA256 of sorted name<TAB>SHA256(bytes)<LF>, including upstream inputs and
// generated controls. This is the original 3,900-case working-tree identity.
const repositoryCorpusIdentity = "09cd6c1c6a0fb423c2bf9cc1a56f5670218eca4088f0ba7a49ac26b58652cae6"

// The corpus files corpus.pin names under stage3/api/node_modules: npm ci --prefix stage3/api installs them from
// stage3/api's lockfile, so no archive of tracked files carries them.
const provisionedCorpusPrefix = "stage3/api/node_modules/"

// gitPinnedInput runs Git on the checkout's own objects and nothing else. A test that reads the network isn't
// deterministic, and a sandboxed runner has none (#arw7837). protocol.allow=never refuses every transport, so
// neither a fetch nor a partial clone's lazy fetch of a missing object can reach a remote on any Git version (-c
// reaches the lazy fetch's own git process through GIT_CONFIG_PARAMETERS); GIT_NO_LAZY_FETCH (Git 2.44) stops that
// fetch before it starts.
func gitPinnedInput(root string, input []byte, arguments ...string) ([]byte, error) {
	command := exec.CommandContext(context.Background(), "git", append([]string{"-C", root, "-c", "protocol.allow=never"}, arguments...)...)
	command.Env = append(os.Environ(), "GIT_NO_LAZY_FETCH=1")
	command.Stdin = bytes.NewReader(input)
	output, err := childguard.CombinedOutput(command, jsonGuard)
	if err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", strings.Join(arguments, " "), err, strings.TrimSpace(string(output)))
	}
	return output, nil
}

// The corpus product holds every input a runner's source archive lacks: the pin's repository JSON (Git objects,
// and Loom's source archive is the tree's tracked files with no history) and the provisioned node_modules files
// corpus.pin names (untracked). Workshop builds it from a checkout that holds both, and the tests read only it, the
// source archive and the cohere submodule's tracked files. Git objects are content addressed, so the pin alone
// names the repository bytes; the lockfile names the provisioned ones, and corpus.pin checks both byte for byte.
func jsonPinnedCorpusInputs() buildcache.Inputs {
	return buildcache.Inputs{
		Name:  "json pinned corpus",
		Files: []string{"stage1/cohere/json/pinned_repository_corpus_test.go", "stage1/cohere/json/testdata/corpus.pin", "stage3/api/package-lock.json"},
		Flags: []string{"repository JSON at " + repositoryCorpusCommit, "provisioned " + provisionedCorpusPrefix + " inputs named by corpus.pin"},
	}
}

func buildJSONPinnedCorpus(directory string) error {
	root, err := filepath.Abs(repository)
	if err != nil {
		return err
	}
	tracked, err := repositoryCasesAtPin(root, repositoryCorpusCommit)
	if err != nil {
		return err
	}
	provisioned, err := provisionedCorpusFiles(root)
	if err != nil {
		return err
	}
	if err := writeCorpusRecords(filepath.Join(directory, "repository"), tracked); err != nil {
		return err
	}
	return writeCorpusRecords(filepath.Join(directory, "provisioned"), provisioned)
}

// The product test and every corpus reader share one recipe and key.
func TestProduct_JSONPinnedCorpus(t *testing.T) {
	t.Parallel()
	buildcache.Product(t, jsonPinnedCorpusInputs(), buildJSONPinnedCorpus)
}

// Every test in a process reads the product once. Callers get slice copies because hash partitioning sorts its
// input in place.
var pinnedCorpusRead struct {
	once                 sync.Once
	tracked, provisioned []textCase
	err                  error
}

func pinnedCorpusProduct(t testing.TB) (tracked, provisioned []textCase) {
	t.Helper()
	pinnedCorpusRead.once.Do(func() {
		directory, err := buildcache.Get(jsonPinnedCorpusInputs(), buildJSONPinnedCorpus)
		if err == nil {
			pinnedCorpusRead.tracked, err = readCorpusRecords(filepath.Join(directory, "repository"))
		}
		if err == nil {
			pinnedCorpusRead.provisioned, err = readCorpusRecords(filepath.Join(directory, "provisioned"))
		}
		if err != nil {
			pinnedCorpusRead.err = fmt.Errorf("json pinned corpus product: %w", err)
		}
	})
	if pinnedCorpusRead.err != nil {
		t.Fatal(pinnedCorpusRead.err)
	}
	return append([]textCase(nil), pinnedCorpusRead.tracked...), append([]textCase(nil), pinnedCorpusRead.provisioned...)
}

// A product file holds its cases the way Git's cat-file --batch does: "<size> <name>\n", the bytes, "\n".
func writeCorpusRecords(path string, cases []textCase) error {
	var records bytes.Buffer
	for _, item := range cases {
		if item.Name == "" || strings.ContainsAny(item.Name, "\r\n") {
			return fmt.Errorf("corpus product %s: unsupported case name %q", path, item.Name)
		}
		fmt.Fprintf(&records, "%d %s\n", len(item.Text), item.Name)
		records.WriteString(item.Text)
		records.WriteByte('\n')
	}
	return os.WriteFile(path, records.Bytes(), 0o644)
}

func readCorpusRecords(path string) ([]textCase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cases []textCase
	for len(data) > 0 {
		header, rest, terminated := bytes.Cut(data, []byte("\n"))
		sizeText, name, named := strings.Cut(string(header), " ")
		size, err := strconv.Atoi(sizeText)
		if !terminated || !named || name == "" || err != nil || size < 0 || len(rest) <= size || rest[size] != '\n' {
			if len(header) > 120 {
				header = header[:120]
			}
			return nil, fmt.Errorf("corpus product %s: malformed record after %d cases at %q", path, len(cases), header)
		}
		cases = append(cases, textCase{Name: name, Text: string(rest[:size])})
		data = rest[size+1:]
	}
	return cases, nil
}

// The corpus.pin manifest beside this file, read from the repository root.
func corpusPinManifest(root string) (corpusPin, error) {
	var manifest corpusPin
	encoded, err := os.ReadFile(filepath.Join(root, "stage1/cohere/json/testdata/corpus.pin"))
	if err != nil {
		return manifest, err
	}
	return manifest, json.Unmarshal(encoded, &manifest)
}

// provisionedCorpusFiles reads the node_modules files corpus.pin names from the product builder's disk, where npm ci
// --prefix stage3/api put them.
func provisionedCorpusFiles(root string) ([]textCase, error) {
	manifest, err := corpusPinManifest(root)
	if err != nil {
		return nil, err
	}
	var cases []textCase
	for _, item := range manifest.Cases {
		if !strings.HasPrefix(item.Path, provisionedCorpusPrefix) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(item.Path)))
		if err != nil {
			return nil, fmt.Errorf("pinned corpus input %s (npm ci --prefix stage3/api provisions it on the corpus product's builder): %w", item.Path, err)
		}
		if !utf8.Valid(data) {
			return nil, fmt.Errorf("non-UTF-8 pinned corpus file %s", item.Path)
		}
		cases = append(cases, textCase{Name: item.Path, Text: string(data)})
	}
	return cases, nil
}

// repositoryCasesAtPin reads the pin's JSON from the checkout's own objects, on the corpus product's builder.
// The TypeScript corpusfiles selector returns live paths, not pinned bytes.
func repositoryCasesAtPin(root, pin string) ([]textCase, error) {
	if len(pin) != 40 || strings.Trim(pin, "0123456789abcdef") != "" {
		return nil, fmt.Errorf("repository JSON: invalid commit pin %q", pin)
	}
	if _, err := gitPinnedInput(root, nil, "cat-file", "-e", pin+"^{commit}"); err != nil {
		return nil, fmt.Errorf("repository JSON pin %s: environment: the commit is not in the checkout at %s, and the corpus never fetches; a shallow checkout carries it with git fetch --depth=1 origin %s before the run", pin, root, pin)
	}
	tree, err := gitPinnedInput(root, nil, "ls-tree", "-r", "-z", pin)
	if err != nil {
		return nil, fmt.Errorf("repository JSON pin %s: %w", pin, err)
	}
	var names []string
	blobs := map[string]string{}
	for _, entry := range strings.Split(string(tree), "\x00") {
		// <mode> <type> <object>TAB<path>
		header, name, found := strings.Cut(entry, "\t")
		fields := strings.Fields(header)
		if found && len(fields) == 3 && fields[1] == "blob" && strings.HasSuffix(name, ".json") {
			names = append(names, name)
			blobs[name] = fields[2]
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("repository JSON pin %s has no JSON files", pin)
	}
	sort.Strings(names)
	if err := pinnedBlobsPresent(root, pin, names, blobs); err != nil {
		return nil, err
	}
	var requests strings.Builder
	for _, name := range names {
		if strings.ContainsAny(name, "\r\n") {
			return nil, fmt.Errorf("repository JSON pin %s: unsupported newline in path %q", pin, name)
		}
		fmt.Fprintf(&requests, "%s:%s\n", pin, name)
	}
	batch, err := gitPinnedInput(root, []byte(requests.String()), "cat-file", "--batch")
	if err != nil {
		return nil, fmt.Errorf("repository JSON pin %s: %w", pin, err)
	}
	reader := bufio.NewReader(bytes.NewReader(batch))
	cases := make([]textCase, 0, len(names))
	for _, name := range names {
		header, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("repository JSON pin %s path %s: %w", pin, name, err)
		}
		fields := strings.Fields(header)
		if len(fields) != 3 || fields[1] != "blob" {
			return nil, fmt.Errorf("repository JSON pin %s path %s: invalid blob header %q", pin, name, header)
		}
		size, err := strconv.Atoi(fields[2])
		if err != nil || size < 0 {
			return nil, fmt.Errorf("repository JSON pin %s path %s: invalid blob size %q", pin, name, fields[2])
		}
		data := make([]byte, size)
		if _, err := io.ReadFull(reader, data); err != nil {
			return nil, fmt.Errorf("repository JSON pin %s path %s: %w", pin, name, err)
		}
		newline, err := reader.ReadByte()
		if err != nil || newline != '\n' {
			return nil, fmt.Errorf("repository JSON pin %s path %s: missing blob terminator", pin, name)
		}
		if !utf8.Valid(data) {
			return nil, fmt.Errorf("non-UTF-8 corpus file %s at pin %s", name, pin)
		}
		cases = append(cases, textCase{Name: name, Text: string(data)})
	}
	if _, err := reader.ReadByte(); err != io.EOF {
		return nil, fmt.Errorf("repository JSON pin %s: trailing batch output", pin)
	}
	return cases, nil
}

// Loom's pool units are blobless clones (--filter=blob:none): the pin's commit and trees are there, but only the
// blobs a checkout wrote. cat-file would try to fetch each missing one on its own, which gitPinnedInput refuses; this
// names every JSON file whose blob is missing first, so the failure says which inputs the checkout lacks.
func pinnedBlobsPresent(root, pin string, names []string, blobs map[string]string) error {
	// --no-walk: the pin's own tree, not every object reachable through its history. --missing=print lists an absent
	// object rather than fetching it.
	objects, err := gitPinnedInput(root, nil, "rev-list", "--objects", "--no-walk", "--missing=print", pin)
	if err != nil {
		return fmt.Errorf("repository JSON pin %s: %w", pin, err)
	}
	absent := map[string]bool{}
	for _, line := range strings.Split(string(objects), "\n") {
		if object, found := strings.CutPrefix(line, "?"); found {
			absent[object] = true
		}
	}
	var missing []string
	for _, name := range names {
		if absent[blobs[name]] {
			missing = append(missing, name)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	named := strings.Join(missing, ", ")
	if len(missing) > 5 {
		named = strings.Join(missing[:5], ", ") + fmt.Sprintf(" and %d more", len(missing)-5)
	}
	return fmt.Errorf("repository JSON pin %s: environment: %d of its %d JSON files have no blob in the checkout at %s, and the corpus never fetches: %s", pin, len(missing), len(names), root, named)
}

// pinnedCorpusFiles is every non-generated input: the pin's repository JSON and the provisioned files from the
// corpus product, and the pinned submodule paths from the checkout's own tracked files.
func pinnedCorpusFiles(t testing.TB, root string) ([]textCase, error) {
	tracked, provisioned := pinnedCorpusProduct(t)
	cases := append(tracked, provisioned...)
	manifest, err := corpusPinManifest(root)
	if err != nil {
		return nil, err
	}
	// Only pinned submodule paths are read from disk. Arbitrary repository JSON,
	// additions, deletions and dirty index entries cannot enter.
	for _, item := range manifest.Cases {
		if strings.HasPrefix(item.Path, "generated/") || strings.HasPrefix(item.Path, provisionedCorpusPrefix) {
			continue
		}
		if !strings.HasPrefix(item.Path, "cohere/") {
			return nil, fmt.Errorf("unexpected disk input in corpus pin: %s", item.Path)
		}
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(item.Path)))
		if err != nil {
			return nil, fmt.Errorf("pinned corpus input %s: %w", item.Path, err)
		}
		if !utf8.Valid(data) {
			return nil, fmt.Errorf("non-UTF-8 pinned corpus file %s", item.Path)
		}
		cases = append(cases, textCase{Name: item.Path, Text: string(data)})
	}
	sort.Slice(cases, func(i, j int) bool {
		a, b := strings.Split(cases[i].Name, "/"), strings.Split(cases[j].Name, "/")
		for k := 0; k < len(a) && k < len(b); k++ {
			if a[k] != b[k] {
				return a[k] < b[k]
			}
		}
		return len(a) < len(b)
	})
	return cases, nil
}

func repositoryCorpusIdentityError(pin string, cases []textCase, count int, identity string) error {
	got := pinForCases(cases)
	if got.Count != count || got.SHA256 != identity {
		return fmt.Errorf("repository JSON pin %s: corpus identity changed: %d cases SHA256 %s; want %d SHA256 %s (bump pin and identity together)", pin, got.Count, got.SHA256, count, identity)
	}
	return nil
}

func TestPinnedRepositoryCorpusIdentity(t *testing.T) {
	t.Parallel()
	cases := corpusCases(t)
	if err := repositoryCorpusIdentityError(repositoryCorpusCommit, cases, 3900, repositoryCorpusIdentity); err != nil {
		t.Fatal(err)
	}
	t.Logf("3900 cases; name/byte identity SHA256 %s", pinForCases(cases).SHA256)
}

func TestPinnedRepositoryCorpusIgnoresWorkingTree(t *testing.T) {
	t.Parallel()
	fixture := newRepositoryFixture(t)
	pin, err := gitCorpus(fixture.root, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	before, err := repositoryCasesAtPin(fixture.root, strings.TrimSpace(string(pin)))
	if err != nil {
		t.Fatal(err)
	}
	fixtureWrite(t, fixture.root, "tracked.json", "{changed:true}")
	fixtureWrite(t, fixture.root, "added.json", "[42]")
	fixtureGit(t, fixture.root, "add", "tracked.json", "added.json")
	if err := os.Remove(filepath.Join(fixture.root, "nested/second.json")); err != nil {
		t.Fatal(err)
	}
	after, err := repositoryCasesAtPin(fixture.root, strings.TrimSpace(string(pin)))
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(before) != fmt.Sprint(after) {
		t.Fatal("working-tree/index mutation changed pinned names or bytes")
	}
	t.Log("staged edit/addition and worktree deletion left pinned names and bytes unchanged")
}

func TestPinnedRepositoryCorpusPinMutant(t *testing.T) {
	t.Parallel()
	fixture := newRepositoryFixture(t)
	pin, err := gitCorpus(fixture.root, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	before, err := repositoryCasesAtPin(fixture.root, strings.TrimSpace(string(pin)))
	if err != nil {
		t.Fatal(err)
	}
	fixtureWrite(t, fixture.root, "pin-added.json", "[42]")
	commitFixtureJSON(t, fixture.root, "pin-added.json")
	mutant, err := gitCorpus(fixture.root, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	after, err := repositoryCasesAtPin(fixture.root, strings.TrimSpace(string(mutant)))
	if err != nil {
		t.Fatal(err)
	}
	if pinForCases(before).SHA256 == pinForCases(after).SHA256 || len(after) != len(before)+1 {
		t.Fatal("pin mutant did not change its JSON set and identity")
	}
	want := pinForCases(before)
	mutantPin := strings.TrimSpace(string(mutant))
	if repositoryCorpusIdentityError(mutantPin, before, want.Count, want.SHA256) != nil {
		t.Fatal("baseline identity failed")
	}
	if repositoryCorpusIdentityError(mutantPin, after, want.Count, want.SHA256) == nil {
		t.Fatal("pin mutant survived the identity check")
	}
	t.Logf("pin mutant %s rejected: %v", strings.TrimSpace(string(mutant)), repositoryCorpusIdentityError(mutantPin, after, want.Count, want.SHA256))
}

func validatePinnedCorpus(t testing.TB, cases []textCase, expected corpusPin) (corpusPin, int, error) {
	tracked, _ := pinnedCorpusProduct(t)
	state := repositoryJSON{paths: map[string]bool{}}
	for _, item := range tracked {
		state.paths[item.Name] = true
	}
	return validateCorpusState(state, cases, expected)
}

func TestJSONPinnedCorpusRecords(t *testing.T) {
	t.Parallel()
	cases := []textCase{{"a.json", "{}"}, {"with space/b.json", ""}, {"c.json", "[1,\n2]\n\n"}, {"d.json", "7 e.json\n"}}
	path := filepath.Join(t.TempDir(), "records")
	if err := writeCorpusRecords(path, cases); err != nil {
		t.Fatal(err)
	}
	read, err := readCorpusRecords(path)
	if err != nil || fmt.Sprint(read) != fmt.Sprint(cases) {
		t.Fatalf("records changed names or bytes: %v %q", err, read)
	}
	encoded, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// A product cut short, or a record whose size is off by one, is refused rather than read as fewer or other cases.
	for _, mutant := range []struct{ name, text string }{
		{"truncated", string(encoded[:len(encoded)-1])},
		{"size", strings.Replace(string(encoded), "2 a.json", "1 a.json", 1)},
	} {
		if err := os.WriteFile(path, []byte(mutant.text), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := readCorpusRecords(path); err == nil || !strings.Contains(err.Error(), "malformed record") {
			t.Fatalf("%s product read without error: %v", mutant.name, err)
		}
	}
	if err := writeCorpusRecords(path, []textCase{{"new\nline.json", "{}"}}); err == nil {
		t.Fatal("a name with a newline was written")
	}
}

// Not parallel: the process environment (GIT_TRACE2_EVENT, set by t.Setenv to trace git's fetches).
func TestPinnedRepositoryCorpusNeverFetches(t *testing.T) {
	fixture := newRepositoryFixture(t)
	head, err := gitCorpus(fixture.root, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	pin := strings.TrimSpace(string(head))
	full, err := repositoryCasesAtPin(fixture.root, pin)
	if err != nil {
		t.Fatal(err)
	}
	// A shallow gate checkout without the pin, whose origin holds it and would serve it to any fetch: a read that
	// fetched would succeed here.
	shallow := t.TempDir()
	fixtureGit(t, shallow, "init", "-q")
	fixtureGit(t, shallow, "remote", "add", "origin", fixture.root)
	var missing error
	if fetches := gitFetchesTracing(t, func() { _, missing = repositoryCasesAtPin(shallow, pin) }); fetches != 0 {
		t.Fatalf("a checkout without the pin fetched %d times, want none", fetches)
	}
	if missing == nil || !strings.Contains(missing.Error(), pin) || !strings.Contains(missing.Error(), "not in the checkout") {
		t.Fatalf("a checkout without the pin read the corpus anyway, or lost the pin's name: %v", missing)
	}
	// The checkout step carries the pin, as a runner's must; the read then fetches nothing of its own.
	fixtureGit(t, shallow, "fetch", "--quiet", "--depth=1", "origin", pin)
	var after []textCase
	if fetches := gitFetchesTracing(t, func() { after, err = repositoryCasesAtPin(shallow, pin) }); fetches != 0 || err != nil {
		t.Fatalf("a shallow checkout carrying the pin: %d fetches, %v", fetches, err)
	}
	if fmt.Sprint(full) != fmt.Sprint(after) {
		t.Fatal("a shallow checkout carrying the pin read different names or bytes")
	}
	t.Logf("without the pin: %v; with it at depth 1: %d identical inputs, no fetch", missing, len(after))
}

// gitFetchesTracing counts the git processes a function starts whose arguments include fetch, from Git's trace2
// event stream.
func gitFetchesTracing(t *testing.T, run func()) int {
	t.Helper()
	trace := filepath.Join(t.TempDir(), "trace2.json")
	t.Setenv("GIT_TRACE2_EVENT", trace)
	run()
	encoded, err := os.ReadFile(trace)
	if err != nil {
		t.Fatal(err)
	}
	fetches := 0
	for _, line := range strings.Split(string(encoded), "\n") {
		var event struct {
			Event string   `json:"event"`
			Argv  []string `json:"argv"`
		}
		if json.Unmarshal([]byte(line), &event) != nil || event.Event != "start" {
			continue
		}
		for _, argument := range event.Argv {
			if argument == "fetch" {
				fetches++
				break
			}
		}
	}
	return fetches
}

// Not parallel: the process environment (GIT_TRACE2_EVENT, set by t.Setenv to trace git's fetches).
func TestPinnedRepositoryCorpusBloblessCloneNeverFetches(t *testing.T) {
	fixture := newRepositoryFixture(t)
	for index := 0; index < 8; index++ {
		fixtureWrite(t, fixture.root, fmt.Sprintf("batch/%d.json", index), fmt.Sprintf("[%d]", index))
	}
	commitFixtureJSON(t, fixture.root, "batch")
	head, err := gitCorpus(fixture.root, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	pin := strings.TrimSpace(string(head))
	fixtureGit(t, fixture.root, "config", "uploadpack.allowFilter", "true")
	fixtureGit(t, fixture.root, "config", "uploadpack.allowAnySHA1InWant", "true")
	// The pool's checkout: every commit and tree, no blobs, and an origin that would serve each one.
	clone := filepath.Join(t.TempDir(), "blobless")
	fixtureGit(t, fixture.root, "clone", "--quiet", "--filter=blob:none", "--no-checkout", "file://"+fixture.root, clone)
	var missing error
	if fetches := gitFetchesTracing(t, func() { _, missing = repositoryCasesAtPin(clone, pin) }); fetches != 0 {
		t.Fatalf("a blobless clone fetched %d times, want none", fetches)
	}
	// A missing input is the checkout's, named, and never reads as a wrong answer (devtools/red-sort's WRONG).
	wrong := regexp.MustCompile(`(?i)compile error|build.failed|undefined:|syntax error|assertion|\b(?:got|want|expected)\b|assert.*diff`)
	if missing == nil || !strings.Contains(missing.Error(), pin) || !strings.Contains(missing.Error(), "10 of its 10 JSON files") ||
		!strings.Contains(missing.Error(), "batch/0.json") || wrong.MatchString(missing.Error()) {
		t.Fatalf("a blobless clone read the pinned corpus anyway, or didn't name what it lacks: %v", missing)
	}
	// A full clone has every blob and fetches nothing.
	if fetches := gitFetchesTracing(t, func() { _, err = repositoryCasesAtPin(fixture.root, pin) }); fetches != 0 || err != nil {
		t.Fatalf("a full clone: %d fetches, %v", fetches, err)
	}
	t.Log(missing)
}
