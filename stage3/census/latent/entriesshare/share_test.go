// Package entriesshare holds the latent census's record answers to one computation a file (#wd9wpyk): latentRun and
// latentFullSelected start each file with a share of its own and make every lowering of the file with the share's
// literal memo as its entriesRecords, and lower.go.txt's entriesProgramRecordsWalk answers every lowering of a file
// after the first from the share.
package entriesshare

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/toolchain"
)

// Each file loop of the census's two lowering-per-unit drivers starts the file with a fresh share before it makes the
// file's first lowering, and makes each lowering with the share's literal memo. Mutants: latentRun's share line
// removed, so a file reads the last file's answers, and latentFullSelected's lowering made without entriesRecords, so
// every unit computes its literals' answers again; each is red.
func TestEachFileStartsAFreshShare(t *testing.T) {
	t.Parallel()
	for _, driver := range []struct{ file, function string }{{"../lower.go.txt", "latentRun"}, {"../units.go.txt", "latentFullSelected"}} {
		source, err := os.ReadFile(driver.file)
		if err != nil {
			t.Fatal(err)
		}
		if problem := freshSharePerFile(source, driver.function); problem != "" {
			t.Errorf("%s's %s: %s", filepath.Base(driver.file), driver.function, problem)
		}
	}
}

// freshSharePerFile says what is wrong with function's loop over program.Files(), or "" when the loop's body sets
// latentEntriesShare to latentEntriesFile() as one of its own statements, before any statement that lowers, and every
// lowering it makes is given the share's literals.
func freshSharePerFile(source []byte, function string) string {
	files := token.NewFileSet()
	file, err := parser.ParseFile(files, "driver.go", source, 0)
	if err != nil {
		return err.Error()
	}
	var body *ast.BlockStmt
	for _, declaration := range file.Decls {
		if candidate, ok := declaration.(*ast.FuncDecl); ok && candidate.Name.Name == function && candidate.Recv == nil {
			body = candidate.Body
		}
	}
	if body == nil {
		return "no such function"
	}
	loops := 0
	problem := ""
	ast.Inspect(body, func(node ast.Node) bool {
		loop, ok := node.(*ast.RangeStmt)
		if !ok || render(files, loop.X) != "program.Files()" {
			return true
		}
		loops++
		for _, statement := range loop.Body.List {
			text := render(files, statement)
			if text == "latentEntriesShare = latentEntriesFile()" {
				return false
			}
			if strings.Contains(text, "lowering{") || strings.Contains(text, "fresh()") {
				problem = "a lowering is made before the file's share is set"
				return false
			}
		}
		problem = "the file loop never sets latentEntriesShare = latentEntriesFile()"
		return false
	})
	if loops != 1 {
		return fmt.Sprintf("%d loops over program.Files(), want 1", loops)
	}
	if problem != "" {
		return problem
	}
	made := 0
	ast.Inspect(body, func(node ast.Node) bool {
		literal, ok := node.(*ast.CompositeLit)
		if !ok || render(files, literal.Type) != "lowering" {
			return true
		}
		made++
		shared := false
		for _, element := range literal.Elts {
			shared = shared || render(files, element) == "entriesRecords: latentEntriesShare.literals"
		}
		if !shared {
			problem = "a lowering is made without entriesRecords: latentEntriesShare.literals"
		}
		return true
	})
	if made == 0 {
		return "no lowering is made"
	}
	return problem
}

func render(files *token.FileSet, node ast.Node) string {
	var text bytes.Buffer
	if err := printer.Fprint(&text, files, node); err != nil {
		return ""
	}
	return text.String()
}

// The census's entriesProgramRecordsWalk, beside main's entriesProgramHasRecords, walks once a file whatever number of
// lowerings ask, replays moduleOrder for each lowering the share answers, keeps the walk's answer, and leaves the share
// unknown when the walk panics. Mutant: the share write (share.known, share.records = true, records) dropped, red.
func TestOneWalkAFile(t *testing.T) {
	t.Parallel()
	probe := shareProbe(t)
	log, err := os.Create(filepath.Join(t.TempDir(), "test.log"))
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(probe, "-test.count=1", "-test.v")
	command.Dir = filepath.Dir(probe)
	command.Stdout = log
	command.Stderr = log
	err = command.Run()
	_ = log.Close()
	if err != nil {
		contents, _ := os.ReadFile(log.Name())
		t.Fatalf("compiled share probe: %v\n%s", err, contents)
	}
}

