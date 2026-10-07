Built: this-alias implementation with four-way original-fixture and source-corpus parity; two suggestion rules remain blocked.
Commits: claim 9ff521a9 pushed before source; this-alias e41b30da; evidence commit is named in the final response.
Commands and outputs: 40 original cases, an option matrix and 218 compiler/stage1 files agree with Go on Node, emitted JavaScript and sanitized native; own lint/format, registry, vet and filtered oracle pass.
Mutants: inverting alias allow-list membership compiles and is caught only by comparison on all three backends; a compiling Go serializer control proves both suggestion-blocker checks can fail.
Not covered: the optional-chain and unnecessary-constraint ports and their rule mutants, default shared integration, malformed configuration diagnostics, and the full repository gate.

## Scope after Ahra's correction

No additional claim is made. The three current names are:

1. @typescript-eslint/no-non-null-asserted-optional-chain
2. @typescript-eslint/no-this-alias
3. @typescript-eslint/no-unnecessary-type-constraint

Before the correction, the prior work was pushed through e50c08c4, then origin
was fetched with --prune --recurse-submodules=no and all heads. Main remains
ef3d907ecdc4c771b016f7d9c52372def057a340. Selection inspected 320 refs and 39
unique claim Markdown blobs, plus main .a/.ts lint sources. All 46 helper-ready
names occur in claims. The listed names are the first remaining inventory
entries with needs_type_information=false. The claim was pushed before source.
selection.py reproduces the pre-claim snapshot using e50c08c4 for this branch.

Ahra's latest instruction authorizes temporary .ts modules while the shared
harness worker adds .a support on codex/lint-harness-dot-a. Only the current
this-alias modules were switched to rule.ts and messages.ts. The later integration
codemod owns their rename. No shared generator, test harness, compiler, parser,
or cohere submodule file was edited. Comparisons reused the previously proposed
scratch compatibility overlay; the production patch remains unapplied. The
other earlier .a candidates and profile API mismatch still prevent this branch's
default lint package gate until the shared work is integrated. This implementation
is validated through the scratch harness, not an unconditional merge claim.

## This alias behavior and validation

The owned directory is rules/typescript-no-this-alias. It contains descriptor,
concrete listener and factory, exact messages, independent Go adapter, mutant,
raw witnesses and reproducible validation. It preserves Go's .ts/.tsx/.mts/.cts
extension gate (.js and .a are excluded), bare-this right-hand sides, all
assignment operators, direct destructuring targets, recursive identifier target
unwrapping, member exclusions, options inversion and allow-list aliases and
precedence. Only the original target's direct object/array shape takes the
assignment destructuring branch. Parenthesized patterns remain silent. An
angle-bracket assertion's expression uses the parser's second child. Identifier
and pattern spans, messages and absence of fixes/suggestions agree exactly.

Captured Go options already use ReportDestructuring and AllowedNames. Public
wire options use allowDestructuring, allowedNames and allowNames. The adapter
and local JSON decoder handle both forms and preserve null/duplicate behavior.
The original fixture capture retains File as well as Source and Options, so the
JavaScript exclusion is actually tested. Forty distinct original combinations
are compared; the original Go test family, including direct option-decoder
assertions, also passes.

All commands source /workspace/adamic-tools/env.sh first. Outputs go directly to
logs. The final shipping commands are:

```sh
go test -overlay=/tmp/lint-wave1-12-batch4-overlay.json ./stage1/cohere/lint \
  -run '^(TestAliasUpstream|TestAliasOptions)$' -count=1 -v -timeout=10m \
  > /tmp/batch4-shipping.log 2>&1

go test -overlay=/tmp/lint-wave1-12-batch4-overlay.json ./stage1/cohere/lint \
  -run '^TestMutants$/this_alias_allowed_name_ignored$' -count=1 -v -timeout=10m \
  > /tmp/batch4-mutant.log 2>&1

ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 \
  go test -overlay=/tmp/lint-wave1-12-batch4-overlay.json ./stage1/cohere/lint \
  -run '^(TestCompilerAndStage1Agree|TestAliasThroughput)$' \
  -count=1 -v -timeout=20m > /tmp/batch4-corpus.log 2>&1

go test -overlay=/tmp/lint-wave1-12-batch4-overlay.json ./stage1/cohere/lint/registry \
  -count=1 -v > /tmp/batch4-registry.log 2>&1

go vet -overlay=/tmp/lint-wave1-12-batch4-overlay.json ./... > /tmp/batch4-vet.log 2>&1

ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m > /tmp/batch4-oracle.log 2>&1

(cd cohere && go test ./internal/lint/rules/typescript \
  -run '^(TestNoNonNullAssertedOptionalChain|TestNoThisAlias|TestNoUnnecessaryTypeConstraint)' \
  -count=1 -v -timeout=10m > /tmp/batch4-upstream.log 2>&1)
```

