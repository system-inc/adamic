package estree

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// Oracle answers are immutable products of the oracle binary and the complete,
// ordered input bytes. Relative paths keep manifests and audit records identical
// on different instances. Audit answer files are written only during the build.
func estreeOracleOutput(t *testing.T, oracle string, arguments ...string) []byte {
	t.Helper()
	mode := "single"
	paths := []string{arguments[0]}
	if strings.HasPrefix(arguments[0], "--") {
		mode = arguments[0]
		listing, err := os.ReadFile(arguments[1])
		if err != nil {
			t.Fatal(err)
		}
		paths = nil
		for _, path := range strings.Split(string(listing), "\n") {
			if path != "" {
				paths = append(paths, path)
			}
		}
	}
	if len(paths) == 0 {
		t.Fatal("empty oracle input enumeration")
	}
	binary, err := os.ReadFile(oracle)
	if err != nil {
		t.Fatal(err)
	}
	inputs := buildcache.Inputs{Name: "estree oracle answers", Files: []string{"stage1/cohere/estree/oracle_answers_test.go"}, Flags: []string{mode, fmt.Sprintf("oracle SHA256 %x", sha256.Sum256(binary))}}
	sources := make([][]byte, len(paths))
	names := make([]string, len(paths))
	for i, path := range paths {
		sources[i], err = os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		names[i] = fmt.Sprintf("%04d%s", i, filepath.Ext(path))
		inputs.Flags = append(inputs.Flags, fmt.Sprintf("%s SHA256 %x", names[i], sha256.Sum256(sources[i])))
	}
	directory := buildcache.Product(t, inputs, func(directory string) error {
		for i, name := range names {
			if err := os.WriteFile(filepath.Join(directory, name), sources[i], 0644); err != nil {
				return err
			}
		}
		args := []string{names[0]}
		if mode != "single" {
			if err := os.WriteFile(filepath.Join(directory, "manifest"), []byte(strings.Join(names, "\n")+"\n"), 0644); err != nil {
				return err
			}
			args = []string{mode, "manifest"}
			if mode == "--audit" {
				args = append(args, "audit")
			}
		}
		command := exec.Command(oracle, args...)
		command.Dir = directory
		answer, err := command.Output()
		if err != nil {
			return fmt.Errorf("oracle %s: %w", mode, err)
		}
		return os.WriteFile(filepath.Join(directory, "answer"), answer, 0644)
	})
	answer, err := os.ReadFile(filepath.Join(directory, "answer"))
	if err != nil {
		t.Fatal(err)
	}
	return answer
}
