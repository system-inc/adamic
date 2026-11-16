#!/usr/bin/env bash
# scratch-72/job.sh: scratch for #m2wxyaz, never committed. The proof job for adaptation 72-readonly-diagnostic-slots,
# run sealed from the tree's root, after the caller's preamble:
#
#   source /opt/adamic-block/environment.sh && source cloud/sealed-job.sh && sealed_typescript_source &&
#     sealed_npm stage3/api stage3/typescript && sealed_git && bash scratch-72/job.sh
#
#   a. regenerate: stage3/lane/generate-api.py rewrites sanctioned-api.json from the pinned TypeScript and every
#      adaptation's api-sanctions.json, then shard-plan.json's runtime is recomputed by shard_plan.runtime_hash() and
#      generate_shards.py rewrites and checks the Go declarations. Prints the whole patch between "=== PATCH" and
#      "=== END PATCH", and the rows of normalized_changes added and removed. PASS is exactly the three readonly
#      members added and nothing else moved.
#   b. lane: stage3-lane-byte-comparisons on the after tree with the regenerated sanction, run by shard.py against a
#      tree built by hereby with stage3/typescript's install, as the gate's product path runs it. It must pass. A
#      control runs the same unit on the mutant tree, where it must fail naming exactly the three members as missing.
#   c. mutant: this tree's stage3 with 72's directory dropped, applied, and both trees' src/compiler measured by the
#      meter's latent census (first error, scoped tail) through report.latent_summary. Family 4 is 192 on the mutant
#      and 0 on the real tree; the latent total is printed for both.
#   d. adapted tree: go test ./stage3/lane/adaptedtree -run '^TestProduct_Stage3AdaptedTree$', its whole log printed
#      when red.
#
# Verdict lines begin "scratch-72: verdict". Logs go to /tmp/scratch-72-<stamp>; the heavy work goes under $TMPDIR,
# never beside the tree. Exit 0 only when every verdict passes; 64 to 66 are preparation failures.
set -u

root=$PWD
[ -f "$root/stage3/apply.sh" ] && [ -f "$root/stage3/adapt/72-readonly-diagnostic-slots/adapt.cjs" ] || {
	echo "scratch-72: $root is not an adamic tree's root carrying 72-readonly-diagnostic-slots, so nothing ran"
	exit 64
}
[ "$(git -C "$root" rev-parse --show-toplevel 2> /dev/null)" = "$(cd "$root" && pwd -P)" ] || {
	echo "scratch-72: $root has no git of its own; run sealed_git first, so the patch can be read"
	exit 64
}
[ -z "$(git -C "$root" status --porcelain)" ] || {
	echo "scratch-72: the tree doesn't read clean before the job, so the patch couldn't be told apart; nothing ran"
	git -C "$root" status --porcelain | head -n 20
	exit 64
}
stamp=$(date -u +%Y%m%dT%H%M%SZ)
out=/tmp/scratch-72-$stamp
mkdir -p "$out" || exit 64
work=$(mktemp -d "${TMPDIR:-/tmp}/scratch-72-work.XXXXXX") || { echo "scratch-72: couldn't make a work directory under ${TMPDIR:-/tmp}"; exit 64; }
echo "scratch-72: logs in $out, work in $work"
# No __pycache__ lands in the tree, so git diff holds only what the regeneration wrote.
export PYTHONDONTWRITEBYTECODE=1

missing=""
for tool in go gofmt node python3 git; do
	command -v "$tool" > /dev/null 2>&1 || missing="$missing $tool"
done
[ -z "$missing" ] || { echo "scratch-72: missing from the toolchain:$missing"; exit 65; }
node --version | grep -q '^v24\.' || { echo "scratch-72: node is $(node --version), adaptations need Node 24"; exit 65; }
echo "scratch-72: $(uname -n), $(nproc) threads, tree $(git -C "$root" rev-parse --short HEAD), node $(node --version), $(go version)"

# step <name> <command...>: the command's whole output in $out/<name>.log, and its exit.
step() {
	local name=$1
	shift
	"$@" > "$out/$name.log" 2>&1
	local code=$?
	echo "scratch-72: $name exited $code ($out/$name.log)"
	return "$code"
}

# say <line>: a line printed and kept in $out/verdicts.log for the summary and the exit.
say() { echo "$*" | tee -a "$out/verdicts.log"; }

