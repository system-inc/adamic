package mutation_aliasing

// cohere's side of Adamic's mutation_aliasing slice (stage1/cohere/mutation_aliasing in Adamic).
// Adamic's test overlays this file into cohere's mutation_aliasing module, beside the module's own
// tests and in their package, and runs it with ADAMIC_PORT_REQUEST naming a request. It writes the cases
// the port reads: the vocabulary, the probes of the module's helper tests, the module's own tests'
// functions built with those tests' own helpers, functions generated from a seed, and the functions a
// request's graphs hold (React Compiler's own fixtures as cohere's high-level IR lowers them, with their
// effects, exported by its TestExportRangesForAdamic). Then it reads the cases back, as the port does,
// runs the module's passes on each, and writes what they made in the port's format (main.ts says what
// it is). A request's graphs are gzipped cases files or directories of *.txt cases files, read in the
// order given.
//
// An exported function carries the ranges cohere's own pass gave it. Before anything is written, each
// is checked against the ranges the module gives it here, so the cases are shown to reproduce what
// cohere computes on its own IR, closures and all.
//
// The passes run here are cohere's, unchanged: this file only builds functions, adapts them, and prints.

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	ssa "github.com/system-inc/cohere/static_single_assignment"
)

// adamicPortRequest is what Adamic's test asks for.
type adamicPortRequest struct {
	Seed      int64    `json:"seed"`
	Generated int      `json:"generated"`
	Graphs    []string `json:"graphs"`
	Cases     string   `json:"cases"`
	Answers   string   `json:"answers"`
}

