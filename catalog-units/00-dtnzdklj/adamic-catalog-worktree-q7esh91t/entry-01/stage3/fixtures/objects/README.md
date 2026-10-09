# Remaining object and syntax fixtures

25 source-derived `.a` programs, stock TypeScript 6.0.3 source inventory, exact
Node observations, and current-main stage-0 observations. No compiler changes or
adapted-tree patches are made here. These are programs for judging the features
and reviewing the adaptations, not a claim that native tsc is working.

Adamic base: `ef3d907ecdc4c771b016f7d9c52372def057a340`.
TypeScript tag: `v6.0.3`, commit `050880ce59e30b356b686bd3144efe24f875ebc8`.
Census read from `origin/codex/tsc-census` at
`429c1177f0130f785c19cf590d1860513b2ddbfc`; pipeline README read from
`origin/codex/stage3-base` at `8728405135d329efc12c837a7a6c293234abbe1c`.
The relevant census inputs are `REPORT.md`, `inventory.cjs`, `data/sites.json`,
`data/shape_additions.json`, `data/function_expandos.json`, and `data/files.json`.
The function-expando ledger is empty. All 77 pinned source hashes are audited.

## Counts and decisions

`inventory.cjs` uses stock TypeScript's parser and checker. `sites.json` gives
every location, including original line text. Counts exclude generated sources
and include original compiler subdirectories and barrels. A count is a node or
write, not a runtime frequency. These are source observations, not latent
lowering outcomes: the census's original tree never passed stage 0's checker.

| Form | Sites | Decision and reason |
| --- | ---: | --- |
| `a spread after the first field` | 17 | Adaptation. The later operand can overwrite a field with an undisclosed key and value. Use reviewed explicit fields or a closed-key merge with the original left-to-right snapshots. Moving the spread first is not a general repair. |
| Post-creation ledger | 16: 13 object, 3 array | Declared object slots are a language representation issue, not proof of illegal expansion. All 16 names exist in the receiver types. Keep optional slots absent until written. Adapt the polling array to a wrapper containing items and metadata; update all queue consumers and preserve alias identity. Do not claim that adaptation has been performed here. |
| `a definite assignment assertion !` | 12 | Adaptation. All 12 are local-variable markers, not class fields. Initialize only where control flow proves assignment, or represent absence as `T \| undefined`. Arbitrary zero/empty defaults change JS semantics. Fixture 15 actually returns undefined in two purportedly required slots. |
| Classes | 9: 8 declarations, 1 expression | Language features: constructors, initialized fields, readonly/private/static members, implements, nominal instances and inheritance. The single superclass relationship is the identifier multimap. Do not invent initialization for `declare` members: the allocator/brand declaration needs a separately proven adaptation. Parameter properties belong to the namespace/parameter-property worker. |
| Binding-element defaults | 20 | Language feature. Evaluate defaults only for undefined, in binding order, with the original receiver. |
| Binding-element rest | 3 | Language feature. All three are array rest: container alternatives and two diagnostic-argument forms in checker.ts. Rest collects the remaining elements in order. No object binding rest was found. |
| `a destructured parameter beside a parameter with a default` | 2 | Language feature. Preserve ordered parameter initialization, including the wildcard matcher's own default and the imports helper's separate default parameter. |
| `a call through ?. (an optional call)` | 448 call-chain nodes; 72 have their own `?.` call token | Language feature. Preserve receiver binding, single evaluation and skipped arguments on a nullish path. 448 is the stock OptionalChain flag count, including calls propagated from an optional receiver; 72 is not an additional count. |
| Getters | 82 | Language feature. Property access invokes the descriptor; repeated reads may change values. 72 are node-factory getters, one is a class getter, the rest are literal getters. Keep narrowing on a captured getter result when repeatability cannot be proved. |
| Setters | 2 | Language feature. Both transformation hook setters check state before updating captured hook functions. Property assignment must call them. |
| `yield (generators)` | 16 | Language feature. Suspension requires an iterator/state-machine representation with owned live values and cancellation cleanup. An eager array rewrite would change evaluation and allocation timing. |
| `in` | 10 | Language feature. Test inherited as well as own presence. These usages include diagnostics' union discrimination and debug augmentation checks. Delegate dictionary representation to records; do not replace presence by truthiness. |
| `delete` | 2 | Dictionary adaptation/records dependency. Both delete dynamic table entries. Use the ordered dictionary's deletion operation, preserving absence and ordering. A fixed-shape field set to undefined is different. |
| `debugger` | 1 | Adaptation for a production native build: remove the debugger trap from Debug.fail, retain Error creation and throwing. This preserves the tested output without an attached debugger; debugging support is a separate host capability. |

