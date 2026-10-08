# Remaining source adaptations

Adaptation 41 removes 39 of135 explicit any tokens and types five untyped locals.
The current main census has 37 direct any observations; the final tree has 20.
One final observation is an already concrete NodeJS.Timeout | undefined parameter,
resolved without any by stock TypeScript; the latent discrepancy is not diagnosed.
No compiler, shared lane, shared census or oracle harness is edited.

The continuation from accepted c6ea4131 is split into timer globals, complete
constructor stages, shallow copies, JSON/config, untyped locals and generic data
comparison. Each has exact owner rules, stock checking, whole-file emitted
JavaScript equality, full census observations and actual restored-any mutants.
Constructor, copy, JSON and comparison bodies also run on Node.
The final fresh apply/lane runs the full upstream oracle with eight workers and
checks the existing 222 public API sanctions. See REPORT.md for its result.

Main is 5fda2d26 after merging it into the assigned branch; its adaptation/compiler
inputs are unchanged from the earlier ef3141e9 measurement. Its changed files are
test scheduling, lane/oracle worker policy and docs, retained in final provenance.
The additional TABLE.md is at dc6b1529. Any-array storage/length is already
supported and has zero census rows; that kind is dropped. All15 source void
expressions are adapted. Debugger remains because deleting it changes inspector
behavior. With has zero table/source sites and is checker-rejected. Variance and
the other ranked source fixes remain for their owners.

The remaining 96 syntax tokens are independently parsed, with exact current
locations in evidence/finished-sites.json. RESIDUE.md explains the20 observed
blockers and related hidden timer, JSON, staged allocator and expando-copy
contracts. Recursive JSON has a static union; this unit does not claim its
remaining public/boundary work impossible. No unknown cast-back, alias for any
or phantom generic inferred from an any-returning host is introduced.

New internal JSON declarations must be exported with @internal at their owners.
Private names caused a real declaration-bundling failure; its failed lane and
repair proof are retained. A final idempotence/LF reconstruction guard proof
also exercises real duplicate/drift/shifted-void mutants.

Commands write to logs:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > setup.log 2>&1
source /workspace/adamic-tools/env.sh
export NODE_PATH="$HOME/.cache/adamic-stage3/api/node_modules"
bash stage3/apply.sh NEW_TREE > apply.log 2>&1
node "$NODE_PATH/typescript/lib/tsc.js" -p NEW_TREE/src/compiler --noEmit > stock.log 2>&1
node stage3/adapt/41-explicit-any-remaining/family-proof.cjs BEFORE AFTER FAMILY proof.json > proof.log 2>&1
node stage3/adapt/41-explicit-any-remaining/guard-proof.cjs MAIN FINAL guards.json > guards.log 2>&1
python3 stage3/census/latent/make_overlay.py "$PWD" NEW_OVERLAY > overlay.log 2>&1
gofmt -w NEW_OVERLAY/*.go
go build -buildvcs=false -overlay=NEW_OVERLAY/overlay.json -o NEW_BINARY ./stage3/census/latent/tool > build.log 2>&1
LATENT_ASSERT_NO_OUTPUT=1 NEW_BINARY NEW_TREE/src/compiler census.jsonl > census.log 2>&1
python3 stage3/adapt/41-explicit-any-remaining/census-report.py census.jsonl NEW_TREE report.json > recount.log 2>&1
bash stage3/lane/run.sh NEW_LANE > lane.log 2>&1
```

Historical family snapshots must use their recorded rule commit; current rules
include the JSON internal-owner repair. family-proof checks only its reviewed
family files. Final guards independently reconstruct every owned source file
from the main input, and final source hashes match the lane's fresh apply tree.
There are no new fixtures or counts.md changes. No full Go-package gate was run.
The user's final rule supersedes per-family pushes: remaining commits stay local
until the finished unit passes; then one push.

See REPORT.md, TIMERS.md, CONSTRUCTORS.md, COPIES.md, JSON-CONFIG.md, LOCALS.md,
DATA.md, ANY-RETURNS.md, ANY-CALLS.md, DEBUGGER.md, DROPPED.md, REMAINING.md,
RESIDUE.md and RANKED.md. The prior c6ea4131 report remains in git history with
its original evidence; current evidence adds family and final observations.
