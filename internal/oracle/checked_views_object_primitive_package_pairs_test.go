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

func TestCheckedViewObjectPrimitivePackagePairs(t *testing.T) {
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
		Pairs        []struct {
			Type, Field string
			Reads       int `json:"read_count"`
			Sites       []json.RawMessage
		}
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Commit != "050880ce59e30b356b686bd3144efe24f875ebc8" {
		t.Fatal("upstream drift")
	}
	for name, hash := range manifest.Declarations {
		data, err := os.ReadFile(checkedViewFixturePath(filepath.Join(declarations, name)))
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != hash {
			t.Fatal("declaration drift", name)
		}
	}
	pairs, reads := 0, 0
	for _, pair := range manifest.Pairs {
		if pair.Type != "PackageJsonInfoContents" {
			continue
		}
		pairs++
		reads += pair.Reads
		if len(pair.Sites) != pair.Reads {
			t.Fatal("site drift")
		}
		for _, mode := range []string{"good", "wrong", "nested"} {
			t.Run(pair.Field+"-"+mode, func(t *testing.T) {
				name := "package-" + pair.Field + "-" + mode + ".a"
				input, err := os.ReadFile(checkedViewFixturePath("../../stage3/interface-downcasts/lane4b/original/" + name))
				if err != nil {
					t.Fatal(err)
				}
				bound := strings.Replace(string(input), "'original-tsc-resolver'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/moduleNameResolver.d.ts"))), 1)
				path := filepath.Join(t.TempDir(), name)
				if err := os.WriteFile(path, []byte(bound), 0600); err != nil {
					t.Fatal(err)
				}
				source := "true\n"
				declared := "false | VersionPaths | undefined"
				nested := "member.version"
				good := "false\nfalse\nabsent\n"
				nestedSource := "false\n"
				nestedMessage := "field read failed: member.version is not a string; expected string, found boolean"
				if pair.Field == "resolvedEntrypoints" {
					declared = "false | string[] | undefined"
					nested = "member[0]"
					good = "false\nstring\nabsent\n"
					nestedSource = "boolean\n"
					nestedMessage = "element read failed: member[0] expected string, found boolean"
				}
				if mode == "good" {
					source = good
				} else if mode == "nested" {
					source = nestedSource
				}
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
				for _, name := range []string{"PackageJsonInfoContents", "VersionPaths"} {
					complete := false
					for _, contract := range program.ViewContracts {
						if contract.Name != name {
							continue
						}
						var fields []string
						for _, field := range contract.Fields {
							fields = append(fields, field.Name)
						}
						slices.Sort(fields)
						complete = complete || len(fields) > 0 && slices.Equal(fields, manifest.Fields[name])
					}
					if !complete {
						t.Fatal("original fields reduced", name)
					}
				}
				want := run{stdout: []byte(source)}
				if mode == "wrong" {
					want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: info." + pair.Field + " matches no member of " + declared + "; expected " + declared + ", found boolean\n")}
				}
				if mode == "nested" {
					want = run{exitCode: 70, stderr: []byte("adamic: panic: " + nestedMessage + "\n")}
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
							t.Fatalf("%s must execute and fail only the pin, backend %d: %#v", label, index, got)
						}
						t.Logf("%s caught backend %d: exit %d stdout %q", label, index, got.exitCode, got.stdout)
					}
				}
				if mode == "nested" {
					changed := 0
					if pair.Field == "versionPaths" {
						changed = changeObjectPrimitiveRead(program, func(read ir.Property) bool { return read.View == nested }, func(read ir.Property) ir.Property { read.View = ""; return read })
					} else {
						mutate := func(node any) any {
							if read, ok := node.(ir.ArrayIndex); ok && read.View == nested {
								read.View = ""
								changed++
								return read
							}
							return node
						}
						for index := range program.Functions {
							program.Functions[index].Body = mutateReadiness(program.Functions[index].Body, mutate)
						}
					}
					if changed != 1 {
						t.Fatalf("want one nested/element mutant, got %d", changed)
					}
					caught("drop nested or element check")
					return
				}
				matches := func(read ir.Property) bool { return read.View == "info."+pair.Field }
				var root ir.Property
				if count := changeObjectPrimitiveRead(program, matches, func(read ir.Property) ir.Property { root = read; return read }); count != 1 {
					t.Fatalf("want one outer read, got %d", count)
				}
				changed := false
				for _, id := range program.ViewContracts[root.ViewContract-1].Members {
					member := &program.ViewContracts[id-1]
					if member.Kind != ir.ViewScalar || member.Of != ir.Boolean || len(member.Allowed) != 1 {
						continue
					}
					original := *member
					member.Allowed = append([]ir.ViewLiteral(nil), member.Allowed...)
					member.Allowed[0].Boolean = true
					caught("accept wrong member")
					*member = original
					changed = true
					break
				}
				if !changed {
					t.Fatal("missing literal member mutant")
				}
				changeObjectPrimitiveRead(program, matches, func(read ir.Property) ir.Property { read.View = ""; return read })
				caught("skip outer check")
			})
		}
	}
	if pairs != 2 || reads != 4 {
		t.Fatalf("package demand drift: %d pairs / %d reads", pairs, reads)
	}
}
