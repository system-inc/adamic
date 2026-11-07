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

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/hostconfiguration"
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
	Configuration *hostconfiguration.Configuration `json:"configuration,omitempty"`
	File          string                           `json:"file"`
	Platform      string                           `json:"platform,omitempty"`
	TSC           []string                         `json:"tsc"`
	Reason        string                           `json:"reason"`
	Node          behavior                         `json:"node"`
	Stage0        stage0                           `json:"stage0"`
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
		// The last setting of a name is the one a child sees, as exec keeps the last of duplicates: toolchain.Env
		// passes a caller's empty CC and then appends CC=clang.
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

func execute(test testing.TB, directory string, environment []string, name string, arguments ...string) behavior {
	test.Helper()
	timeoutContext, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	// A bare name is the program toolchain.Path finds; one it can't find fails as it starts, with its error.
	path, unresolved := toolchain.Path(name)
	if unresolved != nil {
		path = name
	}
	command := exec.CommandContext(timeoutContext, path, arguments...)
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
	stdout, stderr, captureError := childguard.Capture(timeoutContext, command)
	var exited *exec.ExitError
	if timeoutContext.Err() != nil {
		test.Fatalf("%s timed out: %v", name, timeoutContext.Err())
	}
	if captureError != nil && !errors.As(captureError, &exited) {
		test.Fatal(captureError)
	}
	return behavior{string(stdout), string(stderr), command.ProcessState.ExitCode()}
}

func equal(test *testing.T, check string, got, want behavior) {
	test.Helper()
	if got != want {
		test.Errorf(
			"%s byte comparison failed: got exit=%d stdout=%q stderr=%q; want exit=%d stdout=%q stderr=%q",
			check,
			got.Exit,
			got.Stdout,
			got.Stderr,
			want.Exit,
			want.Stdout,
			want.Stderr,
		)
	}
}

