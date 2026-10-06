# Syntax-only lint verification

The recovered tree is intact. The final parent-array initialization change is
included in this verification. All changes are under `stage1/cohere/lint/`;
no compiler, runtime, parser, scanner or submodule source is changed.

## Commands and outcomes

Run from the repository after `source /workspace/adamic-tools/env.sh`:

```sh
ADAMIC_LINT_BENCH=1 ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 \
  go test ./stage1/cohere/lint -count=1 -v -timeout 20m \
  > /tmp/lint-restored.log 2>&1

go vet ./... > /tmp/lint-restored-vet.log 2>&1

go test ./internal/oracle -run TestTheOracleCatchesOneByte -count=1 -timeout 10m \
  > /tmp/lint-restored-oracle.log 2>&1

/workspace/scratch/cohere --no-fix --no-cache \
  stage1/cohere/lint/finding.ts stage1/cohere/lint/lint.ts \
  stage1/cohere/lint/main.ts > /tmp/lint-restored-cohere.log 2>&1

gofmt -l stage1/cohere/lint
git diff --check
```

Every command completed with exit 0. Vet, gofmt and diff checks printed nothing.
The filtered oracle printed `ok github.com/system-inc/adamic/internal/oracle`
in 3.688s. Cohere checked three files with 276 rules and reported 100%
Adamic-ready, no findings. The full repository test gate was not run; the full
new slice and the named filtered oracle were run.

The source corpus is TypeScript v6.0.3, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`, with all 77 compiler files, and all
78 current stage1 `.ts` files. The rule corpus contains 217 distinct captured
cohere source/rule/options combinations plus 24 generated runs. Source paths
and temporary case paths are printed, so exact output byte totals can vary
when temporary path lengths change between runs.

The final comparison totals 12,381,452 identical bytes across Go, Node running
the same TS, and sanitized native Adamic. Canonical output includes exact
finding descriptions, positions, message IDs, repair categories and whole fixed
sources. The Go oracle uses cohere's actual formatter and converging edit engine.

## Mutants and gap proof

Each scratch mutant compiled and ran successfully, then failed the comparison
on both Node and ASan/UBSan native:

| Mutant | What caught it |
| --- | --- |
| Apply an eqeqeq suggestion as an automatic fix | repair category and fixed-source comparison |
| Report an empty function body | finding-position and finding-list comparison |
| Invert duplicate-case membership | finding-position and finding-list comparison |

`TestNestedConstructorGap` requires Node's `1` and stage 0's false constructor
refusal at line 7, column 16. See `GAPS.md` and its 19-line proving program.

## Measurement

Five interleaved rounds on 77 compiler files yielding 451 findings. Count mode
includes process startup, reading, parsing and all five rule visitors, excluding
finding formatting and the repair phase. Native is an unsanitized clang -O2
build; comparisons and mutants use sanitizers with leak checking. Go is 1.27.1,
clang 20.1.8, Node 24.19.0. Machine and load are recorded below. Native is 4.39
times slower than Go and 1.34 times slower than Node on this run.

Original toolchain setup reported Go/clang/Node/submodules ready in 0s each,
build cache warm in 10s, done in 10s. `nproc` was 5; the CPU allocation was four
cores (`cpu.max: 400000 100000`). Shared-machine timing is an observation, not
a performance guarantee.

Selected final log output:

```
--- PASS: TestNestedConstructorGap (0.10s)
lint_test.go:258: cohere cases: 217 unique source/rule/options combinations
lint_test.go:283: Go, Node, native identical: 128446 bytes
--- PASS: TestRulesAgree (17.89s)
lint_test.go:317: compiler and stage1: 155 files
lint_test.go:318: Go, Node, native identical: 12253006 bytes
--- PASS: TestCompilerAndStage1Agree (28.66s)
--- PASS: TestMutants (39.22s)
--- PASS: TestMutants/suggestion_applied_as_fix (13.56s)
--- PASS: TestMutants/empty_function_body_reported (12.17s)
--- PASS: TestMutants/duplicate_case_suppressed (11.49s)
lint_test.go:410: machine Linux 6d3a2813751b 6.18.44 #1 SMP Sat Sep 26 20:02:31 UTC 2026 x86_64 GNU/Linux; model name	: AMD EPYC 9V74 80-Core Processor; load before 0.86 0.31 0.23 1/189 118278
lint_test.go:441: best of 5 Go: 0.222345s, 2028.37 findings/s (77 files, 451 findings)
lint_test.go:441: best of 5 native: 0.976100s, 462.04 findings/s (77 files, 451 findings)
lint_test.go:441: best of 5 Node: 0.730318s, 617.54 findings/s (77 files, 451 findings)
lint_test.go:444: load after 0.88 0.34 0.23 2/189 118361
--- PASS: TestThroughput (16.09s)
ok  	github.com/system-inc/adamic/stage1/cohere/lint	101.969s
```

## Scope limits

Five syntax-only rules only. No general config/suppression integration, type
information, additional rules, JSX or malformed-source recovery is claimed.
Overlapping native repairs stop explicitly; this is not a general competing-fix
engine. The compiler checkout is scratch-only and is not committed.
