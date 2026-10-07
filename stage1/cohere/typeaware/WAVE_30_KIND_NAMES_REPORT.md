Built: removed obsolete numericKinds from three partial JSX declarations and checked their named kinds against pinned Go ast.Kind names.
Commits: implementation c4238fade88e87c8da1b79014b4c456e12aea64d; current main remains c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06; pushed only to codex/typeaware-wave-30.
Commands and outputs: component gate PASS 11.511s; vet PASS; setup 32s, nproc 5; fetched all origin heads, with no landing-base change.
Mutants: three wrong-name metadata mutants are rejected by pinned Go comparison; all four successful-exit behavior mutants are caught by production-Go output comparison.
Not covered: registered complete JSX visitors, bindings/stability, full findings/fixes/suggestions, compiler/repository parity or full-rule native/Go timings.

Ahra's current instruction supersedes the numeric declaration requirement.
The announced shared registry validates strings.TrimPrefix(ast.Kind.String(),
"Kind"), so the existing JSX names already have the correct spelling. The three
owned descriptors now remove numericKinds. Their private component check validates
the names directly and rejects an obsolete numeric field. For each descriptor,
a changed first listener name (SourceFile) is rejected by this comparison.
These are metadata check mutants, not source-rule behavior mutants.

No shared registry, test harness, parser, bridge or compiler was edited. The
announced registry source at origin/lint-rules/harness was inspected read-only.
It supports .a modules and handed nodes, but ab70f38d4 remains unmerged on main.
The descriptors retain explicit partial-component status and are not claimed to
satisfy the full registry's factory, hook, oracle and witness requirements.
The prior numeric parser API blocker wording is superseded: names are the
registry's input. Native JSX extraction and integration remain actual blockers.

All component output bytes again match production Go, sanitizer native and
emitted JavaScript: qualified fragments 1215 bytes, names 12037 bytes and
construction messages 148002 bytes. Existing behavior mutants exit zero with
empty stderr and differ at bytes 98 (fragment object), 22 (dash exclusion),
44006 (function remedy) and 13368 (astral escape). The unpaired-surrogate refusal
probe also passes. This change introduces no checker handles; earlier bridge
released-handle checks remain applicable to unchanged code.

After source /workspace/adamic-tools/env.sh:
ADAMIC_WAVE30_JSX_COMPONENTS=/workspace/wave-30-kindnames-components
 go test ./stage1/cohere/typeaware -run '^TestWave30JsxComponents$' -count=1 -v
passes in 11.511s. go vet ./stage1/cohere/typeaware passes with empty output.
Setup reports Go 0s, clang 1s, Node 1s, submodules 1s, warm 32s, total 32s,
nproc 5 and quota 4 cores. Output goes exclusively to logs. Exact test streams
and uncompressed SHA-256 hashes are preserved beside them in
validation-wave-30-kindnames. No new claims were made. Unchanged whole-rule
prerequisite and older corpus gates were not repeated; their latest observations
remain in the prior reports. No full repository gate was run. Gate durations
are validation costs, not native/Go rule-performance benchmarks.
