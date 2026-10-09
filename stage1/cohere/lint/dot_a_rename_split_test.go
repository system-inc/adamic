package lint

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const testDotARenameShards = 1

// The original pinned corpus is one rename and its first no-var witness.
// Builds are immutable content-addressed products, never shard-local builds.
type dotARenameProducts struct {
	Original, Changed, Oracle, OriginalNative, ChangedNative, OriginalJS, ChangedJS string
	Before, After                                                                   []byte
}

func dotARenameInputs(t *testing.T, name string, flags []string) buildcache.Inputs {
	t.Helper()
	files := []string{"internal", "bridge", "stage1/typescript", "cohere/TypeScript/tsc", "cohere/TypeScript-shim", "go.mod", "go.work", "cohere/go.mod", "cohere/go.sum", "stage1/cohere/lint/dot_a_rename_split_test.go"}
	for _, path := range portFiles(t) {
		files = append(files, filepath.ToSlash(filepath.Join("stage1/cohere/lint", path)))
	}
	return buildcache.Inputs{Name: "dot-a-rename-" + name, Files: files, Flags: flags,
		Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version"), "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT")}}
}

func dotARenameBuild(t *testing.T, directory, name string) (string, string) {
	t.Helper()
	inputs := dotARenameInputs(t, name+"-lowered", []string{"native.C", "javascript.JavaScript", directory})
	lowered := buildcache.Product(t, inputs, func(dir string) error {
		prepareRegistry(t, directory)
		program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
		if err != nil {
			return err
		}
		ir, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "lint.c"), []byte(native.C(ir)), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "lint.mjs"), []byte(javascript.JavaScript(ir)), 0644)
	})
	inputs = dotARenameInputs(t, name+"-native", append(native.Flags(native.Options{Sanitize: true}), directory))
	binary := buildcache.Product(t, inputs, func(dir string) error {
		data, err := os.ReadFile(filepath.Join(lowered, "lint.c"))
		if err != nil {
			return err
		}
		return native.Build(string(data), filepath.Join(dir, "scanner"), native.Options{Sanitize: true})
	})
	return filepath.Join(binary, "scanner"), filepath.Join(lowered, "lint.mjs")
}

func dotARenameSetup(t *testing.T) dotARenameProducts {
	t.Helper()
	started := time.Now()
	defer func() { t.Logf("TestDotARename (setup): %.3f s", time.Since(started).Seconds()) }()
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	p := dotARenameProducts{Original: directory, Oracle: goOracle(t)}
	// Keep the Go build as before: GoBuild has not landed on this base.
	source := buildcache.Product(t, dotARenameInputs(t, "renamed-source", []string{"rules/no-var/rule.a -> rule.ts", directory}), func(dir string) error {
		copyPort(t, dir, "", "")
		return os.Rename(filepath.Join(dir, "rules/no-var/rule.a"), filepath.Join(dir, "rules/no-var/rule.ts"))
	})
	p.Changed = source
	p.Before, err = os.ReadFile(filepath.Join(directory, "rules/no-var/rule.a"))
	if err != nil {
		t.Fatal(err)
	}
	p.After, err = os.ReadFile(filepath.Join(source, "rules/no-var/rule.ts"))
	if err != nil {
		t.Fatal(err)
	}
	p.OriginalNative, p.OriginalJS = dotARenameBuild(t, directory, "original")
	p.ChangedNative, p.ChangedJS = dotARenameBuild(t, source, "renamed")
	return p
}

func dotARenameUnion(t *testing.T) {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "dot_a_rename_split_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, declaration := range file.Decls {
		if fn, ok := declaration.(*ast.FuncDecl); ok && strings.HasPrefix(fn.Name.Name, "TestDotARename_") {
			names[fn.Name.Name] = true
		}
	}
	if len(names) != testDotARenameShards {
		t.Fatalf("enumerated %d shards, declared %d", len(names), testDotARenameShards)
	}
	for i := 0; i < testDotARenameShards; i++ {
		if !names[fmt.Sprintf("TestDotARename_%03d", i)] {
			t.Fatalf("missing shard %d", i)
		}
	}
	// The pinned original enumeration has exactly this one semantic case.
	cases := []string{"no-var-a-to-ts"}
	owners := make(map[string]int)
	for shard := 0; shard < testDotARenameShards; shard++ {
		for index, key := range cases {
			if index%testDotARenameShards == shard {
				owners[key]++
			}
		}
	}
	for _, key := range cases {
		if owners[key] != 1 {
			t.Fatalf("case %s has %d owners", key, owners[key])
		}
	}
	t.Logf("union: %d/%d cases exactly once", len(owners), len(cases))
}

func TestDotARename_000(t *testing.T) {
	t.Parallel()
	dotARenameUnion(t)
	var p dotARenameProducts
	supplied := os.Getenv("ADAMIC_DOT_A_RENAME_PROBE")
	if supplied != "" {
		data, err := os.ReadFile(supplied)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, &p); err != nil {
			t.Fatal(err)
		}
	} else {
		p = dotARenameSetup(t)
	}
	if !bytes.Equal(p.Before, p.After) {
		t.Fatal("rename changed module bytes")
	}
	path := manifest(t, []string{ownedWitnesses(t, p.Original, "no-var")[0] + "\tno-var"})
	want := compareWithJavaScript(t, p.Oracle, p.OriginalNative, p.Original, path, p.OriginalJS)
	got := compareWithJavaScript(t, p.Oracle, p.ChangedNative, p.Changed, path, p.ChangedJS)
	if !bytes.Equal(got, want) {
		t.Fatal("rename changed results")
	}
	t.Logf("rename only: .ts and .a identical on all three runtimes against Go (%d bytes)", len(want))
	if supplied != "" {
		return
	}
	// Run the ordinary owning shard with one emitted result deliberately changed.
	data, err := os.ReadFile(p.ChangedJS)
	if err != nil {
		t.Fatal(err)
	}
	p.ChangedJS = filepath.Join(t.TempDir(), "planted.mjs")
	marker := "planted dot-a rename disagreement"
	if err := os.WriteFile(p.ChangedJS, append(data, []byte(fmt.Sprintf("\nconsole.log(%q);\n", marker))...), 0644); err != nil {
		t.Fatal(err)
	}
	data, err = json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	products := filepath.Join(t.TempDir(), "products.json")
	if err := os.WriteFile(products, data, 0644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestDotARename_000$", "-test.v", "-test.timeout=30s")
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = 5 * time.Second
	command.Env = append(os.Environ(), "ADAMIC_DOT_A_RENAME_PROBE="+products)
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("planted probe exceeded its deadline: %v\n%s", ctx.Err(), output)
	}
	if err == nil || bytes.Count(output, []byte("--- FAIL: TestDotARename_000")) != 1 || !bytes.Contains(output, []byte("emitted JavaScript:")) || !bytes.Contains(output, []byte(marker)) {
		t.Fatalf("wrong planted failure: %v\n%s", err, output)
	}
	t.Log("planted disagreement caught exactly by TestDotARename_000")
}
