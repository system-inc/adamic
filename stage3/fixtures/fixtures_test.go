package fixtures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/toolchain"
)

var update = flag.Bool("update", false, "update stage0 only after native agrees with recorded and current Node")
var fixtureRoot = flag.String("fixtures", ".", "fixture root, including scratch copies for audits")

type behavior struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	Exit   int    `json:"exit"`
}
type stage0 struct {
	Outcome string `json:"outcome"`
	What    string `json:"what"`
	Stdout  string `json:"stdout,omitempty"`
}
type fixture struct {
	File     string   `json:"file"`
	Platform string   `json:"platform,omitempty"`
	TSC      []string `json:"tsc"`
	Reason   string   `json:"reason"`
	Node     behavior `json:"node"`
	Stage0   stage0   `json:"stage0"`
}

// toolchainPath is the PATH setting toolchain.Env gives every tool.
func toolchainPath() string {
	return toolchainSettings("PATH")[0]
}

// toolchainSettings are the settings toolchain.Env gives every tool for names, in its spelling: its PATH, and the C
// compiler go builds cgo with (CC=clang and CXX=clang++ unless the caller names others).
func toolchainSettings(names ...string) []string {
	var settings []string
	for _, name := range names {
		// The last setting of a name is the one a child sees, as exec keeps the last of duplicates: toolchain.Env passes
		// a caller's empty CC and then appends CC=clang.
		last := ""
		for _, setting := range toolchain.Env() {
			if strings.HasPrefix(setting, name+"=") {
				last = setting
			}
		}
		if last == "" {
			panic("toolchain.Env sets no " + name)
		}
		settings = append(settings, last)
	}
	return settings
}

func execute(t testing.TB, directory string, environment []string, name string, arguments ...string) behavior {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	// A bare name is the program toolchain.Path finds; one it can't find fails as it starts, with its error.
	path, unresolved := toolchain.Path(name)
	if unresolved != nil {
		path = name
	}
	command := exec.CommandContext(ctx, path, arguments...)
	if unresolved != nil {
		command.Err = unresolved
	}
	command.Dir = directory
	// The PATH toolchain gives a tool, the declared tools directory's places first, so what a child runs by name itself
	// (the hook scripts' go) is toolchain's too, and its CC and CXX, so the hook's cgo builds and the go env its key
	// reads both name the compiler every other build uses (clang, not Linux's gcc). environment, last, still wins. The
	// rest of this process's environment passes on purpose: the hooks read their ADAMIC_STAGE3_ settings from it, so a
	// node fixture run also takes the caller's NODE_OPTIONS. python3 goes through toolchain.Path, and no box's setup
	// puts it under ADAMIC_TOOLS, so a sandbox fails the python3 hooks, naming it, rather than run PATH's.
	command.Env = append(append(os.Environ(), toolchainSettings("PATH", "CC", "CXX")...), environment...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
	command.WaitDelay = 5 * time.Second
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	var exited *exec.ExitError
	if ctx.Err() != nil {
		t.Fatalf("%s timed out: %v", name, ctx.Err())
	}
	if err != nil && !errors.As(err, &exited) {
		t.Fatal(err)
	}
	return behavior{stdout.String(), stderr.String(), command.ProcessState.ExitCode()}
}

func equal(t *testing.T, check string, got, want behavior) {
	t.Helper()
	if got != want {
		t.Errorf("%s byte comparison failed: got exit=%d stdout=%q stderr=%q; want exit=%d stdout=%q stderr=%q", check, got.Exit, got.Stdout, got.Stderr, want.Exit, want.Stdout, want.Stderr)
	}
}

// replaceStage0 replaces only this JSON value. Node, provenance, extra fields,
// key order and whitespace outside stage0 remain byte-for-byte unchanged.
func replaceStage0(raw json.RawMessage, value stage0) (json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return nil, fmt.Errorf("fixture must be an object: %v", err)
	}
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		var field json.RawMessage
		if err := decoder.Decode(&field); err != nil {
			return nil, err
		}
		if key == "stage0" {
			encoded, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}
			end := int(decoder.InputOffset())
			start := end - len(field)
			result := append([]byte(nil), raw[:start]...)
			result = append(result, encoded...)
			return append(result, raw[end:]...), nil
		}
	}
	return nil, errors.New("fixture has no stage0 field")
}

