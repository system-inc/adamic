#!/usr/bin/env bash
# Record of material commands run from /workspace/adamic, unless a subshell says otherwise.
# This file documents the measurement; it is not a one-command rerun script.
mkdir -p /workspace/scratch
git fetch origin main > /workspace/scratch/fetch.log 2>&1
git switch -c typescript/axolotl-classify origin/main
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1
# Read these files in full before preparing the measurement:
cat stage3/meter/twice-daily.sh
cat stage3/meter/report.py
cat stage3/census/tool/main.go
cat stage3/census/latent/tool/main.go
cat stage3/census/latent/make_overlay.py
cat stage3/census/latent/main.go.txt
cat stage3/census/latent/load.go.txt
cat internal/load/load.go
cat /etc/codex/network-policy.json
git rev-parse HEAD
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /workspace/scratch/upstream > /workspace/scratch/upstream.log 2>&1
(cd stage3/api && npm ci --ignore-scripts --no-audit --no-fund > /workspace/scratch/packages.log 2>&1)
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash stage3/apply.sh /workspace/scratch/adapted > /workspace/scratch/apply.log 2>&1
# The first build attempt used /usr/bin/go and failed with 'Go: Unknown option: build'.
# Setup's generated environment selects the actual Go toolchain; the successful rerun sources it.
source /workspace/adamic-tools/env.sh
# extract.go and extract-overlay.json were created only in /workspace/scratch.
# The overlay substitutes the scratch Go main for stage3/census/tool/main.go at build time.
# No production or census-tool source is edited.
go build -o /workspace/scratch/census ./stage3/census/tool > /workspace/scratch/build-census.log 2>&1
go build -buildvcs=false -overlay=/workspace/scratch/extract-overlay.json -o /workspace/scratch/extract ./stage3/census/tool > /workspace/scratch/build-extract.log 2>&1
/workspace/scratch/census /workspace/scratch/adapted/src/compiler /workspace/scratch/census.jsonl > /workspace/scratch/census.log 2>&1
/workspace/scratch/extract /workspace/scratch/adapted/src/compiler /workspace/scratch/extracted.jsonl > /workspace/scratch/extract.log 2>&1
# Build and classification scripts also stayed under scratch.
PYTHONDONTWRITEBYTECODE=1 python3 /workspace/scratch/build_report.py > /workspace/scratch/build-report.log 2>&1
# The report builder invokes stage3/meter/report.py:report on census.jsonl,
# compares the scratch extraction record by record, then writes these artifacts.
git submodule status --recursive
# Final validation and publication:
PYTHONDONTWRITEBYTECODE=1 python3 /workspace/scratch/validate_report.py > /workspace/scratch/validation.log 2>&1
git diff --check
git add -- stage3/census/axolotl
git commit -m 'stage3/census: classify axolotl checker diagnostics'
git push origin HEAD:refs/heads/typescript/axolotl-classify
