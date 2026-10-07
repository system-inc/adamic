Built: native .a source validators for react-hooks/static-components, set-state-in-render and set-state-in-effect, with an isolated raw HIR bridge; eighteen wave 26 ports now have native Go agreement.
Commits: base e8ba3d5d81de4d3773c723914fccd4c76248b965; preceding reporting commit ae7bfd8cb19cc05a12b4a0eec8ddd3753067f36d; implementation d7a62ceddbbb63191d903cb160d241bc2a24fa5a.
Checks: 129 source controls, 69 findings, 77 compiler files and 287 repository files per rule match production Go under normal native execution and sanitizers; original-source Node matches all controls; previous fifteen ports re-green.
Mutants: three new analysis mutants, four reporting mutants, three no-input refusal mutants, omitted SSA, retained released handle, fifteen prior rule mutants, thirteen prior question runs, five prior handle mutants, seven ABI mutants and Node's one-byte mutant were caught.
Not covered: checker-linked JavaScript emission and shared registration, the full root gate, larger application corpora, or Go's arbitrary creation-message choice when multiple dynamic phi operands have distinct creators.

These are source analyses, not precomputed Go diagnostics. Each rule has its own
rule.a and standalone suite.a entry point. The earlier index.a classes remain
message helpers. Their legacy analyze() has no source parameters and explicitly
refuses; callers use the corresponding Rule class and run() for source analysis.
No extra rules were claimed during this unit.

The react-hir question returns raw lowering, SSA, nested function arenas, capture
edges, syntax coordinates, type alias names and type symbol names. Its memo view
also performs Cohere's memo erasure and immediate-call inlining. The native rules
make every lint decision. No Go lint rule or diagnostic is called by the bridge.
Cohere's internal package cannot be imported from Adamic, so the raw implementation
is isolated under this owned directory. COPIED_SOURCES.json hashes all 28 pinned
source inputs; extracted definitions and rewritten local imports are explicit.
The package contains 9,560 Go lines. No module configuration changed.

Only one line was added to the shared checker facts switch. The bridge question,
contract test, native adapter, raw graph package, validators, runners and validation
tools are separate files. Shared generators, harnesses and protected compiler
files were untouched. The constructor initialization workaround stays in these
owned files.

Static-components propagates creation identity through loads, stores and phis in
one reverse-postorder pass. It preserves the compilation gate, nested-unit walk,
member-tag silence and invisible back edges. Render-setter analysis translates
captures and uses native post-dominance, including useMemo precedence and shadow
resolution. Effect-setter analysis consumes the memo view, translates helper
captures, propagates ref-derived values through destructuring, and computes the
control-dependence exemption natively.

The raw graph comparison calls Cohere's original ForFunction and
ForFunctionWithoutManualMemoization independently. On 22 controls both views
match in 1,046,120 bytes, normally and under sanitizers. Omitting initial SSA
construction compiles and exits 0 with empty stderr but changes those bytes.
Serialized graph data survives checker release. A new query on that handle
panics with exit 70; retaining the registry entry produces exit 0 and is caught.
The question contract checks malformed views, wrong node kinds, raw type fields
and UTF-8 framing. Pinned Cohere is 715ba94f3608a6500086b1076ce5cb7e51b836db;
compiler corpus is TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8.

| Source validator | Controls | Findings | Compiler/repository findings |
| --- | ---: | ---: | ---: |
| static-components | 23 | 14 | 0 / 0 |
| set-state-in-render | 53 | 25 | 0 / 0 |
| set-state-in-effect | 53 | 30 | 0 / 0 |

The independent Go oracles invoke the unmodified production registry rules.
Comparison includes ranges, rule/message identifiers, full descriptions, every
fix and every suggestion; these three rules propose none. Controls are extracted
from pinned Go fixture literals and include typed and untyped render inputs,
renamed imports, memo callbacks, helper chains, ref data and control dependence,
nested compilation units, Unicode and comment-trimmed spans. This is not a claim
that every generated fixture in Cohere's tests was extracted.

