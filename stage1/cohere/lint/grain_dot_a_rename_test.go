package lint

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"hash/fnv"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

const testDotARenameShards = 2

var dotARenameCases = []string{"no-var-a-to-ts:original", "no-var-a-to-ts:renamed"}

func dotARenameOwner(key string) int {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(key))
	return int(hash.Sum64() % testDotARenameShards)
}

// The original pinned corpus is one rename and its first no-var witness.
// Builds are immutable content-addressed products, never shard-local builds.
type dotARenameProducts struct {
	Original, Changed, Oracle, OriginalNative, ChangedNative, OriginalJS, ChangedJS string
	Before, After                                                                   []byte
}

func dotARenameInputs(t *testing.T, name string, flags []string) buildcache.Inputs {
	t.Helper()
	files := append(rulesAgreeSourceInputs(t), "stage1/cohere/lint/testdata/oracle.go")
	// copyPort also copies witnesses and option sidecars; those enter its key.
	for _, descriptor := range prepareRegistry(t, packageDirectory) {
		directory := filepath.Join("stage1/cohere/lint/rules", descriptor.Slug, "testdata")
		if _, err := os.Stat(filepath.Join(repository, directory)); err == nil {
			files = append(files, filepath.ToSlash(directory))
		} else if !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	return buildcache.Inputs{Name: "dot-a-rename-" + name, Files: files, Flags: flags,
		Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version"), buildcache.Tool("go", "env", "GOOS", "GOARCH", "CGO_ENABLED", "GOEXPERIMENT", "CC", "CXX", "CGO_CFLAGS", "CGO_LDFLAGS"), "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT")}}
}

// Each declaration and filtered shard uses the same once-per-process recipe.
var dotARenameProductOnce [6]sync.Once
var dotARenameProductDirectories [6]string

func dotARenameProduct(t *testing.T, slot int, inputs buildcache.Inputs, recipe func(string) error) string {
	t.Helper()
	dotARenameProductOnce[slot].Do(func() {
		dotARenameProductDirectories[slot] = buildcache.Product(t, inputs, recipe)
	})
	if dotARenameProductDirectories[slot] == "" {
		t.Fatal("DotARename product preparation failed")
	}
	return dotARenameProductDirectories[slot]
}

func dotARenameSource(t *testing.T) string {
	t.Helper()
	return dotARenameProduct(t, 0, dotARenameInputs(t, "renamed-source", []string{"rules/no-var/rule.a -> rule.ts", packageDirectory}), func(dir string) error {
		copyPort(t, dir, "", "")
		return os.Rename(filepath.Join(dir, "rules/no-var/rule.a"), filepath.Join(dir, "rules/no-var/rule.ts"))
	})
}

func dotARenameOracle(t *testing.T) string {
	t.Helper()
	return dotARenameProduct(t, 1, dotARenameInputs(t, "oracle", []string{"go build -overlay", packageDirectory}), func(dir string) error {
		_, err := dotARenameGoOracleIn(packageDirectory, dir)
		return err
	})
}

func dotARenameLowered(t *testing.T, renamed bool) string {
	t.Helper()
	name, directory, slot := "original", packageDirectory, 2
	if renamed {
		name, directory, slot = "renamed", dotARenameSource(t), 3
	}
	inputs := dotARenameInputs(t, name+"-lowered", []string{"native.C", "javascript.JavaScript", name})
	return dotARenameProduct(t, slot, inputs, func(dir string) error {
		prepareRegistry(t, directory)
		phase := time.Now()
		program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
		if err != nil {
			return err
		}
		t.Logf("%s load: %.3fs", name, time.Since(phase).Seconds())
		phase = time.Now()
		result, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		t.Logf("%s lower: %.3fs", name, time.Since(phase).Seconds())
		phase = time.Now()
		if err := os.WriteFile(filepath.Join(dir, "lint.c"), []byte(native.C(result)), 0644); err != nil {
			return err
		}
		t.Logf("%s native emit: %.3fs", name, time.Since(phase).Seconds())
		phase = time.Now()
		defer func() { t.Logf("%s JavaScript emit: %.3fs", name, time.Since(phase).Seconds()) }()
		return os.WriteFile(filepath.Join(dir, "lint.mjs"), []byte(javascript.JavaScript(result)), 0644)
	})
}

func dotARenameNative(t *testing.T, renamed bool) string {
	t.Helper()
	name, slot := "original", 4
	if renamed {
		name, slot = "renamed", 5
	}
	lowered := dotARenameLowered(t, renamed)
	options := native.Options{Sanitize: true, Split: true, Jobs: 4}
	inputs := dotARenameInputs(t, name+"-native", append(native.Flags(options), name, "Split=true", "Jobs=4", "ADAMIC_NATIVE_JOBS="+os.Getenv("ADAMIC_NATIVE_JOBS"), "ADAMIC_GATE_UNCACHED="+os.Getenv("ADAMIC_GATE_UNCACHED")))
	return dotARenameProduct(t, slot, inputs, func(dir string) error {
		data, err := os.ReadFile(filepath.Join(lowered, "lint.c"))
		if err != nil {
			return err
		}
		started := time.Now()
		defer func() { t.Logf("%s clang: %.3fs", name, time.Since(started).Seconds()) }()
		return native.Build(string(data), filepath.Join(dir, "scanner"), options)
	})
}

// C emission exceeds the build phase's 60s grain. Keep lowered/native
// products shard-prepared until that production dependency is split.
func TestProduct_DotARenameSource(t *testing.T) {
	t.Parallel()
	started := time.Now()
	dotARenameSource(t)
	t.Logf("%s: %.3fs", t.Name(), time.Since(started).Seconds())
}

func TestProduct_DotARenameOracle(t *testing.T) {
	t.Parallel()
	started := time.Now()
	dotARenameOracle(t)
	t.Logf("%s: %.3fs", t.Name(), time.Since(started).Seconds())
}

func dotARenameGoOracleIn(sourceRoot, directory string) (string, error) {
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		return "", err
	}
	side, err := filepath.Abs(filepath.Join(packageDirectory, "testdata/oracle.go"))
	if err != nil {
		return "", err
	}
	descriptors, err := registry.Generate(sourceRoot)
	if err != nil {
		return "", err
	}
	replacements := map[string]string{}
	var virtualFiles []string
	var failure error
	add := func(name, source string) {
		virtual := filepath.Join(root, "adamic_lint_"+name+".go")
		absolute, err := filepath.Abs(source)
		if err != nil {
			failure = err
			return
		}
		replacements[virtual] = absolute
		virtualFiles = append(virtualFiles, virtual)
	}
	add("oracle", side)
	add("registry", filepath.Join(sourceRoot, ".generated/registry.go"))
	for _, d := range descriptors {
		add(strings.ReplaceAll(d.Slug, "-", "_"), filepath.Join(sourceRoot, "rules", d.Slug, "oracle.go"))
	}
	if failure != nil {
		return "", failure
	}
	overlay, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		return "", err
	}
	path := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		return "", err
	}
	binary := filepath.Join(directory, "oracle")
	args := append([]string{"build", "-overlay=" + path, "-o", binary}, virtualFiles...)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	command := exec.CommandContext(ctx, "go", args...)
	command.Dir = root
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	var diagnostics bytes.Buffer
	command.Stderr = &diagnostics
	if err := command.Run(); err != nil || len(commandDiagnostics("go", diagnostics.Bytes())) != 0 {
		return "", fmt.Errorf("Go oracle build: %v\n%s", err, diagnostics.Bytes())
	}
	return binary, nil
}

