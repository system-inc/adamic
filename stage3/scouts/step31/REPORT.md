Built: a stable real-binder Node dump, three .a witnesses, source mutants and a native-ready process comparator.
Commits: main 45487a809f89885a3fc651cd590e7dabf31362dc; branch codex/step31-scout; delivery SHA is in the final dispatch.
Observed: 301 acceptance and 6262 selected upstream binder comparisons pass; fresh main own-file meter is 56/79, whole-program 2/79, lowering 0/79.
Mutants: three actual binder bodies, three API seams, three process outputs and four research fields are caught independently.
Not covered: native execution/linking, a 78/78 meter, the full upstream suite, complete checker diagnostics coverage or full emitter differential coverage.

Measurements use unmodified main's compiler and main's apply pipeline, TypeScript
v6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. The fresh meter is
20261008T164100Z.xFfE4V. Both compiler and adapted-main pins are 45487a8.
The area measurement was also completed by the required meter command, but this
scout's conclusions use main only. All new repository files are under this scout.
No production compiler, adaptation, fixture registry or oracle counts file changed.

The brief's 78/78 is a roadmap target, not an observed main result. Main now has
79 source roots, including hostErrors.ts, plus three JSON inputs excluded from
the source denominator. The last checked-in main meter had 54/79; this fresh
compiler reports 56/79. Binder and emitter each have zero own-file diagnostics.
Checker has 89. Each of the three loaded root programs fails the checker, so
ordinary lowering is blocked for all three. The all-roots checker is also red.
An own-file green row does not prove that file's runtime dependency closure builds.

| Piece | Own-file diagnostics | Root whole-program | Site-attributed latent Refused | Site-attributed latent NotYet |
|---|---:|---|---:|---:|
| binder.ts | 0 | fail | 389 | 6 |
| checker.ts | 89 | fail | 26 | 9 |
| emitter.ts | 0 | fail | 445 | 17 |

The latent main total is 5,131 Refused, 1,468 NotYet, three panics and three
SkippedDependency sites. All are measured on a checker-rejected program.
Independent-unit attempts stop at their first lowering failure. Diagnosed
function bodies are skipped, and final ownership/backend passes are absent.
The counts cannot establish exhaustive blockers or a working native compiler.

The largest exact families, with an actual adapted-source location for each:

| File | Family | Unique sites | First location |
|---|---|---:|---|
| binder.ts | unchecked runtime cast | 176 | binder.ts:359:29 |
| binder.ts | open numeric enum used as a literal object tag | 104 | binder.ts:375:41 |
| binder.ts | non-null assertion | 61 | binder.ts:594:16 |
| binder.ts | FlowNode viewed as optional FlowNode with incompatible writable node field | 14 | binder.ts:1034:95 |
| checker.ts | unchecked runtime cast in eligible outer declarations | 13 | checker.ts:1155:14 |
| checker.ts | numeric prefix unary expression, NotYet | 3 | checker.ts:1470:9 |
| checker.ts | namespace refusal / ModuleDeclaration NotYet | 2 each | checker.ts:54223:1 |
| checker.ts | detached trackSymbol method | 2 | checker.ts:54345:33 |
| emitter.ts | unchecked runtime cast | 251 | emitter.ts:973:47 |
| emitter.ts | open numeric enum used as a literal object tag | 57 | emitter.ts:454:66 |
| emitter.ts | non-null assertion | 48 | emitter.ts:507:21 |
| emitter.ts | detached parenthesizeExpressionForDisallowedComma method | 12 | emitter.ts:2233:199 |

Every exact reason and location, the 89 checker diagnostics, and each closure's
full-file ledger are in evidence/research.json.gz. Findings encountered while
attempting emitter but located in core belong to core; the accounting audit
specifically rejects a one-site binder attribution mutant.

A raw import of binder/checker/emitter goes through _namespaces/ts.ts and the
performance barrel; checker also imports the moduleSpecifiers namespace, and
emitter imports the ts namespace value. This preserves the broad evaluation
cycle. Stock-symbol import resolution and whole-function declaration gathering
are different measures, both recorded in evidence/entrypoints-and-imports.json.gz
and the three *-slice.json.gz manifests. The existing slice tool retains exact
function bytes, selects namespace members, rewrites imports and retains original
ordered evaluation imports. Its byte/import audits all pass.

