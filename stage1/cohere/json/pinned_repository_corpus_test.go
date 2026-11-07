package json

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"

	"github.com/system-inc/adamic/stage1/cohere/json/oracles"
)

// Freeze repository JSON so selection need not rerun this package on every
// .json edit. These inputs move only when someone explicitly bumps this pin.
// The checkout that builds the corpus product must already hold this commit (a
// shallow one fetches it at checkout, git fetch --depth=1 origin <pin>); the
// corpus never fetches, and a missing pin is fatal.
const repositoryCorpusCommit = oracles.RepositoryCorpusCommit

// SHA256 of sorted name<TAB>SHA256(bytes)<LF>, including upstream inputs and
// generated controls: 3,919 cases at cohere 0183ccf8, where 7945d102 had 3,900.
const repositoryCorpusIdentity = "0ebc69862856a18e2af48e2a27ebfba868f642da3b4e48c218b99bd0a09236bf"

// The corpus files corpus.pin names under stage3/api/node_modules: npm ci --prefix stage3/api installs them from
// stage3/api's lockfile, so no archive of tracked files carries them.
const provisionedCorpusPrefix = oracles.ProvisionedCorpusPrefix

// The corpus product holds every input a runner's source archive lacks: the pin's repository JSON (Git objects,
// and Loom's source archive is the tree's tracked files with no history) and the provisioned node_modules files
// corpus.pin names (untracked). Workshop builds it from a checkout that holds both, and the tests read only it, the
// source archive and the cohere submodule's tracked files. Git objects are content addressed, so the pin alone
// names the repository bytes; the lockfile names the provisioned ones, and corpus.pin checks both byte for byte.
//
// The repository's tracked JSON is declared too. The corpus carries the pin's JSON text, whose reports name the
// machines they ran on (/tmp/adamic-gate/..., /private/tmp/...), and buildcache's relocation check passes such a path
// only when a tracked file the product declares holds it (#arw7837). Those reports are tracked and unchanged since the
// pin, so the product's key also moves when one of them does.
func jsonPinnedCorpusInputs() (buildcache.Inputs, error) { return oracles.PinnedCorpusInputs() }

// jsonPinnedCorpusInputsAt is the corpus product's key with the repository JSON at pin: the whole commit id is a flag,
// so a different pin, which any change to a corpus file at it makes, is a different key.
func jsonPinnedCorpusInputsAt(pin string) (buildcache.Inputs, error) {
	return oracles.PinnedCorpusInputsAt(pin)
}

// A different pin moves both the name key (what a runner finds the corpus by) and the declared key (what this machine
// alone builds it by), so a corpus at another commit is never served for this one. A one-byte change to a corpus file
// at the pin is a different commit id, so it is this case too. Keying builds nothing.
func TestJSONPinnedCorpusRekeysWithThePin(test *testing.T) {
	test.Parallel()
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	keys := func(pin string) (string, string) {
		inputs, jsonPinnedCorpusInputsAtError := jsonPinnedCorpusInputsAt(pin)
		if jsonPinnedCorpusInputsAtError != nil {
			test.Fatal(jsonPinnedCorpusInputsAtError)
		}
		declared, keyError := buildcache.Key(root, inputs)
		if keyError != nil {
			test.Fatal(keyError)
		}
		return buildcache.NameKey(root, inputs), declared
	}
	nameKey, declared := keys(repositoryCorpusCommit)
	// The pin with its last hex digit changed: the smallest move to another commit id.
	other := repositoryCorpusCommit[:39] + "1"
	if other == repositoryCorpusCommit {
		other = repositoryCorpusCommit[:39] + "2"
	}
	otherNameKey, otherDeclared := keys(other)
	if otherNameKey == nameKey || otherDeclared == declared {
		test.Fatalf("pin %s kept name key %s or key %s", other, nameKey[:12], declared[:12])
	}
	test.Logf(
		"pin %s to %s: name key %s to %s, key %s to %s",
		repositoryCorpusCommit,
		other,
		nameKey[:12],
		otherNameKey[:12],
		declared[:12],
		otherDeclared[:12],
	)
}

// jsonTrackedSources is the repository's tracked JSON, listed once per process.

// buildJSONPinnedCorpus is the build for the product inputs keys, and PinnedGit checks the pin against those same
// Inputs: the build never lists them again, so it never lists the tracked JSON, which reads .git, inside its trace.
func buildJSONPinnedCorpus(inputs buildcache.Inputs) func(directory string) error {
	return oracles.BuildPinnedCorpus(inputs)
}

