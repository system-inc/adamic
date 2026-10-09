package json

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"
)

// Freeze repository JSON so selection need not rerun this package on every
// .json edit. These inputs move only when someone explicitly bumps this pin.
// A shallow checkout fetches this exact SHA if absent; failure is fatal.
const repositoryCorpusCommit = "2b4dbde172c1c2f23c065ecae8b906f8fcb72a20"

// SHA256 of sorted name<TAB>SHA256(bytes)<LF>, including upstream inputs and
// generated controls. This is the original 3,900-case working-tree identity.
const repositoryCorpusIdentity = "09cd6c1c6a0fb423c2bf9cc1a56f5670218eca4088f0ba7a49ac26b58652cae6"

// Git objects are immutable. Cache their decoded bytes once per process, and
// return a slice copy because hash partitioning sorts its input in place.
var pinnedRepositoryReads sync.Map

type pinnedRepositoryRead struct {
	once  sync.Once
	cases []textCase
	err   error
}

func cachedRepositoryCasesAtPin(root, pin string) ([]textCase, error) {
	key := struct{ root, pin string }{root, pin}
	value, _ := pinnedRepositoryReads.LoadOrStore(key, &pinnedRepositoryRead{})
	state := value.(*pinnedRepositoryRead)
	state.once.Do(func() { state.cases, state.err = repositoryCasesAtPin(root, pin) })
	return append([]textCase(nil), state.cases...), state.err
}

// Reuse repository_test.go's Git executor for both tree and batch blob reads.
// The TypeScript corpusfiles selector returns live paths, not pinned bytes.
func repositoryCasesAtPin(root, pin string) ([]textCase, error) {
	if len(pin) != 40 || strings.Trim(pin, "0123456789abcdef") != "" {
		return nil, fmt.Errorf("repository JSON: invalid commit pin %q", pin)
	}
	if _, err := gitCorpus(root, "cat-file", "-e", pin+"^{commit}"); err != nil {
		if _, err := gitCorpus(root, "fetch", "--quiet", "--no-tags", "--no-write-fetch-head", "--depth=1", "origin", pin); err != nil {
			return nil, fmt.Errorf("repository JSON pin %s unavailable; fetch failed (no working-tree fallback): %w", pin, err)
		}
	}
	tree, err := gitCorpus(root, "ls-tree", "-r", "--name-only", "-z", pin)
	if err != nil {
		return nil, fmt.Errorf("repository JSON pin %s: %w", pin, err)
	}
	var names []string
	for _, name := range strings.Split(string(tree), "\x00") {
		if strings.HasSuffix(name, ".json") {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("repository JSON pin %s has no JSON files", pin)
	}
	sort.Strings(names)
	var requests strings.Builder
	for _, name := range names {
		if strings.ContainsAny(name, "\r\n") {
			return nil, fmt.Errorf("repository JSON pin %s: unsupported newline in path %q", pin, name)
		}
		fmt.Fprintf(&requests, "%s:%s\n", pin, name)
	}
	batch, err := gitCorpusInput(root, []byte(requests.String()), "cat-file", "--batch")
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

func pinnedCorpusFiles(root, pin string) ([]textCase, error) {
	cases, err := cachedRepositoryCasesAtPin(root, pin)
	if err != nil {
		return nil, err
	}
	encoded, err := os.ReadFile(filepath.Join(root, "stage1/cohere/json/testdata/corpus.pin"))
	if err != nil {
		return nil, err
	}
	var manifest corpusPin
	if err := json.Unmarshal(encoded, &manifest); err != nil {
		return nil, err
	}
	// Only provisioned and pinned submodule paths are read from disk. Arbitrary
	// repository JSON, additions, deletions and dirty index entries cannot enter.
	for _, item := range manifest.Cases {
		if strings.HasPrefix(item.Path, "generated/") {
			continue
		}
		if !strings.HasPrefix(item.Path, "cohere/") && !strings.HasPrefix(item.Path, "stage3/api/node_modules/") {
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

func validatePinnedCorpus(root string, cases []textCase, expected corpusPin) (corpusPin, int, error) {
	tracked, err := cachedRepositoryCasesAtPin(root, repositoryCorpusCommit)
	if err != nil {
		return corpusPin{}, 0, err
	}
	state := repositoryJSON{paths: map[string]bool{}}
	for _, item := range tracked {
		state.paths[item.Name] = true
	}
	return validateCorpusState(state, cases, expected)
}

func TestPinnedRepositoryCorpusUnavailable(t *testing.T) {
	t.Parallel()
	fixture := newRepositoryFixture(t)
	fixtureGit(t, fixture.root, "remote", "add", "origin", t.TempDir())
	missing := strings.Repeat("0", 40)
	if _, err := repositoryCasesAtPin(fixture.root, missing); err == nil || !strings.Contains(err.Error(), missing) || !strings.Contains(err.Error(), "fetch failed") {
		t.Fatalf("missing pin silently skipped or lost pin name: %v", err)
	} else {
		t.Log(err)
	}
}
