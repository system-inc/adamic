package oracle

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func scannerCastFixture(t *testing.T, name string) (*ir.Program, string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "stage3/drivers/scanner/cast-checks", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	return program, path
}

func TestScannerCastDiagnostic(t *testing.T) {
	for _, name := range []string{"good", "undefined", "missing", "wrong"} {
		t.Run(name, func(t *testing.T) {
			program, path := scannerCastFixture(t, "diagnostic-"+name)
			truth := onNode(t, path)
			output := map[string]string{"good": "true\n", "undefined": "undefined\n", "missing": "undefined\n", "wrong": "42\n"}[name]
			if truth.exitCode != 0 || string(truth.stdout) != output {
				t.Fatalf("source Node: %#v", truth)
			}
			expected := truth
			if name == "wrong" {
				expected = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: value.elidedInCompatabilityPyramid matches no member of boolean | undefined; expected boolean | undefined, found number\n")}
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(expected, got); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
				}
			}
			if name != "wrong" {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}

func TestScannerCastDiagnosticMutant(t *testing.T) {
	program, _ := scannerCastFixture(t, "diagnostic-wrong")
	expected := onJavaScriptBackend(t, program)
	if expected.exitCode != 70 {
		t.Fatalf("liar has no check: %#v", expected)
	}
	changed := 0
	omit := func(value ir.Expression) ir.Expression {
		property, ok := value.(ir.Property)
		if ok && property.Name == "elidedInCompatabilityPyramid" && property.View != "" {
			property.View = ""
			changed++
			return property
		}
		return value
	}
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), omit)
	mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), omit)
	if changed != 1 {
		t.Fatalf("want one consumed field check, changed %d", changed)
	}
	for backend, got := range map[string]run{"native release": releasedUncached(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 {
			t.Fatalf("%s mutant must execute unchecked read: %#v", backend, got)
		}
		if difference := disagreement(expected, got); difference == "" {
			t.Fatal("field-check omission survived")
		}
		t.Logf("%s omission caught by exit/message pin: exit=%d stdout=%q", backend, got.exitCode, got.stdout)
	}
}

func TestScannerCastConstructorFrontiers(t *testing.T) {
	for _, test := range []struct{ name, output, reason string }{
		{"error-any", "ok\n", "a cast the runtime can't check"},
		{"error-typed", "ok\n", "captureStackTrace"},
		{"string-typed", "A\n", "String as a value outside equality or typeof"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "stage3/drivers/scanner/cast-checks", test.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != test.output {
				t.Fatalf("source Node: %#v", truth)
			}
			_, err = lowered(t, path)
			if err == nil || !strings.Contains(err.Error(), test.reason) {
				t.Fatalf("want named constructor frontier %q, got %v", test.reason, err)
			}
			t.Logf("Node-valid frontier: %v", err)
		})
	}
}

func scannerCastCounts(t *testing.T) []string {
	t.Helper()
	var rows []string
	for _, name := range []string{"good", "undefined", "missing", "wrong"} {
		rows = append(rows, counted(t, "stage3/drivers/scanner/cast-checks/diagnostic-"+name+".a", false, nil, false, false))
	}
	for _, name := range []string{"string-any", "string-checked"} {
		rows = append(rows, counted(t, "stage3/drivers/scanner/cast-checks/"+name+".a", false, nil, false, false))
	}
	for _, name := range []string{"01_enum_declaration", "01_enum_declaration_fails", "02_name_subunion", "02_name_subunion_fails", "05_extension", "05_extension_fails", "06_parser_keyword", "06_parser_keyword_fails", "09_primitive_string", "09_primitive_string_fails"} {
		rows = append(rows, counted(t, "stage3/fixtures/checked-casts/"+name+".a", false, nil, false, false))
	}
	return rows
}

// Keep the unit rows measurable when unrelated inherited fixtures block the
// repository-wide refresh. The ordinary counts hook still includes these rows.
func TestScannerCastCounts(t *testing.T) {
	rows := scannerCastCounts(t)
	data, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	if !*updateCounts {
		for _, row := range rows {
			if !strings.Contains(string(data), row+"\n") {
				t.Errorf("unrecorded scanner counts: %s", row)
			}
		}
		return
	}
	parts := strings.SplitN(string(data), "\n## Predicate direction counts", 2)
	var lines []string
	for _, line := range strings.Split(strings.TrimSuffix(parts[0], "\n"), "\n") {
		if !strings.HasPrefix(line, "| stage3/drivers/scanner/cast-checks/") && !strings.HasPrefix(line, "| stage3/fixtures/checked-casts/01_") && !strings.HasPrefix(line, "| stage3/fixtures/checked-casts/02_") && !strings.HasPrefix(line, "| stage3/fixtures/checked-casts/05_") && !strings.HasPrefix(line, "| stage3/fixtures/checked-casts/06_") && !strings.HasPrefix(line, "| stage3/fixtures/checked-casts/09_") {
			lines = append(lines, line)
		}
	}
	lines = append(lines, rows...)
	updated := strings.Join(lines, "\n") + "\n"
	if len(parts) == 2 {
		updated += "\n## Predicate direction counts" + parts[1]
	}
	if err := os.WriteFile(countsPath, []byte(updated), 0644); err != nil {
		t.Fatal(err)
	}
}
