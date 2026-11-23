package lane

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// A public-any exemption remains justified only while its pinned API sample still reads the unchecked value.
// This test runs git to read the pinned object, never Go or a compiler. Entries may be removed, never added here.
func TestJSONValues52PinnedAnySampleReadsRemain(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "stage3/adapt/52-json-values/public-any-pinned.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Path     []string `json:"path"`
		Kind     string   `json:"kind"`
		Baseline string   `json:"baseline"`
		Read     string   `json:"read"`
	}
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) > 1 {
		t.Fatal("the reviewed public-any exemption list grew")
	}
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		t.Fatal("ADAMIC_TYPESCRIPT_SOURCE must name the pinned upstream git input")
	}
	sample, err := pinnedSample(source, typescriptPin(t, root), "tests/cases/compiler/APISample_parseConfig.ts")
	if err != nil {
		t.Fatalf("read pinned API sample: %v", err)
	}
	seen := map[string]bool{}
	for _, row := range rows {
		if len(row.Path) != 2 || row.Path[1] != "config" ||
			row.Path[0] != "parseConfigFileTextToJson" ||
			row.Kind != "PropertySignature" || row.Baseline != "APISample_parseConfig.errors.txt" ||
			row.Read != `config.config["compilerOptions"]` {
			t.Fatalf("unreviewed public-any exemption: %+v", row)
		}
		name := strings.Join(row.Path, ".")
		if seen[name] {
			t.Fatalf("duplicate exemption %s", name)
		}
		seen[name] = true
		if !jsonValues52SampleContainsOwnPin(string(sample), row.Path[0], row.Read) {
			t.Errorf("%s is no longer pinned by %s: remove its exemption; missing own call ts.%s( or read %q", name, row.Baseline, row.Path[0], row.Read)
		}
	}
}

// pinnedSample is path's bytes at pin in source, read by git as PinnedGit reads a pinned commit
// (internal/buildcache/pinned.go:33): replace objects and grafts off and no transport, so the answer is the pin's own.
// toolchain's GitEnv already keeps git's configuration out of reach. A replace ref or a graft in the source could
// otherwise answer with another commit's bytes under the pin's id, and the exemption would be judged on them. It is
// written here, not shared with PinnedGit, which takes a product's Inputs: a test that keys no product has none.
func pinnedSample(source, pin, path string) ([]byte, error) {
	command := toolchain.GitCommandIn(source, "-c", "protocol.allow=never", "show", pin+":"+path)
	if command.Err != nil {
		return nil, command.Err
	}
	command.Env = append(command.Env, "GIT_NO_REPLACE_OBJECTS=1", "GIT_GRAFT_FILE="+os.DevNull, "GIT_NO_LAZY_FETCH=1")
	return command.Output()
}

// A replace ref in the source never changes what the pin's sample reads: pinnedSample reads the pin's own blob.
func TestJSONValues52PinnedSampleIgnoresAReplaceRef(t *testing.T) {
	t.Parallel()
	source := t.TempDir()
	git := func(arguments ...string) string {
		t.Helper()
		command := toolchain.GitCommandIn(source, append([]string{"-c", "user.name=test", "-c", "user.email=test@adamic.invalid", "-c", "commit.gpgSign=false"}, arguments...)...)
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("git %s: %v: %s", strings.Join(arguments, " "), err, output)
		}
		return strings.TrimSpace(string(output))
	}
	git("init", "-q")
	write := func(text string) string {
		if err := os.WriteFile(filepath.Join(source, "sample.ts"), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		git("add", "sample.ts")
		git("commit", "-q", "-m", text)
		return git("rev-parse", "HEAD")
	}
	pin := write("pinned\n")
	other := write("replaced\n")
	git("replace", pin, other)
	if plain := git("show", pin+":sample.ts"); plain != "replaced" {
		t.Fatalf("the planted replace ref doesn't answer for the pin (git show read %q), so this test can't fail", plain)
	}
	sample, err := pinnedSample(source, pin, "sample.ts")
	if err != nil {
		t.Fatal(err)
	}
	if string(sample) != "pinned\n" {
		t.Errorf("pinnedSample read %q at the pin, not its own %q", sample, "pinned\n")
	}
}

func jsonValues52SampleContainsOwnPin(sample, function, read string) bool {
	return strings.Contains(sample, "ts."+function+"(") && strings.Contains(sample, read)
}

func TestJSONValues52ExemptionsRequireTheirOwnCall(t *testing.T) {
	t.Parallel()
	const read = `config.config["compilerOptions"]`
	const sample = `const config = ts.parseConfigFileTextToJson("config.json", text); config.config["compilerOptions"];`
	if !jsonValues52SampleContainsOwnPin(sample, "parseConfigFileTextToJson", read) {
		t.Fatal("the parser call and its read must pin the parser")
	}
	if jsonValues52SampleContainsOwnPin(sample, "readConfigFile", read) {
		t.Fatal("another function's call and read cannot pin readConfigFile")
	}
	if jsonValues52SampleContainsOwnPin(read, "parseConfigFileTextToJson", read) {
		t.Fatal("the read alone cannot pin a function")
	}
	if jsonValues52SampleContainsOwnPin(`ts.parseConfigFileTextToJson("config.json", text)`, "parseConfigFileTextToJson", read) {
		t.Fatal("the call alone cannot pin the unchecked read")
	}
}
