# Parser objects and the Program region

This scout follows the October 8 ruling: complete SourceFiles and their parsed
nodes belong to the CLI Program region by type. Detached and synthetic graphs
need the graph-region and checked-weak lifetime rules. The runtime brief was read
from `runtime/cycles-decision`, commit
`d4108344c3ea62ee0208cb26ef6f9eae20af01db`, [docs/cycles-decision.md](https://github.com/system-inc/adamic/blob/d4108344c3ea62ee0208cb26ef6f9eae20af01db/docs/cycles-decision.md).
It calls out explicit keepers for detached children and escaping region owners.

## What the meter observes

Stock TypeScript 6.0.3 source is pinned to
`050880ce59e30b356b686bd3144efe24f875ebc8`. prepare.cjs copies it outside Adamic,
parses the allocation owners with stock TypeScript's AST, wraps 15 actual `new`
expressions, and records parser/factory function entry/exit paths. An independent
constructor-entry audit covers compiler Node/Token/Identifier constructors and
service NodeObject/TokenOrIdentifierObject constructors. It asserts exact object
identity equality with the recorded allocation population at each boundary.
No upstream source is committed. The scratch build uses upstream's locked esbuild
as a Node bundler, with source maps; an uninstrumented bundle is the control.
The standard library files come from the stock npm TypeScript 6.0.3 package.

“Parser” means creation while a function in the actual Parser namespace is active. “Synthetic” means a
factory allocation outside that parsing phase, not a count of NodeFlags.Synthesized bits.
Missing/recovery nodes and abandoned speculative parses remain parser-created.
NodeArray, ordinary arrays, symbols, types, flow records and strings are outside
this node-only allocation census.

The scanner workload parses every regular src/compiler file, including JSON
and generated diagnostics, as the scanner driver's corpus specifies. It uses
createSourceFile with parent fixing enabled and retains all resulting SourceFiles
until its end. scanner-inputs.json records all 81 paths and byte hashes. Its
structural AST digest matches the uninstrumented source bundle.

The acceptance workload reuses corpus.py's exact materialization and options:
300 selection.json cases plus the three-source tiny project. Each is a fresh
createProgram/createCompilerHost followed by getPreEmitDiagnostics, in one Node
process. All default libraries loaded by that project are counted; repeated
standard-library parses are separate allocations, not deduplicated counts.
Config parsing is included: each project creates a JSON SourceFile that is not a
root of its Program. That is counted as parser-created and unreachable at end.
Per-project end means after checking, with that Program's getSourceFiles roots
held, before the next Program. This measures the full noEmit check, not just
user-file parsing or CLI startup. projects.json preserves root counts, AST and
diagnostic hashes, golden results and allocation counts for all 301 projects.

Two reachability measures stay separate. Syntax reachability walks forEachChild
plus attached JSDoc. Graph reachability recursively walks enumerable own object
fields from every SourceFile, including node/symbol/diagnostic links, avoiding
cycles. It does not inspect closure environments, weak-map internals, prototype
properties or arbitrary host keepers. The requested literal detached union is
`parent === undefined OR not graph reachable`; another count exempts rooted
SourceFiles from the parentless half. A nonempty parent is not proof of reachability.

The tracker strongly holds every newly allocated node until this observation,
including garbage from speculation. Counts therefore describe *allocations and
end-state attachment*, not nodes surviving V8 GC or live native heap bytes.
Instrumentation retains additional objects and changes timing. Timings are meter
wall time, including controls and traversal, not stock-tsc performance claims.

## Results

| Workload | Parser-created | Factory outside parsing | No parent | Not reachable from SourceFiles | Literal detached union | Root-exempt detached union |
|---|---:|---:|---:|---:|---:|---:|
| 81 scanner inputs | 973,000 | 0 | 18,797 | 18,716 | 18,797 | 18,716 |
| 301 acceptance projects | 41,410,722 | 4,409 | 1,263,474 | 1,256,718 | 1,263,584 | 1,256,718 |

All 301 goldens and control AST/diagnostic comparisons pass. Scanner measurement
wall time was 21.245s; acceptance was 706.725s;
the final synthetic caller pass was 152.99s. Its counts
agree independently with the corrected full traces for every project.
Phase review reclassified 18 creations from parser work to outside-parser factory work.

There were 7,167 acceptance SourceFiles created: 6,866 Program roots plus the
301 configuration JSON SourceFiles. The latter account for 301 unreachable
SourceFile objects; they are not completed Program roots. SourceFile root
parentlessness is exempted only in the second union, never hidden in raw counts.

## Paths and interpretation

[code-paths.md](code-paths.md) lists creation paths and source-mapped checker call
sites with counts. paths.json.gz preserves every full path with a detached
candidate or synthetic allocation, including matched attached creations at that
same path. The supplemental synthetic-only pass must reproduce every project's
synthetic count from the main census before summarization succeeds. Review tightened phase classification from any
parser.ts function to the actual Parser namespace: the original complete traces
were reclassified without rerunning allocation, and the final independently
instrumented synthetic pass verifies each corrected count. summary.json records
every reclassification. This avoids labeling checker factory calls through
visitNode/visitNodes as parsing. sites.json
lists the 15 exact allocation expressions.

Observation: a SourceFile root has no parent; speculation can allocate an object
that no completed SourceFile owns; a factory result may have a parent without a
forward SourceFile path. None of these alone proves a node is outside the ruled
Program region. Membership is inferred by type and lifetime, not by testing the
parent field or reclaiming nodes when this traversal cannot find them.

Inference under the ruling: completed ordinary trees use Program membership.
Detached/synthetic objects need an independently retained owner of the containing
region, an appropriate graph region, or checked weak backlinks with a keeper.
A graph region can retain an overwritten-away member until teardown. An escaping
interior object cannot be counted independently while its Program storage is
freed. This scout measures the cases to handle; it does not implement membership
inference, graph allocation, promotion, teardown, or a new weak policy.

## Fixtures and checks

Three .a ownership mirrors are held to direct source Node, backend Node, and
counted ASan/UBSan native runs. They use today's explicit Weak for backlinks
while a strong keeper lives. This makes them executable without pretending main
has the new Program allocator or inferred interior pointers. Runtime strings are
built with repeat; allocation/free counters are balanced. counts.md is their
local registry inside this unit's territory, not the central oracle registry.

| Fixture | Witness | Mutant | Catcher |
|---|---|---|---|
| parsed.a | Program owns a SourceFile and its child; parent identity holds | Drop the parsed parent edge | stdout differs on source Node, native, backend Node |
| synthetic.a | Factory node is absent from file children, with an independent parent keeper | Erase synthesized provenance | stdout differs on all three |
| detached.a | Child removed from file children still has a parent and retained payload | Leave the child attached | stdout differs on all three |

tracker-check.cjs additionally checks root exemption, an unreachable synthetic
node and factory creation inside a tree-walk callback, then skips an allocation seam: the independent constructor audit rejects
that mutant. A scratch mutant that omits graph traversal fails the rooted-tree
reachability assertion. A mutant that labels every parser.ts helper as parsing
fails the factory-in-tree-walk phase assertion. Mutation source files and binaries remain in scratch only. The
fixture runner requires every mutant to typecheck, compile and finish normally,
so compiler warnings and sanitizer crashes cannot masquerade as stdout kills.
Fixture checking keeps strict, exactOptionalPropertyTypes,
noUncheckedIndexedAccess, verbatimModuleSyntax and erasableSyntaxOnly enabled.

## Reproduce

Use a new external tree at the pin and new scratch output directories. Install
its locked dependencies and generate diagnostics using upstream's script. Set
STAGE3_TYPESCRIPT to stock npm TypeScript 6.0.3's lib/typescript.js. The commands
below assume tree, scratch, projects, evidence and callers are absolute paths:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/step24-region-setup.log 2>&1
source /workspace/adamic-tools/env.sh
(cd "$tree" && npm ci && node scripts/processDiagnosticMessages.mjs src/compiler/diagnosticMessages.json) > /tmp/step24-region-source.log 2>&1
node stage3/scouts/step24/region/prepare.cjs "$tree" "$scratch" > /tmp/step24-region-prepare.log 2>&1
python3 stage3/scouts/step24/region/materialize.py "$projects" > /tmp/step24-region-materialize.log 2>&1
node --max-old-space-size=12288 stage3/scouts/step24/region/measure.cjs "$scratch" "$tree" "$projects" "$evidence" > /tmp/step24-region-measure.log 2>&1
node --enable-source-maps --max-old-space-size=8192 stage3/scouts/step24/region/callers.cjs "$scratch" "$projects" "$callers" > /tmp/step24-region-callers.log 2>&1
node stage3/scouts/step24/region/summarize.cjs "$evidence" "$callers" > /tmp/step24-region-summary.log 2>&1
node stage3/scouts/step24/region/tracker-check.cjs "$scratch" > /tmp/step24-region-tracker-check.log 2>&1
go build -o /tmp/step24-region-adamic ./cmd/adamic > /tmp/step24-region-build.log 2>&1
node stage3/scouts/step24/region/check.cjs /tmp/step24-region-adamic > /tmp/step24-region-check.log 2>&1
```

No whole package tests, full gate, emit suite, watch/service versions, multiple
sharing Programs or native stock-tsc compilation were run. The ordinary
acceptance diagnostics, control ASTs and requested ownership fixtures were run.
No runtime file or existing corpus/golden file was edited.

Base: origin/main `45487a809f89885a3fc651cd590e7dabf31362dc`. Setup timing lines:
Node 0.019s, Go 0.020s, submodules 0.053s, markdown skip step 0.007s, markdown
ready 0.067s, clang 0.136s, build 8.457s, tests deferred 8.558s, cache warm
8.559s, done 8.585s. nproc=5, cpu.max=`400000 100000` (four CPUs).
Setup and measurements sourced `/workspace/adamic-tools/env.sh`.
