// Package lower turns a checked program into Adamic's IR.
//
// Stage 0 lowers a little and refuses the rest. Every construct it can't lower yet is an error
// naming the construct and where it is, never a skip: a compiler that drops a statement it doesn't
// understand produces a program that looks right and isn't, which is the one thing Adamic must
// never do.
package lower

import (
	"context"
	"errors"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/buildlog"
	"github.com/system-inc/adamic/internal/flow"
	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"time"
)

// Lower lowers a checked program from one entry, in ESM evaluation order.
type Options struct {
	Debug         bool
	ProgramRegion bool
	// OwnershipQuery adds validation after the existing move proof. Default off.
	OwnershipQuery bool
}

func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {
	return LowerWithOptions(ctx, program, Options{
		ProgramRegion:  os.Getenv("ADAMIC_PROGRAM_REGION") == "1",
		OwnershipQuery: os.Getenv("ADAMIC_OWNERSHIP_QUERY") == "1",
	})
}

func LowerWithOptions(ctx context.Context, program *load.Program, options Options) (*ir.Program, error) {
	files := program.Files()
	if len(files) != 1 {
		return nil, fmt.Errorf("lower: stage 0 compiles a program from one entry file, got %d", len(files))
	}
	entry := files[0]
	// The lower phase's line on $ADAMIC_BUILD_LOG, once a program: a program region's discovery pass is in it.
	defer buildlog.Phase("lower", time.Now())
	// Stage 0 checks single-threaded, so one checker answers for every file.
	typeChecker, release := program.Checker(ctx, entry)
	defer release()

	var plan *programMembership
	if options.ProgramRegion {
		discovery, lowerCheckedError := lowerChecked(program, typeChecker, entry, nil, true, false)
		if lowerCheckedError != nil {
			return nil, lowerCheckedError
		}
		plan = discovery.programPlan
	}
	built, lowerCheckedError := lowerChecked(
		program,
		typeChecker,
		entry,
		plan,
		false,
		options.Debug,
		options.OwnershipQuery,
	)
	if lowerCheckedError != nil {
		return nil, lowerCheckedError
	}

	return built.result, nil
}

func lowerChecked(
	program *load.Program,
	typeChecker *checker.Checker,
	entry *ast.SourceFile,
	plan *programMembership,
	discovery, debug bool,
	ownershipQuery ...bool,
) (*lowering, error) {
	return lowerCheckedWithCycleTypes(program, typeChecker, entry, plan, discovery, debug, nil, ownershipQuery...)
}

