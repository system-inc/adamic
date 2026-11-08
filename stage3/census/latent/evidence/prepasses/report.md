# Census prepasses, task sygfnmc

The corrected entry census exposes 18 optional-field write findings across 10
files. Eight units become newly stopped and 13 already-blocked first stops move.
Five old findings are masked by earlier stops; raw and distinct views agree.

Branch: `typescript/census-prepasses`, cut on origin/main
`0816e18511b5e1ef4383543ee81c526fc36acfaf`. The tested code checkpoint
`a2d4b5783` was pushed before the long census. No production compiler or adaptation
change is included. Both sides use the same frozen tree from that main apply,
TypeScript 6.0.3 source `050880ce59e30b356b686bd3144efe24f875ebc8`, and the
pinned checker dependency recorded in provenance.json.

Both first-error streams reached 81 source files, with zero checker diagnostics,
4,059 attempted units and 504 generic deferrals. Headers, unit identities and
attempted eligibility match. These are isolated lowering observations, with
first-error masking and generic coverage limits, not native compilation results.


## first-error

| View | Cleared | New | Changed | Newly stopped units | Changed existing first stops | Panics before | Panics after |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| raw_findings | 5 | 18 | 0 | 8 | 13 | 0 | 0 |
| distinct_reads | 5 | 18 | 0 | 8 | 13 | 0 | 0 |

New findings by exact reason and source file:

| Reason | File | Raw | Distinct |
| --- | --- | ---: | ---: |
| writing a possibly absent optional own field | src/compiler/builder.ts | 2 | 2 |
| writing a possibly absent optional own field | src/compiler/checker.ts | 1 | 1 |
| writing a possibly absent optional own field | src/compiler/executeCommandLine.ts | 3 | 3 |
| writing a possibly absent optional own field | src/compiler/factory/emitNode.ts | 4 | 4 |
| writing a possibly absent optional own field | src/compiler/factory/nodeFactory.ts | 1 | 1 |
| writing a possibly absent optional own field | src/compiler/moduleNameResolver.ts | 2 | 2 |
| writing a possibly absent optional own field | src/compiler/parser.ts | 1 | 1 |
| writing a possibly absent optional own field | src/compiler/utilities.ts | 2 | 2 |
| writing a possibly absent optional own field | src/compiler/watchPublic.ts | 1 | 1 |
| writing a possibly absent optional own field | src/compiler/watchUtilities.ts | 1 | 1 |

Cleared is an old finding with no remaining reason at that identity. Changed retains another reason at the identity. New includes replacement reasons. A cleared first-error observation can be hidden by an earlier new stop. Only lowering-phase findings define unit first stops. Raw and distinct are unique identities, not event multiplicity.

## core-full

| View | Cleared | New | Changed | Newly stopped units | Changed existing first stops | Panics before | Panics after |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| raw_findings | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| distinct_reads | 0 | 0 | 0 | 0 | 0 | 0 | 0 |

New findings by exact reason and source file:

| Reason | File | Raw | Distinct |
| --- | --- | ---: | ---: |
| None | | 0 | 0 |

Cleared is an old finding with no remaining reason at that identity. Changed retains another reason at the identity. New includes replacement reasons. A cleared first-error observation can be hidden by an earlier new stop. Only lowering-phase findings define unit first stops. Raw and distinct are unique identities, not event multiplicity.

## Full core scope

Both full-mode runs load the complete tsc entry program and select core.ts alone.
Each has 59 attempted units and 211 deferred generic declarations. Their complete
records, including witness metadata, are byte-identical:
`c5de384152cc9feb6cfe6c12dc0461d36e136c76e2f5ddc25e5eaa1be0f816bb`.
Each records 12 NotYet findings, 1 Refused finding and 8 boundaries, with zero
panics. The full pair has no cleared, new, changed or newly stopped identities.

## Passes, module lists and ownership

All four templates retain collectViewedElements, noteInheritance,
noteAccessorNames and noteOmittedOptionals in production order. The exemption
table beside the AST guard documents the census's conservative nil element key,
enum validation at declaration/use, and the absence of a whole-module execution
schedule in independent units. The element collector is called, but nil-key mode
checks every element read and requires no selective discovery or retry pipeline.
Enum objects are prepared lazily; an entryless census conservatively treats const
enum objects as built. Namespace and enum reads retain their ordinary runtime
readiness checks. No whole-program readiness coverage is claimed.

The loader preserves original input roots before an entry expands its observation
ledger. Prepasses use ESMModuleOrder, the ordering implementation moduleOrder
uses; directory drivers use the stable deduplicated union of the original roots'
orders. The list helper does not move dynamic-import refusals or cyclic read-proof
effects earlier in units. Those remain at their existing query sites.