// replaceStage0 replaces only this JSON value. Node, provenance, extra fields,
// key order and whitespace outside stage0 remain byte-for-byte unchanged.
func replaceStage0(raw json.RawMessage, value stage0) (json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, tokenError := decoder.Token()
	if tokenError != nil || token != json.Delim('{') {
		return nil, fmt.Errorf("fixture must be an object: %v", tokenError)
	}
	for decoder.More() {
		key, tokenError := decoder.Token()
		if tokenError != nil {
			return nil, tokenError
		}
		var field json.RawMessage
		if decodeError := decoder.Decode(&field); decodeError != nil {
			return nil, decodeError
		}
		if key == "stage0" {
			encoded, encodeError := json.Marshal(value)
			if encodeError != nil {
				return nil, encodeError
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
	return fs.ValidPath(name) &&
		!strings.ContainsAny(name, "\\:") &&
		(filepath.Ext(name) == ".a" || filepath.Ext(name) == ".ts")
}

// Check the supported mode independently of the broader call and URL guards.
// Its fatal must be observable even when another guard rejects the same runner.
func nodeRunnerMode(test *testing.T, text string) int {
	test.Helper()
	erasable := strings.Count(text, "stripTypeScriptTypes(source)")
	// The oracle's transform call also asks for a source map, so Node's stack positions name the program's own lines
	// (#f4qxx5d). A map changes no code: on all 1,637 programs under stage3/fixtures, internal/oracle/testdata and
	// oracle, the stripped output parses to the same syntax with and without it, differing only in comments and the map
	// line.
	transformedCalls := strings.Count(text, "stripTypeScriptTypes(source, { mode: 'transform' })") +
		strings.Count(text, "stripTypeScriptTypes(source, { mode: 'transform', sourceMap: true, sourceUrl: url })")
	if erasable+transformedCalls != 1 {
		test.Fatal("source Node runner changed: review the transform-mode hook")
	}
	return erasable
}

// The enum and namespace branches use Node's transform mode. Derive that runner
// from the current source oracle, preserving its runtime and import hooks while
// accepting either source mode without changing the original runner.
func transformedNodeRunner(test *testing.T, repository string) string {
	test.Helper()
	path := filepath.Join(repository, "oracle/node.mjs")
	source, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	text := string(source)
	// Parameter properties may already require transform mode in the source oracle.
	erasable := nodeRunnerMode(test, text)
	if strings.Count(text, "stripTypeScriptTypes(source") != 1 ||
		strings.Count(text, "new URL('./adamic.mjs', import.meta.url)") != 1 {
		test.Fatal("source Node runner changed: review the transform-mode hook")
	}
	if erasable == 1 {
		text = strings.Replace(
			text,
			"stripTypeScriptTypes(source)",
			"stripTypeScriptTypes(source, { mode: 'transform' })",
			1,
		)
	}
	oracleURL := (&url.URL{Scheme: "file", Path: filepath.ToSlash(path)}).String()
	text = strings.Replace(
		text,
		"new URL('./adamic.mjs', import.meta.url)",
		fmt.Sprintf("new URL('./adamic.mjs', %q)", oracleURL),
		1,
	)
	transformed := filepath.Join(test.TempDir(), "node-transform.mjs")
	if writeError := os.WriteFile(transformed, []byte(text), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	return transformed
}

func TestFixturesAssertions(test *testing.T) {
	test.Parallel()
	testFixtureDirectory(test, "assertions")
}
func TestFixturesCycles(test *testing.T) {
	test.Parallel()
	testFixtureDirectory(test, "cycles")
}
func TestFixturesEntries(test *testing.T) {
	test.Parallel()
	testFixtureDirectory(test, "entries")
}
func TestFixturesEnums(test *testing.T) {
	test.Parallel()
	testFixtureDirectory(test, "enums")
}
func TestFixturesHost(test *testing.T) {
	test.Parallel()
	testFixtureDirectory(test, "host")
}
func TestFixturesNamespaces(test *testing.T) {
	test.Parallel()
	testFixtureDirectory(test, "namespaces")
}
func TestFixturesNestedFunctions(test *testing.T) {
	test.Parallel()
	testFixtureDirectory(test, "nested-functions")
}
func TestFixturesObjects(test *testing.T) {
	test.Parallel()
	testFixtureDirectory(test, "objects")
}
func TestFixturesPredicates(test *testing.T) {
	test.Parallel()
	testFixtureDirectory(test, "predicates")
}
func TestFixturesReal(test *testing.T) {
	test.Parallel()
	testFixtureDirectory(test, "real")
}
func TestFixturesRecords(test *testing.T) {
	test.Parallel()
	testFixtureDirectory(test, "records")
}
func TestFixturesRunner(test *testing.T) {
	test.Parallel()
	testFixtureDirectory(test, "runner")
}
func TestFixturesTaste(test *testing.T) {
	test.Parallel()
	testFixtureDirectory(test, "taste")
}

func testFixtureDirectory(test *testing.T, directory string, selectedFiles ...string) {
	test.Helper()
	repository, absolutePathError := filepath.Abs("../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	root, absolutePathError := filepath.Abs(*fixtureRoot)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	statuses := []string{filepath.Join(root, directory, "status.json")}
	hook := fixtureOracleHook(test, repository)
	shard, count := fixtureShard(test)
	transformedRunner := transformedNodeRunner(test, repository)
	for _, status := range statuses {
		test.Run(filepath.Base(filepath.Dir(status)), func(test *testing.T) {
			test.Parallel()
			data, readError := os.ReadFile(status)
			if readError != nil {
				test.Fatal(readError)
			}
			var entries []json.RawMessage
			if decodeError := json.Unmarshal(data, &entries); decodeError != nil {
				test.Fatal(decodeError)
			}
			if len(entries) == 0 {
				test.Fatal("empty fixture status")
			}
			var mutex sync.Mutex
			replacements := make(map[int]json.RawMessage)
			test.Cleanup(func() {
				if len(replacements) == 0 {
					return
				}
				var result bytes.Buffer
				cursor := 0
				for index, raw := range entries {
					offset := bytes.Index(data[cursor:], raw)
					if offset < 0 {
						test.Fatal("cannot locate original fixture JSON")
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
					test.Fatal(err)
				}
				if !bytes.Equal(current, data) {
					test.Fatal("status.json changed during the test; refusing to overwrite it")
				}
				temporary, err := os.CreateTemp(filepath.Dir(status), ".status-*")
				if err != nil {
					test.Fatal(err)
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
					test.Fatal(err)
				}
				if closeError != nil {
					test.Fatal(closeError)
				}
				if renameError := os.Rename(temporary.Name(), status); renameError != nil {
					test.Fatal(renameError)
				}
			})
			seen := map[string]bool{}
			for index, raw := range entries {
				var entry fixture
				decoder := json.NewDecoder(bytes.NewReader(raw))
				if decodeError := decoder.Decode(&entry); decodeError != nil {
					test.Fatal(decodeError)
				}
				if decodeError := decoder.Decode(new(any)); decodeError != io.EOF {
					test.Fatal("unexpected trailing fixture JSON")
				}
				if !validFixturePath(entry.File) || seen[entry.File] {
					test.Fatalf("invalid or duplicate file: %q", entry.File)
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
					test.Fatalf("missing provenance for %s", entry.File)
				}
				switch entry.Stage0.Outcome {
				case "Checker", "Refused", "NotYet", "Compiles", "CheckedStop":
				default:
					test.Fatalf("invalid outcome: %q", entry.Stage0.Outcome)
				}
				name := filepath.Base(filepath.Dir(status)) + "/" + entry.File
				selected := fixtureShardIndex(name, count) == shard
				if !selected {
					continue
				}
				test.Run(entry.File, func(test *testing.T) {
					test.Parallel()
					if entry.Platform != "" && entry.Platform != runtime.GOOS {
						test.Skipf("fixture records platform %s; current platform is %s", entry.Platform, runtime.GOOS)
					}
					path := filepath.Join(filepath.Dir(status), entry.File)
					nodeRunner := filepath.Join(repository, "oracle/node.mjs")
					bucket := filepath.Base(filepath.Dir(status))
					if bucket == "enums" || bucket == "namespaces" {
						nodeRunner = transformedRunner
					}
					node := entry.Node // Backend checks never depend on an available reference binary.
					nodeChecked := false
					recordedNodeAgrees := test.Run("node", func(test *testing.T) {
						if why := hostconfiguration.Skip(index+1, entry.Configuration); why != "" {
							test.Skipf(
								"fixture row %d: the node step needs node-without-inspector; set ADAMIC_HOST_NODE_WITHOUT_INSPECTOR",
								index+1,
							)
						}
						nodeBinary := hostconfiguration.Require(test, index+1, entry.Configuration)
						actualNode := execute(
							test,
							repository,
							nil,
							nodeBinary,
							"--disable-warning=ExperimentalWarning",
							nodeRunner,
							path,
						)
						equal(test, "recorded Node", actualNode, entry.Node)
						nodeChecked = true
					}) && nodeChecked
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
							test.Fatalf("unexpected stage 0 error: %v", err)
						}
						// Exact diagnostic text with repository-relative paths for portability.
						actual.What = strings.ReplaceAll(
							err.Error(),
							root+string(filepath.Separator),
							"stage3/fixtures/",
						)
						actual.What = strings.ReplaceAll(actual.What, repository+string(filepath.Separator), "")
					}
					nativeAgrees := false
					if actual.Outcome == "Compiles" {
						nativeAgrees = test.Run("native", func(test *testing.T) {
							resultPath := filepath.Join(test.TempDir(), "result.json")
							result := execute(
								test,
								repository,
								[]string{"ADAMIC_STAGE3_FIXTURE=" + path, "ADAMIC_STAGE3_RESULT=" + resultPath},
								hook,
								"-test.run=^TestStage3FixtureHook$",
								"-test.count=1",
								"-test.timeout=85s",
							)
							if result.Exit != 0 {
								test.Fatalf(
									"oracle native/sanitizer/leak hook failed: %s%s",
									result.Stdout,
									result.Stderr,
								)
							}
							data, readError := os.ReadFile(resultPath)
							if readError != nil {
								test.Fatal(readError)
							}
							var fields []json.RawMessage
							if decodeError := json.Unmarshal(data, &fields); decodeError != nil {
								test.Fatal(decodeError)
							}
							if len(fields) != 3 {
								test.Fatal("invalid oracle hook result")
							}
							var stdout, stderr []byte
							var exit int
							for index, target := range []any{&stdout, &stderr, &exit} {
								if decodeError := json.Unmarshal(fields[index], target); decodeError != nil {
									test.Fatal(decodeError)
								}
							}
							observed := behavior{string(stdout), string(stderr), exit}
							if entry.Stage0.Outcome == "CheckedStop" {
								// An inserted check may stop source that Node accepts. Pin the
								// complete stop, including output before it, on both backends.
								checkedPath := filepath.Join(test.TempDir(), "checked.mjs")
								if writeError := os.WriteFile(
									checkedPath,
									[]byte(checkedJavaScript),
									0600,
								); writeError != nil {
									test.Fatal(writeError)
								}
								want := behavior{Stdout: entry.Stage0.Stdout, Stderr: entry.Stage0.What, Exit: 70}
								equal(test, "pinned native check", observed, want)
								if observed == want {
									actual = stage0{
										Outcome: "CheckedStop",
										What:    string(stderr),
										Stdout:  string(stdout),
									}
								}
								test.Run("javascript", func(test *testing.T) {
									if why := hostconfiguration.Skip(index+1, entry.Configuration); why != "" {
										test.Skipf(
											"fixture row %d: the CheckedStop javascript check needs node-without-inspector; set ADAMIC_HOST_NODE_WITHOUT_INSPECTOR",
											index+1,
										)
									}
									nodeBinary := hostconfiguration.Require(test, index+1, entry.Configuration)
									checkedNode := execute(
										test,
										repository,
										nil,
										nodeBinary,
										"--disable-warning=ExperimentalWarning",
										nodeRunner,
										checkedPath,
									)
									equal(test, "pinned Node check", checkedNode, want)
									equal(test, "native versus checked Node", observed, checkedNode)
								})
							} else {
								equal(test, "native versus Node", observed, node)
							}
						})
					}
					if *update && recordedNodeAgrees && nativeAgrees {
						if actual != entry.Stage0 {
							replacement, replaceStage0Error := replaceStage0(raw, actual)
							if replaceStage0Error != nil {
								test.Fatal(replaceStage0Error)
							}
							mutex.Lock()
							replacements[index] = replacement
							mutex.Unlock()
						}
						return
					}
					test.Run("stage0", func(test *testing.T) {
						if actual != entry.Stage0 {
							test.Errorf(
								"gap changed: update status.json and check the native output; got %+v; recorded %+v",
								actual,
								entry.Stage0,
							)
						}
					})
				})
			}
		})
	}
}

func TestFixturePaths(test *testing.T) {
	test.Parallel()
	for _, testCase := range []struct {
		name  string
		valid bool
	}{
		{"01_value.a", true},
		{"01_call_time/main.a", true},
		{"01_call_time/_namespaces/main.a", true},
		{"../main.a", false},
		{"01_case/../main.a", false},
		{"/tmp/main.a", false},
		{"C:/main.a", false},
		{`01_case\main.a`, false},
		{"./main.a", false},
		{"01_case//main.a", false},
		{"main.ts", true},
		{"main.js", false},
		{"main", false},
		{"", false},
	} {
		test.Run(testCase.name, func(test *testing.T) {
			test.Parallel()
			if got := validFixturePath(testCase.name); got != testCase.valid {
				test.Errorf("validFixturePath(%q) = %v, want %v", testCase.name, got, testCase.valid)
			}
		})
	}
}

// Indices are zero based, as in the gate. Invalid selectors must never turn an
// intended run into a successful empty run.
func fixtureShard(test *testing.T) (int, int) {
	test.Helper()
	text := os.Getenv("ADAMIC_TEST_SHARD")
	if text == "" {
		return 0, 1
	}
	parts := strings.Split(text, "/")
	if len(parts) != 2 {
		test.Fatalf("invalid ADAMIC_TEST_SHARD %q: want i/n", text)
	}
	index, indexError := strconv.Atoi(parts[0])
	count, countError := strconv.Atoi(parts[1])
	if indexError != nil || countError != nil || count < 1 || index < 0 || index >= count {
		test.Fatalf("invalid ADAMIC_TEST_SHARD %q: require 0 <= i < n", text)
	}
	return index, count
}

func TestFixtureShardManifest(test *testing.T) {
	test.Parallel()
	statuses, globError := filepath.Glob(filepath.Join(*fixtureRoot, "*", "status.json"))
	if globError != nil {
		test.Fatal(globError)
	}
	if len(statuses) == 0 {
		test.Fatal("no fixture status.json files found")
	}
	_, count := fixtureShard(test)
	seen := map[string]bool{}
	for _, status := range statuses {
		data, readError := os.ReadFile(status)
		if readError != nil {
			test.Fatal(readError)
		}
		var entries []fixture
		if decodeError := json.Unmarshal(data, &entries); decodeError != nil {
			test.Fatal(decodeError)
		}
		for index, entry := range entries {
			_, skip, resolveError := hostconfiguration.Resolve(index+1, entry.Configuration, os.Getenv)
			if resolveError != nil {
				test.Fatal(resolveError)
			}
			if skip != "" {
				// A manifest must keep every row: log the requirement; never Skip it.
				test.Log(skip)
			}
			name := filepath.Base(filepath.Dir(status)) + "/" + entry.File
			if seen[name] || !validFixturePath(entry.File) {
				test.Fatalf("invalid or duplicate fixture %q", name)
			}
			seen[name] = true
			test.Logf("%s shard=%d/%d", name, fixtureShardIndex(name, count), count)
		}
	}
	if len(seen) == 0 {
		test.Fatal("empty fixture shard manifest")
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
// Not parallel: clears CC and CXX through test.Setenv.
func TestExecuteGivesChildrenTheToolchainsCompiler(test *testing.T) {
	test.Setenv("CC", "")
	test.Setenv("CXX", "")
	result := execute(test, test.TempDir(), nil, "/bin/sh", "-c", `echo "$CC $CXX"`)
	if result.Exit != 0 || result.Stdout != "clang clang++\n" {
		test.Fatalf(
			"a child saw %q (exit %d, stderr %q), want toolchain's clang and clang++",
			result.Stdout,
			result.Exit,
			result.Stderr,
		)
	}
}

// Both streams exceed a pipe buffer; the final bytes and nonzero exit must survive the drain.
// Mutant: drop the last captured stdout byte before returning from execute.
func TestFixtureExecuteCapturesWholeOutput(test *testing.T) {
	test.Parallel()
	const size = 1 << 20
	result := execute(test, test.TempDir(), nil, "node", "-e", `
const fs = require('node:fs');
fs.writeFileSync(1, 'o'.repeat(1 << 20));
fs.writeFileSync(2, 'e'.repeat(1 << 20));
setTimeout(() => process.exit(7), 25);
`)
	if result.Exit != 7 || result.Stdout != strings.Repeat("o", size) || result.Stderr != strings.Repeat("e", size) {
		test.Fatalf(
			"captured exit=%d stdout=%d bytes stderr=%d bytes, want exit=7 and %d bytes per stream",
			result.Exit,
			len(result.Stdout),
			len(result.Stderr),
			size,
		)
	}
}
