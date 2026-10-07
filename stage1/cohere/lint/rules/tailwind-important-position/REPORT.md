Built: important-marker positioning, variable syntax, and variant-order semantics in owned .a modules, with temporary rule.ts facades.
Commits: claim 1e405194 was pushed first; important positioning 69bb89e1, variable syntax 30867c28, variant ordering 529bc65e.
Commands and outputs: the six-test gate passed in 253.449s; three supplemental checks passed in 48.012s; see evidence/three-final.log, evidence/configured.log and the final .a-driver rerun in evidence/dot-a-final.log (PASS, 118.171s).
Mutants: all three compiling silence mutants disagreed with Go on Node, emitted JavaScript and ASan/UBSan native; both provider and custom-regexp guard mutants lost their required refusals.
Not covered: raw JSX parsing, custom variable regular expressions, the native Tailwind ProgramReads/design-system provider, and the shared full gate.

The prior five claims were completed and pushed through d4808055 before selecting these three. See ../func-name-matching/REPORT.md. The earlier three TypeScript rules remain pushed through e373f8e0. This batch was selected from inventory array order after all 46 helper-ready names were occupied: all 337 origin refs and 52 distinct Markdown claim blobs were inspected after the older implementations were pushed. The claim snapshot and names are in ../../claims/wave1-04.md.

The three implementations, surface adapter and private driver are .a. The existing registration generator requires rule.ts, so each rule has a temporary facade using Ahra's explicit exception. No shared registration, shared harness, or compiler file was changed. Validation is private to this owned rule directory. It copies the lint sources into scratch space, generates the registry there, compiles a single driver, and compares complete findings, byte ranges, messages, fixes, suggestions, and unchanged fixed source against unmodified Go rules.

The important-marker and variable-syntax implementations reproduce Go's delimiter balancing, whitespace splitting, marker handling, modifiers, default and decoded options, and three class surfaces. The owned surface adapter reads nested value positions but leaves unrelated calls and comparison operands alone. Additional raw TypeScript witnesses exercise calls, default variables, arrays, conditionals, logical operands, assertions, template holes, escapes and astral UTF-8 offsets.

Variant ordering uses stable insertion sorting, preserves the order of two element-scoped variants, and moves globals in descending dense-rank order. Its private validation provider loads a real Go program with Tailwind 4.1.18, parses actual candidates and supplies only unsorted candidate ranks plus Go's global flag. It supplies neither corrected text nor findings. Adamic computes splitting, sorting, range, message and findings. This measures the ported rule conditional on resolved design-system facts; it does not certify an Adamic CSS engine or native ProgramReads implementation.

The original Go tests are captured through scratch Go overlays. Their production rules and expected assertions are unmodified. The machine-specific Tailwind package lookup in the original test fixture is redirected to the pinned installed package, so variant tests run rather than skip. There are 16 distinct important-position, 14 variable-syntax and 15 variant-order upstream cases, all JSX. These are compared using a Go AST projection because the unchanged stage1 parser refuses JSX attributes. Separate tests require the same explicit parser refusal on Node, emitted JavaScript and sanitized native. Raw compiler/stage1 and additional witnesses use the independent parser.

Exact integration blockers:

* Raw JSX returns exit 70 with `parser slice expected GreaterThanToken`. A projection proves rule semantics, not independent parsing.
* Stage 0 rejects a nonconstant `RegExp` pattern. The two default variable patterns are implemented with equivalent suffix checks. Custom patterns explicitly refuse with exit 70 and `NotYet: nonconstant Go variable-pattern regexp provider`; this includes invalid custom patterns Go would skip. Their behavior is not certified. The first build evidence is retained in evidence/first.log.
* The shared context has no Tailwind ProgramReads or resolved design-system adapter. For scale, the Go Tailwind collapse substrate has 35 production files and 17,452 lines; its program package has 35 production files and 9,403 lines. These are observed source sizes, not an estimate of the minimum work needed for this one adapter. Selecting variant ordering without the explicit private provider refuses with exit 70 and `NotYet: shared Tailwind ProgramReads and resolved design-system adapter`. A real Go positive control and a compiling provider-guard mutant prove the refusal can fail. A native Tailwind engine was not implemented here.
* The shared profile test still fails to compile at profile_test.go:32 because it ranges over the portFiles function. The current generator/copy contract still requires .ts facades. codex/lint-harness-dot-a owns these shared changes; this unit has not edited or imported those shared files. Full-root tests were not run; the owned package, bounded vet and filtered oracle evidence are reported instead.

Toolchain: bash cloud/setup.sh reached Go, clang, Node and submodule readiness, each reporting 0s, then failed in that shared profile compile error during cache warming. The working environment is /workspace/adamic-tools/env.sh, Go 1.27.1, clang 20.1.8, Node 24.19.0, nproc 5. The initial setup transcript is ../typescript-no-non-null-asserted-optional-chain/evidence/setup.log. npm installed tailwindcss@4.1.18 in /tmp/wave104-tailwind, reporting 786ms; reproducing the private tests requires ADAMIC_TAILWIND_PACKAGE pointing at that package directory.

