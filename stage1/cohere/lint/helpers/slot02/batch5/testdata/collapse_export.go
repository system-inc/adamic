package tailwind

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type Slot02Pair struct {
	Key   string
	Value bool
}
type Slot02RootKinds struct {
	Root  string
	Kinds []Slot02Pair
}
type Slot02Store struct{ Roots []Slot02RootKinds }
type Slot02SortCase struct {
	Length, Capacity int
	Nodes            []int
	Order            []int
	Count            int
	Visited          []string
}
type Slot02Batch5Corpus struct {
	Texts, RootSet, Kinds []string
	Static, Functional    []Slot02Pair
	Stores                []Slot02Store
	Sorts                 []Slot02SortCase
	Want                  string
	ParsedCandidates      int
}

var Slot02ExpectedNodes []*Node
var Slot02SortCalls int
var Slot02SortNodesMatch bool
var Slot02DelegateSort Sort

// Instrument the wrapper's call site only; the real propertySort body is unchanged.
func AdamicSlot02Batch5PropertySort(nodes []*Node) (Sort, []string) {
	Slot02SortCalls++
	Slot02SortNodesMatch = len(nodes) == len(Slot02ExpectedNodes) && (len(nodes) == 0 || &nodes[0] == &Slot02ExpectedNodes[0])
	value, visited := propertySort(nodes)
	Slot02DelegateSort = value
	return value, visited
}

