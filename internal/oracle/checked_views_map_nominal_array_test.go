package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
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
		if got.exitCode != 70 || !strings.Contains(string(got.stderr), "class identity") || (site == "read" && !strings.Contains(string(got.stderr), "cast failed: field read failed:")) || (site == "producer" && !strings.Contains(string(got.stderr), "Map nominal producer failed:")) {
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
		if got.exitCode != 70 || !strings.Contains(string(got.stderr), map[bool]string{false: "cast failed: field read failed:", true: "Map nominal producer failed:"}[write]) || !strings.Contains(string(got.stderr), "class identity") {
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

func TestCheckedViewOptionalNominalMutants(t *testing.T) {
	for _, site := range []string{"producer", "container", "read"} {
		t.Run(site, func(t *testing.T) {
			fixture := "entry-nominal-optional"
			if site != "read" {
				fixture += "-producer"
			}
			program, path := interfaceFixture(t, "nullish/maps/"+fixture)
			truth := onNode(t, path)
			fake, item := -1, -1
			for i, local := range program.Locals {
				if local.Name == "fake" {
					fake = i
				}
				if local.Name == "item" {
					item = i
				}
			}
			if fake < 0 || item < 0 {
				t.Fatal("missing local")
			}
			changed := false
			for i, statement := range program.Main {
				declaration, ok := statement.(ir.Declare)
				if !ok || program.Locals[declaration.Local].Name != "source" {
					continue
				}
				if site != "read" {
					creation, ok := declaration.Value.(ir.MapNew)
					if !ok {
						t.Fatal("missing Map")
					}
					creation.Entries[0][1] = ir.ObjectLiteral{Fields: []ir.Field{{Name: "child", Value: ir.Read{Local: fake, Of: ir.Object}}}}
					if site == "container" {
						creation.Entries[0][1] = ir.ArrayLiteral{Element: ir.Number, Elements: []ir.Expression{ir.NumberConstant{Value: 7}}}
					}
					declaration.Value = creation
					program.Main[i] = declaration
				} else {
					mutation := ir.SetProperty{Object: ir.Read{Local: item, Of: ir.Object}, Name: "child", Value: ir.Read{Local: fake, Of: ir.Object}}
					remaining := append([]ir.Statement(nil), program.Main[i+1:]...)
					program.Main = append(program.Main[:i+1], mutation)
					program.Main = append(program.Main, remaining...)
				}
				changed = true
				break
			}
			if !changed {
				t.Fatal("mutation missed")
			}
			native, _ := nativelyUncached(t, program)
			expected := "cast failed: field read failed:"
			if site != "read" {
				expected = "Map nominal producer failed:"
			}
			for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if got.exitCode != 70 || !strings.Contains(string(got.stderr), expected) || !strings.Contains(string(got.stderr), "class identity") {
					t.Fatalf("optional nominal lookalike ran on: %#v", got)
				}
			}
			t.Logf("Node=%q; optional lookalike caught at %s", truth.stdout, site)
		})
	}
}

func TestCheckedViewNullableNominalArrayMutants(t *testing.T) {
	for _, kind := range []string{"null", "undefined", "both"} {
		for _, site := range []string{"read", "producer", "write", "widen"} {
			t.Run(kind+"-"+site, func(t *testing.T) {
				program, path := interfaceFixture(t, "nullish/maps/entry-nominal-elements-"+kind+"-"+site)
				truth := onNode(t, path)
				fake := -1
				for i, local := range program.Locals {
					if local.Name == "fake" {
						fake = i
					}
				}
				if fake < 0 && site != "widen" {
					t.Fatal("missing fake")
				}
				changed := false
				for i, statement := range program.Main {
					if site == "write" || site == "widen" {
						if evaluation, ok := statement.(ir.Evaluate); ok {
							if push, ok := evaluation.Value.(ir.ArrayPush); ok {
								push.Value = ir.Read{Local: fake, Of: ir.Object}
								if push.Element == ir.Union {
									push.Value = ir.Box{Value: push.Value}
								}
								if site == "widen" {
									push.Value = ir.Box{Value: ir.Null{}, NullReference: true}
									if kind == "undefined" {
										push.Value = ir.Undefined{Of: ir.Object}
									}
								}
								evaluation.Value = push
								program.Main[i] = evaluation
								changed = true
							}
						}
						continue
					}
					if declaration, ok := statement.(ir.Declare); ok && program.Locals[declaration.Local].Name == "item" {
						array, ok := declaration.Value.(ir.ArrayLiteral)
						if !ok {
							t.Fatal("missing array")
						}
						array.Elements[0] = ir.Read{Local: fake, Of: ir.Object}
						if array.Element == ir.Union {
							array.Elements[0] = ir.Box{Value: array.Elements[0]}
						}
						declaration.Value = array
						program.Main[i] = declaration
						changed = true
					}
				}
				if !changed {
					t.Fatal("mutation missed")
				}
				native, _ := nativelyUncached(t, program)
				for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if got.exitCode != 70 || !strings.Contains(string(got.stderr), "class identity") {
						t.Fatalf("nullable class lookalike ran on: %#v", got)
					}
				}
				t.Logf("Node=%q; nullable class lookalike caught at %s", truth.stdout, site)
			})
		}
	}
}

