source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh && source cloud/gates/steps.sh && gate_begin lane-whole-sealed || exit 1
# ask: --size F64 --ceiling 120m
# lane-whole-sealed: stage3/lane whole, sealed (cloud/gates/diff-to-gates.json). Seats' proven form for family 7's head
# (J-nkj6ja4), at the head's own sha with no patch laid: stage3/lane's products settled, go test over stage3/lane and
# stage3/lane/adaptedtree with every TestShard_ of shards.json passing, the shard plan's runtime digest this tree's, and
# the lane's Python suite whole with no parser override.
# This file is the whole command of a Loom sealed job at the head's own sha, from its first line, which opens sealed so a
# rented machine takes it. cloud/submit-gates.py matches it by sha256 against the job's commandSha256, so any change here
# is a new recipe, and jobs of the old one stop counting.
export STAGE3_CACHE="$gateScratch/stage3-cache"
unset STAGE3_PRODUCT_STORE ADAMIC_BUILD_CACHE_URL GOCACHEPROG
mkdir -p "$STAGE3_CACHE"
gate_preamble repositories sandboxed_repositories
gate_preamble npm sandboxed_npm stage3/api stage3/typescript
gate_preamble typescript sandboxed_typescript_source
if python3 -I -B -c '
import sys
sys.path.insert(0, "stage3/lane")
from shard_plan import load_plan, runtime_hash
plan, tree = load_plan()["runtime"], runtime_hash()
print(f"  shard-plan.json runtime {plan}, tree {tree}")
sys.exit(plan != tree)'; then
	gate_pass "shard-plan.json's runtime digest is this tree's"
else
	gate_fail digest "shard-plan.json's runtime digest isn't this tree's runtime_hash()"
fi
gate_products stage3/lane
began=$SECONDS
rc=0
go test -count=1 -timeout 60m -json ./stage3/lane/... > "$gateScratch/lane.jsonl" 2> "$gateScratch/lane.stderr" || rc=$?
printf 'go test ./stage3/lane/... exit=%s wall_seconds=%s\n' "$rc" "$((SECONDS - began))"
summary=0
python3 -I - "$gateScratch/lane.jsonl" stage3/lane/shards.json << 'PYTHON' || summary=$?
import json, sys
events_path, manifest_path = sys.argv[1:3]
results, outputs, packages = {}, {}, {}
for line in open(events_path, errors='replace'):
    try:
        event = json.loads(line)
    except ValueError:
        continue
    test, action, package = event.get('Test'), event.get('Action'), event.get('Package', '')
    if action == 'output' and test:
        outputs.setdefault((package, test), []).append(event.get('Output', ''))
    if action in ('pass', 'fail', 'skip'):
        if test is None:
            packages[package] = action
        elif '/' not in test:
            results[(package, test)] = action
units = len(json.load(open(manifest_path))['units'])
counted = list(results.values())
shards = sum(1 for (p, t), a in results.items() if t.startswith('TestShard_') and a == 'pass')
print(f'  packages: {json.dumps(packages, sort_keys=True)}')
print(f'  top-level tests: {counted.count("pass")} pass, {counted.count("fail")} fail, {counted.count("skip")} skip; '
      f'TestShard_ passed {shards} of {units} units')
for (package, test), action in sorted(results.items()):
    if action == 'skip':
        print(f'  skipped: {package} {test}')
failed = sorted(key for key, action in results.items() if action == 'fail')
for package, test in failed[:10]:
    print(f'  --- FAIL {package} {test}, last 30 lines:')
    print(''.join(outputs.get((package, test), [])[-30:]).rstrip())
sys.exit(0 if packages and all(a == 'pass' for a in packages.values()) and not failed and shards == units else 1)
PYTHON
if [ "$rc" -eq 0 ] && [ "$summary" -eq 0 ]; then
	gate_pass "lane whole: go test ./stage3/lane/..., every shard"
else
	tail -20 "$gateScratch/lane.stderr"
	gate_fail lane "go test ./stage3/lane/... (exit $rc, summary $summary)"
fi
began=$SECONDS
if env -u CENSUS_TYPESCRIPT -u TSC_ADAPT_TYPESCRIPT PYTHONDONTWRITEBYTECODE=1 STAGE3_CACHE="$gateScratch/discover-cache" \
	python3 -B -m unittest discover -s stage3/lane -p 'test_*.py' > "$gateScratch/python.log" 2>&1; then
	gate_pass "the lane's Python suite: $(grep -E '^Ran [0-9]+ tests' "$gateScratch/python.log" | tail -1) ($((SECONDS - began)) s)"
else
	grep -E '^(Ran |FAILED|ERROR:|FAIL:)' "$gateScratch/python.log" | head -20
	gate_fail python "unittest discover -s stage3/lane"
fi
gate_end