func dotARenameSetup(t *testing.T, renamed bool) dotARenameProducts {
	t.Helper()
	started := time.Now()
	defer func() { t.Logf("TestDotARename (setup): %.3f s", time.Since(started).Seconds()) }()
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	p := dotARenameProducts{Original: directory}
	var readErr error
	var workers sync.WaitGroup
	workers.Add(3)
	go func() {
		defer workers.Done()
		p.Oracle = filepath.Join(dotARenameOracle(t), "oracle")
	}()
	go func() {
		defer workers.Done()
		p.Changed = dotARenameSource(t)
		p.Before, readErr = os.ReadFile(filepath.Join(directory, "rules/no-var/rule.a"))
		if readErr != nil {
			return
		}
		p.After, readErr = os.ReadFile(filepath.Join(p.Changed, "rules/no-var/rule.ts"))
		if readErr != nil {
			return
		}
	}()
	go func() {
		defer workers.Done()
		if renamed {
			p.ChangedNative = filepath.Join(dotARenameNative(t, true), "scanner")
			p.ChangedJS = filepath.Join(dotARenameLowered(t, true), "lint.mjs")
		} else {
			p.OriginalNative = filepath.Join(dotARenameNative(t, false), "scanner")
			p.OriginalJS = filepath.Join(dotARenameLowered(t, false), "lint.mjs")
		}
	}()
	workers.Wait()
	if readErr != nil {
		t.Fatal(readErr)
	}
	if t.Failed() {
		t.Fatal("DotARename preparation failed")
	}
	return p
}

