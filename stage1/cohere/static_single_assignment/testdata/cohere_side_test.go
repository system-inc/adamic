package static_single_assignment_test

// cohere's side of Adamic's static_single_assignment slice (stage1/cohere/static_single_assignment in
// Adamic). Adamic's test overlays this file into cohere's static_single_assignment module, beside the
// module's own tests, and runs it with ADAMIC_PORT_REQUEST naming a request. It writes the cases the
// port reads: the module's own ten tests' functions, built with those tests' own helpers; functions
// generated from a seed; and the functions a request's graphs hold (React Compiler's own fixtures as
// cohere's high-level IR lowers them, exported by its TestExportGraphsForAdamic). Then it reads the cases back, as the port does, runs the module's
// passes on each, and writes what they made in the port's format (main.ts says what it is). A request's
// graphs are gzipped cases files or directories of *.txt cases files, read in the order given.
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
	"slices"
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
		t.Skip("run by Adamic's static_single_assignment slice")
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
	for _, test := range moduleTestFunctions() {
		writeOracleFunction(&cases, oracleFromTest(test.name, test.function, test.passes))
	}
	random := rand.New(rand.NewSource(request.Seed))
	for index := 0; index < request.Generated; index++ {
		writeOracleFunction(&cases, generatedFunction(fmt.Sprintf("generated-%d", index), random))
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

	functions, err := parseOracleCases(cases.String())
	if err != nil {
		t.Fatal(err)
	}
	var answers strings.Builder
	for _, function := range functions {
		runOracleFunction(&answers, function)
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

// The oracle IR is the module tests' IR with a terminal's places added, which cohere's high-level IR
// has and the tests' IR leaves out. Adamic's main.ts holds the same one.

type oraclePlace struct {
	id  ssa.IdentifierId
	tag string
}

type oracleInstruction struct {
	uses, defines []oraclePlace
	contextStore  bool
	result        ssa.IdentifierId
	order         ssa.EvaluationOrder
}

type oracleEdge struct {
	to   ssa.BlockId
	kind ssa.Edge
}

type oracleBlock struct {
	id                            ssa.BlockId
	instructions                  []*oracleInstruction
	edges                         []oracleEdge
	returns                       bool
	predecessors                  []ssa.BlockId
	phis                          []*ssa.Phi[oraclePlace]
	terminalUses, terminalDefines []oraclePlace
	terminalOrder                 ssa.EvaluationOrder
	placeholder                   bool
}

type oracleIdentifier struct {
	declaration ssa.DeclarationId
	name        string
}

type oracleFunction struct {
	name        string
	passes      string
	entry       ssa.BlockId
	bound       int
	blocks      []*oracleBlock
	table       map[ssa.BlockId]*oracleBlock
	identifiers []oracleIdentifier
	params      []oraclePlace
	returns     *oraclePlace
	contextual  map[ssa.DeclarationId]bool
}

func newOracleFunction(name string) *oracleFunction {
	return &oracleFunction{name: name, table: map[ssa.BlockId]*oracleBlock{}, contextual: map[ssa.DeclarationId]bool{}}
}

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
	placeholder := &oracleBlock{id: block.id, placeholder: true,
		predecessors: append([]ssa.BlockId(nil), block.predecessors...)}
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
func (oracleGraph) EndsInReturn(block *oracleBlock) bool { return block.returns }
func (oracleGraph) InstructionCount(f *oracleFunction, block *oracleBlock) int {
	return len(block.instructions)
}
func (oracleGraph) EachInstructionPlace(f *oracleFunction, block *oracleBlock, index int,
	visit func(*oraclePlace, ssa.Role)) {
	instruction := block.instructions[index]
	for i := range instruction.uses {
		visit(&instruction.uses[i], ssa.Use)
	}
	for i := range instruction.defines {
		visit(&instruction.defines[i], ssa.Define)
	}
}
func (oracleGraph) IsContextStore(f *oracleFunction, block *oracleBlock, index int) bool {
	return block.instructions[index].contextStore
}
func (oracleGraph) ContextStoreDefines(f *oracleFunction, block *oracleBlock, index int, place oraclePlace) bool {
	return place.id == block.instructions[index].result
}
func (oracleGraph) SetInstructionOrder(f *oracleFunction, block *oracleBlock, index int, order ssa.EvaluationOrder) {
	block.instructions[index].order = order
}
func (oracleGraph) EachTerminalPlace(block *oracleBlock, visit func(*oraclePlace, ssa.Role)) {
	for i := range block.terminalUses {
		visit(&block.terminalUses[i], ssa.Use)
	}
	for i := range block.terminalDefines {
		visit(&block.terminalDefines[i], ssa.Define)
	}
}
func (oracleGraph) SetTerminalOrder(block *oracleBlock, order ssa.EvaluationOrder) {
	block.terminalOrder = order
}
func (oracleGraph) Params(f *oracleFunction) []oraclePlace { return f.params }
func (oracleGraph) Returns(f *oracleFunction) *oraclePlace { return f.returns }
func (oracleGraph) Declaration(f *oracleFunction, id ssa.IdentifierId) ssa.DeclarationId {
	return f.identifiers[id].declaration
}
func (oracleGraph) Contextual(f *oracleFunction, declaration ssa.DeclarationId) bool {
	return f.contextual[declaration]
}
func (oracleGraph) Mint(f *oracleFunction, original ssa.IdentifierId) ssa.IdentifierId {
	id := ssa.IdentifierId(len(f.identifiers))
	f.identifiers = append(f.identifiers, f.identifiers[original])
	return id
}
func (oracleGraph) Named(f *oracleFunction, id ssa.IdentifierId) bool {
	return f.identifiers[id].name != ""
}
func (oracleGraph) PlaceString(f *oracleFunction, id ssa.IdentifierId) string {
	return fmt.Sprintf("%s$%d", f.identifiers[id].name, id)
}
func (oracleGraph) IdentifierOf(place oraclePlace) ssa.IdentifierId { return place.id }
func (oracleGraph) WithIdentifier(place oraclePlace, id ssa.IdentifierId) oraclePlace {
	place.id = id
	return place
}

// runOracleFunction runs the passes a function names, as main.ts's run does, and prints what they made.
func runOracleFunction(out *strings.Builder, f *oracleFunction) {
	graph := oracleGraph{}
	ssa.ReversePostorder(graph, f)
	ssa.MarkPredecessors(graph, f)
	ssa.MarkEvaluationOrder(graph, f)
	if f.passes == "construct" {
		ssa.Construct(graph, f)
	}

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
		placeholder := "-"
		if block.placeholder {
			placeholder = "placeholder"
		}
		var predecessors []string
		for _, id := range block.predecessors {
			predecessors = append(predecessors, strconv.Itoa(int(id)))
		}
		returns := "0"
		if block.returns {
			returns = "1"
		}
		line("block %d %s predecessors=%s terminal=%d return=%s", block.id, placeholder, strings.Join(predecessors, ","), block.terminalOrder, returns)
		for _, phi := range block.phis {
			var operands strings.Builder
			for _, operand := range phi.Operands {
				fmt.Fprintf(&operands, " %d=%s", operand.Predecessor, oraclePlaceText(operand.Place))
			}
			line("phi %s%s", oraclePlaceText(phi.Place), operands.String())
		}
		for _, instruction := range block.instructions {
			context := "-"
			if instruction.contextStore {
				context = "context"
			}
			line("instruction %d %s uses=%s defines=%s", instruction.order, context, oraclePlacesText(instruction.uses), oraclePlacesText(instruction.defines))
		}
		line("terminal uses=%s defines=%s", oraclePlacesText(block.terminalUses), oraclePlacesText(block.terminalDefines))
	}
	line("params %s", oraclePlacesText(f.params))
	if f.returns == nil {
		line("returns -")
	} else {
		line("returns %s", oraclePlaceText(*f.returns))
	}
	var found []string
	for id := 0; id < f.bound; id++ {
		if _, ok := f.table[ssa.BlockId(id)]; ok {
			found = append(found, strconv.Itoa(id))
		}
	}
	line("table %s", strings.Join(found, " "))
	line("identifiers %d", len(f.identifiers))
	kinds := map[ssa.SSAViolationKind]string{ssa.SSAViolationMultipleDefinitions: "MultipleDefinitions", ssa.SSAViolationUseNotDominated: "UseNotDominated",
		ssa.SSAViolationEntryHasPredecessors: "EntryHasPredecessors"}
	for _, violation := range ssa.VerifySSA(graph, f) {
		line("violation %s %d %d %s", kinds[violation.Kind], violation.Identifier, violation.Block, violation.Detail)
	}
	stats := ssa.CollectSSAStats(graph, f)
	line("stats %d %d %d", stats.Phis, stats.NamedValues, stats.Uses)
	dominance := ssa.ComputeDominance(graph, f)
	for _, block := range f.blocks {
		var dominators []string
		for _, candidate := range f.blocks {
			if dominance.Dominates(candidate.id, block.id) {
				dominators = append(dominators, strconv.Itoa(int(candidate.id)))
			}
		}
		line("dominators %d=%s", block.id, strings.Join(dominators, ","))
	}
}

func oraclePlaceText(place oraclePlace) string {
	tag := place.tag
	if tag == "" {
		tag = "-"
	}
	return fmt.Sprintf("%d:%s", place.id, tag)
}

func oraclePlacesText(places []oraclePlace) string {
	texts := make([]string, 0, len(places))
	for _, place := range places {
		texts = append(texts, oraclePlaceText(place))
	}
	return strings.Join(texts, ",")
}

// writeOracleFunction writes a function as cases records, the format main.ts reads.
func writeOracleFunction(out *strings.Builder, f *oracleFunction) {
	text := func(value string) string {
		if value == "" {
			return "-"
		}
		return value
	}
	fmt.Fprintf(out, "function %s\nentry %d\nbound %d\n", f.name, f.entry, f.bound)
	for _, identifier := range f.identifiers {
		fmt.Fprintf(out, "identifier %d %s\n", identifier.declaration, text(identifier.name))
	}
	var contextual []int
	for declaration, is := range f.contextual {
		if is {
			contextual = append(contextual, int(declaration))
		}
	}
	sort.Ints(contextual)
	for _, declaration := range contextual {
		fmt.Fprintf(out, "contextual %d\n", declaration)
	}
	for _, param := range f.params {
		fmt.Fprintf(out, "param %d %s\n", param.id, text(param.tag))
	}
	if f.returns != nil {
		fmt.Fprintf(out, "returns %d %s\n", f.returns.id, text(f.returns.tag))
	}
	for _, block := range f.blocks {
		fmt.Fprintf(out, "block %d\n", block.id)
		if block.returns {
			out.WriteString("return\n")
		}
		for _, edge := range block.edges {
			fmt.Fprintf(out, "edge %d %s\n", edge.to, oracleEdgeKinds[edge.kind])
		}
		for _, instruction := range block.instructions {
			fmt.Fprintf(out, "instruction %d\n", instruction.result)
			if instruction.contextStore {
				out.WriteString("context\n")
			}
			for _, use := range instruction.uses {
				fmt.Fprintf(out, "use %d %s\n", use.id, text(use.tag))
			}
			for _, define := range instruction.defines {
				fmt.Fprintf(out, "define %d %s\n", define.id, text(define.tag))
			}
		}
		for _, use := range block.terminalUses {
			fmt.Fprintf(out, "terminal-use %d %s\n", use.id, text(use.tag))
		}
		for _, define := range block.terminalDefines {
			fmt.Fprintf(out, "terminal-define %d %s\n", define.id, text(define.tag))
		}
	}
	fmt.Fprintf(out, "passes %s\n", f.passes)
}

var oracleEdgeKinds = map[ssa.Edge]string{ssa.Real: "Real", ssa.Fallthrough: "Fallthrough", ssa.Exceptional: "Exceptional"}

// parseOracleCases reads cases records back into functions, as main.ts reads them.
func parseOracleCases(cases string) ([]*oracleFunction, error) {
	var functions []*oracleFunction
	var current *oracleFunction
	var block *oracleBlock
	var instruction *oracleInstruction
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
			if fields[index] == "-" {
				return ""
			}
			return fields[index]
		}
		place := func() oraclePlace { return oraclePlace{id: ssa.IdentifierId(number(1)), tag: text(2)} }
		switch fields[0] {
		case "function":
			current = newOracleFunction(text(1))
			block, instruction = nil, nil
		case "entry":
			current.entry = ssa.BlockId(number(1))
		case "bound":
			current.bound = number(1)
		case "identifier":
			current.identifiers = append(current.identifiers, oracleIdentifier{declaration: ssa.DeclarationId(number(1)), name: text(2)})
		case "contextual":
			current.contextual[ssa.DeclarationId(number(1))] = true
		case "param":
			current.params = append(current.params, place())
		case "returns":
			returns := place()
			current.returns = &returns
		case "block":
			block = &oracleBlock{id: ssa.BlockId(number(1))}
			current.blocks = append(current.blocks, block)
			current.table[block.id] = block
			instruction = nil
		case "return":
			block.returns = true
		case "edge":
			kind := -1
			for edge, name := range oracleEdgeKinds {
				if name == fields[2] {
					kind = int(edge)
				}
			}
			if kind < 0 {
				return nil, fmt.Errorf("an unknown edge kind: %s", line)
			}
			block.edges = append(block.edges, oracleEdge{to: ssa.BlockId(number(1)), kind: ssa.Edge(kind)})
		case "instruction":
			instruction = &oracleInstruction{result: ssa.IdentifierId(number(1))}
			block.instructions = append(block.instructions, instruction)
		case "context":
			instruction.contextStore = true
		case "use":
			instruction.uses = append(instruction.uses, place())
		case "define":
			instruction.defines = append(instruction.defines, place())
		case "terminal-use":
			block.terminalUses = append(block.terminalUses, place())
		case "terminal-define":
			block.terminalDefines = append(block.terminalDefines, place())
		case "passes":
			current.passes = text(1)
			functions = append(functions, current)
			current = nil
		default:
			return nil, fmt.Errorf("an unknown record: %s", line)
		}
	}
	return functions, nil
}