pin=$(python3 -I -c 'import json, sys; print(json.load(open(sys.argv[1]))["commit"])' "$root/stage3/source.json")
[ -n "${ADAMIC_TYPESCRIPT_SOURCE:-}" ] && [ "$(git -C "$ADAMIC_TYPESCRIPT_SOURCE" rev-parse HEAD 2> /dev/null)" = "$pin" ] || {
	echo "scratch-72: ADAMIC_TYPESCRIPT_SOURCE (${ADAMIC_TYPESCRIPT_SOURCE:-unset}) isn't a checkout at the pin $pin; run sealed_typescript_source"
	exit 66
}
modules="$root/stage3/api/node_modules"
[ "$(node -p 'require(process.argv[1] + "/typescript/package.json").version' "$modules" 2> /dev/null)" = 6.0.3 ] || {
	echo "scratch-72: stage3/api's typescript isn't laid at 6.0.3; run sealed_npm stage3/api stage3/typescript"
	exit 66
}
install="$root/stage3/typescript/node_modules"
[ -f "$install/hereby/bin/hereby.js" ] || { echo "scratch-72: stage3/typescript/node_modules holds no hereby; run sealed_npm stage3/api stage3/typescript"; exit 66; }

# a. Regenerate the sanction and the shard plan's runtime in the tree, as generate-api.py's and shard_plan.py's own
# guards name the commands (test_sanctioned_api_freshness.py, generate_shards.py --check, test_shards.py).
git -C "$root" show HEAD:stage3/lane/sanctioned-api.json > "$out/a-sanctioned-before.json"
git -C "$root" show HEAD:stage3/lane/shard-plan.json > "$out/a-shard-plan-before.json"
regenerated=1
(cd "$root" && NODE_PATH="$modules" step a-generate-api python3 -B stage3/lane/generate-api.py "$ADAMIC_TYPESCRIPT_SOURCE") || regenerated=0
if [ "$regenerated" = 1 ]; then
	# The plan's runtime is the only field a sanction moves (6286add012, e977285dcf); shard_plan.py's own writer
	# serialization, json.dumps(indent=2) and a newline, is held first, so nothing else in the file can move.
	(cd "$root" && step a-shard-runtime python3 -B - << 'PYTHON') || regenerated=0
import json, sys
sys.path.insert(0, 'stage3/lane')
import shard_plan
path = shard_plan.LANE / 'shard-plan.json'
text = path.read_text()
plan = json.loads(text)
if json.dumps(plan, indent=2) + '\n' != text:
    sys.exit('shard-plan.json is not in shard_plan.py\'s serialization, so rewriting it would move more than runtime')
before = plan['runtime']
plan['runtime'] = shard_plan.runtime_hash()
path.write_text(json.dumps(plan, indent=2) + '\n')
print(f'runtime {before} to {plan["runtime"]}')
PYTHON
fi
if [ "$regenerated" = 1 ]; then
	(cd "$root" && step a-generate-shards python3 -B stage3/lane/generate_shards.py) || regenerated=0
	(cd "$root" && step a-generate-shards-check python3 -B stage3/lane/generate_shards.py --check) || regenerated=0
fi
echo "=== PATCH"
git -C "$root" -c core.quotepath=off diff --no-color --no-ext-diff
echo "=== END PATCH"
git -C "$root" -c core.quotepath=off diff --stat --no-color
git -C "$root" status --porcelain > "$out/a-status.txt"
echo "scratch-72: a the tree's status after regeneration: $(tr '\n' ' ' < "$out/a-status.txt")"
python3 -I - "$out/a-sanctioned-before.json" "$root/stage3/lane/sanctioned-api.json" "$out/a-shard-plan-before.json" \
	"$root/stage3/lane/shard-plan.json" "$out/a-status.txt" "$regenerated" << 'PYTHON' | tee -a "$out/verdicts.log"
