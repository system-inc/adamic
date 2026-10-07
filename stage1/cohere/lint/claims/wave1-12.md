# Lint wave 1, slot 12

Branch: codex/lint-wave1-12.

The helper report delegates its ordered 30 plus 16 rule list to HELPERS.md.
Positions 34, 35 and 36 are:

- 34: base/consistency-require-pagination-argument-name. Skipped: already ported on origin/codex/stage1-lint-batch4 in base_consistency_require_pagination_argument_name.ts.
- 35: base/correctness-require-orm-column-declare. Claimed here.
- 36: nexus/consistency-no-boolean-outcome. Claimed here.

No implementation is written before this claim is committed and pushed.
The requested docs/parallel-work.md does not exist on main or the foundations.
The foundations conflict in six shared lint files; the merge retains directory
registration in those files and incorporates the helper files and inventory.

## Unit report

Built two `.a` candidate rule ports; position 34 pagination is skipped because an origin branch already owns a port.
Commits: claim e450447; ORM 80ff908; boolean outcome 203969d; foundation merges 50c61df and a8361f5.
Checks: 256 captured cohere cases and 212 compiler/stage1 files agree across Go, source Node, emitted JavaScript and sanitized native under the proposed compatibility overlay; registry, vet and a filtered uncached oracle pass.
Mutants: ignoring bare ORM decorators and removing the boolean companion requirement both compile/run and are caught only by output comparison on source Node, emitted JavaScript and sanitized native.
Not covered: default registration/build integration and `.a` self-lint are blocked; the full repository test gate was not run.

### Ownership and foundation observations

Started from fetched main d090af5. The checkout originally fetched only main;
explicitly fetched all origin heads to get the foundations and check existing
ports. The broad fetch also started unnecessary historical submodule fetches;
those task-owned fetch processes were stopped after the branch refs arrived.
The pinned cohere worktree remains 715ba94f3608a6500086b1076ce5cb7e51b836db.

The report's ordered list is linked in HELPERS.md rather than enumerated in
helpers/REPORT.md itself. Position 34 is already implemented in
`base_consistency_require_pagination_argument_name.ts` on
origin/codex/stage1-lint-batch4 at d486b03a15f3202acc317dc81c40cc21818e21f7.
It is skipped, not migrated or revalidated by this worker. Searching origin
branches found no implementation of the other two assigned names.

The foundations were not clean together: six conflicts occurred in README.md,
lint.ts, lint_test.go, main.ts, settings.ts and testdata/oracle.go. The merge
retained registration's versions in those files and incorporated the helper
files and inventory. The requested docs/parallel-work.md is absent from main
and both foundations. No compiler, parser, runtime or submodule source is edited.

The registration foundation hardcodes rule.ts and rejects .a mutant modules.
The merged profile_test.go also ranges over the old portFiles value, whereas
registration makes it a function. Default `go run ./cmd/lint-registry` exits 1
opening the new ORM directory's nonexistent rule.ts; default `go test` fails to
compile profile_test.go at 32:23. These failures are recorded separately from
successful overlay runs and are not called green.

A proposed [compatibility.patch](../rules/nexus-consistency-no-boolean-outcome/compatibility.patch)
extends the shared Go registration/discovery code to .a, compares emitted
JavaScript, and fixes the profiling API mismatch. It is deliberately unapplied:
the user limits edits to owned rule directories. A question requesting the
territory exception remains unanswered. The patch and reproducible scratch-only
runner are concrete, reviewable artifacts. They require no hand-edited dispatch,
per-rule oracle list or per-rule corpus list. This branch is a validated candidate,
not ready for the default gate or an unconditional merge.

### Implementations

ORM preserves eight decorator names, first matching source order, bare/called
identifier decorators, the ambient/declare exemption, computed-key unwrapping,
identifier/private/literal names, exact policy messages and spans.

Boolean outcomes preserve interface and bare type-literal alias scope, first-name
order and last duplicate definition, bare boolean recognition, companion fields,
Properties/third-party exemptions, allowedTypeNames, repeated role suffixes,
whole-signature spans and exact policy rendering. Shared StrictOptions validates
the generated target descriptor and OptionsJson decodes values. Both rules have
no automatic fix or suggestion; their fixed sources remain byte-identical.