// probeSources is the probe module: the census's share declarations and wrapper, cut from lower.go.txt, and main's
// entriesProgramHasRecords, cut from internal/lower/entries_records.go, beside a stub lowering and the probe's test.
func probeSources(repository string) (map[string]string, error) {
	census, err := declarations(filepath.Join(repository, "stage3/census/latent/lower.go.txt"),
		"latentEntriesAnswer", "latentEntriesShare", "latentEntriesFile", "latentEntriesProgramWalks", "entriesProgramRecordsWalk")
	if err != nil {
		return nil, err
	}
	compiler, err := declarations(filepath.Join(repository, "internal/lower/entries_records.go"), "entriesProgramHasRecords")
	if err != nil {
		return nil, err
	}
	module, err := os.ReadFile(filepath.Join(repository, "go.mod"))
	if err != nil {
		return nil, err
	}
	directives := ""
	for _, line := range strings.Split(string(module), "\n") {
		if strings.HasPrefix(line, "go ") || strings.HasPrefix(line, "toolchain ") {
			directives += strings.TrimSpace(line) + "\n"
		}
	}
	if !strings.HasPrefix(directives, "go ") {
		return nil, fmt.Errorf("the repository's go.mod declares no go version")
	}
	return map[string]string{
		"go.mod":        "module probe\n\n" + directives,
		"census.go":     "package lower\n\nimport \"probe/ast\"\n\n" + census,
		"ast/ast.go":    "package ast\n\ntype Node struct{}\n",
		"main.go":       "package lower\n\n" + compiler,
		"probe_test.go": probeTest,
		"lowering.go":   probeLowering,
	}, nil
}

// declarations cuts the named top-level declarations from a Go file, in the file's order, and fails on any it lacks.
func declarations(path string, names ...string) (string, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	files := token.NewFileSet()
	file, err := parser.ParseFile(files, filepath.Base(path), source, 0)
	if err != nil {
		return "", err
	}
	wanted := map[string]bool{}
	for _, name := range names {
		wanted[name] = true
	}
	found := map[string]bool{}
	text := ""
	for _, declaration := range file.Decls {
		name := ""
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			name = declaration.Name.Name
		case *ast.GenDecl:
			if len(declaration.Specs) == 1 {
				switch spec := declaration.Specs[0].(type) {
				case *ast.TypeSpec:
					name = spec.Name.Name
				case *ast.ValueSpec:
					if len(spec.Names) == 1 {
						name = spec.Names[0].Name
					}
				}
			}
		}
		if wanted[name] {
			if found[name] {
				return "", fmt.Errorf("%s declares %s twice", filepath.Base(path), name)
			}
			found[name] = true
			text += render(files, declaration) + "\n\n"
		}
	}
	for _, name := range names {
		if !found[name] {
			return "", fmt.Errorf("%s declares no top-level %s", filepath.Base(path), name)
		}
	}
	return text, nil
}

// probeLowering is the stub the cut declarations compile against: a lowering whose raw walk and moduleOrder count
// their calls, and can be told to panic.
const probeLowering = `package lower

type program struct{}

func (*program) Files() []string { return []string{"entry.ts"} }

type lowering struct {
	program                    *program
	entriesProgramRecords      bool
	entriesProgramRecordsKnown bool
	moduleOrders               int
}

var rawAnswer, rawPanics = true, false

func (l *lowering) moduleOrder(entry string) ([]string, error) {
	l.moduleOrders++
	return []string{entry}, nil
}

func (l *lowering) latentEntriesProgramRecordsWalkRaw() bool {
	l.moduleOrder("entry.ts")
	if rawPanics {
		panic("the walk panicked")
	}
	return rawAnswer
}
`

