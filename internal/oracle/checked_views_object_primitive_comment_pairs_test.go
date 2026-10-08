package oracle

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// These witnesses import each complete upstream target, including its unread
// descendants. Candidate reads count provenance, not whole-tsc executions.
func TestCheckedViewObjectPrimitiveCommentPairs(t *testing.T) {
	declarations := os.Getenv("ADAMIC_OBJECT_PRIMITIVE_ORIGINAL_DECLS")
	if declarations == "" {
		t.Skip("set ADAMIC_OBJECT_PRIMITIVE_ORIGINAL_DECLS to prepare.cjs output")
	}
	data, err := os.ReadFile(checkedViewFixturePath(filepath.Join(declarations, "original-manifest.json")))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Commit       string              `json:"upstream_commit"`
		Declarations map[string]string   `json:"declarations"`
		Fields       map[string][]string `json:"fields"`
		ArrayFields  []string            `json:"array_fields"`
		Pairs        []struct {
			ID          int `json:"type_id"`
			Type, Field string
			Reads       int `json:"read_count"`
			Sites       []json.RawMessage
		}
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Commit != "050880ce59e30b356b686bd3144efe24f875ebc8" || len(manifest.Pairs) != 42 {
		t.Fatal("original provenance drift")
	}
	for name, expected := range manifest.Declarations {
		data, err := os.ReadFile(checkedViewFixturePath(filepath.Join(declarations, name)))
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != expected {
			t.Fatal("declaration drift", name)
		}
	}
	pairs, reads := 0, 0
	for _, pair := range manifest.Pairs {
		if pair.Field != "comment" {
			continue
		}
		pairs++
		reads += pair.Reads
		if len(pair.Sites) != pair.Reads {
			t.Fatal("site provenance drift", pair.ID)
		}
		for _, mode := range []string{"good", "wrong", "nested"} {
			t.Run(fmt.Sprintf("%d-%s", pair.ID, mode), func(t *testing.T) {
				name := fmt.Sprintf("jsdoc-%d-%s.a", pair.ID, mode)
				input, err := os.ReadFile(checkedViewFixturePath("../../stage3/interface-downcasts/lane4b/original/" + name))
				if err != nil {
					t.Fatal(err)
				}
				bound := strings.Replace(string(input), "'original-tsc-types'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/types.d.ts"))), 1)
				path := filepath.Join(t.TempDir(), name)
				if err := os.WriteFile(path, []byte(bound), 0600); err != nil {
					t.Fatal(err)
				}
				source := map[string]string{"good": "plain\n7\nabsent\n", "wrong": "wrong\n", "nested": "false\n"}[mode]
				if diff := disagreement(run{stdout: []byte(source)}, onNode(t, path)); diff != "" {
					t.Fatal("source Node: " + diff)
				}
				loaded, err := load.Load([]string{path})
				if err != nil {
					t.Fatal(err)
				}
				program, err := lower.Lower(context.Background(), loaded)
				if err != nil {
					t.Fatal(err)
				}
				for _, name := range append(strings.Split(pair.Type, " | "), "JSDocText", "JSDocLink", "JSDocLinkCode", "JSDocLinkPlain") {
					complete := false
					for _, contract := range program.ViewContracts {
						if contract.Name != name {
							continue
						}
						fields := make([]string, 0, len(contract.Fields))
						for _, field := range contract.Fields {
							fields = append(fields, field.Name)
						}
						slices.Sort(fields)
						complete = complete || len(fields) > 0 && slices.Equal(fields, manifest.Fields[name])
					}
					if !complete {
						t.Fatal("original target fields reduced or omitted", name)
					}
				}
				array := false
				for _, contract := range program.ViewContracts {
					if contract.Kind == ir.ViewArray && contract.Name == "NodeArray<JSDocComment>" {
						fields := make([]string, 0, len(contract.Fields))
						for _, field := range contract.Fields {
							fields = append(fields, field.Name)
						}
						slices.Sort(fields)
						array = contract.Element != 0 && len(fields) > 0 && slices.Equal(fields, manifest.ArrayFields)
					}
				}
				if !array {
					t.Fatal("original NodeArray element or metadata contract omitted")
				}
				want := run{stdout: []byte(source)}
				switch mode {
				case "wrong":
					want = run{exitCode: 70, stderr: []byte("adamic: panic: cast failed: field read failed: node.comment matches no member of string | NodeArray<JSDocComment> | undefined; expected string | NodeArray<JSDocComment> | undefined, found boolean\n")}
				case "nested":
					want = run{exitCode: 70, stderr: []byte("adamic: panic: cast failed: field read failed: first.flags is not a NodeFlags; expected NodeFlags, found boolean\n")}
				}
				objectPrimitiveOriginalCount(t, program, name)
				sanitized, binary := nativelyUncached(t, program)
				for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if diff := disagreement(want, got); diff != "" {
						t.Fatalf("original contract: %s; stderr %q", diff, got.stderr)
					}
				}
				if mode == "good" {
					if report := leaksUncached(t, program, binary); report != "" {
						t.Fatal(report)
					}
					return
				}
				caught := func(label string) {
					for index, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
						if got.exitCode != 0 || disagreement(want, got) == "" {
							t.Fatalf("%s must execute and fail the refusal pin, backend %d: %#v", label, index, got)
						}
						t.Logf("%s caught backend %d: exit %d stdout %q", label, index, got.exitCode, got.stdout)
					}
				}
				if mode == "nested" {
					if count := changeObjectPrimitiveRead(program, func(read ir.Property) bool { return read.View == "first.flags" }, func(read ir.Property) ir.Property { read.View = ""; return read }); count != 1 {
						t.Fatalf("want one nested mutant, got %d", count)
					}
					caught("drop nested check")
					return
				}
				var root ir.Property
				matches := func(read ir.Property) bool { return read.View == "node.comment" }
				if count := changeObjectPrimitiveRead(program, matches, func(read ir.Property) ir.Property { root = read; return read }); count != 1 {
					t.Fatalf("want one outer read, got %d", count)
				}
				changed := false
				for _, id := range program.ViewContracts[root.ViewContract-1].Members {
					member := &program.ViewContracts[id-1]
					if member.Kind != ir.ViewScalar || member.Of != ir.String {
						continue
					}
					original := *member
					member.Of, member.Allowed = ir.Boolean, nil
					// The coarse mask mirrors the member selection; mutate both.
					changeObjectPrimitiveRead(program, matches, func(read ir.Property) ir.Property { read.NullishKinds |= 1 << ir.Boolean; return read })
					caught("accept wrong member")
					changeObjectPrimitiveRead(program, matches, func(read ir.Property) ir.Property { return root })
					*member = original
					changed = true
					break
				}
				if !changed {
					t.Fatal("no independent member mutant")
				}
				changeObjectPrimitiveRead(program, matches, func(read ir.Property) ir.Property { read.View = ""; return read })
				caught("skip outer check")
			})
		}
	}
	if pairs != 16 || reads != 50 {
		t.Fatalf("pair demand drift: %d pairs / %d reads", pairs, reads)
	}
}
