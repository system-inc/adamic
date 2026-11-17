#!/usr/bin/env bash
# factory.sh: the measurement for #v879py5's adaptation, a scratch script that never lands. 26 latent refusals, all in
# nodeFactory.ts, are the lazy getters of createNodeFactory's factory literal handing out a function whose comment
# parameter is NodeArray<JSDocComment> where NodeFactory's member takes string | NodeArray<JSDocComment>.
# stage3/adapt/96-emit-jsdoc-tag-comment-parameters widens the four memoized arrows those getters call to the type of the
# workers they forward to. Compiler accepts it on: the family at 0 with 0 new TS diagnostics, the latent total, the lane
# green, the adaptation dropped bringing the 26 back, and whether T's inference at the getters moves anything else. Run
# from an adamic tree's root, on a sealed job as
#
#   source /opt/adamic-block/environment.sh && source cloud/sealed-job.sh && export GOFLAGS=... && sealed_typescript_source &&
#   sealed_npm stage3/api stage3/typescript && ... && bash scratch-v879py5/factory.sh
#
#   before: stage3/apply.sh over a copy of this tree's stage3 inputs with 96-emit-jsdoc-tag-comment-parameters left out, which
#           is also the mutant: the adaptation dropped
#   after:  this tree's own stage3/apply.sh, which runs 96-emit after 91-patcher-discard, whose pin on createNodeFactory reads upstream text
#
#   sources: the two src trees differ only in src/compiler/factory/nodeFactory.ts, by `string | ` in the four arrows
#   (1) diagnostics: one ts.createProgram per tree over all upstream src, 70's options, stock TypeScript 6.0.3; PASS is
#       no diagnostic of any code new in after
#   (2) the meter: the latent census on both src/compiler trees, read by report.py's latent_summary. The family is every
#       reason naming JSDocComment; for it and for every reason that grows or shrinks, each site's where and text, before
#       and after. PASS is the family at 26 before (the mutant) and 0 after, and no other reason moved, which is the
#       answer on T's inference at the getters
#   (3) the lane, whole, on the after tree: every unit of stage3/lane/shards.json run by shard.py against the after tree
#       built by hereby with stage3/typescript's install, as the gate's product path runs it. 96 carries no API sanction,
#       since no public declaration moves, so stage3-lane-byte-comparisons must pass too
#   (4) emit: both trees built by hereby; PASS is every file under built/local byte-identical, so the JavaScript and
#       typescript.d.ts don't move
#   (5) the lane's own python guards, on a tree whose stage3/lane this adaptation leaves untouched
#   (6) go test ./stage3/lane/adaptedtree -run '^TestProduct_Stage3AdaptedTree$'
#
# Verdict lines begin "factory: verdict". Logs, JSON and the summary go to $TMPDIR/factory-<stamp>; the heavy trees go
# to $TMPDIR too (FACTORY_WORK names another place), never beside the tree, which a sealed job lays read-only, and are
# removed at the end unless FACTORY_KEEP=1. Exit 0 when every part passes, 1 when any fails, 3 when none fails and one
# is skipped; 64 to 66 are preparation failures.
set -u

