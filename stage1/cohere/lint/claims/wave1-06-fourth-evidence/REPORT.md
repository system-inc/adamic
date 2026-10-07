Built: two .a rule candidates with exact Go corpus parity; no-unnecessary-type-constraint remains blocked by distinct suggestion spans.
Commits: claim a01b4c3c pushed before code; no-unsafe candidate ef739a97; no-useless candidate d52fd841; evidence commit follows.
Commands and outputs: setup 30s/nproc 5; 89 upstream cases and 398 compiler/stage1 rows match all four executions; vet and filtered oracle PASS.
Mutants: wrong-global-type-name and omit-empty-export-fix compile/run cleanly and are killed only by byte comparison on Node, emitted JavaScript and sanitized native.
Not covered: default .a registry integration, third-rule implementation/mutant, arbitrary configurations beyond these option-free rules and full repository test gate.

# Selection and ownership

Branch codex/lint-wave1-06. Pushed existing work first: Everything up-to-date.
Fetched all origin heads without recursive submodule fetching. The audit examined
320 refs; origin/main remains ef3d907e. All 46 helper-ready entries are named in
origin claim documents. The first available syntax inventory entries are:

1. @typescript-eslint/no-unnecessary-type-constraint
2. @typescript-eslint/no-unsafe-function-type
3. @typescript-eslint/no-useless-empty-export

Syntax eligibility excludes needs_type_information and binding_only. Inventory
rules-array order is retained. Main lint source modules contain no selected name;
no origin claim contains any selected name before this claim. A separate scan of
selected rule directories on every other origin ref found no implementation.
Nested evidence is excluded from claims. audit.py uses pre-claim 68304ac0 for this
worker's own ref; selection.json records the fetched ref snapshot. Later fetched
claims can change a reproduced audit result, so the snapshot is historical proof.
Claim a01b4c3c was pushed before implementation. No PR. Earlier claims remain.

# What the two candidates decide

Each owns rule.a, messages.a, rule.json with no order, the independent upstream
Go adapter, a positive raw witness and an owned mutant.json. No Adamic .ts module
was authored. Descriptors use the directory registration contract, through a
scratch compatibility overlay; production shared files remain unchanged.

NoUnsafeFunctionType listens to TypeReference, reports only a plain identifier
Function and preserves Go's coarse whole-file suppression if an interface, type
alias, class or enum of that name appears anywhere in the tree. Values and
qualified names are silent. The suppression walk happens before visitation.
NoUselessEmptyExport examines top-level statements only. It preserves the .d.ts
gate, ordinary/import-equals distinction, default/export-equals distinction,
type-only empty clauses, side-effect re-exports and export modifiers. It reports
and safely deletes each redundant empty export without adjoining trivia.

Go cohere bodies, formatter and converging fixer are unchanged. The proposal
only lets the existing registry/test discovery read .a rule modules and owned
mutants. It remains unapplied; ../wave1-06-next-evidence/a-support.patch is the
concrete patch used by overlay.py. Shared .a integration is a prerequisite for
landing these candidates. Default registration exits 1 trying to open
rules/no-unsafe-function-type/rule.ts. This failure is recorded, not hidden.

# Parity and range coverage

The baseline registration suite with the two candidates under the overlay:

    go test -overlay="$wave06_overlay" ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree)$' -count=1 -v -timeout 10m

PASS 101.987s. Inherited/upstream corpus: 302 unique source/rule/options
combinations, 156284 identical output bytes on Go, source Node and sanitized
native. Owned selected/all witnesses: 7738 identical bytes. These inherited
comparisons do not include emitted JavaScript and rewrite captured filenames.
They are supplemented, not substituted for the stronger own checks below.

The owned template is added as a virtual Go test file through overlay.py. It
reuses the foundation build/capture/format utilities without editing their
production files. It preserves original filename basenames when replaying
upstream captures, so declaration-file exemptions remain actual test inputs.

    go test -overlay="$wave06_overlay" ./stage1/cohere/lint -run '^TestWave06(Candidates|Mutants|Throughput)$' -count=1 -v -timeout 10m

PASS 86.841s. TestWave06Candidates: 89 filename-preserving unique upstream cases
plus two positive witnesses, 30176 identical output bytes across Go cohere,
source Node, emitted JavaScript and sanitized native. This includes the upstream
span, exact message, multiple-empty-export, qualified-name, shadowing,
module-marker, declaration-file and all safe-fix witnesses. Every execution
must succeed with empty stderr; sanitizer failures are failures, not comparisons.
Native is built with Sanitize:true (ASan/UBSan and Linux leak checking).

Pinned TypeScript checkout v6.0.3:
050880ce59e30b356b686bd3144efe24f875ebc8, scratch-only at
/workspace/scratch/wave1-06-typescript-6.0.3.

    ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/wave1-06-typescript-6.0.3 go test -overlay="$wave06_overlay" ./stage1/cohere/lint -run '^TestWave06Compiler$' -count=1 -v -timeout 15m

PASS 76.308s. Both rules run separately on all 77 src/compiler .ts files and
122 stage1 .ts/.a files, including generated registry source and existing gap
inputs: 199 files, 398 file/rule rows. All four executions match 24387652 output
bytes, including findings, repairs and fixed-source text. No corpus input was
filtered for a parse refusal. Emitted JavaScript is generated by the actual
javascript backend, not copied source. The clone is not committed.

