package oracle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

type step09Fixture struct {
	File          string `json:"file"`
	Failing       bool   `json:"failing"`
	NodeOutput    string `json:"node_stdout"`
	RuntimeOutput string `json:"runtime_stdout"`
	RuntimeExit   int    `json:"runtime_exit"`
}

func step09Fixtures(t *testing.T) []step09Fixture {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repository, "stage3/fixtures/checked-casts/fixtures.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []step09Fixture
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	return fixtures
}

func TestStep09AcceptanceTagged(t *testing.T) {
	for _, fixture := range step09Fixtures(t) {
		if !strings.HasPrefix(fixture.File, "01_") && !strings.HasPrefix(fixture.File, "02_") {
			continue
		}
		t.Run(fixture.File, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/checked-casts", fixture.File))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != fixture.NodeOutput || len(truth.stderr) != 0 {
				t.Fatalf("Node golden: %#v", truth)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			expected := truth
			if fixture.Failing {
				message := "cast failed: this Node is not a EnumDeclaration"
				if strings.HasPrefix(fixture.File, "02_") {
					message = "cast failed: this DeclarationName is not a Identifier | StringLiteral"
				}
				expected = run{exitCode: 70, stdout: []byte(fixture.RuntimeOutput), stderr: []byte("adamic: panic: " + message + "\n")}
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if diff := disagreement(expected, got); diff != "" {
					t.Fatalf("contract: %s: %#v", diff, got)
				}
			}
			if !fixture.Failing {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
				return
			}
			// Omit only the actual cast-point tag test. Later field checks remain, so
			// the exact cast diagnostic pins the obligation at the cast itself.
			changed := 0
			omit := func(value ir.Expression) ir.Expression {
				if cast, ok := value.(ir.CheckedCast); ok {
					changed++
					return cast.Value
				}
				return value
			}
			mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), omit)
			mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), omit)
			if changed != 1 {
				t.Fatalf("want one tag check, got %d", changed)
			}
			for backend, got := range map[string]run{"native release": releasedUncached(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
				if disagreement(expected, got) == "" {
					t.Fatalf("%s tag omission survived", backend)
				}
				t.Logf("%s tag omission caught by exact cast-point diagnostic: exit=%d stderr=%q", backend, got.exitCode, got.stderr)
			}
		})
	}
}