import json, sys
before, after, planBefore, planAfter = (json.load(open(path)) for path in sys.argv[1:5])
status = [line.rstrip('\n') for line in open(sys.argv[5]) if line.strip()]
regenerated = sys.argv[6] == '1'
problems = []
old = {row['declaration']: row for row in before['normalized_changes']}
new = {row['declaration']: row for row in after['normalized_changes']}
added = sorted(new.keys() - old.keys())
removed = sorted(old.keys() - new.keys())
changed = sorted(key for key in old.keys() & new.keys() if old[key] != new[key])
print(f"scratch-72: a normalized_changes {len(old)} to {len(new)}: {len(added)} added, {len(removed)} removed, {len(changed)} changed")
for key in added:
    print(f"scratch-72: a   + {key}\n                  before {new[key]['before']}\n                  after  {new[key]['after']}")
for key in removed:
    print(f"scratch-72: a   - {key}")
for key in changed:
    print(f"scratch-72: a   ~ {key}\n                  was {old[key]}\n                  now {new[key]}")
# Each of the three is upstream's property signature with ReadonlyKeyword before its name, and nothing else.
types = {'file': ['UnionType', ['TypeReference', ['Identifier', 'SourceFile']], ['UndefinedKeyword']],
         'start': ['UnionType', ['NumberKeyword'], ['UndefinedKeyword']],
         'length': ['UnionType', ['NumberKeyword'], ['UndefinedKeyword']]}
wanted = {f'member:["ts","DiagnosticRelatedInformation","{name}"]#0': name for name in types}
if added != sorted(wanted):
    problems.append(f'added {added}, want {sorted(wanted)}')
for key, name in wanted.items():
    row = new.get(key)
    if row is None or row['before'] is None or row['after'] is None:
        continue
    want = ['PropertySignature', ['Identifier', name], types[name]]
    if json.loads(row['before']) != want:
        problems.append(f'{name} before is {row["before"]}')
    if json.loads(row['after']) != [want[0], ['ReadonlyKeyword'], *want[1:]]:
        problems.append(f'{name} after is {row["after"]}')
if removed or changed:
    problems.append(f'{len(removed)} rows removed and {len(changed)} changed')
for field in ('source_commit', 'counts', 'excluded_method_conversions', 'baseline_diffs'):
    if before.get(field) != after.get(field):
        problems.append(f'{field} moved: {before.get(field)} to {after.get(field)}')
sanction = 'stage3/adapt/72-readonly-diagnostic-slots/api-sanctions.json'
oldProofs = {row['file']: row['sha256'] for row in before['proofs']}
newProofs = {row['file']: row['sha256'] for row in after['proofs']}
if set(newProofs) - set(oldProofs) != {sanction} or set(oldProofs) - set(newProofs) or any(
        oldProofs[file] != newProofs[file] for file in oldProofs.keys() & newProofs.keys()):
    problems.append(f'proofs moved beyond {sanction}')
def owner(row):
    return json.dumps(row, sort_keys=True)
oldOwners, newOwners = {owner(row) for row in before['owners']}, {owner(row) for row in after['owners']}
ownersAdded = [json.loads(row) for row in sorted(newOwners - oldOwners)]
print(f"scratch-72: a owners: {len(ownersAdded)} added, {len(oldOwners - newOwners)} removed")
for row in ownersAdded:
    print(f"scratch-72: a   + owner {row['path']} line {row['line']}: {row['adaptation'][:100]}")
if oldOwners - newOwners or len(ownersAdded) != 3 or {tuple(row['path']) for row in ownersAdded} != {
        ('DiagnosticRelatedInformation', name) for name in types}:
    problems.append('owners moved beyond the three')
diffAdded = [line for line in after['diff_lines'] if line.startswith('+') and not line.startswith('+++')]
diffRemoved = [line for line in after['diff_lines'] if line.startswith('-') and not line.startswith('---')]
print(f"scratch-72: a diff_lines: {len(before['diff_lines'])} to {len(after['diff_lines'])} lines, "
      f"now {len(diffAdded)} added and {len(diffRemoved)} removed")
planMoved = sorted(key for key in planBefore.keys() | planAfter.keys() if planBefore.get(key) != planAfter.get(key))
print(f"scratch-72: a shard-plan.json fields moved: {planMoved}")
if planMoved != ['runtime']:
    problems.append(f'shard-plan.json moved {planMoved}, want runtime alone')
files = sorted(line[3:] for line in status)
if files != ['stage3/lane/sanctioned-api.json', 'stage3/lane/shard-plan.json'] or any(not line.startswith(' M ') for line in status):
    problems.append(f'the tree changed {status}, want sanctioned-api.json and shard-plan.json modified alone')