| Entry closure | Code declarations | Code files | Copied-span lines including facades | Evaluation modules | Whole-file own-green / code files |
|---|---:|---:|---:|---:|---:|
| binder: bindSourceFile | 1407 | 21 | 26463 | 79 | 16/21 |
| checker: createTypeChecker | 3390 | 39 | 124651 | 79 | 31/39 |
| emitter: emitFiles and createPrinter | 2573 | 35 | 59347 | 79 | 26/35 |

Code-file lists and every original copied span are in the manifests. Binder's
closure includes parser, scanner, allocator/type definitions, core, utilities,
Debug, diagnostic messages, tracing/performance and checker ID helpers. It does
not reach createTypeChecker. Emitter also does not reach createTypeChecker;
it receives an EmitResolver from its caller. Retaining checker.ts as a code file
for getNodeId is not evidence that the whole checker body is needed.
Full-file latent ledgers for these closures count binder 2,678 Refused/1,011
NotYet, checker 4,077/1,225, and emitter 3,486/1,181. These include declarations
not reached in the slice and are deliberately not called slice-lowering counts.

No separately named hidden-source ranking artifact was found on this main.
This unit makes that boundary reviewable by ranking the latent census's skipped
bodies by actual source bytes/lines and marking whether each is reached.
The largest reached hidden body is createTypeChecker at checker.ts:1486:
3,097,953 bytes, 52,715 physical lines, 86 body diagnostics. Checker's small
26/9 latent headline excludes this body. Its reached resolution dependencies
also hide getLoadModuleFromTargetExportOrImport at moduleNameResolver.ts:2751
(236 lines), among smaller functions. Binder has no reached skipped top-level
body in this measurement, though its closure's full files contain skipped bodies.
Emitter reaches hidden transformNodes at transformer.ts:248 (419 lines),
createSystemWatchFunctions in sys.ts (364 lines), createSourceMapGenerator in
sourcemap.ts (324 lines), and the same 236-line module resolver function.

The historical untouched-source ranking at ef3d907 is independent corroboration
of hidden scale, not a fresh lowering count: checker has 3,612 non-boolean
conditions starting at checker.ts:1569, 2,443 nested functions starting at
checker.ts:1958, and 406 non-null assertions starting at checker.ts:1758.
Binder has 165 nested functions and 120 non-boolean conditions; emitter has
392 and 287. Main's later adapters/compiler features change these obligations.
The source census's layer labels and all historical locations remain in the JSON.
A diagnosed outer function prevents the current latent pass from proving which
of its inner sites now lowers. The next measurement should admit that body or
use a separately guarded nested-unit census, then recount before planning fixes.

The observed callable seams are concrete. binder.ts:502 bindSourceFile takes a
parsed SourceFile and CompilerOptions, mutating locals, symbols, declarations,
exports/members, flags and bindDiagnostics. It is the stable Node observer's
only semantic operation after createSourceFile. The observer never creates a
Program or checker. Source mutants exercise bindFunctionDeclaration at :3709,
bindVariableDeclarationOrBindingElement at :3648, and addDeclarationToSymbol
at :635 through the real binder, not copied reimplementations.

checker.ts:1486 createTypeChecker takes TypeCheckerHost, then its returned
getDiagnostics takes SourceFile and a cancellation token (worker at :49689).
Its constructor binds host files. Host duties include source lists, compiler
options, module resolution and project redirects, libraries and package/type
lookup. For the fixed import-free probe, empty imports/moduleAugmentations/
ambientModuleNames reproduce program.ts:3318-3361's preparation. Unsupported
host capability access throws, rather than silently returning success. The probe
reports TS2322 at UTF-16 start 29, length 5, with exact message, using neither
createProgram nor the CLI. A noLib probe is explicitly narrower than a complete
library-backed checker harness; global missing-library diagnostics are not
claimed tested by this per-file assertion.

emitter.ts:1211 createPrinter returns printFile(SourceFile), which observes
printing and comments and still prints TypeScript type annotations. It is not
JavaScript transformation coverage. emitter.ts:752 emitFiles takes EmitResolver,
EmitHost, an optional target SourceFile and script/declaration transformers.
writeFile captures output text, output path and BOM. The direct probe obtains
a real resolver from checker.getEmitResolver (checker.ts:2506), uses real
getTransformers, and emits /probe.js with its annotation erased. It compares
emitSkipped, emit diagnostics and exact text. Thus Node can exercise the emitter
without the whole Program/CLI linking, but full typed emit has checker/resolver
and transformer dependencies. Printer-only and full-emit results must stay separate.
The same probes pass on actual adapted source and stock 6.0.3 with equal bytes.