All four newly authored Adamic modules use .a. The raw TypeScript witnesses use
the registration contract's .ts.txt extension. Generated registries remain ignored.

### Commands and observed outputs

Every test wrote directly to a log. Raw logs are copied unchanged into
[owned evidence](../rules/nexus-consistency-no-boolean-outcome/evidence/), with
.txt names so they can be reviewed alongside the code. No test output was piped.

The initial `bash cloud/setup.sh > /tmp/lint-wave1-12-setup.log 2>&1` failed in
cache warming because it overlapped the foundation merge and saw conflict markers
in lint_test.go at 28:1. Go, clang, Node and submodule setup had finished at 0s,
1s, 1s and 2s. The initial dependency warm-up for cohere's oracle then completed
successfully but emitted download notices for regexp2 on stderr, which the strict
execution harness treated as a failure. The rerun uses the now-warm dependencies.
Both failed attempts are retained in evidence.

Workaround:

```sh
GOFLAGS=-overlay=/tmp/lint-wave1-12/overlay.json bash cloud/setup.sh \
  > /tmp/lint-wave1-12/setup-overlay.log 2>&1
source /workspace/adamic-tools/env.sh
nproc
```

Exit 0. Timing output:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (46s)
setup: done in 46s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

nproc: 5. Go 1.27.1, clang 20.1.8, Node 24.19.0, Linux x86_64.

```sh
go test -overlay=/tmp/lint-wave1-12/overlay.json ./stage1/cohere/lint \
  -run '^TestOwnedWitnesses$' -count=1 -v -timeout 10m \
  > /tmp/lint-wave1-12/witness2.log 2>&1

ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 \
  go test -overlay=/tmp/lint-wave1-12/overlay.json ./stage1/cohere/lint \
  -run '^(TestRulesAgree|TestCompilerAndStage1Agree|TestWave12OptionsAndShapes|TestWave12Throughput|TestMutants/(orm_bare_decorator_ignored|boolean_outcome_companion_ignored))$' \
  -count=1 -v -timeout 20m > /tmp/lint-wave1-12/parity.log 2>&1

go test -overlay=/tmp/lint-wave1-12/overlay.json ./stage1/cohere/lint \
  -run '^TestMutants$/(orm_bare_decorator_ignored|boolean_outcome_companion_ignored)$' \
  -count=1 -v -timeout 10m > /tmp/lint-wave1-12/mutants.log 2>&1

go test -overlay=/tmp/lint-wave1-12/overlay.json ./stage1/cohere/lint/registry \
  -count=1 -v > /tmp/lint-wave1-12/registry-tests.log 2>&1

go vet -overlay=/tmp/lint-wave1-12/overlay.json ./... \
  > /tmp/lint-wave1-12/vet.log 2>&1

ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout 10m \
  > /tmp/lint-wave1-12/oracle.log 2>&1
```

All six exit 0. Witnesses: 7,994 identical bytes, PASS in 31.244s. Corpus/shapes/
throughput run: PASS in 321.607s. It contains 256 distinct captured cohere
source/rule/options combinations, 147,239 identical canonical bytes; 212 files,
12,620,096 identical bytes; synthetic options/shapes, 29,882 identical bytes.
The compiler pin is TypeScript v6.0.3 at
050880ce59e30b356b686bd3144efe24f875ebc8, all 77 src/compiler files. Stage1
coverage includes .a and .ts. New rules run alongside the five registration
foundation rules, including their finding/fix comparisons.

Mutants: PASS in 78.364s. Registry: PASS in 0.074s, including descriptor rejection
controls and deterministic generation. Vet: empty log, exit 0. Filtered uncached
oracle: PASS in 13.524s, native/node cache hits 0 and misses 1.

The [runner](../rules/nexus-consistency-no-boolean-outcome/validate.py) provides
separate, simpler corpus and mutant filters for reproduction. Its prepare-only
mode rebuilt all three proposed Go sources byte for byte against those used in
the validation. No shared file is modified when running it.

