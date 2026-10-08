//go:build lintoracle

// Overlay beside Go HIR. Only syntax/checker INPUT facts belong here.
package high_level_intermediate_representation

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/static_single_assignment"
	"reflect"
	"sort"
	"strings"
)

// Register during adapter init, before tests run. A worker receives a fresh graph;
// it chooses its own real Go pipeline/boundaries and retains errors as outcomes.
type OracleConstructionObserver func(key string, function *Function, typeChecker *checker.Checker, caller string) error

var oracleConstructionObservers []OracleConstructionObserver

func OracleRegisterConstructionObserver(observer OracleConstructionObserver) {
	oracleConstructionObservers = append(oracleConstructionObservers, observer)
}
func OracleObserveConstruction(key string, f *Function, tc *checker.Checker, caller string) error {
	for _, observe := range oracleConstructionObservers {
		if err := observe(key, CloneFunction(f), tc, caller); err != nil {
			return err
		}
	}
	return nil
}
func OracleInputFacts(f *Function, tc *checker.Checker) []OracleSidecarRow {
	var rows []OracleSidecarRow
	add := func(path, namespace, kind string, id *int, key, payload string) {
		rows = append(rows, OracleSidecarRow{Namespace: namespace, FunctionPath: path, AnchorKind: kind, AnchorId: id, Key: key, Payload: payload})
	}
	node := func(path, kind string, id *int, n *ast.Node) {
		payload := "nil"
		if n != nil {
			payload = fmt.Sprintf("%s\t%d\t%d", n.Kind.String(), n.Pos(), n.End())
		}
		add(path, "input.ast", kind, id, "node", payload)
	}
	var visit func(*Function, string)
	visit = func(fn *Function, path string) {
		available := "0"
		if tc != nil {
			available = "1"
		}
		add(path, "input.checker", "function", nil, "available", available)
		node(path, "function", nil, fn.Node)
		add(path, "input.identity", "function", nil, "next-block", fmt.Sprint(fn.nextBlock))
		retained := map[int]bool{}
		for _, block := range fn.Blocks {
			retained[int(block.Id)] = true
		}
		ids := []int{}
		for id := range fn.blocksById {
			if !retained[int(id)] {
				ids = append(ids, int(id))
			}
		}
		sort.Ints(ids)
		for _, id := range ids {
			block := fn.blocksById[static_single_assignment.BlockId(id)]
			var body strings.Builder
			fmt.Fprintf(&body, "block %d %s predecessors=", id, block.Kind)
			for i, pred := range block.Predecessors {
				if i > 0 {
					body.WriteByte(',')
				}
				fmt.Fprint(&body, pred)
			}
			body.WriteByte('\n')
			for _, phi := range block.Phis {
				fmt.Fprintf(&body, "phi %s", oraclePlace(phi.Place))
				for _, op := range phi.Operands {
					fmt.Fprintf(&body, " %d=%s", op.Predecessor, oraclePlace(op.Place))
				}
				body.WriteByte('\n')
			}
			for _, instruction := range block.Instructions {
				fmt.Fprintf(&body, "instruction-ref %d\n", instruction)
			}
			if block.Terminal == nil {
				body.WriteString("terminal-null\n")
			} else {
				payload := oraclePayload(block.Terminal)
				if r, ok := block.Terminal.(*Return); ok {
					payload = oraclePlace(r.Value)
				}
				fmt.Fprintf(&body, "terminal %d %s %s\n", TerminalOrder(block.Terminal), reflect.TypeOf(block.Terminal).Elem().Name(), payload)
			}
			add(path, "input.blocks", "function", nil, fmt.Sprint(id), body.String())
		}
		for _, identifier := range fn.Identifiers {
			id := int(identifier.Id)
			node(path, "identifier", &id, identifier.Node)
			aliasName, symbolName := "", ""
			aliasPresent, typePresent := "0", "0"
			if tc != nil && identifier.Node != nil {
				valueType := tc.GetTypeAtLocation(identifier.Node)
				if valueType != nil {
					if alias := checker.Type_alias(valueType); alias != nil {
						if symbol := alias.Symbol(); symbol != nil {
							aliasName = symbol.Name
							aliasPresent = "1"
						}
					}
					if symbol := checker.Type_symbol(valueType); symbol != nil {
						symbolName = symbol.Name
						typePresent = "1"
					}
				}
			}
			add(path, "input.types", "identifier", &id, "alias-present", aliasPresent)
			add(path, "input.types", "identifier", &id, "type-present", typePresent)
			add(path, "input.types", "identifier", &id, "alias-symbol", aliasName)
			add(path, "input.types", "identifier", &id, "type-symbol", symbolName)
		}
		for _, instruction := range fn.Instructions {
			id := int(instruction.Id)
			node(path, "instruction", &id, instruction.Node)
			if d, ok := instruction.Value.(*Destructure); ok {
				same := "0"
				if d.Pattern == d.LValue {
					same = "1"
				}
				add(path, "input.pattern", "instruction", &id, "lvalue-shares-pattern", same)
			}
		}
		for i, nested := range fn.Functions {
			visit(nested, fmt.Sprintf("%s/%d", path, i))
		}
	}
	visit(f, "$")
	return rows
}

