package lint

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegexCompiledOptionCapture(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, slug := range []string{"id-length", "no-inline-comments", "no-warning-comments"} {
		directory := filepath.Join(root, "rules", slug)
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
		source := filepath.Join(packageDirectory, "rules", slug)
		err := filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			relative, err := filepath.Rel(source, path)
			if err != nil {
				return err
			}
			target := filepath.Join(directory, relative)
			if entry.IsDir() {
				return os.MkdirAll(target, 0755)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			return os.WriteFile(target, data, 0644)
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	directory := t.TempDir()
	rows, _, err := captureUpstream(root, directory)
	if err != nil {
		t.Fatal(err)
	}
	sources := 0
	for _, row := range rows {
		fields := strings.Split(row, "\t")
		if fields[1] != "id-length" || fields[5] == "" {
			continue
		}
		var options struct{ ExceptionPatterns []json.RawMessage }
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			t.Fatal(err)
		}
		for _, pattern := range options.ExceptionPatterns {
			var source string
			if err := json.Unmarshal(pattern, &source); err != nil {
				t.Fatalf("compiled option source lost: %s", pattern)
			}
			sources++
		}
	}
	if sources == 0 {
		t.Fatal("capture exercised no compiled exception patterns")
	}
	t.Logf("%d upstream cases replayed through the Go oracle; %d compiled pattern sources preserved", len(rows), sources)
}

func TestTypedCompiledOptionCaptureKey(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	cohere, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		t.Fatal(err)
	}
	base := filepath.Join(cohere, "internal/lint/testing")
	record := typedConfigRecord
	if os.Getenv("ADAMIC_TYPED_CAPTURE_MUTANT") == "1" {
		record = strings.Replace(record, "marshalCapturedOptions(options)", "json.Marshal(options)", 1)
	}
	// Execute the exact injected typed record with a real compiled matcher, then
	// look up its program using the source-preserving docs capture's case key.
	source := `package rule_testing
import (
 "encoding/json"
 "os"
 "path/filepath"
 "testing"
 esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
)
type adamicCaptureSource string
func (s adamicCaptureSource) Text() string {return string(s)}
func TestAdamicTypedCompiledOptions(t *testing.T) {
 directory := t.TempDir()
 if err := os.WriteFile(filepath.Join(directory,"tsconfig.json"),[]byte("{\"compilerOptions\":{\"strict\":false}}"),0644);err != nil {t.Fatal(err)}
 pattern,err := esregexp.Compile("^_","u")
 if err != nil {t.Fatal(err)}
 var options any = struct {ExceptionPatterns []*esregexp.RegExp}{[]*esregexp.RegExp{pattern}}
 subject := struct{Name string}{"typed-pattern-control"}
 subjectFileName := "source.ts"
 sourceFile := adamicCaptureSource("let _a = 1;")
 diagnostics := []int{1}
 files := []int{1}
` + record + "}\n"
	side := filepath.Join(directory, "typed_options_test.go")
	if err := os.WriteFile(side, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{
		filepath.Join(base, "adamic_typed_options_test.go"): side,
		filepath.Join(base, "adamic_capture_options.go"):    filepath.Join(packageDirectory, "regex/testdata/capture_options.go"),
	}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := run(cohere, []string{"COHERE_DOCS_CAPTURE=" + directory}, "go", "test", "-overlay="+overlayPath, "./internal/lint/testing", "-run", "^TestAdamicTypedCompiledOptions$", "-count=1"); err != nil {
		t.Fatal(err)
	}
	configs, err := readTypedConfigs(filepath.Join(directory, typedConfigFile))
	if err != nil {
		t.Fatal(err)
	}
	key := "typed-pattern-control\tsource.ts\t{\"ExceptionPatterns\":[\"^_\"]}\tlet _a = 1;"
	programs := configs[key]
	if len(programs) != 1 || programs[0].CompilerOptions != `{"strict":false}` || programs[0].Findings != 1 {
		t.Fatalf("compiled option lost typed program: source-preserving key %q, captured programs %v", key, configs)
	}
	t.Log("compiled-pattern typed config retains the strict:false program under the docs capture's identical source-preserving key")
}