func TestCheckedViewOptionalMutableNominalMutants(t *testing.T) {
	t.Run("narrow-slot", func(t *testing.T) {
		program, path := interfaceFixture(t, "nullish/maps/entry-nominal-optional-mutable-narrow")
		truth := onNode(t, path)
		native, _ := nativelyUncached(t, program)
		for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
			if got.exitCode != 70 || !strings.Contains(string(got.stderr), "field write failed:") || !strings.Contains(string(got.stderr), "child") {
				t.Fatalf("narrow original slot admitted undefined: %#v", got)
			}
		}
		t.Logf("Node=%q; actual nonnullable slot rejects widened helper write", truth.stdout)
	})
	for _, site := range []string{"read", "producer", "write"} {
		t.Run(site, func(t *testing.T) {
			program, path := interfaceFixture(t, "nullish/maps/entry-nominal-optional-mutable-"+site)
			truth := onNode(t, path)
			fake, item := -1, -1
			for i, local := range program.Locals {
				if local.Name == "fake" {
					fake = i
				}
				if local.Name == "item" {
					item = i
				}
			}
			if fake < 0 || item < 0 {
				t.Fatal("missing locals")
			}
			changed := false
			for i, statement := range program.Main {
				if site == "write" {
					if store, ok := statement.(ir.SetProperty); ok && store.Name == "child" && !store.Define {
						store.Value = ir.Read{Local: fake, Of: ir.Object}
						program.Main[i] = store
						changed = true
					}
					continue
				}
				if declaration, ok := statement.(ir.Declare); ok && program.Locals[declaration.Local].Name == "source" {
					if site == "producer" {
						creation, ok := declaration.Value.(ir.MapNew)
						if !ok {
							t.Fatal("missing Map")
						}
						creation.Entries[0][1] = ir.ObjectLiteral{Fields: []ir.Field{{Name: "child", Value: ir.Read{Local: fake, Of: ir.Object}}}}
						declaration.Value = creation
						program.Main[i] = declaration
					} else {
						mutation := ir.SetProperty{Object: ir.Read{Local: item, Of: ir.Object}, Name: "child", Value: ir.Read{Local: fake, Of: ir.Object}}
						remaining := append([]ir.Statement(nil), program.Main[i+1:]...)
						program.Main = append(program.Main[:i+1], mutation)
						program.Main = append(program.Main, remaining...)
					}
					changed = true
					break
				}
			}
			if !changed {
				t.Fatal("mutation missed")
			}
			native, _ := nativelyUncached(t, program)
			expected := "Map nominal producer failed:"
			if site == "read" {
				expected = "cast failed: field read failed:"
			}
			for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if got.exitCode != 70 || !strings.Contains(string(got.stderr), expected) || !strings.Contains(string(got.stderr), "class identity") {
					t.Fatalf("optional mutable class lookalike ran on: %#v", got)
				}
			}
			t.Logf("Node=%q; optional mutable lookalike caught at %s", truth.stdout, site)
		})
	}
}

