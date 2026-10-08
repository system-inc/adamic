package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

func TestCheckedViewMapCallableProducerMutants(t *testing.T) {
	for _, site := range []string{"constructor", "set"} {
		t.Run(site, func(t *testing.T) {
			program, path := interfaceFixture(t, "nullish/maps/entry-callable-producer")
			truth := onNode(t, path)
			original, _ := nativelyUncached(t, program)
			if diff := disagreement(truth, original); diff != "" {
				t.Fatal(diff)
			}
			wrong := -1
			for index, local := range program.Locals {
				if local.Name == "wrong" {
					wrong = index
				}
			}
			if wrong < 0 {
				t.Fatal("missing wrong producer")
			}
			changed := false
			for index, statement := range program.Main {
				if site == "constructor" {
					if declaration, ok := statement.(ir.Declare); ok {
						if creation, ok := declaration.Value.(ir.MapNew); ok && len(creation.Entries) > 0 {
							creation.Entries[0][1] = ir.Read{Local: wrong, Of: ir.Closure}
							declaration.Value = creation
							program.Main[index] = declaration
							changed = true
							break
						}
					}
				} else if evaluation, ok := statement.(ir.Evaluate); ok {
					if store, ok := evaluation.Value.(ir.MapSet); ok {
						store.Value = ir.Read{Local: wrong, Of: ir.Closure}
						evaluation.Value = store
						program.Main[index] = evaluation
						changed = true
						break
					}
				}
			}
			if !changed {
				t.Fatal("producer mutation missed storage site")
			}
			native, _ := nativelyUncached(t, program)
			for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if got.exitCode != 70 || !strings.Contains(string(got.stderr), "incompatible result representation") {
					t.Fatalf("wrong physical producer ran on: %#v", got)
				}
			}
			t.Logf("Node control stdout=%q; mismatched actual producer stopped at %s", truth.stdout, site)
		})
	}
}

func TestCheckedViewMapTuplePayloadMutant(t *testing.T) {
	program, path := interfaceFixture(t, "nullish/maps/entry-tuple")
	truth := onNode(t, path)
	native, _ := nativelyUncached(t, program)
	if diff := disagreement(truth, native); diff != "" {
		t.Fatal(diff)
	}
	changed := false
	for index, statement := range program.Main {
		if declaration, ok := statement.(ir.Declare); ok && program.Locals[declaration.Local].Name == "pair" {
			literal, ok := declaration.Value.(ir.ObjectLiteral)
			if !ok || len(literal.Fields) != 2 {
				t.Fatal("tuple storage changed")
			}
			literal.Fields[0].Value = literal.Fields[1].Value
			declaration.Value = literal
			program.Main[index] = declaration
			changed = true
			break
		}
	}
	if !changed {
		t.Fatal("tuple payload mutant missed")
	}
	native, _ = nativelyUncached(t, program)
	for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || !strings.Contains(string(got.stderr), "pair[0]") {
			t.Fatalf("tuple helper trusted malformed position: %#v", got)
		}
	}
	t.Logf("Node control=%q; wrong string payload stopped at helper pair[0]", truth.stdout)
}

func TestCheckedViewMapStorageMetadataMutant(t *testing.T) {
	program, path := interfaceFixture(t, "nullish/maps/entry-convert-number-undefined")
	truth := onNode(t, path)
	changed := false
	for index, statement := range program.Main {
		if declaration, ok := statement.(ir.Declare); ok {
			if creation, ok := declaration.Value.(ir.MapNew); ok && program.Locals[declaration.Local].Name == "source" {
				creation.Value = ir.Boolean
				declaration.Value = creation
				program.Main[index] = declaration
				changed = true
				break
			}
		}
	}
	if !changed {
		t.Fatal("storage metadata mutation missed")
	}
	native, _ := nativelyUncached(t, program)
	for _, got := range []run{native, releasedUncached(t, program)} {
		if got.exitCode != 70 || !strings.Contains(string(got.stderr), "Map value storage cannot be converted") {
			t.Fatalf("unavailable read conversion ran on: %#v", got)
		}
	}
	// JavaScript has no physical storage reinterpretation in this mutant.
	if diff := disagreement(truth, onJavaScriptBackend(t, program)); diff != "" {
		t.Fatal(diff)
	}
	t.Logf("incompatible physical Boolean producer stops before a Number|undefined slot read; Node=%q", truth.stdout)
}

