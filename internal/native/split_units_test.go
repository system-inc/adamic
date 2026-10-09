package native

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

const decodePrefixes = 92014
const decodePrefixUnitSize = 2048
const normalizePointUnitSize = 65536

type testUnitRange struct{ first, last int }

func (piece testUnitRange) name() string {
	return fmt.Sprintf("cases_%06d_%06d", piece.first, piece.last)
}
func unitRanges(total, width int) []testUnitRange {
	var pieces []testUnitRange
	for first := 0; first < total; first += width {
		pieces = append(pieces, testUnitRange{first, min(first+width, total)})
	}
	return pieces
}

// Frozen counts describe the old workloads. A missing, duplicated or overlapping
// piece fails independently of whether all the remaining observations agree.
func TestSplitUnitCoverage(t *testing.T) {
	for _, plan := range []struct {
		name                string
		total, width, units int
	}{
		{"decode_prefixes", 92014, decodePrefixUnitSize, 45},
		{"normalize_points", 0x110000, normalizePointUnitSize, 17},
	} {
		t.Run(plan.name, func(t *testing.T) {
			pieces := unitRanges(plan.total, plan.width)
			if len(pieces) != plan.units {
				t.Fatalf("got %d units, want %d", len(pieces), plan.units)
			}
			next := 0
			for _, piece := range pieces {
				if piece.first != next || piece.last <= piece.first || piece.last > plan.total {
					t.Fatalf("invalid coverage at %+v after %d", piece, next)
				}
				next = piece.last
			}
			if next != plan.total {
				t.Fatalf("covered %d pieces, want %d", next, plan.total)
			}
		})
	}
	// Verify the actual top-level wrappers, including their helper and range index.
	// This checks gate discovery and catches a missing, duplicated or wrong wrapper.
	for _, plan := range []struct {
		file, prefix, helper, target string
		units                        int
	}{
		{"decode_ascii_test.go", "TestDecodeASCIIUnit", "runDecodeASCIIUnit", "native", 45},
		{"decode_ascii_test.go", "TestDecodeASCIIWASIUnit", "runDecodeASCIIUnit", "wasi", 45},
		{"normalize_test.go", "TestNormalizeMatchesNodePoints", "runNormalizeUnit", "", 17},
		{"normalize_test.go", "TestNormalizeMatchesNodeContexts", "runNormalizeUnit", "", 1},
	} {
		file, err := parser.ParseFile(token.NewFileSet(), plan.file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		seen := make([]bool, plan.units)
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || !strings.HasPrefix(function.Name.Name, plan.prefix) {
				continue
			}
			index := 0
			if plan.units != 1 {
				index, err = strconv.Atoi(strings.TrimPrefix(function.Name.Name, plan.prefix))
				if err != nil || index < 0 || index >= plan.units || function.Name.Name != fmt.Sprintf("%s%02d", plan.prefix, index) {
					t.Fatalf("unexpected unit %s", function.Name.Name)
				}
			}
			if seen[index] || len(function.Body.List) != 1 {
				t.Fatalf("invalid wrapper %s", function.Name.Name)
			}
			seen[index] = true
			statement, ok := function.Body.List[0].(*ast.ExprStmt)
			if !ok {
				t.Fatalf("invalid wrapper %s", function.Name.Name)
			}
			call, ok := statement.X.(*ast.CallExpr)
			if !ok {
				t.Fatalf("invalid call %s", function.Name.Name)
			}
			helper, ok := call.Fun.(*ast.Ident)
			if !ok || helper.Name != plan.helper {
				t.Fatalf("wrong helper in %s", function.Name.Name)
			}
			wantIndex := index
			if plan.units == 1 {
				wantIndex = 17
			}
			wantArguments := 2
			if plan.target != "" {
				wantArguments = 3
			}
			if len(call.Args) != wantArguments {
				t.Fatalf("wrong arguments in %s", function.Name.Name)
			}
			actual, ok := call.Args[len(call.Args)-1].(*ast.BasicLit)
			if !ok || actual.Value != strconv.Itoa(wantIndex) {
				t.Fatalf("wrong range in %s", function.Name.Name)
			}
			if plan.target != "" {
				target, ok := call.Args[1].(*ast.BasicLit)
				if !ok || target.Value != strconv.Quote(plan.target) {
					t.Fatalf("wrong target in %s", function.Name.Name)
				}
			}
		}
		for index, present := range seen {
			if !present {
				t.Errorf("missing top-level %s unit %d", plan.prefix, index)
			}
		}
	}
}

var testExecutable struct {
	sync.Once
	digest string
	err    error
}

type testBuildCache struct{ inputs buildcache.Inputs }