No class declaration occurs inside a function; the stock walk found zero such
sites. Class inventory locations are checker.ts:1451 and :54329, debug.ts:826,
factory/utilities.ts:1413, semver.ts:46 and :206,
transformers/utilities.ts:389 and :441, and types.ts:4586.
The state-machine class is six constructor parameter properties, already owned
elsewhere. Version's overloaded constructor and DebugTypeMapper's declared kind
are inventoried but not instantiated by these fixtures.

## Fixture coverage

Original statements and method bodies are retained. Dependency-heavy enclosing
functions are reduced to the relevant statement regions, with smaller signatures,
local support types and dependency stubs. Drivers are new. Constants standing
for enum members use their real numeric values. This does not certify all
upstream dependency contracts: the exact reductions below are the scope.

| Fixture | Original usage and reduction |
| --- | --- |
| 01_reference_spreads | parser's type-reference directive has two conditional late spreads. Run missing/present resolution modes and print JSON to distinguish absent keys. Keep the push statement; omit unrelated pragma kinds and parsing scaffolding. |
| 02_polling_levels | sys's leading spread followed by override spread. Keep the whole helper; environment lookup and level conversion are deterministic stubs for the merged numeric levels. |
| 03_resolution_cache_spreads | module resolution combines three cache objects. Retain the merge and selected methods; omit the unused getter closure and setup/cleanup bodies. This is the mutant subject. |
| 04_trace_metadata | tracing's three metadata events, including a final cat override after spread. Retain the actual fs.writeSync statement and JSON serialization; substitute a string sink and deterministic timestamp. |
| 05_polling_array_metadata | sys's entire createPollingIntervalQueue, including both assertions and all three writes. The array remains an array, is pushed to and read through metadata. This is the hardest shape case; it is not flattened into an ordinary object. |
| 06_watch_close | sys's dynamic watchFile body closes an aliased watched object, removes it from the watched list and exposes the closure write. The same isClosed statement represents fixed-chunk watchFile:476. Host mtime and queue insertion are stubs; Date/timeout behavior belongs to host. |
| 07_synthetic_reference | the whole copyFileReferenceAsSynthetic function copies a reference then writes pos and end. Driver verifies the source is untouched. |
| 08_loop_state | all three optional state propagations in createConvertedLoopState and the references write in parseProjectReferenceConfigFile. Omit loop-AST setup and project I/O; keep the creation/write regions. |
| 10_build_options | tscBuild write and the complete absolute-path conversion function. Empty build options omit the irrelevant common-options copying loop. Keep the other function's option loop; its dynamic lookup still needs records. |
| 12_decorator_descriptor | memberInfo creation and both private method/private field chained descriptor writes, then the entire attachFileToDiagnostic function. Decorator factory calls are stubs producing helper names. All related diagnostic inputs use the real same-file branch. |
| 14_comment_pending | the entire iterateCommentRanges function, retaining four markers, scan label and intentional newline switch fallthrough. Character-code constants and ASCII whitespace/shebang helpers are narrowed to the supplied comment inputs. |
| 15_accessor_absence | the entire getAllAccessorDeclarations function with four markers. Helper predicates use a reduced accessor shape; the dynamic getter case demonstrates actual absent second/set slots. |
| 16_version_defaults | Version.with's five defaults and its toString method unchanged. A harness constructor supports the valid numeric versions and prerelease/build strings supplied by the driver; parsing and overloads are omitted. |
| 17_container_rest | the actual alternative-container loop and array rest, with symbol lookup reduced to strings. Return ordering is best containers first, then alternatives. |
| 19_identifier_multimap | complete IdentifierNameMultiMap and complete SymbolTrackerImpl, plus the real superclass size getter and the empty cancellation exception. Map key and surrounding checker types are reduced. The tracker is the hardest class case: constructor unwraps another tracker, stores callbacks, reports diagnostics, tracks tuple state and uses optional methods. |
| 20_queue_optional_call | createQueue's optional items.slice initialization. Queue operations are omitted to isolate initialization, though current stage 0 still stops at mutable-view invariance first. |
| 22_nested_optional_calls | the entire ancestor/global-cache helper. A two-directory walker stands in for the path host; return contracts are preserved, including the callback's implicit undefined. |
| 23_factory_getters | three actual nodeFactory getter members. The binary factory stub preserves memoizeOne's identity contract for numeric keys, and prints a compact operand descriptor instead of allocating an AST. |
| 24_transformation_accessors | both actual hook getter/setter pairs and their assertions. Context setup fixes a pre-initialization state; driver sets and invokes both hooks. |
| 25_map_generator | the entire core mapIterator generator; iterable input and a lazy mapping callback. |
| 26_diagnostic_in | the entire errorOrSuggestion function; reduced location/messages and diagnostic constructors preserve both global and node paths. |
| 27_delete_substitution | onEmitNode's source-file setup and conditional table delete. Omit downstream emission and global resets; driver observes the deleted entry. |
| 28_debugger | Debug.fail's debugger, Error creation and throw. Omit V8-only captureStackTrace and keep a driver catch. |
| 29_parameter_default | the entire usesExtensionsOnImports function with actual destructured/default parameters. File-extension helpers are reduced to JS/TS/JSON inputs used in the driver. |
| 30_host_optional_method | resolutionCache's whole getModuleResolutionHost plus utilities' whole getPathsBasePath, with small host/options shapes. No filesystem work is performed. |

