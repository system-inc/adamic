package lower

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

// Refresh an intended answer change with:
// go test ./internal/lower -run '^TestCyclesCorpus$' -args -update
var updateCycles = flag.Bool("update", false, "rewrite the recorded cycle finder corpus with current answers")

const cyclesCorpusPath = "testdata/cycles.json"

// The initial answers were recorded from the frozen finder at
// 6c891cca313e614ae46de794814bc7c0db4cae01, over every sameness-test entry.
// Each entry runs in a subprocess to release its checker and relation caches.
// Keep failed loads and pre-finder refusals too: the inventory must not silently shrink.
type cycleAnswer struct {
	Types  []string            `json:"types,omitempty"`
	Status string              `json:"status"`
	Error  string              `json:"error,omitempty"`
	Slots  []cycleSlotAnswer   `json:"slots,omitempty"`
	Graph  map[string][]string `json:"graph,omitempty"`
}
type cycleSlotAnswer struct {
	Holder   string       `json:"holder"`
	Slot     string       `json:"slot"`
	Target   string       `json:"target"`
	Weak     bool         `json:"weak"`
	Reaches  bool         `json:"cycle_capable"`
	Unproven *fresh.Write `json:"unproven,omitempty"`
	Refusal  string       `json:"refusal,omitempty"`
}

func cyclesCorpusEntries(t *testing.T) []string {
	t.Helper()
	return cyclesCorpusEntriesUnder(t, []string{"../oracle", "../../stage1"})
}

