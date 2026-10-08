Built checked readonly dictionary consumption for the original BuildOptions and CompilerOptions pairs, 2 pairs / 11 candidate reads in lane 4b.
Commits: shared conversion 8861b316; dictionary selection and controls c77f7326; prior lane control adjustment 29ffa565.
Checks: original declarations, Node, sanitized and release native, JavaScript, leaks and scoped counts pass; command logs accompany this report.
Mutants: skip selected type checking, accept the wrong receiver shape, and drop the transitive element check each lose a pinned exit-70 refusal in C and JavaScript.
Uncovered at this checkpoint: OptionsBase 1 pair / 3 reads, any 1 pair / 20 reads, writable conversions and reached entry-tuple consumers.

The boundary is internal/lower/records.go, recordStorageView. Its
sameRecordStorage check previously required every CompilerOptionsValue member
to share the dictionary target's storage. The input is ir.Union: a counted heap
reference holding a scalar box, array, fixed object or record. The read needs
an own string-key lookup, preserving source identity, checking representation
before following storage, and checking the returned element's declared type.
The minimal original-pair witness is buildoptions-good.a. Its selected value
is consumed as a readonly dictionary, then its array element is read. Changing
the consumer to a mutable index signature still produces the original
fixed-object/record storage refusal.

checkedDictionaryReferenceConversion carries that reference unchanged into
a readonly view. Each dictionary read uses the existing
adamic_view_dictionary_source_read. That function checks the heap kind,
uses adamic_record_is to identify record.c storage, and otherwise reads checked
fixed-object slots. It does not reinterpret fixed slots as a hash table, copy
keys, allocate another record representation, or grant a write certificate.
Invalid array, scalar, Map, function and missing receiver representations stop
at the demanded read. An absent key yields undefined only where allowed.
The reference result's fields and array elements remain checked at later reads.

The separate DictionaryReference selector retains the original descriptor
and type ID. Readiness preserves its selected contract. Descendant schemas
are registered; original unsupported member obligations remain available for
consumers outside the supported subset. Direct String consumers retain the
primitive-only selector because reference ToPrimitive is not implemented.
No checked conversion to a mutable dictionary is admitted.

Both pairs use the complete pinned TypeScript 6.0.3 declarations from
050880ce59e30b356b686bd3144efe24f875ebc8. Original descriptor field inventories
are asserted in the tests. Six controls per pair cover fixed-object and record
producers, missing keys, wrong selected element type, wrong receiver
representation, and wrong nested array element type. Every control is held to
Node, sanitized native, release native and JavaScript; successful cases also
pass leaks. Native mutants run with sanitizers and must finish safely.

Pinned negatives:

    field read failed: dictionary['name']; expected string[] | undefined, found array
    field read failed: dictionary['name']; expected string[] | undefined, found number
    element read failed: values[0] expected string, found number

All have the adamic: panic: prefix, trailing newline, and exit 70.
Skip-type and wrong-shape mutants safely normalize the unchecked value to
undefined, yielding Node's missing output instead of stopping. The transitive
mutant safely converts the actual numeric element to string representation,
yielding 42 instead of stopping. Each original pair catches all three mutants
in both backends. No compiler error or invalid memory access counts as a kill.

| Lane | Newly certified pairs / reads | Left in requested dictionary scope |
| --- | --- | --- |
| 4b | 2 / 11 | 0 / 0 |
| 4 | 2 / 11 overlap | OptionsBase 1 / 3 |
| 6 | 2 / 11 overlap | any 1 / 20 |

Counts are static candidate reads, not measured production tsc reachability.
The lanes overlap and must not be added. any remains excluded because the
subset has no sound erased-value contract. Tuple consumers remain excluded:
entries-fixed-array-{good,wrong,empty}.a and
entries-producer-array-{good,wrong,empty}.a require their producer-specific
ownership and per-position consumption contracts. This checkpoint does not
certify every reference consumer, mutation, or ToPrimitive operation.

Commands:

    ADAMIC_BRAND_ORIGINAL_DECLS=/tmp/views-dictionary-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewDictionaryJoint$|^TestCheckedViewDictionaryJointCounts$' -count=1 -v
    ADAMIC_BRAND_ORIGINAL_DECLS=/tmp/views-dictionary-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewOriginalDictionaryPrimitiveComponents$|^TestCheckedViewDictionaryJoint$' -count=1 -v
    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewDictionaryPropagation$|^TestCheckedViewDictionaryCompilerPaths$' -count=1 -v
    go test ./internal/lower ./internal/ir -run 'TestPrimitiveDictionary' -count=1
    go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts
    ADAMIC_BRAND_ORIGINAL_DECLS=/tmp/views-dictionary-declarations go test ./internal/oracle -run '^TestCheckedViewDictionaryJointCounts$' -count=1 -args -update-counts

First-lane controls and count verification: exit 0, 17.925s. Original primitive
regressions and joint controls: exit 0, 38.847s. Transitive baseline: exit 0,
5.622s; repeated transitive cases pass in the first-lane regression log.
Primitive certificate proofs: lower .492s, IR .009s, exit 0.
The global count refresh failed outside this unit at FS evaluated option
literals, process.exit values and other fixtures (global-counts.log). It left
counts.md unchanged. Scoped declaration-backed fixture counts were refreshed
in joint/counts.md and verified. The new fixtures are outside the global
fixture registry; no registry in a protected file was changed.
No whole package or full gate was run.

Setup used GOPROXY=https://proxy.golang.org|direct, bash cloud/setup.sh, then
/workspace/adamic-tools/env.sh. Ready lines: Node .061s, Go .093s, clang .503s,
markdown 1.942s, submodules 308.155s, Go build 482.968s, test binaries deferred
483.060s, build cache warm 483.062s, done 483.093s. nproc=5; CPU quota is four.
No lane branch was merged, no protected implementation file was edited and no
cohere code was copied. Delivery stays on codex/views-dictionary-joint, based
on 1771221bd232bd08618669e587e4566be8c252b9.