func TestCheckedViewMapUndefinedStorageMutant(t *testing.T) {
	program, path := interfaceFixture(t, "nullish/maps/entry-brand-nullish")
	truth := onNode(t, path)
	changed := false
	for index, statement := range program.Main {
		if declaration, ok := statement.(ir.Declare); ok && program.Locals[declaration.Local].Name == "missing" {
			declaration.Value = ir.ObjectLiteral{Fields: []ir.Field{{Name: "count", Value: ir.NumberConstant{Value: 7}}}}
			program.Main[index] = declaration
			changed = true
			break
		}
	}
	if !changed {
		t.Fatal("undefined producer mutant missed")
	}
	native, _ := nativelyUncached(t, program)
	for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || !strings.Contains(string(got.stderr), "undefined storage conversion failed: missing") {
			t.Fatalf("present payload stored as branded undefined: %#v", got)
		}
	}
	t.Logf("present object producer refused before undefined conversion; Node control=%q", truth.stdout)
}

func TestCheckedViewMapNestedPayloadMutant(t *testing.T) {
	for _, name := range []string{"entry-convert-nested-array", "entry-convert-nested-union", "entry-convert-nested-boolean-optional"} {
		t.Run(name, func(t *testing.T) { checkedViewMapNestedPayloadMutant(t, name) })
	}
}

func checkedViewMapNestedPayloadMutant(t *testing.T, name string) {
	program, path := interfaceFixture(t, "nullish/maps/"+name)
	truth := onNode(t, path)
	changed := false
	for index, statement := range program.Main {
		if declaration, ok := statement.(ir.Declare); ok && program.Locals[declaration.Local].Name == "item" {
			literal, ok := declaration.Value.(ir.ArrayLiteral)
			if !ok {
				t.Fatal("array producer changed")
			}
			literal.Element = ir.String
			constant := len(program.Strings)
			program.Strings = append(program.Strings, "wrong")
			for index := range literal.Elements {
				literal.Elements[index] = ir.StringConstant{Index: constant}
			}
			declaration.Value = literal
			program.Main[index] = declaration
			changed = true
			break
		}
	}
	if !changed {
		t.Fatal("array payload mutant missed")
	}
	native, _ := nativelyUncached(t, program)
	for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || !strings.Contains(string(got.stderr), "values[element]") {
			t.Fatalf("nested helper trusted wrong element: %#v", got)
		}
	}
	t.Logf("Node=%q; wrong string payload stops at helper values[element]", truth.stdout)
}

func TestCheckedViewMapKeyMetadataMutant(t *testing.T) {
	program, path := interfaceFixture(t, "nullish/maps/entry-convert-key")
	truth := onNode(t, path)
	changed := false
	for index, statement := range program.Main {
		if declaration, ok := statement.(ir.Declare); ok && program.Locals[declaration.Local].Name == "source" {
			creation, ok := declaration.Value.(ir.MapNew)
			if !ok {
				t.Fatal("Map producer changed")
			}
			creation.Key = ir.Boolean
			for index := range creation.Entries {
				creation.Entries[index][0] = ir.BooleanConstant{Value: index != 0}
			}
			declaration.Value = creation
			program.Main[index] = declaration
			changed = true
			break
		}
	}
	if !changed {
		t.Fatal("key storage mutation missed")
	}
	native, _ := nativelyUncached(t, program)
	for _, got := range []run{native, releasedUncached(t, program)} {
		if got.exitCode != 70 || !strings.Contains(string(got.stderr), "Map key lookup storage cannot be converted; expected boolean, found number | undefined") {
			t.Fatalf("wrong physical key producer ran on: %#v", got)
		}
	}
	t.Logf("Boolean storage with a Number producer certificate stops before lookup; Node=%q", truth.stdout)
}