Three source mutants compile, exit 0 and have empty stderr:
static-components disables assignment-taint propagation; render and effect change
the accepted Dispatch alias to DispatchMutant. Only the production diagnostic
byte comparison catches them. The four reporter mutants change the effect,
render and static message identifiers, or static's empty-creator branch. Each is
caught by Go diagnostic bytes. The memo message is byte-compared but does not
have its own reporter-identifier mutant. Removing each of three no-input refusals exits 0
and fails the required exit-70 contract. Detailed commands and outputs are saved.

Source Node executes the original .a validators using recorded raw checker
responses; the snapshots are dependencies, not findings. It matches all 129
controls. Native sanitizer runs use ASan/UBSan/LSan and have empty stderr.
The full bridge regression passes 100 C queries, released/zero/distinct handles,
162 positions in 3,261 exact bytes and seven mutants: input length, output length,
retained released handle, wrong source position, removed link opt-in, omitted
output free and unowned region allocation. Sanitizer, byte and refusal checks
catch each as recorded in bridge.log.gz.

The five prior suites pass again with 218 controls and 225 findings, their normal
and sanitizer corpus comparisons, fifteen rule mutants, thirteen raw question
mutant runs and five released-registry mutants. Their complete logs name every
mutation and its catcher. The filtered Node/native/JavaScript regression passes
fifteen fixtures and its one-byte mutant. Worker caches were used there; no
integration uncached gate is claimed. Vet, owned Go formatting and diff checks
pass with empty output.

Three alternating quiet rounds compare full bytes on every timing run:

| Rule | Compiler native / Go median | Repository native / Go median |
| --- | ---: | ---: |
| static-components | 5.907155 / 7.080944 s | 0.870340 / 0.289425 s |
| set-state-in-render | 5.674237 / 6.869089 s | 0.757792 / 0.295794 s |
| set-state-in-effect | 6.641709 / 5.740106 s | 0.790862 / 0.226920 s |

These are process times, including program load and serialization. They measure
three separate rule runners, not a combined eighteen-rule lint. Native is slower
on the repository corpus and for effect linting on the compiler corpus.

Run the owned validate_hir.py, validate_static.py, validate_state.py,
validate_node.py, validate_partial.py and validate_cost.py with the arguments
retained in source_validation/**/runs.json.gz. The five previous suite commands
are in source_validation/previous-fifteen/runs.json.gz. Additional commands were
go test ./bridge/tsgo/... -count=1 -v -timeout 15m, the separate
TestReactHIRQuestionContract, go vet ./..., owned gofmt -l and git diff --check.
Test output always went directly to files. The existing cloud setup was reused:
Go ready 0s, clang 1s, Node 1s, submodules 1s, warm/total 135s; nproc is 5.

The JavaScript source-analysis comparison remains blocked by the shared CLI:
`adamic js static_suite.a` exits 1 with `Adamic 0.1 refuses an unlinked
typescript-go library call; build with --tsgo <checker archive>`. The same happens
for both setter suites. cmd/adamic/main.go currently supports checker linkage
only for native builds and rejects checker use in JavaScript. The original .a
source runs on Node; only emitted checker-linked JavaScript is missing. The
shared compiler was not changed. Reporting-only JavaScript still passes.
Shared suite registration is left to the worker owning the generator/harness.

An early unconditional whole-file graph run was stopped before memory exhaustion;
source runners now apply their native compilation gate before requesting graphs.
Large admitted functions still serialize nested graphs, and larger application
corpora were not measured. Go iterates phi operand maps arbitrarily; when two
paths have different creation spellings, its message can choose either. The
native serialized order is deterministic. Such a case is outside these controls
and frozen corpora; universal byte agreement on that nondeterministic message
is not claimed. No full repository gate or main/area push was performed.

source_validation preserves deterministic gzip outputs, source and diagnostic
hashes, fixture inputs, every command and elapsed observation. Earlier partial
reporting artifacts remain historical evidence. The native source port supersedes
the old raw-HIR engineering blocker documented in the claim.
