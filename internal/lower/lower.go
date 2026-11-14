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
	// ModuleFormat is the module format the build declares its program in. Default ModuleFormatEcmaScript.
	ModuleFormat ModuleFormat
}

// ModuleFormat is a module format a build can declare: what Node would run the program as, and so
// which of Node's module values it has. It's a build input, so internal/lowered's key holds it with
// every other Options field.
type ModuleFormat string

const (
	// ModuleFormatEcmaScript is an ES module, the format every Adamic program is unless its build says
	// otherwise. It's the zero value, so Options that name no format lower as they always have, and
	// one format is never spelled two ways.
	ModuleFormatEcmaScript ModuleFormat = ""
	// ModuleFormatCommonJSBundle is one bundled image of CommonJS modules, as tsc ships as _tsc.js,
	// where every module's __filename and __dirname are the running image's (Adamic's ruling on
	// #0ta3dk9; nodeCommonJSModuleValuesBound). It isn't a module a file each, whose two Node gives
	// each module of its own.
	ModuleFormatCommonJSBundle ModuleFormat = "CommonJSBundle"
)

func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {
	return LowerWithOptions(ctx, program, Options{ProgramRegion: os.Getenv("ADAMIC_PROGRAM_REGION") == "1", OwnershipQuery: os.Getenv("ADAMIC_OWNERSHIP_QUERY") == "1"})
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
		discovery, err := lowerCheckedWithCycleTypes(program, typeChecker, entry, nil, true, false, options.ModuleFormat, nil)
		if err != nil {
			return nil, err
		}
		plan = discovery.programPlan
	}
	built, err := lowerCheckedWithCycleTypes(program, typeChecker, entry, plan, false, options.Debug, options.ModuleFormat, nil, options.OwnershipQuery)
	if err != nil {
		return nil, err
	}

	return built.result, nil
}

// lowerChecked lowers a program as an ES module, the format every caller but LowerWithOptions lowers in.
func lowerChecked(program *load.Program, typeChecker *checker.Checker, entry *ast.SourceFile, plan *programMembership, discovery, debug bool, ownershipQuery ...bool) (*lowering, error) {
	return lowerCheckedWithCycleTypes(program, typeChecker, entry, plan, discovery, debug, ModuleFormatEcmaScript, nil, ownershipQuery...)
}

// Tests retain transient checker types before lazy queries replace them. A checked view's
// element reads are keyed by what the program views (viewElementKey). An array contract the key's
// pre-pass didn't see would leave a read it should check raw, so the key fails closed: the program
// is lowered again with every element read checked.
func lowerCheckedWithCycleTypes(program *load.Program, typeChecker *checker.Checker, entry *ast.SourceFile, plan *programMembership, discovery, debug bool, moduleFormat ModuleFormat, cycleTypes map[int]*checker.Type, ownershipQuery ...bool) (*lowering, error) {
	key := &viewElementKey{}
	built, err := lowerWithElementKey(key, program, typeChecker, entry, plan, discovery, debug, moduleFormat, cycleTypes, ownershipQuery...)
	if len(key.misses) != 0 {
		built, err = lowerWithElementKey(&viewElementKey{every: true}, program, typeChecker, entry, plan, discovery, debug, moduleFormat, cycleTypes, ownershipQuery...)
	}
	if err != nil {
		return built, err
	}
	// A relation the checker gave up on answered "not related", and lowering may have read that answer
	// as a fact, so a program whose checker holds one is NotYet (#apyq5ve). Checked here, after the
	// lowering, so every caller gets it; a refusal or a NotYet already fails closed.
	if overflow := program.RelationOverflow(typeChecker); overflow != "" {
		return nil, &NotYet{Where: built.result.Source, What: "a type relation the checker gave up on as too complex, whose answer lowering read (" + overflow + ")"}
	}
	return built, nil
}