A dry self-lint was attempted with pinned Go cohere:

```sh
/tmp/lint-wave1-12/cohere --no-fix --lint --no-cache \
  stage1/cohere/lint/rules/base-correctness-require-orm-column-declare/rule.a \
  stage1/cohere/lint/rules/base-correctness-require-orm-column-declare/messages.a \
  stage1/cohere/lint/rules/nexus-consistency-no-boolean-outcome/rule.a \
  stage1/cohere/lint/rules/nexus-consistency-no-boolean-outcome/messages.a \
  > /tmp/lint-wave1-12/cohere.log 2>&1
```

Exit 1, "nothing to check": this pinned CLI rejects the .a extension despite
the repository's sourceExtensions setting. It did not perform self-lint. The
stage0 typechecker and lowering do accept these files, as the native and emitted
JavaScript builds demonstrate. No .ts source copy was substituted for this check.

### Mutants

| Mutant | Compiling change | What caught it |
| --- | --- | --- |
| orm_bare_decorator_ignored | Ignore a bare @OrmColumn, while preserving calls | Missing bare-property finding against Go on source Node, emitted JavaScript and sanitized native |
| boolean_outcome_companion_ignored | Remove the companion requirement | Extra finding on ordinary State.succeeded against Go on the same three executions |

Each execution exits 0 without stderr before its output is compared. A compiler
error, clang warning, sanitizer error or panic is not credited as a mutant kill.
The filtered external oracle also proves its one-byte stdout comparator can fail.
The registry suite's preexisting descriptor/adapter mutants are recorded in its
raw evidence; they are not replacements for these two semantic rule mutants.

### Findings per second

Best of five interleaved rounds, count-only end-to-end execution: startup, file
reads, parsing and visitors included; formatting, fixing and builds excluded.
Correctness uses ASan/UBSan/leak checks; timing uses unsanitized native -O2.

Both rules produce zero findings on the 77 compiler files alone. To avoid a
meaningless zero-rate measurement, each benchmark adds one 1,000-declaration
positive fixture. The ORM fixture has two findings per declaration; the boolean
fixture has one. All rounds and all three implementations report the same count.
These are compiler-plus-synthetic workload rates, not isolated visitor speeds.

| Rule | Files/findings | Native findings/s | Node findings/s | Go findings/s |
| --- | --- | ---: | ---: | ---: |
| base/correctness-require-orm-column-declare | 78 / 2,000 | 1,188.33 | 1,332.68 | 5,896.49 |
| nexus/consistency-no-boolean-outcome | 78 / 1,000 | 582.82 | 785.20 | 3,488.82 |

Best elapsed seconds: ORM native 1.683039, Node 1.500734, Go 0.339185;
boolean native 1.715800, Node 1.273556, Go 0.286630. Shared-machine noise is
visible across rounds. No performance guarantee is inferred.

### Remaining limits

Default integration requires the shared .a support and profiling merge repair.
Permission to apply that proposal is still pending. The foundation's ancillary
profiling/volume tests belong to its older monolithic ports; the bounded filtered
suite does not migrate or claim those rules. The full repository gate is not run.
No default package/build green result is claimed.

No fresh pagination validation; no complete arbitrary malformed-option diagnostic
prose, config/suppression frontend parity, general parser recovery or non-UTF-8
input coverage; no pinned CLI self-lint of .a. Cohere's own tests, valid decoded
options, full named corpora, byte spans/messages/fixed sources, and the two rule
mutants are covered as observed above. No PR is opened.

## Next three rules, October 7

All previous work is pushed through 068060e. Fetched origin/main is
ef3d907ecdc4c771b016f7d9c52372def057a340. All 305 origin refs were inspected;
unique claim Markdown blobs and main lint .a/.ts sources were searched.
All 46 helper-ready names occur in claims. Continuing in inventory array order
on origin/codex/lint-inventory, syntax-only means needs_type_information=false,
including entries still waiting on helpers. The first three absent from both
main implementations and all fetched claim Markdown are claimed here:

1. @next/next/google-font-preconnect
2. @next/next/inline-script-id
3. @next/next/next-script-for-ga