if len(new) != 276:
    problems.append(f'{len(new)} sanctioned changes, where test_check.py now counts 276')
if not regenerated:
    problems.insert(0, 'a regeneration step failed')
print(f"scratch-72: verdict a regenerate {'PASS' if not problems else 'FAIL'}: "
      + ('exactly the three readonly members added, nothing else moved' if not problems else '; '.join(problems)))
PYTHON
# The lane's own python guards on the regenerated files: the count, the freshness against generation, the runtime.
(cd "$root/stage3/lane" && step a-lane-guards python3 -B -m unittest -v \
	test_check.LandingLaneTests.test_manifest_includes_only_proved_property_handoffs \
	test_sanctioned_api_freshness \
	test_shards.ShardTests.test_plan_runtime_is_this_trees_runtime_hash \
	test_shards.ShardTests.test_manifest_exact_schema_and_counts)
guards=$?
if [ "$guards" = 0 ]; then
	say "scratch-72: verdict a lane guards PASS: test_check's count, sanction freshness and the plan's runtime hold"
else
	say "scratch-72: verdict a lane guards FAIL: exited $guards"
	tail -n 40 "$out/a-lane-guards.log"
fi

# d starts here, in the background, once a has read the tree's status: the product build is its own apply and hereby,
# under the build cache sealed-job.sh keeps out of git, and it reads none of stage3/lane's files.
(cd "$root" && step d-adapted-tree go test ./stage3/lane/adaptedtree -run '^TestProduct_Stage3AdaptedTree$' -count=1 -timeout 40m -v < /dev/null) &
adaptedTree=$!

# c's mutant stage3: this tree's apply inputs with 72's directory left out, and nothing else changed.
mutant="$work/mutant/stage3"
mkdir -p "$mutant/adapt" || exit 64
for input in apply.sh apply.py caches.py source.json api; do
	cp -a "$root/stage3/$input" "$mutant/" || exit 64
done
for entry in "$root/stage3/adapt"/* "$root/stage3/adapt"/.[!.]*; do
	[ -e "$entry" ] || [ -L "$entry" ] || continue
	[ "$(basename "$entry")" = 72-readonly-diagnostic-slots ] || cp -a "$entry" "$mutant/adapt/" || exit 64
done
diff -rq "$root/stage3/adapt" "$mutant/adapt" > "$out/c-mutant-inputs.diff" 2>&1
echo "scratch-72: c the mutant's adapt differs from this tree's by: $(tr '\n' ' ' < "$out/c-mutant-inputs.diff")"
echo "scratch-72: product keys: after $(cd "$root" && python3 -B stage3/apply.py --key), mutant $(python3 -B "$mutant/apply.py" --key)"

# Both applies from a fresh local product store, so each builds its tree from the pinned upstream here, and the census
# tool, in parallel.
after="$work/after"
before="$work/before"
export STAGE3_PRODUCT_STORE="$work/products"
mkdir -p "$STAGE3_PRODUCT_STORE"
(cd "$root" && STAGE3_CACHE="$work/cache-after" step apply-after bash stage3/apply.sh "$after") &
applyAfter=$!
(STAGE3_CACHE="$work/cache-before" step apply-before bash "$mutant/apply.sh" "$before") &
applyBefore=$!
(cd "$root" && step census-overlay python3 -B stage3/census/latent/make_overlay.py "$root" "$work/latent-overlay" &&
	step census-gofmt bash -c 'gofmt -w "$1"/*.go' _ "$work/latent-overlay" &&
	step census-build go build -buildvcs=false -overlay="$work/latent-overlay/overlay.json" -o "$work/latent-census" ./stage3/census/latent/tool) &
censusBuild=$!
wait "$applyAfter"
applyAfterExit=$?
wait "$applyBefore"
applyBeforeExit=$?
wait "$censusBuild"
censusBuildExit=$?
echo "scratch-72: apply after exited $applyAfterExit, apply before exited $applyBeforeExit, census build exited $censusBuildExit"
haveTrees=1
if [ "$applyAfterExit" != 0 ] || [ "$applyBeforeExit" != 0 ]; then
	haveTrees=0
	failed=$(grep -o 'stage3/adapt/[0-9][0-9]-[^/]*/adapt.cjs' "$out/apply-after.log" | tail -n 1)
	say "scratch-72: verdict apply FAIL: after exited $applyAfterExit${failed:+ in $failed}, before exited $applyBeforeExit"
	tail -n 30 "$out/apply-after.log" "$out/apply-before.log"
	say "scratch-72: verdict b lane SKIP: no trees"
	say "scratch-72: verdict c mutant SKIP: no trees"