Observed final original/options PASS: 46.531s. Original fixture comparison:
40 combinations, 15,665 identical canonical bytes, 25.35s. Expanded matrix:
95,230 identical bytes, 21.17s; six filename extensions, ten option forms,
wrappers, direct/parenthesized patterns, and angle assertions only on .ts/.mts/.cts.
A prior matrix incorrectly put an angle assertion in a .js file. Go's fixer
correctly rejected it as JSX; that failure is retained in final-corpus.log and
not counted as passing. The fixture was corrected without changing shared code.

Final corpus/throughput PASS: 66.869s. All 77 TypeScript src/compiler files and
the stage1 snapshot total 218 files, with 12,631,585 identical canonical bytes
across Go, source Node, emitted JavaScript and ASan/UBSan native, in 41.99s.
Compiler pin: TypeScript v6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8.
Owned witnesses previously passed with 14,208 identical bytes. Final semantic
mutant PASS: 26.400s, caught on all three backends after successful compilation
and execution with no stderr. It changes only allow-list membership polarity.

Registry controls pass, final repository vet is empty with exit 0, original Go
families pass in 0.032s, and filtered uncached oracle passes in 0.493s with
native/node cache hits 0 and misses 1. The full repository gate was not run.

The owned .ts modules pass pinned cohere's lint and format check:

```sh
/tmp/lint-wave1-12/cohere --no-fix \
  stage1/cohere/lint/rules/typescript-no-this-alias/rule.ts \
  stage1/cohere/lint/rules/typescript-no-this-alias/messages.ts \
  > /tmp/batch4-self-gate.log 2>&1
```

Exit 0; 276 rules, two checked modules, formatting included. Initial style
findings were fixed only in the owned modules. An initial --fix --format-only
invocation was rejected as contradictory flags and changed nothing; the correct
formatting-only invocation then passed. Both logs are retained.

Toolchain setup through the existing scratch overlay:
GOFLAGS=-overlay=/tmp/lint-wave1-12-batch4-overlay.json bash cloud/setup.sh.
Exit 0. Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s,
build cache warm 37s, done in 37s on 5 processors; cpu.max 400000 100000,
17.6 GB; nproc 5. The earlier default profile compile failure is recorded in the
preceding batch. Tool versions: Go 1.27.1, clang 20.1.8, Node 24.19.0.

validate.py rebuilds the previously proposed compatibility sources into scratch
and maps the owned validation_test.go.txt into a virtual test source. Its
--prepare-only path was verified. It runs the original-filename corpus rather
than the known-broken shared all-rule capture, and never edits shared repository
files. Source env.sh and supply --scratch and --typescript to reproduce.

## Exact blockers and stop

The two unimplemented rules report real suggestions whose edit ranges are
smaller than the finding ranges:

- Optional-chain assertion: finding on the asserted expression, suggestion
  removing only the bang.
- Unnecessary constraint: finding on the parameter name, suggestion replacing
  the following extends clause, sometimes with a comma for arrow disambiguation.

The shared oracle requires exactly one suggestion with exactly one edit whose
range equals the finding range. Both positive witnesses compile and run the real
Go rules, then fail in the unchanged serializer with panic: unexpected suggestion
shape. Dropping the suggestions or replacing their ranges would violate the
requested byte-for-byte behavior. This needs a shared finding/edit contract and
serializer change beyond .a support. No partial rule is registered.

Independent blocker test:

```sh
go test ./stage1/cohere/lint/claims/wave1-12-batch4-evidence \
  -count=1 -v -timeout=10m > /tmp/batch4-suggestions.log 2>&1

ADAMIC_SUGGESTION_CONTROL=1 \
  go test ./stage1/cohere/lint/claims/wave1-12-batch4-evidence \
  -count=1 -v -timeout=10m > /tmp/batch4-control.log 2>&1
```

The baseline passes in 2.907s. The compiling control suppresses suggestion
serialization only in a temporary Go copy; both witnesses then finish and are
caught by the check, producing the intentional test failure in 7.783s. This is a
blocker-probe control, not either missing rule's semantic mutant. No production
Go oracle or rule is altered. Per Ahra's instruction, work stops at these exact
shared blockers; there are no further claims.

## Findings per second

Release native, source Node and Go, best of five interleaved count-only runs.
All agree on 1,000 findings over 77 compiler files plus one synthetic file with
1,000 aliases. Compiler files alone have zero findings. Startup/read/parse/visit
are included; compilation and formatted findings/fixes are excluded. Correctness
and mutants use sanitized native. Other bounded checks may overlap the runs.

| Backend | Best seconds | Findings per second |
|---|---:|---:|
| Native | 1.275352 | 784.10 |
| Node | 1.071419 | 933.34 |
| Go | 0.219845 | 4548.66 |

No rule throughput or semantic rule mutant is claimed for the two blocked rules.
