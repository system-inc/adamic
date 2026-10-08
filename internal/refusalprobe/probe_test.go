package refusalprobe

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func TestCatalogCoverage(t *testing.T) {
	if err := ValidateCatalog("../.."); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for index := 0; index < 200; index++ {
		p := Generate(1, index)
		seen[p.Entry] = true
		if p.Source == p.Neighbor {
			t.Fatal("identical probe and control")
		}
		if p != Generate(1, index) {
			t.Fatal("writer is not deterministic")
		}
	}
	for _, entry := range Catalog() {
		if entry.Boundary == "" && !seen[entry.Name] {
			t.Errorf("missing %s", entry.Name)
		}
	}
}

func TestCatalogAuditCanFail(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "internal/lower")
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatal(err)
	}
	source := `package lower
var refusals = map[int]refusal{1: {"a newly refused construct", "fix"}}
`
	if err := os.WriteFile(filepath.Join(path, "refusals.go"), []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	if err := ValidateCatalog(directory); err == nil || !strings.Contains(err.Error(), "newly refused") {
		t.Fatalf("missing entry was not caught: %v", err)
	}
}

func TestWrongDiagnosticIsAFinding(t *testing.T) {
	program, err := GenerateEntries(1, 0, []string{"definite-local"})
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"right", &lower.Refused{What: program.Expected, Fix: "initialize"}, ""},
		{"wrong", &lower.Refused{What: "the non-null assertion !", Fix: "narrow"}, "wrong-refusal"},
		{"accepted", nil, "accepted"},
		{"not-yet", &lower.NotYet{What: "assignment"}, "not-yet"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			finding := Diagnose(program, test.err)
			if test.want == "" {
				if finding != nil {
					t.Fatal(finding)
				}
				return
			}
			if finding == nil || finding.Kind != test.want {
				t.Fatalf("want %s, got %v", test.want, finding)
			}
		})
	}
}

func TestExecutableNeighbors(t *testing.T) {
	for index := 0; index < 200; index++ {
		p := Generate(1, index)
		t.Run(p.Entry, func(t *testing.T) {
			if err := Compile(context.Background(), p.Neighbor); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestNeighborFailureIsNotACompilerHole(t *testing.T) {
	p := Generate(1, 0)
	p.Neighbor = "const value: number = 'wrong';"
	finding := Check(context.Background(), p)
	if finding == nil || finding.Kind != "invalid-neighbor" {
		t.Fatalf("broken control became a compiler finding: %v", finding)
	}
}

func TestStrictDiagnosticWithRealLowering(t *testing.T) {
	p, err := GenerateEntries(1, 0, []string{"definite-local"})
	if err != nil {
		t.Fatal(err)
	}
	p.Expected = "@ts-ignore suppression directive"
	finding := Check(context.Background(), p)
	if finding == nil || finding.Kind != "wrong-refusal" {
		t.Fatalf("unrelated real refusal passed: %v", finding)
	}
}

func TestDirectRefusalAuditCanFail(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "internal/lower")
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatal(err)
	}
	source := `package lower
func refuse() error { return &Refused{What: "a new direct refusal"} }
`
	if err := os.WriteFile(filepath.Join(path, "refusals.go"), []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	if err := ValidateCatalog(directory); err == nil {
		t.Fatal("new direct refusal escaped the catalog")
	}
}

// Exercise both sides of every executable promise through the real loader/lowerer.
func TestGeneratedRefusals(t *testing.T) {
	for _, entry := range executable() {
		t.Run(entry.Name, func(t *testing.T) {
			program := generate(1, 0, entry)
			if err := Compile(context.Background(), program.Neighbor); err != nil {
				t.Fatalf("neighbor: %v", err)
			}
			err := Compile(context.Background(), program.Source)
			if finding := Diagnose(program, err); finding != nil {
				t.Fatalf("%s: %s\nsource:\n%s\nneighbor:\n%s", finding.Kind, finding.Diagnostic, program.Source, program.Neighbor)
			}
			t.Logf("bad: %v; neighbor: compiled", err)
		})
	}
}

func TestEnumNotYetBoundaries(t *testing.T) {
	for _, entry := range Catalog() {
		if !strings.HasPrefix(entry.Name, "enum-") || entry.Boundary == "" {
			continue
		}
		t.Run(entry.Name, func(t *testing.T) {
			program := generate(1, 0, entry)
			var notYet *lower.NotYet
			err := Compile(context.Background(), program.Source)
			if !errors.As(err, &notYet) || notYet.What != entry.Diagnostic {
				t.Fatalf("want NotYet %q, got %v", entry.Diagnostic, err)
			}
			if err := Compile(context.Background(), program.Neighbor); err != nil {
				t.Fatal(err)
			}
			if _, err := GenerateEntries(1, 0, []string{entry.Name}); err == nil {
				t.Fatal("boundary counted as executable refusal")
			}
			t.Logf("bad: NotYet %s; neighbor: compiled", notYet.What)
		})
	}
}

func TestAcceptedProgramIsAFindingWithRealLowering(t *testing.T) {
	program, err := GenerateEntries(1, 0, []string{"optional-widening"})
	if err != nil {
		t.Fatal(err)
	}
	program.Source = program.Neighbor
	finding := Check(context.Background(), program)
	if finding == nil || finding.Kind != "accepted" {
		t.Fatalf("accepted forbidden input escaped: %v", finding)
	}
}

// Observe both inputs, but leave the language decision to the owner. Questions
// remain printed boundaries and cannot be selected as forbidden constructs.
func TestUnsettledQuestions(t *testing.T) {
	for _, entry := range Catalog() {
		if !strings.HasPrefix(entry.Boundary, "Question:") {
			continue
		}
		t.Run(entry.Name, func(t *testing.T) {
			program := generate(1, 0, entry)
			if err := Compile(context.Background(), program.Neighbor); err != nil {
				t.Fatal(err)
			}
			err := Compile(context.Background(), program.Source)
			t.Logf("main observation: %v; neighbor: compiled", err)
			if _, err := GenerateEntries(1, 0, []string{entry.Name}); err == nil {
				t.Fatal("unsettled entry counted as an executable refusal")
			}
			t.Skip(entry.Boundary)
		})
	}
}
