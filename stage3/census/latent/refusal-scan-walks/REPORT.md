# Refusal scan investigation, task 2bz5m7v

The hot cost belongs in Lower's semantic queries. There is already one shared
per-node rule dispatcher. The registry does not walk each file once per rule,
and the census does not rerun the refusal scan per unit. No production or
census behavior has been changed. The initial source-only report was pushed
first in f3f18f495, before adding the profiling probes.

Compiler base: fetched origin/main 42dde46209e94bbb5c8bea00593272856fad6882.
Input: TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8, with this base's
complete adaptation set, through src/tsc/tsc.ts. Entry expansion yields 81 own
source files. This is a current-main measurement, not a reproduction of Pupa's
historical whole-census CPU percentage. The enum helpers changed after a47d7504;
the registry and shared refusal dispatcher did not.

## What walks, and when

- `stage3/census/latent/lower.go.txt:259` starts the file loop. It constructs a
  fresh scan lowering at 263-265, calls `latentRefuse(file)` at 273, and only
  then starts the independent unit loop at 275. The later `latentRegister`
  at 312 registers declarations for lowering attempts, not refusal rules.
- `internal/lower/refusal_rules.go:33` installs a wrapper in an init-time
  registry. Its `register.func1` calls the registered check at line 35.
  Registration is not a runtime file walk. Rules register from their own init
  functions, for example `library_string.go:411` and `void_widening.go:45`.
- `internal/lower/refusals.go:104` dispatches the ordered registry at each node
  of one visitor; its recursive child walk is at 114 and file-root call at 118.
  Contracts have a separate visitor and root call at 90. Module preflights at
  56-74 and rule helpers can perform additional walks.
- `stage3/census/latent/refusalrewrite/rewrite.go:139` removes the original
  child walk. Lines 157-169 install one deferred child walk per supported
  visitor, with continuation and diagnosed-body guards. `make_overlay.py:70`
  invokes that transformer; it does not split the registry into separate walks.
  `full.go.txt` has state-copy/program-fact helpers, not another per-unit
  refusal registration loop.

Thus each file has one outer refusal call, with the two core visitors plus
preflights. There is no universal single total-walk count: some Lower helpers
restart file/program walks per query. These are independent of unit count.
Production likewise calls refuse per module after its normal setup
(`internal/lower/lower.go:93`); its registered checks invoke the same helpers.
The census's fresh-per-file state also repeats lowering-local program preflight
memos across files. That lifecycle difference exists, but is not the dominant
registry cost measured here.

## Plain profiles on one box

The probe loads/checks the entire entry, applies the census entry-closure hook,
then calls the actual generated first-error `latentRefuse` on one selected file.
It attempts no units. Timers exclude loading, compilation and evidence encoding.
The runs are sequential, after setup, adaptation and both builds completed.
[REPRODUCE.md](REPRODUCE.md) describes the exact instrumentation and limits.

| File | Source bytes | Scan wall s | Scan CPU s | Scan findings |
| --- | ---: | ---: | ---: | ---: |
| utilities.ts | 515,338 | 125.549 | 133.305 | 252 |
| checker.ts | 3,155,803 | 664.332 | 693.892 | 1,503 |

Selected sampled cumulative CPU seconds from the uninstrumented profiles:

| Function | utilities.ts | checker.ts |
| --- | ---: | ---: |
| registry register.func1 | 107.21 | 634.63 |
| Node.ForEachChild | 113.02 | 636.61 |
| wholeNumericEnums | 98.44 | 586.52 |
| numericEnum | 96.18 | 571.27 |
| stringWideningRefusal | 52.98 | 421.23 |
| refuseVoidWidening | 33.44 | 120.94 |

Cumulative values include descendants and overlap. They are not additive walk
costs. `Node.ForEachChild` itself has only 0.14 / 0.08 s flat CPU; the registry
wrapper has 0.02 / 0.04 s flat CPU. Its name identifies the closure created by
register, not repeated execution of registration. Deep recursive stacks can
truncate outer ancestor attribution, so use the raw profiles and explicit
per-rule timers rather than adding cumulative frames.

Registry-focused profiles put 98.00 s / 584.00 s under wholeNumericEnums,
91.4% / 92.0% of the registry's sampled cumulative CPU. See
[utilities](evidence/registry-utilities-top.txt) and
[checker](evidence/registry-checker-top.txt). This attributes the cost to Lower's
semantic helper even when samples outside the registry are excluded.

Raw profiles, top tables and timer JSON are in [evidence/](evidence/).

## Exact walks and per-rule times

Both counted scans passed and matched the plain scan's complete finding-array
SHA-256, including order and all finding fields. Counts are 252 for utilities.ts
and 1,503 for checker.ts. The hashes are in [validation.json](evidence/validation.json).
This validates the counters, not an overlay optimization or whole-census identity.

| Counted scan metric | utilities.ts | checker.ts |
| --- | ---: | ---: |
| Scan wall s | 121.341 | 656.728 |
| Scan CPU s | 128.478 | 686.038 |
| Registered rules | 47 | 47 |
| Rule checks | 2,554,435 | 14,016,163 |
| Selected file root child-walk calls | 27 | 14 |
| All context file-root child-walk calls | 2,146 | 2,059 |
| All child-walk calls | 18,873,237 | 8,434,032 |

