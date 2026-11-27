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
	// FirstLinkTraps lowers each unit that stops as a trap, which ends the program when it runs, instead of
	// stopping the lowering, so one run of the first link names every wall (first_link_traps.go). It's
	// measurement only: a program lowered so is never a product, and cmd/adamic-firstlink is its one reader.
	// environmentOptions never sets it. Default off.
	FirstLinkTraps bool
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
	return LowerWithOptions(ctx, program, environmentOptions())
}

// environmentOptions are the Options a build reads from its environment, the same for a program and a Worker.
func environmentOptions() Options {
	return Options{ProgramRegion: os.Getenv("ADAMIC_PROGRAM_REGION") == "1", OwnershipQuery: os.Getenv("ADAMIC_OWNERSHIP_QUERY") == "1"}
}

func LowerWithOptions(ctx context.Context, program *load.Program, options Options) (*ir.Program, error) {
	return lowerTarget(ctx, program, options, false, nil)
}

// LowerWorker uses the environment symbol already proved by CheckSignature. It
// never discovers a handle or makes a second interpretation of that contract.
func LowerWorker(ctx context.Context, program *load.Program, environment *ast.Symbol) (*ir.Program, error) {
	return lowerTarget(ctx, program, environmentOptions(), true, environment)
}
func lowerTarget(ctx context.Context, program *load.Program, options Options, worker bool, environment *ast.Symbol) (*ir.Program, error) {
	files := program.Files()
	if len(files) != 1 {
		return nil, fmt.Errorf("lower: stage 0 compiles a program from one entry file, got %d", len(files))
	}
	entry := files[0]
	if !worker {
		if site := program.WorkersImport(); site != nil {
			return nil, &Refused{Where: program.Where(site), What: "an import of adamic/workers outside a Worker", Fix: "compile it with adamic worker"}
		}
	}

	// The lower phase's line on $ADAMIC_BUILD_LOG, once a program: a program region's discovery pass is in it.
	defer buildlog.Phase("lower", time.Now())
	// Stage 0 checks single-threaded, so one checker answers for every file.
	typeChecker, release := program.Checker(ctx, entry)
	defer release()
	if worker {
		// One call, never an else-if: stage3/census/scan-all's overlay marks every `if err := ...` in this file
		// as a pipeline phase by inserting a line before it, which an else-if can't take.
		var environmentType *checker.Type
		if environment != nil {
			environmentType = typeChecker.GetDeclaredTypeOfSymbol(environment)
		}
		if err := checkWorkerBindings(program, typeChecker, environmentType, nil); err != nil {
			return nil, err
		}
	}

	var plan *programMembership
	if options.ProgramRegion {
		discovery, err := lowerCheckedWithCycleTypes(program, typeChecker, entry, nil, true, false, options.ModuleFormat, options.FirstLinkTraps, nil)
		if err != nil {
			return nil, err
		}
		plan = discovery.programPlan
	}
	built, err := lowerCheckedWithCycleTypes(program, typeChecker, entry, plan, false, options.Debug, options.ModuleFormat, options.FirstLinkTraps, nil, options.OwnershipQuery)
	if err != nil {
		return nil, err
	}

	return built.result, nil
}

// lowerChecked lowers a program as an ES module, the format every caller but LowerWithOptions lowers in.
func lowerChecked(program *load.Program, typeChecker *checker.Checker, entry *ast.SourceFile, plan *programMembership, discovery, debug bool, ownershipQuery ...bool) (*lowering, error) {
	return lowerCheckedWithCycleTypes(program, typeChecker, entry, plan, discovery, debug, ModuleFormatEcmaScript, false, nil, ownershipQuery...)
}

// Tests retain transient checker types before lazy queries replace them. A checked view's
// element reads are keyed by what the program views (viewElementKey). An array contract the key's
// pre-pass didn't see would leave a read it should check raw, so the key fails closed: the program
// is lowered again with every element read checked.
func lowerCheckedWithCycleTypes(program *load.Program, typeChecker *checker.Checker, entry *ast.SourceFile, plan *programMembership, discovery, debug bool, moduleFormat ModuleFormat, firstLinkTraps bool, cycleTypes map[int]*checker.Type, ownershipQuery ...bool) (*lowering, error) {
	key := &viewElementKey{}
	built, err := lowerWithElementKey(key, program, typeChecker, entry, plan, discovery, debug, moduleFormat, firstLinkTraps, cycleTypes, ownershipQuery...)
	if len(key.misses) != 0 {
		built, err = lowerWithElementKey(&viewElementKey{every: true}, program, typeChecker, entry, plan, discovery, debug, moduleFormat, firstLinkTraps, cycleTypes, ownershipQuery...)
	}
	if err != nil {
		return built, err
	}
	// A relation the checker gave up on answered "not related", and lowering may have read that answer
	// as a fact, so a program whose checker holds one is NotYet (#apyq5ve). Checked here, after the
	// lowering, so every caller gets it; a refusal or a NotYet already fails closed.
	if overflow := program.RelationOverflow(typeChecker); overflow != "" {
		stop := &NotYet{Where: built.result.Source, What: "a type relation the checker gave up on as too complex, whose answer lowering read (" + overflow + ")"}
		if !built.firstLinkProgramStop("RelationOverflow", stop) {
			return nil, stop
		}
		// A first-link lowering goes on, but past the budget every relation lowering asked answered not
		// related, wherever it was asked, a cycle the finder then didn't follow among them (#c0eypws). So no
		// row is a measurement of the program, this one's included, and every row is Suspect.
		built.firstLinkSuspectEvery()
	}
	return built, nil
}

