#!/usr/bin/env bash
set -euo pipefail
landing_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_dir="$(cd "$landing_dir/../../../../.." && pwd)"
log_dir="${1:-$(mktemp -d -t wave28-checker-landing.XXXXXX)}"
mkdir -p "$log_dir"
log_dir="$(cd "$log_dir" && pwd)"
python3 - "$repo_dir" "$landing_dir" "$log_dir/overlay.json" <<'PYOVERLAY'
import json,sys
root,owned,path=sys.argv[1:]
with open(path,'w') as output:
 json.dump({'Replace':{root+'/stage1/cohere/lint/wave28_owned_landing_test.go':owned+'/owned_cases_test.go'}},output,indent=2)
PYOVERLAY
cd "$repo_dir"
export GOMAXPROCS=4 GOFLAGS=-buildvcs=false
go test -json -count=1 -timeout=60m -overlay="$log_dir/overlay.json" -tags=wave28landing ./stage1/cohere/lint -run '^TestWave28EveryCapturedCase$' > "$log_dir/owned-cases.jsonl" 2>&1
go test -json -count=1 -timeout=60m -parallel=1 ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestMutants)$/(Only_root_nullability_accepted|All_var_uses_treated_as_in_scope|Either_if_branch_credited_as_exit|Conditional_error_possibility_changed_from_or_to_and|Unicode_component_mistaken_for_intrinsic|React_fragment_receiver_renamed|Arrow_insertion_gains_a_trailing_space|Timeout_assignment_recognition_reversed|Capturing_group_duplicated)$' > "$log_dir/mutants-witnesses.jsonl" 2>&1
