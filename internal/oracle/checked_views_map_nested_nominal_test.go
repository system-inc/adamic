package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

func TestCheckedViewMapNestedNominalMutants(t *testing.T) {
	for _, site := range []string{"constructor", "set", "copy", "receiver", "generic", "field"} {
		t.Run(site, func(t *testing.T) {
			fixture := "entry-nominal-nested"
			if site == "constructor" || site == "set" || site == "copy" {
				fixture += "-producer"
			}
			if site == "field" {
				fixture += "-field"
			}
			program, path := interfaceFixture(t, "nullish/maps/"+fixture)
			truth := onNode(t, path)
			locals := map[string]int{}
			for i, l := range program.Locals {
				locals[l.Name] = i
			}
			fake, ok := locals["fake"]
			if !ok {
				t.Fatal("missing fake")
			}
			replacement := ir.Read{Local: fake, Of: ir.Object}
			if site != "receiver" && site != "generic" && site != "field" {
				forged, ok := locals["forged"]
				if !ok {
					t.Fatal("missing forged")
				}
				replacement.Local = forged
			}
			changed := false
			for i, statement := range program.Main {
				if declaration, ok := statement.(ir.Declare); ok {
					creation, ok := declaration.Value.(ir.MapNew)
					if !ok {
						continue
					}
					name := program.Locals[declaration.Local].Name
					if site == "constructor" && name == "source" {
						creation.Entries[0][1] = replacement
					} else if site == "copy" && name == "copy" {
						key := -1
						for i, s := range program.Strings {
							if s == "key" {
								key = i
							}
						}
						if key < 0 {
							t.Fatal("missing key")
						}
						pair := ir.ObjectLiteral{Fields: []ir.Field{{Name: "0", Value: ir.StringConstant{Index: key}}, {Name: "1", Value: replacement}}}
						creation.Pairs = ir.ArrayLiteral{Element: ir.Object, Elements: []ir.Expression{pair}}
					} else if site == "field" && name == "source" {
						item, ok := locals["item"]
						if !ok {
							t.Fatal("missing item")
						}
						mutation := ir.SetProperty{Object: ir.Read{Local: item, Of: ir.Object}, Name: "child", Value: replacement}
						tail := append([]ir.Statement(nil), program.Main[i+1:]...)
						program.Main = append(program.Main[:i+1], mutation)
						program.Main = append(program.Main, tail...)
						changed = true
						break
					} else {
						continue
					}
					declaration.Value = creation
					program.Main[i] = declaration
					changed = true
					break
				}
				if evaluation, ok := statement.(ir.Evaluate); ok {
					if store, ok := evaluation.Value.(ir.MapSet); ok && site == "set" {
						store.Value = replacement
						evaluation.Value = store
						program.Main[i] = evaluation
						changed = true
						break
					}
					if call, ok := evaluation.Value.(ir.Call); ok && (site == "receiver" && program.Functions[call.Function].Name == "inspectNamed" || site == "generic" && strings.HasPrefix(program.Functions[call.Function].Name, "inspectGeneric")) {
						call.Arguments[0] = replacement
						evaluation.Value = call
						program.Main[i] = evaluation
						changed = true
						break
					}
				}
			}
			if !changed {
				t.Fatal("mutation missed")
			}
			native, _ := nativelyUncached(t, program)
			for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				expected := "Map nominal producer failed:"
				if site == "receiver" || site == "generic" || site == "field" {
					expected = "cast failed: field read failed:"
				}
				if got.exitCode != 70 || !strings.Contains(string(got.stderr), expected) || !strings.Contains(string(got.stderr), "class identity") {
					t.Fatalf("nominal lookalike ran on at %s: %#v", site, got)
				}
			}
			t.Logf("Node=%q; lookalike caught at %s", truth.stdout, site)
		})
	}
}
