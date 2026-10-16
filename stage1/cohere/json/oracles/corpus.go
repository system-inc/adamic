// Package oracles owns json's compiler-free Go reference products.
// Both test binaries resolve relative paths from stage1/cohere/json.
package oracles

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/tracked"
)

type CaseIdentity struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type GroupPin struct {
	Count  int    `json:"count"`
	SHA256 string `json:"sha256"`
}

type CorpusPin struct {
	Count  int                 `json:"count"`
	SHA256 string              `json:"sha256"`
	Groups map[string]GroupPin `json:"groups"`
	Cases  []CaseIdentity      `json:"cases"`
}

// CaseGroup is shared with the parent json tests.
func CaseGroup(path string) string {
	if strings.HasPrefix(path, "generated/") {
		return "generated"
	}
	if strings.HasPrefix(path, "cohere/TypeScript/") {
		return "TypeScript"
	}
	if strings.HasPrefix(path, "cohere/") {
		return "cohere"
	}
	return "provisioned"
}

// PinForCases is shared with the parent json tests.
func PinForCases(cases []ComparisonCase) CorpusPin {
	pin := CorpusPin{Count: len(cases), Groups: map[string]GroupPin{}}
	for _, item := range cases {
		pin.Cases = append(pin.Cases, CaseIdentity{item.Name, fmt.Sprintf("%x", sha256.Sum256([]byte(item.Text)))})
	}
	sort.Slice(pin.Cases, func(i, j int) bool {
		if pin.Cases[i].Path == pin.Cases[j].Path {
			return pin.Cases[i].SHA256 < pin.Cases[j].SHA256
		}
		return pin.Cases[i].Path < pin.Cases[j].Path
	})
	var all strings.Builder
	groups := map[string]*strings.Builder{}
	for _, item := range pin.Cases {
		line := item.Path + "\t" + item.SHA256 + "\n"
		all.WriteString(line)
		group := CaseGroup(item.Path)
		if groups[group] == nil {
			groups[group] = &strings.Builder{}
		}
		groups[group].WriteString(line)
		record := pin.Groups[group]
		record.Count++
		pin.Groups[group] = record
	}
	pin.SHA256 = fmt.Sprintf("%x", sha256.Sum256([]byte(all.String())))
	for group, lines := range groups {
		record := pin.Groups[group]
		record.SHA256 = fmt.Sprintf("%x", sha256.Sum256([]byte(lines.String())))
		pin.Groups[group] = record
	}
	return pin
}

// CorpusPinError is shared with the parent json tests.
func CorpusPinError(want, got CorpusPin) error {
	// Recompute the committed manifest too: a corrupted count/digest cannot pass.
	summarize := func(pin CorpusPin) CorpusPin {
		var lines strings.Builder
		for _, item := range pin.Cases {
			lines.WriteString(item.Path + "\t" + item.SHA256 + "\n")
		}
		pin.SHA256 = fmt.Sprintf("%x", sha256.Sum256([]byte(lines.String())))
		return pin
	}
	if summarize(want).SHA256 != want.SHA256 || want.Count != len(want.Cases) {
		return fmt.Errorf("corpus pin manifest/count/digest is inconsistent")
	}
	var differences []string
	for _, group := range []string{"provisioned", "TypeScript", "cohere", "generated"} {
		a, b := want.Groups[group], got.Groups[group]
		if a != b {
			differences = append(differences, fmt.Sprintf("%s: got %d, pinned %d (identity SHA256 got %s, pinned %s)", group, b.Count, a.Count, b.SHA256, a.SHA256))
		}
	}
	if want.Count == got.Count && want.SHA256 == got.SHA256 && len(differences) == 0 {
		return nil
	}
	expected, observed := map[string]string{}, map[string]string{}
	for _, item := range want.Cases {
		expected[item.Path] = item.SHA256
	}
	for _, item := range got.Cases {
		observed[item.Path] = item.SHA256
	}
	for _, item := range want.Cases {
		hash, exists := observed[item.Path]
		if !exists {
			differences = append(differences, "missing "+item.Path)
		} else if hash != item.SHA256 {
			differences = append(differences, "changed text "+item.Path)
		}
	}
	for _, item := range got.Cases {
		if _, exists := expected[item.Path]; !exists {
			differences = append(differences, "extra "+item.Path)
		}
	}
	return fmt.Errorf("corpus pin: got %d, pinned %d; %s", got.Count, want.Count, strings.Join(differences, "; "))
}

