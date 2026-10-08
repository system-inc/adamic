//go:build lintoracle

// Overlay beside Go HIR. Only this lane's pass boundaries and sidecars live here.
package high_level_intermediate_representation

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func init() {
	OracleRegisterConstructionObserver(unit4Observe)
}

func unit4PrimitiveRows(f *Function) ([]OracleExtraIdentity, []OracleSidecarRow, error) {
	primitive := inferPrimitivePropertyReads(f)
	var ids []int
	for id, yes := range primitive {
		if yes {
			ids = append(ids, int(id))
		}
	}
	sort.Ints(ids)
	var rows []OracleSidecarRow
	for _, id := range ids {
		id := id
		rows = append(rows, OracleSidecarRow{Namespace: "unit4.primitive", FunctionPath: "$", AnchorKind: "identifier", AnchorId: &id, Key: "property-read", Payload: "1"})
	}
	return nil, rows, nil
}

func unit4AstRows(f *Function) ([]OracleExtraIdentity, []OracleSidecarRow, error) {
	var rows []OracleSidecarRow
	add := func(path, kind string, id int, key, payload string) {
		rows = append(rows, OracleSidecarRow{Namespace: "input.unit4-ast", FunctionPath: path, AnchorKind: kind, AnchorId: &id, Key: key, Payload: payload})
	}
	var visit func(*Function, string)
	visit = func(fn *Function, path string) {
		for _, identifier := range fn.Identifiers {
			text := ""
			if identifier.Node != nil && ast.IsIdentifier(identifier.Node) {
				text = identifier.Node.Text()
			}
			add(path, "identifier", int(identifier.Id), "identifier-text", text)
		}
		for _, instruction := range fn.Instructions {
			name := ""
			switch value := instruction.Value.(type) {
			case *CallExpression:
				name = calleeName(fn, instruction, value.Callee)
			case *MethodCall:
				name = calleeName(fn, instruction, value.Property)
			case *NewExpression:
				name = calleeName(fn, instruction, value.Callee)
			}
			add(path, "instruction", int(instruction.Id), "callee-name", name)
			add(path, "instruction", int(instruction.Id), "receiver-name", receiverSyntaxName(instruction))
		}
		for i, nested := range fn.Functions {
			visit(nested, fmt.Sprintf("%s/%d", path, i))
		}
	}
	visit(f, "$")
	return nil, rows, nil
}
func unit4Observe(key string, f *Function, tc *checker.Checker, caller string) error {
	destination := os.Getenv("HIR_UNIT4_EXPORT")
	if destination == "" {
		return nil
	}
	if err := os.MkdirAll(destination, 0755); err != nil {
		return err
	}
	write := func(pass, boundary string, extra func(*Function) ([]OracleExtraIdentity, []OracleSidecarRow, error)) error {
		encode := OracleFunctionEncoder(key, "unit4."+pass, tc, extra)
		data, err := encode(f)
		if err != nil {
			return err
		}
		repeat, err := encode(f)
		if err != nil {
			return err
		}
		if string(data) != string(repeat) {
			return fmt.Errorf("unit4 %s/%s nondeterministic: %s", pass, boundary, key)
		}
		return os.WriteFile(filepath.Join(destination, key+"."+pass+"."+boundary+".checkpoint"), data, 0600)
	}
	if err := write("primitive", "before", nil); err != nil {
		return err
	}
	if err := write("primitive", "after", unit4PrimitiveRows); err != nil {
		return err
	}
	if err := write("reactive", "before", unit4AstRows); err != nil {
		return err
	}
	effectInput := CloneFunction(f)
	InferReactive(f, tc)
	if err := write("reactive", "after", unit4AstRows); err != nil {
		return err
	}
	return unit4EffectBoundaries(key, effectInput, tc, destination)
}

func unit4EffectRows(fn *Function, path string, effects *AliasingEffects) []OracleSidecarRow {
	var rows []OracleSidecarRow
	for _, instruction := range fn.Instructions {
		id := int(instruction.Id)
		for i, effect := range effects.Get(instruction.Id) {
			rows = append(rows, OracleSidecarRow{Namespace: "unit4.effects", FunctionPath: path, AnchorKind: "instruction", AnchorId: &id, Key: fmt.Sprintf("%06d", i), Payload: fmt.Sprintf("%s\t%s\t%s\t%s", effect.Kind, oraclePlace(effect.From), oraclePlace(effect.Into), effect.Value)})
		}
	}
	return rows
}
func unit4EffectBoundaries(key string, original *Function, tc *checker.Checker, destination string) error {
	for _, pass := range []string{"effects", "effects-nested"} {
		fn := CloneFunction(original)
		before, err := OracleFunctionEncoder(key, "unit4."+pass, tc, unit4AstRows)(fn)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(destination, key+"."+pass+".before.checkpoint"), before, 0600); err != nil {
			return err
		}
		var rows []OracleSidecarRow
		if pass == "effects" {
			effects := InferAliasingEffects(fn)
			rows = unit4EffectRows(fn, "$", effects)
		} else {
			effects := InferAliasingEffectsForNested(fn)
			for i, nested := range fn.Functions {
				rows = append(rows, unit4EffectRows(nested, fmt.Sprintf("$/%d", i), effects[FunctionId(i)])...)
			}
		}
		extra := func(f *Function) ([]OracleExtraIdentity, []OracleSidecarRow, error) {
			ids, inputs, err := unit4AstRows(f)
			return ids, append(inputs, rows...), err
		}
		encode := OracleFunctionEncoder(key, "unit4."+pass, tc, extra)
		after, err := encode(fn)
		if err != nil {
			return err
		}
		repeat, err := encode(fn)
		if err != nil {
			return err
		}
		if string(after) != string(repeat) {
			return fmt.Errorf("unit4 %s nondeterministic: %s", pass, key)
		}
		if err := os.WriteFile(filepath.Join(destination, key+"."+pass+".after.checkpoint"), after, 0600); err != nil {
			return err
		}
	}
	return nil
}
