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
  at 323 registers declarations for lowering attempts, not refusal rules.
- `internal/lower/refusal_rules.go:33` installs a wrapper in an init-time
  registry. Its `register.func1` calls the registered check at line 35.
  Registration is not a runtime file walk. Rules register from their own init
  functions, for example `library_string.go:411` and `void_widening.go:45`.
- `internal/lower/refusals.go:104` dispatches the ordered registry at each node
  of one visitor; its recursive child walk is at 114 and file-root call at 118.
  Contracts have a separate visitor and root call at 90. Module preflights at
  56-74 and rule helpers can perform additional walks.
- `stage3/census/latent/refusalrewrite/rewrite.go:139` removes the original
  child walk. Lines 166-179 install one deferred child walk per supported
  visitor, with continuation and diagnosed-body guards. `make_overlay.py:79`
  invokes that transformer; it does not split the registry into separate walks.
  `full.go.txt` has state-copy/program-fact helpers, not another per-unit
  refusal registration loop.

Thus each file has one outer refusal call, with the two core visitors plus
preflights. There is no universal single total-walk count: some Lower helpers
restart file/program walks per query. These are independent of unit count.
Production likewise calls refuse per module after its normal setup
(`internal/lower/lower.go:96`); its registered checks invoke the same helpers.
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

Raw profiles, top tables and timer JSON are in [evidence/](evidence/).

## Exact walks and per-rule times

The counted utilities scan completed with the identical 252-finding SHA-256:
`f204600548bb89c678fdb2fc21d71fef3722e69bf48bc98a66d5aba442ecb65c`.
Its scan used 121.341 s wall and 128.478 s CPU. This is an instrumentation
comparison, not an optimization or a claimed speedup.

There are 47 registered rules. The two core visitor root calls each occur once.
Across all Lower helper/preflight calls in this one selected scan, utilities.ts
is the receiver of 27 file-root child walks; all context files together account
for 2,146 such calls. A root call can short circuit, so these are not counts of
completed full-file traversals.

| Rule, utilities.ts | Rule calls | File-root walks | All child-walk calls | Inclusive instrumented wall s |
| --- | ---: | ---: | ---: | ---: |
| stringWideningRefusal | 54,260 | 0 | 0 | 51.969 |
| refuseVoidWidening | 54,213 | 0 | 0 | 34.275 |
| castRefusal | 54,383 | 276 | 33,120 | 12.324 |
| predicateArgumentsRefusal | 54,429 | 526 | 5,540,466 | 1.387 |
| enumRefusal | 54,428 | 81 | 921,555 | 0.663 |
| recordDivergenceRefusal | 54,429 | 163 | 1,896,246 | 0.197 |

The hottest rules do not call ForEachChild themselves at all. Real helper
restarts include `predicates.go:583` (predicateCallbackCallers),
`records_divergence.go:134` (recordPrototypeKey), `interface_cast.go:146`
(view) and `records_observation.go:69` (recordReadDiscarded). They are Lower's
queries, not registry registration. The latter walks were inspected with their
rule/caller files, along with the registry, transformer, README and full overlay.

The counted checker scan is still running at this checkpoint. The final
per-rule, per-source-file and per-call-site CSV ledgers will include both files.

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
