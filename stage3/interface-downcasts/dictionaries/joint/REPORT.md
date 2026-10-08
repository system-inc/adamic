Built checked readonly dictionary conversion and transitive reads: lane 4b 2 pairs / 11 reads, lane 4 3 / 14, lane 6 2 / 11.
Commits: 8f5ac7cb shared consumption checks; dbeb7b5c original-pair certification; a1b0298c lane 6 recheck evidence; first-lane checkpoint bd3ede0d was pushed.
Checks: joint suite passes in 50.460s with Node, sanitized native, release native, JavaScript, successful-case leaks and scoped counts; logs accompany this report.
Mutants: skipped element-type checks, accepted array and nominal receiver shapes, and dropped transitive checks all lose pinned exit-70 refusals in both backends; allocation-origin guard removal also loses its named frontend refusal.
Uncovered: lane 6 any 1 pair / 20 reads, six entry-tuple consumers, writable conversions and other storage-dependent consumers without producer certificates; global count refresh fails outside this unit.

The boundary is internal/lower/records.go, recordStorageView, at its
sameRecordStorage comparison. The selected CompilerOptionsValue arrives as
ir.Union, a counted heap reference whose payload can be a scalar box, array,
fixed object or record. The dictionary consumer needs an own string-key read
that checks representation before following storage, then checks the returned
element against its read contract. buildoptions-good.a is the smaller original
pair's witness. Its essential consumption is:

    const selected=carrier.child['item'];
    const dictionary=selected as {readonly [key:string]:string[]|undefined};
    const values=dictionary['name'];
    console.log(values===undefined ? 'missing' : values[0] ?? 'missing');

checkedDictionaryReferenceConversion in internal/lower/view_dictionary_conversion.go
emits a marked ir.Narrow. Native lowering calls
adamic_view_dictionary_source_conversion in runtime/view_dictionaries.c;
JavaScript calls adamicViewDictionarySource. This checks the heap tag when
the selected reference is consumed, before exposing it as a dictionary.
Arrays, scalar boxes, Maps, functions, null, undefined and nominal or opaque
objects stop with a pinned diagnostic. It returns the same borrowed object;
it neither copies keys nor scans elements.

Each later lookup calls adamic_view_dictionary_source_read. It independently
checks the receiver and uses adamic_record_is to distinguish the existing
runtime/record.c representation from fixed-object storage. Fixed-object fields
use checked slots; records use the existing record lookup. Each result's type
is checked when read. An absent key is undefined when the indexed read type
allows it, including the checker's inferred undefined. A selected reference
retains its child contract, so reading values[0] checks the string element.
No second record representation was written and runtime/record.c was unchanged.

The separate DictionaryReference selector retains the original descriptor
and type ID, registers descendant schemas and survives readiness processing.
Direct String consumers retain the primitive selector because reference
ToPrimitive is outside this subset. Unsupported contracts are not widened.
Only readonly dictionary targets without named properties qualify for this
conversion. Mutable conversion retains its storage refusal. Allocation-flow
inspection also refuses aliased writes, coalescing, spreads, probes and
storage-dependent operations without a producer certificate. Unknown
allocation origins conservatively refuse such operations, which can reject
unrelated operations when the graph cannot establish separation. This is a
restriction of the implemented read subset, not permission to mutate storage.

The complete original TypeScript 6.0.3 declarations are from
050880ce59e30b356b686bd3144efe24f875ebc8. Tests assert original field inventories.
BuildOptions, CompilerOptions and OptionsBase each have eight controls:
fixed-object and record producers, explicit and inferred missing-key cases,
wrong dictionary element value, wrong array element value, array receiver and
nominal receiver. Node establishes the source behavior. Both native modes and
JavaScript must produce the pinned checked-view behavior; successful controls
also pass leak checks. Stopping panic paths terminate before cleanup and their
counts are recorded separately. Three array-write controls retain the checked
producer guard. Six lane 6 array/map/source rechecks now compile and run in
both backends and pass successful-case leaks.

Pinned failures have prefix adamic: panic:, trailing newline and exit 70:

    field read failed: selected; expected { readonly [key: string]: string[] | undefined; }, found array
    field read failed: selected; expected { readonly [key: string]: string[] | undefined; }, found unsupported representation
    field read failed: dictionary['name']; expected string[] | undefined, found number
    element read failed: values[0] expected string, found number
    element read failed: <array write> expected string, found number

