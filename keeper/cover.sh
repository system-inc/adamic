#!/bin/bash
# cover.sh: the test audit's line map for one main sha (#mp71kkr, @system_adamic_tests): the whole Go suite, planned as
# the star's gate plans it, with every part also writing which lines of Adamic's own Go packages its tests executed
# (ADAMIC_GATE_COVERPKG, read by the unit body). replay.sh then sends a mutant only to the parts that ran its line,
# since a compiler mutant reaches about 44 packages by imports and most of them never execute the changed line.
#
#	pilots/adamic-gate/cover.sh <sha> [<units>]    # plans the job and prints its path; Loom runs it on the star's pool
#
# The map covers Go only: the C runtime and the stage1 ports' .a and .ts sources aren't Go coverage's to see, and a
# mutant there replays on its own package by name (REPLAY_PACKAGES). Coverage slows tests, so the units get the gate's
# budget doubled; a part that times out under coverage leaves its lines unmapped, and the map says which.
set -uo pipefail

sha=$1 units=${2:-60}
[[ ${sha} =~ ^[0-9a-f]{40}$ ]] || { echo "cover: a full sha, please"; exit 2; }
bin=${HOME}/.loom/bin gate=${HOME}/Projects/system/adamic-gate work=${HOME}/.loom/cover/${sha}
mkdir -p "${work}"
git -C "${gate}" fetch -q origin "${sha}" || { echo "cover: can't fetch ${sha}"; exit 2; }
reference=$(ls -t "${HOME}"/.loom/pregate/reference-*.jsonl.gz | head -1)
# Adamic's own Go packages that tests reach: internal/, bridge/, cmd/. stage1 and stage3 hold ports and drivers whose
# Go is harness, and mutating the harness tests the oracle.
coverpkg=github.com/system-inc/adamic/internal/...,github.com/system-inc/adamic/bridge/...,github.com/system-inc/adamic/cmd/...
"${bin}/adamic-gate" plan --target codex --remainder --gate-inputs "$(cat "${HOME}/.loom/gate-inputs")" --reference "${reference}" --sha "${sha}" --units "${units}" > "${work}/plain.json" 2> "${work}/plan.err" || { echo "cover: planning failed: $(tail -1 "${work}/plan.err")"; exit 2; }
python3 - "${work}/plain.json" "${work}/job.json" "${coverpkg}" <<'PY'
import json, sys
plain, path, coverpkg = sys.argv[1:]
job = json.load(open(plain))
job["name"] = "adamic-cover"
for unit in job["units"]:
    # The body is argv[2]: the export goes first, so the unit body sees it whatever it opens with.
    unit["argv"][2] = "export ADAMIC_GATE_COVERPKG=%s\n" % coverpkg + unit["argv"][2]
    unit["timeoutSeconds"] = unit.get("timeoutSeconds", 0) * 2
    if not any(output.get("glob") == "loom-out/cover.tgz" for output in unit.get("outputs", [])):
        unit.setdefault("outputs", []).append({"glob": "loom-out/cover.tgz"})
json.dump(job, open(path, "w"))
print("%s: %d units" % (path, len(job["units"])))
PY