func TestCheckedViewNullableNominalSlotMutants(t *testing.T) {
	t.Run("narrow-slot", func(t *testing.T) {
		program, path := interfaceFixture(t, "nullish/maps/entry-nominal-nullable-slot-null-narrow")
		truth := onNode(t, path)
		native, _ := nativelyUncached(t, program)
		for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
			if got.exitCode != 70 || !strings.Contains(string(got.stderr), "field write failed:") || !strings.Contains(string(got.stderr), "child") {
				t.Fatalf("narrow original slot admitted null: %#v", got)
			}
		}
		t.Logf("Node=%q; actual nonnullable slot rejects widened helper write", truth.stdout)
	})
	for _, site := range []string{"read", "producer", "write"} {
		t.Run(site, func(t *testing.T) {
			program, path := interfaceFixture(t, "nullish/maps/entry-nominal-nullable-slot-null-"+site)
			truth := onNode(t, path)
			fake, item := -1, -1
			for i, local := range program.Locals {
				if local.Name == "fake" {
					fake = i
				}
				if local.Name == "item" {
					item = i
				}
			}
			if fake < 0 || item < 0 {
				t.Fatal("missing locals")
			}
			changed := false
			for i, statement := range program.Main {
				if site == "write" {
					if store, ok := statement.(ir.SetProperty); ok && store.Name == "child" && !store.Define && store.WriteContract != 0 && program.ViewContracts[store.WriteContract-1].Kind != ir.ViewNull {
						store.Value = ir.Read{Local: fake, Of: ir.Object}
						program.Main[i] = store
						changed = true
					}
					continue
				}
				if declaration, ok := statement.(ir.Declare); ok && program.Locals[declaration.Local].Name == "source" {
					if site == "producer" {
						creation, ok := declaration.Value.(ir.MapNew)
						if !ok {
							t.Fatal("missing Map")
						}
						creation.Entries[0][1] = ir.ObjectLiteral{Fields: []ir.Field{{Name: "child", Value: ir.Read{Local: fake, Of: ir.Object}}}}
						declaration.Value = creation
						program.Main[i] = declaration
					} else {
						mutation := ir.SetProperty{Object: ir.Read{Local: item, Of: ir.Object}, Name: "child", Value: ir.Read{Local: fake, Of: ir.Object}}
						remaining := append([]ir.Statement(nil), program.Main[i+1:]...)
						program.Main = append(program.Main[:i+1], mutation)
						program.Main = append(program.Main, remaining...)
					}
					changed = true
					break
				}
			}
			if !changed {
				t.Fatal("mutation missed")
			}
			native, _ := nativelyUncached(t, program)
			expected := "Map nominal producer failed:"
			if site == "read" {
				expected = "cast failed: field read failed:"
			}
			for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if got.exitCode != 70 || !strings.Contains(string(got.stderr), expected) || !strings.Contains(string(got.stderr), "class identity") {
					t.Fatalf("nullable mutable class lookalike ran on: %#v", got)
				}
			}
			t.Logf("Node=%q; nullable mutable lookalike caught at %s", truth.stdout, site)
		})
	}
}

func TestCheckedViewNullableNominalSlotBoundaryRefusals(t *testing.T) {
	for _, name := range []string{"null-helper", "both-helper", "both-optional-helper"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "stage3/interface-downcasts/nullish/maps/entry-nominal-nullable-slot-"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			_, err = lowered(t, path)
			if err == nil || !strings.Contains(err.Error(), "nullable nominal value from an unboxed reference") {
				t.Fatalf("unboxed boundary escaped: %v", err)
			}
			t.Logf("Node=%q; named refusal=%v", truth.stdout, err)
		})
	}
}

func TestCheckedViewNominalFieldWidenReadMutants(t *testing.T) {
	for _, variant := range []string{"null", "both"} {
		t.Run(variant, func(t *testing.T) {
			program, path := interfaceFixture(t, "nullish/maps/entry-nominal-slot-widen-"+variant)
			truth := onNode(t, path)
			fake, item := -1, -1
			for i, local := range program.Locals {
				if local.Name == "fake" {
					fake = i
				}
				if local.Name == "item" {
					item = i
				}
			}
			if fake < 0 || item < 0 {
				t.Fatal("missing locals")
			}
			changed := false
			for i, statement := range program.Main {
				if declaration, ok := statement.(ir.Declare); ok && program.Locals[declaration.Local].Name == "source" {
					mutation := ir.SetProperty{Object: ir.Read{Local: item, Of: ir.Object}, Name: "child", Value: ir.Read{Local: fake, Of: ir.Object}}
					remaining := append([]ir.Statement(nil), program.Main[i+1:]...)
					program.Main = append(program.Main[:i+1], mutation)
					program.Main = append(program.Main, remaining...)
					changed = true
					break
				}
			}
			if !changed {
				t.Fatal("mutation missed")
			}
			native, _ := nativelyUncached(t, program)
			for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if got.exitCode != 70 || !strings.Contains(string(got.stderr), "cast failed: field read failed:") || !strings.Contains(string(got.stderr), "class identity") {
					t.Fatalf("widened reference lookalike ran on: %#v", got)
				}
			}
			t.Logf("Node=%q; original Object storage rechecks identity through nullable target", truth.stdout)
		})
	}
}
