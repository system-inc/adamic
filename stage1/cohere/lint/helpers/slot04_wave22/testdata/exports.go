package tailwind

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
)

type AdamicThemeEntry struct {
	Key, Value string
	Options    int
}
type AdamicCase struct {
	Name, Source, OptionRaw string
	Candidate               *ParsedCandidate
	Definitions             map[string]*UtilityDefinition
	Entries                 []AdamicThemeEntry
	Prefix                  string
}
type AdamicAdapted struct {
	Name, OptionRaw, Root, Definition, Walked   string
	Present, Functional, ModifierPresent, Found bool
	Flags                                       int
	Dropped, NonRatio                           []int
	Variants                                    []string
	Valid, ObjectPresent                        bool
	Keys, Sorted                                []string
}

func adamicJSON(v any) string {
	b, e := json.Marshal(v)
	if e != nil {
		panic(e)
	}
	return string(b)
}
func adamicEvaluator(c AdamicCase) *UtilityEvaluator {
	t := NewTheme()
	t.Prefix = c.Prefix
	for _, v := range c.Entries {
		t.values[v.Key] = themeValue{value: v.Value, options: ThemeOptions(v.Options)}
	}
	return &UtilityEvaluator{Theme: t, Definitions: c.Definitions}
}
func adamicIDs(nodes []*Node) map[*Node]int {
	ids := map[*Node]int{}
	var visit func([]*Node)
	visit = func(ns []*Node) {
		for _, n := range ns {
			ids[n] = len(ids)
			visit(n.Nodes)
		}
	}
	visit(nodes)
	return ids
}
func adamicSet(set map[*Node]bool, ids map[*Node]int) []int {
	out := []int{}
	for n, v := range set {
		if v {
			out = append(out, ids[n])
		}
	}
	sort.Ints(out)
	return out
}
func AdamicAdapt(c AdamicCase) AdamicAdapted {
	a := AdamicAdapted{Name: c.Name, OptionRaw: c.OptionRaw, Dropped: []int{}, NonRatio: []int{}, Variants: []string{}, Keys: []string{}, Sorted: []string{}, Definition: "null", Walked: "null"}
	var object map[string]json.RawMessage
	e := json.Unmarshal([]byte(c.OptionRaw), &object)
	a.Valid = e == nil
	a.ObjectPresent = object != nil
	for k := range object {
		a.Keys = append(a.Keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(a.Keys)))
	a.Sorted = append(a.Sorted, a.Keys...)
	sort.Strings(a.Sorted)
	if c.Candidate == nil {
		return a
	}
	a.Present = true
	a.Functional = c.Candidate.Kind == ParsedCandidateKindFunctional
	a.Root = c.Candidate.Root
	a.ModifierPresent = c.Candidate.Modifier != nil
	d, found := c.Definitions[a.Root]
	a.Found = found
	if !found || !a.Functional {
		return a
	}
	a.Definition = adamicJSON(d.Nodes)
	body := cloneNodes(d.Nodes)
	s := &utilityEvaluation{evaluator: adamicEvaluator(c), value: c.Candidate.Value, modifier: c.Candidate.Modifier}
	s.walk(body)
	a.Walked = adamicJSON(body)
	ids := adamicIDs(body)
	a.Dropped = adamicSet(s.droppedDeclarations, ids)
	a.NonRatio = adamicSet(s.nonRatioDeclarations, ids)
	if s.usedValueFunction {
		a.Flags |= 1
	}
	if s.resolvedValueFunction {
		a.Flags |= 2
	}
	if s.usedModifierFunction {
		a.Flags |= 4
	}
	if s.resolvedModifierFunction {
		a.Flags |= 8
	}
	if s.resolvedRatioValue {
		a.Flags |= 16
	}
	for mask := 0; mask < 4; mask++ {
		cloned := cloneNodes(body)
		clonedIDs := adamicIDs(cloned)
		removed := map[*Node]bool{}
		for n, id := range clonedIDs {
			if mask&1 != 0 {
				for _, x := range a.Dropped {
					if x == id {
						removed[n] = true
					}
				}
			}
			if mask&2 != 0 {
				for _, x := range a.NonRatio {
					if x == id {
						removed[n] = true
					}
				}
			}
		}
		a.Variants = append(a.Variants, adamicJSON(removeNodes(cloned, removed)))
	}
	return a
}
func AdamicObserve(c AdamicCase) {
	e := adamicEvaluator(c)
	before := adamicJSON(c.Definitions)
	for mode := 0; mode < 4; mode++ {
		var nodes []*Node
		var ok bool
		if mode == 3 {
			nodes, ok = e.Compile(c.Candidate)
		} else {
			nodes, ok = e.compile(c.Candidate, utilityRemovals(mode))
		}
		fmt.Printf("%t|%s\n", ok, adamicJSON(nodes))
		if adamicJSON(c.Definitions) != before {
			panic("definition mutated")
		}
	}
}

var adamicLock sync.Mutex

func AdamicRecordCompile(e *UtilityEvaluator, c *ParsedCandidate) {
	path := os.Getenv("ADAMIC_SLOT04_COMPILE")
	if path == "" {
		return
	}
	adamicLock.Lock()
	defer adamicLock.Unlock()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	row := AdamicCase{Candidate: c, Definitions: e.Definitions, OptionRaw: "{}", Prefix: e.Theme.Prefix}
	for k, v := range e.Theme.values {
		row.Entries = append(row.Entries, AdamicThemeEntry{k, v.value, int(v.options)})
	}
	sort.Slice(row.Entries, func(i, j int) bool { return row.Entries[i].Key < row.Entries[j].Key })
	if err = json.NewEncoder(f).Encode(row); err != nil {
		panic(err)
	}
}
