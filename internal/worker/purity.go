package worker

import (
	"fmt"
	"reflect"

	"github.com/system-inc/adamic/internal/ir"
)

// CheckPure is the conservative seam for the forthcoming split analysis.
// Reflection visits new IR operands automatically, so an added expression cannot
// silently hide an I/O operation or global read from this interim check.
func CheckPure(program *ir.Program, function int) error {
	visited := make(map[int]bool)
	var inspect func(int) error
	var walk func(reflect.Value) error
	inspect = func(index int) error {
		if visited[index] {
			return nil
		}
		visited[index] = true
		if err := walk(reflect.ValueOf(program.Functions[index].Body)); err != nil {
			function := program.Functions[index]
			position := function.Position
			return fmt.Errorf("%s:%d:%d: %s: %w", position.File, position.Line, position.Column, function.Name, err)
		}
		return nil
	}
	global := func(local int) error {
		if program.Locals[local].Global {
			return fmt.Errorf("accesses module global %s", program.Locals[local].Name)
		}
		return nil
	}
	walk = func(value reflect.Value) error {
		if !value.IsValid() {
			return nil
		}
		if value.CanInterface() {
			switch node := value.Interface().(type) {
			case ir.WriteLine:
				return fmt.Errorf("calls console")
			case ir.ReadTextFile, ir.WriteTextFile, ir.ReadDirectory, ir.FileStatus, ir.ProgramArguments:
				return fmt.Errorf("calls adamic I/O (%T)", node)
			case ir.NodeFSFile:
				return fmt.Errorf("impure: host operation %s", node.Operation)
			case ir.NodeHostCall:
				return fmt.Errorf("impure: host operation %s.%s", node.Module, node.Member)
			case ir.ProcessCall:
				operation := "process." + node.Operation
				if node.Operation == "cwd" {
					operation = "host_cwd"
				}
				return fmt.Errorf("impure: host operation %s", operation)
			case ir.Read:
				return global(node.Local)
			case ir.Assign:
				if err := global(node.Local); err != nil {
					return err
				}
			case ir.Call:
				for _, target := range program.CallTargets(node) {
					if err := inspect(target); err != nil {
						return err
					}
				}
			case ir.CallClosure, ir.ArrayMap, ir.ArrayVisit, ir.ArrayReduce, ir.ArrayFrom, ir.MapForEach, ir.ArraySort:
				targets := program.ClosureTargets(node.(ir.Expression))
				if targets.Unknown {
					return fmt.Errorf("calls an unbounded function value")
				}
				for _, target := range targets.Functions {
					if err := inspect(target); err != nil {
						return err
					}
				}
			}
		}
		switch value.Kind() {
		case reflect.Interface, reflect.Pointer:
			if !value.IsNil() {
				return walk(value.Elem())
			}
		case reflect.Struct:
			for index := 0; index < value.NumField(); index++ {
				if err := walk(value.Field(index)); err != nil {
					return err
				}
			}
		case reflect.Slice, reflect.Array:
			for index := 0; index < value.Len(); index++ {
				if err := walk(value.Index(index)); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := inspect(function); err != nil {
		return fmt.Errorf("worker: refused Wasm function %w; fix: remove I/O and module-global access and use statically known callees", err)
	}
	return nil
}