The two core visitor root calls each occur exactly once in each scan. Helpers
and preflights account for the other root calls. The complete file ledger covers
203 source identities, including 122 declaration files. A root call can short
circuit; these are not counts of completed full-file traversals. No all-81-file
census aggregate is inferred from these two isolated scans.

Selected rules, with inclusive instrumented wall time:

| File | Rule | Calls | File-root walks | Child-walk calls | Wall s |
| --- | --- | ---: | ---: | ---: | ---: |
| utilities | stringWideningRefusal | 54,260 | 0 | 0 | 51.969 |
| utilities | refuseVoidWidening | 54,213 | 0 | 0 | 34.275 |
| utilities | castRefusal | 54,383 | 276 | 33,120 | 12.324 |
| utilities | predicateArgumentsRefusal | 54,429 | 526 | 5,540,466 | 1.387 |
| utilities | enumRefusal | 54,428 | 81 | 921,555 | 0.663 |
| utilities | recordDivergenceRefusal | 54,429 | 163 | 1,896,246 | 0.197 |
| checker | stringWideningRefusal | 297,860 | 0 | 0 | 419.359 |
| checker | refuseVoidWidening | 297,449 | 0 | 0 | 124.480 |
| checker | castRefusal | 298,309 | 1,098 | 131,760 | 46.570 |
| checker | predicateArgumentsRefusal | 298,760 | 162 | 68,959 | 0.353 |
| checker | enumRefusal | 298,739 | 81 | 931,297 | 7.137 |
| checker | recordDivergenceRefusal | 298,760 | 0 | 30 | 0.107 |

Every rule, including zero-walk rules, is in [rules.csv](evidence/rules.csv).
[files.csv](evidence/files.csv) gives every observed file-root count by rule and
[walk-sites.csv](evidence/walk-sites.csv) gives every Lower call site. Timings
include counter overhead and nested checks; they are not disjoint CPU buckets.
The outer bucket has walk counts only and blank timing/call-count CSV cells.

The hottest rules do not call ForEachChild themselves at all. utilities.ts has
more counted child walks than checker.ts while consuming far less CPU. Real
helper restarts include `predicates.go:583` (predicateCallbackCallers),
`records_divergence.go:134` (recordPrototypeKey), `interface_cast.go:146`
(view) and `records_observation.go:69` (recordReadDiscarded). They are Lower's
queries, not registry registration. These walks and their rule/caller files
were inspected alongside the registry, transformer, README and full overlay.

## The Lower handoff

`internal/lower/enums.go:64` loops over union members. At 74 it calls
`numericEnum(identity)` for every numeric enum member, even when many members
have the same identity. `numericEnum` at 98 loops over the enum's entire
member declaration at 102, calling checker.GetConstantValue at 103 every time.
It has no memo. For a whole numeric enum of m members this can do m squared
constant-value queries for one wholeNumericEnums query, repeatedly at many
structural type-pair queries.

The hot callers include string widening (`library_string.go:401`,
`invariance.go:61` and 67) and void widening (`void_widening.go:344`, 383-384).
This explains why substantial semantic work appears under the generic registry
wrapper and the common AST traversal. It is not evidence that every rule owns
a separate full-file visitor.

Route the fix to Lower: lazily memoize numericEnum by enum symbol, including
false answers, and consider checker/type-keyed summaries for whole/open numeric
enum membership. Keep the first-query timing and existing heterogeneous-enum,
alias/deduplicated-value, whole/partial-union and impurity semantics. Run Lower's
enum, widening and arm-filter correctness tests and remeasure these same inputs.
A one-walk dispatcher already exists; adding another traversal-fusion layer in
the overlay would not remove this cost. The overlay must continue calling the
same rules in the same order. No internal/lower or existing overlay source has
been edited, and no speedup or whole-census finding-identity proof is claimed.

## Validation and limits

All four isolated scan tests exited 0. The probe calls t.Parallel first, has a
repository-unique name, and invokes no Go subprocess. Plain and counted finding
hashes match for both files. The complete 746-file source manifest was verified
after the runs. [provenance.json](evidence/provenance.json) pins the compiler,
checker/submodules, binaries, source and box (four-CPU quota, 16 GiB memory cap).
All started setup, adaptation, compilation, profile and analysis processes have
finished; the live-task-process audit is empty. Retained task files stay below 50.

The first test build included ordinary lower tests, which reference entry helpers
that make_overlay deliberately removes. The profiling-only manifest now excludes
those test declarations. A first counter build exposed two embedded Node value
receivers; the AST instrumenter now preserves their implicit address-taking. Both
final probe binaries built and all four runs passed. Failed and final build logs
are retained compressed. The interrupted full-history source clone was stopped;
the successful shallow checkout has exactly the requested source commit.

Only files under stage3/census/latent/refusal-scan-walks are changed. No existing
census overlay behavior or internal/lower code was altered. This completes the
measured Lower handoff; implementing that semantic cache belongs to Lower.
