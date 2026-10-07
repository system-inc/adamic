Built: rebased all eleven owned rules onto current main, preserving Go comparisons through an owned frozen test contract; shared harness integration is parked on #zmh9v36.
Commits: main f8013f0baac41ddc340d76f83bddde38536a8f07; helper branch beeb653309e21ed3f300d6866d870e8461ca30ed; rebased rule implementation e09c70ca3, followed by this parking evidence commit.
Commands and outputs: setup PASS 38s, nproc 5; owned packages PASS 152.803s, 246.039s and 218.749s; vet PASS; 3,916 compiler/stage1 rule/file pairs, 147,374,518 identical bytes.
Mutants: all eleven rule-semantic mutants, numeric subscription mutant, suggestion guard mutant, custom-regexp guard mutant and Tailwind provider guard mutant compiled and were caught on Node, emitted JavaScript and ASan/UBSan native.
Not covered: production shared registration, numeric handed-node execution, independent JSX support, native Tailwind ProgramReads/design-system providers, complete shared suggestions, fresh throughput measurements or the full repository gate.

# Parking status

Fresh fetch of every origin head still names main f8013f0. Both owned branches
contain it. The rule branch now carries only owned rule commits above the already
rebased helper branch. The incompatible inherited registration commits were not
replayed onto main. No shared finding.ts, context.ts, main.ts, lint.ts, registry
generator, oracle or lint_test comparison changed relative to main.

The remaining integration blocker is missing shared directory registration and its
contract, being unified on #zmh9v36. This branch is parked under Ahra's instruction,
with owned tests green using the prior contract as a frozen scratch fixture.
The fixture is documented in parking/README.md. It is not a production harness
replacement or a claim that main's monolithic driver can execute these directories.
When the shared harness handoff SHA is named, rebase and replace this temporary
fixture with that contract before any further helper reservation.

Existing explicit adapter refusals remain. Seventeen font/physical-direction JSX
cases and forty-five important/variable/variant JSX cases compare through real Go
AST projection; they also prove the independent parser refuses with exit 70.
Variant facts are supplied by real Go ProgramReads/design-system observations,
not synthesized answers. Missing providers and nonconstant custom regexes refuse.
Suggestion comparisons retain all edits; the incomplete shared model still refuses.
These are conditional semantic certificates and explicit boundaries, not completed
production adapter claims. Numeric declarations are checked against pinned ast.Kind;
the index/string-kind legacy contract does not provide the requested speedup.

# Commands and retained evidence

Source /workspace/adamic-tools/env.sh for each command. The compiler source pin is
050880ce59e30b356b686bd3144efe24f875ebc8 at /tmp/lint-wave1-04-typescript.

- bash cloud/setup.sh: PASS; Go/clang/Node 0s, submodules 1s, cache warm 38s,
  done 38s, nproc 5, cgroup cpu.max 400000 100000. Go 1.27.1, clang 20.1.8,
  Node 24.19.0. Log: evidence/parking-setup.log.
- ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-04-typescript go test
  ./stage1/cohere/lint/rules/typescript-no-non-null-asserted-optional-chain
  ./stage1/cohere/lint/rules/func-name-matching
  ./stage1/cohere/lint/rules/tailwind-important-position -p=1 -count=1 -v
  -timeout=20m: first two packages PASS 152.803s and 246.039s. The process
  exited 1 because I omitted ADAMIC_TAILWIND_PACKAGE for the final package.
  This invocation error is retained, not counted as a semantic mutant or green gate.
  Log: evidence/parking-full.log.
- ADAMIC_TAILWIND_PACKAGE=/tmp/wave104-tailwind/node_modules/tailwindcss
  ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-04-typescript go test
  ./stage1/cohere/lint/rules/tailwind-important-position -count=1 -v
  -timeout=20m: PASS 218.749s, all checks in that package. The pinned installed
  package is Tailwind 4.1.18. Log: evidence/parking-tailwind.log.
- go vet ./stage1/cohere/lint/rules/typescript-no-non-null-asserted-optional-chain/...
  ./stage1/cohere/lint/rules/func-name-matching
  ./stage1/cohere/lint/rules/tailwind-important-position: PASS, empty log.
- git diff --check: PASS. Named shared-file diff against origin/main: empty.

Each corpus contains all 356 current compiler/stage1 sources. The three suites
cover 1,068 / 1,780 / 1,068 rule/file pairs and compare respectively 41,331,548 /
66,299,597 / 39,743,373 bytes. Their upstream captures, independent/projection
controls, complete repair serialization and all compiling mutants also pass.
Benchmarks are intentionally not remeasured during parking; earlier owned reports
contain their measured native, Node and Go findings/second, with their original bases.

# Rule-semantic mutants caught only by Go comparison

- consistent-this: reverse consistent-this judgement
- eslint-comments-require-description: invert eslint-comments-require-description guard
- func-name-matching: reverse func-name-matching judgement
- next-google-font-display: omit blocking font display
- tailwind-important-position: silence important-position
- tailwind-no-physical-direction: invert tailwind-no-physical-direction guard
- tailwind-variable-syntax: silence variable-syntax
- tailwind-variant-order: silence variant ordering
- typescript-no-non-null-asserted-optional-chain: optional chain suggestion deletes preceding token
- typescript-no-non-null-assertion: non null assertion loses second suggestion edit
- typescript-no-this-alias: this alias initializer unwraps parentheses

All finish cleanly and differ from Go on all three targets. Guard mutants separately
prove required refusals; the numeric mutant substitutes kind 261 for kind 307.