func Slot02SortedValues(values map[string]bool) []string {
	result := []string{}
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func AdamicSlot02Batch5Observe(input []string) []byte {
	// Oracle-process-only boundary controls. Nil declarations still mean present;
	// false functional entries do not. Restore globals before returning.
	staticControl := "slot02-empty-static"
	functionalControl := "slot02-false-functional"
	if _, exists := FrameworkStaticDeclarations[staticControl]; exists {
		panic("control collision")
	}
	if _, exists := frameworkFunctionalRoots[functionalControl]; exists {
		panic("control collision")
	}
	FrameworkStaticDeclarations[staticControl] = nil
	frameworkFunctionalRoots[functionalControl] = false
	defer delete(FrameworkStaticDeclarations, staticControl)
	defer delete(frameworkFunctionalRoots, functionalControl)
	corpus := Slot02Batch5Corpus{Kinds: []string{"static", "functional", "alien", "", "STATIC"}}
	texts := map[string]bool{"": true, "@": true, "@-2xl": true, "@max-foo": true, "@max-": true, "border-b": true, "border-b-": true, "border-b--x": true, "bg-": true, "-bg-x": true, "a-b-c": true, "a-b-c-": true, "a--b": true, "é-😀-x": true, "é-😀-": true, "😀-x": true, "\x00-a": true, "--": true, "-": true, staticControl: true, functionalControl: true}
	for _, text := range input {
		texts[text] = true
		for _, field := range strings.Fields(text) {
			texts[field] = true
			for _, part := range segment(field, ':') {
				texts[part] = true
			}
		}
	}
	roots := map[string]bool{"@": true, "@max": true, "a": true, "a-b": true, "a-b-c": true, "border": true, "border-b": true, "bg": true, "é": true, "é-😀": true, "😀": true, "-": true}
	for root, declarations := range FrameworkStaticDeclarations {
		roots[root] = true
		texts[root] = true
		corpus.Static = append(corpus.Static, Slot02Pair{root, len(declarations) > 0})
	}
	for root, value := range frameworkFunctionalRoots {
		roots[root] = true
		texts[root] = true
		corpus.Functional = append(corpus.Functional, Slot02Pair{root, value})
	}
	sort.Slice(corpus.Static, func(i, j int) bool { return corpus.Static[i].Key < corpus.Static[j].Key })
	sort.Slice(corpus.Functional, func(i, j int) bool { return corpus.Functional[i].Key < corpus.Functional[j].Key })
	registry := NewVariantRegistry()
	registry.RegisterFrameworkVariants(FrameworkVariantRegistrations)
	for _, registration := range registry.Registrations() {
		roots[registration.Name] = true
		texts[registration.Name] = true
	}
	system := &LoadedDesignSystem{variants: registry, theme: NewTheme()}
	for _, text := range Slot02SortedValues(texts) {
		for _, candidate := range ParseCandidate(text, system) {
			corpus.ParsedCandidates++
			texts[candidate.Root] = true
			for _, variant := range candidate.Variants {
				for current := &variant; current != nil; current = current.Variant {
					texts[current.Root] = true
				}
			}
		}
	}
	corpus.Texts = Slot02SortedValues(texts)
	corpus.RootSet = Slot02SortedValues(roots)
	stores := []map[string]map[UtilityKind]bool{
		nil,
		{"flex": nil, "bg": {UtilityKindStatic: true}, "slot02-custom": {UtilityKindStatic: true, UtilityKindFunctional: true}, "": {UtilityKind("alien"): true}},
		{}, {}, {},
	}
	// Real stylesheet ingestion exercises the repository registration boundary.
	collector := &stylesheetCollector{theme: NewTheme(), utilityRoots: map[string]map[UtilityKind]bool{}, staticUtilityNodes: map[string][]*Node{}, customVariants: map[string]bool{}}
	nodes, err := ParseCSS("@utility flex-* { opacity: 1; } @utility bg { display: block; } @utility slot02-custom { color: red; } @utility slot02-custom-* { color: blue; }")
	if err != nil {
		panic(err)
	}
	if err = collector.ingest(nodes, "/fixture.css"); err != nil {
		panic(err)
	}
	stores[4] = collector.utilityRoots
	for _, root := range corpus.Texts {
		stores[2][root] = nil
		stores[3][root] = map[UtilityKind]bool{UtilityKindStatic: false, UtilityKindFunctional: true, UtilityKind("alien"): true, UtilityKind(""): false}
	}
	var want strings.Builder
	for _, store := range stores {
		projection := Slot02Store{Roots: []Slot02RootKinds{}}
		keys := map[string]bool{}
		for root := range store {
			keys[root] = true
		}
		for _, root := range Slot02SortedValues(keys) {
			row := Slot02RootKinds{Root: root, Kinds: []Slot02Pair{}}
			kindKeys := map[string]bool{}
			for kind := range store[root] {
				kindKeys[string(kind)] = true
			}
			for _, kind := range Slot02SortedValues(kindKeys) {
				row.Kinds = append(row.Kinds, Slot02Pair{kind, store[root][UtilityKind(kind)]})
			}
			projection.Roots = append(projection.Roots, row)
		}
		corpus.Stores = append(corpus.Stores, projection)
		system.utilityRoots = store
		for _, root := range corpus.Texts {
			for _, kind := range corpus.Kinds {
				fmt.Fprintf(&want, "utility:%t\n", system.HasUtility(root, UtilityKind(kind)))
			}
		}
	}
	for _, text := range corpus.Texts {
		for mode := 0; mode < 4; mode++ {
			calls := []string{}
			exists := func(root string) bool {
				calls = append(calls, root)
				switch mode {
				case 0:
					return roots[root]
				case 1:
					return true
				case 2:
					return len(calls)%2 == 1
				default:
					return false
				}
			}
			candidates := findRoots(text, exists)
			fmt.Fprintf(&want, "roots:%d\n", len(candidates))
			for _, candidate := range candidates {
				fmt.Fprintf(&want, "root:%s:%t:%s\n", candidate.root, candidate.hasValue, candidate.value)
			}
			fmt.Fprintf(&want, "calls:%d\n", len(calls))
			for _, call := range calls {
				fmt.Fprintf(&want, "call:%s\n", call)
			}
		}
	}
	// Dependency inputs are computed by the real private traversal. The wrapper
	// itself is independently called and checked for call, value and slice semantics.
	sortTrees := [][]*Node{nil, {}, {Declaration("display", "block"), Declaration("display", "flex")}}
	for _, text := range input {
		if tree, err := ParseCSS(text); err == nil {
			sortTrees = append(sortTrees, tree)
		}
	}
	for _, text := range []string{"display: block; opacity:; color: red;", "a { display: block; --tw-sort: width; } color: red;", "--tw-sort: not-a-property; opacity: 1; --tw-sort: width; color: red;", "@supports (display: grid) { display: grid; } opacity: 1;"} {
		tree, err := ParseCSS(text)
		if err != nil {
			panic(err)
		}
		sortTrees = append(sortTrees, tree)
	}
	for _, pair := range corpus.Static {
		sortTrees = append(sortTrees, nodesFromStaticDeclarations(FrameworkStaticDeclarations[pair.Key]))
	}
	for _, tree := range sortTrees {
		dependency, visited := propertySort(tree)
		row := Slot02SortCase{Nodes: []int{}, Order: append([]int{}, dependency.Order[:cap(dependency.Order)]...), Length: len(dependency.Order), Capacity: cap(dependency.Order), Count: dependency.Count, Visited: append([]string{}, visited...)}
		for index := range tree {
			row.Nodes = append(row.Nodes, index)
		}
		corpus.Sorts = append(corpus.Sorts, row)
		Slot02ExpectedNodes = tree
		Slot02SortCalls = 0
		Slot02SortNodesMatch = false
		result := PropertySort(tree)
		fmt.Fprintf(&want, "sort:%d:%v\ntrace:%d:%t\n", result.Count, result.Order, Slot02SortCalls, Slot02SortNodesMatch)
		Slot02DelegateSort.Count += 1000
		if len(Slot02DelegateSort.Order) > 0 {
			Slot02DelegateSort.Order[0] += 1000
		}
		fmt.Fprintf(&want, "alias:%d:%v\n", result.Count, result.Order)
		Slot02DelegateSort.Order = Slot02DelegateSort.Order[:0:0]
		fmt.Fprintf(&want, "header:%d:%d:%v\n", len(result.Order), cap(result.Order), result.Order)
	}
	corpus.Want = want.String()
	data, err := json.Marshal(corpus)
	if err != nil {
		panic(err)
	}
	return data
}