# Mutants

The checked-in final mutants:

| Rule | Mutation | What comparison catches |
|---|---|---|
| no-unsafe-function-type | Match NotFunction instead of Function in TypeReference | Missing bannedFunctionType finding on the real positive witness |
| no-useless-empty-export | Keep the correct finding but omit its fix marker | Same id, message and span, different repair and unchanged fixed source |

Both compile and run successfully on source Node, emitted JavaScript and
sanitized native, with exit 0 and no stderr. Only the independent Go output
comparison kills them. The fix-only mutant was run separately after strengthening
an earlier inverted-module-gate mutant:

    go test -overlay="$wave06_overlay" ./stage1/cohere/lint -run '^TestWave06Mutants$/^no-useless-empty-export$' -count=1 -v -timeout 10m

PASS 23.589s. fix-mutant.txt pins the first mismatch at range 24 34: Go repair
fix, mutant repair empty. candidates.txt records the earlier detector mutants;
that log is historical evidence, not the final checked-in empty-export metadata.
No mutant or implementation for no-unnecessary-type-constraint is credited.

# Findings per second

One process run per side over 77 pinned compiler files, including startup and
file reads. Both rules produce zero findings on this corpus on all three sides;
therefore findings/s is genuinely zero and is not a useful speed comparison.
Native here is optimized without sanitizers; sanitizer correctness is above.

| Rule | Go seconds / findings per second | Native seconds / findings per second | Node seconds / findings per second |
|---|---|---|---|
| no-unsafe-function-type | 0.277938 / 0.00 | 1.639600 / 0.00 | 1.206134 / 0.00 |
| no-useless-empty-export | 0.253846 / 0.00 | 1.539124 / 0.00 | 1.204082 / 0.00 |

    ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/wave1-06-typescript-6.0.3 go test -overlay="$wave06_overlay" ./stage1/cohere/lint -run '^TestWave06CompilerThroughput$' -count=1 -v -timeout 10m

PASS 13.448s, compiler-throughput.txt. A bounded positive metric, separately
labelled, repeats each owned witness 500 times, 500 findings. It includes process
startup and reads; it is not compiler-corpus throughput and is not best-of-five:

| Rule | Go findings/s | Native findings/s | Node findings/s |
|---|---:|---:|---:|
| no-unsafe-function-type | 23483.65 | 33320.01 | 2088.48 |
| no-useless-empty-export | 23408.22 | 19997.98 | 2029.65 |

# Third rule: observed contract mismatch

All three upstream sources and their test families were read. Independent Go
constraint_oracle.go.txt runs the unchanged NoUnnecessaryTypeConstraint on:

    function data<T extends any>() {}

Observed diagnostic range 14..15: T. Suggestion edit range 15..27:
space + extends any. The default Finding stores a single start/end pair and the
existing oracle explicitly panics when suggestion.Fixes[0].Range != d.Range.
Thus this is not representable faithfully by an owned listener on the retained
foundation. Reporting on the constraint instead, discarding the suggestion or
making it an unattended fix would each silently change Go behavior. No such
substitute implementation was written. The final Go probe exits 0 and asserts
the two ranges differ; constraint.txt records its output.

Completing this rule requires shared separate diagnostic/edit spans and support
in the formatter/oracle. Applying only the .a registry patch does not solve it.
CLAUDE.md says "Never edit a dispatch, oracle, corpus or copied-file list."
Those shared changes are outside this worker's stated rule-directory territory;
no scope exception has been received. The .a integration patch is concrete and
reviewable, while separate-span support remains unimplemented. Original helper
merge conflicts and missing docs/parallel-work.md remain documented in earlier
batch reports.

# Setup and bounded gate

    bash cloud/setup.sh > /tmp/wave1-06-fourth-setup.log 2>&1
    source /workspace/adamic-tools/env.sh
    nproc

Setup succeeds; nproc 5. Go 1.27.1, clang 20.1.8, Node v24.19.0. Timing:

    setup: go ready (0s)
    setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
    setup: node ready (0s)
    setup: submodules ready (0s)
    setup: build cache warm (30s)
    setup: done in 30s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB

From cohere, all three selected Go reference test families:

    go test ./internal/lint/rules/typescript -run '^(TestNoUnnecessaryTypeConstraint|TestNoUnsafeFunctionType|TestNoUselessEmptyExport)' -count=1 -v -timeout 10m

PASS 0.030s, upstream.txt. From Adamic root:

    go vet -overlay="$wave06_overlay" ./... > /tmp/wave1-06-fourth-vet.log 2>&1
    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout 10m > /tmp/wave1-06-fourth-oracle.log 2>&1

Vet exit 0 with empty log. Filtered external oracle PASS 7.150s, six probe misses,
no cached probes. Full test gate not run; default lint package registration is
known to fail on .a. No compiler core or shared production source was edited.
Every test wrote directly to a log; none was piped. Logs are copied unchanged.
Early constraint-probe attempts used an incorrect working directory, then a
relative filename, then ran without the tool environment sourced. All were
corrected; only final exit 0 is credited. An early throughput selection matched
no test before its template was appended; only the final actual run is credited.