### All 16 post-creation ledger entries

| Census locations | Fixture |
| --- | --- |
| commandLineParser.ts:2985 | 10 |
| program.ts:4069 | 08 |
| sys.ts:273, :476 | 06 (same close/write form) |
| sys.ts:283, :284, :285 | 05 |
| transformers/declarations.ts:551, :552 | 07 |
| transformers/es2015.ts:3584, :3589, :3594 | 08 |
| transformers/esDecorators.ts:1352, :1372 | 12 (both branches) |
| tsbuildPublic.ts:336 | 10 |
| utilities.ts:8657 | 12 |

The stock checker recount matches the census ledger exactly, including object,
property, location, initial kind and whether the receiver declares the property.
It is intentionally the census's narrow direct-literal scan, not a whole-program
expando analysis through arbitrary aliases or factory calls.

## Observations, mutants and reproduction

Use the Node/tool paths printed by cloud/setup.sh. This worker used
`/workspace/adamic-tools/env.sh`, Go 1.27.1, clang 20.1.8 and Node 24.19.0.
Setup timing: Go/clang/Node/submodules ready at 1s, cache warm and done at 117s;
`nproc` 5, cgroup CPU quota 4. No setup failure.

```sh
source /workspace/adamic-tools/env.sh
CENSUS_TYPESCRIPT=/tmp/objects-api/node_modules/typescript/lib/typescript.js node stage3/fixtures/objects/inventory.cjs /tmp/objects-typescript stage3/fixtures/objects/sites.json > /tmp/objects-inventory.log 2>&1
python3 stage3/fixtures/objects/observe.py > /tmp/objects-observe.log 2>&1
python3 stage3/fixtures/objects/mutant.py > /tmp/objects-mutant.log 2>&1
python3 stage3/fixtures/objects/audit.py /tmp/objects-typescript /tmp/objects-census/stage3/census --mutants > /tmp/objects-audit.log 2>&1
```

