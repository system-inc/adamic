# Step 18: optional calls and the rest of `?`

Branch `codex/scout-optional-calls`, based on the explicit area-next landing
candidate `8cb5e7c1`. This report serves roadmap step 18 and task #ht2nwj5.

## Where it bites in TypeScript

Observed: a fresh no-output latent census of the current stage3-adapted
TypeScript compiler directory on this base. TypeScript is pinned to 6.0.3,
upstream `050880ce59e30b356b686bd3144efe24f875ebc8`; `stage3/apply.sh`
produced the adapted source. The census measurement says **checker-rejected**:
it inventories compiler barriers without claiming those programs are accepted.

The supplied hidden ranking at `6c4fc1af` was measured with compiler
`ed6e29751ee47d86fad450cd1674139883bc0f70`, not this base. Its hidden-byte
numbers below are historical exclusive bytes revealed if that reason alone were
fixed, not measured bytes retired on this branch. Keep that provenance separate
from the fresh root counts. Ranking boundary counts are distinct boundary spans,
not the fresh diagnostic-root counts. Full normalized roots and attribution
metadata are in [inventory.json](step-18/inventory.json).

A root here is a distinct `(kind, file:line:column, reason, text)` finding across
the compiler census records. Repeated entry observations count once. Witnesses
are source positions in the pinned adapted compiler; `H` marks a historical
ranking witness when fewer than three fresh positions exist. No witness is
invented to meet a quota.

| Kind and exact reason | Base roots | Historical hidden bytes | Historical boundary spans | Three file:line witnesses |
|---|---:|---:|---:|---|
| NotYet: `a call through ?. (an optional call)` | 136 | 4053 | 136 | `builder.ts:670`; `builder.ts:710`; `builder.ts:1852` |
| NotYet: `an optional chain longer than one step` | 11 | 2106 | 11 | `checker.ts:4745`; `checker.ts:36469`; `checker.ts:53798` |
| NotYet: `?. to a number, which would be number | undefined` | 3 | 1964 | 3 | `checker.ts:10416`; `checker.ts:10445`; `utilities.ts:8102` |
| NotYet: `optional chaining to .size on a value` | 20 | 341 | 20 | `builder.ts:705`; `builder.ts:756`; `builder.ts:778` |
| NotYet: `?.[] on a value` | 6 | 58 | 6 | `checker.ts:9101`; `checker.ts:22931`; `checker.ts:33479` |

The five syntax reasons account for 176 fresh
roots and 8,522 historical exclusive hidden bytes. All five are classified
`compiler lesson` by the supplied NotYet table; its owner column is absent.
The ranking's refusal table is pinned to `d35a81d`, and its NotYet table to
`dc6b1529`. Neither supplies an optional-call acceptance policy change.

There is no separate Refused reason for these optional-syntax shapes in the
ranking or the fresh census: zero such roots, zero attributed hidden bytes,
and no witnesses. Refused relations mentioning an *optional field* are the
optional-widening and mutable-relation soundness rules, not optional syntax.
`writing a possibly absent optional own field`, `reading optional`, and checked
view `_optionalChainBrand` representation are also separate lessons. They stay
outside this retirement count. A generic refusal can block a program that also
contains `?.`; that does not turn its refusal into a step-18 syntax root.

### Reproduction and limits

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/scout-optional-setup.log 2>&1
source /workspace/adamic-tools/env.sh
STAGE3_CACHE=/workspace/scratch/scout-stage3-cache bash stage3/apply.sh /workspace/scratch/scout-optional-adapted > /tmp/scout-optional-apply.log 2>&1
python3 stage3/census/latent/make_overlay.py "$PWD" /workspace/scratch/scout-optional-overlay > /tmp/scout-optional-overlay.log 2>&1
go build -buildvcs=false -overlay=/workspace/scratch/scout-optional-overlay/overlay.json -o /workspace/scratch/scout-optional-census ./stage3/census/latent/tool > /tmp/scout-optional-census-build.log 2>&1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /workspace/scratch/scout-optional-census /workspace/scratch/scout-optional-adapted/src/compiler /workspace/scratch/scout-optional-base.jsonl > /tmp/scout-optional-base-census.log 2>&1
```

The output guard returned measurement records without usable IR. This run
covers the compiler-directory census, not the tsc CLI entry census. A retired
root means this exact local barrier disappeared; later NotYet or Refused
barriers and checker errors can remain. Historical hidden bytes must not be
reported as successful compilation or additive porting progress.
