Built: a 67-stop rule map, two adaptation-20 patterns, and 13 Node-held .a witnesses; parser checker stops fall to 53.
Commits: 6eba289c (builder assignments) and fac53684 (nullable contextual targets), from main 031a1259.
Commands: full apply exits 0; both full oracles have 106366 passing/1 sanctioned failure/0 pending; both unchanged lane checks PASS; parser stops 67 to 53.
Mutants: wrong builder initializer, wrong question-token declaration, 13 Node output mutations, 26 measured-header mutations, and three emitted-byte comparison negative controls.
Uncovered: native parser execution and 53 remaining stops; internal declaration output changes, while published API bytes stay identical.

The input is f053ef44a3b0ee7c52fcc3e2e7e3caa035918228's project-entry
comparison, not the .a driver's separate loader refusal. TypeScript is pinned
to v6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8. The publication branch
starts at 031a1259bc7973934792dc6cb1bd4074fc2204b9. No compiler source,
shared apply/oracle/lane code, or upstream source is committed by this unit.

[STOPS.md](STOPS.md) maps all 67 original locations to all four requested
adaptation families, an owner, a disposition and a Node witness. Its JSON
counterpart retains full diagnostic chains, source expressions, contextual
types, declaration kinds and owner coordinates. Original coordinates are from
the adapted input; type insertions can move later columns. The stock compiler
API resolves 59 owner records. Eight required-result/call or nested relation
sites retain their full expression and message rather than claiming a direct
optional property declaration resolution. Named-contract lookups are labeled
separately from contextual symbol resolutions.

The two patterns remain inside adaptation 20's present-undefined declaration
rule. The first resolves the two reviewed builder state assignments independently
of the first incompatible structural member. The second examines the non-null
object constituent of a nullable contextual target. Both use the existing
optional-owner, compatible-value, indexed-seed and generic exclusion checks.
The new nullable pattern additionally excludes public declarations. Public
present-undefined slots may be truthful widening candidates, but editing them
would change the API bytes this unit explicitly requires unchanged.