// probeTest drives two files of two lowerings each, as latentRun and latentFullSelected do: a fresh share a file, a
// fresh lowering (or a snapshot of one that hasn't asked) a unit.
const probeTest = `package lower

import "testing"

func fresh() *lowering { return &lowering{program: &program{}} }

func TestShare(t *testing.T) {
	latentEntriesShare, latentEntriesProgramWalks, rawAnswer, rawPanics = nil, 0, true, false
	if share := latentEntriesFile(); share.literals == nil || share.known {
		t.Fatal("a file's share starts without a literal memo, or known")
	}
	if !fresh().entriesProgramHasRecords() || !fresh().entriesProgramHasRecords() || latentEntriesProgramWalks != 2 {
		t.Fatalf("no share: %d walks, want one a lowering", latentEntriesProgramWalks)
	}
	for file, answer := range []bool{true, false} {
		latentEntriesShare, latentEntriesProgramWalks, rawAnswer = latentEntriesFile(), 0, answer
		first := fresh()
		if first.entriesProgramHasRecords() != answer || first.entriesProgramHasRecords() != answer {
			t.Fatalf("file %d: the first lowering's answer isn't the walk's", file)
		}
		rawAnswer = !answer
		second := fresh()
		if second.entriesProgramHasRecords() != answer {
			t.Fatalf("file %d: the second lowering's answer isn't the file's first walk's", file)
		}
		if !second.entriesProgramRecordsKnown || second.moduleOrders != 1 {
			t.Fatalf("file %d: the second lowering kept known=%v and ran moduleOrder %d times, want true and 1", file, second.entriesProgramRecordsKnown, second.moduleOrders)
		}
		if latentEntriesProgramWalks != 1 {
			t.Fatalf("file %d: %d walks for two lowerings, want 1", file, latentEntriesProgramWalks)
		}
	}
	latentEntriesShare, latentEntriesProgramWalks, rawAnswer, rawPanics = latentEntriesFile(), 0, true, true
	func() {
		defer func() { _ = recover() }()
		fresh().entriesProgramHasRecords()
	}()
	if latentEntriesShare.known {
		t.Fatal("a walk that panicked left its share known")
	}
	rawPanics = false
	if !fresh().entriesProgramHasRecords() || latentEntriesProgramWalks != 2 {
		t.Fatalf("after a panicked walk: %d walks, want the next lowering to walk again", latentEntriesProgramWalks)
	}
}
`

// probeToolchain is the probe's go, asked as refusalrewrite's probe asks it: toolchain's, with GOWORK off, at the
// repository, whose go and toolchain directives the probe's module declares.
func probeToolchain(t *testing.T) []string {
	t.Helper()
	var reports []string
	for _, arguments := range [][]string{{"version"}, {"env", "GOOS", "GOARCH", "CGO_ENABLED", "CC", "GOFLAGS", "GOEXPERIMENT"}} {
		command := toolchain.CommandIn("../../../..", "go", arguments...)
		command.Env = append(command.Env, "GOWORK=off")
		output, err := command.Output()
		if err != nil {
			t.Fatalf("go %s: %v", strings.Join(arguments, " "), err)
		}
		reports = append(reports, strings.TrimSpace(string(output)))
	}
	return reports
}

// shareProbe is the probe module's test binary, a product built ahead (go test -c with GOWORK off, which vets it as go
// test did), so the unit runs it without go. Its sources are cut in this process, so the files they're cut from and
// this file key it.
func shareProbe(t *testing.T) string {
	t.Helper()
	inputs := buildcache.Inputs{
		Name:      "latent entries share probe",
		Files:     []string{"go.mod", "stage3/census/latent/lower.go.txt", "internal/lower/entries_records.go", "stage3/census/latent/entriesshare/share_test.go"},
		Flags:     []string{"go test -c -trimpath -ldflags=-buildid= -buildvcs=false", "GOWORK=off"},
		Toolchain: probeToolchain(t),
	}
	directory := buildcache.Product(t, inputs, func(directory string) error {
		sources, err := probeSources("../../../..")
		if err != nil {
			return err
		}
		module := filepath.Join(directory, "module")
		for name, text := range sources {
			path := filepath.Join(module, name)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
				return err
			}
		}
		command := toolchain.CommandIn(module, "go", "test", "-c", "-trimpath", "-ldflags=-buildid=", "-buildvcs=false", "-o", filepath.Join(directory, "probe.test"), ".")
		command.Env = append(command.Env, "GOWORK=off")
		if combined, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("go test -c of the share probe: %v\n%s", err, combined)
		}
		return nil
	})
	return filepath.Join(directory, "probe.test")
}

func TestProduct_EntriesShareProbe(t *testing.T) {
	t.Parallel()
	shareProbe(t)
}
