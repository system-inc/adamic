#!/usr/bin/env bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
mkdir -p /tmp/decode-ascii/service /tmp/decode-ascii/runtime-before
# Before differs in exactly input.c; objects.json holds the independent all-object check.
cp internal/native/runtime/*.c internal/native/runtime/*.h /tmp/decode-ascii/runtime-before/
git show 4d86c305dda261768b35687d199c1b2188c7ab71:internal/native/runtime/input.c > /tmp/decode-ascii/runtime-before/input.c
git show f4ec96c:internal/native/wasm/service/service.a > /tmp/decode-ascii/service/service.a
git show f4ec96c:internal/native/wasm/service/host.mjs > /tmp/decode-ascii/service/host.mjs
git show a4e0902:cloud/reports/wasm-requests-profile/command.a > /tmp/decode-ascii/service/command.a
python3 - <<'PY'
from pathlib import Path
p=Path('/tmp/decode-ascii/service/command.a')
p.write_text(p.read_text().replace('../../../internal/native/wasm/service/service.a','./service.a'))
PY
go build -o /tmp/decode-ascii/build-service cloud/reports/decode-ascii/build.go
/tmp/decode-ascii/build-service /tmp/decode-ascii/runtime-before /tmp/decode-ascii/service/service.a wasi /tmp/decode-ascii/service-before.wasm
/tmp/decode-ascii/build-service internal/native/runtime /tmp/decode-ascii/service/service.a wasi /tmp/decode-ascii/service-after.wasm
/tmp/decode-ascii/build-service /tmp/decode-ascii/runtime-before /tmp/decode-ascii/service/command.a native /tmp/decode-ascii/native-before
/tmp/decode-ascii/build-service internal/native/runtime /tmp/decode-ascii/service/command.a native /tmp/decode-ascii/native-after
