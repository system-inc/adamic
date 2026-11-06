# How these outputs were made

On Golem, Nov 5, Node 24.21.0 from `~/.adamic-tools/env.sh`, in a worktree at
main `521ca101` with stage3/api/node_modules at its pinned digest
(`03b25a14...`). This directory lives on cli/host-require-adapt-evidence only,
because every file under the adaptation's directory feeds apply's input key.

```sh
export NODE_PATH=$PWD/stage3/api/node_modules
unit=stage3/adapt/99-host-require
# Control: main without this directory (481 s, exit 0).
stage3/apply.sh ~/cli-ha/tree-main > apply-main.log 2>&1
# The whole chain with this directory (487 s, exit 0); patch-set.md is its table.
stage3/apply.sh ~/cli-ha/tree-after > apply-after.log 2>&1
node $unit/adapt.cjs ~/cli-ha/tree-after > idempotence.log 2>&1
node $unit/witnesses.cjs ~/cli-ha/tree-main ~/cli-ha/tree-after witnesses.json > witnesses.log 2>&1
node $unit/witnesses.cjs --mutant ~/cli-ha/tree-main ~/cli-ha/tree-after mutant.json > mutant.log 2>&1 # exits 1
node refusals.cjs $PWD/$unit/adapt.cjs ~/cli-ha/tree-main ~/cli-ha/tree-after refusals refusals.json > refusals.log 2>&1
node require-census.cjs ~/cli-ha/tree-main require-census-before.json
node require-census.cjs ~/cli-ha/tree-trial require-census-after.json
# Upstream's install and build, in a copy of tree-after, with ~/fast-gate/npm/bin on PATH.
npm ci --no-audit --no-fund && npm run build > build.log 2>&1
```

tree-trial is tree-main with this adapter run on it alone. Its src is
byte-identical to tree-after's (`diff -r -q`), so the after census holds for
both.
