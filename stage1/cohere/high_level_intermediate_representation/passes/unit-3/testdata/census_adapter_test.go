//go:build lintoracle

package high_level_intermediate_representation

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

var unit3FixtureMutex sync.Mutex

func init() {
	if os.Getenv("HIR_UNIT3_EXPORT") != "" {
		OracleRegisterConstructionObserver(unit3ObserveCensus)
	}
}
func unit3SyntaxRows(f *Function) []OracleSidecarRow {
	rows := []OracleSidecarRow{}
	var visit func(*Function, string)
	visit = func(fn *Function, path string) {
		for _, identifier := range fn.Identifiers {
			if identifier.Node != nil && ast.IsIdentifier(identifier.Node) {
				id := int(identifier.Id)
				rows = append(rows, OracleSidecarRow{Namespace: "unit3.ast", FunctionPath: path, AnchorKind: "identifier", AnchorId: &id, Key: "text", Payload: identifier.Node.Text()})
			}
		}
		for i, child := range fn.Functions {
			visit(child, fmt.Sprintf("%s/%d", path, i))
		}
	}
	visit(f, "$")
	return rows
}
func unit3Result(result any) string {
	switch value := result.(type) {
	case int:
		return fmt.Sprint(value)
	case ManualMemoization:
		return fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d", value.Recognised, value.Marked, value.Dependencies, value.WithoutDepsArray, value.UnextractableDeps, value.NotAnArrayLiteral, value.NotAnInlineFunction)
	case DeadCodeEliminationResult:
		return fmt.Sprintf("%d,%d", value.Instructions, value.Phis)
	case []FunctionId:
		fields := []string{}
		for _, id := range value {
			fields = append(fields, fmt.Sprint(id))
		}
		return strings.Join(fields, ",")
	case struct {
		Remap  *InlineRemap
		Copied bool
	}:
		if !value.Copied {
			return "false"
		}
		var body strings.Builder
		fmt.Fprintf(&body, "true\nentry %d", value.Remap.Entry)
		identifiers := []int{}
		for id := range value.Remap.Identifiers {
			identifiers = append(identifiers, int(id))
		}
		sort.Ints(identifiers)
		for _, id := range identifiers {
			for old, mapped := range value.Remap.Identifiers {
				if int(old) == id {
					fmt.Fprintf(&body, "\nidentifier %d=%d", id, mapped)
				}
			}
		}
		blocks := []int{}
		for id := range value.Remap.Blocks {
			blocks = append(blocks, int(id))
		}
		sort.Ints(blocks)
		for _, id := range blocks {
			for old, mapped := range value.Remap.Blocks {
				if int(old) == id {
					fmt.Fprintf(&body, "\nblock %d=%d", id, mapped)
				}
			}
		}
		instructions := []int{}
		for id := range value.Remap.Instructions {
			instructions = append(instructions, int(id))
		}
		sort.Ints(instructions)
		for _, id := range instructions {
			for old, mapped := range value.Remap.Instructions {
				if int(old) == id {
					fmt.Fprintf(&body, "\ninstruction %d=%d", id, mapped)
				}
			}
		}
		return body.String()
	default:
		panic(fmt.Sprintf("unknown unit-3 result %T", result))
	}
}
func unit3ObserveCensus(key string, f *Function, tc *checker.Checker, caller string) error {
	var result any
	run := func(pass string, fn, nested *Function, captures []Place, nestedOrdinal string) error {
		result = nil
		extra := func(current *Function) ([]OracleExtraIdentity, []OracleSidecarRow, error) {
			rows := unit3SyntaxRows(current)
			if strings.Split(pass, ":")[0] == "inline_remap" {
				places := []string{}
				for _, capture := range captures {
					places = append(places, oraclePlace(capture))
				}
				rows = append(rows, OracleSidecarRow{Namespace: "unit3.input", FunctionPath: "$", AnchorKind: "function", Key: "nested", Payload: nestedOrdinal}, OracleSidecarRow{Namespace: "unit3.input", FunctionPath: "$", AnchorKind: "function", Key: "captures", Payload: strings.Join(places, " ")})
			}
			if result != nil {
				rows = append(rows, OracleSidecarRow{Namespace: "unit3.result", FunctionPath: "$", AnchorKind: "function", Key: "result", Payload: unit3Result(result)})
			}
			return nil, rows, nil
		}
		encode := OracleFunctionEncoder(key, pass, tc, extra)
		checkpoint, err := unit3Observe(strings.Split(pass, ":")[0], fn, nested, captures, encode)
		if err != nil {
			return fmt.Errorf("%s/%s (%s): %w", key, pass, caller, err)
		}
		result = checkpoint.Result
		// The observer snapshots immediately after Go; add the returned counts there.
		after, err := encode(fn)
		if err != nil {
			return err
		}
		destination := filepath.Join(os.Getenv("HIR_UNIT3_EXPORT"), key, pass)
		if err := os.MkdirAll(destination, 0755); err != nil {
			return err
		}
		for name, body := range map[string][]byte{"before.checkpoint": checkpoint.Before, "after.checkpoint": after} {
			if err := os.WriteFile(filepath.Join(destination, name), body, 0600); err != nil {
				return err
			}
		}
		return nil
	}
	unit3FixtureMutex.Lock()
	defer unit3FixtureMutex.Unlock()
	original := CloneFunction(f)
	if err := run("outline_functions", f, nil, nil, ""); err != nil {
		return err
	}
	InferReactive(f, tc)
	if err := run("drop_manual_memoization", f, nil, nil, ""); err != nil {
		return err
	}
	guarded := CloneFunction(f)
	if err := run("inline_iife", guarded, nil, nil, ""); err != nil {
		return err
	}
	invoked := CloneFunction(f)
	if err := run("inline_iife_including_memo_callbacks", f, nil, nil, ""); err != nil {
		return err
	}
	// This is the exact preservation merge boundary. Direct no-op calls are still
	// recorded when the enclosing pipeline would skip them, retaining all records.
	inlined := result.(int)
	if err := run("merge_consecutive_blocks", CloneFunction(f), nil, nil, ""); err != nil {
		return err
	}
	if inlined > 0 {
		MergeConsecutiveBlocks(f)
	}
	if err := run("dead_code_elimination", f, nil, nil, ""); err != nil {
		return err
	}
	if err := run("invoked_functions", invoked, nil, nil, ""); err != nil {
		return err
	}
	if len(original.Functions) == 0 {
		return run("inline_remap", original, nil, nil, "absent")
	}
	for ordinal := range original.Functions {
		parent := CloneFunction(original)
		nested := parent.Functions[ordinal]
		var captures []Place
		for _, instruction := range parent.Instructions {
			if value, ok := instruction.Value.(*FunctionExpression); ok && int(value.Function) == ordinal {
				captures = value.Captures
				break
			}
		}
		label := "inline_remap"
		if ordinal > 0 {
			label += fmt.Sprintf(":%d", ordinal)
		}
		if err := run(label, parent, nested, captures, fmt.Sprint(ordinal)); err != nil {
			return err
		}
	}
	return nil
}
