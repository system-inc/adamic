#!/bin/bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
cd /tmp/meter-stall-20261105/source
while ! test -f cohere/TypeScript/tsc/go.mod; do sleep 5; done
python3 stage3/census/latent/make_overlay.py . /tmp/meter-stall-20261105/latent-overlay > /tmp/meter-stall-20261105/overlay.log 2>&1
gofmt -w /tmp/meter-stall-20261105/latent-overlay/*.go
python3 stage3/meter/entry_overlay.py /tmp/meter-stall-20261105/latent-overlay /tmp/meter-stall-20261105/entry-overlay > /tmp/meter-stall-20261105/entry-overlay.log 2>&1
gofmt -w /tmp/meter-stall-20261105/entry-overlay/*.go
go build -buildvcs=false -overlay=/tmp/meter-stall-20261105/entry-overlay/overlay.json -o /tmp/meter-stall-20261105/entry-census ./stage3/census/latent/tool > /tmp/meter-stall-20261105/build.log 2>&1