On the two measured inputs, module membership equals Files exactly: 79 files for
checker.ts and 81 for tsc. Files keeps the entry first; ESM order places its
dependencies first. The three walking fact collectors produce sets, so that order
difference does not change their answers here. Membership can change inheritance,
accessor and omission facts on other inputs with type-only or unrelated roots.
Enum validation is order-independent; namespace readiness is order-sensitive.
collectViewedElements is a no-op with the census's nil key.

Omission discovery reads contextual checker types before explicit unit slots exist
and writes only omittedOptionals. A registered snapshot captures that program-only
map; latentUseProgramFacts installs an owned clone in every unit. Statecopy's table,
type guard, capture/install and owned-map guards include its exact symbol-keyed
shape. Statement snapshots copy it too. The historical predicates driver uses a
minimal helper with the same owned-map policy. The completed entry stream records
one omission discovery call; the other three retained calls run per fresh unit.

## Per-unit costs

Eight fresh preparations were timed for each requested entry, in production pass
order. The two exempt readiness passes were timed on disposable lowerings solely
for this probe. Neither returned an error. Both checker programs have zero
diagnostics. These sampled wall times include timer/logging overhead; the nil-key
collector mostly measures that overhead. First samples, means and ranges are in
costs.json.

| Prepass | checker.ts median ms | tsc entry median ms |
| --- | ---: | ---: |
| collectViewedElements | 0.007 | 0.006 |
| noteInheritance | 23.461 | 20.337 |
| enumInitialization | 5.199 | 4.379 |
| noteAccessorNames | 21.233 | 19.092 |
| noteOmittedOptionals | 26.110 | 25.615 |
| namespaceInitialization | 78.784 | 125.586 |

At the tsc median, repeating omission discovery adds 25.615 ms to 39.429 ms of
inheritance/accessor walks, about 1.65 times that walking subtotal. Over 4,059
fresh units, that is approximately 104.0 seconds of repeated discovery,
before map-copy costs. The implementation computes it once. This is an estimate
from sampled warm costs, not a measured whole-census speedup.

## Guard and validation

The Go AST guard follows Lower's local delegates to the one moduleOrder-to-refuse
window, extracts every lowering method call there, and checks each template in
order against explicit exemptions. It has no Go/compiler subprocess and runs in
parallel. The omission mutant rejects noteOmittedOptionals in the named template;
the fake production pass rejects all four templates. Both mutant files retain their
in-process red assertions and diagnostic logs.

The entire latent package and predicates Go package pass, including
TestLatentSpeculativePrivateIRFields and the state-copy guards. Predicates family
and continuation/body-boundary audits, the latent no-output/continuation audit,
and all 15 pairing tests pass. Both entry overlays and the standard overlay build
successfully outside gate tests. The zero-diagnostic omission probe is missed by
main and observed twice by changed first-error, full and speculative modes, including
the second independent unit. probe-result.json retains those assertions.

The earlier absent-optional report and literal/decodeJson prepass proofs at
2c62ea4e motivated the correction; their feature compiler was not merged here.
The early cold statecopy publication timeout is retained along with the later
passing whole-package run. The inefficient full-history source clone was stopped
with its own exit 143 and replaced by a successful exact pinned shallow fetch.

## Resources, recovery and limits

Every census child exited 0. The 30-minute silence stop and 14 GiB memory ceiling
did not trigger. The core pair ran alongside the baseline; the changed entry run
started near its end under a combined census VmRSS ceiling. Wall times are
shared-host resource observations. resources.csv retains child wall/CPU/peak data.

A transport reconnect ended the baseline's original outer runner while its census
continued unchanged. adopt.py recovered its actual Linux kernel exit status from
/proc after zombie state. Its lifetime CPU ticks and sampled lifetime VmHWM are
identified explicitly; the final polling interval can miss a peak. Later runners
use detached supervisors and wait/reap their own children. Runner checks preserve
exit 7 and verify signal and memory stops report the child's -15 and reap it.

Refusal-scan finding sets are identical. 6 generic witness flags changed;
unit-witness-changes.json retains them without normalization. Raw streams preserve
all event contexts. Raw/distinct counts are unique identities, not event multiplicity.
The pair tool performs stock TypeScript outermost-access-chain normalization.

None of the 73 production first stops would move from core.ts:2599:15: production
already runs omission discovery before its refusal scan, and production code is
unchanged. This follows from unchanged code and pass order; the 73-root production
gate was not rerun. Census unit-stop changes do not change those production results.

See commands.md, provenance.json, source-hashes.json, the two paired JSON/Markdown
ledgers, compressed raw streams and timing journals, and final-packages.log.txt.
