# Readonly enum views, task 52zja36

Status: verification resumed with the corrected paired replay condition.
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

## Check 1: composition and paired replay

Both full applies with --build-product exit 0. The after apply ran every later
adapter in order over adaptation 40's amended debug.ts. Therefore there is no
later pinned-owner conflict in that apply. The adapters after 40 that change
debug.ts, established by the apply's per-adapter before/after Git objects, are:

- 41-explicit-any-remaining
- 76-truthful-casts
- 99-cli-debug-assertion-level
- 99-correctness-axolotl-n
- 99-correctness-axolotl-t

The original whole-tree idempotence condition was corrected by the requester.
Adaptation 30's addRange refusal at core.ts:992:21 occurs on main's own completed
tree too, because later adapters change that owner. It is not an amendment
blocker. The new check replays every adapter on isolated copies of both complete
trees, continues after refusals, and compares normalized outcomes, allowing only
adaptation 40's expected difference. Both records and their diff will be retained.

## Resumed checks

The first bounded latent piece, full mode restricted to before debug.ts, has
completed with exit 0. Its observations and timings are under
`evidence/readonly-views/pieces/latent-before/`. Further results will replace the
previously unmeasured site ledger as they complete. The original partial ordinary
before census and the original stop evidence remain retained as historical data.

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