else
	say "scratch-72: verdict apply PASS: both applies exited 0; after's row: $(grep -F '72-readonly-diagnostic-slots' "$after/patch-set.md")"
	# The sources: src before and after differ only in types.ts, by `readonly ` before the three members.
	python3 -I - "$before" "$after" "$out/c-sources.json" << 'PYTHON' | tee -a "$out/verdicts.log"
import hashlib, json, sys
from pathlib import Path
before, after, report = Path(sys.argv[1]), Path(sys.argv[2]), Path(sys.argv[3])
def manifest(tree):
    root = tree / 'src'
    return {p.relative_to(root).as_posix(): hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(root.rglob('*')) if p.is_file()}
old, new = manifest(before), manifest(after)
differing = sorted(name for name in old.keys() | new.keys() if old.get(name) != new.get(name))
oldLines = (before / 'src/compiler/types.ts').read_bytes().decode().split('\n')
newLines = (after / 'src/compiler/types.ts').read_bytes().decode().split('\n')
changed = [(i + 1, a, b) for i, (a, b) in enumerate(zip(oldLines, newLines)) if a != b]
slots = {'file: SourceFile | undefined;', 'start: number | undefined;', 'length: number | undefined;'}
def inserted(a, b):
    indent = len(a) - len(a.lstrip())
    return b == a[:indent] + 'readonly ' + a[indent:]
ok = (differing == ['compiler/types.ts'] and len(oldLines) == len(newLines) and len(changed) == 3 and
      all(inserted(a, b) for _, a, b in changed) and {a.strip() for _, a, _ in changed} == slots)
json.dump({'files': len(new), 'differing': differing, 'types_ts_changed_lines': changed}, open(report, 'w'), indent=2)
print(f"scratch-72: verdict c sources {'PASS' if ok else 'FAIL'}: {len(new)} src files, differing {differing}, "
      f"types.ts lines changed {[(line, b.strip()) for line, _, b in changed]}")
PYTHON
fi

# c. The latent census as three-slots.sh ran it for J-4m3dv80 (first error, scoped tail, no output), on both trees in
# parallel, before any build writes into them.
if [ "$haveTrees" = 1 ] && [ "$censusBuildExit" != 0 ]; then
	say "scratch-72: verdict c mutant FAIL: the latent census didn't build"
	tail -n 30 "$out/census-overlay.log" "$out/census-gofmt.log" "$out/census-build.log" 2> /dev/null
elif [ "$haveTrees" = 1 ]; then
	mkdir -p "$work/census-before" "$work/census-after"
	(LATENT_TIMES="$work/census-before/latent-times.jsonl" LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TAIL=1 \
		step c-census-before "$work/latent-census" "$before/src/compiler" "$work/census-before/latent.jsonl") &
	censusBefore=$!
	(LATENT_TIMES="$work/census-after/latent-times.jsonl" LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TAIL=1 \
		step c-census-after "$work/latent-census" "$after/src/compiler" "$work/census-after/latent.jsonl") &
	censusAfter=$!
	wait "$censusBefore"
	censusBeforeExit=$?
	wait "$censusAfter"
	censusAfterExit=$?
	gzip -c "$work/census-before/latent.jsonl" > "$out/c-before.jsonl.gz" 2> /dev/null
	gzip -c "$work/census-after/latent.jsonl" > "$out/c-after.jsonl.gz" 2> /dev/null
	python3 -I - "$root/stage3/meter" "$before" "$work/census-before" "$after" "$work/census-after" "$out" "$censusBeforeExit" "$censusAfterExit" << 'PYTHON' | tee -a "$out/verdicts.log"