// oracleFromTest is a module test's function as an oracle function, unchanged: its blocks in their
// order, its edges, instructions, identifiers, parameters, returns and contextual bindings.
func oracleFromTest(name string, f *testFunction, passes string) *oracleFunction {
	oracle := newOracleFunction(name)
	oracle.passes = passes
	oracle.entry = f.entry
	oracle.bound = int(f.next)
	for _, identifier := range f.identifiers {
		oracle.identifiers = append(oracle.identifiers, oracleIdentifier{declaration: identifier.declaration, name: identifier.name})
	}
	for declaration, is := range f.contextual {
		oracle.contextual[declaration] = is
	}
	for _, param := range f.params {
		oracle.params = append(oracle.params, oraclePlace{id: param.id, tag: param.tag})
	}
	if f.returns != nil {
		oracle.returns = &oraclePlace{id: f.returns.id, tag: f.returns.tag}
	}
	for _, block := range f.blocks {
		copied := &oracleBlock{id: block.id, returns: block.returns}
		for _, edge := range block.edges {
			copied.edges = append(copied.edges, oracleEdge{to: edge.to, kind: edge.kind})
		}
		for _, instruction := range block.instructions {
			made := &oracleInstruction{contextStore: instruction.contextStore, result: instruction.result}
			for _, use := range instruction.uses {
				made.uses = append(made.uses, oraclePlace{id: use.id, tag: use.tag})
			}
			for _, define := range instruction.defines {
				made.defines = append(made.defines, oraclePlace{id: define.id, tag: define.tag})
			}
			copied.instructions = append(copied.instructions, made)
		}
		oracle.blocks = append(oracle.blocks, copied)
		oracle.table[copied.id] = copied
	}
	return oracle
}

