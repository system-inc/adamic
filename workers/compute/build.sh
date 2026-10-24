#!/usr/bin/env bash
set -euo pipefail
if [[ $# != 1 ]]; then
  echo 'usage: workers/compute/build.sh <out-directory>' >&2
  exit 2
fi
"${ADAMIC:-adamic}" worker workers/compute/handler.a --out "$1"