func cyclesCorpusEntriesUnder(t *testing.T, roots []string) []string {
	t.Helper()
	var paths []string
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				// Generated lint registries are ignored build artifacts, not corpus sources.
				if entry.Name() == "node_modules" || entry.Name() == ".generated" {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(path, ".a") || strings.HasSuffix(path, ".ts") && !strings.HasSuffix(path, ".d.ts") {
				paths = append(paths, filepath.ToSlash(path))
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	sort.Strings(paths)
	return paths
}

func TestCyclesCorpusInventoryIgnoresGenerated(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{"fixture.a", "source.ts", "types.d.ts", ".generated/registry.ts", "node_modules/library.ts"} {
		absolute := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(absolute), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(absolute, []byte("console.log('fixture');\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	wanted := []string{filepath.ToSlash(filepath.Join(root, "fixture.a")), filepath.ToSlash(filepath.Join(root, "source.ts"))}
	if got := cyclesCorpusEntriesUnder(t, []string{root}); !reflect.DeepEqual(got, wanted) {
		t.Fatalf("corpus includes generated or dependency sources: got %v; want %v", got, wanted)
	}
}

func cyclesReadCorpus(t *testing.T) map[string]cycleAnswer {
	t.Helper()
	data, err := os.ReadFile(cyclesCorpusPath)
	if err != nil {
		t.Fatal(err)
	}
	var corpus map[string]cycleAnswer
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	return corpus
}
func cyclesPrepareRegistry(t *testing.T, root string) {
	t.Helper()
	if _, err := registry.Generate(root); err != nil {
		t.Fatal(err)
	}
}

func TestCyclesCorpusPreparesGeneratedDependency(t *testing.T) {
	root := t.TempDir()
	rule := filepath.Join(root, "rules", "no-debugger")
	if err := os.MkdirAll(rule, 0755); err != nil {
		t.Fatal(err)
	}
	source := "../../stage1/cohere/lint/rules/no-debugger"
	if err := filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(rule, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0644)
	}); err != nil {
		t.Fatal(err)
	}
	cyclesPrepareRegistry(t, root)
	generated := filepath.Join(root, ".generated", "registry.ts")
	wanted, err := os.ReadFile(generated)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(wanted), "../rules/no-debugger/rule.ts") {
		t.Fatal("generated registry lost fixture rule")
	}
	if err := os.WriteFile(generated, []byte("stale registry"), 0644); err != nil {
		t.Fatal(err)
	}
	cyclesPrepareRegistry(t, root)
	got, err := os.ReadFile(generated)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, wanted) {
		t.Fatal("corpus depends on a stale generated dependency")
	}
}

func TestCyclesCorpus(t *testing.T) {
	// lint.ts imports its generated dispatch even though that artifact is not an entry.
	cyclesPrepareRegistry(t, "../../stage1/cohere/lint")
	paths := cyclesCorpusEntries(t)
	var wanted map[string]cycleAnswer
	if !*updateCycles {
		wanted = cyclesReadCorpus(t)
	}
	recorded := map[string]cycleAnswer{}
	reached, ports := 0, 0
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			command := exec.Command(os.Args[0], "-test.run=^TestCyclesCorpusFixture$", "-test.timeout=10m")
			destination := filepath.Join(t.TempDir(), "answer.json")
			command.Env = append(os.Environ(), "CYCLES_FIXTURE="+path, "CYCLES_ANSWER="+destination, "GOMEMLIMIT=6GiB")
			output, err := command.CombinedOutput()
			if err != nil {
				t.Fatalf("%v\n%s", err, output)
			}
			data, err := os.ReadFile(destination)
			if err != nil {
				t.Fatal(err)
			}
			var answer cycleAnswer
			if err := json.Unmarshal(data, &answer); err != nil {
				t.Fatal(err)
			}
			recorded[path] = answer
			if answer.Status == "finder" {
				reached++
				if strings.HasPrefix(path, "../../stage1/") {
					ports++
				}
			}
			if !*updateCycles {
				prior, ok := wanted[path]
				if !ok || !reflect.DeepEqual(prior, answer) {
					before, _ := json.Marshal(prior)
					after, _ := json.Marshal(answer)
					at := 0
					for at < len(before) && at < len(after) && before[at] == after[at] {
						at++
					}
					snippet := func(data []byte) string {
						start := max(0, at-80)
						end := min(len(data), at+240)
						if start > end {
							start = end
						}
						return string(data[start:end])
					}
					t.Fatalf("cycle corpus differs for %s at byte %d\nrecorded: %s\ncurrent: %s", path, at, snippet(before), snippet(after))
				}
			}
		})
	}
	if len(recorded) == len(paths) && (reached == 0 || ports == 0) {
		t.Fatal("corpus did not exercise both fixtures and stage1 ports")
	}
	if !*updateCycles && len(wanted) != len(paths) {
		t.Fatalf("corpus inventory differs: recorded=%d current=%d; use -update", len(wanted), len(paths))
	}
	if *updateCycles && len(recorded) != len(paths) {
		t.Fatal("-update requires the complete corpus; do not filter subtests")
	}
	if *updateCycles && !t.Failed() {
		data, err := json.MarshalIndent(recorded, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(cyclesCorpusPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(cyclesCorpusPath, append(data, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("%d entries recorded, %d finder runs including %d stage1 entries", len(paths), reached, ports)
}

func TestCyclesCorpusFixture(t *testing.T) {
	path := os.Getenv("CYCLES_FIXTURE")
	if path == "" {
		t.Skip("run through TestCyclesCorpus")
	}
	answer := cycleAnswer{Status: "load"}
	program, err := load.Load([]string{path})
	if err == nil {
		answer.Status = "lower"
		_, err = cyclesProbeLower(context.Background(), program, func(l *lowering, modules []*ast.SourceFile) error {
			answer.Status = "finder"
			finder := l.cycleTypes(modules)
			err := finder.graphTypes(modules)
			seen, shapes, where := finder.seen, finder.shapes, finder.where
			slots, weak, reaches, unproven, links := finder.slotsOf, finder.weak, finder.reaches, finder.cycleUnproven, finder.graphLinks
			describe := func(proven *checker.Type) string {
				if proven == nil {
					return ""
				}
				// Render modifiers as well as the type text: readonly/optional slots can
				// have identical value types, but different ownership behavior.
				// Higher ObjectFlags bits describe lazily populated checker caches.
				name := fmt.Sprintf("%s [flags=%d object=%d]", l.checker.TypeToString(proven), proven.Flags(), proven.ObjectFlags()&0xffff)
				for _, field := range l.checker.GetPropertiesOfType(proven) {
					name += fmt.Sprintf(";%s:optional=%t,readonly=%t", field.Name, field.Flags&ast.SymbolFlagsOptional != 0, l.checker.IsReadonlySymbol(field))
				}
				return name
			}
			check := func(holder, target *checker.Type, kind fresh.WriteKind, name string, refusal string) {
				slot := cycleSlotAnswer{Holder: describe(holder), Slot: name, Target: describe(target), Weak: weak(target), Reaches: reaches(target, cycleNode{proven: holder}), Refusal: refusal}
				if !slot.Weak && slot.Reaches {
					writeName := name
					if kind != fresh.WriteField {
						writeName = ""
					}
					slot.Unproven = unproven(kind, holder, writeName)
				}
				answer.Slots = append(answer.Slots, slot)
			}
			for _, holder := range seen {
				if holder.Flags()&checker.TypeFlagsObject == 0 || finder.isFunction(holder) {
					continue
				}
				refusal := cyclesError(slots(holder))
				switch {
				case l.checker.IsArrayType(holder) || checker.IsTupleType(holder):
					for _, target := range l.checker.GetTypeArguments(holder) {
						check(holder, target, fresh.WriteElement, "[]", refusal)
					}
				case l.isLibraryType(holder, "Map", "ReadonlyMap"):
					for i, target := range l.checker.GetTypeArguments(holder) {
						check(holder, target, fresh.WriteMapEntry, fmt.Sprintf("map[%d]", i), refusal)
					}
				case l.isLibraryType(holder, "Set", "ReadonlySet"):
					for _, target := range l.checker.GetTypeArguments(holder) {
						check(holder, target, fresh.WriteSetElement, "set", refusal)
					}
				default:
					for _, field := range finder.fields(holder) {
						check(holder, l.checker.GetTypeOfSymbol(field), fresh.WriteField, field.Name, refusal)
					}
				}
			}
			for local, declared := range l.result.Locals {
				if !declared.Captured || declared.Global {
					continue
				}
				if proven := l.localTypes[local]; proven != nil {
					answer.Slots = append(answer.Slots, cycleSlotAnswer{Holder: fmt.Sprintf("cell:%d", local+1), Target: describe(proven), Weak: weak(proven), Reaches: reaches(proven, cycleNode{cell: local + 1})})
				}
			}
			// Resolve checker identities by their structural description. Fresh literal
			// views have process-local IDs; preserve their multiplicity, not those IDs.
			types := map[int]string{}
			visited := map[cycleNode]bool{}
			var queue []cycleNode
			for proven := range where {
				queue = append(queue, cycleNode{proven: proven})
			}
			for _, proven := range shapes {
				queue = append(queue, cycleNode{proven: proven})
			}
			for len(queue) > 0 {
				node := queue[len(queue)-1]
				queue = queue[:len(queue)-1]
				if visited[node] {
					continue
				}
				visited[node] = true
				if node.proven != nil {
					types[int(node.proven.Id())] = describe(node.proven)
				}
				queue = append(queue, links(node)...)
			}
			identity := func(id int) string {
				if name, ok := types[id]; ok {
					return name
				}
				if id > 0 {
					t.Fatalf("graph identity %d has no recorded type", id)
				}
				return fmt.Sprintf("allocation:%d", id)
			}
			answer.Graph = map[string][]string{}
			var visit func(reflect.Value, string)
			visit = func(value reflect.Value, path string) {
				switch value.Kind() {
				case reflect.Interface, reflect.Pointer:
					if !value.IsNil() {
						visit(value.Elem(), path)
					}
				case reflect.Struct:
					for i := 0; i < value.NumField(); i++ {
						field := value.Field(i)
						name := value.Type().Field(i).Name
						next := path + "." + name
						if name == "GraphTypes" {
							var selected []string
							if field.Kind() == reflect.Map {
								iter := field.MapRange()
								for iter.Next() {
									selected = append(selected, fmt.Sprintf("%s=%t", identity(int(iter.Key().Int())), iter.Value().Bool()))
								}
							} else {
								for j := 0; j < field.Len(); j++ {
									id := int(field.Index(j).Int())
									if l.result.GraphTypes[id] {
										selected = append(selected, identity(id))
									}
								}
							}
							sort.Strings(selected)
							if len(selected) != 0 {
								answer.Graph[next] = selected
							}
						} else {
							visit(field, next)
						}
					}
				case reflect.Slice, reflect.Array:
					for i := 0; i < value.Len(); i++ {
						visit(value.Index(i), fmt.Sprintf("%s[%d]", path, i))
					}
				}
			}
			visit(reflect.ValueOf(l.result), "IR")
			sort.Slice(answer.Slots, func(i, j int) bool {
				a, _ := json.Marshal(answer.Slots[i])
				b, _ := json.Marshal(answer.Slots[j])
				return string(a) < string(b)
			})
			return err
		})
	}
	answer.Error = cyclesError(err)
	cyclesCompact(&answer)
	data, err := json.Marshal(answer)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(os.Getenv("CYCLES_ANSWER"), data, 0600); err != nil {
		t.Fatal(err)
	}
}

// Diagnostic paths are relative to the repository, independent of checkout root.
func cyclesError(err error) string {
	if err == nil {
		return ""
	}
	root, e := filepath.Abs("../..")
	if e != nil {
		panic(e)
	}
	return strings.ReplaceAll(filepath.ToSlash(err.Error()), filepath.ToSlash(root)+"/", "")
}

// Intern repeated descriptions in a sorted per-program table. References retain
// multiplicity and never depend on checker IDs or map iteration order.
func cyclesCompact(answer *cycleAnswer) {
	names := map[string]bool{}
	for _, slot := range answer.Slots {
		names[slot.Holder] = true
		names[slot.Target] = true
	}
	for _, ids := range answer.Graph {
		for _, id := range ids {
			names[id] = true
		}
	}
	for name := range names {
		answer.Types = append(answer.Types, name)
	}
	sort.Strings(answer.Types)
	refs := map[string]string{}
	for i, name := range answer.Types {
		refs[name] = fmt.Sprintf("type:%d", i)
	}
	for i := range answer.Slots {
		answer.Slots[i].Holder = refs[answer.Slots[i].Holder]
		answer.Slots[i].Target = refs[answer.Slots[i].Target]
	}
	for path, ids := range answer.Graph {
		for i := range ids {
			ids[i] = refs[ids[i]]
		}
		sort.Strings(ids)
		answer.Graph[path] = ids
	}
}
