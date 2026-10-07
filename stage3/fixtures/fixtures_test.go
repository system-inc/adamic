package fixtures

import (
	"bytes"
	"context"
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
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
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
}
type fixture struct {
	File     string   `json:"file"`
	Platform string   `json:"platform,omitempty"`
	TSC      []string `json:"tsc"`
	Reason   string   `json:"reason"`
	Node     behavior `json:"node"`
	Stage0   stage0   `json:"stage0"`
}

func execute(t *testing.T, directory string, environment []string, name string, arguments ...string) behavior {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, name, arguments...)
	command.Dir = directory
	command.Env = append(os.Environ(), environment...)
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

// The enum and namespace branches use Node's transform mode. Derive that runner
// from the current source oracle, preserving its runtime and import hooks while
// leaving the ordinary erasable runner untouched.
func transformedNodeRunner(t *testing.T, repository string) string {
	t.Helper()
	path := filepath.Join(repository, "oracle/node.mjs")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	if strings.Count(text, "stripTypeScriptTypes(source)") != 1 || strings.Count(text, "new URL('./adamic.mjs', import.meta.url)") != 1 {
		t.Fatal("source Node runner changed: review the transform-mode hook")
	}
	text = strings.Replace(text, "stripTypeScriptTypes(source)", "stripTypeScriptTypes(source, { mode: 'transform' })", 1)
	oracleURL := (&url.URL{Scheme: "file", Path: filepath.ToSlash(path)}).String()
	text = strings.Replace(text, "new URL('./adamic.mjs', import.meta.url)", fmt.Sprintf("new URL('./adamic.mjs', %q)", oracleURL), 1)
	transformed := filepath.Join(t.TempDir(), "node-transform.mjs")
	if err := os.WriteFile(transformed, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	return transformed
}

func TestFixtures(t *testing.T) {
	t.Parallel()
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs(*fixtureRoot)
	if err != nil {
		t.Fatal(err)
	}
	statuses, err := filepath.Glob(filepath.Join(root, "*", "status.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(statuses) == 0 {
		t.Fatal("no fixture status.json files found")
	}
	// These helpers live in oracle's test files. Build their small exported test
	// hook once, then run one process per compiling fixture in parallel.
	hook := filepath.Join(t.TempDir(), "oracle.test")
	built := execute(t, repository, nil, "go", "test", "-c", "-o", hook, "./internal/oracle")
	if built.Exit != 0 {
		t.Fatalf("building oracle hook: %s%s", built.Stdout, built.Stderr)
	}
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
				if len(entry.TSC) == 0 || entry.Reason == "" {
					t.Fatalf("missing provenance for %s", entry.File)
				}
				switch entry.Stage0.Outcome {
				case "Checker", "Refused", "NotYet", "Compiles":
				default:
					t.Fatalf("invalid outcome: %q", entry.Stage0.Outcome)
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
					if err == nil {
						_, err = lower.Lower(context.Background(), program)
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
							result := execute(t, repository, []string{"ADAMIC_STAGE3_FIXTURE=" + path, "ADAMIC_STAGE3_RESULT=" + resultPath}, hook, "-test.run=^TestStage3FixtureHook$", "-test.count=1")
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
							equal(t, "native versus Node", behavior{string(stdout), string(stderr), exit}, node)
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