Three builder owners and seven nullable-context owners change in the final tree.
At numeric step 20, full apply selects 387 declarations versus main's 373.
The four additional early selections are AutoGenerateInfo.prefix/suffix,
AllDecorators.parameters (already covered by 32's handoff ledger), and
SymbolVisibilityResult.errorModuleName (already covered by 75). They converge
to main's final annotations. Thus 14 early selections are not 14 new final
source edits. [apply20-comparison.json](apply20-comparison.json) records every
added owner and both actual totals. The late replay
and exact nullable-owner ledger are independent selection evidence, not an
inference from diagnostic codes. Full apply's earlier diagnostic population is
different from the late replay; declaration totals from those runs are not
subtracted to estimate the source delta.

Observed parser result: **67 to 53**, 14 original identities clear, zero new
identities. The new first stop is **checker.ts:1683:9 TS2322**, the NodeBuilder
signature result's optional typeArguments relation. Ten clearances are the
selected optional-code sites; four are incidental shared-owner clearances:
checker.ts:46159 TS2345, checker.ts:54329 TS2420, and resolutionCache.ts:1282
and 1297 TS2322. [parser-comparison.json](parser-comparison.json) retains both
complete results. The compiler is the never-pushed main-plus-stricter-options
scratch used in the input survey, index tree
6be232f53831a0d1faf8def593ebb8b382c73862. Both runs use original parser.ts,
its project config, and the same external Node 25.3.3 declarations. These are
checker-rejected builds; no native execution or lowering clearance is claimed.

| Diagnostic | Before | After | Cleared |
|---|---:|---:|---:|
| TS2322 | 9 | 7 | 2 |
| TS2345 | 7 | 6 | 1 |
| TS2375 | 16 | 12 | 4 |
| TS2379 | 10 | 4 | 6 |
| TS2412 | 24 | 24 | 0 |
| TS2420 | 1 | 0 | 1 |

The 53 remaining sites comprise 35 explicit rule declines, 13 outside the
optional adaptation's seed codes, four public API compatibility constraints,
and one nested generic environment contract needing a new structural rule.
Methods cannot union the callable with undefined without changing declaration
kind/variance. Optional tuple-label indexing cannot be made required. A reset
through generic T is not proven by widening its base. A fileWatcher union has
both required and optional owners. The readonly and optional-widening adapters
do not repair those value/presence contracts merely because they share a file.

The indexed-read rules 30–33 do not cover the remaining tuple-label read:
31 already owns checker.ts, but namedMemberDeclarations?.[i] can legitimately
be absent. Adding an assertion would invent a required label. The other stops
are optional contracts or nullable relations, rather than direct indexed-read
sites. Rule 70's readonly views preserve these undefined/presence relations;
it supplies no repair. No chain among the 67 targets rule 75's named Type caches, JSON configuration
metadata or errorModuleName slots. Consequently neither a
file-list extension nor a pattern extension for those rules is supported.
Every row's JSON contains the separate explanations for 20, 30–33, 70 and 75.

The .a witnesses reproduce the structural causes and are linked per stop.
They are minimal relation witnesses, not claims that a real compiler path is
reachable with the same tiny input. Each ran from source on Node 24.19.0 and
was checked by the real current-main Adamic compiler to set its first-line
header. All thirteen Node runs exit zero with the recorded bytes. Thirteen changed
initializers/reads change those bytes and fail the comparator. Header controls
exercise the recorded-diagnostic predicate, not a claimed independent Gate.aCheck
invocation. No internal/oracle runtime fixture was added; counts.md records
these scout witnesses locally.

The builder mutant replaces buildInfo.outSignature with true on real source;
the value compatibility guard rejects it before edits and restores the bytes.
The nullable mutant changes the real internal questionToken annotation to
number; its ledger has five edits instead of the required seven and omits the
question-token owner. The exact-owner assertion catches it. Both mutants live
only in external scratch. verify.py recursively compares all 10 JavaScript files and 715 declarations.
All 110 published/library declarations and 601 internal declarations are
byte-identical to clean main. Four internal artifacts change: the bundled
typescript.internal.d.ts and the per-source compiler/builder.d.ts,
compiler/resolutionCache.d.ts and compiler/types.d.ts. These encode the six
exported internal contract changes; four local checker declarations do not
emit. Their exact diffs are internal-api.diff and internal-module-api.diff.
An unreviewed internal difference fails the comparator. JavaScript, published
API and unreviewed internal API byte mutations are all caught. Full equality
of every internal API artifact is therefore not claimed or achievable while
those internal declarations are corrected.

A contaminated early control was discarded: a replay overlapped its apply.
Its oracle was stopped with SIGTERM and is not counted as valid evidence.
The replacement control uses an immutable archive of main's entire stage3
folder, with separate output and logs. It receives no experimental writes.
The initial full lane uses this branch's unmodified shared runners. It failed
three 40-second timeouts (largeControlFlowGraph,
codeFixClassImplementInterfaceNoTruncation and reallyLargeFile), plus the
sanctioned API-baseline failure; no additional baseline differences appeared.
The first clean-main oracle was interrupted by an environment reconnect that
killed its parent, leaving orphan test workers; those were stopped. Neither
run is substituted for a passing lane. Both are preserved as failure evidence.
The detached measure-gates.py then runs clean main and the extension sequentially,
with fresh full oracle outputs and the unchanged lane checker. Their apply
exit codes are from the independently completed full applies, not invented
reruns. Shared source hashes and final measured gate reports are preserved.
The first isolated extended run was also rejected: 106,360 passing, two
failing, zero pending, with a 40-second relationComplexityError before-all
hook timeout in addition to the sanctioned API failure. It used the prior
run's .parallelperf.json (time-based 90% batching); clean main had no timing
cache and used file-size batching. The fresh extended repeat preserves and
removes only that transient scratch cache, then uses the same unmodified
oracle and eight workers. No source, timeout, test filter, expected count or
sanction is changed. All failed observations remain in evidence.

Final comparison: both complete applies exit **0**. Both full oracles run
all runners, no test filter, eight workers on linux/x64/Node v24.19.0; install
and build exit **0**, tests and the oracle exit **1** for exactly the sanctioned
public-API baseline failure. Both have **106,366 passing, one failing, zero
pending**. Both unchanged landing checks report **PASS stage3 landing lane**,
exit **0**. The API composition is unchanged: 28 reference declarations, 222
composed/sanctioned declarations, 194 added and 193 removed diff lines. The
complete baseline diff is byte-identical, SHA-256
3da14ced9f2c7eea0ae8b4dabbe6067a91510b0ab0d950ddb83903a9eb50fa28.
Main's test phase takes 543.796 seconds; the fresh extension takes 538.885.
[gate-comparison.json](gate-comparison.json) retains the exact common result.
The comparator also rejects the first timed-out lane as unequal to main.

All 75 shared apply/oracle/lane files are byte-identical to archived main.
The full emitted-artifact comparator covers 725 files (10 JavaScript, 715
declarations), with exactly the four reviewed internal declaration differences
listed above. This proves published API equality, not equality of internal
compiler contracts that the unit deliberately corrected. Compressed raw logs,
phase exits, patch tables, API diffs, both failed lanes and the interrupted
control live in evidence/, whose manifest hashes their uncompressed bytes. The exact CRLF-preserving
source diff is evidence/source.patch.gz (`gzip -dc` to read/apply).
No incomplete or timed-out observation is used as the passing result.

Setup passed with GOPROXY=https://proxy.golang.org|direct. Timing lines:
Node 0.072s, Go 0.071s, markdown dependencies 0.180s, submodules 0.192s,
clang 0.474s, Go build 101.040s, cache warm 101.525s, total 101.626s.
`nproc` is 5; cpu.max is 400000 100000. The isolated control/repeat shells source
/workspace/adamic-tools/env.sh. Both initial and isolated runs record
Node 24.19.0. Go 1.27.1 and clang 20.1.8 are the setup toolchain.

Principal commands, all test output redirected to logs:

```sh
source /workspace/adamic-tools/env.sh
bash /workspace/cache/remaining67-baseline-tools/stage3/apply.sh /workspace/cache/remaining67-main-clean > /tmp/remaining67-main-clean-apply.log 2>&1
bash stage3/lane/run.sh /workspace/cache/remaining67-final-lane > /tmp/remaining67-final-lane.log 2>&1
/workspace/cache/checker-stops-topic build /workspace/cache/checker-stops-adapted/src/compiler/parser.ts -o /workspace/cache/remaining67-before-parser > /tmp/remaining67-before-parser.log 2>&1
/workspace/cache/checker-stops-topic build /workspace/cache/remaining67-measure-after/src/compiler/parser.ts -o /workspace/cache/remaining67-after-parser > /tmp/remaining67-after-parser.log 2>&1
python3 stage3/scouts/step24/remaining-67/check-fixtures.py /workspace/cache/remaining67-adamic > /tmp/remaining67-fixtures-final.log 2>&1
TSC_ADAPT_TYPESCRIPT=/workspace/cache/tsc-census/npm/node_modules/typescript/lib/typescript.js python3 stage3/scouts/step24/remaining-67/extension-mutant.py /workspace/cache/remaining67-after > /tmp/remaining67-extension-mutant.log 2>&1
python3 stage3/scouts/step24/remaining-67/check-selector.py stage3/scouts/step24/remaining-67/nullable-source-mutant-report.json > /tmp/remaining67-selector.log 2>&1
python3 stage3/scouts/step24/remaining-67/audit.py > /tmp/remaining67-audit.log 2>&1
nohup setsid python3 stage3/scouts/step24/remaining-67/measure-gates.py /workspace/cache > /tmp/remaining67-isolated-gates.log 2>&1 < /dev/null &
nohup setsid python3 stage3/scouts/step24/remaining-67/measure-gates.py /workspace/cache fresh-extended > /tmp/remaining67-fresh-gates.log 2>&1 < /dev/null &
python3 stage3/scouts/step24/remaining-67/gate-comparison.py /workspace/cache/remaining67-isolated-main /workspace/cache/remaining67-isolated-extended-fresh > /tmp/remaining67-gate-comparison.log 2>&1
python3 stage3/scouts/step24/remaining-67/verify.py /workspace/cache/remaining67-main-clean /workspace/cache/remaining67-final-lane/adapted-tree > /tmp/remaining67-emitted-identity.log 2>&1
```

The fresh-extended command runs after the first detached sequence finishes;
it preserves/removes that sequence's timing cache before repeating only the
extended full oracle. These commands require the documented prepared cache
trees and refuse to replace existing result directories.

No whole Go package or full repository gate is used as confirmation. The full
upstream oracle and lane are the requested stage-3 measurements. Stock tooling
and upstream source stay in external caches; only the adaptations, scout tools,
.a witnesses and measured evidence are publication changes.
