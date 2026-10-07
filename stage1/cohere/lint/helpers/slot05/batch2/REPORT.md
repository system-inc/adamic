Built two retained .a helpers: Tailwind calleeValues and variableValues; the dispatcher duplicate is withdrawn.
Commits: 4550ef38776e040e97b8a427bf0d37621d8db8e5; reservations f5028c4 and 7b33442 were pushed before code; the original three helpers were already pushed at 72358fc/77b2029.
Checks: retained package PASS in 107.151s; 1,185 distinct Go verdicts matched Node source and sanitized native; vet, types, formatting and six filtered oracle fixtures passed.
Mutants: callee kind inversion, dropped later arguments, and variable match inversion all compiled, ran cleanly and differed from Go; the withdrawn dispatch mutant also failed.
Not covered: third continuation helper, full rule diagnostics/fixes, emitted-JavaScript comparison, production adapters/regex/collection, dynamically assembled/external fixtures, and full repository gate.

## Scope and ownership

Branch codex/lint-helpers-05. The original three helpers were already committed and pushed at 72358fc and 77b2029. The continuation began with another push, which reported Everything up-to-date.

Every selection explicitly fetched all origin codex/lint-helpers* refs and inspected all six remote claim directories. The shared comment bundle was also respected through its older HELPERS.md claim. The readiness ranking uses dependent-rule counts; removal of one helper dependency is not a claim that its whole rule is implemented.

The first continuation factory claim cba50e2 collided with slot 01's earlier c250b83. It was withdrawn in dd9e8c3. No factory code is delivered or credited. A temporary factory runner failed native lowering of a nested local compiler callback before any factory mutant ran. The experiment was discarded after ownership yielded; this is not a completed check.

The dispatcher was reserved in dd9e8c3 at 00:48:53 UTC. A later refresh found slot 03's ce7f300 at 00:48:49 UTC. Slot 03 retains it. This slot's comparison and compiling mutant passed, but its duplicate implementation and test were removed. Evidence of that experiment remains explicitly named withdrawn-dispatcher-experiment.log; its dependency and mutant are not credited to the retained delivery.

Callee values f5028c4 and variable values 7b33442 remain uniquely owned. Ahra's correction instructed workers to finish existing claims, keep changes inside their own directories, and claim nothing more. Accordingly no replacement for the lost third slot was claimed. This continuation stops with two retained helpers. No shared registration generator, shared harness, compiler implementation, other worker directory or Go cohere worktree was edited. All new Adamic sources are .a. The local independent oracle runner lives entirely under slot05/batch2.

## Delivered behavior and dependency accounting

| Helper | File | Consumers | Final blockers removed alone |
|---|---|---:|---:|
| tailwind.*ClassLiteralReader.calleeValues | tailwind_callee_values.a | 11 | 0 |
| tailwind.*ClassLiteralReader.variableValues | tailwind_variable_values.a | 11 | 0 |

There are 22 dependency occurrences across the same 11 rules. [CONSUMERS.md](CONSUMERS.md) lists every rule for both helpers; [readiness.json](readiness.json) retains every residual dependency. This batch does not make any additional rule completely helper-ready. Counting only this slot's five delivered helpers across both batches gives 76 removed occurrences across 49 distinct rules, with the original two final blockers removed. Other workers' implementations are not assumed integrated.

Callee matching uses only the direct decoded Identifier text. Parenthesized callees, member calls and other expression kinds decline. A present argument list is required, all arguments are collected in order, and the caller supplies collectClassValues with Callee origin and false outer edges.

Variable matching requires an initializer and Identifier declaration name. Pattern callbacks preserve Go regexp.MatchString semantics and input order, stop after the first match, and delegate to classValuesUnder with Variable origin and false outer edges. The regex implementation and downstream traversal are explicitly owned elsewhere.

SurfaceNode uses exact named AST fields and numeric non-owning indexes. ClassValues indexes an immutable full-payload pool retaining node identity, text, range, origin and edges. Callee accumulation copies ordered entries; variable delegation preserves the dependency's arrays. Nil slices are represented by empty arrays. No ownership cycle is introduced.

Observation: Go's private AST AsCallExpression/AsVariableDeclaration assertions panic on nil and wrong kinds before the apparent nil guards. Go probes establish refusal, and both Node and sanitized native must exit 70 with an explicit panic. Exact runtime panic prose is not promised.

## Oracle and checks

