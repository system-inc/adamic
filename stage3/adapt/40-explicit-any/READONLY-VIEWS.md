# Readonly enum views, task 52zja36

Status: stopped at check 1, with the amendment retained for review.
Base: origin/main 63e14a0b2b61cf251ad74292e42d5fd09af804b3.

Exactly 21 replacement strings change in class-rules.json:

- Rules 72 through 89 and 91 use
  `Readonly<Record<"K", Readonly<Record<string, string | number>>>>`.
  Rule 91 is the GeneratedIdentifierFlags view in emitter.ts; the other
  18 casts are in debug.ts, including two separate FlowFlags reads.
- Rules 69 and 71 use `Readonly<Record<string, string | number>>` for
  formatEnum and getEnumMembers.

The outer Readonly describes a namespace field that cannot be assigned, and
its inner Readonly describes a read view of the enum table. No cast is added.
Every other rule and every non-replacement field is unchanged. Rule 70's Map
key is unchanged; no further body edit has been established as necessary.
class-proof.cjs's declaration projection and Boolean-map mutant now use the
readonly parameter spelling. Earlier enum-display evidence remains historical.

## Stop

Both full applies with --build-product exit 0. Ordered replay over the completed
after tree records zero changed files for 00-setup, 10-type-imports and
20-optional-declarations. It then stops, without writes, at 30-indexed-reads:

```
declined site changed: core.ts:992:21
```

The pinned owner is addRange, expression from[i], occurrence 2 of 2, action
decline. The completed tree contains `to.push(from[i]!)`. The same adapter
command refuses identically on the baseline full apply. This is a pre-existing
composition conflict, not evidence of an enum-view conflict. No other adaptation
directory is edited. Later adapters were not replayed, so no conclusion about
their debug.ts owner pins is claimed.

The remaining verification stops here. The ordinary before census was already
running and was terminated with exit 143; its retained JSONL is partial. There
is no complete before/after census, no latent run, and no measured per-site
before/after stop. All 19 sites are explicitly marked unmeasured in
[evidence/readonly-views/site-stops.json](evidence/readonly-views/site-stops.json).
The types check, full nine-file artifact comparison, stage3 consumer tests and
inner-only Readonly mutant have not run. The branch does not meet the full
six-check acceptance condition.

## Completed evidence

[evidence/readonly-views/status.json](evidence/readonly-views/status.json) records
all six checks individually. The directory retains compressed setup, source,
build, apply and partial census logs, plus per-adapter replay logs.

The focused enum planner proof passes against the real adaptation-40-before
input. It checks all 23 existing enum-display rules, idempotence, a foreign type
mutant, and byte-identical complete-file stock transpilation with comments kept
and removed. That covers debug.ts, emitter.ts and program.ts. It is not the
required comparison of all nine normally emitted files. The baseline ordinary
compiler build passes; no after build is claimed.

Commands run after sourcing /workspace/adamic-tools/env.sh, with each check's
stdout and stderr redirected to its named log:

```sh
bash cloud/setup.sh > /workspace/scratch/ro-setup.log 2>&1
npm ci --prefix stage3/api > /workspace/scratch/ro-npm.log 2>&1
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source
bash stage3/apply.sh /workspace/scratch/ro-before --build-product > /workspace/scratch/ro-before-apply.log 2>&1
# Change only the 21 rule replacement strings and matching proof strings.
bash stage3/apply.sh /workspace/scratch/ro-after --build-product > /workspace/scratch/ro-after-apply.log 2>&1
python3 stage3/adapt/40-explicit-any/evidence/readonly-views/replay.py /workspace/scratch/ro-after /workspace/scratch/ro-replay > /workspace/scratch/ro-replay.log 2>&1
NODE_PATH=stage3/api/node_modules node stage3/adapt/30-indexed-reads/adapt.cjs /workspace/scratch/ro-before > /workspace/scratch/ro-baseline-replay30.log 2>&1
NODE_PATH=stage3/api/node_modules node stage3/adapt/40-explicit-any/evidence/readonly-views/focused-proof.cjs /workspace/scratch/ro-before > /workspace/scratch/ro-focused-proof.log 2>&1
```

The original focused run used the same script in scratch; the retained script
resolves the repository relative to itself and pins the baseline rules to the
base SHA. rules-scope.json records every replacement before and after.
All started long-running processes were completed or terminated before delivery.

A retained-script rerun initially omitted NODE_PATH and failed to resolve the
parser from classes.cjs. That invocation is retained in
focused-proof-missing-node-path.log; the corrected rerun passes.
