set -euo pipefail
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
for side in base head; do
 cd /workspace/scratch/$side
 mkdir -p /workspace/scratch/$side-results
 if [[ ! -x /workspace/scratch/$side-results/latent-census ]]; then
  python3 stage3/census/latent/make_overlay.py "$PWD" /workspace/scratch/$side-overlay > /workspace/scratch/$side-results/overlay.log 2>&1
  gofmt -w /workspace/scratch/$side-overlay/*.go
  go build -buildvcs=false -overlay=/workspace/scratch/$side-overlay/overlay.json -o /workspace/scratch/$side-results/latent-census ./stage3/census/latent/tool > /workspace/scratch/$side-results/build.log 2>&1
 fi
 python3 /workspace/scratch/measure.py /workspace/scratch/$side-results/rss.json env LATENT_TIMES=/workspace/scratch/$side-results/times.jsonl LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 /workspace/scratch/$side-results/latent-census /workspace/scratch/adapted/src/compiler /workspace/scratch/$side-results/latent.jsonl > /workspace/scratch/$side-results/run.log 2>&1
done
