#!/usr/bin/env bash
# Usage: bash reproduce.sh ADAPTED_COMPILER_DIRECTORY NEW_OUTPUT_DIRECTORY
set -euo pipefail
root=$(realpath "$1")
out=$(realpath -m "$2")
repo=$(git rev-parse --show-toplevel)
mkdir -p "$out"
cd "$repo"
export GOMEMLIMIT=${GOMEMLIMIT:-3GiB}
python3 stage3/census/latent/make_overlay.py "$repo" "$out/overlay" > "$out/overlay.log" 2>&1
go build -buildvcs=false -overlay="$out/overlay/overlay.json" -o "$out/census" ./stage3/census/latent/tool > "$out/build.log" 2>&1
python3 stage3/census/speculative/audit_binding.py "$repo" "$out/overlay" "$out/census" "$out/binding" > "$out/binding.log" 2>&1
python3 stage3/census/speculative/audit_snapshot.py "$repo" "$out/overlay" "$out/snapshot" > "$out/snapshot.log" 2>&1
python3 stage3/census/speculative/audit.py "$out/census" "$out/control" > "$out/control.log" 2>&1
python3 stage3/census/speculative/audit_signature.py "$out/census" "$out/signature-control" > "$out/signature-control.log" 2>&1
python3 stage3/census/speculative/audit_dependency.py "$out/census" "$out/dependency-control" > "$out/dependency-control.log" 2>&1
python3 stage3/census/speculative/prove_identity.py "$repo" "$out/overlay" "$out/census" "$out/identity" > "$out/identity.log" 2>&1
node stage3/census/speculative/stock.cjs "$out/control/source" "$out/control/speculative.jsonl" "$out/control/stock.json" > "$out/control/stock.log" 2>&1
python3 stage3/census/speculative/audit_report_controls.py "$out/control/source" "$out/control/speculative.jsonl" "$out/control/stock.json" "$out/control/full.jsonl" "$out/control/no-stubs-mutant.jsonl" "$out/report-guards" > "$out/report-guards.log" 2>&1
python3 stage3/census/speculative/isolation.py "$repo" "$out/isolation" > "$out/isolation.log" 2>&1
python3 stage3/census/latent/audit_output_guards.py "$repo" "$out/overlay" "$out/guards" > "$out/guards.log" 2>&1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 "$out/census" "$root" "$out/full.jsonl" > "$out/full.log" 2>&1
LATENT_SPECULATIVE=1 LATENT_MUTANT_NO_STUBS=1 LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 "$out/census" "$root" "$out/no-stubs.jsonl" > "$out/no-stubs.log" 2>&1
LATENT_SPECULATIVE=1 LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 "$out/census" "$root" "$out/speculative.jsonl" > "$out/speculative.log" 2>&1
node stage3/census/speculative/stock.cjs "$root" "$out/speculative.jsonl" "$out/stock.json" > "$out/stock.log" 2>&1
python3 stage3/census/speculative/report.py "$root" "$out/speculative.jsonl" "$out/stock.json" "$out/results" > "$out/report.log" 2>&1
python3 stage3/census/speculative/verify.py "$out/speculative.jsonl" "$out/results/RESULT.json" "$root" > "$out/verify.log" 2>&1
python3 stage3/census/speculative/compare.py "$out/full.jsonl" "$out/no-stubs.jsonl" "$out/speculative.jsonl" > "$out/compare.log" 2>&1
