Built: live escaped namespace objects with staged ordered exports and checked readonly writes toward step 15.
Commit: this implementation commit follows the recorded boundary d5feab7d; delivery SHA is in the push report.
Checks: TypeScript emit on Node, both backends, native sanitizers and leak checks, focused mutants and landing sweep; logs in live-evidence.
Mutants: snapshot, second object, early member, reordered assignments and independently removed native/JavaScript write checks are caught.
Limits: global-script linking, descriptor reflection, nullable exports and several exported declaration forms remain unsupported.

## Representation and execution

An escaped or reopened namespace has one canonical object per checker symbol, initialized only when its first executable declaration block runs. An ordered record table stores boxed exported values. A hidden readonly-member table protects fixed exports. Function installation happens after its declaration statement; uninitialized variables emit no installation. Internal references, qualified references and aliases read the same table. Compound writes evaluate the receiver once. Imported aliases share identity across files.

The native runtime uses data storage rather than presenting accessors. JavaScript emits ordinary data properties with hidden metadata. Descriptor operations, freezing and object spreads in programs with escaped namespace objects are conservatively named NotYet. Nullable exports are named NotYet because staged absence and null must remain distinct. Runtime enum, class, nested namespace and generic function exports are named NotYet. Existing strong ownership-cycle checks still refuse a self-containing namespace. The __proto__ export is named NotYet because TypeScript's property assignment invokes the prototype setter.

No protected emitter, lower.go, native.go or oracle_test.go implementation was edited. Hooks are in namespace lowering, reads and assignments, IR field metadata, JavaScript field helpers and native record/view field helpers. Existing nonescaping single-block namespaces retain their previous lowering.

## Oracle and fixtures

`typescript-node.mjs` executes TypeScript 6.0.3 createProgram emission, including const enum erasure, on Node 24.19.0. ADAMIC_TYPESCRIPT_RUNTIME can select the installed package; otherwise the runner resolves the pinned stage3/api package or its home-directory cache. On a fresh machine, npm ci --prefix stage3/api installs the repository-pinned oracle dependency. The source is unchanged; no TypeScript or cohere implementation was copied. Each fitting witness is compared to this emitter's stdout, stderr and exit status in both backends. Native binaries use sanitizers and leak checks.

Stored, passed and returned witnesses now pass. Identity, staged initialization, two merged blocks, assignment order, structural receivers, receiver effects, inherited membership, imported aliases and reference-valued readonly views have separate witnesses. `10_tracing_escape.a` now compiles and its recorded fixture outcome changes from NotYet to Compiles. The generic Node runner also covers fitting witnesses that do not depend on const enum erasure. The erased-block witness uses the dedicated TypeScript-emitter oracle and main's existing additionalFixtureCounts hook, because the strip transform incorrectly preserves const enums. This is an oracle-adapter distinction, not a program divergence. Its strip transform mishandles an unqualified cross-block name in the initial merged witness; the final witness uses qualified accesses and is also held to the actual TypeScript emitter.

Direct const mutation is TS2540. Contrary to the ruling's example diagnostic, TypeScript accepts qualified namespace-function property replacement and its emitted program prints 2. Adamic explicitly refuses replacing this fixed member, as the ruling requires. Wider-view const mutation prints 2 on Node but terminates in both backends with exit 70 and `adamic: panic: namespace write failed: value is readonly`. This divergence is recorded in ruled-divergences.json. Shared oracle registry integration awaits compiler/per-backend-stops, which was absent on the starting base; the dedicated oracle test pins both backends now.

Two executable global-script namespace declarations in separate files are not demonstrated: the current loader forces source files to be modules, and TypeScript then gives them distinct symbols. The cross-file fixture imports one exported namespace containing two merged blocks and proves shared identity and writes through imports. A question proposing preservation of module semantics versus adding global-script linking was sent; no linker policy was assumed.

## Mutants

Three source snapshots produce their own clean Node-matching output but disagree with the original live program. Four independent IR mutants create a second object, install a member early, reorder two merged-block assignments and remove readonly installation metadata. The first three disagree with the original Node output; the fourth exposes Node's clean overwrite instead of the production terminal check. Separate Go overlays remove the actual native record-write check and JavaScript field-write check independently: TestNamespaceReadonlyWiderViewStop fails at the pinned write because each mutant exits 0 and prints 2 instead of exit 70. Removing the erased-block creation guard independently makes the erased-block witness fail with object instead of undefined before runtime initialization. Existing namespace semantic mutants (wrong scoped function, wrong scoped constant, wrong namespace enum, wrong body order, wrong exported state, wrong debug initialization, drop returned assignment, wrong factory binding) are caught by Node comparisons. State mutants lose assignment and lose hoisting are caught by Node, and skip ready check is caught by the checked JavaScript exit comparison.

## Replay and remaining wall

The reduced tracing assignment originating at tracing.ts:95:19 now passes TypeScript emit and both backends. The unchanged stock project at 050880ce59e30b356b686bd3144efe24f875ebc8 does not reach that statement: its first loader diagnostic is src/compiler/binder.ts:1109:17, TS2412, assigning undefined to FlowNode under exactOptionalPropertyTypes. This is a named stop outside tracing, not evidence that the full stock compiler lowers.

## Environment and verification

GOPROXY was https://proxy.golang.org|direct. cloud/setup.sh reported Node ready 0.019s, Go ready 0.022s, markdown dependencies ready 0.066s, submodules ready 0.067s, clang ready 0.150s, Go build ready 31.679s, test binaries deferred 31.866s, build cache warm 31.868s and done 31.896s. nproc was 5 with CPU quota 4.

Build and vet ./internal/...; stage1 Gap|Gaps|Probes; stage3 fixtures; focused namespace lowering, own oracle tests, existing namespace controls and mutants; changed-file a-check; and TestCountsAreRecorded are recorded separately. No full repository gate was run. An earlier stage3 attempt failed because the filesystem filled with rebuildable Go cache objects. Removing old cache artifacts restored space, and the same sweep passed. Test output was written directly to logs.

## Integration conflicts

Current main 6f933fa04292441026791bbf3143bdf2dbaa3f02 was merged before delivery. Two conflicts arose when restoring the implementation: object.c includes retained both main's view_unions_mixed.h and namespace.h; namespaces_test.go retained main's parallel scheduling while changing tracing escape from an expected NotYet to successful lowering. No implementation from another unlanded worker branch was merged.

## Counts rows

Fourteen rows are added to counts.md, with no existing namespace control row changed. Stored, passed and returned register the formerly refused live escapes. Identity records the canonical object comparison; staged records before/after keys and membership; merged records two reopened blocks; structural-view records an ordinary receiver passed through the namespace shape; order records assignment-order enumeration; effects records once-only receiver evaluation; prototype-presence records inherited membership versus own membership; imports/main records cross-module alias identity and writes; wider-reference records protected reference fields and closures; 10_tracing_escape records the newly admitted roadmap witness; erased-block records omission of type-only and const-enum-only blocks. The six numeric columns are generated IR/check counts, not revealed bytes. Negative/refused witnesses and mutants are not counted as fitting programs.

The final focused namespace oracle run passed in 33.777s; stage3 fixtures passed in 19.565s; counts regeneration passed in 51.352s. The changed-file a-check covered 24 .a paths. The environment restarted during an earlier sweep, interrupting its remaining processes; those checks and the actual runtime mutants were resumed and completed. Raw TypeScript-emitter observations for 17 programs are in live-evidence/typescript-node-observations.json. Terminal exit-70 cases use the existing _exit policy; leak checks cover successful fitting programs and successful semantic mutants.
