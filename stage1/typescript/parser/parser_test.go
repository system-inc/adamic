package parser

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const repository = "../../.."
const compilerCommit = "050880ce59e30b356b686bd3144efe24f875ebc8"

var portFiles = []string{"nodes.ts", "grammar.ts", "lookahead.ts", "statements.ts", "jsx.ts", "parser.ts", "main.ts"}

type execution struct {
	output   []byte
	duration time.Duration
}

func goOracle(t *testing.T) string {
	t.Helper()
	return parserGuardOracle(t)
}

func buildPort(t *testing.T, directory string, sanitize bool) string {
	t.Helper()
	program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "scanner")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: sanitize}); err != nil {
		t.Fatal(err)
	}
	return binary
}

func copyPort(t *testing.T, file, from, to string) string {
	t.Helper()
	directory := t.TempDir()
	for _, name := range portFiles {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		source := string(data)
		scannerDirectory, err := filepath.Abs("../scanner/scanner.ts")
		if err != nil {
			t.Fatal(err)
		}
		source = strings.ReplaceAll(source, "../scanner/scanner.ts", scannerDirectory)
		if name == file {
			if strings.Count(source, from) != 1 {
				t.Fatalf("mutant must change exactly one site in %s: %q", file, from)
			}
			source = strings.Replace(source, from, to, 1)
		}
		if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return directory
}

func node(t *testing.T, directory, manifest string, count bool) execution {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", manifest}
	if count {
		args = append(args, "--count")
	}
	return execute(t, "", "node", args...)
}

func difference(got, want []byte) string {
	if bytes.Equal(got, want) {
		return ""
	}
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) || i < len(b); i++ {
		left, right := "<EOF>", "<EOF>"
		if i < len(a) {
			left = a[i]
		}
		if i < len(b) {
			right = b[i]
		}
		if left != right {
			caseLine := ""
			for j := i; j >= 0; j-- {
				if j < len(b) && strings.HasPrefix(b[j], "case ") {
					caseLine = b[j]
					break
				}
			}
			start := i - 5
			if start < 0 {
				start = 0
			}
			end := i + 5
			if end > len(b) {
				end = len(b)
			}
			return fmt.Sprintf("%s line %d: port %q, Go %q\nGo context:\n%s", caseLine, i+1, left, right, strings.Join(b[start:end], "\n"))
		}
	}
	return "different bytes"
}

func compilerManifest(t *testing.T) (string, int) {
	t.Helper()
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		t.Skip("set ADAMIC_TYPESCRIPT_SOURCE to the pinned v6.0.3 checkout")
	}
	output, err := exec.Command("git", "-C", source, "rev-parse", "HEAD").Output()
	if err != nil || strings.TrimSpace(string(output)) != compilerCommit {
		t.Fatalf("corpus pin differs: %q %v", output, err)
	}
	var manifest strings.Builder
	files := 0
	err = filepath.WalkDir(filepath.Join(source, "src/compiler"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".ts") {
			manifest.WriteString(path + "\n")
			files++
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "compiler.txt")
	if err := os.WriteFile(path, []byte(manifest.String()), 0644); err != nil {
		t.Fatal(err)
	}
	return path, files
}

// Not parallel: initializes shared compilerExpressionsProducts before the parallel shards run.
func TestCompilerExpressionsAgree_Setup(t *testing.T) {
	compilerExpressionsSetup(t)
}