The oracle overlay exposes the real private Go methods at cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. Pin drift and missing consumer coverage fail the tests. It reads all nonempty Go string literals from every inventory-listed test file for every consumer, deduplicates, adds controls, parses TSX, and repeats with default, empty and custom settings. Custom settings include an invalid regex that Go skips.

Each helper uses 1,039 distinct test-file strings and controls and 3,117 parser/settings cases. These are not counts of complete lint fixtures: prose, option and malformed-source literals are included.

- Callee reader: all 162 CallExpression verdicts match Go, Node source and sanitized native.
- Variable reader: all 1,023 VariableDeclaration verdicts match Go, Node source and sanitized native.
- Both readers: independent nil/wrong-kind refusal probes match Go's refusal behavior.
- Actual Go collection results and regex verdicts are supplied to explicitly external callback dependencies. Payload IDs are interned by their complete Go-record contents.

Commands from the repository root, after sourcing /workspace/adamic-tools/env.sh. Every test output went directly to a log file, never through a pipe:

    ADAMIC_SLOT05_BATCH2_EVIDENCE=/workspace/adamic/stage1/cohere/lint/helpers/slot05/batch2/evidence ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch2 -count=1 -v -timeout=15m > /tmp/lint05-batch2-retained.log 2>&1

PASS in 107.151s. The entire final touched package runs both baselines, both refusal contracts and all three retained compiling mutants. See evidence/retained-package.log. Corpus hashes and per-consumer counts are committed under evidence/.

    go vet ./stage1/cohere/lint/helpers/slot05/batch2 > /tmp/lint05-batch2-vet-retained.log 2>&1
    go run ./cmd/adamic types stage1/cohere/lint/helpers/slot05/batch2/main.a > /tmp/lint05-batch2-types-retained.log 2>&1
    gofmt -l stage1/cohere/lint/helpers/slot05/batch2 > /tmp/lint05-batch2-format-retained.log
    git diff --check > /tmp/lint05-batch2-diff-retained.log

Final checks exit 0. Vet, formatting and whitespace logs are empty. The first staged whitespace check found a trailing blank line in CONSUMERS.md; it was removed in the documentation commit before the final branch check. Types printed the checked entry-point variable, parameter and function types. These checks were repeated after removal of the withdrawn dispatcher.

    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m > /tmp/lint05-batch2-oracle.log 2>&1

PASS in 1.011s: all six fixtures, zero cache hits and six probe misses. See evidence/filtered-oracle.log. The full repository gate was not run. The original helper package passed in the prior report; this continuation's Go package is new and isolated.

## Every compiling mutant

Only a mutant that compiles, exits 0, emits no stderr and differs from the external Go result is credited. Native builds enable ASan/UBSan and Linux leak checking.

| Retained helper | Mutation | Witness |
|---|---|---|
| Callee | Invert the direct-Identifier guard | Verdict 22: mutant empty, Go literal 0 |
| Callee | Collect only the first argument | Verdict 37: mutant literal 0, Go literals 0,1 |
| Variable | Invert the successful-pattern requirement | Verdict 22: mutant empty, Go literal 0 |

The callee selection mutant passed in the initial standalone run (41.825s), the variable mutant passed in its standalone run (35.173s), and all retained mutants passed again in the final package. The extra callee argument mutant demonstrates that later arguments cannot silently disappear.

Withdrawn dispatcher experiment: disabling variable-declaration dispatch was caught at verdict 12,790, mutant empty versus Go literal 0. It passed standalone in 41.471s and again with the two delivered readers wired in. The four-test experiment package passed in 147.118s. These are recorded for completeness, not counted as retained work.

## Setup and limits

The continuation reused the successful original unit setup; no redundant reinstall was needed. nproc again printed 5. Go 1.27.1, clang 20.1.8, Node 24.19.0. /opt/adamic-tools/env.sh is absent here; the setup-installed path is /workspace/adamic-tools/env.sh.

Original setup timing lines:

    setup: go ready (0s)
    setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
    setup: node ready (1s)
    setup: submodules ready (1s)
    setup: build cache warm (88s)
    setup: done in 88s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB

Not covered: whole-rule findings, fixes, emitted-JavaScript comparison, production AST/payload adapter integration, Go-compatible regex compilation/matching implementation, downstream collection, dynamic concatenation of Go fixture sources, runtime-loaded external corpora, arbitrary malformed arenas, and the full repository gate. The generator explicitly refuses namespace-attribute assertion inputs rather than inventing a clean result. The third continuation slot cannot be delivered without a replacement claim, which Ahra's correction forbids at this point.