// A pass's own sidecar encoder runs afresh at each boundary, including failures.
// Include all required analysis tables in extra; construction scopes '-' is no substitute.
func OracleFunctionEncoder(key, pass string, tc *checker.Checker, extra func(*Function) ([]OracleExtraIdentity, []OracleSidecarRow, error)) func(*Function) ([]byte, error) {
	return func(f *Function) ([]byte, error) {
		if f == nil {
			return nil, fmt.Errorf("%s: nil function", pass)
		}
		c := OracleCheckpoint{Key: key, Pass: pass, Graph: oracleDump(f), Sidecars: OracleInputFacts(f, tc)}
		c.Identities = OracleDormantIdentities(f)
		if source := ast.GetSourceFileOfNode(f.Node); source != nil {
			extension := ".ts"
			switch source.ScriptKind {
			case core.ScriptKindTSX:
				extension = ".tsx"
			case core.ScriptKindJSX:
				extension = ".jsx"
			case core.ScriptKindJS:
				extension = ".js"
			}
			c.Sidecars = append(c.Sidecars, OracleSourceFacts(source.Text(), extension)...)
		}
		if extra != nil {
			ids, rows, err := extra(f)
			if err != nil {
				return nil, err
			}
			c.Identities = append(OracleDormantIdentities(f), ids...)
			c.Sidecars = append(c.Sidecars, rows...)
		}
		return []byte(OracleWriteCheckpoint(c)), nil
	}
}
func OracleSourceFacts(source, extension string) []OracleSidecarRow {
	return []OracleSidecarRow{{Namespace: "input.source", FunctionPath: "$", AnchorKind: "function", Key: "kind", Payload: strings.TrimPrefix(extension, ".")}, {Namespace: "input.source", FunctionPath: "$", AnchorKind: "function", Key: "text", Payload: source}}
}

func OracleDormantIdentities(f *Function) []OracleExtraIdentity {
	var ids []OracleExtraIdentity
	var visit func(*Function, string)
	visit = func(fn *Function, path string) {
		retained := map[int]bool{}
		for _, b := range fn.Blocks {
			retained[int(b.Id)] = true
		}
		for id := 1; id < int(fn.nextBlock); id++ {
			if !retained[id] {
				ids = append(ids, OracleExtraIdentity{FunctionPath: path, Kind: "block", Id: id})
			}
		}
		for i, nested := range fn.Functions {
			visit(nested, fmt.Sprintf("%s/%d", path, i))
		}
	}
	visit(f, "$")
	return ids
}
