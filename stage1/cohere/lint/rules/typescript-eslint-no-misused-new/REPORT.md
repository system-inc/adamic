Built: @typescript-eslint/no-misused-new in owned .a modules; the other two claimed rules have proven shared repair-range blockers.
Commits: pre-code claim 6dce77a4; implementation 8b003543; evidence commit follows.
Commands and outputs: four-way validation PASS 120.036s; all-rule rerun PASS 22.270s; vet and filtered oracle PASS; setup 20s, nproc 5.
Mutant: skipping getters compiles and exits cleanly, then fails only Go comparison on Node source, emitted JavaScript and sanitized native.
Not covered: the two blocked ports, their mutants/throughput, the full repository gate, and ordinary production registration without .a compatibility.

## Selection and published claim

Pushed the existing branch first: Everything up-to-date at 18449284. Then fetched
all origin heads, without recursive submodule fetching. Scanned 310 origin refs.
All helper-ready names remain represented in direct claim records, including
existing-origin-port skips covered by the original instruction. Reports and logs
are not treated as claims.

The next three eligible entries from the inventory's `syntax ready for AST/API
adaptation` wave, in inventory order, were:

1. @typescript-eslint/no-extra-non-null-assertion
2. @typescript-eslint/no-misused-new
3. @typescript-eslint/no-unnecessary-parameter-property-assignment

None was selected by an implementation on main or named in an origin claim before
this update. Claim 6dce77a4 was committed and pushed before implementation. The
selection audit reconstructs our own claim file at its prior revision 18449284,
so it verifies the exact first three without counting our new claims against us.
The main and inventory pins and every matching claim record are in
[evidence/selection.json](evidence/selection.json).

## Completed implementation

The directory owns rule.json without order, rule.a, messages.a, the real upstream
oracle adapter, mutant.json targeting rule.a and a raw TypeScript witness. No
shared dispatch, registry, context, finding model, compiler or cohere source was
edited. Every new Adamic module has the .a extension.

The three judgments remain distinct:

- An interface construct signature returning its own written bare name reports
  the three-byte `new` keyword. Type-literal construct signatures remain silent.
- A method signature with the static identifier `constructor` reports the key,
  including signatures in type literals. Quoted and computed keys stay silent.
- A bodyless named-class method or getter named `new` returning the class's
  written bare name reports the key. Class expressions are included; setters
  remain excluded, including the upstream parser-recovery fixture.

Return names are compared by syntax, not binding resolution. Type arguments do
not change the bare name. Qualified, parenthesized, array and other return types
remain silent. A method body makes the final child a body rather than a return
annotation, so it is excluded. Exact descriptions are copied from the pinned Go
rule. This rule has no options, fixes or suggestions.

## Four-way findings and fixed-source comparison

The independent oracle uses unmodified Go cohere
715ba94f3608a6500086b1076ce5cb7e51b836db, its canonical formatter and converging
fixer. Original tests run through the capture overlay, including direct range
assertions. The complete capture for no-misused-new has 45 distinct source/options
combinations. No case for this rule is excluded.

The same manifests run on Go, Node source, emitted JavaScript and native with
ASan/UBSan and leak checking. Every successful side must exit 0 with no stderr
before its output is compared byte for byte.

| Comparison | Findings | Identical bytes on all four sides |
|---|---:|---:|
| 45 upstream cases plus selected-rule witness | 16 | 8,902 |
| Same cases plus selected and all-rule witness | 19 | 10,409 |
| Entire compiler and stage1 corpus | 0 | 12,345,503 |

The corpus is all 77 compiler files at TypeScript commit
050880ce59e30b356b686bd3144efe24f875ebc8 and all 141 stage1 .ts/.a sources, including
generated registry sources. Reports, diagnostic byte ranges, repair metadata and
fixed source all match. Capture also executes older registered rules and logs the
previously documented module-alias JSX exclusion; that fixture belongs to neither
this port nor either new repair blocker.

The main run passes all third-batch fixtures, corpus, mutant, throughput and
blocker checks in 120.036s. The subsequent fixture rerun adds explicit all-rule
dispatch coverage and passes in 22.270s. Only test comments and that additional
witness row changed between these runs; the port implementation did not change.

## Semantic mutant

Mutant: remove GetAccessor from the class member filter. The witness's
`declare abstract class C { get new(): C; }` loses its classNew finding, while
both interface findings remain as positive controls. The variant builds and runs
with exit 0 and no stderr on Node source, emitted JavaScript and sanitized native.
All three disagree with the independently executed Go oracle. Compilation or
sanitizer failures are not counted as caught mutants.

## Findings per second

The natural compiler/stage1 corpus has zero findings. Throughput therefore adds
500 copies of the three-finding witness: 1,500 findings over 219 files. This is a
synthetic supplement, not a natural compiler findings-rate claim.

Five interleaved Go/native/Node rounds must agree on the full count. The table
uses each runtime's best elapsed time, including startup, file reads and parsing.
Native timing uses a release build; parity and mutant checks use sanitizers.

| Runtime | Best seconds | Findings per second |
|---|---:|---:|
| native | 1.330574 | 1,127.33 |
| Node source | 0.905665 | 1,656.24 |
| Go cohere | 0.226439 | 6,624.31 |

The full timing rounds are in evidence/validation.log. These observations do not
show Adamic outperforming Go.

## Proven repair-range blockers

Both rules remain claimed and blocked, with no partial listener or placeholder
descriptor. The owned blocker tests build the current independent serializer
with a virtual selection of each real upstream rule. Count mode first succeeds
with exactly one finding, proving valid parsing and listener execution. Full
serialization then exits 2 at the exact shape guard:

| Rule | Input | Observed refusal |
|---|---|---|
| no-extra-non-null-assertion | `const r = foo!!;` | `panic: unexpected fix shape` |
| no-unnecessary-parameter-property-assignment | `class C { constructor(public foo: string) { this.foo = foo; } }` | `panic: unexpected suggestion shape` |

Observation: the assertion rule reports the whole inner NonNullExpression but
fixes only its trailing `!`. The serializer requires a fix range equal to the
diagnostic range. Linter.fixed applies diagnostic start/end and ignores the
existing editStart/editEnd fields; overlapping diagnostic spans also cannot
stand in for disjoint bang edits on a triple assertion.

Observation: the parameter-property rule's suggestion range extends one byte
past the diagnostic, deliberately consuming the semicolon or another trailing
byte. The shared serializer requires equality with the diagnostic range and
cannot expose that independent edit range. Ignoring the extra byte would pass a
findings-only comparison while producing the wrong suggested rewrite.

All original Go tests for both rules pass separately in 0.037s, including exact
fixed sources and applied suggestions. This evidence is in blocked-upstream.log.
No mutant, Node/emitted-JS/native parity or throughput is claimed for a blocked
rule whose implementation was not built.

Inference: full ports require shared independent repair-range serialization and
fix application support. CLAUDE.md says "Never edit a dispatch, oracle, corpus or
copied-file list." The user's territory is each owned rule directory. No shared
production change was made to work around those constraints.

## Reproduction and bounded gate

Use `source /workspace/adamic-tools/env.sh` in each shell. Go 1.27.1,
clang 20.1.8, Node 24.19.0; nproc 5; cpu.max 400000 100000; memory 17.6 GB.

validate.py creates the same temporary compatibility overlay used by the earlier
ports: .a discovery/imports/mutants/copying and the inherited profile_test.go
range-over-function typo. It injects this owned suite into the parent package and
reuses earlier owned emitted-JS/corpus helpers. No production .a integration is
claimed. The already committed concrete patch remains at
../nexus-import-require-module-alias/integration.patch, unapplied.

All test output was written directly to files, never piped:

```
ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-14-typescript-pinned python3 stage1/cohere/lint/rules/typescript-eslint-no-misused-new/validate.py > /tmp/lint-wave1-14-third-run.log 2>&1
# Runs go test -overlay=<printed overlay> ./stage1/cohere/lint -run '^TestWave14Third' -count=1 -v -timeout=20m
GOFLAGS=-overlay=<printed overlay> bash cloud/setup.sh > <owned>/evidence/setup.log 2>&1
go vet -overlay=<printed overlay> ./... > <owned>/evidence/vet.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -timeout=10m > <owned>/evidence/oracle.log 2>&1
go test -overlay=<printed overlay> ./stage1/cohere/lint -run '^TestWave14ThirdRules$' -count=1 -v -timeout=10m > <owned>/evidence/all-rules.log 2>&1
# From cohere/:
go test ./internal/lint/rules/typescript -run '^(TestNoExtraNonNullAssertion|TestNoUnnecessaryParameterPropertyAssignment)' -count=1 -v -timeout=10m > <owned>/evidence/blocked-upstream.log 2>&1
```

Vet exits 0 with an empty log. The filtered external oracle passes in 0.089s.
Setup uses the known profile compatibility workaround and succeeds: Go ready 0s,
clang ready 0s, Node ready 0s, submodules ready 0s, cache warm 20s, done 20s on
5 processors. The full repository gate was not run. This is the touched lint
package's focused gate, repository vet and a filtered external oracle.