root=$PWD
[ -f "$root/stage3/apply.sh" ] && [ -f "$root/stage3/adapt/96-emit-jsdoc-tag-comment-parameters/adapt.cjs" ] || {
	echo "factory: $root is not an adamic tree's root carrying 96-emit-jsdoc-tag-comment-parameters, so nothing ran"
	exit 64
}
stamp=$(date -u +%Y%m%dT%H%M%SZ)
scratch="${TMPDIR:-/tmp}"
out="${scratch%/}/factory-$stamp"
work="${FACTORY_WORK:-${scratch%/}/v879py5-$stamp.factory}"
mkdir -p "$out" || exit 64
mkdir "$work" || { echo "factory: $work is there already or can't be made, so nothing ran"; exit 64; }
jobEnd() {
	local code=$?
	trap - EXIT
	if [ "${FACTORY_KEEP:-0}" != 1 ]; then
		case "$work" in /*/*.factory) chmod -R u+w -- "$work" 2> /dev/null; rm -rf -- "$work" ;; esac
	fi
	echo "factory: evidence in $out"
	exit "$code"
}
trap jobEnd EXIT
trap 'exit 143' TERM
trap 'exit 129' HUP
trap 'exit 130' INT
mkdir "$work/tmp" || exit 64
export TMPDIR="$work/tmp"

if [ -f "$HOME/adamic-tools/env.sh" ]; then
	. "$HOME/adamic-tools/env.sh" || { echo "factory: $HOME/adamic-tools/env.sh didn't source, so nothing ran"; exit 65; }
fi
missing=""
for tool in go gofmt node python3 git; do
	command -v "$tool" > /dev/null 2>&1 || missing="$missing $tool"
done
[ -z "$missing" ] || { echo "factory: missing from the toolchain:$missing"; exit 65; }
hasNpm=0; command -v npm > /dev/null 2>&1 && hasNpm=1
node --version | grep -q '^v24\.' || { echo "factory: node is $(node --version), adaptations need Node 24"; exit 65; }
[ -n "$(ls -A "$root/cohere" 2> /dev/null)" ] || { echo "factory: cohere isn't laid in $root"; exit 66; }
echo "factory: $(uname -n), $(nproc) threads, $(git -C "$root" rev-parse HEAD 2> /dev/null || echo 'no checkout'), node $(node --version), $(go version), npm $([ "$hasNpm" = 1 ] && npm --version || echo absent)"

# step <name> <command...>: the command's whole output in $out/<name>.log, and its exit.
step() {
	local name=$1
	shift
	"$@" > "$out/$name.log" 2>&1
	local code=$?
	echo "factory: $name exited $code ($out/$name.log)"
	return "$code"
}

# say <line>: a line printed and kept in $out/verdicts.log for the summary and the exit.
say() { echo "$*" | tee -a "$out/verdicts.log"; }

# The pinned upstream checkout apply reads: the caller's, the laid cohere/TypeScript when it is the pin, else a clone.
pin() { python3 -I -c 'import json, sys; print(json.load(open(sys.argv[1]))[sys.argv[2]])' "$root/stage3/source.json" "$1"; }
pinCommit=$(pin commit)
if [ -n "${ADAMIC_TYPESCRIPT_SOURCE:-}" ] && [ "$(git -C "$ADAMIC_TYPESCRIPT_SOURCE" rev-parse HEAD 2> /dev/null)" = "$pinCommit" ]; then
	echo "factory: upstream is the caller's ADAMIC_TYPESCRIPT_SOURCE"
elif [ "$(git -C "$root/cohere/TypeScript" rev-parse HEAD 2> /dev/null)" = "$pinCommit" ]; then
	export ADAMIC_TYPESCRIPT_SOURCE="$root/cohere/TypeScript"
	echo "factory: upstream is the laid cohere/TypeScript, at the pin"
else
	step upstream-clone git clone --depth 1 --branch "$(pin tag)" "$(pin repository)" "$work/upstream" || exit 66
	export ADAMIC_TYPESCRIPT_SOURCE="$work/upstream"
fi
[ "$(git -C "$ADAMIC_TYPESCRIPT_SOURCE" rev-parse HEAD)" = "$pinCommit" ] || { echo "factory: upstream isn't at $pinCommit"; exit 66; }

# stage3/api's lockfile-pinned typescript@6.0.3, which apply puts on NODE_PATH for every adaptation.
modules="$root/stage3/api/node_modules"
[ "$(node -p 'require(process.argv[1] + "/typescript/package.json").version' "$modules" 2> /dev/null)" = 6.0.3 ] || { echo "factory: stage3/api's typescript isn't laid at 6.0.3"; exit 66; }

# Before's stage3: this tree's apply inputs with 96-emit-jsdoc-tag-comment-parameters left out, and nothing else changed.
mutant="$work/mutant/stage3"
mkdir -p "$mutant/adapt" || exit 64
for input in apply.sh apply.py caches.py source.json api; do
	cp -a "$root/stage3/$input" "$mutant/" || exit 64
done
for entry in "$root/stage3/adapt"/* "$root/stage3/adapt"/.[!.]*; do
	[ -e "$entry" ] || [ -L "$entry" ] || continue
	[ "$(basename "$entry")" = 96-emit-jsdoc-tag-comment-parameters ] || cp -a "$entry" "$mutant/adapt/" || exit 64
done
diff -rq "$root/stage3/adapt" "$mutant/adapt" > "$out/before-inputs.diff" 2>&1
echo "factory: before's adapt differs from this tree's by: $(tr '\n' ' ' < "$out/before-inputs.diff")"
echo "factory: product keys: after $(cd "$root" && python3 stage3/apply.py --key), before $(python3 "$mutant/apply.py" --key)"

# (5) The lane's own python guards. This adaptation carries no sanction, so stage3/lane is the tree's as committed and
# each guard reads it as main's own run does.
(cd "$root/stage3/lane" && step 5-lane-guards python3 -B -m unittest -v \
	test_check.LandingLaneTests.test_manifest_includes_only_proved_property_handoffs \
	test_sanctioned_api_freshness \
	test_shards.ShardTests.test_plan_runtime_is_this_trees_runtime_hash \
	test_shards.ShardTests.test_manifest_exact_schema_and_counts)
guards=$?
if [ "$guards" = 0 ]; then
	say "factory: verdict (5) lane guards PASS: test_check's count, sanction freshness and the plan's runtime hold"
else
	say "factory: verdict (5) lane guards FAIL: exited $guards"
	tail -n 40 "$out/5-lane-guards.log"
fi

# (6) The adapted tree product, in the background from here: its own apply and hereby, reading none of this run's trees.
(cd "$root" && step 6-adapted-tree go test ./stage3/lane/adaptedtree -run '^TestProduct_Stage3AdaptedTree$' -count=1 -timeout 40m -v < /dev/null) &
adaptedTree=$!

# Phase 1, in parallel: both applies from a fresh local product store, and the census tool. A local store makes each
# apply build its tree from the pinned upstream here; nothing is fetched as a product.
before="$work/before"
after="$work/after"
export STAGE3_PRODUCT_STORE="$work/products"
mkdir -p "$STAGE3_PRODUCT_STORE"
(STAGE3_CACHE="$work/cache-before" step apply-before bash "$mutant/apply.sh" "$before") &
applyBefore=$!
(cd "$root" && STAGE3_CACHE="$work/cache-after" step apply-after bash stage3/apply.sh "$after") &
applyAfter=$!
(cd "$root" && step census-overlay python3 stage3/census/latent/make_overlay.py "$root" "$work/latent-overlay" &&
	step census-gofmt bash -c 'gofmt -w "$1"/*.go' _ "$work/latent-overlay" &&
	step census-build go build -buildvcs=false -overlay="$work/latent-overlay/overlay.json" -o "$work/latent-census" ./stage3/census/latent/tool) &
censusBuild=$!
wait "$applyBefore"
applyBeforeExit=$?
wait "$applyAfter"
applyAfterExit=$?
wait "$censusBuild"
censusBuildExit=$?
echo "factory: apply before exited $applyBeforeExit, apply after exited $applyAfterExit, census build exited $censusBuildExit"
[ "$applyBeforeExit" = 0 ] || { echo "factory: before's apply failed, so there is no before tree; its last lines:"; tail -n 30 "$out/apply-before.log"; exit 66; }
if [ "$applyAfterExit" != 0 ]; then
	# apply.py names the adaptation it was running in the failed command.
	failed=$(grep -o 'stage3/adapt/[0-9][0-9]-[^/]*/adapt.cjs' "$out/apply-after.log" | tail -n 1)
	say "factory: verdict apply FAIL: the after tree's apply exited $applyAfterExit${failed:+ in $failed}, so there is no after tree"
	tail -n 30 "$out/apply-after.log"
	say "factory: verdict (1) diagnostics SKIP: no after tree"
	say "factory: verdict (2) meter SKIP: no after tree"
	say "factory: verdict (3) lane SKIP: no after tree"
	say "factory: verdict (4) emit SKIP: no after tree"
	exit 1
fi
say "factory: verdict apply PASS: both applies exited 0; after's patch-set row: $(grep -F '96-emit-jsdoc-tag-comment-parameters' "$after/patch-set.md")"

# The sources: src before and after differ only in factory/nodeFactory.ts, by `string | ` before the comment parameter's
# type in the four arrows, one line each, and nothing else on those lines.
python3 -I - "$before" "$after" "$out/sources.json" << 'PYTHON' | tee -a "$out/verdicts.log"
import hashlib, json, sys
from pathlib import Path
before, after, report = Path(sys.argv[1]), Path(sys.argv[2]), Path(sys.argv[3])
def manifest(tree):
    root = tree / 'src'
    return {p.relative_to(root).as_posix(): hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(root.rglob('*')) if p.is_file()}
old, new = manifest(before), manifest(after)
differing = sorted(name for name in old.keys() | new.keys() if old.get(name) != new.get(name))
name = 'src/compiler/factory/nodeFactory.ts'
oldLines = (before / name).read_bytes().decode().split('\n')
newLines = (after / name).read_bytes().decode().split('\n')
changed = [(i + 1, a, b) for i, (a, b) in enumerate(zip(oldLines, newLines)) if a != b]
arrows = sorted(['getJSDocSimpleTagCreateFunction', 'getJSDocSimpleTagUpdateFunction',
                 'getJSDocTypeLikeTagCreateFunction', 'getJSDocTypeLikeTagUpdateFunction'])
narrow, wide = 'comment?: NodeArray<JSDocComment>)', 'comment?: string | NodeArray<JSDocComment>)'
def arrow(line):
    words = line.split()
    return words[1] if len(words) > 2 and words[0] == 'const' and words[2] == '=' else None
def widened(a, b):
    return a.count(narrow) == 1 and b == a.replace(narrow, wide)
found = sorted(arrow(a) for _, a, _ in changed)
ok = (differing == ['compiler/factory/nodeFactory.ts'] and len(oldLines) == len(newLines) and len(changed) == 4 and
      all(widened(a, b) for _, a, b in changed) and found == arrows)
json.dump({'files': len(new), 'differing': differing, 'changed_lines': changed}, open(report, 'w'), indent=2)
print(f"factory: verdict sources {'PASS' if ok else 'FAIL'}: {len(new)} src files, differing {differing}, "
      f"nodeFactory.ts lines changed {[(line, arrow(a)) for line, a, _ in changed]}")
PYTHON

# (1) Diagnostics: the program 70's audit builds, on each tree, and every syntactic, global and semantic diagnostic.
# A message naming the tree's own path names it as <tree>, so the same diagnostic reads the same in both.
cat > "$out/diagnostics.cjs" << 'NODE'
'use strict';
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
const tree = path.resolve(process.argv[2]);
const output = path.resolve(process.argv[3]);
// Adaptation 70's options, so the program is the one its audit checks.
const options = {
    strict: true, exactOptionalPropertyTypes: true, noUncheckedIndexedAccess: true,
    verbatimModuleSyntax: true, erasableSyntaxOnly: false,
    allowImportingTsExtensions: true, noEmit: true,
    module: ts.ModuleKind.ESNext, moduleDetection: ts.ModuleDetectionKind.Force,
    moduleResolution: ts.ModuleResolutionKind.Bundler, target: ts.ScriptTarget.ES2024,
    lib: ['lib.es2024.d.ts'], types: [],
};
const roots = ts.sys.readDirectory(path.join(tree, 'src'), ['.ts'], ['**/lib/**']);
const program = ts.createProgram(roots, options);
const groups = { syntactic: program.getSyntacticDiagnostics(), global: program.getGlobalDiagnostics(), semantic: program.getSemanticDiagnostics() };
const rows = [];
for (const [group, diagnostics] of Object.entries(groups)) {
    for (const diagnostic of diagnostics) {
        const row = { group, code: diagnostic.code, category: ts.DiagnosticCategory[diagnostic.category],
            message: ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n').split(tree).join('<tree>') };
        if (diagnostic.file) {
            const point = diagnostic.file.getLineAndCharacterOfPosition(diagnostic.start ?? 0);
            const name = path.resolve(diagnostic.file.fileName);
            row.file = name.startsWith(tree + path.sep) ? path.relative(tree, name) : name;
            row.line = point.line + 1;
            row.column = point.character + 1;
        }
        rows.push(row);
    }
}
fs.writeFileSync(output, JSON.stringify({ typescript: ts.version, roots: roots.length, diagnostics: rows }) + '\n');
console.log(`${roots.length} roots, ${rows.length} diagnostics`);
NODE
(NODE_PATH="$modules" step 1-diagnostics-before node "$out/diagnostics.cjs" "$before" "$out/1-diagnostics-before.json") &
diagnosticsBefore=$!
(NODE_PATH="$modules" step 1-diagnostics-after node "$out/diagnostics.cjs" "$after" "$out/1-diagnostics-after.json") &
diagnosticsAfter=$!
wait "$diagnosticsBefore"
diagnosticsBeforeExit=$?
wait "$diagnosticsAfter"
diagnosticsAfterExit=$?
if [ "$diagnosticsBeforeExit" != 0 ] || [ "$diagnosticsAfterExit" != 0 ]; then
	say "factory: verdict (1) diagnostics FAIL: the program didn't build (exits $diagnosticsBeforeExit before, $diagnosticsAfterExit after)"
	tail -n 20 "$out/1-diagnostics-before.log" "$out/1-diagnostics-after.log"
else
	python3 -I - "$out/1-diagnostics-before.json" "$out/1-diagnostics-after.json" << 'PYTHON' | tee -a "$out/verdicts.log"
import json, sys
from collections import Counter
before, after = (json.load(open(path)) for path in sys.argv[1:3])
def key(row):
    return (row.get('file') or '', row.get('line') or 0, row.get('column') or 0, row['code'], row['message'])
old, new = Counter(map(key, before['diagnostics'])), Counter(map(key, after['diagnostics']))
added, removed = new - old, old - new
print(f"factory: (1) diagnostics: before {len(before['diagnostics'])} over {before['roots']} roots, "
      f"after {len(after['diagnostics'])} over {after['roots']} roots")
print(f"factory: (1) diagnostics new in after by code {dict(sorted(Counter(k[3] for k in added.elements()).items()))}, "
      f"gone from after by code {dict(sorted(Counter(k[3] for k in removed.elements()).items()))}")
for label, rows in (('new', added), ('gone', removed)):
    for file, line, column, code, message in sorted(rows.elements())[:40]:
        print(f"factory: (1)   {label} TS{code} {file}:{line}:{column} {message.replace(chr(10), ' / ')[:300]}")
fresh = sum(added.values())
print(f"factory: verdict (1) diagnostics {'PASS' if not fresh else 'FAIL'}: {fresh} new in after, {sum(removed.values())} gone")
PYTHON
fi

# (2) The latent census as the meter runs it (first error, scoped tail, no output), before and after, in parallel.
if [ "$censusBuildExit" != 0 ]; then
	say "factory: verdict (2) meter FAIL: the latent census didn't build"
	tail -n 30 "$out/census-overlay.log" "$out/census-gofmt.log" "$out/census-build.log" 2> /dev/null
else
	mkdir -p "$work/census-before" "$work/census-after"
	(LATENT_TIMES="$work/census-before/latent-times.jsonl" LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TAIL=1 \
		step 2-census-before "$work/latent-census" "$before/src/compiler" "$work/census-before/latent.jsonl") &
	censusBefore=$!
	(LATENT_TIMES="$work/census-after/latent-times.jsonl" LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TAIL=1 \
		step 2-census-after "$work/latent-census" "$after/src/compiler" "$work/census-after/latent.jsonl") &
	censusAfter=$!
	wait "$censusBefore"
	censusBeforeExit=$?
	wait "$censusAfter"
	censusAfterExit=$?
	gzip -c "$work/census-before/latent.jsonl" > "$out/2-before.jsonl.gz" 2> /dev/null
	gzip -c "$work/census-after/latent.jsonl" > "$out/2-after.jsonl.gz" 2> /dev/null
	python3 -I - "$root/stage3/meter" "$before" "$work/census-before" "$after" "$work/census-after" "$out" "$censusBeforeExit" "$censusAfterExit" << 'PYTHON' | tee -a "$out/verdicts.log"
import json, sys
from collections import Counter
from pathlib import Path
meter = Path(sys.argv[1])
sys.path.insert(0, str(meter))
import report  # The meter's own reader: its coverage checks, its unique-site count and its site_where.
beforeTree, beforeRun, afterTree, afterRun = map(Path, sys.argv[2:6])
out, codes = Path(sys.argv[6]), sys.argv[7:9]
# The family: #v879py5's 26, a function taking NodeArray<JSDocComment> seen as one taking a string too, and any like them.
def watched(key):
    return 'JSDocComment' in key
def sites(run):
    # Each site as lowering_summary counts it: every finding of every file record and of the relation report trailer,
    # a tail observation left out, unique by (kind, where, reason, text); per_reason counts NotYet and Refused only.
    # Keyed by per_reason's key, each site's where read from src/ on with report.site_where.
    records = [json.loads(line) for line in (run / 'latent.jsonl').read_text().splitlines()]
    unique = set()
    for record in records[1:]:
        for finding in record.get('findings') or []:
            if finding.get('phase') != 'tail' and finding['kind'] in ('NotYet', 'Refused'):
                unique.add((finding['kind'], finding['where'], finding['reason'], finding['text']))
    byReason = {}
    for kind, where, reason, text in unique:
        byReason.setdefault(kind + ': ' + reason, []).append((report.site_where(where), text))
    return {key: sorted(found) for key, found in byReason.items()}
def measure(tree, run):
    summary = report.latent_summary(tree, run)
    reasons = Counter(summary['per_reason'])
    found = sites(run)
    # The sites read here are the ones the meter counts, or this part says so rather than print a different set.
    drift = sorted(key for key in reasons.keys() | found.keys() if reasons.get(key, 0) != len(found.get(key, [])))
    result = {'totals': summary['totals'], 'latent_total': summary['totals']['NotYet'] + summary['totals']['Refused'],
              'source_files': summary['source_files'],
              'family': {key: count for key, count in reasons.items() if watched(key)}, 'site_drift': drift}
    return result, reasons, found
try:
    before, beforeReasons, beforeSites = measure(beforeTree, beforeRun)
    after, afterReasons, afterSites = measure(afterTree, afterRun)
except Exception as error:
    print(f"factory: verdict (2) meter FAIL: a census didn't read whole (exits {codes}): {error}")
    sys.exit(0)
for name, r in (('before', before), ('after', after)):
    print(f"factory: (2) {name}: {r['source_files']} files, latent total {r['latent_total']} {r['totals']}")
    print(f"factory: (2) {name}: the JSDocComment family {sum(r['family'].values())} sites over {len(r['family'])} reasons")
    for key, count in sorted(r['family'].items(), key=lambda item: (-item[1], item[0])):
        print(f"factory: (2)   {name} {count} {key}")
    if r['site_drift']:
        print(f"factory: (2) {name}: the sites read here disagree with per_reason for {len(r['site_drift'])} reasons, "
              f"first {r['site_drift'][:3]}")
delta = Counter(afterReasons)
delta.subtract(beforeReasons)
grown = sorted(((count, key) for key, count in delta.items() if count > 0), key=lambda item: (-item[0], item[1]))
shrunk = sorted(((count, key) for key, count in delta.items() if count < 0), key=lambda item: (item[0], item[1]))
print(f"factory: (2) reasons that appear or grow in after: {len(grown)}")
for count, key in grown:
    print(f"factory: (2)   +{count} {key} ({beforeReasons.get(key, 0)} to {afterReasons.get(key, 0)})")
print(f"factory: (2) reasons that shrink or go in after: {len(shrunk)}, {sum(c for c, _ in shrunk)} sites")
for count, key in shrunk:
    print(f"factory: (2)   {count} {key} ({beforeReasons.get(key, 0)} to {afterReasons.get(key, 0)})")
# Every site of every watched reason, in either tree, and of every reason that grew: where and text, before then after.
# The two trees have the same lines (the sources check), so a site at the same where and text in both is kept; one only
# before is gone, one only after is new.
def flat(text):
    line = ' '.join(text.split())
    return line if len(line) <= 1500 else f"{line[:1500]} ... ({len(line)} characters)"
shown = sorted({key for key in beforeReasons.keys() | afterReasons.keys() if watched(key)} | {key for _, key in grown} |
               {key for _, key in shrunk})
print(f"factory: (2) sites of {len(shown)} reasons, the JSDocComment family and every reason that grew or shrank:")
for key in shown:
    old, new = beforeSites.get(key, []), afterSites.get(key, [])
    print(f"factory: (2) reason {key}: before {len(old)}, after {len(new)}")
    for label, mine, theirs, other in (('before', old, new, 'gone'), ('after', new, old, 'new')):
        for where, text in mine:
            state = 'kept' if (where, text) in theirs else other
            print(f"factory: (2)   {label} {state} {where} {flat(text)}")
family = [sum(r['family'].values()) for r in (before, after)]
drifted = before['site_drift'] or after['site_drift']
# Any reason outside the family that moves is something the widened parameters moved beyond the 26: T's inference at the
# getters, or anything else. None may.
others = sorted(key for _, key in grown + shrunk if not watched(key))
ok = family[0] == 26 and family[1] == 0 and not grown and not others and not drifted
print(f"factory: verdict (2) meter {'PASS' if ok else 'FAIL'}: latent total {before['latent_total']} to {after['latent_total']}, "
      f"the JSDocComment family {family[0]} to {family[1]} (the adaptation dropped brings back {family[0]}), "
      f"{len(grown)} reasons grew, {len(others)} reasons outside the family moved"
      f"{', and the sites read disagree with the meter' if drifted else ''}")
(out / '2-summary.json').write_text(json.dumps({'before': before, 'after': after,
    'reason_delta': {key: count for key, count in delta.items() if count},
    'sites': {key: {'before': beforeSites.get(key, []), 'after': afterSites.get(key, [])} for key in shown}}, indent=2) + '\n')
PYTHON
fi

# (3) The lane, whole, on the after tree. The gate runs it as shards.json's units, each shard.py against an adapted tree
# product built by hereby with stage3/typescript's install linked in as node_modules (stage3/lane/adaptedtree); the
# older run.sh does its own apply and runs npm ci from the registry. Neither path may be faked.
install="$root/stage3/typescript/node_modules"
installReason=""
if [ ! -d "$install" ]; then
	installReason="stage3/typescript/node_modules isn't laid"
elif [ ! -f "$install/hereby/bin/hereby.js" ]; then
	installReason="stage3/typescript/node_modules holds no hereby"
else
	# apply.py's own reading of an install against its lockfile. The adapted tree product doesn't make it, so a
	# mismatch is printed and the lane still runs; the control on the before tree tells a box's red from the slots'.
	cat > "$out/3-install-check.py" << 'PYTHON'
import sys
from pathlib import Path
stage = Path(sys.argv[1])
sys.path.insert(0, str(stage))
import apply
apply.installed_packages(stage / 'typescript')
print('stage3/typescript/node_modules is an install of its lockfile')
PYTHON
	python3 -I "$out/3-install-check.py" "$root/stage3" > "$out/3-install-check.log" 2>&1
	echo "factory: (3) install check exited $?: $(tail -n 1 "$out/3-install-check.log")"
fi

# buildTree <tree> <label>: hereby local and tests, as the adapted tree product builds, then the link is removed.
buildTree() {
	local tree=$1 label=$2 real code
	real=$(cd "$install" && pwd -P)
	ln -s "$real" "$tree/node_modules" || return 1
	mkdir -p "$work/home-$label"
	HOME="$work/home-$label" XDG_CONFIG_HOME="$work/home-$label" NODE_DISABLE_COMPILE_CACHE=1 \
		bash -c 'cd "$1" && node "$2/hereby/bin/hereby.js" local && node "$2/hereby/bin/hereby.js" tests' _ "$tree" "$real" \
		> "$out/3-build-$label.log" 2>&1
	code=$?
	unlink "$tree/node_modules"
	echo "factory: 3-build-$label exited $code ($out/3-build-$label.log)"
	[ "$code" = 0 ] && [ -f "$tree/built/local/run.js" ]
}

# The units' driver: every unit, or the named ones, against one tree, its reports kept in $out.
cat > "$out/lane.py" << 'PYTHON'
import json, os, signal, subprocess, sys, time, shutil
from collections import Counter
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path
root, tree, results, evidence = map(Path, sys.argv[1:5])
concurrency, seconds, label = int(sys.argv[5]), float(sys.argv[6]), sys.argv[7]
only = sys.argv[8].split(',') if len(sys.argv) > 8 and sys.argv[8] else None
units = json.loads((root / 'stage3/lane/shards.json').read_text())['units']
plan = json.loads((root / 'stage3/lane/shard-plan.json').read_text())
if only:
    units = [unit for unit in units if unit['name'] in only]
results.mkdir(parents=True, exist_ok=False)
evidence.mkdir(parents=True, exist_ok=True)
def run(unit):
    env = dict(os.environ, STAGE3_ADAPTED_TREE=str(tree), STAGE3_RESULTS=str(results), STAGE3_SHARD_SECONDS=str(seconds),
               NODE_DISABLE_COMPILE_CACHE='1', PYTHONDONTWRITEBYTECODE='1')
    started = time.monotonic()
    with (evidence / (unit['name'] + '.log')).open('w') as stream:
        child = subprocess.Popen(unit['command'], cwd=root, env=env, stdout=stream, stderr=subprocess.STDOUT,
                                 start_new_session=True)
        try:
            code = child.wait(timeout=seconds + 120)
        except subprocess.TimeoutExpired:
            os.killpg(child.pid, signal.SIGKILL)
            child.wait()
            code = 124
    path = results / unit['name'] / 'report.json'
    report = json.loads(path.read_text()) if path.exists() else {'status': 'missing', 'errors': ['no report.json']}
    if path.exists():
        shutil.copyfile(path, evidence / (unit['name'] + '.json'))
    row = dict(name=unit['name'], exit=code, seconds=round(time.monotonic() - started, 1), status=report.get('status'),
               counts=report.get('counts') or {}, expect=unit['expect'], errors=report.get('errors') or [])
    with (evidence / 'ledger.jsonl').open('a') as stream:
        stream.write(json.dumps(row) + '\n')
    return row
with ThreadPoolExecutor(max_workers=concurrency) as pool:
    rows = list(pool.map(run, units))
failing = [row for row in rows if row['exit'] != 0 or row['status'] != 'pass']
total = Counter()
for row in rows:
    total.update(row['counts'])
counts = {key: total.get(key, 0) for key in ('pass', 'fail', 'skip')}
(evidence / 'failing.txt').write_text(','.join(row['name'] for row in failing))
if only is None:
    ok = not failing and counts == plan['counts']
    print(f"factory: verdict (3) lane {'PASS' if ok else 'FAIL'}: {len(rows)} units on the {label} tree, {len(failing)} failing, "
          f"counts {counts} where the plan holds {plan['counts']}")
else:
    print(f"factory: (3) control: {len(rows)} units rerun on the {label} tree, {len(failing)} failing there")
for row in failing:
    errors = ' | '.join(error.splitlines()[0][:300] for error in row['errors'][:4] if error)
    print(f"factory: (3)   {label} {row['name']} exit {row['exit']} {row['status']}, counts {row['counts']} "
          f"where the manifest expects {row['expect']}: {errors}")
if only is not None:
    for name in sorted(set(only) - {row['name'] for row in failing}):
        print(f"factory: (3)   {label} {name} passes")
PYTHON

if [ -z "$installReason" ]; then
	# Half the threads, from one to eight: each unit runs one upstream worker.
	concurrency=$(( $(nproc) / 2 ))
	[ "$concurrency" -ge 1 ] || concurrency=1
	[ "$concurrency" -le 8 ] || concurrency=8
	concurrency=${FACTORY_LANE_CONCURRENCY:-$concurrency}
	# shard.py's own limit is 30 s, the gate's timing contract; this run measures verdicts, so units get more room.
	seconds=${FACTORY_SHARD_SECONDS:-120}
	echo "factory: (3) the lane's product path: shards.json's units, $concurrency at a time, $seconds s each"
	# Both trees are built, one after the other, so (4) can compare what hereby emits; the before tree is also (3)'s control.
	buildTree "$after" after
	afterBuilt=$?
	buildTree "$before" before
	beforeBuilt=$?
	if [ "$afterBuilt" = 0 ] && [ "$beforeBuilt" = 0 ]; then
		python3 -I - "$before/built/local" "$after/built/local" << 'PYTHON' | tee -a "$out/verdicts.log"
import hashlib, sys
from pathlib import Path
before, after = Path(sys.argv[1]), Path(sys.argv[2])
def manifest(root):
    return {p.relative_to(root).as_posix(): hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(root.rglob('*')) if p.is_file()}
old, new = manifest(before), manifest(after)
differing = sorted(name for name in old.keys() | new.keys() if old.get(name) != new.get(name))
named = [name for name in ('typescript.js', 'typescript.d.ts', 'tsc.js', '_tsc.js') if name in new]
print(f"factory: (4) built/local holds {len(new)} files after and {len(old)} before; {named} among them")
for name in differing[:40]:
    print(f"factory: (4)   differs {name}: before {old.get(name, 'absent')[:12]}, after {new.get(name, 'absent')[:12]}")
ok = not differing and 'typescript.js' in new and 'typescript.d.ts' in new
print(f"factory: verdict (4) emit {'PASS' if ok else 'FAIL'}: {len(differing)} of {len(new)} built files differ")
PYTHON
	else
		say "factory: verdict (4) emit FAIL: a tree didn't build with hereby (after $afterBuilt, before $beforeBuilt)"
	fi
	if [ "$afterBuilt" != 0 ]; then
		say "factory: verdict (3) lane FAIL: the after tree didn't build with hereby; see $out/3-build-after.log"
		tail -n 30 "$out/3-build-after.log"
	else
		python3 -I "$out/lane.py" "$root" "$after" "$work/lane-after" "$out/3-lane-after" "$concurrency" "$seconds" after \
			> "$out/3-lane-after.log" 2>&1
		echo "factory: 3-lane-after exited $? ($out/3-lane-after.log)"
		grep '^factory: ' "$out/3-lane-after.log" | tee -a "$out/verdicts.log"
		grep -q '^factory: verdict (3) lane ' "$out/3-lane-after.log" ||
			say "factory: verdict (3) lane FAIL: the units' driver printed no verdict; see $out/3-lane-after.log"
		failing=$(cat "$out/3-lane-after/failing.txt" 2> /dev/null)
		# A control: the failing units again on the before tree, so a red reads as the adaptation's or the box's.
		if [ -n "$failing" ]; then
			if [ "$beforeBuilt" = 0 ]; then
				python3 -I "$out/lane.py" "$root" "$before" "$work/lane-before" "$out/3-lane-before" "$concurrency" "$seconds" before "$failing" \
					> "$out/3-lane-before.log" 2>&1
				grep '^factory: ' "$out/3-lane-before.log" | tee -a "$out/verdicts.log"
			else
				echo "factory: (3) control: the before tree didn't build; see $out/3-build-before.log"
			fi
		fi
	fi
elif [ "$hasNpm" = 1 ]; then
	platform=$(node -p 'process.platform + "/" + process.arch + "/" + process.version')
	if ! python3 -I -c 'import json, sys; sys.exit(0 if sys.argv[2] in json.load(open(sys.argv[1]))["platforms"] else 1)' "$root/stage3/lane/expected.json" "$platform"; then
		say "factory: verdict (3) lane SKIP: $installReason, and run.sh's expected.json holds no counts for $platform"
		say "factory: verdict (4) emit SKIP: $installReason"
	elif ! npm ping > "$out/3-npm-ping.log" 2>&1; then
		say "factory: verdict (3) lane SKIP: $installReason, and run.sh's npm ci can't reach the registry (npm ping failed)"
		say "factory: verdict (4) emit SKIP: $installReason"
	else
		# run.sh applies this tree's own stage3 again; the store holds after's product under the same key, so it's restored.
		(cd "$root" && STAGE3_CACHE="$work/cache-after" step 3-lane-run bash stage3/lane/run.sh "$work/lane-run")
		for file in verdict.txt verdict.json report.json execution.json failures.md; do
			cp "$work/lane-run/$file" "$out/3-lane-run-$file" 2> /dev/null
		done
		verdict=$(cat "$work/lane-run/verdict.txt" 2> /dev/null)
		case "$verdict" in
			PASS*) say "factory: verdict (3) lane PASS: run.sh: $verdict" ;;
			*) say "factory: verdict (3) lane FAIL: run.sh: ${verdict:-no verdict.txt}" ;;
		esac
		say "factory: verdict (4) emit SKIP: $installReason; run.sh builds one tree"
	fi
else
	say "factory: verdict (3) lane SKIP: $installReason, and there is no npm for run.sh's npm ci; lay stage3/typescript/node_modules (the pinned devDependencies) to run the product path"
	say "factory: verdict (4) emit SKIP: $installReason"
fi

# (6) The adapted tree product, started before phase 1.
wait "$adaptedTree"
adaptedTreeExit=$?
if [ "$adaptedTreeExit" = 0 ]; then
	say "factory: verdict (6) adapted tree PASS: TestProduct_Stage3AdaptedTree in ./stage3/lane/adaptedtree"
else
	say "factory: verdict (6) adapted tree FAIL: go test exited $adaptedTreeExit; its last lines follow"
	tail -n 60 "$out/6-adapted-tree.log"
fi

echo "factory: summary, $SECONDS s:"
grep '^factory: verdict ' "$out/verdicts.log"
grep -q '^factory: verdict .* FAIL:' "$out/verdicts.log" && exit 1
grep -q '^factory: verdict .* SKIP:' "$out/verdicts.log" && exit 3
exit 0