// Tests retain transient checker types before lazy queries replace them.
func lowerCheckedWithCycleTypes(
	program *load.Program,
	typeChecker *checker.Checker,
	entry *ast.SourceFile,
	plan *programMembership,
	discovery, debug bool,
	cycleTypes map[int]*checker.Type,
	ownershipQuery ...bool,
) (*lowering, error) {
	lowering := &lowering{
		cycleTypes:       cycleTypes,
		program:          program,
		checker:          typeChecker,
		result:           &ir.Program{},
		this:             -1,
		functionIndex:    -1,
		entry:            entry,
		programPlan:      plan,
		programDiscovery: discovery,
	}
	lowering.debug = debug
	lowering.result.ProgramRegion = discovery || plan != nil
	if plan != nil {
		lowering.result.ProgramTypes = plan.types
	}

	// The base name only, so the same program emits the same C on every machine.
	lowering.result.Source = filepath.Base(program.FileName(entry))
	modules, moduleOrderError := lowering.moduleOrder(entry)
	if moduleOrderError != nil {
		return nil, moduleOrderError
	}
	lowering.noteInheritance(modules)
	if enumInitializationError := lowering.enumInitialization(modules); enumInitializationError != nil {
		return nil, enumInitializationError
	}
	lowering.noteAccessorNames(modules)
	lowering.noteOmittedOptionals(modules)
	if namespaceInitializationError := lowering.namespaceInitialization(modules); namespaceInitializationError != nil {
		return nil, namespaceInitializationError
	}
	for _, module := range modules {
		if refuseError := lowering.refuse(module); refuseError != nil {
			return nil, refuseError
		}
	}
	// Link every declaration before lowering any function body, including across back edges.
	var declarations []*ast.Node
	for _, module := range modules {
		declarations = append(declarations, module.Statements.Nodes...)
	}
	if declareModuleError := lowering.declareModule(declarations); declareModuleError != nil {
		return nil, declareModuleError
	}
	lowering.recordExports(entry)
	for _, module := range modules {
		body, moduleStatementsError := lowering.moduleStatements(module)
		if moduleStatementsError != nil {
			return nil, moduleStatementsError
		}
		lowering.result.Main = append(lowering.result.Main, body...)
	}
	if lowering.unlowerable != nil {
		return nil, lowering.unlowerable
	}
	lowering.result.Main = append(lowering.forwarderValues, lowering.result.Main...)
	lowering.finishClassCalls()
	if finishAccessorsError := lowering.finishAccessors(); finishAccessorsError != nil {
		return nil, finishAccessorsError
	}
	lowering.finishToPrimitive()
	// Before exceptions: a view a generator feeds is rewritten into its dispatch here, and the
	// dispatch's steps are throws the fixed point must see, each its own generator's ResumeMayThrow.
	if checkIterableViewsError := lowering.checkIterableViews(); checkIterableViewsError != nil {
		return nil, checkIterableViewsError
	}
	if exceptionsError := lowering.exceptions(); exceptionsError != nil {
		return nil, exceptionsError
	}
	if checkLibrarySpreadsError := lowering.checkLibrarySpreads(modules); checkLibrarySpreadsError != nil {
		return nil, checkLibrarySpreadsError
	}
	if checkAccessorSpreadsError := lowering.checkAccessorSpreads(); checkAccessorSpreadsError != nil {
		return nil, checkAccessorSpreadsError
	}
	if checkRecordGuardedReadsError := lowering.checkRecordGuardedReads(); checkRecordGuardedReadsError != nil {
		return nil, checkRecordGuardedReadsError
	}
	if checkDynamicKeyedGettersError := lowering.checkDynamicKeyedGetters(); checkDynamicKeyedGettersError != nil {
		return nil, checkDynamicKeyedGettersError
	}
	if checkGeneratorRoadsError := checkGeneratorRoads(lowering.result); checkGeneratorRoadsError != nil {
		return nil, checkGeneratorRoadsError
	}
	if normalizeAsyncError := flow.NormalizeAsync(lowering.result); normalizeAsyncError != nil {
		where := entry.AsNode()
		var located *flow.AsyncError
		if errors.As(normalizeAsyncError, &located) && located.Site > 0 && located.Site <= len(lowering.asyncSites) {
			where = lowering.asyncSites[located.Site-1]
		}
		return nil, &NotYet{Where: lowering.program.Where(where), What: normalizeAsyncError.Error()}
	}
	if lowering.result.AsyncEntry > 0 {
		// NormalizeAsync moved module bodies into one scheduler entry. Main now
		// holds only its generated invocation; retain boundaries of the actual Main.
		source := ir.SourceIdentity{
			Module:      lowering.modulePath(entry),
			Declaration: []string{"module"},
			Role:        "async_module",
		}
		lowering.result.Functions[lowering.result.AsyncEntry-1].Source = source
		lowering.result.MainModules = []ir.MainModule{{Module: source.Module, Statements: len(lowering.result.Main)}}
	}
	if findCyclesError := lowering.findCycles(modules); findCyclesError != nil {
		return nil, findCyclesError
	}
	if checkViewOptionalWritesError := lowering.checkViewOptionalWrites(); checkViewOptionalWritesError != nil {
		return nil, checkViewOptionalWritesError
	}
	if checkViewCopiesError := lowering.checkViewCopies(); checkViewCopiesError != nil {
		return nil, checkViewCopiesError
	}
	if checkViewWidenedReadsError := lowering.checkViewWidenedReads(); checkViewWidenedReadsError != nil {
		return nil, checkViewWidenedReadsError
	}
	if len(lowering.asyncPayloadFailures) > 0 {
		return nil, lowering.asyncPayloadFailures[0]
	}
	if checkPromiseConsumptionError := lowering.checkPromiseConsumption(); checkPromiseConsumptionError != nil {
		return nil, checkPromiseConsumptionError
	}
	if checkViewLogicalWritesError := lowering.checkViewLogicalWrites(); checkViewLogicalWritesError != nil {
		return nil, checkViewLogicalWritesError
	}
	if discovery || plan != nil {
		if lowering.result.HasAsync() {
			return nil, &Refused{
				Where: program.Where(entry.AsNode()),
				What:  "Program region with async work",
				Fix:   "keep Program members in the one-shot synchronous CLI",
			}
		}
		// A program with generators may have a Program region (#8xcz03h), since no generator frame
		// can join one or reach a member by a road the cycle walk doesn't see:
		//  1. A frame has no captured cell: checkGeneratorBody stops a captured local and a nested
		//     function in a generator, and FrameEnvironment holds only captured locals
		//     (TestProgramRegionGeneratorCapturedLocalStops, TestProgramRegionGeneratorNestedFunctionStops).
		//  2. A generator's call is an ir.Call, which programAllocation never stamps, and
		//     internal/native/generator.go adopts nothing, so no frame is adopted
		//     (internal/oracle's TestProgramRegionGeneratorsNotAdopted). A nested generator's call, or
		//     one written in place, is an ir.CallClosure through its carrier; a written-in-place one's
		//     carrier is the MakeClosure nestedGeneratorCall makes, never through programAllocation,
		//     and its function's type, a generator function's, which no slot holds, joins no
		//     ProgramTypes, so that carrier stays counted too (Memory on #qk8rztp S3b,
		//     TestProgramRegionGeneratorInPlaceNotAdopted). Its function would still join through a
		//     cell it captures, were that cell a member, which only a cycle through the cell makes it
		//     (selectProgramMembership's environment rule); no pin reaches one yet.
		//  3. No generator value reaches a slot: expression.go stops every value use but an Iterable<T>
		//     argument, and checkGeneratorRoads holds the lowered program to that
		//     (TestProgramRegionGeneratorStoredInFieldStops, TestProgramRegionGeneratorPushedIntoArrayStops,
		//     TestProgramRegionGeneratorHeldInLocalStops), and an Iterable<T> parameter one is
		//     passed into stays one for-of's alone (TestProgramRegionGeneratorParameterStoredInFieldStops,
		//     TestProgramRegionGeneratorParameterCapturedByClosureStops).
		// When a generator can be stored in a slot (#qk8rztp S2c-4), the cycle walk needs the frame's
		// links, and these pins are that work's done-when.
	}
	if plan != nil {
		if applyProgramMetadataError := lowering.applyProgramMetadata(plan); applyProgramMetadataError != nil {
			return nil, applyProgramMetadataError
		}
	}
	if len(ownershipQuery) > 0 && ownershipQuery[0] && lowering.hasMovedParallelMap {
		for _, transfer := range fresh.QueryOwnershipTransfers(lowering.result) {
			if transfer.Answer.Verdict != fresh.Proven {
				where := lowering.moveSites[transfer.Site-1]
				return nil, lowering.moveRefused(where, "whole reachable ownership is not proven (ownership query "+
					transfer.Answer.Verdict.String()+
					": "+
					fmt.Sprint(transfer.Answer.Evidence)+
					")")
			}
		}
	}
	readiness(lowering.result)
	borrow(lowering.result)
	counters(lowering.result)
	// Last, once nothing more changes the program: its walk of the method tables is the one the
	// emitter reads too (ir.Program's MethodsNamed).
	if voidDiscardsError := lowering.voidDiscards(); voidDiscardsError != nil {
		return nil, voidDiscardsError
	}
	return lowering, nil
}