Owned directories use next-google-font-preconnect, next-inline-script-id and
next-next-script-for-ga under rules/. This update is pushed before rule code.

### Next-three outcome

The three reserved Next.js ports are blocked on the shared JSX parser and AST
adapter. Go cohere reports on all three positive witnesses; source Node, emitted
JavaScript and sanitized native produce byte-identical parser panics instead.
No incomplete registration is installed. Reproducible proof, a compiling probe
control and exact coverage limits are in
[the next-three report](wave1-12-next-evidence/REPORT.md). The claims remain owned
by this branch pending that foundation. No per-rule parity, mutants or throughput
result is claimed for these three names.

## Third batch, October 7

Prior work is pushed through b4656c7c. Main remains
ef3d907ecdc4c771b016f7d9c52372def057a340. After fetching all origin heads,
310 refs and 32 unique claim Markdown blobs were inspected alongside main lint
.a/.ts source. All 46 helper-ready names occur in claims. The first remaining
inventory entries with needs_type_information=false are claimed here:

1. @next/next/no-before-interactive-script-outside-document
2. @next/next/no-css-tags
3. @next/next/no-document-import-in-page

This ownership update is pushed before new source or probes. Reserved rule
directories are next-no-before-interactive-script-outside-document,
next-no-css-tags and next-no-document-import-in-page.

### Third-batch outcome

Document-import candidate committed as dfab0924. Four-way witnesses, compiler/
stage1 sources, path cases and a compiling rule mutant pass through the proposed
scratch overlay. Full original fixture parity fails because the shared Go
oracle forces TS on JSX and capture drops filenames; the stage1 parser also
refuses JSX. The other two new rules remain blocked on JSX. No full-parity port
is claimed. Exact commands, rates and independent proofs are in
[the third-batch report](wave1-12-batch3-evidence/REPORT.md).

## Fourth batch, October 7

Previous work pushed through e50c08c4. Fetched main remains ef3d907e.
320 origin refs, 39 unique claim Markdown blobs and main .a/.ts lint sources
were inspected. All helper-ready names occur in claims. The first three
remaining inventory entries with needs_type_information=false are claimed:

1. @typescript-eslint/no-non-null-asserted-optional-chain
2. @typescript-eslint/no-this-alias
3. @typescript-eslint/no-unnecessary-type-constraint

This claim is pushed before new code.

### Fourth-batch outcome after Ahra's correction

This-alias implementation e41b30da uses temporary .ts modules as authorized.
Forty original cases with filenames/options preserved, the option matrix and
218 compiler/stage1 files agree across Go, Node, emitted JavaScript and sanitized
native through the previously proposed scratch overlay. Its compiling semantic
mutant is caught by comparison on all three backends; own lint/format passes.

The optional-chain and constraint rules are blocked because the unchanged shared
serializer rejects their legitimate suggestion subranges. Dropping those edits
would break parity. No shared files are edited and no additional claim is made.
Work stops at those exact blockers, as instructed. Full evidence and rates are
in [the fourth-batch report](wave1-12-batch4-evidence/REPORT.md).

## Fifth batch, October 7

Previous claimed work is pushed through 97b9922b. The two suggestion rules are
implemented in 1fe40525 and a93ba8ef and compare byte for byte against the exact
published suggestion-capable harness. The five JSX claims now have tested
independent decision candidates in 6d7aded4; extraction remains explicitly blocked.
See the optional-chain rule REPORT.md for full evidence and limits.

Fetched main remains ef3d907e. All 339 origin refs and 52 unique claim Markdown
blobs were searched along with main lint .a/.ts sources and rule descriptors.
All 46 helper-ready names are unavailable. In the inventory array order,
the first remaining needs_type_information=false names are claimed here:

1. better-tailwindcss/enforce-shorthand-classes
2. better-tailwindcss/no-concatenated-classes
3. better-tailwindcss/no-conflicting-classes

Owned directories are better-tailwindcss-enforce-shorthand-classes,
better-tailwindcss-no-concatenated-classes and
better-tailwindcss-no-conflicting-classes. This claim is pushed before new code.