// The product test and every corpus reader share one recipe and key.
func TestProduct_JSONPinnedCorpus(test *testing.T) {
	test.Parallel()
	inputs, jsonPinnedCorpusInputsError := jsonPinnedCorpusInputs()
	if jsonPinnedCorpusInputsError != nil {
		test.Fatal(jsonPinnedCorpusInputsError)
	}
	buildcache.Product(test, inputs, buildJSONPinnedCorpus(inputs))
}

// Every test in a process reads the product once. Callers get slice copies because hash partitioning sorts its
// input in place.

func pinnedCorpusProduct(test testing.TB) (tracked, provisioned []textCase) {
	trackedCases, provisionedCases := oracles.PinnedCorpusProduct(test)
	return localCases(trackedCases), localCases(provisionedCases)
}

// A product file holds its cases the way Git's cat-file --batch does: "<size> <name>\n", the bytes, "\n".
func writeCorpusRecords(path string, cases []textCase) error {
	return oracles.WriteCorpusRecords(path, oracleCases(cases))
}

func readCorpusRecords(path string) ([]textCase, error) {
	cases, readCorpusRecordsError := oracles.ReadCorpusRecords(path)
	return localCases(cases), readCorpusRecordsError
}

// The corpus.pin manifest beside this file, read from the repository root.

// provisionedCorpusFiles reads the node_modules files corpus.pin names from the product builder's disk, where npm ci
// --prefix stage3/api put them. They're untracked, so a traced build could never key them; the reads run as key-time
// work instead (buildcache.Keying), and that is sound only because each file must match the SHA256 corpus.pin gives it,
// byte for byte: corpus.pin is a declared input, so any bytes accepted here are exactly what the key already names.

// repositoryCasesAtPin reads the pin's JSON from the checkout's own objects, on the corpus product's builder, through
// buildcache.PinnedFiles: as key-time work, since inputs names the pin, so a traced build never reads .git (#jwn3n9q).
// It runs git with no transport and no lazy fetch, so the corpus never fetches. The TypeScript corpusfiles selector
// returns live paths, not pinned bytes.
func repositoryCasesAtPin(root string, inputs buildcache.Inputs, pin string) ([]textCase, error) {
	cases, repositoryCasesAtPinError := oracles.RepositoryCasesAtPin(root, inputs, pin)
	return localCases(cases), repositoryCasesAtPinError
}

// pinnedCorpusFiles is every non-generated input: the pin's repository JSON and the provisioned files from the
// corpus product, and the pinned submodule paths from the checkout's own tracked files.

func repositoryCorpusIdentityError(pin string, cases []textCase, count int, identity string) error {
	return oracles.RepositoryCorpusIdentityError(pin, oracleCases(cases), count, identity)
}

func TestPinnedRepositoryCorpusIdentity(test *testing.T) {
	test.Parallel()
	cases := corpusCases(test)
	if err := repositoryCorpusIdentityError(repositoryCorpusCommit, cases, 3919, repositoryCorpusIdentity); err != nil {
		test.Fatal(err)
	}
	test.Logf("3919 cases; name/byte identity SHA256 %s", pinForCases(cases).SHA256)
}

// fixtureCasesAtPin is repositoryCasesAtPin on a fixture repository, under a key that names the fixture's pin, as the
// product's names its own.
func fixtureCasesAtPin(root, pin string) ([]textCase, error) {
	return repositoryCasesAtPin(
		root,
		buildcache.Inputs{Name: "json pinned corpus fixture", Flags: []string{"repository JSON at " + pin}},
		pin,
	)
}

func TestPinnedRepositoryCorpusIgnoresWorkingTree(test *testing.T) {
	test.Parallel()
	fixture := newRepositoryFixture(test)
	pin, gitCorpusError := gitCorpus(fixture.root, "rev-parse", "HEAD")
	if gitCorpusError != nil {
		test.Fatal(gitCorpusError)
	}
	before, fixtureCasesAtPinError := fixtureCasesAtPin(fixture.root, strings.TrimSpace(string(pin)))
	if fixtureCasesAtPinError != nil {
		test.Fatal(fixtureCasesAtPinError)
	}
	fixtureWrite(test, fixture.root, "tracked.json", "{changed:true}")
	fixtureWrite(test, fixture.root, "added.json", "[42]")
	fixtureGit(test, fixture.root, "add", "tracked.json", "added.json")
	if removeError := os.Remove(filepath.Join(fixture.root, "nested/second.json")); removeError != nil {
		test.Fatal(removeError)
	}
	after, fixtureCasesAtPinError := fixtureCasesAtPin(fixture.root, strings.TrimSpace(string(pin)))
	if fixtureCasesAtPinError != nil {
		test.Fatal(fixtureCasesAtPinError)
	}
	if fmt.Sprint(before) != fmt.Sprint(after) {
		test.Fatal("working-tree/index mutation changed pinned names or bytes")
	}
	test.Log("staged edit/addition and worktree deletion left pinned names and bytes unchanged")
}