// moduleTest is one of the module's own tests' functions, built as that test builds it, and the passes
// that test runs on it (construct is finalize, then construction).
type moduleTest struct {
	name     string
	function *testFunction
	passes   string
}

// moduleTestFunctions builds each function static_single_assignment_test.go builds, with its helpers,
// line for line as each test does, before its passes.
func moduleTestFunctions() []moduleTest {
	var tests []moduleTest

	{ // TestDiamondPlacesOnePhiAtTheJoin
		f := newTestFunction()
		entry, left, right, join := f.block(), f.block(), f.block(), f.block()
		f.entry = entry.id
		f.define(entry, "x")
		f.define(left, "x")
		f.use(join, "x")
		goTo(entry, left, right)
		goTo(left, join)
		goTo(right, join)
		tests = append(tests, moduleTest{"diamond-places-one-phi-at-the-join", f, "construct"})
	}
	{ // TestLoopPhiTakesTheBackEdge
		f := newTestFunction()
		entry, header, body, exit := f.block(), f.block(), f.block(), f.block()
		f.entry = entry.id
		f.define(entry, "i")
		f.use(header, "i")
		f.define(body, "i")
		f.use(exit, "i")
		goTo(entry, header)
		goTo(header, body, exit)
		goTo(body, header)
		tests = append(tests, moduleTest{"loop-phi-takes-the-back-edge", f, "construct"})
	}
	{ // TestLoopThatNeverWritesLeavesNoPhi
		f := newTestFunction()
		entry, header, body, exit := f.block(), f.block(), f.block(), f.block()
		f.entry = entry.id
		f.define(entry, "i")
		f.use(header, "i")
		f.use(body, "i")
		goTo(entry, header)
		goTo(header, body, exit)
		goTo(body, header)
		tests = append(tests, moduleTest{"loop-that-never-writes-leaves-no-phi", f, "construct"})
	}
	{ // TestUnreachableBlockIsDropped
		f := newTestFunction()
		entry, exit, orphan := f.block(), f.block(), f.block()
		f.entry = entry.id
		goTo(entry, exit)
		goTo(orphan, exit)
		tests = append(tests, moduleTest{"unreachable-block-is-dropped", f, "finalize"})
	}
	{ // TestFallthroughNothingReachesBecomesAPlaceholder
		f := newTestFunction()
		entry, body, continuation := f.block(), f.block(), f.block()
		f.entry = entry.id
		entry.edges = []testEdge{{to: continuation.id, kind: ssa.Fallthrough}, {to: body.id, kind: ssa.Real}}
		body.returns = true
		tests = append(tests, moduleTest{"fallthrough-nothing-reaches-becomes-a-placeholder", f, "finalize"})
	}
	{ // TestContextStoreReusesItsDefinition
		f := newTestFunction()
		entry := f.block()
		f.entry = entry.id
		first := f.define(entry, "n")
		first.contextStore = true
		f.contextual[f.variables["n"]] = true
		second := f.define(entry, "n")
		second.contextStore = true
		tests = append(tests, moduleTest{"context-store-reuses-its-definition", f, "construct"})
	}
	{ // TestExceptionalEdgeIsAnEdge
		f := newTestFunction()
		entry, next, handler := f.block(), f.block(), f.block()
		f.entry = entry.id
		f.define(entry, "x")
		f.define(entry, "x")
		f.use(handler, "x")
		entry.edges = []testEdge{{to: next.id, kind: ssa.Real}, {to: handler.id, kind: ssa.Exceptional}}
		tests = append(tests, moduleTest{"exceptional-edge-is-an-edge", f, "construct"})
	}
	{ // TestEvaluationOrderFollowsTheSlice
		f := newTestFunction()
		entry, exit := f.block(), f.block()
		f.entry = entry.id
		f.define(entry, "a")
		f.use(exit, "a")
		goTo(entry, exit)
		tests = append(tests, moduleTest{"evaluation-order-follows-the-slice", f, "finalize"})
	}
	{ // TestVerifierCatchesAnUndominatedUse
		f := newTestFunction()
		entry, left, right := f.block(), f.block(), f.block()
		f.entry = entry.id
		write := f.define(left, "x")
		read := f.use(right, "x")
		read.uses[0].id = write.defines[0].id
		goTo(entry, left, right)
		tests = append(tests, moduleTest{"verifier-catches-an-undominated-use", f, "finalize"})
	}
	{ // TestDominanceIsTheEntrysAndTheJoins
		f := newTestFunction()
		entry, left, right, join := f.block(), f.block(), f.block(), f.block()
		f.entry = entry.id
		goTo(entry, left, right)
		goTo(left, join)
		goTo(right, join)
		tests = append(tests, moduleTest{"dominance-is-the-entrys-and-the-joins", f, "finalize"})
	}
	return tests
}