func TestAdamicPortCases(t *testing.T) {
	requestPath := os.Getenv("ADAMIC_PORT_REQUEST")
	if requestPath == "" {
		t.Skip("run by Adamic's mutation_aliasing slice")
	}
	encoded, err := os.ReadFile(requestPath)
	if err != nil {
		t.Fatal(err)
	}
	var request adamicPortRequest
	if err := json.Unmarshal(encoded, &request); err != nil {
		t.Fatal(err)
	}

	var cases strings.Builder
	cases.WriteString(moduleTestProbes())
	for _, test := range moduleTestFunctions() {
		writeOracleFunction(&cases, test)
	}
	random := rand.New(rand.NewSource(request.Seed))
	var earlier []string
	for index := 0; index < request.Generated; index++ {
		f := generatedFunction(fmt.Sprintf("generated-%d", index), random, earlier)
		earlier = append(earlier, f.name)
		writeOracleFunction(&cases, f)
	}
	for _, graphs := range request.Graphs {
		contents, err := readGraphs(graphs)
		if err != nil {
			t.Fatal(err)
		}
		cases.WriteString(contents)
	}
	if err := os.WriteFile(request.Cases, []byte(cases.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := parseOracleCases(cases.String())
	if err != nil {
		t.Fatal(err)
	}
	var answers strings.Builder
	var unfaithful []string
	for _, item := range items {
		if item.function == nil {
			describeStandalone(&answers, item.fields)
			continue
		}
		unfaithful = append(unfaithful, describeOracleFunction(&answers, item.function)...)
	}
	if len(unfaithful) > 0 {
		shown := unfaithful
		if len(shown) > 10 {
			shown = shown[:10]
		}
		t.Fatalf("%d exported functions get other ranges here than cohere's own pass gave them, so the cases don't reproduce cohere:\n%s",
			len(unfaithful), strings.Join(shown, "\n"))
	}
	if err := os.WriteFile(request.Answers, []byte(answers.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}

// readGraphs is the cases in a gzipped file of them, or in every *.txt of a directory, in name order.
func readGraphs(path string) (string, error) {
	information, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !information.IsDir() {
		file, err := os.Open(path)
		if err != nil {
			return "", err
		}
		defer file.Close()
		reader, err := gzip.NewReader(file)
		if err != nil {
			return "", err
		}
		contents, err := io.ReadAll(reader)
		return string(contents), err
	}
	names, err := filepath.Glob(filepath.Join(path, "*.txt"))
	if err != nil {
		return "", err
	}
	sort.Strings(names)
	var contents strings.Builder
	for _, name := range names {
		file, err := os.ReadFile(name)
		if err != nil {
			return "", err
		}
		contents.Write(file)
	}
	return contents.String(), nil
}

// The oracle IR is the module tests' IR with what cohere's high-level IR answers beside it: edges of
// every kind, an absent instruction, and a closure rule. Adamic's main.ts holds the same one.

type oraclePlace struct{ id ssa.IdentifierId }

type oracleVisit struct {
	place oraclePlace
	role  ssa.Role
}

type oracleClosure struct {
	into     ssa.IdentifierId
	rule     string
	answer   bool
	nested   string
	captures []oraclePlace
}

type oracleInstruction struct {
	visits        []oracleVisit
	effects       []AliasingEffect[oraclePlace]
	storedContext *ssa.IdentifierId
	closures      []*oracleClosure
	absent        bool
	order         ssa.EvaluationOrder
}

type oracleEdge struct {
	to   ssa.BlockId
	kind ssa.Edge
}

type oracleBlock struct {
	id            ssa.BlockId
	instructions  []*oracleInstruction
	edges         []oracleEdge
	returnValue   *oraclePlace
	predecessors  []ssa.BlockId
	phis          []*ssa.Phi[oraclePlace]
	terminalOrder ssa.EvaluationOrder
}

type oracleFunction struct {
	name                     string
	passes                   string
	entry                    ssa.BlockId
	bound                    int
	blocks                   []*oracleBlock
	table                    map[ssa.BlockId]*oracleBlock
	params, context          []oraclePlace
	returns                  *oraclePlace
	frozenParameters         bool
	parametersDefinedOnEntry bool
	contextKinds             map[ssa.IdentifierId]EffectValueKind
	contextKindOrder         []ssa.IdentifierId
	expect                   map[ssa.IdentifierId]MutableRange
	expectLength             int
	hasExpect                bool
}

func newOracleFunction(name string) *oracleFunction {
	return &oracleFunction{name: name, table: map[ssa.BlockId]*oracleBlock{}, contextKinds: map[ssa.IdentifierId]EffectValueKind{},
		expect: map[ssa.IdentifierId]MutableRange{}}
}

// oracleFunctions are every function the cases hold, by name, for a closure's nested rule.
var oracleFunctions = map[string]*oracleFunction{}

type oracleGraph struct{}

func (oracleGraph) Entry(f *oracleFunction) ssa.BlockId { return f.entry }
func (oracleGraph) BlockBound(f *oracleFunction) int    { return f.bound }
func (oracleGraph) Block(f *oracleFunction, id ssa.BlockId) (*oracleBlock, bool) {
	block, ok := f.table[id]
	return block, ok
}
func (oracleGraph) Blocks(f *oracleFunction) []*oracleBlock            { return f.blocks }
func (oracleGraph) SetBlocks(f *oracleFunction, blocks []*oracleBlock) { f.blocks = blocks }
func (oracleGraph) Retain(f *oracleFunction, keep func(ssa.BlockId) bool) {
	for id := range f.table {
		if !keep(id) {
			delete(f.table, id)
		}
	}
}
func (oracleGraph) Placeholder(f *oracleFunction, block *oracleBlock) *oracleBlock {
	placeholder := &oracleBlock{id: block.id, predecessors: append([]ssa.BlockId(nil), block.predecessors...)}
	f.table[block.id] = placeholder
	return placeholder
}
func (oracleGraph) Id(block *oracleBlock) ssa.BlockId             { return block.id }
func (oracleGraph) Predecessors(block *oracleBlock) []ssa.BlockId { return block.predecessors }
func (oracleGraph) SetPredecessors(block *oracleBlock, predecessors []ssa.BlockId) {
	block.predecessors = predecessors
}
func (oracleGraph) Phis(block *oracleBlock) []*ssa.Phi[oraclePlace]          { return block.phis }
func (oracleGraph) SetPhis(block *oracleBlock, phis []*ssa.Phi[oraclePlace]) { block.phis = phis }
func (oracleGraph) EachEdge(block *oracleBlock, visit func(ssa.BlockId, ssa.Edge)) {
	for _, edge := range block.edges {
		visit(edge.to, edge.kind)
	}
}
func (oracleGraph) EndsInReturn(block *oracleBlock) bool { return block.returnValue != nil }
func (oracleGraph) InstructionCount(f *oracleFunction, block *oracleBlock) int {
	return len(block.instructions)
}
func (oracleGraph) EachInstructionPlace(f *oracleFunction, block *oracleBlock, index int, visit func(*oraclePlace, ssa.Role)) {
	instruction := block.instructions[index]
	for i := range instruction.visits {
		visit(&instruction.visits[i].place, instruction.visits[i].role)
	}
}
func (oracleGraph) IsContextStore(f *oracleFunction, block *oracleBlock, index int) bool {
	return block.instructions[index].storedContext != nil
}
func (oracleGraph) ContextStoreDefines(f *oracleFunction, block *oracleBlock, index int, place oraclePlace) bool {
	return false
}
func (oracleGraph) SetInstructionOrder(f *oracleFunction, block *oracleBlock, index int, order ssa.EvaluationOrder) {
	block.instructions[index].order = order
}
func (oracleGraph) EachTerminalPlace(block *oracleBlock, visit func(*oraclePlace, ssa.Role)) {}
func (oracleGraph) SetTerminalOrder(block *oracleBlock, order ssa.EvaluationOrder) {
	block.terminalOrder = order
}
func (oracleGraph) Params(f *oracleFunction) []oraclePlace { return f.params }
func (oracleGraph) Returns(f *oracleFunction) *oraclePlace { return f.returns }
func (oracleGraph) Declaration(f *oracleFunction, id ssa.IdentifierId) ssa.DeclarationId {
	return ssa.DeclarationId(id)
}
func (oracleGraph) Contextual(f *oracleFunction, declaration ssa.DeclarationId) bool { return false }
func (oracleGraph) Mint(f *oracleFunction, original ssa.IdentifierId) ssa.IdentifierId {
	panic(fmt.Sprintf("%s: the ranges never mint a value (asked for %d)", f.name, original))
}
func (oracleGraph) Named(f *oracleFunction, id ssa.IdentifierId) bool { return true }
func (oracleGraph) PlaceString(f *oracleFunction, id ssa.IdentifierId) string {
	return strconv.Itoa(int(id))
}
func (oracleGraph) IdentifierOf(place oraclePlace) ssa.IdentifierId { return place.id }
func (oracleGraph) WithIdentifier(place oraclePlace, id ssa.IdentifierId) oraclePlace {
	return oraclePlace{id: id}
}

func (oracleGraph) InstructionOrder(f *oracleFunction, block *oracleBlock, index int) (ssa.EvaluationOrder, bool) {
	instruction := block.instructions[index]
	return instruction.order, !instruction.absent
}
func (oracleGraph) TerminalOrder(block *oracleBlock) ssa.EvaluationOrder { return block.terminalOrder }
func (oracleGraph) Effects(f *oracleFunction, block *oracleBlock, index int) []AliasingEffect[oraclePlace] {
	return block.instructions[index].effects
}
func (oracleGraph) ParametersFrozen(f *oracleFunction) bool { return f.frozenParameters }
func (oracleGraph) Context(f *oracleFunction) []oraclePlace { return f.context }
func (oracleGraph) ReturnValue(block *oracleBlock) (oraclePlace, bool) {
	if block.returnValue == nil {
		return oraclePlace{}, false
	}
	return *block.returnValue, true
}
func (oracleGraph) StoredContextValue(f *oracleFunction, block *oracleBlock, index int) (ssa.IdentifierId, bool) {
	if stored := block.instructions[index].storedContext; stored != nil {
		return *stored, true
	}
	return 0, false
}

// Closure answers from the instruction's first closure record for into, under its rule, as main.ts's
// closureOf does. The captures are never nil, so the pass reads the frozen answer even for a closure
// that captures nothing.
func (oracleGraph) Closure(f *oracleFunction, block *oracleBlock, index int, into ssa.IdentifierId,
	kinds map[ssa.IdentifierId]EffectValueKind) ([]oraclePlace, bool) {
	for _, closure := range block.instructions[index].closures {
		if closure.into != into {
			continue
		}
		captures := closure.captures
		if captures == nil {
			captures = []oraclePlace{}
		}
		if closure.rule == "answer" {
			return captures, closure.answer
		}
		immutable := len(captures) > 0
		for _, capture := range captures {
			_, known := kinds[capture.id]
			immutable = immutable && known
		}
		if closure.rule == "immutable" || !immutable {
			return captures, immutable
		}
		nested, ok := oracleFunctions[closure.nested]
		if !ok {
			panic("no function " + closure.nested)
		}
		if len(nested.context) != len(captures) {
			return captures, false
		}
		contextKinds := map[ssa.IdentifierId]EffectValueKind{}
		for position, context := range nested.context {
			contextKinds[context.id] = kinds[captures[position].id]
		}
		graph := BuildAliasingGraph(oracleGraph{}, nested, Options{ContextKinds: contextKinds})
		return captures, graph.MutationCount() == 0
	}
	return nil, false
}

// oracleItem is one thing the cases ask for, in their order: a function, or a record standing alone.
type oracleItem struct {
	function *oracleFunction
	fields   []string
}

var (
	oracleEdgeKinds  = map[ssa.Edge]string{ssa.Real: "Real", ssa.Fallthrough: "Fallthrough", ssa.Exceptional: "Exceptional"}
	oracleRangeGaps  = map[RangeGap]string{RangeGapLoopCarriedInversion: "loop-carried-inversion"}
	oracleEdgeNames  = map[aliasingEdgeKind]string{aliasingEdgeCapture: "capture", aliasingEdgeAlias: "alias", aliasingEdgeMaybeAlias: "maybe-alias"}
	oracleNodeValues = map[aliasingNodeValue]string{aliasingNodeObject: "object", aliasingNodePhi: "phi"}
)

func oracleEffectKind(name string) (AliasingEffectKind, bool) {
	for kind := AliasingEffectCreate; kind <= AliasingEffectApply; kind++ {
		if kind.String() == name {
			return kind, true
		}
	}
	return 0, false
}

func oracleValueKind(name string) (EffectValueKind, bool) {
	for kind := EffectValueMutable; kind <= EffectValueGlobal; kind++ {
		if kind.String() == name {
			return kind, true
		}
	}
	return 0, false
}

// parseOracleCases reads cases records back into what they ask for, as main.ts reads them, and every
// function into oracleFunctions.
func parseOracleCases(cases string) (items []oracleItem, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("%v", recovered)
		}
	}()
	var current *oracleFunction
	var block *oracleBlock
	var instruction *oracleInstruction
	var closure *oracleClosure
	for _, line := range strings.Split(cases, "\n") {
		if line == "" {
			continue
		}
		fields := strings.Split(line, " ")
		number := func(index int) int {
			if index >= len(fields) {
				panic("a record without field " + strconv.Itoa(index) + ": " + line)
			}
			value, err := strconv.Atoi(fields[index])
			if err != nil {
				panic(line)
			}
			return value
		}
		text := func(index int) string {
			if index >= len(fields) {
				panic("a record without field " + strconv.Itoa(index) + ": " + line)
			}
			return fields[index]
		}
		place := func(index int) oraclePlace { return oraclePlace{id: ssa.IdentifierId(number(index))} }
		switch fields[0] {
		case "function":
			if _, exists := oracleFunctions[text(1)]; exists {
				return nil, fmt.Errorf("a second function named %s", text(1))
			}
			current = newOracleFunction(text(1))
			oracleFunctions[current.name] = current
			items = append(items, oracleItem{function: current})
			block, instruction, closure = nil, nil, nil
		case "vocabulary", "probe":
			items = append(items, oracleItem{fields: fields})
		case "entry":
			current.entry = ssa.BlockId(number(1))
		case "bound":
			current.bound = number(1)
		case "frozen-parameters":
			current.frozenParameters = true
		case "parameters-defined-on-entry":
			current.parametersDefinedOnEntry = true
		case "context-kind":
			kind, ok := oracleValueKind(text(2))
			if !ok {
				return nil, fmt.Errorf("an unknown value kind: %s", line)
			}
			id := ssa.IdentifierId(number(1))
			if _, exists := current.contextKinds[id]; !exists {
				current.contextKindOrder = append(current.contextKindOrder, id)
			}
			current.contextKinds[id] = kind
		case "param":
			current.params = append(current.params, place(1))
		case "context":
			current.context = append(current.context, place(1))
		case "returns":
			returns := place(1)
			current.returns = &returns
		case "block":
			block = &oracleBlock{id: ssa.BlockId(number(1)), terminalOrder: ssa.EvaluationOrder(number(2))}
			current.blocks = append(current.blocks, block)
			current.table[block.id] = block
			instruction, closure = nil, nil
		case "edge":
			kind := -1
			for edge, name := range oracleEdgeKinds {
				if name == text(2) {
					kind = int(edge)
				}
			}
			if kind < 0 {
				return nil, fmt.Errorf("an unknown edge kind: %s", line)
			}
			block.edges = append(block.edges, oracleEdge{to: ssa.BlockId(number(1)), kind: ssa.Edge(kind)})
		case "return":
			returned := place(1)
			block.returnValue = &returned
		case "phi":
			phi := &ssa.Phi[oraclePlace]{Place: place(1)}
			for _, operand := range fields[2:] {
				predecessor, value, found := strings.Cut(operand, "=")
				if !found {
					return nil, fmt.Errorf("a phi operand without =: %s", line)
				}
				predecessorId, err := strconv.Atoi(predecessor)
				if err != nil {
					return nil, err
				}
				valueId, err := strconv.Atoi(value)
				if err != nil {
					return nil, err
				}
				phi.Operands.Set(ssa.BlockId(predecessorId), oraclePlace{id: ssa.IdentifierId(valueId)})
			}
			block.phis = append(block.phis, phi)
		case "instruction":
			instruction = &oracleInstruction{}
			if text(1) == "-" {
				instruction.absent = true
			} else {
				instruction.order = ssa.EvaluationOrder(number(1))
			}
			block.instructions = append(block.instructions, instruction)
			closure = nil
		case "use":
			instruction.visits = append(instruction.visits, oracleVisit{place: place(1), role: ssa.Use})
		case "define":
			instruction.visits = append(instruction.visits, oracleVisit{place: place(1), role: ssa.Define})
		case "effect":
			kind, ok := oracleEffectKind(text(1))
			if !ok {
				return nil, fmt.Errorf("an unknown effect kind: %s", line)
			}
			value, ok := oracleValueKind(text(4))
			if !ok {
				return nil, fmt.Errorf("an unknown value kind: %s", line)
			}
			effect := AliasingEffect[oraclePlace]{Kind: kind, Into: place(2), Value: value}
			if text(3) != "-" {
				effect.From = place(3)
				effect.HasFrom = true
			}
			instruction.effects = append(instruction.effects, effect)
		case "stored-context":
			stored := ssa.IdentifierId(number(1))
			instruction.storedContext = &stored
		case "closure":
			closure = &oracleClosure{into: ssa.IdentifierId(number(1)), rule: text(2)}
			switch closure.rule {
			case "answer":
				closure.answer = number(3) == 1
			case "nested":
				closure.nested = text(3)
			case "immutable":
			default:
				return nil, fmt.Errorf("an unknown closure rule: %s", line)
			}
			instruction.closures = append(instruction.closures, closure)
		case "capture":
			closure.captures = append(closure.captures, place(1))
		case "expect":
			current.expect[ssa.IdentifierId(number(1))] = MutableRange{Start: ssa.EvaluationOrder(number(2)), End: ssa.EvaluationOrder(number(3))}
		case "expect-length":
			current.expectLength = number(1)
			current.hasExpect = true
		case "passes":
			current.passes = text(1)
			current = nil
		default:
			return nil, fmt.Errorf("an unknown record: %s", line)
		}
	}
	return items, nil
}

func bit(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

func oracleRangeText(r MutableRange) string {
	if !r.IsSet() {
		return "-"
	}
	return fmt.Sprintf("%d,%d", r.Start, r.End)
}

func oracleListText(items []string) string {
	if len(items) == 0 {
		return "-"
	}
	return strings.Join(items, ",")
}

func oracleEntriesText(entries map[ssa.IdentifierId]int, order []ssa.IdentifierId) string {
	var items []string
	for _, id := range order {
		items = append(items, fmt.Sprintf("%d@%d", id, entries[id]))
	}
	return oracleListText(items)
}

// oracleValues are the identifiers a function names anywhere, ascending.
func oracleValues(f *oracleFunction) []ssa.IdentifierId {
	seen := map[ssa.IdentifierId]bool{}
	for _, place := range f.params {
		seen[place.id] = true
	}
	for _, place := range f.context {
		seen[place.id] = true
	}
	if f.returns != nil {
		seen[f.returns.id] = true
	}
	for id := range f.contextKinds {
		seen[id] = true
	}
	for _, block := range f.blocks {
		if block.returnValue != nil {
			seen[block.returnValue.id] = true
		}
		for _, phi := range block.phis {
			seen[phi.Place.id] = true
			for _, operand := range phi.Operands {
				seen[operand.Place.id] = true
			}
		}
		for _, instruction := range block.instructions {
			for _, visit := range instruction.visits {
				seen[visit.place.id] = true
			}
			for _, effect := range instruction.effects {
				seen[effect.Into.id] = true
				if effect.HasFrom {
					seen[effect.From.id] = true
				}
			}
			if instruction.storedContext != nil {
				seen[*instruction.storedContext] = true
			}
			for _, closure := range instruction.closures {
				for _, capture := range closure.captures {
					seen[capture.id] = true
				}
			}
		}
	}
	values := make([]ssa.IdentifierId, 0, len(seen))
	for id := range seen {
		values = append(values, id)
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	return values
}

// describeOracleFunction runs the passes a function names, as main.ts's describe does, prints what they
// made, and returns how its ranges differ from the ones cohere's own pass gave it, for an exported one.
func describeOracleFunction(out *strings.Builder, f *oracleFunction) []string {
	graph := oracleGraph{}
	switch f.passes {
	case "finalize-ranges":
		ssa.ReversePostorder(graph, f)
		ssa.MarkPredecessors(graph, f)
		ssa.MarkEvaluationOrder(graph, f)
	case "ranges":
	default:
		panic(f.name + ": unknown passes " + f.passes)
	}
	options := Options{ParametersDefinedOnEntry: f.parametersDefinedOnEntry, ContextKinds: f.contextKinds}

	line := func(format string, arguments ...any) {
		fmt.Fprintf(out, format, arguments...)
		out.WriteByte('\n')
	}
	line("function %s", f.name)
	var order []string
	for _, block := range f.blocks {
		order = append(order, strconv.Itoa(int(block.id)))
	}
	line("order %s", strings.Join(order, " "))
	for _, block := range f.blocks {
		var orders []string
		for _, instruction := range block.instructions {
			if instruction.absent {
				orders = append(orders, "-")
				continue
			}
			orders = append(orders, strconv.Itoa(int(instruction.order)))
		}
		line("block %d first=%d terminal=%d instructions=%s", block.id, BlockFirstOrder(graph, f, block), block.terminalOrder, oracleListText(orders))
	}

	aliasing := BuildAliasingGraph(graph, f, options)
	nodeIds := make([]ssa.IdentifierId, 0, len(aliasing.state.nodes))
	for id := range aliasing.state.nodes {
		nodeIds = append(nodeIds, id)
	}
	sort.Slice(nodeIds, func(i, j int) bool { return nodeIds[i] < nodeIds[j] })
	for _, id := range nodeIds {
		node := aliasing.state.nodes[id]
		var edges []string
		for _, edge := range node.edges {
			edges = append(edges, fmt.Sprintf("%d@%d:%s", edge.node, edge.index, oracleEdgeNames[edge.kind]))
		}
		line("node %d %s created-from=%s captures=%s aliases=%s maybe-aliases=%s edges=%s", node.id, oracleNodeValues[node.value],
			oracleEntriesText(node.createdFrom, node.createdFromOrder), oracleEntriesText(node.captures, node.capturesOrder),
			oracleEntriesText(node.aliases, node.aliasesOrder), oracleEntriesText(node.maybeAliases, node.maybeAliasOrder), oracleListText(edges))
	}
	for _, mutation := range aliasing.mutations {
		line("mutation %d %d %d %s %d", mutation.index, mutation.place, mutation.end, bit(mutation.transitive), mutation.kind)
	}
	line("mutations %d", aliasing.MutationCount())

	widened := &MutableRanges{}
	aliasing.Widen(widened)
	ranges := InferMutableRanges(graph, f, options)
	again := InferMutableRanges(graph, f, options)
	idempotent := ranges.Len() == again.Len()
	var unfaithful []string
	for _, id := range oracleValues(f) {
		r := ranges.Get(id)
		idempotent = idempotent && r == again.Get(id)
		contains := "-"
		if r.IsSet() {
			contains = bit(r.Contains(r.Start-1)) + bit(r.Contains(r.Start)) + bit(r.Contains(r.End-1)) + bit(r.Contains(r.End))
		}
		line("value %d range=%s widened=%s kind=%s contains=%s", id, oracleRangeText(r), oracleRangeText(widened.Get(id)), aliasing.Kind(id), contains)
		if f.hasExpect && r != f.expect[id] {
			unfaithful = append(unfaithful, fmt.Sprintf("%s: value %d has %s here, cohere's pass gave it %s", f.name, id, oracleRangeText(r), oracleRangeText(f.expect[id])))
		}
	}
	line("length %d", ranges.Len())
	var invalid []string
	for _, id := range ValidateMutableRanges(ranges) {
		invalid = append(invalid, strconv.Itoa(int(id)))
	}
	line("invalid %s", oracleListText(invalid))
	line("idempotent %s", bit(idempotent))
	if f.hasExpect && ranges.Len() != f.expectLength {
		unfaithful = append(unfaithful, fmt.Sprintf("%s: %d ranges here, cohere's pass held %d", f.name, ranges.Len(), f.expectLength))
	}
	return unfaithful
}

// describeStandalone prints what a record outside any function asks for, as main.ts's standalone does.
func describeStandalone(out *strings.Builder, fields []string) {
	line := func(format string, arguments ...any) {
		fmt.Fprintf(out, format, arguments...)
		out.WriteByte('\n')
	}
	if fields[0] == "vocabulary" {
		line("vocabulary")
		for kind := AliasingEffectCreate; kind <= AliasingEffectApply; kind++ {
			line("effect-kind %s mutation=%s aliasing=%s", kind, bit(kind.IsMutation()), bit(kind.IsAliasing()))
		}
		for kind := EffectValueMutable; kind <= EffectValueGlobal; kind++ {
			line("value-kind %s", kind)
		}
		var gaps []string
		for _, gap := range RangeGaps() {
			gaps = append(gaps, oracleRangeGaps[gap])
		}
		line("gaps %s", oracleListText(gaps))
		return
	}
	number := func(index int) ssa.EvaluationOrder {
		value, err := strconv.Atoi(fields[index])
		if err != nil {
			panic(strings.Join(fields, " "))
		}
		return ssa.EvaluationOrder(value)
	}
	r, at := MutableRange{Start: number(2), End: number(3)}, number(4)
	switch fields[1] {
	case "range":
		line("range %d %d %d set=%s valid=%s contains=%s", r.Start, r.End, at, bit(r.IsSet()), bit(r.IsValid()), bit(r.Contains(at)))
	case "phi-opens-before":
		line("phi-opens-before %d %d %d %s", r.Start, r.End, at, bit(phiOpensBefore(r, at)))
	case "phi-opened-range":
		opened, moved := phiOpenedRange(r, at)
		text := "-"
		if moved {
			text = fmt.Sprintf("%d,%d", opened.Start, opened.End)
		}
		line("phi-opened-range %d %d %d %s", r.Start, r.End, at, text)
	default:
		panic("an unknown probe: " + strings.Join(fields, " "))
	}
}

// writeOracleFunction writes a function as cases records, the format main.ts reads.
func writeOracleFunction(out *strings.Builder, f *oracleFunction) {
	fmt.Fprintf(out, "function %s\nentry %d\nbound %d\n", f.name, f.entry, f.bound)
	if f.frozenParameters {
		out.WriteString("frozen-parameters\n")
	}
	if f.parametersDefinedOnEntry {
		out.WriteString("parameters-defined-on-entry\n")
	}
	for _, id := range f.contextKindOrder {
		fmt.Fprintf(out, "context-kind %d %s\n", id, f.contextKinds[id])
	}
	for _, param := range f.params {
		fmt.Fprintf(out, "param %d\n", param.id)
	}
	for _, context := range f.context {
		fmt.Fprintf(out, "context %d\n", context.id)
	}
	if f.returns != nil {
		fmt.Fprintf(out, "returns %d\n", f.returns.id)
	}
	for _, block := range f.blocks {
		fmt.Fprintf(out, "block %d %d\n", block.id, block.terminalOrder)
		for _, edge := range block.edges {
			fmt.Fprintf(out, "edge %d %s\n", edge.to, oracleEdgeKinds[edge.kind])
		}
		if block.returnValue != nil {
			fmt.Fprintf(out, "return %d\n", block.returnValue.id)
		}
		for _, phi := range block.phis {
			fmt.Fprintf(out, "phi %d", phi.Place.id)
			for _, operand := range phi.Operands {
				fmt.Fprintf(out, " %d=%d", operand.Predecessor, operand.Place.id)
			}
			out.WriteByte('\n')
		}
		for _, instruction := range block.instructions {
			if instruction.absent {
				out.WriteString("instruction -\n")
			} else {
				fmt.Fprintf(out, "instruction %d\n", instruction.order)
			}
			for _, visit := range instruction.visits {
				role := "use"
				if visit.role == ssa.Define {
					role = "define"
				}
				fmt.Fprintf(out, "%s %d\n", role, visit.place.id)
			}
			for _, effect := range instruction.effects {
				from := "-"
				if effect.HasFrom {
					from = strconv.Itoa(int(effect.From.id))
				}
				fmt.Fprintf(out, "effect %s %d %s %s\n", effect.Kind, effect.Into.id, from, effect.Value)
			}
			if instruction.storedContext != nil {
				fmt.Fprintf(out, "stored-context %d\n", *instruction.storedContext)
			}
			for _, closure := range instruction.closures {
				switch closure.rule {
				case "answer":
					fmt.Fprintf(out, "closure %d answer %s\n", closure.into, bit(closure.answer))
				case "nested":
					fmt.Fprintf(out, "closure %d nested %s\n", closure.into, closure.nested)
				default:
					fmt.Fprintf(out, "closure %d %s\n", closure.into, closure.rule)
				}
				for _, capture := range closure.captures {
					fmt.Fprintf(out, "capture %d\n", capture.id)
				}
			}
		}
	}
	fmt.Fprintf(out, "passes %s\n", f.passes)
}

// oracleFromTest is a module test's function as an oracle function, unchanged: its blocks in their
// order, their successors as real edges, returned values and phis, its instructions' places (uses
// before definitions, as testGraph visits them), effects and stores into captured bindings, and its
// parameters, context, returns place and parameter rule. The module's tests finalize before the pass, so
// it runs as finalize-ranges.
func oracleFromTest(name string, f *testFunction, options Options) *oracleFunction {
	oracle := newOracleFunction(name)
	oracle.passes = "finalize-ranges"
	oracle.entry = f.entry
	oracle.bound = int(f.next)
	oracle.frozenParameters = f.parametersFrozen
	oracle.parametersDefinedOnEntry = options.ParametersDefinedOnEntry
	for _, param := range f.params {
		oracle.params = append(oracle.params, oraclePlace{id: param.id})
	}
	for _, context := range f.context {
		oracle.context = append(oracle.context, oraclePlace{id: context.id})
	}
	if f.returns != nil {
		oracle.returns = &oraclePlace{id: f.returns.id}
	}
	convert := func(effect AliasingEffect[testPlace]) AliasingEffect[oraclePlace] {
		return AliasingEffect[oraclePlace]{Kind: effect.Kind, From: oraclePlace{id: effect.From.id}, Into: oraclePlace{id: effect.Into.id},
			HasFrom: effect.HasFrom, Value: effect.Value}
	}
	for _, block := range f.blocks {
		copied := &oracleBlock{id: block.id}
		for _, successor := range block.successors {
			copied.edges = append(copied.edges, oracleEdge{to: successor, kind: ssa.Real})
		}
		if block.returns != nil {
			copied.returnValue = &oraclePlace{id: block.returns.id}
		}
		for _, phi := range block.phis {
			made := &ssa.Phi[oraclePlace]{Place: oraclePlace{id: phi.Place.id}}
			for _, operand := range phi.Operands {
				made.Operands.Set(operand.Predecessor, oraclePlace{id: operand.Place.id})
			}
			copied.phis = append(copied.phis, made)
		}
		for _, instruction := range block.instructions {
			made := &oracleInstruction{storedContext: instruction.storedContext}
			for _, use := range instruction.uses {
				made.visits = append(made.visits, oracleVisit{place: oraclePlace{id: use.id}, role: ssa.Use})
			}
			for _, define := range instruction.defines {
				made.visits = append(made.visits, oracleVisit{place: oraclePlace{id: define.id}, role: ssa.Define})
			}
			for _, effect := range instruction.effects {
				made.effects = append(made.effects, convert(effect))
			}
			copied.instructions = append(copied.instructions, made)
		}
		oracle.blocks = append(oracle.blocks, copied)
		oracle.table[copied.id] = copied
	}
	return oracle
}

// moduleTestProbes are the module's tests of its helpers, as probe records: the phi helpers'
// cases exactly as TestRangesPhiOpensBeforeItsBlockWhenWidened and TestPhiOpensOneBeforeItsBlock ask
// them, a range's predicates, and the vocabulary TestTheVocabularyNamesEveryKind names.
func moduleTestProbes() string {
	return strings.Join([]string{
		"vocabulary",
		"probe phi-opens-before 0 0 10", "probe phi-opens-before 5 10 10", "probe phi-opens-before 5 11 10", "probe phi-opens-before 1 99 0",
		"probe phi-opened-range 0 20 10", "probe phi-opened-range 5 20 10", "probe phi-opened-range 0 0 10", "probe phi-opened-range 0 2 1",
		"probe range 0 0 0", "probe range 0 0 5", "probe range 3 21 3", "probe range 3 21 10", "probe range 3 21 21", "probe range 3 21 2",
		"probe range 5 5 5", "probe range 6 5 5", "probe range 0 4 0", "probe range 4 0 4", "",
	}, "\n")
}

// moduleTestFunctions builds each function mutation_aliasing_test.go builds, with its helpers, line for
// line as each test does, before its passes. The two tests of the phi kinds and of freezing call the
// state's methods directly; their cases are built here as functions whose effects make the same state,
// one per case of theirs, since a function is what the port's driver reads.
func moduleTestFunctions() []*oracleFunction {
	var tests []*oracleFunction

	{ // TestParametersDefinedOnEntryBothWays
		build := func() *testFunction {
			f := newTestFunction()
			p, q, r := ssa.IdentifierId(1), ssa.IdentifierId(2), ssa.IdentifierId(3)
			f.params = []testPlace{place(p), place(q), place(r)}
			entry := f.block()
			entry.add(&testInstruction{uses: []testPlace{place(q)},
				effects: []AliasingEffect[testPlace]{MutationEffect(AliasingEffectMutateTransitiveConditionally, place(p))}})
			entry.add(&testInstruction{uses: []testPlace{place(p)}})
			entry.add(&testInstruction{uses: []testPlace{place(p)},
				effects: []AliasingEffect[testPlace]{MutationEffect(AliasingEffectMutate, place(p))}})
			return f
		}
		tests = append(tests, oracleFromTest("parameters-defined-on-entry-without", build(), Options{}))
		tests = append(tests, oracleFromTest("parameters-defined-on-entry-with", build(), Options{ParametersDefinedOnEntry: true}))
	}
	{ // TestAMutationReachesWhatItIsAnAliasOf
		f := newTestFunction()
		source, alias := ssa.IdentifierId(1), ssa.IdentifierId(2)
		entry := f.block()
		entry.add(&testInstruction{defines: []testPlace{place(source)},
			effects: []AliasingEffect[testPlace]{CreateEffect(place(source), EffectValueMutable)}})
		entry.add(&testInstruction{uses: []testPlace{place(source)}, defines: []testPlace{place(alias)},
			effects: []AliasingEffect[testPlace]{
				CreateEffect(place(alias), EffectValueMutable),
				FlowEffect(AliasingEffectAlias, place(source), place(alias)),
			}})
		entry.add(&testInstruction{uses: []testPlace{place(alias)},
			effects: []AliasingEffect[testPlace]{MutationEffect(AliasingEffectMutate, place(alias))}})
		tests = append(tests, oracleFromTest("a-mutation-reaches-what-it-is-an-alias-of", f, Options{}))
	}
	{ // TestAnAliasMadeAfterAMutationDoesNotWiden
		f := newTestFunction()
		source, alias := ssa.IdentifierId(1), ssa.IdentifierId(2)
		entry := f.block()
		entry.add(&testInstruction{defines: []testPlace{place(source)},
			effects: []AliasingEffect[testPlace]{CreateEffect(place(source), EffectValueMutable)}})
		entry.add(&testInstruction{defines: []testPlace{place(alias)},
			effects: []AliasingEffect[testPlace]{CreateEffect(place(alias), EffectValueMutable)}})
		entry.add(&testInstruction{uses: []testPlace{place(alias)},
			effects: []AliasingEffect[testPlace]{MutationEffect(AliasingEffectMutate, place(alias))}})
		entry.add(&testInstruction{uses: []testPlace{place(source), place(alias)},
			effects: []AliasingEffect[testPlace]{FlowEffect(AliasingEffectAlias, place(source), place(alias))}})
		tests = append(tests, oracleFromTest("an-alias-made-after-a-mutation-does-not-widen", f, Options{}))
	}
	{ // TestFrozenParametersDropAConditionalMutation
		for _, frozen := range []bool{true, false} {
			f := newTestFunction()
			f.parametersFrozen = frozen
			p := ssa.IdentifierId(1)
			f.params = []testPlace{place(p)}
			entry := f.block()
			entry.add(&testInstruction{uses: []testPlace{place(p)}})
			entry.add(&testInstruction{uses: []testPlace{place(p)},
				effects: []AliasingEffect[testPlace]{MutationEffect(AliasingEffectMutateConditionally, place(p))}})
			tests = append(tests, oracleFromTest(map[bool]string{true: "frozen-parameters-drop-a-conditional-mutation",
				false: "mutable-parameters-widen-on-a-conditional-mutation"}[frozen], f, Options{}))
		}
	}
	{ // TestAStoreIntoACapturedBindingWidensFromItsShape
		f := newTestFunction()
		value := ssa.IdentifierId(1)
		entry := f.block()
		entry.add(&testInstruction{defines: []testPlace{place(value)}})
		entry.add(&testInstruction{uses: []testPlace{place(value)}})
		stored := value
		entry.add(&testInstruction{uses: []testPlace{place(value)}, storedContext: &stored})
		tests = append(tests, oracleFromTest("a-store-into-a-captured-binding-widens-from-its-shape", f, Options{}))
	}
	{ // TestMutatingAReturnedValueReachesTheReturnsPlace, with its returns place and without
		for _, withReturns := range []bool{true, false} {
			f := newTestFunction()
			value, returned := ssa.IdentifierId(1), ssa.IdentifierId(2)
			returns := place(returned)
			f.returns = &returns
			entry := f.block()
			entry.add(&testInstruction{defines: []testPlace{place(value)},
				effects: []AliasingEffect[testPlace]{CreateEffect(place(value), EffectValueMutable)}})
			entry.returns = &testPlace{id: value}
			name := "mutating-a-returned-value-reaches-the-returns-place"
			if !withReturns {
				f.returns = nil
				name = "a-function-without-a-returns-place-has-no-returns-node"
			}
			tests = append(tests, oracleFromTest(name, f, Options{}))
		}
	}
	{ // TestAMutationOfAPhiReachesItsOperands
		f := newTestFunction()
		left, right, merged := ssa.IdentifierId(1), ssa.IdentifierId(2), ssa.IdentifierId(3)
		entry, leftBlock, rightBlock, join := f.block(), f.block(), f.block(), f.block()
		entry.successors = []ssa.BlockId{leftBlock.id, rightBlock.id}
		leftBlock.successors = []ssa.BlockId{join.id}
		rightBlock.successors = []ssa.BlockId{join.id}
		leftBlock.add(&testInstruction{defines: []testPlace{place(left)},
			effects: []AliasingEffect[testPlace]{CreateEffect(place(left), EffectValueMutable)}})
		rightBlock.add(&testInstruction{defines: []testPlace{place(right)},
			effects: []AliasingEffect[testPlace]{CreateEffect(place(right), EffectValueMutable)}})
		phi := &ssa.Phi[testPlace]{Place: place(merged)}
		phi.Operands.Set(leftBlock.id, place(left))
		phi.Operands.Set(rightBlock.id, place(right))
		join.phis = []*ssa.Phi[testPlace]{phi}
		join.add(&testInstruction{uses: []testPlace{place(merged)},
			effects: []AliasingEffect[testPlace]{MutationEffect(AliasingEffectMutate, place(merged))}})
		tests = append(tests, oracleFromTest("a-mutation-of-a-phi-reaches-its-operands", f, Options{}))
	}
	{ // TestPhiValueKindsPreserveMixedFrozenValues, one function per case: the two operands created with
		// their kinds in the join's two predecessors, and for the unvisited predecessor, the right operand
		// reaching the join by a back edge, so the join is walked before it.
		for _, testCase := range []struct {
			name        string
			left, right EffectValueKind
			unseen      bool
		}{
			{"frozen-mutable", EffectValueFrozen, EffectValueMutable, false},
			{"mutable-frozen", EffectValueMutable, EffectValueFrozen, false},
			{"mixed-primitive", EffectValueMaybeFrozen, EffectValuePrimitive, false},
			{"mixed-global", EffectValueMaybeFrozen, EffectValueGlobal, false},
			{"mixed-frozen", EffectValueMaybeFrozen, EffectValueFrozen, false},
			{"frozen-primitive", EffectValueFrozen, EffectValuePrimitive, false},
			{"frozen-global", EffectValueFrozen, EffectValueGlobal, false},
			{"global-primitive", EffectValueGlobal, EffectValuePrimitive, false},
			{"global-mutable", EffectValueGlobal, EffectValueMutable, false},
			{"primitive-mutable", EffectValuePrimitive, EffectValueMutable, false},
			{"unvisited-predecessor", EffectValueFrozen, EffectValueMutable, true},
		} {
			f := newTestFunction()
			entry, leftBlock, join, rightBlock := f.block(), f.block(), f.block(), f.block()
			if testCase.unseen {
				entry.successors = []ssa.BlockId{leftBlock.id}
				leftBlock.successors = []ssa.BlockId{join.id}
				join.successors = []ssa.BlockId{rightBlock.id}
				rightBlock.successors = []ssa.BlockId{join.id}
			} else {
				entry.successors = []ssa.BlockId{leftBlock.id, rightBlock.id}
				leftBlock.successors = []ssa.BlockId{join.id}
				rightBlock.successors = []ssa.BlockId{join.id}
			}
			leftBlock.add(&testInstruction{defines: []testPlace{place(1)},
				effects: []AliasingEffect[testPlace]{CreateEffect(place(1), testCase.left)}})
			rightBlock.add(&testInstruction{defines: []testPlace{place(2)},
				effects: []AliasingEffect[testPlace]{CreateEffect(place(2), testCase.right)}})
			phi := &ssa.Phi[testPlace]{Place: place(3)}
			phi.Operands.Set(leftBlock.id, place(1))
			phi.Operands.Set(rightBlock.id, place(2))
			join.phis = []*ssa.Phi[testPlace]{phi}
			join.add(&testInstruction{uses: []testPlace{place(3)}})
			tests = append(tests, oracleFromTest("phi-value-kinds-"+testCase.name, f, Options{}))
		}
	}
	{ // cohere's TestCreateFromMutationPropagatesTransitively (its IR's ranges_test.go), rebuilt in the
		// module tests' IR: a mutation of a value derived from a source reaches what the source captured,
		// since created-from forces it transitive. Neither the corpus nor the generator reaches that path
		// reliably, and only this function shows a created-from that keeps the mutation's own kind.
		f := newTestFunction()
		leaf, source, derived, result := ssa.IdentifierId(1), ssa.IdentifierId(2), ssa.IdentifierId(3), ssa.IdentifierId(4)
		returns := place(result)
		f.returns = &returns
		entry := f.block()
		entry.add(&testInstruction{defines: []testPlace{place(leaf)},
			effects: []AliasingEffect[testPlace]{CreateEffect(place(leaf), EffectValueMutable)}})
		entry.add(&testInstruction{defines: []testPlace{place(source)},
			effects: []AliasingEffect[testPlace]{CreateEffect(place(source), EffectValueMutable), FlowEffect(AliasingEffectCapture, place(leaf), place(source))}})
		entry.add(&testInstruction{uses: []testPlace{place(source)}, defines: []testPlace{place(derived)},
			effects: []AliasingEffect[testPlace]{FlowEffect(AliasingEffectCreateFrom, place(source), place(derived))}})
		entry.add(&testInstruction{defines: []testPlace{place(result)},
			effects: []AliasingEffect[testPlace]{CreateEffect(place(result), EffectValuePrimitive), MutationEffect(AliasingEffectMutate, place(derived))}})
		entry.returns = &testPlace{id: result}
		tests = append(tests, oracleFromTest("create-from-mutation-propagates-transitively", f, Options{}))
	}
	{ // TestFreezeRetainsAlreadyImmutableKinds, one function per kind: created with it, then frozen.
		for _, kind := range []EffectValueKind{EffectValuePrimitive, EffectValueGlobal, EffectValueFrozen} {
			f := newTestFunction()
			entry := f.block()
			entry.add(&testInstruction{defines: []testPlace{place(1)},
				effects: []AliasingEffect[testPlace]{CreateEffect(place(1), kind)}})
			entry.add(&testInstruction{uses: []testPlace{place(1)},
				effects: []AliasingEffect[testPlace]{MutationEffect(AliasingEffectFreeze, place(1))}})
			tests = append(tests, oracleFromTest("freeze-retains-"+kind.String(), f, Options{}))
		}
	}
	return tests
}

// generatedFunction makes a function of the shapes the pass meets, from the random source, already
// finalized (reverse postorder, predecessors and evaluation order, by the single assignment module): one
// to eight blocks, loops, diamonds, exceptional edges and blocks nothing reaches; parameters, captured
// values and a returns place; phis at joins, including an operand from a block that never precedes them;
// every effect kind over values defined before, after or never; Creates of every kind, closures under
// each rule, stores into captured bindings, returned values, an instruction the function doesn't hold,
// and a block finalize never numbered; and both options. Some run as finalize-ranges, so the driver
// finalizes them again.
func generatedFunction(name string, random *rand.Rand, earlier []string) *oracleFunction {
	f := newOracleFunction(name)
	f.passes = "ranges"
	finalizeAgain := random.Intn(6) == 0
	if finalizeAgain {
		f.passes = "finalize-ranges"
	}

	next := ssa.IdentifierId(1)
	var values []ssa.IdentifierId
	fresh := func() ssa.IdentifierId {
		id := next
		next++
		values = append(values, id)
		return id
	}
	// someValue is a value the function names: one made already, one made later, or one nothing makes.
	someValue := func() ssa.IdentifierId {
		switch random.Intn(12) {
		case 0:
			return next + ssa.IdentifierId(random.Intn(4))
		case 1:
			return 900 + ssa.IdentifierId(random.Intn(3))
		}
		if len(values) == 0 {
			return fresh()
		}
		return values[random.Intn(len(values))]
	}
	valueKinds := []EffectValueKind{EffectValueMutable, EffectValueMutable, EffectValueMutable, EffectValuePrimitive, EffectValueFrozen,
		EffectValueMaybeFrozen, EffectValueGlobal}
	valueKind := func() EffectValueKind { return valueKinds[random.Intn(len(valueKinds))] }

	for count := random.Intn(3); count > 0; count-- {
		f.params = append(f.params, oraclePlace{id: fresh()})
	}
	for count := random.Intn(3); count > 0; count-- {
		f.context = append(f.context, oraclePlace{id: fresh()})
	}
	if random.Intn(2) == 0 {
		f.returns = &oraclePlace{id: fresh()}
	}
	f.frozenParameters = random.Intn(3) == 0
	f.parametersDefinedOnEntry = random.Intn(4) == 0
	for _, context := range f.context {
		if random.Intn(3) == 0 {
			f.contextKindOrder = append(f.contextKindOrder, context.id)
			f.contextKinds[context.id] = valueKind()
		}
	}
	if random.Intn(10) == 0 {
		id := 950 + ssa.IdentifierId(random.Intn(3))
		f.contextKindOrder = append(f.contextKindOrder, id)
		f.contextKinds[id] = valueKind()
	}

	blockCount := 1 + random.Intn(8)
	var ids []ssa.BlockId
	nextBlock := ssa.BlockId(1)
	for index := 0; index < blockCount; index++ {
		if random.Intn(6) == 0 {
			nextBlock++
		}
		ids = append(ids, nextBlock)
		nextBlock++
	}
	f.bound = int(nextBlock) + random.Intn(2)
	f.entry = ids[0]
	for _, id := range ids {
		block := &oracleBlock{id: id}
		f.blocks = append(f.blocks, block)
		f.table[id] = block
		for count := random.Intn(4); count > 0; count-- {
			if len(ids) > 1 {
				block.edges = append(block.edges, oracleEdge{to: ids[1+random.Intn(len(ids)-1)], kind: ssa.Real})
			}
		}
		if random.Intn(10) == 0 && len(ids) > 1 {
			block.edges = append(block.edges, oracleEdge{to: ids[1+random.Intn(len(ids)-1)], kind: ssa.Exceptional})
		}
		for count := random.Intn(6); count > 0; count-- {
			block.instructions = append(block.instructions, &oracleInstruction{})
		}
	}
	graph := oracleGraph{}
	ssa.ReversePostorder(graph, f)
	ssa.MarkPredecessors(graph, f)
	ssa.MarkEvaluationOrder(graph, f)

	for blockIndex, block := range f.blocks {
		if len(block.predecessors) >= 2 || (len(block.predecessors) == 1 && random.Intn(5) == 0) {
			for count := random.Intn(3); count > 0; count-- {
				phi := &ssa.Phi[oraclePlace]{Place: oraclePlace{id: fresh()}}
				for _, predecessor := range block.predecessors {
					phi.Operands.Set(predecessor, oraclePlace{id: someValue()})
				}
				if random.Intn(10) == 0 {
					phi.Operands.Set(nextBlock+ssa.BlockId(random.Intn(3)), oraclePlace{id: someValue()})
				}
				block.phis = append(block.phis, phi)
			}
		}
		for _, instruction := range block.instructions {
			if random.Intn(25) == 0 {
				instruction.absent = true
			}
			var creates []ssa.IdentifierId
			for count := random.Intn(3); count > 0; count-- {
				id := fresh()
				if random.Intn(10) == 0 && len(values) > 1 {
					id = values[random.Intn(len(values)-1)]
				}
				instruction.visits = append(instruction.visits, oracleVisit{place: oraclePlace{id: id}, role: ssa.Define})
				if random.Intn(5) != 0 {
					instruction.effects = append(instruction.effects, CreateEffect(oraclePlace{id: id}, valueKind()))
					creates = append(creates, id)
				}
			}
			for count := random.Intn(4); count > 0; count-- {
				role := ssa.Use
				if random.Intn(8) == 0 {
					role = ssa.Define
				}
				visit := oracleVisit{place: oraclePlace{id: someValue()}, role: role}
				position := random.Intn(len(instruction.visits) + 1)
				instruction.visits = append(instruction.visits[:position], append([]oracleVisit{visit}, instruction.visits[position:]...)...)
			}
			for count := random.Intn(5); count > 0; count-- {
				kind := AliasingEffectKind(random.Intn(int(AliasingEffectApply) + 1))
				switch {
				case kind == AliasingEffectCreate:
					id := someValue()
					instruction.effects = append(instruction.effects, CreateEffect(oraclePlace{id: id}, valueKind()))
					creates = append(creates, id)
				case kind.IsAliasing() || kind == AliasingEffectImmutableCapture:
					instruction.effects = append(instruction.effects, FlowEffect(kind, oraclePlace{id: someValue()}, oraclePlace{id: someValue()}))
				default:
					instruction.effects = append(instruction.effects, MutationEffect(kind, oraclePlace{id: someValue()}))
				}
			}
			if random.Intn(12) == 0 {
				stored := someValue()
				instruction.storedContext = &stored
			}
			if len(creates) > 0 && random.Intn(4) == 0 {
				closure := &oracleClosure{into: creates[random.Intn(len(creates))]}
				switch choice := random.Intn(4); {
				case choice == 0:
					closure.rule, closure.answer = "answer", random.Intn(2) == 0
				case choice == 1 && len(earlier) > 0:
					closure.rule, closure.nested = "nested", earlier[len(earlier)-1-random.Intn(min(len(earlier), 20))]
				default:
					closure.rule = "immutable"
				}
				for count := random.Intn(4); count > 0; count-- {
					closure.captures = append(closure.captures, oraclePlace{id: someValue()})
				}
				instruction.closures = append(instruction.closures, closure)
			}
		}
		hasSuccessor := len(block.edges) > 0
		if (!hasSuccessor && random.Intn(4) != 0) || random.Intn(12) == 0 {
			returned := someValue()
			block.returnValue = &oraclePlace{id: returned}
		}
		// A block finalize never numbered, which only a function the driver doesn't finalize again can
		// keep: the pass must skip its instructions and widen nothing to it.
		if !finalizeAgain && blockIndex > 0 && random.Intn(15) == 0 {
			for _, instruction := range block.instructions {
				instruction.order = 0
			}
			block.terminalOrder = 0
		}
	}
	return f
}
