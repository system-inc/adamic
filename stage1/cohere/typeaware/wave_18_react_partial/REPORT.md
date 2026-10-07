Built: native diagnostic reporters for three React rules; source analysis remains explicitly refused.
Commits: claim ee56c863; reporting code 57614fd8; prior blocker evidence through 7e877108.
Checks: four production findings match 2443 bytes, normal and ASan/UBSan; setup 19s, nproc 5.
Mutants: three diagnostic-ID mutations caught by Go bytes; three removed refusals caught by exit-70 checks.
Not covered: source analysis, native corpus parity, full-rule mutants, checker handles or comparable lint timing.

The owned files implement the independently testable reporting portion of
react-hooks/set-state-in-effect, react-hooks/set-state-in-render and
react-hooks/static-components. Each diagnostic is rendered in native Adamic,
including both render-state message IDs, creation-site text for static-components,
and the canonical zero-fix and zero-suggestion fields. The classes deliberately
refuse analyze() with a specific NotYet panic. They are not registered as completed
rules, and must not increase the coverage count.

The validation driver runs the unmodified production Go rules against four positive
controls: synchronous effect setter, unconditional render setter, memo-render
setter and a locally created JSX component. It then passes their span coordinates
to the native reporters. It does not pass the Go message text. Every complete
serialized finding stream matches in 2443 bytes. This proves reporting, not native
verdicts or span discovery. The static creation spelling is the fixture's known
`() => null`; discovery of that spelling from native HIR is unimplemented.

The reporter structure and independent Go oracle were adapted from the newly
published origin/codex/typeaware-wave-06 partial implementation. Its own report
also marks these rules unfinished. The current origin fetch includes 417 refs;
wave 06 and wave 29 publish evidence of the same HIR blocker. No further claims
were made. Our claim ee56c863 was already pushed before this implementation.

## Exact remaining dependency

The existing WAVE_18_REACT_BLOCKER_REPORT.md identifies the Go entry points.
All three validators depend on React HIR Lower/Construct and AsCompilationUnit,
including typed SSA values, phi operands, nested-function capture translation and
source-node identity. Effect analysis additionally needs the memo-erased/inlined
view, ref taint and ControlDominators; render analysis needs UnconditionalBlocks
and its memo callback behavior; static-components needs JSX lowering and
creation-site propagation in graph order. No such native HIR implementation was
found in the origin audit. Adding .a module loading, shared profiles or suggestion
serialization does not supply these semantics. No checker query was added: this
compiler analysis is not a missing typescript-go type-checker operation.

The native JSX probe also exits 70 with:

    adamic: panic: parser slice expected GreaterThanToken, got SlashToken at 55 in fixture.tsx

The corresponding Go positive control reports a static-components finding. An
AST approximation or a zero-finding corpus run cannot establish agreement.
Shared parser, harness, registration generator and protected compiler files were
untouched. Work stops at these dependencies under Ahra's instruction to report
other blockers and stop.

## Verification

Reproduce from the repository:

    source /workspace/adamic-tools/env.sh
    python3 stage1/cohere/typeaware/wave_18_react_partial/validate_partial.py --scratch /workspace/wave18-react-partial-validation-formatted > /tmp/wave18-react-partial-validation-formatted.log 2>&1

The final run prints four matching production findings, one byte-comparison kill
and one refusal-contract kill per rule, the parser refusal, and PARTIAL PASS.
Native and ASan/UBSan reporters exit 0 with empty stderr. Sanitizers cover reporter
execution and memory ownership only, not the missing analyses.

Each diagnostic-ID mutant builds and exits 0 with empty stderr, preserving the
number of findings; only comparison with the independent Go stream rejects it.
Exact differing bytes are in validation/mutant-results.json. Each unmodified
analyze() exits 70. Removing its panic produces exit 0 and empty output, violating
that refusal contract. The generator asserts exactly one panic was removed;
formatting exposed an earlier whitespace assumption, which was fixed before the
successful final run. These six mutants do not count as full-rule analysis mutants.

All subprocess stdout/stderr and command/exit/time records are preserved under
validation/. Transcripts are gzip-compressed without truncation; inputs and source
hashes are retained as JSON. The four Adamic files were formatted with CoHere's
native formatter through the existing private adapter, which uses a virtual .ts
path while writing the original .a file. No Adamic .ts file was created.
Python syntax validation, Go formatting and git diff --check passed.

No checker bridge is used by these reporters, so no released-handle test applies
to this partial work. No native/Go lint time ratio is reported: Go performs source
loading and rule analysis while native currently formats supplied findings.
The corpus byte comparisons and timing bar remain unmet for all three rules.
No full root gate was run for this isolated reporting portion.