func newTestBuildCache(repository string, directories ...string) (*testBuildCache, error) {
	testExecutable.Do(func() {
		path, err := os.Executable()
		if err != nil {
			testExecutable.err = err
			return
		}
		file, err := os.Open(path)
		if err != nil {
			testExecutable.err = err
			return
		}
		defer file.Close()
		h := sha256.New()
		_, testExecutable.err = io.Copy(h, file)
		testExecutable.digest = fmt.Sprintf("%x", h.Sum(nil))
	})
	if testExecutable.err != nil {
		return nil, testExecutable.err
	}
	in := buildcache.Inputs{Name: "native tests", Files: append([]string{"go.mod", "cmd/adamic"}, directories...), Flags: []string{"test=" + testExecutable.digest}, Toolchain: []string{buildcache.Tool("go", "version"), buildcache.Tool("clang", "--version"), buildcache.Tool("node", "--version")}}
	for _, name := range []string{"GOFLAGS", "CGO_CFLAGS", "CGO_LDFLAGS", "CC", "GOTOOLCHAIN"} {
		in.Flags = append(in.Flags, name+"="+os.Getenv(name))
	}
	return &testBuildCache{in}, nil
}
func (c *testBuildCache) Tree(label string, inputs [][]byte, build func(string) error) (string, error) {
	in := c.inputs
	in.Name = label
	in.Flags = append([]string(nil), c.inputs.Flags...)
	for index, input := range inputs {
		in.Flags = append(in.Flags, fmt.Sprintf("input[%d]=%x", index, sha256.Sum256(input)))
	}
	return buildcache.Get(in, build)
}

// Command caches compiler output only. It refuses commands without a single -o.
// Explicit input files and Go overlays contribute their bytes, not temporary paths.
func (c *testBuildCache) Command(command *exec.Cmd, scratch string, extraInputs ...[]byte) ([]byte, error) {
	outputIndex := -1
	for i, arg := range command.Args {
		if arg == "-o" && i+1 < len(command.Args) {
			if outputIndex != -1 {
				return nil, fmt.Errorf("multiple outputs")
			}
			outputIndex = i + 1
		}
	}
	if outputIndex < 0 {
		return nil, fmt.Errorf("build command has no output")
	}
	output := command.Args[outputIndex]
	inputs := append([][]byte(nil), extraInputs...)
	for i, arg := range command.Args {
		if i == outputIndex {
			continue
		}
		normalized := strings.ReplaceAll(arg, scratch, "$SCRATCH")
		inputs = append(inputs, []byte(normalized))
		if data, err := os.ReadFile(arg); err == nil {
			if strings.HasSuffix(arg, ".json") {
				var overlay struct{ Replace map[string]string }
				if json.Unmarshal(data, &overlay) == nil && len(overlay.Replace) > 0 {

					data = nil
					originals := make([]string, 0, len(overlay.Replace))
					for original := range overlay.Replace {
						originals = append(originals, original)
					}
					sort.Strings(originals)
					for _, original := range originals {
						contents, err := os.ReadFile(overlay.Replace[original])
						if err != nil {
							return nil, err
						}
						inputs = append(inputs, []byte(original), contents)
					}
				}
			}
			inputs = append(inputs, data)
		}
	}
	// Command.Env only differs for the sanitized archive; do not persist environment values.
	for _, variable := range command.Env {
		for _, name := range []string{"CC=", "CGO_CFLAGS=", "CGO_LDFLAGS="} {
			if strings.HasPrefix(variable, name) {
				inputs = append(inputs, []byte(variable))
			}
		}
	}
	name := filepath.Base(output)
	entry, err := c.Tree(name, inputs, func(directory string) error {
		command.Args[outputIndex] = filepath.Join(directory, name)
		defer func() { command.Args[outputIndex] = output }()
		data, err := command.CombinedOutput()
		os.WriteFile(filepath.Join(directory, "build.log"), data, 0644)
		if err != nil {
			return fmt.Errorf("%s: %w\n%s", command.Args, err, data)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(entry, name))
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(output, data, 0755); err != nil {
		return nil, err
	}
	return os.ReadFile(filepath.Join(entry, "build.log"))
}

// Unit timing is diagnostic unless the reference audit explicitly enables it.
func checkGrainBudget(t *testing.T) func() {
	t.Helper()
	started := time.Now()
	return func() {
		elapsed := time.Since(started)
		t.Logf("unit elapsed %.3fs", elapsed.Seconds())
		if os.Getenv("ADAMIC_UNIT_BUDGET") == "1" && elapsed > 60*time.Second {
			t.Errorf("unit over 60s budget: %s", elapsed)
		}
	}
}
