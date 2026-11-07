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

func (verdict Verdict) String() string { return [...]string{"Unknown", "Proven", "Refused"}[verdict] }

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
func QueryOwnership(snapshot OwnershipSnapshot, owner string) OwnershipAnswer {
	out := OwnershipAnswer{Verdict: Proven, Snapshot: snapshot}
	unknown := func(why string) {
		if out.Verdict == Proven {
			out.Verdict = Unknown
		}
		out.Evidence = append(out.Evidence, why)
	}
	refuse := func(why string) { out.Verdict = Refused; out.Evidence = append(out.Evidence, why) }
	if len(snapshot.Objects)+len(snapshot.Edges)+len(snapshot.Roots) > 1000000 {
		unknown("ownership witness budget exhausted")
		return out
	}
	if !snapshot.Complete {
		unknown("incomplete root, edge or region inventory")
	}
	objects := map[int]OwnershipObject{}
	regions := map[int][]int{}
	forward := map[int][]OwnershipEdge{}
	for _, object := range snapshot.Objects {
		if _, ok := objects[object.ID]; ok {
			unknown("duplicate object identity")
		}
		objects[object.ID] = object
		if object.Region != 0 {
			regions[object.Region] = append(regions[object.Region], object.ID)
		}
	}
	for _, edge := range snapshot.Edges {
		forward[edge.From] = append(forward[edge.From], edge)
	}
	for id := range forward {
		sort.Slice(forward[id], func(left, right int) bool {
			a, b := forward[id][left], forward[id][right]
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
	for index := range snapshot.Roots {
		if snapshot.Roots[index].Name == owner {
			if selected != nil {
				unknown("ambiguous candidate root")
			}
			selected = &snapshot.Roots[index]
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
		object, ok := objects[id]
		if !ok || object.Unknown {
			unknown(fmt.Sprintf("unresolved object %d", id))
		}
		for _, edge := range forward[id] {
			if !edge.Weak {
				queue = append(queue, edge.To)
			}
		}
		if object.Region != 0 && !expandedRegions[object.Region] {
			expandedRegions[object.Region] = true
			queue = append(queue, regions[object.Region]...)
		}
	}
	for id := range members {
		out.Members = append(out.Members, id)
	}
	sort.Ints(out.Members)

	// Index backwards once rather than walking the heap separately for every root.
	// Two states preserve strong ownership separately from weak observation.
	reverse := map[int][]OwnershipEdge{}
	for _, edge := range snapshot.Edges {
		reverse[edge.To] = append(reverse[edge.To], edge)
	}
	for id := range reverse {
		sort.Slice(reverse[id], func(left, right int) bool {
			a, b := reverse[id][left], reverse[id][right]
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
		for _, edge := range reverse[at.id] {
			key := stateKey{edge.From, at.weak || edge.Weak}
			if _, seen := paths[key]; seen {
				continue
			}
			paths[key] = predecessor{at, edge.Label, paths[at].distance + 1}
			pending = append(pending, key)
		}
	}
	// An opaque frontier reachable from any root independently prevents proof.
	unknownReach := map[int]bool{}
	unknownQueue := []int{}
	seedUnknown := func(id int) {
		object, ok := objects[id]
		if (!ok || object.Unknown) && !unknownReach[id] {
			unknownReach[id] = true
			unknownQueue = append(unknownQueue, id)
		}
	}
	for _, object := range snapshot.Objects {
		seedUnknown(object.ID)
	}
	for _, edge := range snapshot.Edges {
		seedUnknown(edge.From)
		seedUnknown(edge.To)
	}
	for _, root := range snapshot.Roots {
		seedUnknown(root.Object)
	}
	for len(unknownQueue) > 0 {
		id := unknownQueue[0]
		unknownQueue = unknownQueue[1:]
		for _, edge := range reverse[id] {
			if !unknownReach[edge.From] {
				unknownReach[edge.From] = true
				unknownQueue = append(unknownQueue, edge.From)
			}
		}
	}
	roots := append([]OwnershipRoot(nil), snapshot.Roots...)
	sort.Slice(roots, func(left, right int) bool { return roots[left].Name < roots[right].Name })
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
		out.Witnesses = append(
			out.Witnesses,
			OwnershipWitness{root.Name, path, observedWeak, root.Borrowed, root.UsedAfter, root.UseInstruction},
		)
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
		out.Evidence = append(
			out.Evidence,
			"complete private closure; one surviving strong owner; no weak observer or continuation use",
		)
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
	return QueryOwnershipTransfers(program)
}

// QueryOwnershipTransfers evaluates ownership at moved parallelMap boundaries.
// Callers that already recorded move demand during IR construction can use this
// entry directly, avoiding OwnershipTransfers' whole-IR discovery walk.
func QueryOwnershipTransfers(program *ir.Program) []TransferAnswer {
	return prepare(program, true).queries
}
func (analysis *analysis) transferAnswer(move ir.ParallelMap, items value) TransferAnswer {
	result := TransferAnswer{
		Graph:       analysis.graph,
		Site:        move.Site,
		At:          analysis.graph.Instructions[analysis.instruction].At,
		Function:    analysis.function,
		Instruction: analysis.instruction,
		Owner:       -1,
	}
	read, ok := move.Items.(ir.Read)
	if !ok {
		result.Answer = OwnershipAnswer{Verdict: Unknown, Evidence: []string{"candidate is not an owned local"}}
		return result
	}
	result.Owner = read.Local
	snapshot := OwnershipSnapshot{Complete: !analysis.proof.top}
	all := objects{}
	for key, fields := range analysis.state.heap {
		all[key] = true
		for _, held := range fields {
			for target := range held.strong {
				all[target] = true
			}
			for target := range held.weak {
				all[target] = true
			}
		}
	}
	for _, held := range analysis.state.locals {
		for key := range held.strong {
			all[key] = true
		}
		for key := range held.weak {
			all[key] = true
		}
	}
	for key := range items.strong {
		all[key] = true
	}
	allocationInstructions := map[int]flow.InstructionId{}
	for key, site := range analysis.sites {
		allocationInstructions[site] = key.instruction
	}
	for key := range all {
		var allocation *ir.Statement
		var instruction flow.InstructionId
		if key > 0 {
			instruction = allocationInstructions[siteOf(key)]
			allocation = analysis.graph.Instructions[instruction].At
		}
		snapshot.Objects = append(snapshot.Objects, OwnershipObject{
			ID:                    int(key),
			AllocationInstruction: instruction,
			Allocation:            allocation,
			Unknown:               key <= 0 || key%2 == 0 || analysis.state.exposed(key),
		})
		for field, held := range analysis.state.heap[key] {
			label := "." + field
			if field == elementKey {
				label = "[]"
			}
			for to := range held.strong {
				snapshot.Edges = append(snapshot.Edges, OwnershipEdge{int(key), int(to), label, false})
			}
			for to := range held.weak {
				snapshot.Edges = append(snapshot.Edges, OwnershipEdge{int(key), int(to), label, true})
			}
		}
	}
	// The existing interpreter has no dynamic region inventory. Never pretend
	// graph type identities are runtime region identities.
	candidateReach := analysis.state.reach(items)
	for key := range candidateReach {
		if key > 0 && (!analysis.classifiedSites[siteOf(key)] || analysis.graphSites[siteOf(key)]) {
			snapshot.Complete = false
		}
	}
	for local, held := range analysis.state.locals {
		if local >= len(analysis.proof.program.Locals) {
			if !held.empty() {
				snapshot.Complete = false
			}
			continue
		}
		declared := analysis.proof.program.Locals[local]
		name := fmt.Sprintf("%s#%d", declared.Name, local)
		after, use := analysis.continuationUse(local)
		for key := range held.strong {
			snapshot.Roots = append(snapshot.Roots, OwnershipRoot{name, int(key), false, declared.Borrowed, after, use})
		}
		for key := range held.weak {
			snapshot.Roots = append(snapshot.Roots, OwnershipRoot{name, int(key), true, declared.Borrowed, after, use})
		}
	}
	sort.Slice(
		snapshot.Objects,
		func(left, right int) bool { return snapshot.Objects[left].ID < snapshot.Objects[right].ID },
	)
	sort.Slice(snapshot.Edges, func(left, right int) bool {
		x, y := snapshot.Edges[left], snapshot.Edges[right]
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
	result.Answer = QueryOwnership(
		snapshot,
		fmt.Sprintf("%s#%d", analysis.proof.program.Locals[read.Local].Name, read.Local),
	)
	return result
}

func (analysis *analysis) continuationUse(local int) (bool, flow.InstructionId) {
	type position struct {
		block flow.BlockId
		index int
	}
	var pending []position
	for _, block := range analysis.graph.Blocks {
		for index, id := range block.Instructions {
			if id == analysis.instruction {
				pending = append(pending, position{block.Id, index + 1})
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
		block, _ := analysis.graph.Block(at.block)
		killed := false
		for index := at.index; index < len(block.Instructions); index++ {
			instruction := analysis.graph.Instructions[block.Instructions[index]]
			for _, use := range instruction.Uses {
				if analysis.graph.Identifiers[use.Identifier].Declaration == declaration {
					return true, instruction.Id
				}
			}
			for _, define := range instruction.Defines {
				if analysis.graph.Identifiers[define.Identifier].Declaration == declaration {
					killed = true
				}
			}
			if killed {
				if terminal, ok := block.Terminal.(*flow.MayThrow); ok && index == len(block.Instructions)-1 {
					pending = append(pending, position{terminal.Handler, 0})
				}
				break
			}
		}
		if !killed {
			flow.EachSuccessor(block.Terminal, func(next flow.BlockId) { pending = append(pending, position{next, 0}) })
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
		for index := 0; index < node.NumField(); index++ {
			if hasMoved(node.Field(index)) {
				return true
			}
		}
	case reflect.Slice, reflect.Array:
		for index := 0; index < node.Len(); index++ {
			if hasMoved(node.Index(index)) {
				return true
			}
		}
	}
	return false
}

func (analysis *analysis) tagOwnershipGraph(made value, ids []int, program bool) {
	if !analysis.proof.ownership {
		return
	}
	for key := range made.strong {
		if key > 0 {
			analysis.classifiedSites[siteOf(key)] = true
			if program {
				analysis.graphSites[siteOf(key)] = true
			}
		}
	}
	for _, id := range ids {
		if analysis.proof.program.GraphTypes[id] {
			for key := range made.strong {
				if key > 0 {
					analysis.graphSites[siteOf(key)] = true
				}
			}
		}
	}
}