Evidence/variant-first.log records an initial private-adapter empty-object JSON error, subsequently corrected. Evidence/final.log is the passing two-rule run before variant integration; evidence/three-final.log is the six-test three-rule run. The private driver was then renamed from .ts to .a, leaving only the registration facades as temporary .ts, and its raw fixtures, complete source corpus and JSX projection checks were rerun in evidence/dot-a-final.log. Test output is written directly to log files.

Validation observations

Go cohere is pinned at 715ba94f3608a6500086b1076ce5cb7e51b836db; TypeScript's compiler is pinned at 050880ce59e30b356b686bd3144efe24f875ebc8. The main run compared 11,374 identical bytes over six raw witness cases, 37,296,202 identical bytes over all 241 compiler/stage1 files and 723 rule/file pairs, and 15,246 identical bytes over all 45 original JSX cases using Go AST geometry. No source files were excluded. The three rule mutants passed in 51.77s; explicit provider refusals and the provider guard mutant passed in 52.09s. Supplemental configured-callee cases compared 1,642 identical bytes with exactly three findings. The custom-regexp guard mutant passed in 21.89s.

Each `mutant.json` replaces the facade's delegated visit with a condition that accepts only negative node indices. Actual walks have nonnegative indices. Each mutant compiles, exits zero with empty stderr and loses the real positive finding; only Go's full-output comparison catches it on each execution path. The provider guard mutant replaces `completeFacts && facts >= 0` with `facts >= 0`: it produces the correct Go finding despite the provider not being explicitly enabled, so the required-refusal check alone catches it. The custom-regexp guard mutant replaces the `NotYet` panic with an empty branch: it finishes silently and is caught both by the missing-refusal check and the real Go finding.

Whole-source throughput, best of three unsanitized native/Node/Go runs, includes source reading, parsing and traversal. All three rules had zero findings on the 241-file source corpus, so all three report 0.00 findings/s. Observed seconds:

| Rule | Native | Node | Go |
| --- | ---: | ---: | ---: |
| Important position | 1.292319 | 0.896885 | 0.235075 |
| Variable syntax | 1.324688 | 0.901402 | 0.235904 |
| Variant order | 1.330805 | 0.940578 | 0.241322 |

A separate best-of-three benchmark repeats each raw positive witness 300 times and requires exactly 300 findings on every side. Observed findings/s:

| Rule | Native | Node | Go |
| --- | ---: | ---: | ---: |
| Important position | 40,753.30 | 2,312.05 | 24,511.18 |
| Variable syntax | 36,500.68 | 2,228.88 | 25,915.26 |
| Variant order | 14,979.15 | 1,982.90 | 11,603.94 |

The timing observations above were made before the private driver filename was changed from .ts to .a; the final .a driver separately passed the full raw and projected output comparisons. These positive measurements include process startup and repeated parsing/I/O, and are not a substitute for the compiler/stage1 corpus. For variant ordering, Node and native consume previously resolved Go facts while Go resolves candidates in a real loaded program. Fact preparation is outside those Node/native timing intervals. The variant numbers therefore measure the sorting/reporting slice, not comparative complete design-system implementations.

Exact commands run, with environment activated by `source /workspace/adamic-tools/env.sh`:

* `ADAMIC_TAILWIND_PACKAGE=/tmp/wave104-tailwind/node_modules/tailwindcss ADAMIC_WAVE104_EVIDENCE=$PWD/stage1/cohere/lint/rules/tailwind-important-position/evidence ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-04-typescript ADAMIC_LINT_BENCH=1 go test ./stage1/cohere/lint/rules/tailwind-important-position -count=1 -v -timeout=15m > stage1/cohere/lint/rules/tailwind-important-position/evidence/three-final.log 2>&1`: PASS, 253.449s (six checks at that commit).
* `ADAMIC_TAILWIND_PACKAGE=/tmp/wave104-tailwind/node_modules/tailwindcss ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-04-typescript ADAMIC_LINT_BENCH=1 go test ./stage1/cohere/lint/rules/tailwind-important-position -run '^Test(ConfiguredCallees|CustomRegexGuardMutant|PositiveThroughput)$' -count=1 -v -timeout=10m > stage1/cohere/lint/rules/tailwind-important-position/evidence/configured.log 2>&1`: PASS, 48.012s.
* `ADAMIC_TAILWIND_PACKAGE=/tmp/wave104-tailwind/node_modules/tailwindcss ADAMIC_WAVE104_EVIDENCE=$PWD/stage1/cohere/lint/rules/tailwind-important-position/evidence ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-04-typescript go test ./stage1/cohere/lint/rules/tailwind-important-position -run '^Test(RulesAndSuggestions|CompilerAndStage1|JsxProjectedSemanticsAndParserBlocker)$' -count=1 -v -timeout=10m > stage1/cohere/lint/rules/tailwind-important-position/evidence/dot-a-final.log 2>&1`: PASS, 118.171s for the final .a-driver comparison.
* `go vet ./stage1/cohere/lint/rules/tailwind-important-position > stage1/cohere/lint/rules/tailwind-important-position/evidence/vet.log 2>&1`: exit zero, empty log.

The earlier filtered uncached input oracle passed in 1.073s with all six probe misses and no cache hits, recorded in ../func-name-matching/evidence/filtered-oracle.log. No changes were made to compiler, parser, runtime, shared helpers, registration generator or shared tests.