The inventory checker is used for symbol resolution only; this run does not
assert that an unprepared upstream project passes pre-emit diagnostics.

The source and API scratch paths were created by cloning the pinned tag and
`npm install --prefix /tmp/objects-api --no-audit --no-fund typescript@6.0.3`.
The census scratch path is a git archive of the branch named above.
`observe.py` runs exactly `node --disable-warning=ExperimentalWarning
oracle/node.mjs <fixture>` and `go run ./cmd/adamic build <fixture> -o <binary>`.
No non-erasable Node runner change is needed. Compiler diagnostics retain their
text and final newline; only go run's `exit status 1` launcher trailer is removed.
`native.json` records each successful binary's stdout/stderr/exit comparison.
`manifest.json` is source provenance input, not a replacement for `status.json`.

Final observations: all 25 Node runs exit 0 with empty stderr. Stage 0 reports
13 Refused, 7 NotYet, 2 Checker and 3 Compiles. All three native binaries match
Node byte for byte: synthetic reference, memoized factory getters and
transformation accessors. No silent miscompile was observed among these builds.

Additional checks, redirected to logs:

```sh
go test ./internal/lower -run 'TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat|TestWhatZeroOneRefusesIsRefusedWithAFix' -count=1 > /tmp/objects-lower-test.log 2>&1
go test ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(class_features_accessors|class_inheritance|spread_snapshot)\.a$' -count=1 -timeout 30m -v > /tmp/objects-oracle.log 2>&1
```

Both pass: lowering 1.006s; oracle 7.511s, three named fixtures plus the existing
one-byte oracle mutant check. The focused oracle uses its ordinary cache:
7 native misses/3 hits and 7 Node misses. It includes sanitized native runs and
the JS backend. No Go package was changed; these are relevant regression checks.
Final logs are retained in `logs/`.

The plain-field mutant changes only the three spreads of fixture 03 into three
explicit field reads. Original: Refused for late spread. Mutant: Compiles; Node
and native both print `1:2:3:1\n`, stderr empty, exit 0. A check expecting the
recorded Refused outcome therefore fails. `mutant.patch`, `mutant.json` and
`mutant.py` preserve the exact rewrite, observations and rerun. An initial attempt
also exposed an unused getter's closure-cycle refusal; that unused member was
cut from both versions before the successful single-form experiment.

The audit's independent inventory mutant drops one late-spread site while
retaining recorded counts. Source-site recount catches it. Source SHA256s,
ledger equality, status schema, provenance and native-byte comparisons are also
audited. The shared fixtures_test.go is owned elsewhere and was not edited.

## Limits

The native outcomes are first blockers. The markers in fixture 14 are masked by
TS7029 fallthrough; the optional chain in 22 by TS7030 implicit return; queue
initialization by invariance; the polling array by a cast; several object writes
by truthiness, chained assignment or dictionary representation. Those results
are deliberately not relabeled as observations of the later target feature.
Fixture 15, 29 and 30 isolate a marker, a paired parameter default and an optional
call respectively. Current main already compiles the literal getter/setter cases.

All source sites are inventoried, but not every site has its own runnable
program. Not covered dynamically: regex capture-array defaults; the other two
rest sites' complete diagnostic pyramid; full Version parsing/overloaded
construction and VersionRange; DebugTypeMapper's injected kind; anonymous
SymbolLinks' any brand; the parameter-property state machine; full tracing type
descriptors with six late spreads; Unicode comment scanning; inherited-property
presence and dictionary deletion order; suspended generator cancellation.
These require other buckets' types/host/predicate/any work or larger independent
fixtures. This unit supplies representative programs and the complete location
ledger, not proof of those omitted runtime contracts.

No source adaptation was installed, no upstream TypeScript suite or complete
uncached integration gate was run, and no typed/ownership refusal census beyond
the named source forms is claimed. The adapted-tree pipeline remains a separate
integration task. Taste, nested functions, enums, namespaces/parameter properties,
cycles, records, non-null assertions/casts, predicates, host and explicit any are
owned by the other workers. Overlaps remain visible in the real extracts.
