# Lint wave 1 slot 03

Branch: codex/lint-wave1-03. Main base: d090af5.

Assigned positions 7, 8 and 9 in the helper-ready ordering published in
stage1/cohere/lint/HELPERS.md on origin/codex/lint-helpers (5d13f5b).
The helper REPORT.md links to this list rather than containing it itself.

1. @typescript-eslint/no-unused-expressions
2. @typescript-eslint/unified-signatures
3. class-methods-use-this

No implementations of these rules were found in stage1/cohere/lint on main or
any fetched origin branch. Occurrences in historical frequency logs are not ports.
No rule is skipped as already ported. This claim precedes all implementation.

## Foundation blocker

Registration (48ecd93) merges into current main. Helpers (5d13f5b) then conflicts
in README.md, lint.ts, lint_test.go, main.ts, settings.ts and testdata/oracle.go
under stage1/cohere/lint. Its history also introduces the earlier monolithic
rule engine and shared tests. The attempted helper merge was aborted to keep
registration intact. Neither shared engine implementation was silently discarded.

The requested docs/parallel-work.md is absent from main and both foundation
branches. The available registration contract is docs/lint-registration.md.

These rules are reserved by this claim but not implemented or certified.
The foundation integration must be resolved before this unit can honestly
provide the requested registered ports and byte-for-byte evidence.

A second independently reproduced blocker: registry.Discover opens rule.ts,
registry.Render emits imports to rule.ts, and mutant validation accepts only
.ts modules. A temporary rule.a-only copy of the existing no-debugger rule
fails generation with exit 1 and a missing rule.ts error. The copied-port
harness also discovers only .ts files. Supporting .a rule modules requires
shared registration and harness changes outside these three rule directories.
See wave1-03-evidence/extension.log and merge.log for observed failures.

## Continuation batch 2, October 7

All preceding work was pushed before fetching origin. Fetched main:
ef3d907e. Helpers: 6769b88e. Inventory: 73ac2eb0.

The first 45 helper-ready names are mentioned in claims under
stage1/cohere/lint/claims on fetched origin branches. Position 46 is available.
After it, use inventory.json's syntax ready for AST/API adaptation queue in
its published rule order. The next two available names in that queue follow.

1. structure/tailwind-no-physical-direction
2. @next/next/no-assign-module-variable
3. @typescript-eslint/default-param-last

None is named in any fetched origin claim Markdown or implemented on fetched
main. The claim search includes report Markdown under claims conservatively.
Availability follows the user's current main-plus-claims criterion; older ports
on other branches are not silently substituted for main ports. This update is
committed and pushed before writing these three rule implementations.


### Continuation outcome

Candidate implementations were committed separately: module variable e19fd35b,
default parameter 10bab9b6, Tailwind and raw evidence bf7c6799. All are pushed.
The first two match their complete captured upstream corpora. Tailwind matches
53 of 54 cases; the remaining JSX case is an independently proven parser refusal.
All three match Go over 200 compiler/stage1 files and each has a compiling mutant
caught solely by output comparison on Node, emitted JavaScript and sanitized native.

These remain candidates: normal shared registration still requires .ts, and the
owned compatibility proposal is tested only through scratch Go overlays. The
shared-file territory exception question has not been answered. No approval is
inferred and no shared compatibility changes are applied. Full commands, raw
logs, findings/s and coverage limits are in
../rules/structure-tailwind-no-physical-direction/REPORT.md.

## Continuation batch 3

Fetched origin/main ef3d907ecdc4c771b016f7d9c52372def057a340 after pushing
all previous work. All 46 helper-ready names now occur in origin claim Markdown.
From the inventory's syntax ready for AST/API adaptation queue, the first three
names neither implemented on main nor mentioned in any origin claim are:

1. @typescript-eslint/no-unnecessary-type-constraint
2. @typescript-eslint/prefer-as-const
3. @typescript-eslint/prefer-enum-initializers

The search examined 39 distinct claim Markdown blobs, including reports under
the claims directory, across all fetched origin refs. These three are reserved
here before any implementation. Same branch and .a source requirement apply.
The existing shared .a registration limitation is still recorded; no approval
for shared edits is inferred from this continuation instruction.

### Batch 3 outcome

All three remain reserved and unported. Each needs a complete repair shape that
the shared Finding/RuleContext and Go oracle cannot represent. Minimal valid
witnesses reproduce two unexpected suggestion shape refusals and one unexpected
fix shape refusal. The independent full-shape probe succeeds; all 16 selected
upstream test functions pass. No semantic port mutant or cross-backend parity
is claimed. Evidence and setup timing are in
wave1-03-evidence/batch3/REPORT.md. No shared ownership exception was inferred.

## Existing claims implemented before another selection

All six previously unimplemented claims now have owned .a listeners, descriptors,
exact messages, upstream adapters, witnesses and compiling semantic mutants.
The original three match 354 comparable fixture/option profiles; their seven
JSX profiles are independently proven shared-parser refusals. The repair-heavy
three match 142 fixture/corner cases with complete typed repair records. Both
batches match Go over compiler/stage1 snapshots on Node, emitted JavaScript and
sanitized native. All six mutants are caught by output comparison on all paths.

The three older module/default-parameter/Tailwind candidates remain pushed; the
prior Tailwind JSX gap remains explicit. Normal .a registration and full repair
serialization are still shared-harness work. Unsupported repair-heavy registered
execution fails explicitly rather than discarding repairs. Full report and raw
evidence: ../rules/typescript-no-unnecessary-type-constraint/REPORT.md.

## Continuation batch 4

All previous owned rule work was pushed through 6c1993fa before this fetch.
All original 46 helper-ready names are claimed. From the syntax-ready inventory
queue, the first three names absent from main ports and all origin claims are:

1. no-async-promise-executor
2. no-case-declarations
3. no-compare-neg-zero

This update is pushed before writing these rules. Shared integration/JSX limits
for previous claims remain documented rather than treated as matching results.