func verifyCorpusPin(t *testing.T, cases []ComparisonCase) {
	t.Helper()
	encoded, err := os.ReadFile("testdata/corpus.pin")
	if err != nil {
		t.Fatalf("%s corpus pin: %v", t.Name(), err)
	}
	var expected CorpusPin
	if err := json.Unmarshal(encoded, &expected); err != nil {
		t.Fatalf("%s corpus pin: %v", t.Name(), err)
	}
	actual, count, err := ValidatePinnedCorpus(t, cases, expected)
	t.Logf("corpus repository: %d tracked JSON files; read from the corpus product at Git pin %s", count, RepositoryCorpusCommit)
	if err != nil {
		t.Fatalf("%s: %v", t.Name(), err)
	}
	for _, group := range []string{"provisioned", "TypeScript", "cohere", "generated"} {
		t.Logf("corpus pin group %s: %d, SHA256 %s", group, actual.Groups[group].Count, actual.Groups[group].SHA256)
	}
}

// PinnedCorpusInputs is shared with the parent json tests.
func PinnedCorpusInputs() (buildcache.Inputs, error) {
	return PinnedCorpusInputsAt(RepositoryCorpusCommit)
}

// PinnedCorpusInputsAt is shared with the parent json tests.
func PinnedCorpusInputsAt(pin string) (buildcache.Inputs, error) {
	jsonTrackedSources.once.Do(func() {
		root, err := filepath.Abs(repository)
		if err != nil {
			jsonTrackedSources.err = err
			return
		}
		entries, err := tracked.Files(root)
		if err != nil {
			jsonTrackedSources.err = err
			return
		}
		for _, entry := range entries {
			if !entry.Submodule() && strings.HasSuffix(entry.Path, ".json") {
				jsonTrackedSources.files = append(jsonTrackedSources.files, entry.Path)
			}
		}
	})
	if jsonTrackedSources.err != nil {
		return buildcache.Inputs{}, jsonTrackedSources.err
	}
	return buildcache.Inputs{
		Name:  "json pinned corpus",
		Files: append([]string{"stage1/cohere/json/pinned_repository_corpus_test.go", "stage1/cohere/json/oracles/corpus.go", "stage1/cohere/json/testdata/corpus.pin", "stage3/api/package-lock.json"}, jsonTrackedSources.files...),
		Flags: []string{"repository JSON at " + pin, "provisioned " + ProvisionedCorpusPrefix + " inputs named by corpus.pin"},
	}, nil
}

// BuildPinnedCorpus is shared with the parent json tests.
func BuildPinnedCorpus(inputs buildcache.Inputs) func(directory string) error {
	return func(directory string) error {
		root, err := filepath.Abs(repository)
		if err != nil {
			return err
		}
		tracked, err := RepositoryCasesAtPin(root, inputs, RepositoryCorpusCommit)
		if err != nil {
			return err
		}
		provisioned, err := provisionedCorpusFiles(root)
		if err != nil {
			return err
		}
		if err := WriteCorpusRecords(filepath.Join(directory, "repository"), tracked); err != nil {
			return err
		}
		return WriteCorpusRecords(filepath.Join(directory, "provisioned"), provisioned)
	}
}

// PinnedCorpusProduct is shared with the parent json tests.
func PinnedCorpusProduct(t testing.TB) (tracked, provisioned []ComparisonCase) {
	t.Helper()
	pinnedCorpusRead.once.Do(func() {
		inputs, err := PinnedCorpusInputs()
		if err != nil {
			pinnedCorpusRead.err = fmt.Errorf("json pinned corpus inputs: %w", err)
			return
		}
		directory, err := buildcache.Get(inputs, func(string) error { return fmt.Errorf("settle stage1/cohere/json TestProduct_JSONPinnedCorpus first") })
		if err == nil {
			pinnedCorpusRead.tracked, err = ReadCorpusRecords(filepath.Join(directory, "repository"))
		}
		if err == nil {
			pinnedCorpusRead.provisioned, err = ReadCorpusRecords(filepath.Join(directory, "provisioned"))
		}
		if err != nil {
			pinnedCorpusRead.err = fmt.Errorf("json pinned corpus product: %w", err)
		}
	})
	if pinnedCorpusRead.err != nil {
		t.Fatal(pinnedCorpusRead.err)
	}
	return append([]ComparisonCase(nil), pinnedCorpusRead.tracked...), append([]ComparisonCase(nil), pinnedCorpusRead.provisioned...)
}