The proposed differential sequence is:

1. Binder: materialize the existing 301 acceptance projects exactly as the CLI
   driver does, plus compiler/conformance units and eventually the full virtual
   test-host corpus. Run parse/bind on fresh files/options and compare
   adamic-binder-v1 bytes, stderr and exit. Standard libraries are not required
   for binding. Add upstream JS/JSDoc, namespace/enum merges, script mode,
   computed/private names, duplicate bindings and rebind/cache tests in later
   batches. Keep option variants and multi-file units as separate projects.
2. Checker: use a deterministic in-memory TypeCheckerHost over those project
   files and pinned standard libraries. Preserve resolution modes, package
   manifests, redirected references and case/newline behavior. Compare per-file
   and global diagnostics as ordered records: virtual file, start/length,
   category/code, complete message chains and related locations. Keep parse,
   bind, checker, option and emit diagnostics distinct. Add suggestions and
   cancellation as separately requested modes. A request/response transcript
   must fail on an unsupported host operation, not substitute an empty answer.
3. Emitter: begin with printFile/printNode across actual parsed files, then
   transformed emit with real resolver responses. Compare ordered writeFile
   records including virtual output paths, BOM and all text bytes; include JS,
   declaration text, source/declaration maps, diagnostics and emitSkipped.
   The acceptance corpus was selected for noEmit: its existing goldens cannot
   judge emitted text. Record independent stock emit baselines and use upstream
   emit baselines/virtual host with their documented conventions. Exercise target,
   module, JSX, declaration, comments, helpers and source-map option combinations.

Before step 32, gather each callable entry's actual declaration closure and
build a small .a executable around it and its required runtime/dependencies.
A piece executable can link its closure without implementing the full tsc CLI,
watch/build or project orchestration. Binder can consume a parsed AST graph
prepared by the already-held parser or a checked fixture decoder; checker can
consume prepared project/source metadata and libraries through a native host;
emitter can accept a real resolver implementation. This is a proposed component
boundary, not permission to drop required dependencies or stub successful answers.

For earlier isolation, a typed graph protocol could encode AST fields, child
order, parents and bound symbol references with handles, then reconstruct them
natively. A checker snapshot must include the metadata bindSourceFile expects
and honor its already-bound-file path; compiling the constructor still brings
its binder reference into the closure. A resolver transcript can isolate emitter
printing/transformation only if every native query and argument is checked
against the recorded request, and unexpected queries fail. Label such a result
as emitter-with-recorded-resolver, not a native checker or complete compiler.
These protocols need explicit ownership/ABI decisions before implementation.
The delivered process comparator already accepts a native executable and its
request pathname, with no Node dependency in that execution slot.

Observed native admission is currently blocked. Three `adamic c` attempts rooted
at the gathered slice-entry.a each exit 1 and produce zero C bytes. Binder first
reports Debug.captureStackTrace lacking ErrorConstructor declarations; checker
also reports actual optional/indexed-read diagnostics; emitter also reports
resolver contract errors. Adding the scanner-style node:util declaration marker,
then a pinned node_modules symlink, did not repair admission: main reports TS2591
for that marker. These are scratch-only probes, not source adaptations or a
native execution result. Full logs and copied-span hashes are retained.