// Fixture paths use portable slash-separated components. ValidPath rejects
// absolute paths, empty components, and both . and .. components before joining.
func validFixturePath(name string) bool {
	return fs.ValidPath(name) && !strings.ContainsAny(name, "\\:") && filepath.Ext(name) == ".a"
}

// Check the supported mode independently of the broader call and URL guards.
// Its fatal must be observable even when another guard rejects the same runner.
func nodeRunnerMode(t *testing.T, text string) int {
	t.Helper()
	erasable := strings.Count(text, "stripTypeScriptTypes(source)")
	transformedCalls := strings.Count(text, "stripTypeScriptTypes(source, { mode: 'transform' })")
	if erasable+transformedCalls != 1 {
		t.Fatal("source Node runner changed: review the transform-mode hook")
	}
	return erasable
}

// The enum and namespace branches use Node's transform mode. Derive that runner
// from the current source oracle, preserving its runtime and import hooks while
// accepting either source mode without changing the original runner.
func transformedNodeRunner(t *testing.T, repository string) string {
	t.Helper()
	path := filepath.Join(repository, "oracle/node.mjs")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	// Parameter properties may already require transform mode in the source oracle.
	erasable := nodeRunnerMode(t, text)
	if strings.Count(text, "stripTypeScriptTypes(source") != 1 || strings.Count(text, "new URL('./adamic.mjs', import.meta.url)") != 1 {
		t.Fatal("source Node runner changed: review the transform-mode hook")
	}
	if erasable == 1 {
		text = strings.Replace(text, "stripTypeScriptTypes(source)", "stripTypeScriptTypes(source, { mode: 'transform' })", 1)
	}
	oracleURL := (&url.URL{Scheme: "file", Path: filepath.ToSlash(path)}).String()
	text = strings.Replace(text, "new URL('./adamic.mjs', import.meta.url)", fmt.Sprintf("new URL('./adamic.mjs', %q)", oracleURL), 1)
	transformed := filepath.Join(t.TempDir(), "node-transform.mjs")
	if err := os.WriteFile(transformed, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	return transformed
}

func TestFixturesAssertions(t *testing.T) {
	t.Parallel()
	testFixtureDirectory(t, "assertions")
}
func TestFixturesCycles(t *testing.T) {
	t.Parallel()
	testFixtureDirectory(t, "cycles")
}
func TestFixturesEntries(t *testing.T) {
	t.Parallel()
	testFixtureDirectory(t, "entries")
}
func TestFixturesEnums(t *testing.T) {
	t.Parallel()
	testFixtureDirectory(t, "enums")
}
func TestFixturesHost(t *testing.T) {
	t.Parallel()
	testFixtureDirectory(t, "host")
}
func TestFixturesNamespaces(t *testing.T) {
	t.Parallel()
	testFixtureDirectory(t, "namespaces")
}
func TestFixturesNestedFunctions(t *testing.T) {
	t.Parallel()
	testFixtureDirectory(t, "nested-functions")
}
func TestFixturesObjects(t *testing.T) {
	t.Parallel()
	testFixtureDirectory(t, "objects")
}
func TestFixturesPredicates(t *testing.T) {
	t.Parallel()
	testFixtureDirectory(t, "predicates")
}
func TestFixturesReal(t *testing.T) {
	t.Parallel()
	testFixtureDirectory(t, "real")
}
func TestFixturesRecords(t *testing.T) {
	t.Parallel()
	testFixtureDirectory(t, "records")
}
func TestFixturesRunner(t *testing.T) {
	t.Parallel()
	testFixtureDirectory(t, "runner")
}
func TestFixturesTaste(t *testing.T) {
	t.Parallel()
	testFixtureDirectory(t, "taste")
}

func testFixtureDirectory(t *testing.T, directory string, selectedFiles ...string) {
	t.Helper()
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs(*fixtureRoot)
	if err != nil {
		t.Fatal(err)
	}
	statuses := []string{filepath.Join(root, directory, "status.json")}
	hook := fixtureOracleHook(t, repository)
	shard, count := fixtureShard(t)
	transformedRunner := transformedNodeRunner(t, repository)
	for _, status := range statuses {
		t.Run(filepath.Base(filepath.Dir(status)), func(t *testing.T) {
			t.Parallel()
			data, err := os.ReadFile(status)
			if err != nil {
				t.Fatal(err)
			}
			var entries []json.RawMessage
			if err := json.Unmarshal(data, &entries); err != nil {
				t.Fatal(err)
			}
			if len(entries) == 0 {
				t.Fatal("empty fixture status")
			}
			var mutex sync.Mutex
			replacements := make(map[int]json.RawMessage)
			t.Cleanup(func() {
				if len(replacements) == 0 {
					return
				}
				var result bytes.Buffer
				cursor := 0
				for index, raw := range entries {
					offset := bytes.Index(data[cursor:], raw)
					if offset < 0 {
						t.Fatal("cannot locate original fixture JSON")
					}
					start := cursor + offset
					result.Write(data[cursor:start])
					if replacement, ok := replacements[index]; ok {
						result.Write(replacement)
					} else {
						result.Write(raw)
					}
					cursor = start + len(raw)
				}
				result.Write(data[cursor:])
				current, err := os.ReadFile(status)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(current, data) {
					t.Fatal("status.json changed during the test; refusing to overwrite it")
				}
				temporary, err := os.CreateTemp(filepath.Dir(status), ".status-*")
				if err != nil {
					t.Fatal(err)
				}
				defer os.Remove(temporary.Name())
				info, err := os.Stat(status)
				if err == nil {
					err = temporary.Chmod(info.Mode().Perm())
				}
				if err == nil {
					_, err = temporary.Write(result.Bytes())
				}
				closeError := temporary.Close()
				if err != nil {
					t.Fatal(err)
				}
				if closeError != nil {
					t.Fatal(closeError)
				}
				if err := os.Rename(temporary.Name(), status); err != nil {
					t.Fatal(err)
				}
			})
			seen := map[string]bool{}
			for index, raw := range entries {
				var entry fixture
				decoder := json.NewDecoder(bytes.NewReader(raw))
				if err := decoder.Decode(&entry); err != nil {
					t.Fatal(err)
				}
				if err := decoder.Decode(new(any)); err != io.EOF {
					t.Fatal("unexpected trailing fixture JSON")
				}
				if !validFixturePath(entry.File) || seen[entry.File] {
					t.Fatalf("invalid or duplicate file: %q", entry.File)
				}
				seen[entry.File] = true
				if len(selectedFiles) != 0 {
					selected := false
					for _, file := range selectedFiles {
						selected = selected || file == entry.File
					}
					if !selected {
						continue
					}
				}
				if len(entry.TSC) == 0 || entry.Reason == "" {
					t.Fatalf("missing provenance for %s", entry.File)
				}
				switch entry.Stage0.Outcome {
				case "Checker", "Refused", "NotYet", "Compiles", "CheckedStop":
				default:
					t.Fatalf("invalid outcome: %q", entry.Stage0.Outcome)
				}
				name := filepath.Base(filepath.Dir(status)) + "/" + entry.File
				selected := fixtureShardIndex(name, count) == shard
				if !selected {
					continue
				}
				t.Run(entry.File, func(t *testing.T) {
					t.Parallel()
					if entry.Platform != "" && entry.Platform != runtime.GOOS {
						t.Skipf("fixture records platform %s; current platform is %s", entry.Platform, runtime.GOOS)
					}
					path := filepath.Join(filepath.Dir(status), entry.File)
					nodeRunner := filepath.Join(repository, "oracle/node.mjs")
					bucket := filepath.Base(filepath.Dir(status))
					if bucket == "enums" || bucket == "namespaces" {
						nodeRunner = transformedRunner
					}
					node := execute(t, repository, nil, "node", "--disable-warning=ExperimentalWarning", nodeRunner, path)
					recordedNodeAgrees := t.Run("node", func(t *testing.T) { equal(t, "recorded Node", node, entry.Node) })
					program, err := load.Load([]string{path})
					actual := stage0{Outcome: "Compiles"}
					var checkedJavaScript string
					if err == nil {
						lowered, lowerError := lower.Lower(context.Background(), program)
						err = lowerError
						if err == nil && entry.Stage0.Outcome == "CheckedStop" {
							checkedJavaScript = javascript.JavaScript(lowered)
						}
					}
					if err != nil {
						var checker *load.CheckError
						var refused *lower.Refused
						var notYet *lower.NotYet
						switch {
						case errors.As(err, &checker):
							actual.Outcome = "Checker"
						case errors.As(err, &refused):
							actual.Outcome = "Refused"
						case errors.As(err, &notYet):
							actual.Outcome = "NotYet"
						default:
							t.Fatalf("unexpected stage 0 error: %v", err)
						}
						// Exact diagnostic text with repository-relative paths for portability.
						actual.What = strings.ReplaceAll(err.Error(), root+string(filepath.Separator), "stage3/fixtures/")
						actual.What = strings.ReplaceAll(actual.What, repository+string(filepath.Separator), "")
					}
					nativeAgrees := false
					if actual.Outcome == "Compiles" {
						nativeAgrees = t.Run("native", func(t *testing.T) {
							resultPath := filepath.Join(t.TempDir(), "result.json")
							result := execute(t, repository, []string{"ADAMIC_STAGE3_FIXTURE=" + path, "ADAMIC_STAGE3_RESULT=" + resultPath}, hook, "-test.run=^TestStage3FixtureHook$", "-test.count=1", "-test.timeout=85s")
							if result.Exit != 0 {
								t.Fatalf("oracle native/sanitizer/leak hook failed: %s%s", result.Stdout, result.Stderr)
							}
							data, err := os.ReadFile(resultPath)
							if err != nil {
								t.Fatal(err)
							}
							var fields []json.RawMessage
							if err := json.Unmarshal(data, &fields); err != nil {
								t.Fatal(err)
							}
							if len(fields) != 3 {
								t.Fatal("invalid oracle hook result")
							}
							var stdout, stderr []byte
							var exit int
							for index, target := range []any{&stdout, &stderr, &exit} {
								if err := json.Unmarshal(fields[index], target); err != nil {
									t.Fatal(err)
								}
							}
							observed := behavior{string(stdout), string(stderr), exit}
							if entry.Stage0.Outcome == "CheckedStop" {
								// An inserted check may stop source that Node accepts. Pin the
								// complete stop, including output before it, on both backends.
								checkedPath := filepath.Join(t.TempDir(), "checked.mjs")
								if err := os.WriteFile(checkedPath, []byte(checkedJavaScript), 0600); err != nil {
									t.Fatal(err)
								}
								checkedNode := execute(t, repository, nil, "node", "--disable-warning=ExperimentalWarning", nodeRunner, checkedPath)
								want := behavior{Stdout: entry.Stage0.Stdout, Stderr: entry.Stage0.What, Exit: 70}
								equal(t, "pinned Node check", checkedNode, want)
								equal(t, "native versus checked Node", observed, checkedNode)
								equal(t, "pinned native check", observed, want)
								if observed == want {
									actual = stage0{Outcome: "CheckedStop", What: string(stderr), Stdout: string(stdout)}
								}
							} else {
								equal(t, "native versus Node", observed, node)
							}
						})
					}
					if *update && recordedNodeAgrees && nativeAgrees {
						if actual != entry.Stage0 {
							replacement, err := replaceStage0(raw, actual)
							if err != nil {
								t.Fatal(err)
							}
							mutex.Lock()
							replacements[index] = replacement
							mutex.Unlock()
						}
						return
					}
					t.Run("stage0", func(t *testing.T) {
						if actual != entry.Stage0 {
							t.Errorf("gap changed: update status.json and check the native output; got %+v; recorded %+v", actual, entry.Stage0)
						}
					})
				})
			}
		})
	}
}

func TestFixturePaths(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name  string
		valid bool
	}{
		{"01_value.a", true}, {"01_call_time/main.a", true},
		{"01_call_time/_namespaces/main.a", true},
		{"../main.a", false}, {"01_case/../main.a", false},
		{"/tmp/main.a", false}, {"C:/main.a", false}, {`01_case\main.a`, false},
		{"./main.a", false}, {"01_case//main.a", false}, {"main.ts", false}, {"", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := validFixturePath(test.name); got != test.valid {
				t.Errorf("validFixturePath(%q) = %v, want %v", test.name, got, test.valid)
			}
		})
	}
}

