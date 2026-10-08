package fresh

import (
	"fmt"
	"github.com/system-inc/adamic/internal/flow"
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
	"sort"
	"strings"
)

type Verdict uint8

const (
	Unknown Verdict = iota
	Proven
	Refused
)

func (v Verdict) String() string { return [...]string{"Unknown", "Proven", "Refused"}[v] }

type OwnershipEdge struct {
	From, To int
	Label    string
	Weak     bool
}
type OwnershipObject struct {
	ID, Region            int
	AllocationInstruction flow.InstructionId
	Allocation            *ir.Statement
	Unknown               bool
}
type OwnershipRoot struct {
	Name                      string
	Object                    int
	Weak, Borrowed, UsedAfter bool
	UseInstruction            flow.InstructionId
}

// Complete certifies all root, edge and dynamic region inventories. Type SCCs
// cannot supply region membership; incomplete inventories yield Unknown.
type OwnershipSnapshot struct {
	Objects  []OwnershipObject
	Edges    []OwnershipEdge
	Roots    []OwnershipRoot
	Complete bool
}
type OwnershipWitness struct {
	Root, Path                string
	Weak, Borrowed, UsedAfter bool
	UseInstruction            flow.InstructionId
}
type OwnershipAnswer struct {
	Verdict   Verdict
	Members   []int
	Witnesses []OwnershipWitness
	Evidence  []string
	Snapshot  OwnershipSnapshot
}

// QueryOwnership retains the finite may-reach graph and shortest deterministic
// witnesses. A second surviving owner refuses even without a continuation read.
func QueryOwnership(s OwnershipSnapshot, owner string) OwnershipAnswer {
	out := OwnershipAnswer{Verdict: Proven, Snapshot: s}
	unknown := func(why string) {
		if out.Verdict == Proven {
			out.Verdict = Unknown
		}
		out.Evidence = append(out.Evidence, why)
	}
	refuse := func(why string) { out.Verdict = Refused; out.Evidence = append(out.Evidence, why) }
	if len(s.Objects)+len(s.Edges)+len(s.Roots) > 1000000 {
		unknown("ownership witness budget exhausted")
		return out
	}
	if !s.Complete {
		unknown("incomplete root, edge or region inventory")
	}
	objects := map[int]OwnershipObject{}
	regions := map[int][]int{}
	forward := map[int][]OwnershipEdge{}
	for _, o := range s.Objects {
		if _, ok := objects[o.ID]; ok {
			unknown("duplicate object identity")
		}
		objects[o.ID] = o
		if o.Region != 0 {
			regions[o.Region] = append(regions[o.Region], o.ID)
		}
	}
	for _, e := range s.Edges {
		forward[e.From] = append(forward[e.From], e)
	}
	for id := range forward {
		sort.Slice(forward[id], func(i, j int) bool {
			a, b := forward[id][i], forward[id][j]
			if a.Label != b.Label {
				return a.Label < b.Label
			}
			if a.To != b.To {
				return a.To < b.To
			}
			return !a.Weak && b.Weak
		})
	}
	var selected *OwnershipRoot
	for i := range s.Roots {
		if s.Roots[i].Name == owner {
			if selected != nil {
				unknown("ambiguous candidate root")
			}
			selected = &s.Roots[i]
		}
	}
	if selected == nil {
		unknown("candidate root missing")
		return out
	}
	members := map[int]bool{}
	expandedRegions := map[int]bool{}
	queue := []int{selected.Object}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if members[id] {
			continue
		}
		members[id] = true
		o, ok := objects[id]
		if !ok || o.Unknown {
			unknown(fmt.Sprintf("unresolved object %d", id))
		}
		for _, e := range forward[id] {
			if !e.Weak {
				queue = append(queue, e.To)
			}
		}
		if o.Region != 0 && !expandedRegions[o.Region] {
			expandedRegions[o.Region] = true
			queue = append(queue, regions[o.Region]...)
		}
	}
	for id := range members {
		out.Members = append(out.Members, id)
	}
	sort.Ints(out.Members)

	// Index backwards once rather than walking the heap separately for every root.
	// Two states preserve strong ownership separately from weak observation.
	reverse := map[int][]OwnershipEdge{}
	for _, e := range s.Edges {
		reverse[e.To] = append(reverse[e.To], e)
	}
	for id := range reverse {
		sort.Slice(reverse[id], func(i, j int) bool {
			a, b := reverse[id][i], reverse[id][j]
			if a.From != b.From {
				return a.From < b.From
			}
			if a.Label != b.Label {
				return a.Label < b.Label
			}
			return !a.Weak && b.Weak
		})
	}
	type stateKey struct {
		id   int
		weak bool
	}
	type predecessor struct {
		next     stateKey
		label    string
		distance int
	}
	paths := map[stateKey]predecessor{}
	pending := []stateKey{}
	for _, id := range out.Members {
		key := stateKey{id, false}
		paths[key] = predecessor{}
		pending = append(pending, key)
	}
	for len(pending) > 0 {
		at := pending[0]
		pending = pending[1:]
		for _, e := range reverse[at.id] {
			key := stateKey{e.From, at.weak || e.Weak}
			if _, seen := paths[key]; seen {
				continue
			}
			paths[key] = predecessor{at, e.Label, paths[at].distance + 1}
			pending = append(pending, key)
		}
	}
	// An opaque frontier reachable from any root independently prevents proof.
	unknownReach := map[int]bool{}
	unknownQueue := []int{}
	seedUnknown := func(id int) {
		o, ok := objects[id]
		if (!ok || o.Unknown) && !unknownReach[id] {
			unknownReach[id] = true
			unknownQueue = append(unknownQueue, id)
		}
	}
	for _, o := range s.Objects {
		seedUnknown(o.ID)
	}
	for _, e := range s.Edges {
		seedUnknown(e.From)
		seedUnknown(e.To)
	}
	for _, r := range s.Roots {
		seedUnknown(r.Object)
	}
	for len(unknownQueue) > 0 {
		id := unknownQueue[0]
		unknownQueue = unknownQueue[1:]
		for _, e := range reverse[id] {
			if !unknownReach[e.From] {
				unknownReach[e.From] = true
				unknownQueue = append(unknownQueue, e.From)
			}
		}
	}
	roots := append([]OwnershipRoot(nil), s.Roots...)
	sort.Slice(roots, func(i, j int) bool { return roots[i].Name < roots[j].Name })
	for _, root := range roots {
		if unknownReach[root.Object] {
			unknown("unresolved frontier reachable from " + root.Name)
		}
		key := stateKey{root.Object, false}
		strong, strongOK := paths[key]
		weakKey := stateKey{root.Object, true}
		weak, weakOK := paths[weakKey]
		if !strongOK && !weakOK {
			continue
		}
		if weakOK && (!strongOK || weak.distance < strong.distance) {
			key = weakKey
		}
		observedWeak := root.Weak || key.weak
		var labels []string
		for paths[key].distance > 0 {
			link := paths[key]
			labels = append(labels, link.label)
			key = link.next
		}
		path := root.Name + strings.Join(labels, "")
		out.Witnesses = append(out.Witnesses, OwnershipWitness{root.Name, path, observedWeak, root.Borrowed, root.UsedAfter, root.UseInstruction})
		switch {
		case observedWeak:
			refuse("weak observer via " + path)
		case root.Borrowed:
			refuse("borrowed owner via " + path)
		case root.Name != owner:
			refuse("second surviving owner via " + path)
		case root.UsedAfter:
			refuse("use after move via " + path)
		}
	}
	if out.Verdict == Proven {
		out.Evidence = append(out.Evidence, "complete private closure; one surviving strong owner; no weak observer or continuation use")
	}
	return out
}