type lowering struct {
	debug      bool
	cycleTypes map[int]*checker.Type // Optional differential-test capture; nil in normal builds.

	// Source-only sites for graph adoption diagnostics; never stored in emitted IR.
	graphViewSite, graphDynamicSite *ast.Node
	// dynamicKeyedSite is the first run-time key read (reflectGet or viewKeyedRead), where
	// checkDynamicKeyedGetters says why a class getter keeps the program from reading by a run-time key.
	dynamicKeyedSite *ast.Node
	// getterDeclarations is each class getter's declaration, by its function (registerAccessor), which
	// checkDynamicKeyedGetters reads.
	getterDeclarations map[int]*ast.Node
	// functionNodes is the declaration each function's body was lowered from (lowerBody), by the
	// function's index, so a judgment about what lies inside a declaration reads the source's own
	// nesting (nestedGeneratorWrites). A helper lowering makes with no declaration isn't here.
	functionNodes map[int]*ast.Node
	// slots are what a value is judged against where the checker's contextual type isn't the whole
	// answer (slotType, impliedTarget).
	slots slotTargets
	// voidHolders memoizes holdsVoid: whether a type holds void at a position (#p7t8tmh).
	voidHolders map[*checker.Type]bool
	// constEnumReach is every const enum the program reaches as a value (reachedConstEnums), by its
	// symbol, found once.
	constEnumReach map[*ast.Symbol]bool
	// voidBound is, while a generic body instantiated at void is judged, each type parameter bound to
	// void (true) or to a type holding void (false), and genericVoidVerdicts what each judgment found.
	voidBound           map[*checker.Type]bool
	genericVoidVerdicts map[string]error
	// voidMessageMethods are the methods whose bodies voidMethodReads is judging for a message, so a
	// method that reaches itself through this stops there instead of judging itself again.
	voidMessageMethods map[*ast.Node]bool
	// A per-lowering query seam lets tests count actual checker calls without globals.
	// Nil uses checker.GetTypeAtLocation.
	impliedTargetQuery func(*ast.Node) *checker.Type

	librarySpreads      []librarySpread
	wholeObjectCopies   []wholeObjectCopy
	parallelMoveDemand  map[parallelMoveSite]bool
	hasMovedParallelMap bool
	moveSites           []*ast.Node

	// viewSources is each view's source and target, and viewTypes the checker types a read's view
	// ids name, its receiver's and its field's, so the passes after lowering can judge and name them
	// (view_widened.go).
	viewSources []viewSource
	viewTypes   map[int]*checker.Type
	// viewOptionalWrites are the writes to a field their receiver declares optional, judged once every
	// view is in (view_optional_writes.go).
	viewOptionalWrites []viewOptionalWrite
	// narrowedReads are, for each local, the types the checker narrowed a read of it to, and
	// narrowedFields the types it narrowed a field read to: where a predicate's view takes its
	// value (view_widened.go).
	narrowedReads  map[int]map[int]*checker.Type
	narrowedFields []*checker.Type
	// iterableViews are the for-ofs over an Iterable<T> or IterableIterator<T> slot lowered as a
	// collection's iterator, each proven by checkIterableViews once every call is in
	// (iterable_views.go).
	iterableViews []iterableView
	// handNext is the callee of the hand-driven `x.next()` callClosure is lowering, which property
	// lets past its library-member guard (handDrivenNext); nil otherwise.
	handNext *ast.Node
	// generatorArguments are the generators' calls handed into an Iterable<T> parameter, by
	// IterableOf's Site (iterableOfGenerator), which checkIterableViews proves.
	generatorArguments []generatorArgument
	// inPlaceGenerators is the function each generator written in place lowered to, by its expression
	// and the function it's written in (inPlaceGenerator). It is this lowering's, never global: the
	// latent census and a relower each build a fresh lowering.
	inPlaceGenerators map[inPlaceGeneratorKey]int

	programPlan          *programMembership
	programDiscovery     bool
	programCandidates    map[*checker.Type]*ast.Node
	entry                *ast.SourceFile
	asyncSites           []*ast.Node
	asyncPayloadFailures []error

	// V2 retains exact logical producer types for union member certification.
	untaggedCallableTargets map[ir.ViewContractID]*checker.Type
	// objectIntersections are the intersections objectIntersection is proving, so a field that
	// holds one of them again ends there.
	objectIntersections map[*checker.Type]bool
	// elementFits is elementsHeldOtherwise's walk answered once per value type and slot type, so
	// judging every value handed to a slot costs each pair once (element_fit.go).
	elementFits map[elementFitKey]*elementMismatch
	// intersectionViewDepth is how many intersections' view contracts are being built, each inside
	// a field of the one before (view_lazy.go's intersectionViewDepthBound).
	intersectionViewDepth int

	program          *load.Program
	checker          *checker.Checker
	result           *ir.Program
	sourceIdentities map[*ast.Node]ir.SourceIdentity
	sourceModules    map[*ast.SourceFile]string
	// typeStrings is each type's TypeToString, held for the lowering (typeString).
	typeStrings map[*checker.Type]string
	// predicateProofs is every type predicate's proof, run once per verifier mode (predicates_proof.go).
	predicateProofs *predicateProofs

	// cyclicModules keeps unresolved reads checked throughout a cyclic graph.
	cyclicModules     bool
	provenModuleReads map[*ast.Node]bool
	// moduleOrdered is moduleOrder's held answer, and moduleOrders how many times it was computed.
	moduleOrdered *moduleOrdering
	moduleOrders  int
	// assignedSymbols is every symbol an identifier in the program assigns to (predicates.go's assignedSymbol).
	assignedSymbols map[*ast.Symbol]bool
	// writtenAccessorNames is every name a getter or setter in the program has, "" for a computed one
	// (enum_tag_flow.go's accessorName). accessorNames below holds getters alone.
	writtenAccessorNames map[string]bool
	// crossFunctionWrites is every binding a function other than its own writes (narrowed_nullish.go).
	crossFunctionWrites map[*ast.Symbol]bool
	// propertyWrites is every property write in the program, by what it can reach (narrowed_nullish.go).
	propertyWrites *propertyWrites
	// topLevelOnlyFunctions is calledOnlyFromTopLevel's answer for each function declaration asked
	// about (narrowed_nullish.go).
	topLevelOnlyFunctions map[*ast.Node]bool

	// predicateFalse is the false-direction walk's switch and memo (predicate_false_flow.go).
	predicateFalse *predicateFalseFlow

	// refusal is what refuse's per-node rules leave for one another (refusal_rules.go).
	refusal refusalPass

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

	// nodeHostViews remembers each value type and view nodeHostViewAt has judged, and
	// nodeHostHolders which types may hold a host object (library_node.go).
	nodeHostViews   map[nodeHostViewKey]string
	nodeHostHolders map[*checker.Type]bool

	// For the cycle finder (cycles.go): the checker's type of each local, and where it's declared;
	// every function value made, with its type; and the class type being instantiated, for this.
	localTypes     map[int]*checker.Type
	localAlso      map[int][]*checker.Type
	localNodes     map[int]*ast.Node
	closureRecords []closureRecord
	classType      *checker.Type
	classNode      *ast.Node // The name identifier, declaration.Name(); Parent is the class declaration.

	// entriesRecords is each object literal's entriesRecordLiteral answer, and entriesProgramRecords, once
	// entriesProgramRecordsKnown, entriesProgramHasRecords' (entries_records.go).
	entriesRecords             map[*ast.Node]bool
	entriesProgramRecords      bool
	entriesProgramRecordsKnown bool
	// entriesRecordTypes, once entriesRecordTypesKnown, are the types the program's records are made
	// as, and recordGuardedReads the union and maybe field reads admitted beside them, each checked
	// again once every view is registered (entries_records.go).
	entriesRecordTypes      []*checker.Type
	entriesRecordTypesKnown bool
	recordGuardedReads      []recordGuardedRead
	// mergedClassTypes, once mergedClassTypesKnown, are the instance types of the program's classes
	// merged with an interface, and plainAsked each receiver id plainReceiver has judged
	// (entries_records.go, #9jnjwrw).
	mergedClassTypes      []*checker.Type
	mergedClassTypesKnown bool
	plainAsked            map[int]bool

	// typeMapper is what the type parameters of the instantiation being lowered, and of those it's
	// inside, stand for; instantiated is every class type an instantiation was made for (instantiate.go).
	typeMapper   *typeMapper
	instantiated []*checker.Type

	// declaredKeys is set while recordDeclaredKeyRefusal judges an .a module's views, where widened
	// holds a declared property invariant against an index signature (records_declared_keys.go).
	declaredKeys bool

	// droppedContractsRead is set while recordStorageView judges a .ts module's view, where a record seen
	// as a record may drop a named contract the reads hold (droppedContractRead, records_named.go).
	droppedContractsRead bool

	// writeSites is every write into a slot lowering made, by its IR node's Site less one (fresh.go).
	writeSites       []writeSite
	accessorCaptures []accessorCapture
	accessorNames    map[string]bool
	// contractGetters is each view's getters of a contract field's name, judged by reach in
	// finishAccessors (class_accessors.go).
	contractGetters []contractGetters

	// toPrimitive is the program's to_string_hint helper plus one, or zero before a String of a union
	// that may hold an object needs it (to_primitive.go).
	toPrimitive int
}
