#!/usr/bin/env bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
unset GOCACHEPROG
export GOPROXY='https://proxy.golang.org|direct'
export GOMAXPROCS=4
export GOFLAGS="${GOFLAGS:-} -p=4"
out=review/compiler/loom-slice-5-3
clang --version > "$out/toolchain.log"
go version >> "$out/toolchain.log"
node --version >> "$out/toolchain.log"
nproc >> "$out/toolchain.log"
timeout 180 go test ./internal/ir ./internal/lower ./internal/oracle ./cmd/adamic-gate -run '^(TestEveryMapAProgramHoldsIsSorted|TestRuntimeSetsAreEachUsed|TestNativeAgreementRowsNameRegisteredPrograms|TestEveryTestIsParallelOrSaysWhy)$' -count=1 -json -timeout 90s > "$out/guards.jsonl" 2>&1
timeout 240 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -json -timeout 180s -args -update-counts > "$out/counts.jsonl" 2>&1
timeout 600 python3 "$out/run-focused.py" > "$out/focused.log" 2>&1
timeout 240 python3 "$out/run-fixtures.py" > "$out/fixtures.log" 2>&1
timeout 180 go test ./internal/oracle -run '^TestSlice5Witness' -count=1 -json -timeout 90s > "$out/ledger.jsonl" 2>&1
for runner in product-mutants step24-parser-main-mutants placeholder-nonnull-main-mutants backend-stop-mutants; do
 timeout 600 python3 "$out/$runner.py" > "$out/$runner.log" 2>&1
done
timeout 180 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -json -timeout 90s > "$out/readers.jsonl" 2>&1
timeout 180 go vet ./cmd/adamic ./internal/flow ./internal/fresh ./internal/ir ./internal/javascript ./internal/lower ./internal/native ./internal/oracle > "$out/vet.log" 2>&1
