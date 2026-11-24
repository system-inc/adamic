package cloud

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// sandboxed-job.sh keeps the kit's own files outside the tree it lays, so a walker that reads every file under $PWD never
// meets them (internal/acheck read the cache's runtime.a as programs on CLI's J-0pmqbfg; Adamic's rule on #qee4r6g). The
// build cache and the TypeScript copy go to LOOM_JOB_SCRATCH where the runner names one and to TMPDIR where it doesn't,
// and a scratch inside the tree is refused before the command.
//
// Mutants: ADAMIC_BUILD_CACHE_DIR exported under $PWD again (both outside cases red on the cache); the TypeScript copy
// under $PWD again (both red on the copy); the inside-the-tree refusal taken out (the inside case sources clean).
func TestSandboxedJobKeepsTheKitOutsideTheTree(t *testing.T) {
	t.Parallel()
	script, err := filepath.Abs("sandboxed-job.sh")
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name    string
		scratch func(tree, outside string) string
		refused string
	}{
		{"the runner names no scratch", func(tree, outside string) string { return "" }, ""},
		{"the runner names a scratch beside the tree", func(tree, outside string) string { return filepath.Join(outside, "scratch") }, ""},
		{"the runner names a scratch inside the tree", func(tree, outside string) string { return filepath.Join(tree, "scratch") }, "is inside the tree"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			tree, outside, source := filepath.Join(directory, "tree"), filepath.Join(directory, "outside"), filepath.Join(directory, "typescript")
			for _, made := range []string{tree, outside, source} {
				if err := os.Mkdir(made, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(filepath.Join(source, "pin"), []byte("x\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			scratch := test.scratch(tree, outside)
			sourced := `. "$SANDBOXED_JOB_SCRIPT" || { echo "sourced=failed"; exit 0; }
echo "sourced=ok"
sandboxed_typescript_source > /dev/null || exit 1
printf 'cache=%s\ncopy=%s\n' "$ADAMIC_BUILD_CACHE_DIR" "$ADAMIC_TYPESCRIPT_SOURCE"
`
			command := toolchain.Command("bash", "-c", sourced)
			if command.Err != nil {
				t.Fatal(command.Err)
			}
			command.Dir = tree
			environment := command.Env
			if environment == nil {
				environment = os.Environ()
			}
			kept := make([]string, 0, len(environment))
			for _, variable := range environment {
				name, _, _ := strings.Cut(variable, "=")
				if name != "LOOM_JOB_SCRATCH" && name != "TMPDIR" && name != "ADAMIC_BUILD_CACHE_DIR" && name != "ADAMIC_TYPESCRIPT_SOURCE" {
					kept = append(kept, variable)
				}
			}
			kept = append(kept, "SANDBOXED_JOB_SCRIPT="+script, "TMPDIR="+filepath.Join(outside, "tmp"), "ADAMIC_TYPESCRIPT_SOURCE="+source)
			if scratch != "" {
				kept = append(kept, "LOOM_JOB_SCRATCH="+scratch)
			}
			command.Env = kept
			output, err := command.CombinedOutput()
			said := string(output)
			if err != nil {
				t.Fatalf("bash: %v\n%s", err, said)
			}
			if test.refused != "" {
				if !strings.Contains(said, "sourced=failed") || !strings.Contains(said, test.refused) {
					t.Fatalf("want the source refused saying %q, got:\n%s", test.refused, said)
				}
				return
			}
			if !strings.Contains(said, "sourced=ok") {
				t.Fatalf("want the source to pass, got:\n%s", said)
			}
			realTree, err := filepath.EvalSymlinks(tree)
			if err != nil {
				t.Fatal(err)
			}
			for _, key := range []string{"cache", "copy"} {
				value := ""
				for _, line := range strings.Split(said, "\n") {
					if found, ok := strings.CutPrefix(line, key+"="); ok {
						value = found
					}
				}
				if value == "" {
					t.Fatalf("want a %s line, got:\n%s", key, said)
				}
				resolved, err := filepath.EvalSymlinks(value)
				if err != nil {
					t.Fatalf("the %s %s wasn't made: %v", key, value, err)
				}
				if resolved == realTree || strings.HasPrefix(resolved, realTree+string(filepath.Separator)) {
					t.Errorf("the %s %s is inside the tree %s", key, value, tree)
				}
				if scratch != "" && !strings.HasPrefix(value, scratch+string(filepath.Separator)) {
					t.Errorf("the %s %s isn't under the runner's scratch %s", key, value, scratch)
				}
			}
		})
	}
}

// sandboxed-job.sh puts Go's work directory ($WORK, where every action of a build writes its scratch files) under
// LOOM_JOB_SCRATCH where the runner names one. go makes it under GOTMPDIR, else TMPDIR, and a house box's TMPDIR is a
// 16 GB /tmp held in memory, which stage1/cohere/markdownblocks' oracle builds filled while the job's scratch on the data
// disk sat empty (J-3y63xbt, J-ramwydp; #7k79mkn). Where the runner names none, GOTMPDIR stays the caller's (unset
// here) and go works under TMPDIR as before. The test sources the script in a temporary tree holding a module of one
// empty main, runs a real `go build -work` there with GOTOOLCHAIN=local and nothing to fetch, and reads the WORK= line go
// prints on stderr.
//
// Mutant: the GOTMPDIR export dropped (the scratch case red, its WORK under TMPDIR).
func TestSandboxedJobGivesGoTheScratchToWorkIn(t *testing.T) {
	t.Parallel()
	script, err := filepath.Abs("sandboxed-job.sh")
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name    string
		scratch bool
	}{
		{"the runner names a scratch beside the tree", true},
		{"the runner names no scratch", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			tree, outside := filepath.Join(directory, "tree"), filepath.Join(directory, "outside")
			for _, made := range []string{tree, outside} {
				if err := os.Mkdir(made, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			module := map[string]string{"go.mod": "module example.invalid/probe\n\ngo 1.22\n", "main.go": "package main\n\nfunc main() {}\n"}
			for name, content := range module {
				if err := os.WriteFile(filepath.Join(tree, name), []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			scratch, temporary := filepath.Join(outside, "scratch"), filepath.Join(outside, "tmp")
			command := toolchain.Command("bash", "-c", `. "$SANDBOXED_JOB_SCRIPT" || exit 1
if [ -n "${GOTMPDIR+set}" ]; then echo "gotmpdir=$GOTMPDIR"; else echo "gotmpdir unset"; fi
go build -work -o "$TMPDIR/probe" .
`)
			if command.Err != nil {
				t.Fatal(command.Err)
			}
			command.Dir = tree
			environment := command.Env
			if environment == nil {
				environment = os.Environ()
			}
			// The caller's GOFLAGS, workspace, toolchain policy and shared cache (GOCACHEPROG, which may read a store
			// over HTTP) are left out, so the build reads its own module and the local release's standard library alone.
			kept := make([]string, 0, len(environment)+8)
			for _, variable := range environment {
				name, _, _ := strings.Cut(variable, "=")
				switch name {
				case "TMPDIR", "GOTMPDIR", "LOOM_JOB_SCRATCH", "ADAMIC_BUILD_CACHE_DIR", "GOFLAGS", "GOWORK", "GOTOOLCHAIN", "GOPROXY", "GOCACHEPROG":
				default:
					if !strings.HasPrefix(name, "ADAMIC_GOCACHE_") {
						kept = append(kept, variable)
					}
				}
			}
			kept = append(kept, "SANDBOXED_JOB_SCRIPT="+script, "TMPDIR="+temporary, "GOFLAGS=-buildvcs=false", "GOWORK=off", "GOTOOLCHAIN=local",
				"GOPROXY=off")
			if test.scratch {
				kept = append(kept, "LOOM_JOB_SCRATCH="+scratch)
			}
			command.Env = kept
			output, err := command.CombinedOutput()
			said := string(output)
			if err != nil {
				t.Fatalf("bash: %v\n%s", err, said)
			}
			work := ""
			for _, line := range strings.Split(said, "\n") {
				if found, ok := strings.CutPrefix(line, "WORK="); ok {
					work = found
				}
			}
			if work == "" {
				t.Fatalf("want go's WORK= line, got:\n%s", said)
			}
			// -work keeps the directory, so it resolves, and t.TempDir's cleanup takes it with the rest
			resolved, err := filepath.EvalSymlinks(work)
			if err != nil {
				t.Fatalf("go's work directory %s wasn't kept: %v", work, err)
			}
			want := temporary
			if test.scratch {
				want = scratch
			}
			resolvedWant, err := filepath.EvalSymlinks(want)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(resolved, resolvedWant+string(filepath.Separator)) {
				t.Errorf("go worked in %s, want it under %s:\n%s", work, want, said)
			}
			if !test.scratch && !strings.Contains(said, "gotmpdir unset") {
				t.Errorf("want GOTMPDIR left unset where the runner names no scratch, got:\n%s", said)
			}
		})
	}
}

// sandboxed-job.sh gives the job's TMPDIR to the scratch's own tmp where the runner names LOOM_JOB_SCRATCH, made 0700 and
// the job's, so every temporary directory a job makes lands on the scratch's disk: internal/native's adamic-build-*,
// adamic-units-* and adamic-runtime-headers-* filled Lunatic's 16 GB /tmp held in memory while its scratch sat nearly
// empty (J-54dascs; #7k79mkn). A scratch's tmp that resolves inside the tree keeps the caller's TMPDIR and says so, and
// where the runner names no scratch TMPDIR stays the caller's. The test sources the script twice in a temporary tree,
// under `umask 022`, makes a directory with mktemp -d between, and reads TMPDIR after each source; TMPDIR is set
// outright, so both reads agree.
//
// Mutants: the `export TMPDIR="$sandboxedJobTemporary"` line dropped (red on the scratch beside the tree, its TMPDIR the
// caller's); the `case` that refuses a path inside the tree dropped, so it always exports (red on the link into the
// tree, its TMPDIR the link); `chmod 0700` dropped (red on the scratch beside the tree's mode, since `umask 022` leaves
// mkdir's 0755); the chmod moved ahead of the refusal (red on the link into the tree, whose inside turns 0700).
func TestSandboxedJobGivesTheScratchItsTemporaryDirectory(t *testing.T) {
	t.Parallel()
	script, err := filepath.Abs("sandboxed-job.sh")
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name string
		// scratch names LOOM_JOB_SCRATCH; linked makes the scratch's tmp a link to a directory inside the tree before the
		// source; stays is what the output says where TMPDIR stays the caller's, or "" where nothing says it stays.
		scratch bool
		linked  bool
		stays   string
	}{
		{"the runner names a scratch beside the tree", true, false, ""},
		{"the runner names no scratch", false, false, ""},
		{"the scratch's tmp is a link into the tree", true, true, "resolves inside the tree"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			tree, outside := filepath.Join(directory, "tree"), filepath.Join(directory, "outside")
			for _, made := range []string{tree, outside} {
				if err := os.Mkdir(made, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			scratch, temporary := filepath.Join(outside, "scratch"), filepath.Join(outside, "tmp")
			if test.linked {
				inside := filepath.Join(tree, "inside")
				for _, made := range []string{inside, scratch} {
					if err := os.Mkdir(made, 0o755); err != nil {
						t.Fatal(err)
					}
				}
				if err := os.Symlink(inside, filepath.Join(scratch, "tmp")); err != nil {
					t.Fatal(err)
				}
			}
			command := toolchain.Command("bash", "-c", `umask 022
. "$SANDBOXED_JOB_SCRIPT" || exit 1
printf 'tmpdir=%s\n' "$TMPDIR"
made=$(mktemp -d) || exit 1
printf 'made=%s\n' "$made"
. "$SANDBOXED_JOB_SCRIPT" || exit 1
printf 'tmpdir-again=%s\n' "$TMPDIR"
`)
			if command.Err != nil {
				t.Fatal(command.Err)
			}
			command.Dir = tree
			environment := command.Env
			if environment == nil {
				environment = os.Environ()
			}
			kept := make([]string, 0, len(environment)+3)
			for _, variable := range environment {
				name, _, _ := strings.Cut(variable, "=")
				switch name {
				case "TMPDIR", "LOOM_JOB_SCRATCH", "ADAMIC_BUILD_CACHE_DIR", "GOTMPDIR":
				default:
					kept = append(kept, variable)
				}
			}
			kept = append(kept, "SANDBOXED_JOB_SCRIPT="+script, "TMPDIR="+temporary)
			if test.scratch {
				kept = append(kept, "LOOM_JOB_SCRATCH="+scratch)
			}
			command.Env = kept
			output, err := command.CombinedOutput()
			said := string(output)
			if err != nil {
				t.Fatalf("bash: %v\n%s", err, said)
			}
			values := map[string]string{}
			for _, line := range strings.Split(said, "\n") {
				for _, key := range []string{"tmpdir", "made", "tmpdir-again"} {
					if found, ok := strings.CutPrefix(line, key+"="); ok {
						values[key] = found
					}
				}
			}
			for _, key := range []string{"tmpdir", "made", "tmpdir-again"} {
				if values[key] == "" {
					t.Fatalf("want a %s line, got:\n%s", key, said)
				}
			}
			if values["tmpdir-again"] != values["tmpdir"] {
				t.Errorf("TMPDIR is %s after the first source and %s after the second, want one value:\n%s", values["tmpdir"],
					values["tmpdir-again"], said)
			}
			want := temporary
			if test.scratch && !test.linked {
				want = filepath.Join(scratch, "tmp")
			}
			resolvedWant, err := filepath.EvalSymlinks(want)
			if err != nil {
				t.Fatalf("the TMPDIR wanted, %s, wasn't made: %v\n%s", want, err, said)
			}
			resolved, err := filepath.EvalSymlinks(values["tmpdir"])
			if err != nil {
				t.Fatalf("TMPDIR %s wasn't made: %v\n%s", values["tmpdir"], err, said)
			}
			if resolved != resolvedWant {
				t.Errorf("TMPDIR is %s, want %s:\n%s", values["tmpdir"], want, said)
			}
			// mktemp -d keeps the directory, so it resolves, and t.TempDir's cleanup takes it with the rest
			resolvedMade, err := filepath.EvalSymlinks(values["made"])
			if err != nil {
				t.Fatalf("mktemp -d's %s wasn't kept: %v\n%s", values["made"], err, said)
			}
			if !strings.HasPrefix(resolvedMade, resolvedWant+string(filepath.Separator)) {
				t.Errorf("mktemp -d made %s, want it under %s:\n%s", values["made"], want, said)
			}
			if test.stays != "" {
				if !strings.Contains(said, test.stays) {
					t.Errorf("want %q said, got:\n%s", test.stays, said)
				}
				// The refusal comes before any chmod, so the directory inside the tree keeps its own mode.
				if test.linked {
					info, err := os.Stat(filepath.Join(tree, "inside"))
					if err != nil {
						t.Fatal(err)
					}
					if info.Mode().Perm() != 0o755 {
						t.Errorf("the tree's inside is %v after the source, want 0755 left alone", info.Mode())
					}
				}
				return
			}
			if strings.Contains(said, "TMPDIR stays") {
				t.Errorf("want nothing said of TMPDIR staying, got:\n%s", said)
			}
			if !test.scratch {
				return
			}
			info, err := os.Stat(want)
			if err != nil {
				t.Fatal(err)
			}
			if !info.IsDir() || info.Mode().Perm() != 0o700 {
				t.Errorf("the scratch's tmp %s is %v, want a directory of mode 0700", want, info.Mode())
			}
		})
	}
}

// sandboxed_products gives its go test no clock but the job's own ceiling: go's default 10 minutes killed internal/oracle's
// products before they settled on Lunatic's 8 cores (Library's J-1xphvcc, routed by Runtime on #jobs), so a slow box
// read as the head's red. The test sources the script in a temporary tree beside a stand-in go that lists one product
// test and writes down the arguments its run was given, and holds that the products' go test asked `-timeout 0`.
//
// Mutant: the `-timeout 0` taken out of sandboxed_products' go test (the stand-in's arguments name no timeout, red here).
func TestSandboxedProductsGivesGoNoClockButTheJobs(t *testing.T) {
	t.Parallel()
	script, err := filepath.Abs("sandboxed-job.sh")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	tree, bin := filepath.Join(directory, "tree"), filepath.Join(directory, "bin")
	for _, made := range []string{filepath.Join(tree, "cloud", "fast-gate"), filepath.Join(tree, "internal", "oracle"), bin} {
		if err := os.MkdirAll(made, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(tree, "cloud", "fast-gate", "compiler-dependencies.json"), []byte(`{"packages":{}}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	said := filepath.Join(directory, "go-arguments")
	stand := "#!/bin/sh\ncase \" $* \" in *\" -list \"*) echo TestProduct_Oracle; exit 0 ;; esac\nprintf '%s\\n' \"$@\" > \"$GO_ARGUMENTS\"\n"
	if err := os.WriteFile(filepath.Join(bin, "go"), []byte(stand), 0o755); err != nil {
		t.Fatal(err)
	}
	// The TypeScript copy is the image's road, not this test's: it stands in as done.
	sourced := `. "$SANDBOXED_JOB_SCRIPT" || exit 1
sandboxed_typescript_source() { :; }
sandboxed_products ./internal/oracle
`
	command := toolchain.Command("bash", "-c", sourced)
	if command.Err != nil {
		t.Fatal(command.Err)
	}
	command.Dir = tree
	environment := command.Env
	if environment == nil {
		environment = os.Environ()
	}
	path := os.Getenv("PATH")
	kept := make([]string, 0, len(environment)+4)
	for _, variable := range environment {
		name, value, _ := strings.Cut(variable, "=")
		switch name {
		case "PATH":
			path = value
		case "TMPDIR", "ADAMIC_BUILD_CACHE_DIR", "LOOM_JOB_SCRATCH":
		default:
			kept = append(kept, variable)
		}
	}
	command.Env = append(kept, "PATH="+bin+":"+path, "TMPDIR="+filepath.Join(directory, "tmp"), "GO_ARGUMENTS="+said,
		"SANDBOXED_JOB_SCRIPT="+script)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("sandboxed_products: %v\n%s", err, output)
	}
	data, err := os.ReadFile(said)
	if err != nil {
		t.Fatalf("the stand-in go was never run: %v", err)
	}
	arguments := strings.Split(strings.TrimSpace(string(data)), "\n")
	for index, argument := range arguments {
		if argument == "-timeout" && index+1 < len(arguments) && arguments[index+1] == "0" {
			return
		}
		if argument == "-timeout=0" {
			return
		}
	}
	t.Fatalf("sandboxed_products' go test named no -timeout 0, so go's own 10 minutes would end a slow settle: %q", arguments)
}

// sandboxed_products settles product tests wave by wave in the order cloud/fast-gate/product-order.json gives, a package
// at a time, since one go test over every package ran them at once and a test that reads another's product got there
// first (internal/worker's TestProduct_WorkerWasm_005). A wait on a test in a package the union doesn't hold brings that
// test alone, as stage1/cohere/estree's tests wait on stage1/cohere/estree/oracles'. The test sources the script in a
// temporary tree beside a stand-in go that lists each package's product tests from a fixed set, among go's own lines,
// and writes down each run's package and -run expression. app is named, compiler-dependencies.json lists lib under it,
// and oracles is reached only through a wait. A wait on a test lib doesn't list is dropped; a cycle refuses before
// anything runs; a wave that fails stops every wave after it.
//
// Mutants: the closure over waits dropped (`unlisted.add(package)` made `pass`, red on the order: oracles never runs and
// app's tests run with it unsettled); the waves run backwards (`enumerate(waves, 1)` made `enumerate(reversed(waves), 1)`,
// red on the order).
func TestSandboxedProductsSettlesInWaves(t *testing.T) {
	t.Parallel()
	script, err := filepath.Abs("sandboxed-job.sh")
	if err != nil {
		t.Fatal(err)
	}
	waits := map[string][]string{
		"app TestProduct_App":        {"lib TestProduct_Lib", "oracles TestProduct_Oracle"},
		"app TestProduct_AppLate":    {"app TestProduct_App", "lib TestProduct_Gone"},
		"oracles TestProduct_Oracle": {"lib TestProduct_Lib"},
	}
	for _, test := range []struct {
		name string
		// cycle adds a wait that closes a loop; fails names the package whose run exits 1.
		cycle   bool
		fails   string
		runs    []string
		settled bool
		said    string
	}{
		{"waits in order", false, "", []string{"lib ^(TestProduct_Lib)$", "oracles ^(TestProduct_Oracle)$", "app ^(TestProduct_App)$",
			"app ^(TestProduct_AppLate)$"}, true, "products settled for 3 packages in 4 waves, 4 product tests"},
		{"a cycle refuses", true, "", nil, false, "waits hold a cycle"},
		{"a failing wave stops the waves after it", false, "oracles", []string{"lib ^(TestProduct_Lib)$", "oracles ^(TestProduct_Oracle)$"},
			false, "wave 2 of 4 failed in ./oracles/"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			tree, bin := filepath.Join(directory, "tree"), filepath.Join(directory, "bin")
			for _, made := range []string{filepath.Join(tree, "cloud", "fast-gate"), filepath.Join(tree, "app"), filepath.Join(tree, "lib"),
				filepath.Join(tree, "oracles"), bin} {
				if err := os.MkdirAll(made, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			order := map[string][]string{}
			for name, named := range waits {
				order[name] = named
			}
			if test.cycle {
				order["lib TestProduct_Lib"] = []string{"app TestProduct_AppLate"}
			}
			orderData, err := json.Marshal(map[string]any{"version": 1, "waits": order})
			if err != nil {
				t.Fatal(err)
			}
			// oracles lists a test nothing waits on, which never runs: a package reached by a wait brings only the tests
			// waited on.
			listing := "app TestProduct_App\napp TestProduct_AppLate\nlib TestProduct_Lib\noracles TestProduct_Oracle\noracles TestProduct_OracleUnasked\n"
			for _, file := range [][2]string{
				{filepath.Join(tree, "cloud", "fast-gate", "compiler-dependencies.json"), `{"packages":{"app":["lib"]}}` + "\n"},
				{filepath.Join(tree, "cloud", "fast-gate", "product-order.json"), string(orderData)},
				{filepath.Join(directory, "listing"), listing},
			} {
				if err := os.WriteFile(file[0], []byte(file[1]), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			// The package is go test's last argument, ./<path>/. A listing ends with go's own ok line, as go's does.
			stand := `#!/bin/sh
for argument; do last=$argument; done
package=${last#./}
package=${package%/}
case " $* " in
*" -list "*)
	sed -n "s|^$package ||p" "$GO_LISTING"
	printf 'ok  \texample.invalid/%s\t0.01s\n' "$package"
	exit 0
	;;
esac
while [ $# -gt 0 ]; do
	[ "$1" != -run ] || expression=$2
	shift
done
echo "$package $expression" >> "$GO_RUNS"
[ "$package" != "$GO_FAILS" ]
`
			if err := os.WriteFile(filepath.Join(bin, "go"), []byte(stand), 0o755); err != nil {
				t.Fatal(err)
			}
			// The TypeScript copy is the image's road, not this test's: it stands in as done.
			command := toolchain.Command("bash", "-c", `. "$SANDBOXED_JOB_SCRIPT" || exit 1
sandboxed_typescript_source() { :; }
if sandboxed_products ./app; then echo "settled=ok"; fi
`)
			if command.Err != nil {
				t.Fatal(command.Err)
			}
			command.Dir = tree
			environment := command.Env
			if environment == nil {
				environment = os.Environ()
			}
			path := os.Getenv("PATH")
			kept := make([]string, 0, len(environment)+6)
			for _, variable := range environment {
				name, value, _ := strings.Cut(variable, "=")
				switch name {
				case "PATH":
					path = value
				case "TMPDIR", "ADAMIC_BUILD_CACHE_DIR", "LOOM_JOB_SCRATCH":
				default:
					kept = append(kept, variable)
				}
			}
			runs := filepath.Join(directory, "runs")
			command.Env = append(kept, "PATH="+bin+":"+path, "TMPDIR="+filepath.Join(directory, "tmp"), "SANDBOXED_JOB_SCRIPT="+script,
				"GO_LISTING="+filepath.Join(directory, "listing"), "GO_RUNS="+runs, "GO_FAILS="+test.fails)
			output, err := command.CombinedOutput()
			said := string(output)
			if err != nil {
				t.Fatalf("bash: %v\n%s", err, said)
			}
			if settled := strings.Contains(said, "settled=ok"); settled != test.settled {
				t.Fatalf("want sandboxed_products to return 0 %v, got:\n%s", test.settled, said)
			}
			if !strings.Contains(said, test.said) {
				t.Errorf("want %q said, got:\n%s", test.said, said)
			}
			data, err := os.ReadFile(runs)
			if err != nil && !os.IsNotExist(err) {
				t.Fatal(err)
			}
			var ran []string
			if text := strings.TrimSpace(string(data)); text != "" {
				ran = strings.Split(text, "\n")
			}
			if strings.Join(ran, "\n") != strings.Join(test.runs, "\n") {
				t.Errorf("want the runs\n  %s\ngot\n  %s\n%s", strings.Join(test.runs, "\n  "), strings.Join(ran, "\n  "), said)
			}
		})
	}
}

// sandboxed_git leaves a tree that is already a repository as it is and returns 0, so an `&&` chain goes on to its command:
// the runner lays sandboxed trees with a .git since Nov 16, and the refusal it made then cost Emit's jobs a run each
// (#6zvxhjq). A tree with no .git still becomes one commit holding the tree as laid.
//
// Mutant: the old refusal put back (`[ ! -e .git ] || sandboxedJobFail ... || return 1`), red on the repository case.
func TestSandboxedGitLeavesARepositoryAsItIs(t *testing.T) {
	t.Parallel()
	script, err := filepath.Abs("sandboxed-job.sh")
	if err != nil {
		t.Fatal(err)
	}
	for _, repository := range []bool{true, false} {
		t.Run(map[bool]string{true: "a tree laid with its .git", false: "a tree laid with none"}[repository], func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			tree := filepath.Join(directory, "tree")
			if err := os.MkdirAll(tree, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(tree, "go.mod"), []byte("module example.invalid/tree\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			git := func(arguments ...string) string {
				command := toolchain.Command("git", append([]string{"-C", tree, "-c", "user.name=test", "-c", "user.email=test@loom.invalid", "-c", "commit.gpgsign=false"}, arguments...)...)
				if command.Err != nil {
					t.Fatal(command.Err)
				}
				output, err := command.CombinedOutput()
				if err != nil {
					t.Fatalf("git %v: %v\n%s", arguments, err, output)
				}
				return strings.TrimSpace(string(output))
			}
			head := ""
			if repository {
				git("init", "-q")
				git("add", "-A")
				git("commit", "-q", "-m", "the runner's own commit")
				head = git("rev-parse", "HEAD")
			}
			command := toolchain.Command("bash", "-c", `. "$SANDBOXED_JOB_SCRIPT" || exit 1
sandboxed_git && echo "sandboxed_git=ok"`)
			if command.Err != nil {
				t.Fatal(command.Err)
			}
			command.Dir = tree
			environment := command.Env
			if environment == nil {
				environment = os.Environ()
			}
			kept := make([]string, 0, len(environment)+2)
			for _, variable := range environment {
				name, _, _ := strings.Cut(variable, "=")
				if name != "TMPDIR" && name != "ADAMIC_BUILD_CACHE_DIR" && name != "LOOM_JOB_SCRATCH" {
					kept = append(kept, variable)
				}
			}
			command.Env = append(kept, "TMPDIR="+filepath.Join(directory, "tmp"), "SANDBOXED_JOB_SCRIPT="+script)
			output, err := command.CombinedOutput()
			said := string(output)
			if err != nil || !strings.Contains(said, "sandboxed_git=ok") {
				t.Fatalf("want sandboxed_git to return 0, got %v:\n%s", err, said)
			}
			if repository {
				if now := git("rev-parse", "HEAD"); now != head {
					t.Errorf("the tree's HEAD moved from %s to %s: its history wasn't left alone", head, now)
				}
				if !strings.Contains(said, "already a repository") {
					t.Errorf("want it said that the tree is already a repository, got:\n%s", said)
				}
				return
			}
			if count := git("rev-list", "--count", "HEAD"); count != "1" {
				t.Errorf("want one commit holding the tree as laid, got %s", count)
			}
		})
	}
}

// sandboxed_repositories lays each submodule as a repository at its pin from the files the runner laid and its manifest,
// .tracked/<path>/{HEAD,commit,files}, and proves the pin rather than taking the manifest's word (Operations on
// #52cxm6c): the gitlink in the parent's index, the commit object hashing to it, and the laid files writing its tree.
// The test makes a stand-in tree as the runner lays one: a .git whose index holds sub's gitlink, sub's files and those of
// its own submodule, inner, unpacked from git archive, and the manifest, git's own output for each pin. sub's pin has a
// parent, as cohere's does, a CRLF blob under `* text=auto eol=lf`, as four of cohere's are, an executable file, and a
// tracked file its .gitignore names. A tree laid as it was becomes two repositories at their pins that read clean, named
// or not, sub's index holding inner as the gitlink its pin names, and a second call leaves them as they are. A .tracked naming another commit whole, a commit object forged
// under the pin's name, and a file changed under sub or under inner each refuse.
//
// Mutants: the info/attributes line dropped (red on the trees laid as they were: sub's CRLF blob is added as LF, so its
// tree isn't the pin's); the commit's hash check taken out (`[ "$made" = "$gitlink" ] ||` and its refusal made `:`, red on
// the forged commit, which then refuses only at HEAD, saying something else); the pin read from .tracked/<path>/HEAD
// rather than from the parent's index (`gitlink=$(head -c 64 "$manifest/HEAD")` after the gitlink's refusal, red on the
// .tracked naming another commit, which is then laid); the tree check taken out (`[ "$built" = "$tree" ] ||` and its
// refusal made `:`, red on both changed files, which then refuse only at git status, whose HEAD names a tree the
// repository lacks, saying something else); the gitlinks-first update-index step taken out (red on the trees laid as
// they were, but before the index's assertion: inner's files go in as sub's own, and the tree check refuses first, so no
// mutant that leaves that check reaches the assertion, which holds the shape lay.sh's repository has).
func TestSandboxedRepositoriesProvesEachPin(t *testing.T) {
	t.Parallel()
	script, err := filepath.Abs("sandboxed-job.sh")
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name string
		// paths are sandboxed_repositories' arguments; forge changes the laid tree before it runs, given the tree and
		// the stand-in's commits; refused is what it says refusing, or nil for a tree it lays.
		paths   string
		forge   func(t *testing.T, tree string, pins standInPins)
		refused func(pins standInPins) string
	}{
		{"a tree laid as it was", "", nil, nil},
		{"a tree laid as it was, its submodule named", "sub", nil, nil},
		{"a .tracked naming another commit whole", "", func(t *testing.T, tree string, pins standInPins) {
			writeStandInFile(t, filepath.Join(tree, "sub", "a.txt"), "old\n", 0o644)
			writeTracked(t, tree, "sub", pins.origin, pins.other, pins.other)
		}, func(pins standInPins) string {
			return ".tracked/sub/HEAD names " + pins.other[:12] + ", not sub's gitlink " + pins.sub[:12]
		}},
		{"a commit object forged under the pin's name", "", func(t *testing.T, tree string, pins standInPins) {
			writeStandInFile(t, filepath.Join(tree, "sub", "a.txt"), "old\n", 0o644)
			writeTracked(t, tree, "sub", pins.origin, pins.sub, pins.other)
		}, func(pins standInPins) string {
			return ".tracked/sub/commit hashes to " + pins.other[:12] + ", not sub's gitlink " + pins.sub[:12]
		}},
		{"a file changed under the submodule", "", func(t *testing.T, tree string, pins standInPins) {
			writeStandInFile(t, filepath.Join(tree, "sub", "a.txt"), "a\nchanged\n", 0o644)
		}, func(pins standInPins) string { return "the files laid under sub write the tree" }},
		{"a file changed under the submodule's own submodule", "", func(t *testing.T, tree string, pins standInPins) {
			writeStandInFile(t, filepath.Join(tree, "sub", "inner", "i.txt"), "changed\n", 0o644)
		}, func(pins standInPins) string { return "the files laid under sub/inner write the tree" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			tree, pins := layStandInTree(t, directory)
			if test.forge != nil {
				test.forge(t, tree, pins)
			}
			command := toolchain.Command("bash", "-c", `. "$SANDBOXED_JOB_SCRIPT" || exit 1
sandboxed_repositories $SANDBOXED_REPOSITORIES_PATHS || exit 0
echo "laid=ok"
sandboxed_repositories $SANDBOXED_REPOSITORIES_PATHS && echo "again=ok"
`)
			if command.Err != nil {
				t.Fatal(command.Err)
			}
			command.Dir = tree
			environment := command.Env
			if environment == nil {
				environment = os.Environ()
			}
			kept := make([]string, 0, len(environment)+3)
			for _, variable := range environment {
				name, _, _ := strings.Cut(variable, "=")
				switch name {
				case "TMPDIR", "ADAMIC_BUILD_CACHE_DIR", "LOOM_JOB_SCRATCH", "GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE":
				default:
					kept = append(kept, variable)
				}
			}
			command.Env = append(kept, "TMPDIR="+filepath.Join(directory, "tmp"), "SANDBOXED_JOB_SCRIPT="+script,
				"SANDBOXED_REPOSITORIES_PATHS="+test.paths)
			output, err := command.CombinedOutput()
			said := string(output)
			if err != nil {
				t.Fatalf("bash: %v\n%s", err, said)
			}
			if test.refused != nil {
				if strings.Contains(said, "laid=ok") {
					t.Fatalf("want sandboxed_repositories to refuse, got it laid:\n%s", said)
				}
				if want := test.refused(pins); !strings.Contains(said, want) {
					t.Errorf("want %q said, got:\n%s", want, said)
				}
				return
			}
			if !strings.Contains(said, "laid=ok") || !strings.Contains(said, "again=ok") {
				t.Fatalf("want sandboxed_repositories to lay the tree and then leave it, got:\n%s", said)
			}
			for _, want := range []string{"submodules at their pins: 2 laid as repositories from their files, 0 already one",
				"submodules at their pins: 0 laid as repositories from their files, 2 already one",
				"sub " + pins.sub[:12] + " is its own repository at the pin, left as it is"} {
				if !strings.Contains(said, want) {
					t.Errorf("want %q said, got:\n%s", want, said)
				}
			}
			for path, want := range map[string]string{"sub": pins.sub, "sub/inner": pins.inner} {
				if head := standInGit(t, filepath.Join(tree, path), "rev-parse", "HEAD"); head != want {
					t.Errorf("%s's HEAD is %s, want its pin %s", path, head, want)
				}
				if count := standInGit(t, filepath.Join(tree, path), "rev-list", "--count", "HEAD"); count != "1" {
					t.Errorf("%s holds %s commits, want its pin alone", path, count)
				}
			}
			// sub's index holds inner as the gitlink its pin names, never inner's files as sub's own
			if entry := standInGit(t, filepath.Join(tree, "sub"), "ls-files", "-s", "inner"); entry != "160000 "+pins.inner+" 0\tinner" {
				t.Errorf("sub's index holds inner as %q, want the gitlink 160000 %s", entry, pins.inner)
			}
			// The tree reads its submodule at its gitlink, and sub reads inner at its own
			for _, path := range []string{".", "sub", "sub/inner"} {
				if status := standInGit(t, filepath.Join(tree, path), "status", "--porcelain"); status != "" {
					t.Errorf("%s doesn't read clean:\n%s", path, status)
				}
			}
		})
	}
}

// sandboxed_repositories on a tree the runner's lay didn't lay says it's the box's gap, not the head's red (the trust
// seat's ruling on #cq8rn2s): a job whose command didn't open with its sealed `source /opt/adamic-block/environment.sh`
// reached a house box's keeper, which laid a git worktree with no submodules and no .tracked, and the refusal that the
// manifest wasn't there read as the head's exit 1. The test lays layStandInTree's tree and takes its manifest away: gone
// with sub an empty directory, as a keeper's worktree holds it, or a link to the manifest moved aside, which is no
// manifest either. Each exits 3 with one `GAP sandboxed:` line naming the sealed opening, through a caller's `|| exit 1`,
// and nothing is written under sub. A tree with no .tracked whose submodules are already repositories at their pins, as
// lay.sh's --repositories leaves them, still passes.
//
// Mutant: the gap's `exit 3` made `return 1` (red on both gaps: the caller's `|| exit 1` ends the job with 1).
func TestSandboxedRepositoriesSaysAKeeperLaidTreeIsTheBoxsGap(t *testing.T) {
	t.Parallel()
	script, err := filepath.Abs("sandboxed-job.sh")
	if err != nil {
		t.Fatal(err)
	}
	const gap = "GAP sandboxed: this tree carries no runner's manifest (.tracked)"
	for _, test := range []struct {
		name string
		// forge changes the laid tree before the script runs; laidFirst lays it with its manifest and then moves the
		// manifest aside, in the script, before the call the test reads; exit is the script's exit and said a line it says.
		forge     func(t *testing.T, directory, tree string)
		laidFirst bool
		exit      int
		said      string
	}{
		{"a keeper's worktree, sub an empty directory and no .tracked", func(t *testing.T, directory, tree string) {
			for _, gone := range []string{filepath.Join(tree, ".tracked"), filepath.Join(tree, "sub")} {
				if err := os.RemoveAll(gone); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Mkdir(filepath.Join(tree, "sub"), 0o755); err != nil {
				t.Fatal(err)
			}
		}, false, 3, gap},
		{"a .tracked that is a link to the manifest", func(t *testing.T, directory, tree string) {
			aside := filepath.Join(directory, "manifest")
			if err := os.Rename(filepath.Join(tree, ".tracked"), aside); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(aside, filepath.Join(tree, ".tracked")); err != nil {
				t.Fatal(err)
			}
		}, false, 3, gap},
		{"no .tracked, its submodules already repositories at their pins", nil, true, 0,
			"submodules at their pins: 0 laid as repositories from their files, 2 already one"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			tree, _ := layStandInTree(t, directory)
			if test.forge != nil {
				test.forge(t, directory, tree)
			}
			// The `|| exit 1` is a caller's chain: only an exit inside the call ends the job with 3
			command := toolchain.Command("bash", "-c", `. "$SANDBOXED_JOB_SCRIPT" || exit 1
if [ -n "$SANDBOXED_REPOSITORIES_LAID_FIRST" ]; then
	sandboxed_repositories || exit 1
	mv .tracked "$TMPDIR/runner-manifest" || exit 1
fi
sandboxed_repositories || exit 1
echo "returned=ok"
`)
			if command.Err != nil {
				t.Fatal(command.Err)
			}
			command.Dir = tree
			environment := command.Env
			if environment == nil {
				environment = os.Environ()
			}
			kept := make([]string, 0, len(environment)+3)
			for _, variable := range environment {
				name, _, _ := strings.Cut(variable, "=")
				switch name {
				case "TMPDIR", "ADAMIC_BUILD_CACHE_DIR", "LOOM_JOB_SCRATCH", "GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE":
				default:
					kept = append(kept, variable)
				}
			}
			laidFirst := ""
			if test.laidFirst {
				laidFirst = "yes"
			}
			command.Env = append(kept, "TMPDIR="+filepath.Join(directory, "tmp"), "SANDBOXED_JOB_SCRIPT="+script,
				"SANDBOXED_REPOSITORIES_LAID_FIRST="+laidFirst)
			output, err := command.CombinedOutput()
			said := string(output)
			exit := 0
			if err != nil {
				var exitError *exec.ExitError
				if !errors.As(err, &exitError) {
					t.Fatalf("bash: %v\n%s", err, said)
				}
				exit = exitError.ExitCode()
			}
			if exit != test.exit {
				t.Fatalf("want the script to exit %d, got %d:\n%s", test.exit, exit, said)
			}
			if test.exit == 0 {
				if !strings.Contains(said, "returned=ok") || !strings.Contains(said, test.said) {
					t.Errorf("want sandboxed_repositories to return 0 saying %q, got:\n%s", test.said, said)
				}
				return
			}
			var gapLines []string
			for _, line := range strings.Split(said, "\n") {
				if strings.HasPrefix(line, "GAP sandboxed:") {
					gapLines = append(gapLines, line)
				}
			}
			if len(gapLines) != 1 || !strings.HasPrefix(gapLines[0], test.said) {
				t.Fatalf("want one line starting %q, got:\n%s", test.said, said)
			}
			for _, want := range []string{"sub's pin can't be proved here", "'source /opt/adamic-block/environment.sh && ...'"} {
				if !strings.Contains(gapLines[0], want) {
					t.Errorf("want the GAP line to say %q, got:\n%s", want, gapLines[0])
				}
			}
			if strings.Contains(said, "returned=ok") {
				t.Errorf("want nothing after the gap, got:\n%s", said)
			}
			if _, err := os.Lstat(filepath.Join(tree, "sub", ".git")); !os.IsNotExist(err) {
				t.Errorf("want nothing written under sub, but its .git reads as there or unreadable (%v)", err)
			}
		})
	}
}

// standInPins are the commits of layStandInTree's stand-in: sub's pin, the commit before it (other), inner's pin, and
// the repository sub's commits are in.
type standInPins struct {
	sub, other, inner, origin string
}

// layStandInTree makes, under directory, a tree as the runner lays a sandboxed job's: its own .git, a commit whose
// index holds the gitlink sub, the files of sub's pin and of inner's inside it unpacked from git archive, and
// .tracked/<path>/{HEAD,commit,files} for each. sub's pin is its second commit, which differs from its first (other)
// in a.txt alone.
func layStandInTree(t *testing.T, directory string) (string, standInPins) {
	t.Helper()
	inner, origin, tree := filepath.Join(directory, "inner"), filepath.Join(directory, "origin"), filepath.Join(directory, "tree")
	for _, made := range []string{inner, origin, tree} {
		if err := os.MkdirAll(made, 0o755); err != nil {
			t.Fatal(err)
		}
		standInGit(t, made, "init", "-q")
	}
	writeStandInFile(t, filepath.Join(inner, "i.txt"), "i\n", 0o644)
	standInGit(t, inner, "add", "-A")
	standInGit(t, inner, "commit", "-q", "-m", "inner")
	pins := standInPins{inner: standInGit(t, inner, "rev-parse", "HEAD"), origin: origin}
	// Each blob as its bytes are, so crlf.txt's CRLF is the blob's, as cohere's four are; the runner's mirror has no
	// such file, so it goes before the archive
	attributes := filepath.Join(origin, ".git", "info", "attributes")
	writeStandInFile(t, attributes, "* -text\n", 0o644)
	for name, file := range map[string]struct {
		content string
		mode    os.FileMode
	}{
		"a.txt": {"old\n", 0o644}, "run.sh": {"#!/bin/sh\n", 0o755}, "crlf.txt": {"one\r\ntwo\r\n", 0o644},
		".gitattributes": {"* text=auto eol=lf\n", 0o644}, ".gitignore": {"ignored.txt\n", 0o644}, "ignored.txt": {"kept\n", 0o644},
	} {
		writeStandInFile(t, filepath.Join(origin, name), file.content, file.mode)
	}
	standInGit(t, origin, "add", "-A", "-f")
	standInGit(t, origin, "update-index", "--add", "--cacheinfo", "160000,"+pins.inner+",inner")
	standInGit(t, origin, "commit", "-q", "-m", "other")
	pins.other = standInGit(t, origin, "rev-parse", "HEAD")
	// a.txt alone is added: commit -a would take inner, a gitlink with no directory here, as gone
	writeStandInFile(t, filepath.Join(origin, "a.txt"), "a\n", 0o644)
	standInGit(t, origin, "add", "a.txt")
	standInGit(t, origin, "commit", "-q", "-m", "the pin")
	pins.sub = standInGit(t, origin, "rev-parse", "HEAD")
	if err := os.Remove(attributes); err != nil {
		t.Fatal(err)
	}
	// The tree's own commit, its index read from it, and .tracked out of its git status, as the runner's tipExclude has it
	writeStandInFile(t, filepath.Join(tree, "go.mod"), "module example.invalid/tree\n", 0o644)
	writeStandInFile(t, filepath.Join(tree, ".gitmodules"), "[submodule \"sub\"]\n\tpath = sub\n\turl = ../origin.git\n", 0o644)
	standInGit(t, tree, "add", "go.mod", ".gitmodules")
	standInGit(t, tree, "update-index", "--add", "--cacheinfo", "160000,"+pins.sub+",sub")
	standInGit(t, tree, "commit", "-q", "-m", "the tree")
	writeStandInFile(t, filepath.Join(tree, ".git", "info", "exclude"), "/.tracked/\n", 0o644)
	unpackStandInArchive(t, origin, pins.sub, filepath.Join(tree, "sub"))
	unpackStandInArchive(t, inner, pins.inner, filepath.Join(tree, "sub", "inner"))
	writeTracked(t, tree, "sub", origin, pins.sub, pins.sub)
	writeTracked(t, tree, "sub/inner", inner, pins.inner, pins.inner)
	return tree, pins
}

// writeTracked writes .tracked/<path> in tree as the runner's writeManifest does, each file git's own output byte for
// byte from repository: HEAD naming head, and the commit object and listing of commit.
func writeTracked(t *testing.T, tree, path, repository, head, commit string) {
	t.Helper()
	at := filepath.Join(tree, ".tracked", filepath.FromSlash(path))
	for name, content := range map[string][]byte{"HEAD": []byte(head + "\n"),
		"commit": standInGitBytes(t, repository, "cat-file", "commit", commit),
		"files":  standInGitBytes(t, repository, "ls-tree", "-r", "-z", "--full-tree", commit)} {
		if err := os.MkdirAll(at, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(at, name), content, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// unpackStandInArchive unpacks `git archive --format=tar <commit>` of repository into directory, as the runner's
// ExecStartPre unpacks a submodule's archive: each directory, file and link, a file executable where its mode says so.
// The archive holds no member for its root, and needn't hold one for a member's directory, so each is made here.
func unpackStandInArchive(t *testing.T, repository, commit, directory string) {
	t.Helper()
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	reader := tar.NewReader(bytes.NewReader(standInGitBytes(t, repository, "archive", "--format=tar", commit)))
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(directory, filepath.FromSlash(header.Name))
		switch header.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(target, 0o755)
		case tar.TypeReg:
			var content []byte
			if content, err = io.ReadAll(reader); err == nil {
				if err = os.MkdirAll(filepath.Dir(target), 0o755); err == nil {
					err = os.WriteFile(target, content, os.FileMode(header.Mode)&0o777)
				}
			}
		case tar.TypeSymlink:
			if err = os.MkdirAll(filepath.Dir(target), 0o755); err == nil {
				err = os.Symlink(header.Linkname, target)
			}
		case tar.TypeXGlobalHeader:
			// git archive's pax header, naming the commit, is no member to lay
		}
		if err != nil {
			t.Fatalf("unpacking %s of %s at %s: %v", header.Name, repository, commit, err)
		}
	}
}

// writeStandInFile writes one file of the stand-in, its directory made first.
func writeStandInFile(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
	// WriteFile keeps the mode of a file that was there, so an executable bit is set as asked
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
}

// standInGit is git in directory, its configuration out of reach and a test's own identity, its output trimmed.
func standInGit(t *testing.T, directory string, arguments ...string) string {
	t.Helper()
	return strings.TrimSpace(string(standInGitBytes(t, directory, arguments...)))
}

// standInGitBytes is standInGit's output byte for byte, as the runner's manifest holds git's.
func standInGitBytes(t *testing.T, directory string, arguments ...string) []byte {
	t.Helper()
	command := toolchain.GitCommandIn(directory, append([]string{"-c", "user.name=test", "-c", "user.email=test@loom.invalid",
		"-c", "commit.gpgsign=false", "-c", "init.defaultBranch=main"}, arguments...)...)
	if command.Err != nil {
		t.Fatal(command.Err)
	}
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", arguments, directory, err, stderr.String())
	}
	return output
}

// sandboxed-job.sh takes the node's laid Go module cache only when loom-node's cache proxy laid this very commit whole
// (#nodecache): a lay file for the runner's repository and commit, naming the laid cache at its one path, whose goSums
// are the tree's go.sum and go.work.sum files exactly, each with its sha256. The lay is found by the name the cache
// gives it, the hex sha256 of the repository, a newline and the commit, in a lays/ of mode 0711, and a file at the old
// per-repository path is never read. Every other case keeps the image's GOMODCACHE and says why,
// a lay that doesn't parse included, and the preamble still returns 0. GOPROXY, GOSUMDB and GOFLAGS are the image's in
// every case. The test sources the script in a temporary tree with two go.sum files (one in a submodule) and a root
// go.work.sum, and decoys under the .git directories the walk skips, beside a temporary lays directory and laid cache.
//
// Mutants: the no-lay branch taking the laid cache (`image('no lay for this commit')` made `sys.exit(print('laid 214'))`,
// red on no lay); the sha256 check taken out (its `if hashlib.sha256(...) != named[path]:` made `if False:`, red on a
// go.sum that differs); the set check taken out (`found ^ set(named)` made `set()`, red on a go.sum the lay doesn't
// name); the walk finding go.sum alone (`name in ('go.sum', 'go.work.sum')` made `name == 'go.sum'`, red on the whole
// lay, whose go.work.sum the tree then seems to lack); the lay looked for at the old per-repository path
// (`os.path.join(lays, repository, commit + '.json')`, red on the whole lay and on the old path's).
func TestSandboxedJobTakesTheNodesLaidModulesOnlyWhole(t *testing.T) {
	t.Parallel()
	script, err := filepath.Abs("sandboxed-job.sh")
	if err != nil {
		t.Fatal(err)
	}
	const commit = "0123456789abcdef0123456789abcdef01234567"
	const goFlags = "-buildvcs=false -trimpath -mod=readonly"
	sums := map[string]string{"go.sum": "example.invalid/a v1.0.0 h1:a=\n", "cohere/go.sum": "example.invalid/b v1.0.0 h1:b=\n",
		"go.work.sum": "example.invalid/w v1.0.0 h1:w=\n"}
	for _, test := range []struct {
		name string
		// lay returns the lay file's bytes, or nil for no file; tree changes the tree after its go.sum files are written.
		lay  func(t *testing.T, laid string, goSums []map[string]string) []byte
		tree func(t *testing.T, tree string)
		laid bool
		why  string
	}{
		{"a lay for this commit, whole", nil, nil, true, ""},
		{"no lay for this commit", func(*testing.T, string, []map[string]string) []byte { return nil }, nil, false, "no lay for this commit"},
		{"a lay naming another commit", func(t *testing.T, laid string, goSums []map[string]string) []byte {
			return layOf(t, "adamic", "fedcba9876543210fedcba9876543210fedcba98", laid, goSums)
		}, nil, false, "the lay names another commit"},
		{"a lay naming another repository", func(t *testing.T, laid string, goSums []map[string]string) []byte {
			return layOf(t, "cohere", commit, laid, goSums)
		}, nil, false, "the lay names another repository"},
		{"a lay naming another cache", func(t *testing.T, laid string, goSums []map[string]string) []byte {
			return layOf(t, "adamic", commit, laid+"-elsewhere", goSums)
		}, nil, false, "the lay names the cache"},
		{"a go.sum that differs", nil, func(t *testing.T, tree string) {
			if err := os.WriteFile(filepath.Join(tree, "cohere", "go.sum"), []byte("example.invalid/b v1.0.1 h1:c=\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}, false, "cohere/go.sum differs"},
		{"a go.work.sum that differs", nil, func(t *testing.T, tree string) {
			if err := os.WriteFile(filepath.Join(tree, "go.work.sum"), []byte("example.invalid/w v1.0.1 h1:x=\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}, false, "go.work.sum differs"},
		{"a go.sum the lay doesn't name", nil, func(t *testing.T, tree string) {
			if err := os.MkdirAll(filepath.Join(tree, "tools"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(tree, "tools", "go.sum"), []byte("example.invalid/c v1.0.0 h1:c=\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}, false, "tools/go.sum is in the tree alone"},
		{"a go.work.sum the lay doesn't name", nil, func(t *testing.T, tree string) {
			if err := os.WriteFile(filepath.Join(tree, "cohere", "go.work.sum"), []byte("example.invalid/v v1.0.0 h1:v=\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}, false, "cohere/go.work.sum is in the tree alone"},
		{"a go.sum the tree lacks", nil, func(t *testing.T, tree string) {
			if err := os.Remove(filepath.Join(tree, "cohere", "go.sum")); err != nil {
				t.Fatal(err)
			}
		}, false, "cohere/go.sum is in the lay alone"},
		{"a lay that doesn't parse", func(*testing.T, string, []map[string]string) []byte { return []byte(`{"repository":"adamic",`) }, nil, false, "the lay doesn't parse"},
		// A file at the per-repository path the cache wrote before its name moved, lays/adamic/<commit>.json, and none at
		// the name: were the old path read, its why would be the stand-in's, never "no lay"
		{"a lay at the old per-repository path, which the cache no longer writes", func(*testing.T, string, []map[string]string) []byte { return nil }, func(t *testing.T, tree string) {
			old := filepath.Join(filepath.Dir(tree), "lays", "adamic")
			if err := os.Mkdir(old, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(old, commit+".json"), []byte(`{"repository":"adamic",`), 0o444); err != nil {
				t.Fatal(err)
			}
		}, false, "no lay for this commit"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			tree, lays, laid, image := filepath.Join(directory, "tree"), filepath.Join(directory, "lays"), filepath.Join(directory, "mod"), filepath.Join(directory, "image-mod")
			for _, made := range []string{filepath.Join(tree, "cohere"), filepath.Join(tree, ".git", "modules", "cohere"), filepath.Join(tree, "cohere", ".git"),
				laid, image} {
				if err := os.MkdirAll(made, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			// lays/ as the cache makes it, 0711: a job opens its own lay by name and lists nothing
			if err := os.Mkdir(lays, 0o711); err != nil {
				t.Fatal(err)
			}
			nameDigest := sha256.Sum256([]byte("adamic\n" + commit))
			layPath := filepath.Join(lays, hex.EncodeToString(nameDigest[:])+".json")
			goSums := []map[string]string{}
			for _, path := range []string{"cohere/go.sum", "go.sum", "go.work.sum"} {
				if err := os.WriteFile(filepath.Join(tree, path), []byte(sums[path]), 0o644); err != nil {
					t.Fatal(err)
				}
				digest := sha256.Sum256([]byte(sums[path]))
				goSums = append(goSums, map[string]string{"path": path, "sha256": hex.EncodeToString(digest[:])})
			}
			// Under the root's .git and a submodule's, where the walk never looks: go.sum files the lay doesn't name, so
			// reading either would refuse the lay.
			for _, decoy := range []string{filepath.Join(tree, ".git", "modules", "cohere", "go.sum"), filepath.Join(tree, "cohere", ".git", "go.sum")} {
				if err := os.WriteFile(decoy, []byte("not the tree's\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if test.tree != nil {
				test.tree(t, tree)
			}
			lay := layOf(t, "adamic", commit, laid, goSums)
			if test.lay != nil {
				lay = test.lay(t, laid, goSums)
			}
			if lay != nil {
				if err := os.WriteFile(layPath, lay, 0o444); err != nil {
					t.Fatal(err)
				}
			}
			command := toolchain.Command("bash", "-c", `. "$SANDBOXED_JOB_SCRIPT" || { echo "sourced=failed"; exit 0; }
echo "sourced=ok"
printf 'gomodcache=%s\ngoproxy=%s\ngosumdb=%s\ngoflags=%s\n' "$GOMODCACHE" "$GOPROXY" "$GOSUMDB" "$GOFLAGS"
`)
			if command.Err != nil {
				t.Fatal(command.Err)
			}
			command.Dir = tree
			environment := command.Env
			if environment == nil {
				environment = os.Environ()
			}
			kept := make([]string, 0, len(environment)+10)
			for _, variable := range environment {
				name, _, _ := strings.Cut(variable, "=")
				switch name {
				case "TMPDIR", "ADAMIC_BUILD_CACHE_DIR", "LOOM_JOB_SCRATCH", "LOOM_JOB_REPOSITORY", "LOOM_JOB_COMMIT", "SANDBOXED_JOB_LAYS",
					"SANDBOXED_JOB_LAID_GOMODCACHE", "GOMODCACHE", "GOPROXY", "GOSUMDB", "GOFLAGS":
				default:
					kept = append(kept, variable)
				}
			}
			command.Env = append(kept, "SANDBOXED_JOB_SCRIPT="+script, "TMPDIR="+filepath.Join(directory, "tmp"), "LOOM_JOB_REPOSITORY=adamic",
				"LOOM_JOB_COMMIT="+commit, "SANDBOXED_JOB_LAYS="+lays, "SANDBOXED_JOB_LAID_GOMODCACHE="+laid, "GOMODCACHE="+image,
				"GOPROXY=off", "GOSUMDB=off", "GOFLAGS="+goFlags)
			output, err := command.CombinedOutput()
			said := string(output)
			if err != nil || !strings.Contains(said, "sourced=ok") {
				t.Fatalf("want the source to pass, got %v:\n%s", err, said)
			}
			values := map[string]string{}
			for _, line := range strings.Split(said, "\n") {
				if name, value, ok := strings.Cut(line, "="); ok {
					values[name] = value
				}
			}
			for name, want := range map[string]string{"goproxy": "off", "gosumdb": "off", "goflags": goFlags} {
				if values[name] != want {
					t.Errorf("%s is %q, want the image's %q", name, values[name], want)
				}
			}
			if test.laid {
				if values["gomodcache"] != laid {
					t.Errorf("GOMODCACHE is %q, want the node's laid cache %s", values["gomodcache"], laid)
				}
				if want := "sandboxed-job: go modules: the node's laid cache for " + commit[:12] + " (214 modules)"; !strings.Contains(said, want) {
					t.Errorf("want %q said, got:\n%s", want, said)
				}
				return
			}
			if values["gomodcache"] != image {
				t.Errorf("GOMODCACHE is %q, want the image's %s", values["gomodcache"], image)
			}
			if want := "sandboxed-job: go modules: the image's cache (" + test.why; !strings.Contains(said, want) {
				t.Errorf("want %q said, got:\n%s", want, said)
			}
		})
	}
}

// layOf is a lay file as loom-node's cache proxy writes one.
func layOf(t *testing.T, repository, commit, gomodcache string, goSums []map[string]string) []byte {
	t.Helper()
	data, err := json.Marshal(map[string]any{"repository": repository, "commit": commit, "gomodcache": gomodcache, "goSums": goSums,
		"modules": 214, "laidAt": 1763337600000})
	if err != nil {
		t.Fatal(err)
	}
	return data
}
