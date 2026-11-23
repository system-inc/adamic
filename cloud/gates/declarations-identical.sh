source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh && source cloud/gates/steps.sh && gate_begin declarations-identical || exit 1
# ask: --size F64 --ceiling 90m
# declarations-identical: the adapted tree's public .d.ts byte-identical to stock, or the head carries Adamic's ruling on
# the declaration diff (cloud/gates/diff-to-gates.json). Read as the lane reads it: stage3/lane/sanctioned-api.json holds
# every public declaration difference from stock that Adamic has sanctioned, so a head with no public line leaves it the
# base's byte for byte, and the lane's public shard (stage3-lane-byte-comparisons) composes the adapted tree's
# declarations and finds exactly the sanctioned ones, no more. Seats' step 6 for family 7 (J-nkj6ja4). A head that moves
# the sanction is red here by design: its ruling is read by hand.
# This file is the whole command of a Loom sealed job at the head's own sha, from its first line, which opens sealed so a
# rented machine takes it. cloud/submit-gates.py matches it by sha256 against the job's commandSha256, so any change here
# is a new recipe, and jobs of the old one stop counting.
gate_base
export STAGE3_CACHE="$gateScratch/stage3-cache"
unset STAGE3_PRODUCT_STORE ADAMIC_BUILD_CACHE_URL GOCACHEPROG
mkdir -p "$STAGE3_CACHE"
git show "$gateBase:stage3/lane/sanctioned-api.json" > "$gateScratch/sanctioned-api.base.json" ||
	gate_abort sanction "git show of the base's stage3/lane/sanctioned-api.json"
if cmp -s "$gateScratch/sanctioned-api.base.json" stage3/lane/sanctioned-api.json; then
	gate_pass "sanctioned-api.json is the base's, byte for byte"
else
	gate_fail sanction "sanctioned-api.json differs from the base's: a public declaration difference, Adamic's to rule"
fi
gate_preamble repositories sandboxed_repositories
gate_preamble npm sandboxed_npm stage3/api stage3/typescript
gate_preamble typescript sandboxed_typescript_source
gate_products stage3/lane
go test -count=1 -json -run '^TestPerformanceCoreYewmv46PinsEqualAdaptedProductBytes$' ./stage3/lane/adaptedtree/ \
	> "$gateScratch/product.jsonl" 2>&1
product=$(python3 -I -c '
import json, re, sys
for line in open(sys.argv[1], errors="replace"):
    try:
        found = re.search(r"adapted product: (\S+)", json.loads(line).get("Output", ""))
    except ValueError:
        continue
    if found:
        print(found.group(1))
        break' "$gateScratch/product.jsonl")
[ -n "$product" ] && [ -d "$product/tree" ] || gate_abort product "the adapted tree's product path wasn't logged"
gate_pass "adapted tree $product"
rc=0
STAGE3_ADAPTED_TREE="$product/tree" STAGE3_RESULTS="$gateScratch/public-shard" STAGE3_SHARD_SECONDS=55 \
	python3 -E -S -B stage3/lane/shard.py stage3-lane-byte-comparisons "$gateScratch/public-shard" > "$gateScratch/public-shard.log" 2>&1 ||
	rc=$?
printf 'public shard exit=%s\n' "$rc"
if python3 -I - "$gateScratch/public-shard/stage3-lane-byte-comparisons/report.json" "$gateScratch/sanctioned-api.base.json" \
	stage3/lane/shards.json stage3/lane/expected.json << 'PYTHON'; then
import json, sys
report_path, base_path, manifest_path, expected_path = sys.argv[1:5]
ok = True
def check(label, good, shown):
    global ok
    ok = ok and good
    print(f'  {"ok" if good else "wrong"} {label}: {shown}')
report = json.load(open(report_path))
expected = json.load(open(expected_path))
units = [unit for unit in json.load(open(manifest_path))['units'] if unit['name'] == 'stage3-lane-byte-comparisons']
base = json.load(open(base_path))['normalized_changes']
check('status', report.get('status') == 'pass', report.get('status'))
check('errors', report.get('errors') == [], report.get('errors'))
check('counts (shards.json expect)', len(units) == 1 and report.get('counts') == units[0]['expect'],
      [report.get('counts'), [unit['expect'] for unit in units]])
check('failed tests (expected.json)', report.get('failed_tests') == expected['failed_tests'], report.get('failed_tests'))
check('baseline diffs (expected.json)', report.get('baseline_diffs') == expected['baseline_diffs'], report.get('baseline_diffs'))
api = report.get('api') or {}
check('declarations (composed, sanctioned, the base sanction rows)',
      api.get('composed_declarations') == api.get('sanctioned_declarations') == len(base),
      [api.get('composed_declarations'), api.get('sanctioned_declarations'), len(base)])
sys.exit(0 if ok else 1)
PYTHON
	gate_pass "public shard: no declaration difference beyond the base's sanction"
else
	tail -20 "$gateScratch/public-shard.log"
	gate_fail public "public shard: its report or its declarations differ from the base's sanction"
fi
gate_end
