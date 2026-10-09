package oracle

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
)

type intersectionStopExpectation struct {
	Kind      string `json:"kind"`
	Where     string `json:"where"`
	What      string `json:"what"`
	Fix       string `json:"fix"`
	Fixture   string `json:"fixture"`
	SourceSHA string `json:"source_sha256"`
}

func intersectionCanonicalSource(source string) string {
	if declarations := os.Getenv("ADAMIC_INTERSECTION_ORIGINAL_DECLS"); declarations != "" {
		source = strings.ReplaceAll(source, fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/types.d.ts"))), "'original-tsc-types'")
	}
	return source
}

// A stop is pinned independently of the returned IR. Admission cannot replace it.
func intersectionExpectedStop(t *testing.T, path string, err error) bool {
	t.Helper()
	key := strings.ReplaceAll(t.Name(), "/", "__") + "__" + filepath.Base(path)
	base := filepath.Join(repository, "stage3/interface-downcasts/lane7/refusals")
	expectationPath := filepath.Join(base, key+".json")
	var refusal *lower.Refused
	var notYet *lower.NotYet
	actual := intersectionStopExpectation{}
	if errors.As(err, &refusal) {
		actual.Kind = "Refused"
		actual.Where = refusal.Where
		actual.What = refusal.What
		actual.Fix = refusal.Fix
	}
	if errors.As(err, &notYet) {
		actual.Kind = "NotYet"
		actual.Where = notYet.Where
		actual.What = notYet.What
		actual.Fix = notYet.Fix
	}
	data, readErr := os.ReadFile(expectationPath)
	if os.IsNotExist(readErr) {
		if err != nil {
			t.Fatalf("unrecorded stop: %v", err)
		}
		return false
	}
	if readErr != nil {
		t.Fatal(readErr)
	}
	var want intersectionStopExpectation
	if decodeErr := json.Unmarshal(data, &want); decodeErr != nil {
		t.Fatal(decodeErr)
	}
	if actual.Kind != want.Kind || actual.Where != path+want.Where || actual.What != want.What || actual.Fix != want.Fix {
		t.Fatalf("want %s at %s%s: %s; %s, got %v", want.Kind, path, want.Where, want.What, want.Fix, err)
	}
	input, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	source := intersectionCanonicalSource(string(input))
	fixture, readErr := os.ReadFile(filepath.Join(repository, want.Fixture))
	if readErr != nil {
		t.Fatal(readErr)
	}
	if source != string(fixture) || fmt.Sprintf("%x", sha256.Sum256(fixture)) != want.SourceSHA {
		t.Fatal("refusal program differs from its preserved .a fixture")
	}
	if actual.Kind == "Refused" && actual.Fix == "" {
		t.Fatal("a ruled refusal must name its fix")
	}
	t.Logf("pinned %s at %s: %s; %s", actual.Kind, actual.Where, actual.What, actual.Fix)
	return true
}

func intersectionLaneFixture(t *testing.T, name string) (*ir.Program, string, bool) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, interfaceFixturePath(name)))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	stopped := intersectionExpectedStop(t, path, err)
	return program, path, stopped
}
