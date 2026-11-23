source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh && source cloud/gates/steps.sh && gate_begin admission-delta || exit 1
# ask: --size F64
# admission-delta: the admission delta against main's tip, its verdict pass and its admitted programs listed
# (cloud/gates/diff-to-gates.json). Its base is HEAD^, main's tip when the head was cut, and submit-gates.py holds it to
# main's tip at submit by the map's fresh rule. cmd/adamic-admission-delta checks out each revision with its pinned
# submodules from the tree's own git (checkoutPinnedFrom), so the submodules are laid as repositories at their pins
# first, and stage3/api's install is laid for the head checkout to copy (its README's first road).
# This file is the whole command of a Loom sealed job at the head's own sha, from its first line, which opens sealed so a
# rented machine takes it. cloud/submit-gates.py matches it by sha256 against the job's commandSha256, so any change here
# is a new recipe, and jobs of the old one stop counting.
gate_base
gate_preamble repositories sandboxed_repositories
gate_preamble npm sandboxed_npm stage3/api
gate_preamble typescript sandboxed_typescript_source
python3 cloud/admission-corpus/manifest.py --sha HEAD > "$gateScratch/manifest.json" 2> "$gateScratch/manifest.stderr" ||
	{ tail -20 "$gateScratch/manifest.stderr"; gate_abort manifest "cloud/admission-corpus/manifest.py --sha HEAD"; }
gate_pass "manifest at the head"
began=$SECONDS
rc=0
go run ./cmd/adamic-admission-delta --base "$gateBase" --head HEAD --manifest "$gateScratch/manifest.json" \
	--manifest-generator cloud/admission-corpus/manifest.py --json > "$gateScratch/delta.json" 2> "$gateScratch/delta.stderr" || rc=$?
printf 'admission delta exit=%s wall_seconds=%s\n' "$rc" "$((SECONDS - began))"
grep -E '^(admitted|verdict|crashes fixed|crash fixed)' "$gateScratch/delta.stderr" | head -60
if python3 -I - "$gateScratch/delta.json" << 'PYTHON'; then
import json, sys
try:
    delta = json.load(open(sys.argv[1]))
except (OSError, ValueError) as error:
    print(f'  the delta wrote no JSON to read: {error}')
    sys.exit(1)
paths = delta.get('admitted_paths') or []
print(f"  verdict {delta.get('verdict')}, admitted {delta.get('admitted')}, admitted_paths listed {len(paths)}")
for path in paths[:40]:
    print(f'    admitted {path}')
sys.exit(0 if delta.get('verdict') == 'pass' and isinstance(delta.get('admitted'), int) and len(paths) == delta['admitted'] else 1)
PYTHON
	gate_pass "admission delta: verdict pass, its admitted programs listed"
else
	tail -20 "$gateScratch/delta.stderr"
	gate_fail delta "admission delta against $gateBase (exit $rc): its verdict isn't pass, or its admitted programs aren't all listed"
fi
gate_end