func TestCheckedViewMapNominalProducerMutants(t *testing.T) {
	for _, key := range []bool{false, true} {
		for _, site := range []string{"constructor", "set", "copy"} {
			name := site
			if key {
				name = "key-" + site
			}
			t.Run(name, func(t *testing.T) {
				fixture := "entry-nominal"
				if key {
					fixture = "entry-nominal-key"
				}
				program, path := interfaceFixture(t, "nullish/maps/"+fixture)
				truth := onNode(t, path)
				fake := -1
				for index, local := range program.Locals {
					if local.Name == "fake" {
						fake = index
					}
				}
				if fake < 0 {
					t.Fatal("missing lookalike producer")
				}
				position := 1
				if key {
					position = 0
				}
				replacement := ir.Read{Local: fake, Of: ir.Object}
				changed := false
				for index, statement := range program.Main {
					if declaration, ok := statement.(ir.Declare); ok {
						creation, ok := declaration.Value.(ir.MapNew)
						if !ok {
							continue
						}
						if site == "constructor" && program.Locals[declaration.Local].Name == "source" {
							creation.Entries[0][position] = replacement
						} else if site == "copy" && program.Locals[declaration.Local].Name == "copy" {
							elements := []ir.Expression{ir.StringConstant{Index: 0}, ir.NumberConstant{Value: 7}}
							if key {
								elements[0] = replacement
							} else {
								elements[1] = replacement
							}
							pair := ir.ObjectLiteral{Fields: []ir.Field{{Name: "0", Value: elements[0]}, {Name: "1", Value: elements[1]}}}
							creation.Pairs = ir.ArrayLiteral{Element: ir.Object, Elements: []ir.Expression{pair}}
						} else {
							continue
						}
						declaration.Value = creation
						program.Main[index] = declaration
						changed = true
						break
					} else if evaluation, ok := statement.(ir.Evaluate); ok && site == "set" {
						if store, ok := evaluation.Value.(ir.MapSet); ok {
							if key {
								store.Key = replacement
							} else {
								store.Value = replacement
							}
							evaluation.Value = store
							program.Main[index] = evaluation
							changed = true
							break
						}
					}
				}
				if !changed {
					t.Fatal("nominal producer mutation missed")
				}
				native, _ := nativelyUncached(t, program)
				for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if got.exitCode != 70 || !strings.Contains(string(got.stderr), "Map nominal producer failed:") || !strings.Contains(string(got.stderr), "class identity") {
						t.Fatalf("structural lookalike acquired a nominal certificate: %#v", got)
					}
				}
				t.Logf("Node=%q; structural lookalike caught at %s", truth.stdout, name)
			})
		}
	}
}

func TestCheckedViewMapNestedUnionLiteralMutant(t *testing.T) {
	for _, name := range []string{"entry-convert-nested-union-finite", "entry-convert-nested-boolean-finite"} {
		t.Run(name, func(t *testing.T) { checkedViewMapNestedLiteralMutant(t, name) })
	}
}

func checkedViewMapNestedLiteralMutant(t *testing.T, name string) {
	program, path := interfaceFixture(t, "nullish/maps/"+name)
	truth := onNode(t, path)
	changed := false
	for index, statement := range program.Main {
		if declaration, ok := statement.(ir.Declare); ok && program.Locals[declaration.Local].Name == "item" {
			literal, ok := declaration.Value.(ir.ArrayLiteral)
			if !ok {
				t.Fatal("array producer changed")
			}
			literal.Elements[0] = ir.NumberConstant{Value: 2}
			if strings.Contains(name, "boolean") {
				literal.Elements[0] = ir.BooleanConstant{Value: false}
			}
			declaration.Value = literal
			program.Main[index] = declaration
			changed = true
			break
		}
	}
	if !changed {
		t.Fatal("finite payload mutant missed")
	}
	native, _ := nativelyUncached(t, program)
	for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || !strings.Contains(string(got.stderr), "values[0]") {
			t.Fatalf("finite union helper trusted wrong number: %#v", got)
		}
	}
	t.Logf("Node=%q; nonmember payload with unchanged finite declaration stopped at helper values[0]", truth.stdout)
}