For every original pair, accepting a wrong array or nominal representation
makes the typeof-only receiver control finish with object instead of stopping.
Skipping the selected element check safely normalizes the numeric value to
undefined and finishes with missing. Dropping the transitive array check
safely returns string representation 42 and finishes with 42. All mutants
compile; native mutants run with sanitizers. A compiler error or invalid
memory access is never credited as a kill. Removing the conversion-origin
marker from the allocation-guard unit control loses the frontend refusal.
Existing primitive first-member, null/undefined and unsupported-function
mutants remain covered by the prior lane's regression tests.

| Lane | Newly certified pairs / candidate reads | Left in requested dictionary scope |
| --- | --- | --- |
| 4b | BuildOptions and CompilerOptions, 2 / 11 | 0 / 0 |
| 4 | Those two plus OptionsBase, 3 / 14 | 0 / 0 |
| 6 | BuildOptions and CompilerOptions, 2 / 11 | any, 1 / 20 |

These are static candidate counts, not measured production tsc execution.
The overlapping lanes must not be added. Lane 4b's overall ledger advances
from 37 / 165 to 39 / 176; unrelated EmitHelper.text 1 / 3 and
fileInfos.forEach 2 / 2 remain outside this dictionary unit. Lane 6's ledger
advances from 26 / 197 to 28 / 208 of 29 / 228. The any-valued pair remains
excluded because the subset has no checked erased-value contract; its named
frontend refusal is asserted. The six excluded entry-tuple fixtures are
entries-fixed-array-{good,wrong,empty}.a and
entries-producer-array-{good,wrong,empty}.a. Each still produces the named
unsupported tuple-contract refusal at result[element]. They require
producer-specific ownership and per-position consumption contracts.

Commands ran with source /workspace/adamic-tools/env.sh, original declarations
at /tmp/views-dictionary-declarations and test output redirected to log files:

    ADAMIC_BRAND_ORIGINAL_DECLS=/tmp/views-dictionary-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewDictionaryJoint($|Counts$|Recheck$|ArrayWrite$)' -count=1 -v
    ADAMIC_BRAND_ORIGINAL_DECLS=/tmp/views-dictionary-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewOriginalDictionaryPrimitiveComponents$|^TestCheckedViewDictionaryValuesEntries$' -count=1 -v
    go test ./internal/lower ./internal/ir -run 'TestDictionaryConversionRecordOperationOriginMutant|TestDictionaryReferenceProjection|TestSelectedDictionary' -count=1
    go test ./internal/ir -run '^TestPrimitiveDictionary' -count=1
    ADAMIC_BRAND_ORIGINAL_DECLS=/tmp/views-dictionary-declarations go test ./internal/oracle -run '^TestCheckedViewDictionaryJointCounts$' -count=1 -args -update-counts
    go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts

The joint suite passes in 50.460s, scoped refresh in 14.546s, and lower proof
controls in .420s; primitive/tuple regression in 47.913s and IR in .007s. The first lower/IR command matched no IR tests; the exact
primitive IR tests were then run separately. joint/counts.md records all 25
fixtures and the final suite verifies it. The required global refresh failed
in 57.718s at FS evaluated option literals, process.exit values and other
fixtures, leaving root counts.md unchanged. See global-counts.log for exact
errors. Declaration-backed fixtures use their own count registration; no
protected oracle registry was edited. No whole package or full gate was run.

Setup: GOPROXY=https://proxy.golang.org|direct, bash cloud/setup.sh, then the
printed env.sh. Timing lines: Node .061s, Go .093s, clang .503s, markdown
1.942s, submodules 308.155s, Go build 482.968s, test binaries deferred
483.060s, build cache warm 483.062s, done 483.093s. nproc=5, CPU quota four.
No other lane branch was merged, no protected implementation file was edited,
and no cohere code was copied. Shared boundary changes and lane 6 evidence
have separate commits. Delivery is codex/views-dictionary-joint, based on
1771221bd232bd08618669e587e4566be8c252b9. The first lane's passing checkpoint
was pushed once; the completed unit carries the correction that checks
representation before exposing the dictionary reference.