import json, re, sys
from collections import Counter
from pathlib import Path
sys.path.insert(0, sys.argv[1])
import report  # The meter's own reader: its coverage checks and its unique-site count.
beforeTree, beforeRun, afterTree, afterRun = map(Path, sys.argv[2:6])
out, codes = Path(sys.argv[6]), sys.argv[7:9]
# Family 4 as w6-job.sh and three-slots.sh counted it for J-4m3dv80's 192: a reason seen as a type naming Diagnostic or
# DiagnosticRelatedInformation, which can write SourceFile | undefined where SourceFile is read.
family4 = re.compile(r'seen as (.*), which can write SourceFile \| undefined where SourceFile is read$')
def measure(tree, run):
    summary = report.latent_summary(tree, run)
    reasons = Counter(summary['per_reason'])
    members = {key: count for key, count in reasons.items()
               if (m := family4.search(key)) and re.search(r'\bDiagnostic(RelatedInformation)?\b', m.group(1))}
    return {'totals': summary['totals'], 'latent_total': summary['totals']['NotYet'] + summary['totals']['Refused'],
            'source_files': summary['source_files'], 'family4': sum(members.values()), 'family4_reasons': members}, reasons
try:
    before, beforeReasons = measure(beforeTree, beforeRun)
    after, afterReasons = measure(afterTree, afterRun)
except Exception as error:
    print(f"scratch-72: verdict c mutant FAIL: a census didn't read whole (exits {codes}): {error}")
    sys.exit(0)
for name, r in (('mutant (72 dropped)', before), ('real tree', after)):
    print(f"scratch-72: c {name}: {r['source_files']} files, latent total {r['latent_total']} {r['totals']}, family 4 {r['family4']}")
    for key, count in sorted(r['family4_reasons'].items(), key=lambda item: (-item[1], item[0])):
        print(f"scratch-72: c   {name} family 4 {count} {key}")
delta = Counter(afterReasons)
delta.subtract(beforeReasons)
grown = sorted(((count, key) for key, count in delta.items() if count > 0), key=lambda item: (-item[0], item[1]))
shrunk = sorted(((count, key) for key, count in delta.items() if count < 0), key=lambda item: (item[0], item[1]))
print(f"scratch-72: c reasons that appear or grow on the real tree: {len(grown)}, {sum(c for c, _ in grown)} sites "
      f"(J-4m3dv80 found three, each whose readonly field file becomes writable, owed to the subtype slice)")
for count, key in grown:
    print(f"scratch-72: c   +{count} {key}")
print(f"scratch-72: c reasons that shrink or go on the real tree: {len(shrunk)}, {sum(c for c, _ in shrunk)} sites")
for count, key in shrunk[:60]:
    print(f"scratch-72: c   {count} {key}")
ok = before['family4'] == 192 and after['family4'] == 0
print(f"scratch-72: verdict c mutant {'PASS' if ok else 'FAIL'}: family 4 {before['family4']} with 72 dropped (want 192), "
      f"{after['family4']} on the real tree (want 0); latent total {before['latent_total']} with 72 dropped, "
      f"{after['latent_total']} on the real tree (J-4m3dv80: 4138 and 3940)")
(out / 'c-summary.json').write_text(json.dumps({'mutant': before, 'real': after,
    'reason_delta': {key: count for key, count in delta.items() if count}}, indent=2) + '\n')
PYTHON
fi