func lowerWithElementKey(key *viewElementKey, program *load.Program, typeChecker *checker.Checker, entry *ast.SourceFile, plan *programMembership, discovery, debug bool, moduleFormat ModuleFormat, firstLinkTraps bool, cycleTypes map[int]*checker.Type, ownershipQuery ...bool) (built *lowering, err error) {
	lowering := &lowering{cycleTypes: cycleTypes, program: program, checker: typeChecker, result: &ir.Program{}, this: -1, functionIndex: -1, entry: entry, programPlan: plan, programDiscovery: discovery, elementKey: key}
	// A stop before checkHostMemberWrites judges a deferred host member write is main's stop at that write.
	defer func() { err = lowering.hostMemberPending(lowering.regexpServedPending(err)) }()
	// The program hands back every trap the lowering recorded, wherever it returns (first_link_traps.go).
	defer func() { lowering.result.Traps = lowering.traps }()
	lowering.debug = debug
	lowering.moduleFormat = moduleFormat
	lowering.firstLinkTraps = firstLinkTraps
	lowering.result.ProgramRegion = discovery || plan != nil
	if plan != nil {
		lowering.result.ProgramTypes = plan.types
	}

	// The base name only, so the same program emits the same C on every machine.
	lowering.result.Source = filepath.Base(program.FileName(entry))
	// Before any representation is asked, so a closed range type is held one way for the whole lowering
	// (range_types.go).
	lowering.rangeArraysPresent()
	modules, err := lowering.moduleOrder(entry)
	if err != nil {
		return nil, err
	}
	// Before any walk over the program's types, so a generic that instantiates itself deeper at each
	// level is refused where it's declared and no walk meets it (expansive_generics.go).
	if err := lowering.expansiveGenerics(modules); err != nil {
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
	if err := lowering.recordExports(entry); err != nil {
		return nil, err
	}
	for _, module := range modules {
		body, err := lowering.moduleStatements(module)
		if err != nil {
			return nil, err
		}
		lowering.result.Main = append(lowering.result.Main, body...)
	}
	// A first-link lowering records the stops lowering held where it met them (first_link_passes.go).
	lowering.firstLinkHeldStops()
	if lowering.unlowerable != nil {
		return nil, lowering.unlowerable
	}
	lowering.checkOverloadResults(modules)
	if err := lowering.checkOverloadVerdictCalls(modules); err != nil {
		if err = lowering.firstLinkPass("checkOverloadVerdictCalls", false, err, func() error { return lowering.checkOverloadVerdictCalls(modules) }); err != nil {
			return nil, err
		}
	}
	lowering.result.Main = append(lowering.forwarderValues, lowering.result.Main...)
	lowering.finishClassCalls()
	// Before the accessors' dispatchers: a read this makes checked can't go through a dispatcher's
	// unchecked fallback, which finishAccessors stops by asking CheckedField.
	if err := lowering.certifySourceSlots(); err != nil {
		if err = lowering.firstLinkPass("certifySourceSlots", true, err, lowering.certifySourceSlots); err != nil {
			return nil, err
		}
	}
	// Before finishAccessors, while a getter's or setter's read is still the ir.Property lowering made: a
	// read of a prototype member from a generic spread's copy stops (#y4gjeb8).
	if err := lowering.checkGenericSpreadCopyReads(); err != nil {
		if err = lowering.firstLinkPass("checkGenericSpreadCopyReads", false, err, lowering.checkGenericSpreadCopyReads); err != nil {
			return nil, err
		}
	}
	if err := lowering.finishAccessors(); err != nil {
		if err = lowering.firstLinkPass("finishAccessors", true, err, lowering.finishAccessors); err != nil {
			return nil, err
		}
	}
	lowering.finishToPrimitive()
	// Before exceptions: a view a generator feeds is rewritten into its dispatch here, and the
	// dispatch's steps are throws the fixed point must see, each its own generator's ResumeMayThrow.
	if err := lowering.checkIterableViews(); err != nil {
		if err = lowering.firstLinkPass("checkIterableViews", true, err, lowering.checkIterableViews); err != nil {
			return nil, err
		}
	}
	if err := lowering.exceptions(); err != nil {
		if err = lowering.firstLinkPass("exceptions", true, err, lowering.exceptions); err != nil {
			return nil, err
		}
	}
	if err := lowering.checkLibrarySpreads(modules); err != nil {
		if err = lowering.firstLinkPass("checkLibrarySpreads", false, err, func() error { return lowering.checkLibrarySpreads(modules) }); err != nil {
			return nil, err
		}
	}
	// All relations are recorded before judging whether a served method escapes as a value.
	if err := lowering.checkRegExpServedValues(); err != nil {
		if err = lowering.firstLinkPass("checkRegExpServedValues", false, err, lowering.checkRegExpServedValues); err != nil {
			return nil, err
		}
	}
	if err := lowering.checkAccessorSpreads(); err != nil {
		if err = lowering.firstLinkPass("checkAccessorSpreads", false, err, lowering.checkAccessorSpreads); err != nil {
			return nil, err
		}
	}
	if err := lowering.checkRecordGuardedReads(); err != nil {
		if err = lowering.firstLinkPass("checkRecordGuardedReads", false, err, lowering.checkRecordGuardedReads); err != nil {
			return nil, err
		}
	}
	if err := lowering.checkDynamicKeyedGetters(); err != nil {
		if err = lowering.firstLinkPass("checkDynamicKeyedGetters", false, err, lowering.checkDynamicKeyedGetters); err != nil {
			return nil, err
		}
	}
	if err := checkGeneratorRoads(lowering.result); err != nil {
		if err = lowering.firstLinkPass("checkGeneratorRoads", false, err, func() error { return checkGeneratorRoads(lowering.result) }); err != nil {
			return nil, err
		}
	}
	if err := flow.NormalizeAsync(lowering.result); err != nil {
		where := entry.AsNode()
		var located *flow.AsyncError
		if errors.As(err, &located) && located.Site > 0 && located.Site <= len(lowering.asyncSites) {
			where = lowering.asyncSites[located.Site-1]
		}
		stop := &NotYet{Where: lowering.program.Where(where), What: err.Error()}
		if err = lowering.firstLinkPass("NormalizeAsync", true, stop, nil); err != nil {
			return nil, err
		}
	}
	if lowering.result.AsyncEntry > 0 {
		// NormalizeAsync moved module bodies into one scheduler entry. Main now
		// holds only its generated invocation; retain boundaries of the actual Main.
		source := ir.SourceIdentity{Module: lowering.modulePath(entry), Declaration: []string{"module"}, Role: "async_module"}
		lowering.result.Functions[lowering.result.AsyncEntry-1].Source = source
		lowering.result.MainModules = []ir.MainModule{{Module: source.Module, Statements: len(lowering.result.Main)}}
	}
	if err := lowering.findCycles(modules); err != nil {
		if err = lowering.firstLinkPass("findCycles", false, err, func() error { return lowering.findCycles(modules) }); err != nil {
			return nil, err
		}
	}
	if err := lowering.checkBoundArguments(); err != nil {
		if err = lowering.firstLinkPass("checkBoundArguments", false, err, lowering.checkBoundArguments); err != nil {
			return nil, err
		}
	}
	if err := lowering.checkSlotWrites(); err != nil {
		if err = lowering.firstLinkPass("checkSlotWrites", false, err, lowering.checkSlotWrites); err != nil {
			return nil, err
		}
	}
	// Once every checked view is made: no view puts a getter-bearing object, unread, at a type an enumeration
	// the accessor guard admitted reads (#1atvfp5).
	if err := lowering.checkEnumerationReachViews(); err != nil {
		if err = lowering.firstLinkPass("checkEnumerationReachViews", false, err, lowering.checkEnumerationReachViews); err != nil {
			return nil, err
		}
	}
	// Once every checked view is made: no view puts a value stored as an array behind an arm Array.isArray's brand
	// decides, unread (#9z2ksv7).
	if err := lowering.checkArrayArmViews(); err != nil {
		if err = lowering.firstLinkPass("checkArrayArmViews", false, err, lowering.checkArrayArmViews); err != nil {
			return nil, err
		}
	}
	// Each admitted allocator getter road, by the seal, the uses and the reach of every new it feeds (#837zjkt).
	if err := lowering.checkFunctionConstructorRoads(); err != nil {
		if err = lowering.firstLinkPass("checkFunctionConstructorRoads", false, err, lowering.checkFunctionConstructorRoads); err != nil {
			return nil, err
		}
	}
	// Each built function constructor's instance laid by its construction window, once the program's exceptions
	// are known (#w8nanzw (B)).
	if err := lowering.layFunctionConstructorWindows(); err != nil {
		if err = lowering.firstLinkPass("layFunctionConstructorWindows", true, err, lowering.layFunctionConstructorWindows); err != nil {
			return nil, err
		}
	}
	if err := lowering.checkCheckedPredicateParameters(modules); err != nil {
		if err = lowering.firstLinkPass("checkCheckedPredicateParameters", false, err, func() error { return lowering.checkCheckedPredicateParameters(modules) }); err != nil {
			return nil, err
		}
	}
	// After findCycles, whose lazy view rule keeps a checked view's read of an optional callable its own
	// stop, and after finishAccessors, whose dispatcher a presence test mustn't have become.
	if err := lowering.checkOptionalMemberPresence(); err != nil {
		if err = lowering.firstLinkPass("checkOptionalMemberPresence", false, err, lowering.checkOptionalMemberPresence); err != nil {
			return nil, err
		}
	}
	if err := lowering.checkViewCopies(); err != nil {
		if err = lowering.firstLinkPass("checkViewCopies", false, err, lowering.checkViewCopies); err != nil {
			return nil, err
		}
	}
	// Before checkViewWidenedReads weighs the views: a relation rule 3300 let cross that no lowering viewed
	// is refused in that rule's words, never read raw.
	if err := lowering.checkRelationViewsTaken(); err != nil {
		if err = lowering.firstLinkPass("checkRelationViewsTaken", false, err, lowering.checkRelationViewsTaken); err != nil {
			return nil, err
		}
	}
	if err := lowering.checkViewWidenedReads(); err != nil {
		if err = lowering.firstLinkPass("checkViewWidenedReads", false, err, lowering.checkViewWidenedReads); err != nil {
			return nil, err
		}
	}
	// Every road that reads an object operand's layout raw, by range_unions.go's table, where that operand
	// may be a range type's Union a NodeArray stands for (#hae423v, Lower's Must 2).
	if err := lowering.checkRangeUnionReads(); err != nil {
		if err = lowering.firstLinkPass("checkRangeUnionReads", false, err, lowering.checkRangeUnionReads); err != nil {
			return nil, err
		}
	}
	if err := lowering.checkHostMemberWrites(); err != nil {
		if err = lowering.firstLinkPass("checkHostMemberWrites", false, err, lowering.checkHostMemberWrites); err != nil {
			return nil, err
		}
	}
	if len(lowering.asyncPayloadFailures) > 0 {
		return nil, lowering.asyncPayloadFailures[0]
	}
	if err := lowering.checkPromiseConsumption(); err != nil {
		if err = lowering.firstLinkPass("checkPromiseConsumption", false, err, lowering.checkPromiseConsumption); err != nil {
			return nil, err
		}
	}
	if err := lowering.checkViewLogicalWrites(); err != nil {
		if err = lowering.firstLinkPass("checkViewLogicalWrites", false, err, lowering.checkViewLogicalWrites); err != nil {
			return nil, err
		}
	}
	// After findCycles, whose resolveCountTypes fills FunctionTypeTargets, the target set each closure call's
	// layout packs for: a spread through a function value is judged against every target its layout admits
	// (#wzr4atz).
	if err := lowering.checkClosureSpreadLayouts(); err != nil {
		if err = lowering.firstLinkPass("checkClosureSpreadLayouts", false, err, lowering.checkClosureSpreadLayouts); err != nil {
			return nil, err
		}
	}
	if discovery || plan != nil {
		if lowering.result.HasAsync() {
			stop := &Refused{Where: program.Where(entry.AsNode()), What: "Program region with async work", Fix: "keep Program members in the one-shot synchronous CLI"}
			if !lowering.firstLinkProgramStop("ProgramRegion", stop) {
				return nil, stop
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
		if err = lowering.firstLinkPass("voidDiscards", false, err, lowering.voidDiscards); err != nil {
			return nil, err
		}
	}
	if err := lowering.checkInspectorDoomedContracts(); err != nil {
		if err = lowering.firstLinkPass("checkInspectorDoomedContracts", false, err, lowering.checkInspectorDoomedContracts); err != nil {
			return nil, err
		}
	}
	// After the last pass: no function reference is added past here, so every function a flow graph
	// took to never run must still be named by nothing (view_flow.go's runningFunctions).
	if err := checkFlowLiveness(lowering.result, lowering.flowDead); err != nil {
		if err = lowering.firstLinkPass("checkFlowLiveness", false, err, func() error { return checkFlowLiveness(lowering.result, lowering.flowDead) }); err != nil {
			return nil, err
		}
	}
	// Last, after every pass that lowers a body or makes IR: a write Rule C held at the view walk stops only
	// where its statement was lowered, and none is lowered past here (source_slot_certificate.go, #5qx6qv4).
	if err := lowering.checkSourceSlotPending(); err != nil {
		if err = lowering.firstLinkPass("checkSourceSlotPending", false, err, lowering.checkSourceSlotPending); err != nil {
			return nil, err
		}
	}
	// Once no pass builds or changes a view contract: each admission reads its fields' finished
	// contracts, and only the emitters read it (finishViewStaticsAdmitted).
	lowering.finishViewStaticsAdmitted()
	return lowering, nil
}

type lowering struct {
	inspectorRequireKnown     bool // Lazy, immutable program facts; no IR or unit ownership.
	inspectorRequirePresent   bool
	programClassesKnown       bool // programClassesConvert's answer, read once per program.
	programClassesAnswer      bool
	inspectorSkippedContracts map[*ast.Node]int // Skipped AST reads -> unique Panic slots, owned by this IR snapshot.
	bindGetters               map[string]bool   // source getter names, memoized for bind receiver admission
	debug                     bool
	cycleTypes                map[int]*checker.Type // Optional differential-test capture; nil in normal builds.
	// moduleFormat is the module format the build declares (Options.ModuleFormat). A lowering a test
	// builds by hand leaves it ModuleFormatEcmaScript.
	moduleFormat ModuleFormat

	// firstLinkTraps is Options.FirstLinkTraps: a unit that stops lowers as a trap instead
	// (first_link_traps.go). traps are the rows its traps recorded, in the order met, which the lowering
	// hands back as its program's Traps, and firstLinkSuspect says one of them was a recovered Go panic or
	// an error that's neither a NotYet nor a Refused, either of which may leave lowering's state part
	// written, so each row after it is Suspect.
	firstLinkTraps   bool
	firstLinkSuspect bool
	traps            []ir.Trap
	// firstLinkTrapped is every unit trapped, by its node, and firstLinkStubs the trap each lowers as
	// (first_link_units.go). signatureTraps is each module function whose signature trapped, by its symbol,
	// with its trap's message, which a reference to it stops naming (signatureTrap).
	firstLinkTrapped map[*ast.Node]bool
	firstLinkStubs   map[*ast.Node]ir.Panic
	signatureTraps   map[*ast.Symbol]string
	// firstLinkMain is where each top-level statement's IR lies in Main (firstLinkStatements), which a
	// whole-program pass's trap stubs, and firstLinkRewriteSkipped says such a pass that rewrites the IR
	// skipped its work past a stop, so each row after it is Suspect (first_link_passes.go).
	// firstLinkJudgmentAbandoned says a pass that only judges gave up partway, which makes each row after
	// it Suspect the same way (firstLinkJudgmentAbandon).
	firstLinkMain              map[*ast.Node]firstLinkMainRange
	firstLinkRewriteSkipped    bool
	firstLinkJudgmentAbandoned bool

	// flowDead is every function a flow graph built during lowering took to never run (flowGraph,
	// iterableViewFlowGraph), which checkFlowLiveness holds to that once lowering is done.
	flowDead map[int]bool

	// firstLinkHelperUnits is, in a first-link lowering, the unit whose lowering made each helper, a function with no
	// declaration, by its index (firstLinkNoteHelpers), which a trap stubs once every unit owning it is trapped
	// (first_link_helpers.go).
	firstLinkHelperUnits map[int]*ast.Node

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
	// graphRecords are the record literals, each with its own record type, which graphTypes reads in a
	// graph program: a record of a type it didn't select never joins a graph region, so one its own
	// values can reach back is a counted node on a cycle (#t5q4prw, #1vb0x5s). Like graphNodeArrays, a
	// site from a literal that then fails stays listed, which can only over-stop.
	graphRecords []graphRecordSite
	// dynamicKeyedSite is the first readonly view read by a run-time key, where
	// checkDynamicKeyedGetters says why a class getter keeps the program from reading by a run-time key.
	dynamicKeyedSite *ast.Node
	reflectKeyedSite *ast.Node
	// getterDeclarations is each class getter's declaration, by its function (registerAccessor), which
	// checkDynamicKeyedGetters reads.
	getterDeclarations map[int]*ast.Node
	// staticMethodNames names every public static method (never an accessor) any class lowering
	// lowered declares, by its plain name (staticInstance), which viewStaticsAdmitted reads: the run
	// time keeps a static method by index with no name, so its walk can't see one (#24kvb6k).
	staticMethodNames map[string]bool
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

	// regexpServedView records whether any relation serves RegExp test through a program view;
	// regexpServedReads holds candidates for the program-wide checkRegExpServedValues judgment.
	regexpServedView  bool
	regexpServedReads []regexpServedRead
	// viewSources is each view's source and target, and viewTypes the checker types a read's view
	// ids name, its receiver's and its field's, so the passes after lowering can judge and name them
	// (view_widened.go).
	viewSources []viewSource
	viewTypes   map[int]*checker.Type
	// argumentTypes is, at the number an ArgumentTypeIDs, InitialTypeID or ArrayTypeID names less one, an
	// argument's checker type and the slot it's passed into, so graphFlows can judge an argument handed to
	// a function value whose targets are Unknown (argumentTypeID, #x490w4a).
	argumentTypes [][]*checker.Type
	// viewSites is every property access, keyed write and named getter of the program's modules, in
	// view's preorder, built once and shared by every lowering a latent run makes (view_sites.go,
	// #q1hyhzc).
	// checkedFieldWrites counts, by field name, the CheckedFields keys markCheckedField added;
	// viewSiteSeen is a name's count when view last judged its accesses, and viewSiteTrue the accesses
	// it found checked, which stay checked since keys only grow. viewSiteJudged is the refusal
	// judgeViewAccess gave a checked access, nil for none, kept where no substitution is in force, so a
	// later walk reads it instead of judging the access again (viewWalkIndexed, #715zjph).
	viewSites          *viewSiteIndex
	checkedFieldWrites map[string]int
	viewSiteSeen       map[string]int
	viewSiteTrue       map[*ast.Node]bool
	viewSiteJudged     map[*ast.Node]error
	// checkedFieldRecordFree says, per CheckedFields key markCheckedField added, whether every view that
	// registered it was one no record can be read through (viewRecordFree): true while each was, false
	// once one wasn't, absent for a key added another way. checkRecordGuardedReads stops a read only on a
	// key that isn't (#5d9y3rs). viewContractTargets is each type viewContract was asked to describe, by
	// id, so a view's walk can name the types its contracts stand for.
	checkedFieldRecordFree map[string]bool
	viewContractTargets    map[int]*checker.Type
	// placeholderSlots is notePlaceholderSlots' answer by checker and program, shared by every lowering a
	// latent run makes; production leaves it nil, and each lowering walks (placeholder.go, #6xt6p1f).
	placeholderSlots *placeholderSlotCache
	// viewedElements is walkViewedElements' answer by checker and program, shared by every lowering a latent
	// run makes; production leaves it nil, and each lowering walks (viewed_elements_cache.go, #hjspb3b).
	viewedElements *viewedElementsCache
	// inheritance is walkInheritance's answer by checker and program, shared by every lowering a latent run
	// makes; production leaves it nil, and each lowering walks (inheritance_cache.go, #hjspb3b).
	inheritance *inheritanceCache
	// accessorNamesHeld is walkAccessorNames' answer by program, shared by every lowering a latent run makes;
	// production leaves it nil, and each lowering walks (accessor_names_cache.go, #hjspb3b). accessorNames,
	// below, is the answer this lowering holds.
	accessorNamesHeld *accessorNamesCache
	// expansiveGenericsHeld is walkExpansiveGenerics' stop by checker and program, shared by every lowering a latent
	// run makes; production leaves it nil, and each lowering walks (expansive_generics_cache.go, #bmqm0k1). It's
	// Held because expansiveGenerics is the method that reads it.
	expansiveGenericsHeld *expansiveGenericsCache
	// overloadCheckCalls holds every overload result check function (overloadResultCheck), whose call is
	// where its value enters a view (checkViewWidenedReads).
	overloadCheckCalls map[int]bool
	// narrowOverloads memoizes narrowOverload; overloadCalls holds each call to a narrow overload the
	// lowering lowered, true once overloadResultCheck settled it, which checkOverloadResults reads after
	// lowering (overload_result_check.go, #z5d0jye).
	narrowOverloads map[*ast.Node]bool
	overloadCalls   map[*ast.Node]bool
	// overloadCertificates holds, for each overload judged, whether censusOverload accepted it by
	// certificate A, B or D, whose calls overloadCertificateCall holds to a direct entry
	// (overload_certificate_array.go, overload_certificate_map.go, #5fws6q8, and
	// overload_certificate_filter.go, #4z7y6pq).
	overloadCertificates map[*ast.Node]bool
	// overloadNodeTests holds, for each overload judged, whether censusOverload accepted it by the
	// node-test proof, whose direct calls overloadNodeTestCall judges one by one
	// (overload_parse_tree_node.go, #0kkpw2e).
	overloadNodeTests map[*ast.Node]bool
	// overloadGuardedArrays holds, for each overload judged, whether censusOverload accepted it by the
	// guarded-array certificate, whose direct calls overloadGuardedArrayCall judges one by one, and
	// overloadGuardedViews what that judgment reads of the whole program, collected once
	// (overload_certificate_visit_nodes.go, #ytagqs5).
	overloadGuardedArrays map[*ast.Node]bool
	overloadGuardedViews  *overloadGuardedViews
	// overloadVerdicts holds, for each overload of a function declaration's set the census judged, its
	// verdict: censusOverload's Refused or NotYet, or nil when it's served. A set's implementation is keyed
	// too, always nil, once its set is judged (censusOverloads, #4czdn4q).
	overloadVerdicts map[*ast.Node]error
	// overloadVerdictCalls holds each call, new or tagged template overloadCallVerdict let through to an
	// overload of a function declaration's set, which checkOverloadVerdictCalls reads after lowering.
	overloadVerdictCalls map[*ast.Node]bool
	// predicateParametersAdmitted holds each predicate parameter predicateArgumentProof admitted for its
	// checked closure, and predicateParametersBuilt each one lowering passed as that closure, which
	// checkCheckedPredicateParameters holds alike once lowering is done (predicate_parameter_value.go).
	predicateParametersAdmitted map[*ast.Node]bool
	predicateParametersBuilt    map[*ast.Node]bool
	// optionalPresenceReads are the presence tests of an optional member lowering read off its slot,
	// judged by what their receivers may hold once the program is lowered (optional_member_presence.go).
	optionalPresenceReads []optionalPresenceRead
	// dispatchedPresenceReads are the presence reads finishAccessors made a dispatcher's call, each with the
	// key the read had before the dispatcher's fallback zeroed it and what its receiver may hold, read
	// there from the reach and owners the dispatch was decided by (optional_member_presence.go). It holds
	// plain values, never the graph, so the latent census copies it as it copies any lowering field.
	dispatchedPresenceReads []dispatchedPresenceRead
	// dispatchedReads are the field reads finishAccessors made a dispatcher's call, each as lowering made it,
	// its view fields and receiver whole, which the dispatcher's fallback zeroed: checkViewWidenedReads weighs
	// them as it weighs a read that stayed a Property (#3jj9dzg).
	dispatchedReads []ir.Property
	// hostMemberWrites are the writes to a method a program's interface or type literal declares, each
	// with main's stop for it, which lowering stores as fields and checkHostMemberWrites judges once
	// lowering is done (host_member_write.go).
	hostMemberWrites []hostMemberWrite
	// hostMemberJudged says checkHostMemberWrites has run, so a stop after it is the program's own
	// (hostMemberPending).
	hostMemberJudged bool
	// hostMemberValues are the reads of such a method as a value, each with main's stop for it, which
	// lowering reads as fields and checkHostMemberValues judges with the writes (host_member_value.go).
	hostMemberValues []hostMemberValue
	// hostMemberThroughs is the thisArg of each .call of such a value lowering made, as it lowered it, by
	// the call, which checkHostMemberValues holds to the literal the value was read off (host_member_value.go).
	hostMemberThroughs map[*ast.Node]ir.Expression
	// hostMemberFirst is the first stop deferred, a write's or a value read's: main's stop, which a
	// lowering that stops before they're judged gives back (hostMemberPending).
	hostMemberFirst error
	// sourceSlotWrites are the writes of an object field through a checked view, each with the type
	// it writes, which certifySourceSlots holds every read of the field's name to once lowering is
	// done, and sourceSlotsJudged says it has, so a write found after that keeps its stop
	// (source_slot_certificate.go, #2msw6t8).
	sourceSlotWrites  []sourceSlotWrite
	sourceSlotsJudged bool
	// sourceSlotPending are the plain writes to a checked field that may hold undefined or null, each with Rule
	// C's NotYet for it, in the view walk's order, which recordSourceSlotWrite holds instead of stopping there,
	// and loweredStatements every expression statement lowering began to lower (statement): once lowering is
	// done, checkSourceSlotPending gives back the first held stop whose statement was lowered (#5qx6qv4).
	sourceSlotPending []sourceSlotPendingStop
	loweredStatements map[*ast.Node]bool
	// relationViews are the relations rule 3300 let cross as a checked view (rulings.md:72, #0b9xqhh), by
	// the node whose lowered value enters the slot, recorded before lowering and never deleted, and
	// relationViewsTaken the nodes some lowering viewed, which checkRelationViewsTaken holds every entry
	// to once lowering is done (relation_views.go).
	relationViews      map[*ast.Node]relationView
	relationViewsTaken map[*ast.Node]bool
	// relationContainerOrigins are the arrays a relation view views as a readonly array (rulings.md:72's
	// containers), each where its value enters the view, whose elements checkViewArrays holds every
	// Object.keys, values and entries operand away from.
	relationContainerOrigins []ir.Expression
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
	// optionalChains are the optional chains optionalChain is guarding, innermost last, each held only
	// while its steps lower (optional_chain.go).
	optionalChains []optionalChainBinding
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
	// rangeArrays is the program bit, that some interface it declares may have the NodeArray layout, once
	// rangeArraysKnown; rangeShapes memoizes rangeShape per type, and rangeUnionsHeld is whether any type was
	// held as a range type's Union (range_types.go). rangeFields, once rangeFieldsKnown, is each field name a
	// value held as a range type's Union may be read by, and rangeFieldsAny that a field of unknown name may be
	// one (range_unions.go).
	rangeArrays, rangeArraysKnown bool
	rangeShapes                   map[*checker.Type]bool
	rangeUnionsHeld               bool
	rangeFields                   map[string]bool
	rangeFieldsKnown              bool
	rangeFieldsAny                bool
	// objectIntersections are the intersections objectIntersection is proving, each at its depth on the
	// walk's path, so a field that holds one of them again ends there.
	objectIntersections map[*checker.Type]int
	// objectIntersectionVerdicts is objectIntersection's answer per intersection and type mapper, kept once
	// it's settled (#vp9f4yd): a failure always, a success once its walk leaned on no intersection still
	// being proven above it. objectIntersectionLow is the shallowest depth such a lean reached, for the
	// walk in progress.
	objectIntersectionVerdicts map[objectIntersectionKey]objectIntersectionVerdict
	objectIntersectionLow      int
	// elementFits is elementsHeldOtherwise's walk answered once per value type and slot type, so
	// judging every value handed to a slot costs each pair once (element_fit.go).
	elementFits map[elementFitKey]*elementMismatch
	// fieldFitFresh is set while elementsHeldOtherwise walks a value whose every object is made right there
	// (freshObjectValue), so fieldsSeenAs leaves its scalar fields to their own slots.
	fieldFitFresh bool
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
	// numericKeyTables is numericKeyTableClosed's answer per const binding of a literal with numeric computed keys,
	// and numericKeyTableWalks how many times the program was walked for one (numeric_keys.go).
	numericKeyTables     map[*ast.Symbol]bool
	numericKeyTableWalks int
	// predicateProofs is every type predicate's proof, run once per verifier mode (predicates_proof.go).
	predicateProofs *predicateProofs

	// foreignFunctions are the functions Apple calls lower to, by class, selector and call shape
	// (foreign.go).
	foreignFunctions map[string]int

	// foreignDelegates are the Objective-C classes made for the program's classes handed to Apple, by
	// instantiation (foreign_delegate.go).
	foreignDelegates map[*instance]*ir.Delegate

	// delegateHolds are the delegates handed to Apple, each with what holds it from then on, which
	// the cycle finder holds to the same rule as a field (cycles.go).
	delegateHolds []delegateHold

	// appleTypes are the classes and protocols the program's binding files declare, by Objective-C
	// name, and appleHanded the types of the closures the program hands Apple as blocks and actions:
	// what the cycle finder follows into when an Apple object may hold one (cycles.go).
	appleTypes  map[string]*checker.Type
	appleHanded []*checker.Type

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
	// enumTagCheckedViewsDone holds each declared union and member pair the swing registered a view for
	// (enumTagCheckedViews), so a pair registers once. A pair whose view failed isn't held, so each site asks again.
	enumTagCheckedViewsDone map[enumTagCheckedViewKey]bool
	// closedKindAnswers is closedKinds' answer by type and field, and closedKindDeclared and closedKindBuilt
	// the program's declared types and build sites it judges each type against, found once; closedKindNewOnly
	// is each function declaration with a declared this, and whether every use of it is a new callee or a
	// cast to any; closedKindEverywhere is every declared type in every file the program holds, its
	// declaration files included, and closedKindFits each closedKindSourcesFit answer (enum_tag_closed_kind.go).
	closedKindAnswers    map[closedKindKey]closedKindAnswer
	closedKindDeclared   []closedKindSite
	closedKindBuilt      []closedKindSite
	closedKindNewOnly    map[*ast.Symbol]bool
	closedKindEverywhere []closedKindSite
	closedKindFits       map[closedKindFitKey]bool
	// functionConstructorScan is the program's admitted allocator getter roads, found once, each function
	// constructor among them built where a new first needs it; functionConstructorNews is each new lowered with
	// a function constructor among its candidates, and anyRoads each stop an admitted road deferred, all judged
	// once lowering is done (checkFunctionConstructorRoads, #837zjkt).
	functionConstructorScan *allocatorRoads
	functionConstructorNews []functionConstructorNew
	anyRoads                []anyRoad
	// dispatchCastScan is the program's dispatch lookups under ruling 17, found once: each admitted one and each
	// declined candidate's reason, by its assertion (dispatch_cast.go).
	dispatchCastScan *dispatchCastRoads
	// kindTagCarriers is every declared type carrying a literal on a tag field, by the field and the literal's
	// runtime value, and kindTagFields the fields it holds (cast_kind_union_writes.go, #b5w3ycg head 2).
	kindTagCarriers map[string][]*checker.Type
	kindTagFields   map[string]bool
	// arrayStoredNames is every property name a type stored as an array can have, found once, the first time
	// Array.isArray meets a plain object arm (library_array_predicate.go, #s8gy89r).
	arrayStoredNames map[string]bool
	// arrayStoredIndex is whether a type stored as an array in the closed world has an index signature other than a
	// number's, found by the same walk, at the same time (arrayExcludedIndexArm, #gcec0bx).
	arrayStoredIndex bool
	// arrayArmSites are the tests lowering built whose array brand decides an arm arrayExcludedArm or
	// arrayExcludedIndexArm admitted, which checkArrayArmViews holds against every checked view once lowering is done
	// (library_array_predicate.go).
	arrayArmSites []arrayArmSite
	// closureSpreadSites are the calls with a spread through a function value callClosure made, which
	// checkClosureSpreadLayouts judges against every target their layouts pack for once FunctionTypeTargets
	// is filled (arguments_length.go, #wzr4atz).
	closureSpreadSites []closureSpreadSite
	// slotWrites is each write admitted only where every allocation reaching it holds a slot for its field,
	// judged once lowering is done (checkSlotWrites), slotWritesKept each kept, by its fields but its proof
	// (recordSlotWrite), and builtTypes the type each object literal is built at, by the id of its own type, its
	// GraphTypes' first (noteBuiltType).
	slotWrites     []slotWrite
	slotWritesKept map[slotWrite]bool
	builtTypes     map[int]builtType
	// enumerationReachSites are the enumerations the accessor guard admitted beside an accessor, each with the
	// read types it asked by, which checkEnumerationReachViews holds against every checked view once lowering is
	// done (enumeration_accessor_reach.go, #1atvfp5).
	enumerationReachSites []enumerationReachSite
	// crossFunctionWrites holds the first write to each binding from outside its own function.
	crossFunctionWrites map[*ast.Symbol]*ast.Node
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
	// A concrete class layout can omit a merged field whose type has no storage, or whose slot could
	// point onward (mergedLeaf): each class's omitted names, with the words a read or write of one stops in.
	missingMergedSlots map[int]map[string]string
	classes            map[*ast.Symbol]*ast.Node
	statics            map[*ast.Symbol]*instance
	staticGlobals      map[*ast.Symbol]int
	instances          map[string]*instance
	derivedAncestors   map[*ast.Symbol]bool
	optionalClasses    *optionalClassHierarchy

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

	// dictionaryViewUses is the dictionary view's use test's answers, by cast and by declaration, made the first
	// time it's asked for (cast_dictionary_view.go).
	dictionaryViewUses *dictionaryViewMemo

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
	// ownAnswerValues, once ownAnswerValuesKnown, is each type a call, new, property or element read in the
	// program has that libraryObjectOwnAnswer names, a match's groups aside, with that name, first seen first
	// (prototypeHazard, #v414yhr).
	ownAnswerValues      []ownAnswerValue
	ownAnswerValuesKnown bool
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

	// droppedContractsRead is set while recordStorageView, or the optional-widening rule, judges a .ts
	// module's view, where a record seen as a record may drop or add a named contract the reads hold
	// (droppedContractRead and addedContractRead, records_named.go).
	droppedContractsRead bool

	// readonlyKinds is what the program's first view site into each readonly string-index type put there,
	// a record or a fixed object (records_readonly.go), and readonlyDeciding is set while it's walked.
	readonlyKinds    map[*checker.Type]readonlyDecision
	readonlyDeciding bool

	// looseInits is every literal field, element or entry lowering made into a slot typed unknown or
	// object, judged by the loose-slot stop (loose_slots.go).
	looseInits []looseInit

	// writeSites is every write into a slot lowering made, by its IR node's Site less one (fresh.go).
	writeSites       []writeSite
	accessorCaptures []accessorCapture
	accessorNames    map[string]bool
	// contractGetters is each view's getters of a contract field's name, judged by reach in
	// finishAccessors (class_accessors.go).
	contractGetters                      []contractGetters
	callbackReadProofs                   map[*ast.Node][]overloadCallbackReadProof
	callbackReadUses                     []overloadCallbackReadProof
	callbackActivated                    map[*ast.Node]bool
	callbackSliceKnown, callbackOwnSlice bool

	// spreadSources is the type each object literal's spread sees its source as, by the literal's
	// SpreadSite less one, and classHolders each class's instance type by its one-based index, the
	// two sides of the assignability spreadGetters asks (class_accessors.go). spreadNodes is each spread
	// itself by the same index, where a SpreadWith literal's accessor stop is said (spreadAccessorClass).
	spreadSources []*checker.Type
	spreadNodes   []*ast.Node
	classHolders  map[int]*checker.Type
	// genericSpreadSites are the SpreadSites of the spreads whose source the checker types with a type
	// parameter, or with a type that may hide a program class's instance, each with how a stop names its
	// copy, and prototypeMemberUses the class method calls and presence tests lowering made on a
	// receiver other than this, which checkGenericSpreadCopyReads judges (generic_spread_copy_reads.go).
	genericSpreadSites  map[int]string
	prototypeMemberUses []prototypeMemberUse

	// spreadWithCopied is spreadCopiedGetters' copied for SpreadWith literals alone: each getter name one may
	// leave out of its copy, with the class holding it (spreadWithSourceGetter).
	spreadWithCopied map[string]string

	// toPrimitive is the program's to_string_hint helper plus one, or zero before a String of a union
	// that may hold an object needs it (to_primitive.go).
	toPrimitive int

	// keptBufferUnionView is set when a keeping walk met a nodeBufferUnion seen as another type at some
	// depth (nodeBufferUnionSeenAs), so valueSlotKeeps can stop it with the Buffer words.
	keptBufferUnionView bool
}
