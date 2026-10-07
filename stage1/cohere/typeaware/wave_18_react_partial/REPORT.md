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

Native validator cores are now available in 76e453e7. CORE_REPORT.md records
91 prepared-HIR controls, 48 findings, both corpora, sanitizers and four core
mutants. Source lowering/gates remain missing; these are not end-to-end ports.

## Numeric listener declaration follow-up

Current main remains e8ba3d5d and is an ancestor of this branch; the fetched
own origin branch matched afffb233 before this change. The prior full oracle
and prepared-HIR checks remain applicable to unchanged algorithms.

Each of the three active claimed entry points now exposes readonly syntaxKinds
with numeric 307, the pinned parser's SyntaxKind.SourceFile. All three production
Go rules register that kind. The declarations are available for the incoming
shared driver. These entry points do not read source kind strings or refetch
a source node. Their analyze methods still refuse rather than misreport.

The shared ParseNode currently has only kind: string. Numeric source dispatch
and native source-to-HIR Lower/Construct/SSA, unit gates and memo shadowing remain
missing; implementing those would require shared parser/compiler work outside
this unit. A fresh audit examined 479 origin refs; the React source-syntax
modules found on wave 03 also expose string kinds and do not supply the missing
React HIR adapter. Prepared HIR instruction tags are distinct from SyntaxKind.
Older completed wave-18 source rules retain their string-based parser API;
this increment does not claim to migrate those rules or demonstrate faster lint.
No new claims were taken.

Commands:

    bash cloud/setup.sh > /tmp/wave18-listener-setup.log 2>&1
    source /workspace/adamic-tools/env.sh
    python3 stage1/cohere/typeaware/wave_18_react_partial/validate_partial.py --scratch /workspace/wave18-react-listeners > /tmp/wave18-react-listeners.log 2>&1

Setup printed Go 0s, clang 0s, Node 0s, submodules 0s, cache warm 28s,
done 28s; nproc 5, CPU quota 4. Partial validation exits 0: four production
findings match 2367 bytes including fixes/suggestions, normal and ASan/UBSan.
Three diagnostic-ID mutants compile and exit 0 but differ from Go bytes;
three removed-refusal mutants exit 0 instead of the required 70. The base
JSX parser still exits 70. A separate sanitized native listener probe prints
307 three times, exits 0 and has empty stderr. Changing 307 to 308 separately
in each rule compiles and exits 0 with empty stderr; all three mutations
are caught only by numeric output comparison. No source-analysis or full
gate coverage is inferred from these checks. Exact logs, generated .a probes,
mutation results and hashes are retained in validation-listeners/. Comparable
end-to-end native/Go timing remains unavailable as described in CORE_REPORT.md.

## rule.json listener manifests

Added rule.json containing numeric kinds: [307] in each of the three owned
rule directories, matching their native syntaxKinds declarations and the pinned
Go rules' SourceFile listeners. The shared kind-indexed driver is not changed.
Source dispatch remains unimplemented; these entry points refuse rather than
claiming to consume a supplied source node. No rule relevance decision reads
a source kind string in these three entry points. Prepared HIR cores consume
supplied graphs; that is still not a native source port.

Fetched 495 origin refs. Main remains e8ba3d5d, already an ancestor of own
pushed tip c5bbfee0 before this metadata-only change. No rebase was needed.
No native React source Lower/Construct/SSA adapter was found; wave 03's syntax
node still exposes a string kind. No batch-8 Diagnostic integration SHA was
provided in the instruction, and no finding model migration was attempted.
No new claims were taken, and only codex/typeaware-wave-18 is pushed.

Manifest validation parsed all three JSON files, compared them with numeric
SourceFile 307 from the pinned Go AST constants and production Go listener
registrations, and checked consistency with native declarations. For each rule,
a valid JSON mutant changing 307 to 308 was rejected only by the numeric
comparison. This is manifest verification, not a new full-rule oracle run.
Existing native algorithm, sanitizer and byte-only mutant evidence remains
unchanged in CORE_REPORT.md and validation-listeners/. No full gate or
end-to-end lint timing was rerun for this metadata-only increment.

Setup printed Go 0s, clang 1s, Node 1s, submodules 1s, cache warm 27s,
done 27s; nproc 5, CPU quota 4. Validation exited 0 and printed PASS for
three declarations and three parsed mutants. Exact logs are retained below.
