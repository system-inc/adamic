The existing setup, main binary and adapted tree from evidence commit `9c5c120c371ab16049225693108fd5446d0ebc84` were reused on the same box. All 79 compiler source hashes and the main binary SHA-256 were verified against that commit before running.

```bash
cd /workspace/adamic
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
git fetch origin optimizer/branded-supertype-road2 > /workspace/scratch/fetch-final.log 2>&1
git rev-parse 65146f1e^{commit}
# 65146f1eff049a8ce9a35fe67c842f79bbe1407c
git worktree add --detach /workspace/scratch/compiler-65146f1e 65146f1eff049a8ce9a35fe67c842f79bbe1407c > /workspace/scratch/paired-65146f1e/worktree.log 2>&1
rmdir /workspace/scratch/compiler-65146f1e/cohere
ln -s /workspace/adamic/cohere /workspace/scratch/compiler-65146f1e/cohere
cd /workspace/scratch/compiler-65146f1e
python3 stage3/census/latent/make_overlay.py /workspace/scratch/compiler-65146f1e /workspace/scratch/overlay-65146f1e > /workspace/scratch/paired-65146f1e/overlay.log 2>&1
gofmt -w /workspace/scratch/overlay-65146f1e/*.go
go build -buildvcs=false -overlay=/workspace/scratch/overlay-65146f1e/overlay.json -o /workspace/scratch/census-65146f1e ./stage3/census/latent/tool > /workspace/scratch/paired-65146f1e/build.log 2>&1
sha256sum /workspace/scratch/census-A /workspace/scratch/census-65146f1e
cd /workspace/adamic
NODE_PATH=/workspace/adamic/stage3/api/node_modules node stage3/census/changeb-road2-meter/65146f1e/some-generic-sites.cjs > /workspace/scratch/paired-65146f1e/some-inventory.log 2>&1
bash stage3/census/changeb-road2-meter/65146f1e/run-censuses.sh > /workspace/scratch/paired-65146f1e/runner.log 2>&1
```

The saved runner launches these commands concurrently, with `wait4` peak RSS/CPU and monotonic wall time, SIGTERM after 5400 seconds, SIGKILL after another 60 seconds. GNU time is absent, as established in the previous pair.

```bash
LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES=/workspace/scratch/paired-65146f1e/A-times.jsonl /workspace/scratch/census-A /workspace/scratch/tree/src/compiler /workspace/scratch/paired-65146f1e/A.jsonl
LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES=/workspace/scratch/paired-65146f1e/B-times.jsonl /workspace/scratch/census-65146f1e /workspace/scratch/tree/src/compiler /workspace/scratch/paired-65146f1e/B.jsonl
python3 stage3/adapt/40-explicit-any/census-delta.py /workspace/scratch/paired-65146f1e/A.jsonl /workspace/scratch/paired-65146f1e/B.jsonl stage3/census/changeb-road2-meter/65146f1e > /workspace/scratch/paired-65146f1e/delta.log 2>&1
PYTHONDONTWRITEBYTECODE=1 python3 stage3/census/changeb-road2-meter/65146f1e/analyze.py > /workspace/scratch/paired-65146f1e/analysis.log 2>&1
```

A is main `141ad1d244d6a2d9c5403858086d505dfdef5fd5`; B is final road 2 `65146f1eff049a8ce9a35fe67c842f79bbe1407c`. Their meter/recorder code is identical; no backport or compiler edits were applied. The new compiler worktree uses the same initialized `cohere` dependency symlink as before. The stock API inventory completed before the paired census started.

Evidence-only commit and foreground push:

```bash
git add stage3/census/changeb-road2-meter/65146f1e
git add -f stage3/census/changeb-road2-meter/65146f1e/logs
git diff --cached --check
git commit -m "Measure final road-2 head 65146f1e against main"
git push origin typescript/changeb-road2-meter
```

After both censuses finished, the stock API inventory was rerun to preserve its diagnostic details (same 305 calls and one generic-argument site):

```bash
NODE_PATH=/workspace/adamic/stage3/api/node_modules node stage3/census/changeb-road2-meter/65146f1e/some-generic-sites.cjs > /workspace/scratch/paired-65146f1e/some-inventory-diagnostics.log 2>&1
```

Packaging (includes gzip round-trip checks, unchanged input hashes, same box and byte-identical main ledger checks):

```bash
PYTHONDONTWRITEBYTECODE=1 python3 stage3/census/changeb-road2-meter/65146f1e/package.py > /workspace/scratch/package-65146f1e.log 2>&1
```