# b. The byte-comparisons unit as the gate's product path runs it: shard.py with STAGE3_ADAPTED_TREE naming a tree built
# by hereby local and hereby tests with stage3/typescript's install linked in, which the product removes after.
buildTree() {
	local tree=$1 label=$2 real code
	real=$(cd "$install" && pwd -P)
	ln -s "$real" "$tree/node_modules" || return 1
	mkdir -p "$work/home-$label"
	HOME="$work/home-$label" XDG_CONFIG_HOME="$work/home-$label" NODE_DISABLE_COMPILE_CACHE=1 \
		bash -c 'cd "$1" && node "$2/hereby/bin/hereby.js" local && node "$2/hereby/bin/hereby.js" tests' _ "$tree" "$real" \
		> "$out/b-build-$label.log" 2>&1
	code=$?
	unlink "$tree/node_modules"
	echo "scratch-72: b-build-$label exited $code ($out/b-build-$label.log)"
	[ "$code" = 0 ] && [ -f "$tree/built/local/run.js" ]
}
# runUnit <tree> <label>: the unit's report in $out/b-unit-<label>.json.
runUnit() {
	local tree=$1 label=$2 results="$work/lane-$2"
	mkdir -p "$results"
	(cd "$root" && STAGE3_ADAPTED_TREE="$tree" STAGE3_RESULTS="$results" STAGE3_SHARD_SECONDS="${SCRATCH72_SHARD_SECONDS:-300}" \
		NODE_DISABLE_COMPILE_CACHE=1 step "b-unit-$label" python3 -E -S -B stage3/lane/shard.py stage3-lane-byte-comparisons "$results")
	cp "$results/stage3-lane-byte-comparisons/report.json" "$out/b-unit-$label.json" 2> /dev/null
	cp "$results/stage3-lane-byte-comparisons/oracle/baseline.diff" "$out/b-unit-$label-baseline.diff" 2> /dev/null
}
if [ "$haveTrees" = 1 ]; then
	(buildTree "$after" after) &
	buildAfter=$!
	(buildTree "$before" before) &
	buildBefore=$!
	wait "$buildAfter"
	buildAfterExit=$?
	wait "$buildBefore"
	buildBeforeExit=$?
	if [ "$buildAfterExit" != 0 ]; then
		say "scratch-72: verdict b lane FAIL: the after tree didn't build with hereby; see $out/b-build-after.log"
		tail -n 30 "$out/b-build-after.log"
	else
		runUnit "$after" after
		[ "$buildBeforeExit" = 0 ] && runUnit "$before" mutant
		python3 -I - "$root/stage3/lane/shards.json" "$out" "$buildBeforeExit" << 'PYTHON' | tee -a "$out/verdicts.log"
import json, sys
from pathlib import Path
units = {unit['name']: unit for unit in json.load(open(sys.argv[1]))['units']}
expect = units['stage3-lane-byte-comparisons']['expect']
out = Path(sys.argv[2])
def read(label):
    path = out / f'b-unit-{label}.json'
    return json.loads(path.read_text()) if path.exists() else None
after = read('after')
if after is None:
    print('scratch-72: verdict b lane FAIL: the unit wrote no report on the after tree')
else:
    ok = after['status'] == 'pass' and after['counts'] == expect
    print(f"scratch-72: verdict b lane {'PASS' if ok else 'FAIL'}: stage3-lane-byte-comparisons {after['status']} on the after tree, "
          f"counts {after['counts']} where the manifest expects {expect}, api {after.get('api')}")
    for error in after['errors']:
        print(f"scratch-72: b   after: {error.splitlines()[0][:300] if error else error}")
# The control: the mutant tree lacks the three readonly members, so the regenerated sanction must name each one missing.
wanted = sorted(f'missing sanctioned API declaration: member:["ts","DiagnosticRelatedInformation","{name}"]#0'
                for name in ('file', 'start', 'length'))
mutant = read('mutant')
if sys.argv[3] != '0' or mutant is None:
    print(f"scratch-72: verdict b control FAIL: the mutant tree {'did not build' if sys.argv[3] != '0' else 'wrote no unit report'}")
else:
    errors = sorted(mutant['errors'])
    ok = mutant['status'] == 'fail' and errors == wanted
    print(f"scratch-72: verdict b control {'PASS' if ok else 'FAIL'}: on the mutant tree the unit is {mutant['status']} with "
          f"{len(errors)} errors{', exactly the three members missing' if ok else ''}")
    if not ok:
        for error in errors:
            print(f"scratch-72: b   mutant: {error.splitlines()[0][:300] if error else error}")
PYTHON
	fi
fi

# d. The adapted tree product, started after a.
wait "$adaptedTree"
adaptedTreeExit=$?
if [ "$adaptedTreeExit" = 0 ]; then
	say "scratch-72: verdict d adapted tree PASS: TestProduct_Stage3AdaptedTree in ./stage3/lane/adaptedtree"
else
	say "scratch-72: verdict d adapted tree FAIL: go test exited $adaptedTreeExit; its whole log follows"
	cat "$out/d-adapted-tree.log"
fi

echo "scratch-72: summary, $SECONDS s:"
grep '^scratch-72: verdict ' "$out/verdicts.log"
echo "scratch-72: evidence in $out; work left in $work"
grep -q '^scratch-72: verdict .* \(FAIL\|SKIP\):' "$out/verdicts.log" && exit 1
exit 0