Actual commands, all test output redirected to files:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/step31-setup.log 2>&1
source /workspace/adamic-tools/env.sh
STAGE3_METER_RUNS=/tmp/step31-meter bash stage3/meter/twice-daily.sh > /tmp/step31-meter.log 2>&1
bash stage3/apply.sh /tmp/step31-adapted > /tmp/step31-apply.log 2>&1
export STEP31_TYPESCRIPT=/home/agent/.cache/adamic-stage3/api/node_modules/typescript/lib/typescript.js
bash stage3/scouts/step31/run.sh /tmp/step31-adapted /tmp/step31-proof > /tmp/step31-proof.log 2>&1
```

The finished one-command proof exits 0: 6,563 projects (301 acceptance plus
6,262 upstream), 120,330,103 output bytes on each side, SHA256
8c4cb86d3858a9f13788225abb7974bef6becbcbd46a87a16c8843051c18cd5f.
The acceptance-only comparison is 301/301, SHA256
3f83fbf1a9c1587e2923484f7a22911f21c3ac47382ba80dbca597af04d846aa.
The upstream selection explicitly excludes 6,182 input files for virtual-host,
variant, emit-only and other declared reasons. No full upstream npm test or
whole-package/full integration gate was run. Raw large dumps remain in /tmp;
request preparation, pins, sizes/hashes and small fixture goldens are committed.

Closure measurements used `SLICE_TYPESCRIPT=$STEP31_TYPESCRIPT bash
stage3/slice/run.sh TREE NEW_SLICE <entry selectors> --no-adapt`, with selectors
listed in the closure table, followed by `node stage3/slice/verify.cjs SLICE`.
Audits report 1,492 binder, 3,485 checker and 2,664 emitter byte-identical spans,
each with 79 ordered module import lists. Native admission used the locally
built `/tmp/step31-adamic c SLICE/slice-entry.a`; marker retries used admission.a.
Research was regenerated with research.py, then independently checked with
audit-research.py against the exact meter tree. Its four --mutant runs each
exit 1 at the specified recount assertion. The final focused rerun used `bash stage3/scouts/step31/run.sh
/tmp/step31-adapted /tmp/step31-proof-final` with output in
/tmp/step31-proof-final.log and passed with the same corpus hash. JS syntax,
Python AST, shell syntax and diff whitespace checks passed.

Setup printed Go ready 0.065s, Node ready 0.070s, clang ready 0.411s,
markdown dependencies ready 0.843s, submodules ready 12.013s, Go build ready
187.020s, cache warm 187.145s and done 187.220s. nproc=5, cgroup quota
400000/100000, 17.6 GB. Versions: Go 1.27.1, clang 20.1.8, Node 24.19.0.
The requested GOPROXY workaround was set before setup; setup succeeded.

| Mutant actually run | Catcher and observed result |
|---|---|
| bindFunctionDeclaration uses BlockScopedVariable instead of Function flags | stock-golden dump bytes; aliases/scopes differ, Node exit 0, empty stderr |
| bindVariableDeclarationOrBindingElement uses FunctionScopedVariable for block bindings | stock-golden dump bytes; scopes differs, Node exit 0, empty stderr |
| addDeclarationToSymbol clears flags | stock-golden dump bytes; all three files differ, Node exit 0, empty stderr |
| checker input changes string initializer to number | expected TS2322 assertion fails, exit 1 |
| printer removes preserved comment | printFile text assertion fails, exit 1 |
| emitter omits script transformers | annotation remains in emitted text; emit bytes assertion fails, exit 1 |
| exactly one stdout byte changes | stdout comparison only fails, exit 1 |
| stderr gains one byte | stderr comparison only fails, exit 1 |
| process exit changes 0 to 1 | exit comparison only fails, exit 1 |
| own-file total increases by one | independent meter recount fails, exit 1 |
| binder family's count increases by one | unique-site family recount fails, exit 1 |
| hidden checker's line count increases by one | source-byte extent recount fails, exit 1 |
| one extra NotYet assigned to binder | actual diagnostic-file attribution fails, exit 1 |

The .a fixtures are Node binder inputs, not native oracle registrations.
counts.md records 23/5 AST nodes/symbols for aliases, 45/12 for members and
40/8 for scopes, each with zero bind diagnostics. No native allocation counts,
leak checks or native mutant kills are claimed. Parser, checker and emitter
native implementation, JS/JSDoc breadth, full control-flow graphs, incremental
binding, all TypeScript tests and native ownership remain outside this unit.

Questions for @system_adamic, intentionally undecided:

- Does the roadmap retain the old 78-file population, or adopt main's 79 roots
  including hostErrors.ts? Which checker/entry admission profile should each
  component use while preserving Adamic's fixed options?
- What proof should admit SyntaxKind/FlowFlags-tagged AST and flow unions when
  the current open-enum-tag and unchecked-cast refusals reject them? Which
  writable field views require source adaptation rather than a language change?
- Can a per-project arena own AST/symbol/type graphs while parent, declaration,
  symbol-parent and cache edges are typed non-owning handles? Which edges escape
  into incremental caches, and what prevents use after that arena is destroyed?
  This needs a no-GC lifetime proof, not an implicit cycle collector.
- Should native component boundaries exchange typed handles/graphs, or compile
  dependencies together initially? Is emitter-with-recorded-resolver an approved
  intermediate proof boundary, with every resolver query checked?
- How should the retained cyclic barrel evaluation graph be initialized, and
  how should Node declaration capabilities such as captureStackTrace be provided
  to component roots without weakening checker options or inventing host behavior?