// WriteCorpusRecords is shared with the parent json tests.
func WriteCorpusRecords(path string, cases []ComparisonCase) error {
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

// ReadCorpusRecords is shared with the parent json tests.
func ReadCorpusRecords(path string) ([]ComparisonCase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cases []ComparisonCase
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
		cases = append(cases, ComparisonCase{Name: name, Text: string(rest[:size])})
		data = rest[size+1:]
	}
	return cases, nil
}

func corpusPinManifest(root string) (CorpusPin, error) {
	var manifest CorpusPin
	encoded, err := os.ReadFile(filepath.Join(root, "stage1/cohere/json/testdata/corpus.pin"))
	if err != nil {
		return manifest, err
	}
	return manifest, json.Unmarshal(encoded, &manifest)
}

func provisionedCorpusFiles(root string) ([]ComparisonCase, error) {
	manifest, err := corpusPinManifest(root)
	if err != nil {
		return nil, err
	}
	var cases []ComparisonCase
	for _, item := range manifest.Cases {
		if !strings.HasPrefix(item.Path, ProvisionedCorpusPrefix) {
			continue
		}
		var data []byte
		buildcache.Keying(func() { data, err = os.ReadFile(filepath.Join(root, filepath.FromSlash(item.Path))) })
		if err != nil {
			return nil, fmt.Errorf("pinned corpus input %s (npm ci --prefix stage3/api provisions it on the corpus product's builder): %w", item.Path, err)
		}
		if digest := fmt.Sprintf("%x", sha256.Sum256(data)); digest != item.SHA256 {
			return nil, fmt.Errorf("pinned corpus input %s: SHA256 %s, corpus.pin names %s", item.Path, digest, item.SHA256)
		}
		if !utf8.Valid(data) {
			return nil, fmt.Errorf("non-UTF-8 pinned corpus file %s", item.Path)
		}
		cases = append(cases, ComparisonCase{Name: item.Path, Text: string(data)})
	}
	return cases, nil
}

// RepositoryCasesAtPin is shared with the parent json tests.
func RepositoryCasesAtPin(root string, inputs buildcache.Inputs, pin string) ([]ComparisonCase, error) {
	files, err := buildcache.PinnedFiles(root, inputs, pin, "JSON", func(name string) bool { return strings.HasSuffix(name, ".json") })
	if err != nil {
		return nil, err
	}
	cases := make([]ComparisonCase, 0, len(files))
	for _, file := range files {
		if !utf8.Valid(file.Data) {
			return nil, fmt.Errorf("non-UTF-8 corpus file %s at pin %s", file.Path, pin)
		}
		cases = append(cases, ComparisonCase{Name: file.Path, Text: string(file.Data)})
	}
	return cases, nil
}