type TransferAnswer struct {
	Graph       *flow.Function
	Site        int
	At          *ir.Statement
	Function    int
	Instruction flow.InstructionId
	Owner       int
	Answer      OwnershipAnswer
}

func OwnershipTransfers(program *ir.Program) []TransferAnswer {
	if !hasMoved(reflect.ValueOf(program.Main)) {
		found := false
		for _, function := range program.Functions {
			if hasMoved(reflect.ValueOf(function.Body)) {
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
	return prepare(program, true).queries
}
func (a *analysis) transferAnswer(move ir.ParallelMap, items value) TransferAnswer {
	result := TransferAnswer{Graph: a.graph, Site: move.Site, At: a.graph.Instructions[a.instruction].At, Function: a.function, Instruction: a.instruction, Owner: -1}
	read, ok := move.Items.(ir.Read)
	if !ok {
		result.Answer = OwnershipAnswer{Verdict: Unknown, Evidence: []string{"candidate is not an owned local"}}
		return result
	}
	result.Owner = read.Local
	s := OwnershipSnapshot{Complete: !a.proof.top}
	all := objects{}
	for o, fields := range a.state.heap {
		all[o] = true
		for _, held := range fields {
			for target := range held.strong {
				all[target] = true
			}
			for target := range held.weak {
				all[target] = true
			}
		}
	}
	for _, held := range a.state.locals {
		for o := range held.strong {
			all[o] = true
		}
		for o := range held.weak {
			all[o] = true
		}
	}
	for o := range items.strong {
		all[o] = true
	}
	allocationInstructions := map[int]flow.InstructionId{}
	for key, site := range a.sites {
		allocationInstructions[site] = key.instruction
	}
	for o := range all {
		var allocation *ir.Statement
		var instruction flow.InstructionId
		if o > 0 {
			instruction = allocationInstructions[siteOf(o)]
			allocation = a.graph.Instructions[instruction].At
		}
		s.Objects = append(s.Objects, OwnershipObject{ID: int(o), AllocationInstruction: instruction, Allocation: allocation, Unknown: o <= 0 || o%2 == 0 || a.state.exposed(o)})
		for field, held := range a.state.heap[o] {
			label := "." + field
			if field == elementKey {
				label = "[]"
			}
			for to := range held.strong {
				s.Edges = append(s.Edges, OwnershipEdge{int(o), int(to), label, false})
			}
			for to := range held.weak {
				s.Edges = append(s.Edges, OwnershipEdge{int(o), int(to), label, true})
			}
		}
	}
	// The existing interpreter has no dynamic region inventory. Never pretend
	// graph type identities are runtime region identities.
	candidateReach := a.state.reach(items)
	for o := range candidateReach {
		if o > 0 && (!a.classifiedSites[siteOf(o)] || a.graphSites[siteOf(o)]) {
			s.Complete = false
		}
	}
	for local, held := range a.state.locals {
		if local >= len(a.proof.program.Locals) {
			if !held.empty() {
				s.Complete = false
			}
			continue
		}
		declared := a.proof.program.Locals[local]
		name := fmt.Sprintf("%s#%d", declared.Name, local)
		after, use := a.continuationUse(local)
		for o := range held.strong {
			s.Roots = append(s.Roots, OwnershipRoot{name, int(o), false, declared.Borrowed, after, use})
		}
		for o := range held.weak {
			s.Roots = append(s.Roots, OwnershipRoot{name, int(o), true, declared.Borrowed, after, use})
		}
	}
	sort.Slice(s.Objects, func(i, j int) bool { return s.Objects[i].ID < s.Objects[j].ID })
	sort.Slice(s.Edges, func(i, j int) bool {
		x, y := s.Edges[i], s.Edges[j]
		if x.From != y.From {
			return x.From < y.From
		}
		if x.Label != y.Label {
			return x.Label < y.Label
		}
		if x.To != y.To {
			return x.To < y.To
		}
		return !x.Weak && y.Weak
	})
	result.Answer = QueryOwnership(s, fmt.Sprintf("%s#%d", a.proof.program.Locals[read.Local].Name, read.Local))
	return result
}

func (a *analysis) continuationUse(local int) (bool, flow.InstructionId) {
	type position struct {
		block flow.BlockId
		index int
	}
	var pending []position
	for _, b := range a.graph.Blocks {
		for i, id := range b.Instructions {
			if id == a.instruction {
				pending = append(pending, position{b.Id, i + 1})
			}
		}
	}
	seen := map[position]bool{}
	declaration := flow.DeclarationId(local + 1)
	for len(pending) > 0 {
		at := pending[0]
		pending = pending[1:]
		if seen[at] {
			continue
		}
		seen[at] = true
		b, _ := a.graph.Block(at.block)
		killed := false
		for i := at.index; i < len(b.Instructions); i++ {
			instruction := a.graph.Instructions[b.Instructions[i]]
			for _, use := range instruction.Uses {
				if a.graph.Identifiers[use.Identifier].Declaration == declaration {
					return true, instruction.Id
				}
			}
			for _, define := range instruction.Defines {
				if a.graph.Identifiers[define.Identifier].Declaration == declaration {
					killed = true
				}
			}
			if killed {
				if terminal, ok := b.Terminal.(*flow.MayThrow); ok && i == len(b.Instructions)-1 {
					pending = append(pending, position{terminal.Handler, 0})
				}
				break
			}
		}
		if !killed {
			flow.EachSuccessor(b.Terminal, func(next flow.BlockId) { pending = append(pending, position{next, 0}) })
		}
	}
	return false, 0
}

// Demand preparation only for actual move boundaries; stage-1 ports without
// parallel transfers do not need a second whole-program interpretation.
func hasMoved(node reflect.Value) bool {
	switch node.Kind() {
	case reflect.Interface, reflect.Pointer:
		if !node.IsNil() {
			return hasMoved(node.Elem())
		}
	case reflect.Struct:
		if move, ok := node.Interface().(ir.ParallelMap); ok && move.Moved {
			return true
		}
		for i := 0; i < node.NumField(); i++ {
			if hasMoved(node.Field(i)) {
				return true
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < node.Len(); i++ {
			if hasMoved(node.Index(i)) {
				return true
			}
		}
	}
	return false
}

func (a *analysis) tagOwnershipGraph(made value, ids []int) {
	if !a.proof.ownership {
		return
	}
	for o := range made.strong {
		if o > 0 {
			a.classifiedSites[siteOf(o)] = true
		}
	}
	for _, id := range ids {
		if a.proof.program.GraphTypes[id] {
			for o := range made.strong {
				if o > 0 {
					a.graphSites[siteOf(o)] = true
				}
			}
		}
	}
}
