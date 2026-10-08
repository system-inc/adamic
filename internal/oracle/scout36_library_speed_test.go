package oracle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

type librarySpeedCase struct {
	Name   string `json:"name"`
	From   string `json:"mutant_from"`
	To     string `json:"mutant_to"`
	Source string `json:"source"`
	Lowers bool   `json:"lowers"`
}

// The measurement runner and this oracle consume the same manifest and programs.
func librarySpeedCases() []librarySpeedCase {
	data, err := os.ReadFile(filepath.Join(repository, "bench/library-speed/cases.json"))
	if err != nil {
		panic(err)
	}
	var cases []librarySpeedCase
	if err := json.Unmarshal(data, &cases); err != nil {
		panic(err)
	}
	return cases
}

func init() {
	for _, one := range librarySpeedCases() {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"bench/library-speed/" + one.Name + ".a", one.Lowers, false})
	}
}

// Change one operation, keep the types and ownership valid, and run with sanitizers.
// A successful compile/run cannot detect these wrong results; the original source on Node does.
func TestScout36LibrarySpeedMutants(t *testing.T) {
	for _, one := range librarySpeedCases() {
		if !one.Lowers {
			continue
		}
		t.Run(one.Name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "bench/library-speed", one.Name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(source), one.From) != 1 {
				t.Fatalf("mutant must change one operation: %q", one.From)
			}
			mutant := filepath.Join(t.TempDir(), "mutant.a")
			if err := os.WriteFile(mutant, []byte(strings.Replace(string(source), one.From, one.To, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, mutant)
			if err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			got := executeInput(t, inputRun{directory: filepath.Dir(path)}, nodeFSDirectorySanitizerEnvironment(), binary)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("mutant failed outside Node comparison: exit=%d stderr=%s", got.exitCode, got.stderr)
			}
			if difference := disagreement(onNode(t, path), got); difference != "stdout differs" {
				t.Fatalf("caught by %q, want stdout differs", difference)
			}
			t.Log("valid operation mutant caught only by output comparison with original source on Node")
		})
	}
}
