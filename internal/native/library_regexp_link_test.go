package native

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// Not parallel: the linking paths share one sanitizer archive and runtime cache.
func TestRegExpCheckerLinkedRuntime(t *testing.T) {
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	archive := os.Getenv("ADAMIC_CLANG_TSGO_ARCHIVE")
	if archive == "" {
		archive = filepath.Join(directory, "checker.a")
		command := exec.Command("go", "build", "-buildmode=c-archive", "-ldflags=-w", "-o", archive, "./bridge/tsgo/archive")
		command.Dir = repository
		command.Env = append(os.Environ(), "CC=clang", "CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all")
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("checker archive: %v\n%s", err, output)
		}
	}
	entry := filepath.Join(directory, "dynamic.a")
	// The lint port's unchanged dynamic_gap.a, so neither path needs an attributed literal.
	fixture := "import { programArguments } from 'adamic';\nconst pattern = programArguments()[0] ?? 'TODO';\nconsole.log(`${new RegExp(pattern, 'u').test('TODO')}`);\n"
	if err := os.WriteFile(entry, []byte(fixture), 0644); err != nil {
		t.Fatal(err)
	}
	node := exec.Command("node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), entry, "TODO")
	want, err := node.CombinedOutput()
	if err != nil {
		t.Fatalf("source Node: %v\n%s", err, want)
	}
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	source := C(lowered)
	for _, builder := range []struct {
		name  string
		build func(string, string, string, Options) error
	}{
		{"checker", BuildTSGo}, {"split-checker", BuildSplitTSGo},
		{"split", func(source, output, archive string, options Options) error {
			options.Split = true
			return Build(source, output, options)
		}},
	} {
		t.Run(builder.name, func(t *testing.T) {
			binary := filepath.Join(directory, builder.name)
			if err := builder.build(source, binary, archive, Options{Sanitize: true, Jobs: 2}); err != nil {
				t.Fatal(err)
			}
			got, err := exec.Command(binary, "TODO").CombinedOutput()
			if err != nil || string(got) != string(want) {
				t.Fatalf("native: %v\nsource Node=%q native=%q", err, want, got)
			}
		})
	}
}
