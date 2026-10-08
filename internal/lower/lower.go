// Package lower turns a checked program into Adamic's IR.
//
// Stage 0 lowers a little and refuses the rest. Every construct it can't lower yet is an error
// naming the construct and where it is, never a skip: a compiler that drops a statement it doesn't
// understand produces a program that looks right and isn't, which is the one thing Adamic must
// never do.
package lower

import (
	"context"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"path/filepath"
)

// Lower lowers a checked program from one entry, in ESM evaluation order.
func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {
	files := program.Files()
	if len(files) != 1 {
		return nil, fmt.Errorf("lower: stage 0 compiles a program from one entry file, got %d", len(files))
	}
	entry := files[0]
	// Stage 0 checks single-threaded, so one checker answers for every file.
	typeChecker, release := program.Checker(ctx, entry)
	defer release()

	lowering := &lowering{program: program, checker: typeChecker, result: &ir.Program{}, this: -1, functionIndex: -1}
	// The base name only, so the same program emits the same C on every machine.
	lowering.result.Source = filepath.Base(program.FileName(entry))
	modules, err := lowering.moduleOrder(entry)
	if err != nil {
		return nil, err
	}
	lowering.noteInheritance(modules)
	if err := lowering.enumInitialization(modules); err != nil {
		return nil, err
	}
	lowering.noteAccessorNames(modules)
	lowering.noteOmittedOptionals(modules)
	if err := lowering.namespaceInitialization(modules); err != nil {
		return nil, err
	}
	for _, module := range modules {
		if err := lowering.refuse(module); err != nil {
			return nil, err
		}
	}
	// Link every declaration before lowering any function body, including across back edges.
	var declarations []*ast.Node
	for _, module := range modules {
		declarations = append(declarations, module.Statements.Nodes...)
	}
	if err := lowering.declareModule(declarations); err != nil {
		return nil, err
	}
	for _, module := range modules {
		body, err := lowering.statements(module.Statements.Nodes)
		if err != nil {
			return nil, err
		}
		lowering.result.Main = append(lowering.result.Main, body...)
	}
	if lowering.unlowerable != nil {
		return nil, lowering.unlowerable
	}
	lowering.result.Main = append(lowering.forwarderValues, lowering.result.Main...)
	lowering.finishClassCalls()
	if err := lowering.finishAccessors(); err != nil {
		return nil, err
	}
	if err := lowering.exceptions(); err != nil {
		return nil, err
	}
	if err := lowering.checkAccessorSpreads(); err != nil {
		return nil, err
	}
	if err := lowering.findCycles(modules); err != nil {
		return nil, err
	}
	readiness(lowering.result)
	borrow(lowering.result)
	counters(lowering.result)
	return lowering.result, nil
}

type lowering struct {
	viewCallablePayloadTypes     map[int]viewCallablePayloadType
	viewCallableProducerPayloads map[int][]viewCallablePayloadType

	program *load.Program
	checker *checker.Checker
	result  *ir.Program

	// cyclicModules keeps unresolved reads checked throughout a cyclic graph.
	cyclicModules     bool
	provenModuleReads map[*ast.Node]bool
	cycleReadFindings []Finding

	// strings indexes result.Strings, so a constant used twice is stored once.
	strings map[string]int

	// locals maps each variable's symbol to its index in result.Locals.
	locals map[*ast.Symbol]int

	// functions maps each function declaration's symbol to its index in result.Functions.
	functions map[*ast.Symbol]int

	// function is the function being lowered, or nil for the module's top level.
	function *ir.Function

	// this is the local this is in a method or constructor, or -1.
	this int

	// functionIndex is the function being lowered, -1 for the module's top level, and closures the
	// closures being lowered, outermost first, each inside the one before.
	functionIndex int
	closures      []int

	// classes maps each module class's symbol to its declaration, and instances each instantiation
	// already lowered (class.go).
	// omittedOptionals is every optional property some object literal typed by its context leaves out:
	// that object has no slot for it (class.go's absentOptionalWrite).
	omittedOptionals map[*ast.Symbol]bool
	classes          map[*ast.Symbol]*ast.Node
	statics          map[*ast.Symbol]*instance
	staticGlobals    map[*ast.Symbol]int
	instances        map[string]*instance
	derivedAncestors map[*ast.Symbol]bool

	// privateOwners numbers each class that declares a private name, in the order first met, to
	// qualify its private members' names (privateName).
	privateOwners map[*ast.Node]int

	// instance is the class instantiation being lowered, if any.
	instance *instance

	// substitution is what each type parameter stands for in the instantiation being lowered.
	substitution map[*checker.Type]ir.Type

	// unlowerable is the first construct found not lowerable somewhere that can't return an error.
	unlowerable error

	// alwaysUndefined are parameters that only ever receive undefined, whatever the checker calls
	// their type: the first of an Array.from callback's (from.go). Each is a reference that's always
	// missing.
	alwaysUndefined map[*ast.Symbol]bool

	// signed are the functions whose signatures are written and whose bodies aren't lowered yet, by
	// index, each with what its body still needs (signature).
	signed map[int]signed

	// forwarders are the globals holding the function values made for module functions read as values,
	// by the function each forwards to, and forwarderValues the declarations that make them, which run
	// before anything else (functionValue).
	forwarders      map[int]int
	forwarderValues []ir.Statement

	// unsetUntil is where, in the constructor being lowered, its last assignment of a field without an
	// initializer ends: before it, this is only for reading and writing fields (useOfThis).
	unsetUntil int

	// generics maps each generic module function's symbol to its declaration, and genericInstances
	// each instantiation already lowered to its function (generic.go). genericDepth counts the
	// instantiations being lowered inside one another.
	generics              map[*ast.Symbol]*ast.Node
	genericInstances      map[string]int
	classGenericInstances map[string]map[string]int
	genericDepth          int

	// caught are the variables a catch binds, each always an Error; tries are the try statements
	// lowered (exceptions.go).
	caught map[*ast.Symbol]bool
	tries  []tryRecord

	// initializing is the variables whose initializers are being lowered, by where each is declared.
	initializing map[int]*ast.Node

	// For the cycle finder (cycles.go): the checker's type of each local, and where it's declared;
	// every function value made, with its type; and the class type being instantiated, for this.
	localTypes     map[int]*checker.Type
	localAlso      map[int][]*checker.Type
	localNodes     map[int]*ast.Node
	closureRecords []closureRecord
	// Untagged callable read targets are certified after all producers are lowered.
	untaggedCallableTargets map[ir.ViewContractID]*checker.Type
	classType               *checker.Type
	classNode               *ast.Node

	// typeMapper is what the type parameters of the instantiation being lowered, and of those it's
	// inside, stand for; instantiated is every class type an instantiation was made for (instantiate.go).
	typeMapper   *typeMapper
	instantiated []*checker.Type

	// writeSites is every write into a slot lowering made, by its IR node's Site less one (fresh.go).
	writeSites       []writeSite
	accessorCaptures []accessorCapture
	accessorNames    map[string]bool
}
