set -euo pipefail
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
cd /workspace/scratch/head
mkdir -p /workspace/scratch/accessor-results
python3 /workspace/scratch/trace-accessors.py > /workspace/scratch/accessor-results/overlay.log 2>&1
gofmt -w /workspace/scratch/accessor-trace.go /workspace/scratch/accessor-hook.go /workspace/scratch/accessor-enums.go
go build -buildvcs=false -overlay=/workspace/scratch/accessor-overlay.json -o /workspace/scratch/accessor-results/latent-census ./stage3/census/latent/tool > /workspace/scratch/accessor-results/build.log 2>&1
python3 /workspace/scratch/measure.py /workspace/scratch/accessor-results/rss.json env CENSUS_ACCESSOR_SCAN_ONLY=1 CENSUS_ACCESSOR_FILES=/workspace/scratch/accessor-files.json CENSUS_ACCESSOR_TRACE=/workspace/scratch/accessor-results/trace.jsonl LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 /workspace/scratch/accessor-results/latent-census /workspace/scratch/adapted/src/compiler /workspace/scratch/accessor-results/latent.jsonl > /workspace/scratch/accessor-results/run.log 2>&1
python3 /workspace/scratch/summarize-accessors.py > /workspace/scratch/accessor-results/identity.log 2>&1