func lowerWithElementKey(key *viewElementKey, program *load.Program, typeChecker *checker.Checker, entry *ast.SourceFile, plan *programMembership, discovery, debug bool, moduleFormat ModuleFormat, cycleTypes map[int]*checker.Type, ownershipQuery ...bool) (built *lowering, err error) {
	lowering := &lowering{cycleTypes: cycleTypes, program: program, checker: typeChecker, result: &ir.Program{}, this: -1, functionIndex: -1, entry: entry, programPlan: plan, programDiscovery: discovery, elementKey: key}
	// A stop before checkHostMemberWrites judges a deferred host member write is main's stop at that write.
	defer func() { err = lowering.hostMemberPending(err) }()
	lowering.debug = debug
	lowering.moduleFormat = moduleFormat
	lowering.result.ProgramRegion = discovery || plan != nil
	if plan != nil {
		lowering.result.ProgramTypes = plan.types
	}

	// The base name only, so the same program emits the same C on every machine.
	lowering.result.Source = filepath.Base(program.FileName(entry))
	modules, err := lowering.moduleOrder(entry)
	if err != nil {
		return nil, err
	}
	// Before any view is made or any element read lowers, so every read is keyed by the whole program.
	if err := lowering.collectViewedElements(modules); err != nil {
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
	lowering.recordExports(entry)
	for _, module := range modules {
		body, err := lowering.moduleStatements(module)
		if err != nil {
			return nil, err
		}
		lowering.result.Main = append(lowering.result.Main, body...)
	}
	if lowering.unlowerable != nil {
		return nil, lowering.unlowerable
	}
	lowering.checkOverloadResults(modules)
	lowering.result.Main = append(lowering.forwarderValues, lowering.result.Main...)
	lowering.finishClassCalls()
	// Before the accessors' dispatchers: a read this makes checked can't go through a dispatcher's
	// unchecked fallback, which finishAccessors stops by asking CheckedField.
	if err := lowering.certifySourceSlots(); err != nil {
		return nil, err
	}
	if err := lowering.finishAccessors(); err != nil {
		return nil, err
	}
	lowering.finishToPrimitive()
	// Before exceptions: a view a generator feeds is rewritten into its dispatch here, and the
	// dispatch's steps are throws the fixed point must see, each its own generator's ResumeMayThrow.
	if err := lowering.checkIterableViews(); err != nil {
		return nil, err
	}
	if err := lowering.exceptions(); err != nil {
		return nil, err
	}
	if err := lowering.checkLibrarySpreads(modules); err != nil {
		return nil, err
	}
	if err := lowering.checkAccessorSpreads(); err != nil {
		return nil, err
	}
	if err := lowering.checkRecordGuardedReads(); err != nil {
		return nil, err
	}
	if err := lowering.checkDynamicKeyedGetters(); err != nil {
		return nil, err
	}
	if err := checkGeneratorRoads(lowering.result); err != nil {
		return nil, err
	}
	if err := flow.NormalizeAsync(lowering.result); err != nil {
		where := entry.AsNode()
		var located *flow.AsyncError
		if errors.As(err, &located) && located.Site > 0 && located.Site <= len(lowering.asyncSites) {
			where = lowering.asyncSites[located.Site-1]
		}
		return nil, &NotYet{Where: lowering.program.Where(where), What: err.Error()}
	}
	if lowering.result.AsyncEntry > 0 {
		// NormalizeAsync moved module bodies into one scheduler entry. Main now
		// holds only its generated invocation; retain boundaries of the actual Main.
		source := ir.SourceIdentity{Module: lowering.modulePath(entry), Declaration: []string{"module"}, Role: "async_module"}
		lowering.result.Functions[lowering.result.AsyncEntry-1].Source = source
		lowering.result.MainModules = []ir.MainModule{{Module: source.Module, Statements: len(lowering.result.Main)}}
	}
	if err := lowering.findCycles(modules); err != nil {
		return nil, err
	}
	if err := lowering.checkBoundArguments(); err != nil {
		return nil, err
	}
	if err := lowering.checkViewOptionalWrites(); err != nil {
		return nil, err
	}
	// After findCycles, whose lazy view rule keeps a checked view's read of an optional callable its own
	// stop, and after finishAccessors, whose dispatcher a presence test mustn't have become.
	if err := lowering.checkOptionalMemberPresence(); err != nil {
		return nil, err
	}
	if err := lowering.checkViewCopies(); err != nil {
		return nil, err
	}
	if err := lowering.checkViewWidenedReads(); err != nil {
		return nil, err
	}
	if err := lowering.checkHostMemberWrites(); err != nil {
		return nil, err
	}
	if len(lowering.asyncPayloadFailures) > 0 {
		return nil, lowering.asyncPayloadFailures[0]
	}
	if err := lowering.checkPromiseConsumption(); err != nil {
		return nil, err
	}
	if err := lowering.checkViewLogicalWrites(); err != nil {
		return nil, err
	}
	if discovery || plan != nil {
		if lowering.result.HasAsync() {
			return nil, &Refused{Where: program.Where(entry.AsNode()), What: "Program region with async work", Fix: "keep Program members in the one-shot synchronous CLI"}
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
		if err := lowering.applyProgramMetadata(plan); err != nil {
			return nil, err
		}
	}
	if len(ownershipQuery) > 0 && ownershipQuery[0] && lowering.hasMovedParallelMap {
		for _, transfer := range fresh.QueryOwnershipTransfers(lowering.result) {
			if transfer.Answer.Verdict != fresh.Proven {
				where := lowering.moveSites[transfer.Site-1]
				return nil, lowering.moveRefused(where, "whole reachable ownership is not proven (ownership query "+transfer.Answer.Verdict.String()+": "+fmt.Sprint(transfer.Answer.Evidence)+")")
			}
		}
	}
	readiness(lowering.result)
	borrow(lowering.result)
	counters(lowering.result)
	// Last, once nothing more changes the program: its walk of the method tables is the one the
	// emitter reads too (ir.Program's MethodsNamed).
	if err := lowering.voidDiscards(); err != nil {
		return nil, err
	}
	if err := lowering.checkInspectorDoomedContracts(); err != nil {
		return nil, err
	}
	// After the last pass: no function reference is added past here, so every function a flow graph
	// took to never run must still be named by nothing (view_flow.go's runningFunctions).
	if err := checkFlowLiveness(lowering.result, lowering.flowDead); err != nil {
		return nil, err
	}
	return lowering, nil
}

type lowering struct {
	inspectorRequireKnown     bool // Lazy, immutable program facts; no IR or unit ownership.
	inspectorRequirePresent   bool
	inspectorSkippedContracts map[*ast.Node]int // Skipped AST reads -> unique Panic slots, owned by this IR snapshot.
	bindGetters               map[string]bool   // source getter names, memoized for bind receiver admission
	debug                     bool
	cycleTypes                map[int]*checker.Type // Optional differential-test capture; nil in normal builds.
	// moduleFormat is the module format the build declares (Options.ModuleFormat). A lowering a test
	// builds by hand leaves it ModuleFormatEcmaScript.
	moduleFormat ModuleFormat

	// flowDead is every function a flow graph built during lowering took to never run (flowGraph,
	// iterableViewFlowGraph), which checkFlowLiveness holds to that once lowering is done.
	flowDead map[int]bool

	// Source-only sites for graph adoption diagnostics; never stored in emitted IR.
	graphDynamicSite *ast.Node
	// graphDecodeSite is the first decodeJson whose value can hold a container, and graphDecodeType
	// the type it decodes, where graphAdoptionBoundary names a decode a graph program can't hold yet.
	graphDecodeSite *ast.Node
	graphDecodeType string
	// graphUnplanned are allocations the graph plan can't hold whole, over elements a graph program could
	// close a cycle through, each with what it is; graphAdoptionBoundary stops the first in a graph
	// program. Only a record's entries is recorded now (unplannedAllocation says why).
	graphUnplanned []graphUnplannedSite
	// graphPairs are the Map entries sites, each with its pair's tuple type, which graphAdoptionBoundary
	// reads when a graph program holds a pair as some other type its plan can't name (#dps0kp6).
	graphPairs []graphPairSite
	// graphNodeArrays are the NodeArray literals, each with its NodeArray type, which
	// graphAdoptionBoundary reads: native construction never adopts one (#djwd1t2, row 3). A site is
	// recorded before its literal lowers, so one whose literal then fails, or that a rolled-back
	// speculative lowering made, stays listed; that can only over-stop, which fails closed.
	graphNodeArrays []graphNodeArraySite
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
	// classInitializers holds the expression and block roots actually lowered for each non-generic
	// class, including one nested in a generic function or method, which is recorded once by node though
	// it lowers per instance. Unconstructed instance fields and skipped initializers never enter this
	// audit set.
	classInitializers map[*ast.Node]map[*ast.Node]bool
	// slots are what a value is judged against where the checker's contextual type isn't the whole
	// answer (slotType, impliedTarget).
	slots slotTargets
	// voidHolders memoizes holdsVoid: whether a type holds void at a position (#p7t8tmh).
	voidHolders map[*checker.Type]bool
	// constEnumReach is every const enum the program reaches as a value (reachedConstEnums), by its
	// symbol, found once.
	constEnumReach map[*ast.Symbol]bool
	// namespaceValues is every program namespace whose name the program reads as a value, by its
	// symbol, found once before the refusal pass (noteNamespaceValues), and namespaceContainers the
	// global holding each one's object (namespace_value.go).
	namespaceValues     map[*ast.Symbol]bool
	namespaceContainers map[*ast.Symbol]int
	// containerFits is every pair container_fit.go's walk judged whole, at no depth bound and leaning on
	// no pair still in progress, with its stop or "" (containerPairGoes, #zrnkpdz).
	containerFits map[containerFitPair]string
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
	jsonReferenceReads  []jsonReferenceRead
	parallelMoveDemand  map[parallelMoveSite]bool
	hasMovedParallelMap bool
	moveSites           []*ast.Node

	// viewSources is each view's source and target, and viewTypes the checker types a read's view
	// ids name, its receiver's and its field's, so the passes after lowering can judge and name them
	// (view_widened.go).
	viewSources []viewSource
	viewTypes   map[int]*checker.Type
	// viewSites is every property access and named getter of the program's modules, in view's preorder,
	// built once and shared by every lowering a latent run makes (view_sites.go, #q1hyhzc).
	// checkedFieldWrites counts, by field name, the CheckedFields keys markCheckedField added;
	// viewSiteSeen is a name's count when view last judged its accesses, and viewSiteTrue the accesses
	// it found checked, which stay checked since keys only grow.
	viewSites          *viewSiteIndex
	checkedFieldWrites map[string]int
	viewSiteSeen       map[string]int
	viewSiteTrue       map[*ast.Node]bool
	// overloadCheckCalls holds every overload result check function (overloadResultCheck), whose call is
	// where its value enters a view (checkViewWidenedReads).
	overloadCheckCalls map[int]bool
	// narrowOverloads memoizes narrowOverload; overloadCalls holds each call to a narrow overload the
	// lowering lowered, true once overloadResultCheck settled it, which checkOverloadResults reads after
	// lowering (overload_result_check.go, #z5d0jye).
	narrowOverloads map[*ast.Node]bool
	overloadCalls   map[*ast.Node]bool
	// viewOptionalWrites are the writes to a field their receiver declares optional, judged once every
	// view is in (view_optional_writes.go).
	viewOptionalWrites []viewOptionalWrite
	// optionalPresenceReads are the presence tests of an optional member lowering read off its slot,
	// judged by what their receivers may hold once the program is lowered (optional_member_presence.go).
	optionalPresenceReads []optionalPresenceRead
	// dispatchedPresenceReads are the presence reads finishAccessors made a dispatcher's call, each with the
	// key the read had before the dispatcher's fallback zeroed it and what its receiver may hold, read
	// there from the reach and owners the dispatch was decided by (optional_member_presence.go). It holds
	// plain values, never the graph, so the latent census copies it as it copies any lowering field.
	dispatchedPresenceReads []dispatchedPresenceRead
	// hostMemberWrites are the writes to a method a program's interface or type literal declares, each
	// with main's stop for it, which lowering stores as fields and checkHostMemberWrites judges once
	// lowering is done (host_member_write.go).
	hostMemberWrites []hostMemberWrite
	// hostMemberJudged says checkHostMemberWrites has run, so a stop after it is the program's own
	// (hostMemberPending).
	hostMemberJudged bool
	// sourceSlotWrites are the writes of an object field through a checked view, each with the type
	// it writes, which certifySourceSlots holds every read of the field's name to once lowering is
	// done, and sourceSlotsJudged says it has, so a write found after that keeps its stop
	// (source_slot_certificate.go, #2msw6t8).
	sourceSlotWrites  []sourceSlotWrite
	sourceSlotsJudged bool
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
	// nestedCallee is the bare callee of the call callClosure is lowering, which nestedReference lets
	// past its overload-set stop, since that call resolves to the overload and is checked; nil
	// otherwise (#z5d0jye).
	nestedCallee *ast.Node

	// optionalReceiverCall identifies the call whose receiver optionalIntrinsic guards.
	optionalReceiverCall *ast.Node
	// optionalReceiverBinding is scoped to the builtin of that call, never carried.
	optionalReceiverBinding optionalReceiverBinding
	// generatorArguments are the generators' calls handed into an Iterable<T> parameter, by
	// IterableOf's Site (iterableOfGenerator), which checkIterableViews proves.
	generatorArguments []generatorArgument
	// inPlaceGenerators is the function each generator written in place lowered to, by its expression
	// and the function it's written in (inPlaceGenerator). It is this lowering's, never global: the
	// latent census and a relower each build a fresh lowering.
	inPlaceGenerators map[inPlaceGeneratorKey]int
	// storedGenerators is the generator function whose call each stored generator's local holds, by
	// the local (storedGenerator, #qk8rztp S2c-4-1), so its one read is lowered as a move.
	storedGenerators map[int]int

	programPlan          *programMembership
	programDiscovery     bool
	programCandidates    map[*checker.Type]*ast.Node
	entry                *ast.SourceFile
	asyncSites           []*ast.Node
	stringSites          []*ast.Node
	asyncPayloadFailures []error

	// V2 retains exact logical producer types for union member certification.
	untaggedCallableTargets map[ir.ViewContractID]*checker.Type
	// elementKey is which element reads a checked view's element check is built for (view_arrays.go).
	// Nil, in a lowering a test builds by hand, checks every read.
	elementKey *viewElementKey
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
	// numericEnums is numericEnum's answer per enum symbol, and numericEnumWalks how many times it was computed.
	numericEnums     map[*ast.Symbol]bool
	numericEnumWalks int
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
	// omittedOptionals names every optional property some producer leaves out, by name: a contextually
	// typed object literal, or any type a decodeJson schema reaches. That object has no slot for it, and
	// a view of it through any type naming the field reaches the same object (class.go's
	// absentOptionalWrite).
	omittedOptionals map[string]bool
	classes          map[*ast.Symbol]*ast.Node
	statics          map[*ast.Symbol]*instance
	staticGlobals    map[*ast.Symbol]int
	instances        map[string]*instance
	derivedAncestors map[*ast.Symbol]bool
	optionalClasses  *optionalClassHierarchy

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

	// neverRest is the identity-marker proof's reference index and answers, and each proven function's
	// marker, made the first time it's asked for (never_rest_marker.go).
	neverRest *neverRestMarks

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

	// hostBindingViewTypes, once hostBindingViewsKnown, is each Provided host binding's declared type's
	// symbol by its model export's type, which nodeHostObject judges it as (host_bindings.go).
	hostBindingViewTypes  map[*ast.Symbol]*checker.Type
	hostBindingViewsKnown bool

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
	// entriesRecordTypes, once entriesRecordTypesKnown, are the types the program's records are held
	// as, and recordGuardedReads the union and maybe field reads admitted beside them, each checked
	// again once every view is registered (entries_records.go). emptyObject, once emptyObjectKnown, is
	// the type `{}`, every record's implicit source (recordEmptyObject).
	entriesRecordTypes      []*checker.Type
	entriesRecordTypesKnown bool
	recordGuardedReads      []recordGuardedRead
	emptyObject             *checker.Type
	emptyObjectKnown        bool
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

	// spreadSources is the type each object literal's spread sees its source as, by the literal's
	// SpreadSite less one, and classHolders each class's instance type by its one-based index, the
	// two sides of the assignability spreadGetters asks (class_accessors.go).
	spreadSources []*checker.Type
	classHolders  map[int]*checker.Type

	// toPrimitive is the program's to_string_hint helper plus one, or zero before a String of a union
	// that may hold an object needs it (to_primitive.go).
	toPrimitive int
}