// Indices are zero based, as in the gate. Invalid selectors must never turn an
// intended run into a successful empty run.
func fixtureShard(t *testing.T) (int, int) {
	t.Helper()
	text := os.Getenv("ADAMIC_TEST_SHARD")
	if text == "" {
		return 0, 1
	}
	parts := strings.Split(text, "/")
	if len(parts) != 2 {
		t.Fatalf("invalid ADAMIC_TEST_SHARD %q: want i/n", text)
	}
	index, indexError := strconv.Atoi(parts[0])
	count, countError := strconv.Atoi(parts[1])
	if indexError != nil || countError != nil || count < 1 || index < 0 || index >= count {
		t.Fatalf("invalid ADAMIC_TEST_SHARD %q: require 0 <= i < n", text)
	}
	return index, count
}

func TestFixtureShardManifest(t *testing.T) {
	t.Parallel()
	statuses, err := filepath.Glob(filepath.Join(*fixtureRoot, "*", "status.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(statuses) == 0 {
		t.Fatal("no fixture status.json files found")
	}
	_, count := fixtureShard(t)
	seen := map[string]bool{}
	for _, status := range statuses {
		data, err := os.ReadFile(status)
		if err != nil {
			t.Fatal(err)
		}
		var entries []fixture
		if err := json.Unmarshal(data, &entries); err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			name := filepath.Base(filepath.Dir(status)) + "/" + entry.File
			if seen[name] || !validFixturePath(entry.File) {
				t.Fatalf("invalid or duplicate fixture %q", name)
			}
			seen[name] = true
			t.Logf("%s shard=%d/%d", name, fixtureShardIndex(name, count), count)
		}
	}
	if len(seen) == 0 {
		t.Fatal("empty fixture shard manifest")
	}
}

// Assignment depends only on the full fixture identity, so -run selection and
// scheduling cannot move a fixture into a different shard.
func fixtureShardIndex(name string, count int) int {
	digest := sha256.Sum256([]byte(name))
	return int(binary.BigEndian.Uint64(digest[:8]) % uint64(count))
}

// A child execute runs, as the oracle hook's scripts are, builds cgo with the C compiler toolchain gives every tool:
// with no CC or CXX of the caller's, clang and clang++, not the gcc Linux's go would default to. The child here is a
// shell printing both. Mutant: execute without toolchainSettings' CC and CXX, and the child sees neither.
// Not parallel: clears CC and CXX through t.Setenv.
func TestExecuteGivesChildrenTheToolchainsCompiler(t *testing.T) {
	t.Setenv("CC", "")
	t.Setenv("CXX", "")
	result := execute(t, t.TempDir(), nil, "/bin/sh", "-c", `echo "$CC $CXX"`)
	if result.Exit != 0 || result.Stdout != "clang clang++\n" {
		t.Fatalf("a child saw %q (exit %d, stderr %q), want toolchain's clang and clang++", result.Stdout, result.Exit, result.Stderr)
	}
}