func TestPinnedRepositoryCorpusPinMutant(test *testing.T) {
	test.Parallel()
	fixture := newRepositoryFixture(test)
	pin, gitCorpusError := gitCorpus(fixture.root, "rev-parse", "HEAD")
	if gitCorpusError != nil {
		test.Fatal(gitCorpusError)
	}
	before, fixtureCasesAtPinError := fixtureCasesAtPin(fixture.root, strings.TrimSpace(string(pin)))
	if fixtureCasesAtPinError != nil {
		test.Fatal(fixtureCasesAtPinError)
	}
	fixtureWrite(test, fixture.root, "pin-added.json", "[42]")
	commitFixtureJSON(test, fixture.root, "pin-added.json")
	mutant, gitCorpusError := gitCorpus(fixture.root, "rev-parse", "HEAD")
	if gitCorpusError != nil {
		test.Fatal(gitCorpusError)
	}
	after, fixtureCasesAtPinError := fixtureCasesAtPin(fixture.root, strings.TrimSpace(string(mutant)))
	if fixtureCasesAtPinError != nil {
		test.Fatal(fixtureCasesAtPinError)
	}
	if pinForCases(before).SHA256 == pinForCases(after).SHA256 || len(after) != len(before)+1 {
		test.Fatal("pin mutant did not change its JSON set and identity")
	}
	want := pinForCases(before)
	mutantPin := strings.TrimSpace(string(mutant))
	if repositoryCorpusIdentityError(mutantPin, before, want.Count, want.SHA256) != nil {
		test.Fatal("baseline identity failed")
	}
	if repositoryCorpusIdentityError(mutantPin, after, want.Count, want.SHA256) == nil {
		test.Fatal("pin mutant survived the identity check")
	}
	test.Logf(
		"pin mutant %s rejected: %v",
		strings.TrimSpace(string(mutant)),
		repositoryCorpusIdentityError(mutantPin, after, want.Count, want.SHA256),
	)
}

func validatePinnedCorpus(test testing.TB, cases []textCase, expected corpusPin) (corpusPin, int, error) {
	return oracles.ValidatePinnedCorpus(test, oracleCases(cases), expected)
}

func TestJSONPinnedCorpusRecords(test *testing.T) {
	test.Parallel()
	cases := []textCase{
		{"a.json", "{}"},
		{"with space/b.json", ""},
		{"c.json", "[1,\n2]\n\n"},
		{"d.json", "7 e.json\n"},
	}
	path := filepath.Join(test.TempDir(), "records")
	if writeCorpusRecordsError := writeCorpusRecords(path, cases); writeCorpusRecordsError != nil {
		test.Fatal(writeCorpusRecordsError)
	}
	read, readCorpusRecordsError := readCorpusRecords(path)
	if readCorpusRecordsError != nil || fmt.Sprint(read) != fmt.Sprint(cases) {
		test.Fatalf("records changed names or bytes: %v %q", readCorpusRecordsError, read)
	}
	encoded, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	// A product cut short, or a record whose size is off by one, is refused rather than read as fewer or other cases.
	for _, mutant := range []struct{ name, text string }{
		{"truncated", string(encoded[:len(encoded)-1])},
		{"size", strings.Replace(string(encoded), "2 a.json", "1 a.json", 1)},
	} {
		if writeError := os.WriteFile(path, []byte(mutant.text), 0o644); writeError != nil {
			test.Fatal(writeError)
		}
		if _, readCorpusRecordsError := readCorpusRecords(
			path,
		); readCorpusRecordsError == nil || !strings.Contains(readCorpusRecordsError.Error(), "malformed record") {
			test.Fatalf("%s product read without error: %v", mutant.name, readCorpusRecordsError)
		}
	}
	if writeCorpusRecordsError := writeCorpusRecords(
		path,
		[]textCase{{"new\nline.json", "{}"}},
	); writeCorpusRecordsError == nil {
		test.Fatal("a name with a newline was written")
	}
}

