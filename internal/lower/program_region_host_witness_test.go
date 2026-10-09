package lower

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Host 14 exercises captured cells and closures, not parser Node identities.
// Keep its actual selector output explicit instead of calling these trees.
func TestProgramRegionHost14MemberSites(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("../../stage3/fixtures/host/14_getCurrentDirectory.a")
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	entry := loaded.Files()[0]
	typeChecker, release := loaded.Checker(context.Background(), entry)
	defer release()
	discovery, err := lowerChecked(loaded, typeChecker, entry, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	members := []string{}
	for local, selected := range discovery.programPlan.cells {
		if !selected {
			continue
		}
		name := discovery.result.Locals[local].Name
		at := loaded.Where(discovery.localNodes[local])
		label := typeChecker.TypeToStringEx(discovery.localTypes[local], nil, checker.TypeFormatFlagsNoTruncation, nil)
		members = append(members, fmt.Sprintf("cell %s: %s at %s", name, label, at))
		if name != "callback" && name != "value" {
			t.Fatalf("unexpected host member cell %s", name)
		}
	}
	// Allocation membership also uses contextual types, so the input callback
	// closure can be adopted even though its function declaration is not a member.
	built, err := lowerChecked(loaded, typeChecker, entry, discovery.programPlan, false)
	if err != nil {
		t.Fatal(err)
	}
	closureMembers := map[int]bool{}
	inspect := func(node any) bool {
		if closure, ok := node.(ir.MakeClosure); ok && (closure.ProgramRegion || built.result.Functions[closure.Function].ProgramRegion) {
			closureMembers[closure.Function] = true
		}
		return true
	}
	walk(built.result.Main, inspect)
	for _, function := range built.result.Functions {
		walk(function.Body, inspect)
	}
	for _, record := range discovery.closureRecords {
		if !closureMembers[record.function] {
			continue
		}
		label := typeChecker.TypeToStringEx(record.proven, nil, checker.TypeFormatFlagsNoTruncation, nil)
		members = append(members, fmt.Sprintf("closure %s at %s", label, loaded.Where(record.node)))
	}
	sort.Strings(members)
	for _, member := range members {
		t.Log(member)
	}
	if len(members) != 4 {
		t.Fatalf("host 14 member allocation sites %d, want two cells and two closures: %v", len(members), members)
	}
	cells, functions := 0, len(closureMembers)
	for _, selected := range discovery.programPlan.cells {
		if selected {
			cells++
		}
	}
	if cells != 2 || functions != 2 {
		t.Fatalf("selector allocation sites cells=%d closures=%d, want 2 and 2", cells, functions)
	}
	t.Logf("selector selected type identities: %d; fixture 14 has no Node or NodeArray declarations or allocations", len(discovery.programPlan.types))
}

// The original scout driver is pinned evidence, not a rewritten parser witness.
// Only import extensions change while reconstructing it as .a modules.
func TestProgramRegionScoutParseSelectionEmpty(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../review/compiler/program-region-node-check/parse-source.json")
	if err != nil {
		t.Fatal(err)
	}
	var bundle struct {
		Pin   string
		Files map[string]struct {
			SHA256 string
			Text   string
		}
	}
	if err := json.Unmarshal(data, &bundle); err != nil {
		t.Fatal(err)
	}
	if bundle.Pin != "9e062aac7a55e117199c1cc5fbb7a0d59bedf07a" {
		t.Fatal("parse source pin changed")
	}
	root := t.TempDir()
	imports := regexp.MustCompile(`(from\s+['"][^'"]+)\.ts(['"])`)
	for path, file := range bundle.Files {
		if fmt.Sprintf("%x", sha256.Sum256([]byte(file.Text))) != file.SHA256 {
			t.Fatalf("parse source hash changed: %s", path)
		}
		if strings.HasSuffix(path, ".ts") {
			path = strings.TrimSuffix(path, ".ts") + ".a"
		}
		path = filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		text := imports.ReplaceAllString(file.Text, "${1}.a${2}")
		if err := os.WriteFile(path, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	loaded, err := load.Load([]string{filepath.Join(root, "stage1/cohere/parse/parse.a")})
	if err != nil {
		t.Fatal(err)
	}
	entry := loaded.Files()[0]
	typeChecker, release := loaded.Checker(context.Background(), entry)
	defer release()
	discovery, err := lowerChecked(loaded, typeChecker, entry, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	plan := discovery.programPlan
	if len(plan.types) != 0 {
		t.Fatalf("acyclic parse.a selector selected %d type identities", len(plan.types))
	}
	for local, member := range plan.cells {
		if member {
			t.Fatalf("acyclic parse.a cell %s selected", discovery.result.Locals[local].Name)
		}
	}
	for function, member := range plan.functions {
		if member {
			t.Fatalf("acyclic parse.a function %s selected", discovery.result.Functions[function].Name)
		}
	}
	for class, member := range plan.classes {
		if member {
			t.Fatalf("acyclic parse.a class %d selected", class)
		}
	}
	t.Log("parse.a selector member set empty: all parser allocations stay counted")
}