// generatedFunction makes a function of the shapes the passes meet, from the random source: diamonds,
// loops and self loops, unreachable blocks, structural fallthroughs, exceptional edges, edges to blocks
// that aren't there, block ids with gaps, parameters, a returns place, captured bindings written by
// context stores, terminals with places, and places naming a value directly (a temporary read where its
// definition may not dominate, or defined a second time, which the verifier has to judge).
func generatedFunction(name string, random *rand.Rand) *oracleFunction {
	f := newOracleFunction(name)
	f.passes = "construct"
	if random.Intn(7) == 0 {
		f.passes = "finalize"
	}
	f.identifiers = append(f.identifiers, oracleIdentifier{})

	variableNames := []string{"a", "b", "c", "d"}[:1+random.Intn(4)]
	declarations := map[string]ssa.DeclarationId{}
	nextDeclaration := ssa.DeclarationId(1)
	for _, variable := range variableNames {
		declarations[variable] = nextDeclaration
		if random.Intn(5) == 0 {
			f.contextual[nextDeclaration] = true
		}
		nextDeclaration++
	}
	value := func(variable string) oraclePlace {
		id := ssa.IdentifierId(len(f.identifiers))
		f.identifiers = append(f.identifiers, oracleIdentifier{declaration: declarations[variable], name: variable})
		return oraclePlace{id: id, tag: variable}
	}
	// A temporary is a value of a binding of its own, with no name.
	var temporaries []ssa.IdentifierId
	temporary := func() oraclePlace {
		id := ssa.IdentifierId(len(f.identifiers))
		f.identifiers = append(f.identifiers, oracleIdentifier{declaration: nextDeclaration})
		nextDeclaration++
		temporaries = append(temporaries, id)
		return oraclePlace{id: id, tag: "t"}
	}
	pick := func() string { return variableNames[random.Intn(len(variableNames))] }

	for count := random.Intn(3); count > 0; count-- {
		f.params = append(f.params, value(pick()))
	}
	hasReturns := random.Intn(2) == 0
	if hasReturns {
		declarations["result"] = nextDeclaration
		nextDeclaration++
		returns := value("result")
		f.returns = &returns
	}

	blockCount := 1 + random.Intn(10)
	var ids []ssa.BlockId
	next := ssa.BlockId(1)
	for index := 0; index < blockCount; index++ {
		if random.Intn(6) == 0 {
			next++
		}
		ids = append(ids, next)
		next++
	}
	f.bound = int(next) + random.Intn(2)
	f.entry = ids[0]
	if random.Intn(15) == 0 {
		f.entry = ids[random.Intn(len(ids))]
	}
	// No edge enters the entry, in either real IR, and construction relies on it: it refuses a function
	// whose entry has predecessors, where its lookup around a cycle of single-predecessor blocks through
	// the entry would never end. So only a function the finalizer alone runs on, which the verifier
	// judges, may have an edge into its entry.
	target := func() ssa.BlockId {
		for {
			var id ssa.BlockId
			switch random.Intn(30) {
			case 0:
				id = ssa.BlockId(f.bound + 2)
			case 1:
				id = ssa.BlockId(random.Intn(f.bound))
			default:
				id = ids[random.Intn(len(ids))]
			}
			if id != f.entry || f.passes == "finalize" {
				return id
			}
		}
	}

	for _, id := range ids {
		block := &oracleBlock{id: id}
		f.blocks = append(f.blocks, block)
		f.table[id] = block

		for count := random.Intn(7); count > 0; count-- {
			instruction := &oracleInstruction{}
			switch random.Intn(9) {
			case 0, 1: // a store: x = ...
				instruction.defines = append(instruction.defines, value(pick()))
			case 2: // a load into a temporary
				instruction.uses = append(instruction.uses, value(pick()))
				instruction.defines = append(instruction.defines, temporary())
			case 3: // x = x + 1
				variable := pick()
				instruction.uses = append(instruction.uses, value(variable))
				instruction.defines = append(instruction.defines, value(variable))
			case 4: // a context store, its own result a temporary or the binding's value
				variable := pick()
				written := value(variable)
				instruction.contextStore = true
				instruction.defines = append(instruction.defines, written)
				if random.Intn(2) == 0 {
					result := temporary()
					instruction.defines = append(instruction.defines, result)
					instruction.result = result.id
				} else {
					instruction.result = written.id
				}
			case 5: // a temporary read by its value, wherever it was defined
				if len(temporaries) > 0 {
					instruction.uses = append(instruction.uses, oraclePlace{id: temporaries[random.Intn(len(temporaries))], tag: "t"})
				}
				instruction.defines = append(instruction.defines, temporary())
			case 6: // a write of the returns place
				if hasReturns {
					instruction.defines = append(instruction.defines, value("result"))
				} else {
					instruction.uses = append(instruction.uses, value(pick()))
				}
			case 7: // a read of a variable
				instruction.uses = append(instruction.uses, value(pick()))
			case 8: // a temporary defined again, which only the verifier should notice
				if len(temporaries) > 0 {
					instruction.defines = append(instruction.defines, oraclePlace{id: temporaries[random.Intn(len(temporaries))], tag: "t"})
				} else {
					instruction.defines = append(instruction.defines, temporary())
				}
			}
			block.instructions = append(block.instructions, instruction)
		}

		if random.Intn(7) == 0 {
			block.edges = append(block.edges, oracleEdge{to: target(), kind: ssa.Fallthrough})
		}
		for count := random.Intn(3); count > 0; count-- {
			block.edges = append(block.edges, oracleEdge{to: target(), kind: ssa.Real})
		}
		if random.Intn(10) == 0 {
			block.edges = append(block.edges, oracleEdge{to: target(), kind: ssa.Exceptional})
		}
		if !slices.ContainsFunc(block.edges, func(edge oracleEdge) bool { return edge.kind != ssa.Fallthrough }) && random.Intn(4) != 0 {
			block.returns = true
		}
		if random.Intn(3) == 0 {
			block.terminalUses = append(block.terminalUses, value(pick()))
		}
		if random.Intn(20) == 0 {
			block.terminalDefines = append(block.terminalDefines, temporary())
		}
	}
	return f
}