func pinnedCorpusFiles(t testing.TB, root string) ([]ComparisonCase, error) {
	tracked, provisioned := PinnedCorpusProduct(t)
	cases := append(tracked, provisioned...)
	manifest, err := corpusPinManifest(root)
	if err != nil {
		return nil, err
	}
	// Only pinned submodule paths are read from disk. Arbitrary repository JSON,
	// additions, deletions and dirty index entries cannot enter.
	for _, item := range manifest.Cases {
		if strings.HasPrefix(item.Path, "generated/") || strings.HasPrefix(item.Path, ProvisionedCorpusPrefix) {
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
		cases = append(cases, ComparisonCase{Name: item.Path, Text: string(data)})
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

// RepositoryCorpusIdentityError is shared with the parent json tests.
func RepositoryCorpusIdentityError(pin string, cases []ComparisonCase, count int, identity string) error {
	got := PinForCases(cases)
	if got.Count != count || got.SHA256 != identity {
		return fmt.Errorf("repository JSON pin %s: corpus identity changed: %d cases SHA256 %s; want %d SHA256 %s (bump pin and identity together)", pin, got.Count, got.SHA256, count, identity)
	}
	return nil
}

// ValidatePinnedCorpus is shared with the parent json tests.
func ValidatePinnedCorpus(t testing.TB, cases []ComparisonCase, expected CorpusPin) (CorpusPin, int, error) {
	tracked, _ := PinnedCorpusProduct(t)
	state := repositoryJSON{paths: map[string]bool{}}
	for _, item := range tracked {
		state.paths[item.Name] = true
	}
	return ValidateCorpusState(state.paths, state.dirty, cases, expected)
}

var jsonTrackedSources struct {
	once  sync.Once
	files []string
	err   error
}

var pinnedCorpusRead struct {
	once                 sync.Once
	tracked, provisioned []ComparisonCase
	err                  error
}

const RepositoryCorpusCommit = "2b4dbde172c1c2f23c065ecae8b906f8fcb72a20"

const ProvisionedCorpusPrefix = "stage3/api/node_modules/"

type repositoryJSON struct {
	paths map[string]bool
	dirty []string
}

// ValidateCorpusState is shared with the parent json tests.
func ValidateCorpusState(paths map[string]bool, dirty []string, cases []ComparisonCase, expected CorpusPin) (CorpusPin, int, error) {
	state := repositoryJSON{paths: paths, dirty: dirty}
	if len(state.dirty) > 0 {
		return CorpusPin{}, 0, fmt.Errorf("repository dirty JSON: %s", strings.Join(state.dirty, ", "))
	}
	allowed := map[string]bool{}
	for _, item := range expected.Cases {
		allowed[item.Path] = true
	}
	seen, tracked := map[string]bool{}, map[string]bool{}
	var pinned []ComparisonCase
	for _, item := range cases {
		if seen[item.Name] {
			return CorpusPin{}, len(tracked), fmt.Errorf("duplicate corpus path: %s", item.Name)
		}
		seen[item.Name] = true
		if state.paths[item.Name] {
			tracked[item.Name] = true
			continue
		}
		group := CaseGroup(item.Name)
		if group == "cohere" || group == "TypeScript" || allowed[item.Name] {
			pinned = append(pinned, item)
			continue
		}
		return CorpusPin{}, len(tracked), fmt.Errorf("unexpected untracked JSON: %s", item.Name)
	}
	var missing []string
	for path := range state.paths {
		if !tracked[path] {
			missing = append(missing, path)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return CorpusPin{}, len(tracked), fmt.Errorf("repository missing tracked JSON from walk: %s", strings.Join(missing, ", "))
	}
	pin := PinForCases(pinned)
	return pin, len(tracked), CorpusPinError(expected, pin)
}

// CorpusCases is shared with the parent json tests.
func CorpusCases(t *testing.T) []ComparisonCase {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	cases, err := pinnedCorpusFiles(t, root)
	if err != nil {
		t.Fatal(err)
	}
	// Generated boundary cases exercise both filename-selected printers, including the extension's
	// JavaScript literal forms that a JSON.parse/JSON.stringify substitute would lose.
	generated := []string{
		"[]", "{}", "null", "true", "false", `{"a":1,"b":[2,3]}`,
		strings.Repeat("[", 500) + "0" + strings.Repeat("]", 500),
		"[" + strings.Repeat("1,", 4999) + "1]",
		`["\u0000","\uD800","\uD83D\uDE00","é😀","\\","\""]`,
		`[0,-0,1,-1,1.50,1E+03,1e-003,1e309,1e-400,9007199254740993]`,
		`[+1,.5,1.,0x1f,0o17,0b101,1_000,Infinity,NaN,undefined]`,
		"// leading\n{a:1,/* inside */b:[1,2,],}",
		`[1__0]`, `[1_]`, `[0x_1]`, `{"a":`, ``,
	}
	for index, text := range generated {
		for _, name := range []string{"probe.json", "package.json"} {
			cases = append(cases, ComparisonCase{fmt.Sprintf("generated/%d/%s", index, name), text})
		}
	}
	t.Logf("full corpus: %d cases before gate sampling", len(cases))
	verifyCorpusPin(t, cases)
	return cases
}