// Not parallel: the process environment (GIT_TRACE2_EVENT, set by test.Setenv to trace git's fetches).
func TestPinnedRepositoryCorpusNeverFetches(test *testing.T) {
	fixture := newRepositoryFixture(test)
	head, err := gitCorpus(fixture.root, "rev-parse", "HEAD")
	if err != nil {
		test.Fatal(err)
	}
	pin := strings.TrimSpace(string(head))
	full, err := fixtureCasesAtPin(fixture.root, pin)
	if err != nil {
		test.Fatal(err)
	}
	// A shallow gate checkout without the pin, whose origin holds it and would serve it to any fetch: a read that
	// fetched would succeed here.
	shallow := test.TempDir()
	fixtureGit(test, shallow, "init", "-q")
	fixtureGit(test, shallow, "remote", "add", "origin", fixture.root)
	var missing error
	if fetches := gitFetchesTracing(test, func() { _, missing = fixtureCasesAtPin(shallow, pin) }); fetches != 0 {
		test.Fatalf("a checkout without the pin fetched %d times, want none", fetches)
	}
	if missing == nil ||
		!strings.Contains(missing.Error(), pin) ||
		!strings.Contains(missing.Error(), "not in the checkout") {
		test.Fatalf("a checkout without the pin read the corpus anyway, or lost the pin's name: %v", missing)
	}
	// The checkout step carries the pin, as a runner's must; the read then fetches nothing of its own.
	fixtureGit(test, shallow, "fetch", "--quiet", "--depth=1", "origin", pin)
	var after []textCase
	if fetches := gitFetchesTracing(
		test,
		func() { after, err = fixtureCasesAtPin(shallow, pin) },
	); fetches != 0 || err != nil {
		test.Fatalf("a shallow checkout carrying the pin: %d fetches, %v", fetches, err)
	}
	if fmt.Sprint(full) != fmt.Sprint(after) {
		test.Fatal("a shallow checkout carrying the pin read different names or bytes")
	}
	test.Logf("without the pin: %v; with it at depth 1: %d identical inputs, no fetch", missing, len(after))
}

// gitFetchesTracing counts the git processes a function starts whose arguments include fetch, from Git's trace2
// event stream.
func gitFetchesTracing(test *testing.T, run func()) int {
	test.Helper()
	trace := filepath.Join(test.TempDir(), "trace2.json")
	test.Setenv("GIT_TRACE2_EVENT", trace)
	run()
	encoded, readError := os.ReadFile(trace)
	if readError != nil {
		test.Fatal(readError)
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

// Not parallel: the process environment (GIT_TRACE2_EVENT, set by test.Setenv to trace git's fetches).
func TestPinnedRepositoryCorpusBloblessCloneNeverFetches(test *testing.T) {
	fixture := newRepositoryFixture(test)
	for index := 0; index < 8; index++ {
		fixtureWrite(test, fixture.root, fmt.Sprintf("batch/%d.json", index), fmt.Sprintf("[%d]", index))
	}
	commitFixtureJSON(test, fixture.root, "batch")
	head, err := gitCorpus(fixture.root, "rev-parse", "HEAD")
	if err != nil {
		test.Fatal(err)
	}
	pin := strings.TrimSpace(string(head))
	fixtureGit(test, fixture.root, "config", "uploadpack.allowFilter", "true")
	fixtureGit(test, fixture.root, "config", "uploadpack.allowAnySHA1InWant", "true")
	// The pool's checkout: every commit and tree, no blobs, and an origin that would serve each one.
	clone := filepath.Join(test.TempDir(), "blobless")
	fixtureGit(
		test,
		fixture.root,
		"clone",
		"--quiet",
		"--filter=blob:none",
		"--no-checkout",
		"file://"+fixture.root,
		clone,
	)
	var missing error
	if fetches := gitFetchesTracing(test, func() { _, missing = fixtureCasesAtPin(clone, pin) }); fetches != 0 {
		test.Fatalf("a blobless clone fetched %d times, want none", fetches)
	}
	// A missing input is the checkout's, named, and never reads as a wrong answer (devtools/red-sort's WRONG).
	wrong := regexp.MustCompile(
		`(?i)compile error|build.failed|undefined:|syntax error|assertion|\b(?:got|want|expected)\b|assert.*diff`,
	)
	if missing == nil ||
		!strings.Contains(missing.Error(), pin) ||
		!strings.Contains(missing.Error(), "10 of its 10 JSON files") ||
		!strings.Contains(missing.Error(), "batch/0.json") ||
		wrong.MatchString(missing.Error()) {
		test.Fatalf("a blobless clone read the pinned corpus anyway, or didn't name what it lacks: %v", missing)
	}
	// A full clone has every blob and fetches nothing.
	if fetches := gitFetchesTracing(
		test,
		func() { _, err = fixtureCasesAtPin(fixture.root, pin) },
	); fetches != 0 || err != nil {
		test.Fatalf("a full clone: %d fetches, %v", fetches, err)
	}
	test.Log(missing)
}
