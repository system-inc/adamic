package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

func TestCheckedViewNominalArrayReadMutant(t *testing.T)     { nominalArrayMutant(t, "read") }
func TestCheckedViewNominalArrayWriteMutant(t *testing.T)    { nominalArrayMutant(t, "write") }
func TestCheckedViewNominalArrayProducerMutant(t *testing.T) { nominalArrayMutant(t, "producer") }
func nominalArrayMutant(t *testing.T, site string) {
	write := site == "write"
	fixture := "entry-nominal-array-" + site
	program, path := interfaceFixture(t, "nullish/maps/"+fixture)
	truth := onNode(t, path)
	fake := -1
	for i, local := range program.Locals {
		if local.Name == "fake" {
			fake = i
		}
	}
	if fake < 0 {
		t.Fatal("missing fake")
	}
	changed := false
	for i, statement := range program.Main {
		if write {
			if evaluation, ok := statement.(ir.Evaluate); ok {
				if push, ok := evaluation.Value.(ir.ArrayPush); ok {
					push.Value = ir.Read{Local: fake, Of: ir.Object}
					evaluation.Value = push
					program.Main[i] = evaluation
					changed = true
				}
			}
			continue
		}
		declaration, ok := statement.(ir.Declare)
		if !ok || program.Locals[declaration.Local].Name != "item" {
			continue
		}
		array, ok := declaration.Value.(ir.ArrayLiteral)
		if !ok {
			t.Fatal("missing array literal")
		}
		array.Elements[0] = ir.Read{Local: fake, Of: ir.Object}
		declaration.Value = array
		program.Main[i] = declaration
		changed = true
	}
	if !changed {
		t.Fatal("mutation missed")
	}
	native, _ := nativelyUncached(t, program)
	for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || !strings.Contains(string(got.stderr), "class identity") || (site == "read" && !strings.Contains(string(got.stderr), "field read failed:")) || (site == "producer" && !strings.Contains(string(got.stderr), "Map nominal producer failed:")) {
			t.Fatalf("array lookalike ran on: %#v", got)
		}
	}
	t.Logf("Node=%q; lookalike caught at %s", truth.stdout, site)
}

func TestCheckedViewMutableNominalFieldMutant(t *testing.T) { mutableNominalFieldMutant(t, false) }
func TestCheckedViewMutableNominalWriteMutant(t *testing.T) { mutableNominalFieldMutant(t, true) }
func mutableNominalFieldMutant(t *testing.T, write bool) {
	fixture := "entry-nominal-mutable"
	if write {
		fixture += "-write"
	}
	program, path := interfaceFixture(t, "nullish/maps/"+fixture)
	truth := onNode(t, path)
	fake := -1
	for i, local := range program.Locals {
		if local.Name == "fake" {
			fake = i
		}
	}
	if fake < 0 {
		t.Fatal("missing fake")
	}
	changed := false
	for i, statement := range program.Main {
		if store, ok := statement.(ir.SetProperty); ok && store.Name == "child" && !store.Define {
			store.Value = ir.Read{Local: fake, Of: ir.Object}
			if !write {
				store.WriteContract = 0
			}
			program.Main[i] = store
			changed = true
		}
	}
	if !changed {
		t.Fatal("mutation missed")
	}
	native, _ := nativelyUncached(t, program)
	for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || !strings.Contains(string(got.stderr), map[bool]string{false: "field read failed:", true: "Map nominal producer failed:"}[write]) || !strings.Contains(string(got.stderr), "class identity") {
			t.Fatalf("mutable nominal lookalike ran on: %#v", got)
		}
	}
	t.Logf("Node=%q; forged alias store caught at next field read", truth.stdout)
}

func TestCheckedViewNullableNominalAggregateMutants(t *testing.T) {
	for _, kind := range []string{"null", "undefined", "both"} {
		sites := []string{"child"}
		if kind != "both" {
			sites = append(sites, "opposite")
		}
		for _, site := range sites {
			t.Run(kind+"-"+site, func(t *testing.T) {
				program, path := interfaceFixture(t, "nullish/maps/entry-nominal-aggregate-"+kind+"-producer")
				truth := onNode(t, path)
				fake := -1
				for i, local := range program.Locals {
					if local.Name == "fake" {
						fake = i
					}
				}
				if fake < 0 {
					t.Fatal("missing fake")
				}
				changed := false
				for i, statement := range program.Main {
					declaration, ok := statement.(ir.Declare)
					if !ok || program.Locals[declaration.Local].Name != "source" {
						continue
					}
					creation, ok := declaration.Value.(ir.MapNew)
					if !ok {
						t.Fatal("missing Map")
					}
					if site == "child" {
						creation.Entries[0][1] = ir.ObjectLiteral{Fields: []ir.Field{{Name: "child", Value: ir.Read{Local: fake, Of: ir.Object}}}}
					} else if kind == "null" {
						creation.Entries[1][1] = ir.Undefined{Of: ir.Union}
					} else {
						creation.Entries[1][1] = ir.Box{Value: ir.Null{}, NullReference: true}
					}
					declaration.Value = creation
					program.Main[i] = declaration
					changed = true
				}
				if !changed {
					t.Fatal("mutation missed")
				}
				native, _ := nativelyUncached(t, program)
				for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if got.exitCode != 70 || !strings.Contains(string(got.stderr), "Map nominal producer failed:") || !strings.Contains(string(got.stderr), "class identity") {
						t.Fatalf("nullable aggregate ran on: %#v", got)
					}
				}
				t.Logf("Node=%q; %s rejected for %s", truth.stdout, site, kind)
			})
		}
	}
}