func dotARenameUnion(t *testing.T) {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "grain_dot_a_rename_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, declaration := range file.Decls {
		if fn, ok := declaration.(*ast.FuncDecl); ok && strings.HasPrefix(fn.Name.Name, "TestDotARename_") && fn.Name.Name != "TestDotARename_Setup" {
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
	cases := dotARenameCases
	owners := make(map[string]int)
	for shard := 0; shard < testDotARenameShards; shard++ {
		for _, key := range cases {
			if dotARenameOwner(key) == shard {
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

var dotARenameSetupOnce [2]sync.Once
var dotARenamePrepared [2]*dotARenameProducts

// Filtered shards prepare their own real products before the case deadline.
func dotARenameReady(t *testing.T, renamed bool) dotARenameProducts {
	t.Helper()
	slot := 0
	if renamed {
		slot = 1
	}
	dotARenameSetupOnce[slot].Do(func() {
		products := dotARenameSetup(t, renamed)
		dotARenamePrepared[slot] = &products
	})
	if dotARenamePrepared[slot] == nil {
		t.Fatal("DotARename setup did not complete")
	}
	return *dotARenamePrepared[slot]
}

func dotARenameChild(ctx context.Context, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, os.Args[0], args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = 5 * time.Second
	return command
}

type dotARenameProbe struct {
	Key      string
	Products dotARenameProducts
}

func TestDotARenameUnion(t *testing.T) {
	t.Parallel()
	dotARenameUnion(t)
}

// Both variants must equal the same Go oracle: this preserves the original
// rename-equivalence assertion without building both ports in one cold unit.
func dotARenameRunShard(t *testing.T, shard int) {
	t.Helper()
	dotARenameUnion(t)
	var probe dotARenameProbe
	supplied := os.Getenv("ADAMIC_DOT_A_RENAME_PROBE")
	if supplied != "" {
		data, err := os.ReadFile(supplied)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, &probe); err != nil {
			t.Fatal(err)
		}
	}
	for _, key := range dotARenameCases {
		if dotARenameOwner(key) != shard || (supplied != "" && probe.Key != key) {
			continue
		}
		renamed := strings.HasSuffix(key, ":renamed")
		p := probe.Products
		if supplied == "" {
			p = dotARenameReady(t, renamed)
		}
		started := time.Now()
		deadline := time.AfterFunc(90*time.Second, func() { panic("P0: DotARename shard exceeded 90s") })
		defer deadline.Stop()
		defer func() { t.Logf("%s (own work): %.3fs", key, time.Since(started).Seconds()) }()
		if !bytes.Equal(p.Before, p.After) {
			t.Fatal("rename changed module bytes")
		}
		path := manifest(t, []string{ownedWitnesses(t, p.Original, "no-var")[0] + "\tno-var"})
		directory, binary, module := p.Original, p.OriginalNative, p.OriginalJS
		if renamed {
			directory, binary, module = p.Changed, p.ChangedNative, p.ChangedJS
		}
		want := compareWithJavaScript(t, p.Oracle, binary, directory, path, module)
		t.Logf("%s: .a/.ts variant matches the shared Go oracle on all three runtimes (%d bytes)", key, len(want))
		if supplied == "" {
			dotARenamePlant(t, key, p, module)
		}
		if time.Since(started) >= 60*time.Second {
			t.Fatal("DotARename own work over 60s; split smaller")
		}
	}
}

func dotARenamePlant(t *testing.T, key string, p dotARenameProducts, module string) {
	t.Helper()
	data, err := os.ReadFile(module)
	if err != nil {
		t.Fatal(err)
	}
	planted := filepath.Join(t.TempDir(), "planted.mjs")
	marker := "planted dot-a rename disagreement"
	if err := os.WriteFile(planted, append(data, []byte(fmt.Sprintf("\nconsole.log(%q);\n", marker))...), 0644); err != nil {
		t.Fatal(err)
	}
	if strings.HasSuffix(key, ":renamed") {
		p.ChangedJS = planted
	} else {
		p.OriginalJS = planted
	}
	data, err = json.Marshal(dotARenameProbe{Key: key, Products: p})
	if err != nil {
		t.Fatal(err)
	}
	products := filepath.Join(t.TempDir(), "products.json")
	if err := os.WriteFile(products, data, 0644); err != nil {
		t.Fatal(err)
	}
	caught := 0
	for shard := 0; shard < testDotARenameShards; shard++ {
		name := fmt.Sprintf("TestDotARename_%03d", shard)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		command := dotARenameChild(ctx, "-test.run=^"+name+"$", "-test.v", "-test.timeout=10s")
		command.Env = append(os.Environ(), "ADAMIC_DOT_A_RENAME_PROBE="+products)
		output, err := command.CombinedOutput()
		expired := ctx.Err()
		cancel()
		if expired != nil {
			t.Fatalf("planted probe exceeded its deadline: %v\n%s", expired, output)
		}
		if err != nil {
			if bytes.Count(output, []byte("--- FAIL: "+name)) != 1 || !bytes.Contains(output, []byte("emitted JavaScript:")) || !bytes.Contains(output, []byte(marker)) {
				t.Fatalf("wrong planted failure: %v\n%s", err, output)
			}
			caught++
		}
	}
	if caught != 1 {
		t.Fatalf("planted disagreement caught %d times", caught)
	}
	t.Logf("planted disagreement caught exactly by TestDotARename_%03d", dotARenameOwner(key))
}

func TestDotARename_000(t *testing.T) {
	t.Parallel()
	dotARenameRunShard(t, 0)
}

func TestDotARename_001(t *testing.T) {
	t.Parallel()
	dotARenameRunShard(t, 1)
}
