#!/usr/bin/env bash
set -euo pipefail
cd /workspace/adamic
exec python3 stage3/census/changeb-final-meter/run-censuses.py
