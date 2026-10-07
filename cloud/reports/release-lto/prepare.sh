#!/usr/bin/env bash
# Run at the repository root on the report branch. All new Adamic sources are .a.
set -euo pipefail
scratch=${1:?pass an absolute scratch directory}
repository=$(git rev-parse --show-toplevel)
mkdir -p "$scratch/harness/testdata" "$scratch/service" /tmp/wasm-requests-profile
git fetch origin 4189abd3490757e8abe13722ceb365c451293e92 2cdf8dd3faff065390720cfc8bc9b09a1c9e62d2 a4e0902afc35cdc79c09fa7e58fa8d56b2189f55
if [ ! -d "$scratch/typescript/.git" ]; then
    git init "$scratch/typescript"
    git -C "$scratch/typescript" remote add origin https://github.com/microsoft/TypeScript.git
    git -C "$scratch/typescript" fetch --depth=1 origin 050880ce59e30b356b686bd3144efe24f875ebc8
    git -C "$scratch/typescript" checkout 050880ce59e30b356b686bd3144efe24f875ebc8
fi
git show 2cdf8dd3:internal/native/performance/parse-speed/prepare.py > "$scratch/harness/prepare.py"
git show 2cdf8dd3:internal/native/performance/parse-speed/testdata/go_parse.go > "$scratch/harness/testdata/go_parse.go"
git show a4e0902:cloud/reports/wasm-requests-profile/command.a > "$scratch/command.a"
git show a4e0902:cloud/reports/wasm-requests-profile/measure.mjs > "$scratch/measure.mjs"
git show a4e0902:internal/native/wasm/service/service.a > "$scratch/service/service.a"
git show a4e0902:internal/native/wasm/service/host.mjs > "$scratch/service/host.mjs"
python3 - "$scratch" "$repository" <<'PY'
from pathlib import Path
import sys
scratch, repository = map(Path, sys.argv[1:])
p = scratch / 'harness/prepare.py'
p.write_text(p.read_text().replace('repo = Path(__file__).resolve().parents[4]', 'repo = Path(' + repr(str(repository)) + ')'))
p = scratch / 'command.a'
p.write_text(p.read_text().replace('../../../internal/native/wasm/service/service.a', './service/service.a'))
p = scratch / 'measure.mjs'
p.write_text(p.read_text().replace('internal/native/wasm/service/', str(scratch / 'service') + '/'))
PY
python3 "$scratch/harness/prepare.py" "$scratch" "$scratch/typescript"
go build -o "$scratch/adamic" ./cmd/adamic
"$scratch/adamic" c "$scratch/batch8/parse.a" > "$scratch/parse.c" 2> "$scratch/parse-emit.log"
"$scratch/adamic" c "$scratch/command.a" > "$scratch/service.c" 2> "$scratch/service-emit.log"
node "$scratch/measure.mjs" prepare > "$scratch/service-prepare.log" 2>&1
