#!/usr/bin/env bash
# three-slots.sh: #m2wxyaz's measurement of DiagnosticRelatedInformation's file, start and length made readonly
# together and unaudited, by the scratch adaptation stage3/adapt/71-scratch-diagnostic-slots. Run from an adamic tree's
# root on a house box, after a lay line that lays cohere, cohere/TypeScript and stage3/api's node_modules, and after
# sourcing ~/adamic-tools/env.sh (go, node 24, python3, git; npm is optional).
#
#   before: stage3/apply.sh over a copy of this tree's stage3 inputs with 71-scratch-diagnostic-slots left out
#   after:  this tree's own stage3/apply.sh, which runs 71 after 70-readonly-views and before 71-writable-views
#
#   sources: the two src trees differ only in src/compiler/types.ts, by the three `readonly ` insertions
#   (1) TS2540: one ts.createProgram per tree, 70's options, stock TypeScript 6.0.3; PASS is no new TS2540
#   (2) the meter: the latent census on both src/compiler trees, read by report.py's latent_summary; PASS is both
#       Diagnostic families at 0 after and no reason grown
#   (3) the lane, whole, on the after tree: every unit of stage3/lane/shards.json run by shard.py against the after tree
#       built by hereby with stage3/typescript's install, as the gate's product path runs it; or, on a box with npm and
#       no laid install, stage3/lane/run.sh; SKIP with the reason when neither can run here
#
# Verdict lines begin "three-slots: verdict". Logs, JSON and the summary go to /tmp/three-slots-<stamp>; the heavy
# trees go beside the tree and are removed at the end unless THREE_SLOTS_KEEP=1. Exit 0 when every part passes, 1 when
# any fails, 3 when none fails and one is skipped; 64 to 66 are preparation failures.
set -u

root=$PWD
[ -f "$root/stage3/apply.sh" ] && [ -f "$root/stage3/adapt/71-scratch-diagnostic-slots/adapt.cjs" ] || {
	echo "three-slots: $root is not an adamic tree's root carrying 71-scratch-diagnostic-slots, so nothing ran"
	exit 64
}
[ "$(grep -c "declaration: 'DiagnosticRelatedInformation'" "$root/stage3/adapt/70-readonly-views/adapt.cjs")" = 0 ] || {
	echo "three-slots: adaptation 70 still carries a DiagnosticRelatedInformation target, so before and after would both audit it; nothing ran"
	exit 64
}
stamp=$(date -u +%Y%m%dT%H%M%SZ)
out=/tmp/three-slots-$stamp
work="${THREE_SLOTS_WORK:-$root.three-slots}"
mkdir -p "$out" || exit 64
mkdir "$work" || { echo "three-slots: $work is there already or can't be made, so nothing ran"; exit 64; }
jobEnd() {
	local code=$?
	trap - EXIT
	if [ "${THREE_SLOTS_KEEP:-0}" != 1 ]; then
		case "$work" in /*/*.three-slots) chmod -R u+w -- "$work" 2> /dev/null; rm -rf -- "$work" ;; esac
	fi
	echo "three-slots: evidence in $out"
	exit "$code"
}
trap jobEnd EXIT
trap 'exit 143' TERM
trap 'exit 129' HUP
trap 'exit 130' INT
mkdir "$work/tmp" || exit 64
export TMPDIR="$work/tmp"

if [ -f "$HOME/adamic-tools/env.sh" ]; then
	. "$HOME/adamic-tools/env.sh" || { echo "three-slots: $HOME/adamic-tools/env.sh didn't source, so nothing ran"; exit 65; }
fi
missing=""
for tool in go gofmt node python3 git; do
	command -v "$tool" > /dev/null 2>&1 || missing="$missing $tool"
done
[ -z "$missing" ] || { echo "three-slots: missing from the toolchain:$missing"; exit 65; }
hasNpm=0; command -v npm > /dev/null 2>&1 && hasNpm=1
node --version | grep -q '^v24\.' || { echo "three-slots: node is $(node --version), adaptations need Node 24"; exit 65; }
[ -n "$(ls -A "$root/cohere" 2> /dev/null)" ] || { echo "three-slots: cohere isn't laid in $root"; exit 66; }
echo "three-slots: $(uname -n), $(nproc) threads, $(git -C "$root" rev-parse HEAD 2> /dev/null || echo 'no checkout'), node $(node --version), $(go version), npm $([ "$hasNpm" = 1 ] && npm --version || echo absent)"

# step <name> <command...>: the command's whole output in $out/<name>.log, and its exit.
step() {
	local name=$1
	shift
	"$@" > "$out/$name.log" 2>&1
	local code=$?
	echo "three-slots: $name exited $code ($out/$name.log)"
	return "$code"
}

# say <line>: a line printed and kept in $out/verdicts.log for the summary and the exit.
say() { echo "$*" | tee -a "$out/verdicts.log"; }

# The pinned upstream checkout apply reads: the caller's, the laid cohere/TypeScript when it is the pin, else a clone.
pin() { python3 -I -c 'import json, sys; print(json.load(open(sys.argv[1]))[sys.argv[2]])' "$root/stage3/source.json" "$1"; }
pinCommit=$(pin commit)
if [ -n "${ADAMIC_TYPESCRIPT_SOURCE:-}" ] && [ "$(git -C "$ADAMIC_TYPESCRIPT_SOURCE" rev-parse HEAD 2> /dev/null)" = "$pinCommit" ]; then
	echo "three-slots: upstream is the caller's ADAMIC_TYPESCRIPT_SOURCE"
elif [ "$(git -C "$root/cohere/TypeScript" rev-parse HEAD 2> /dev/null)" = "$pinCommit" ]; then
	export ADAMIC_TYPESCRIPT_SOURCE="$root/cohere/TypeScript"
	echo "three-slots: upstream is the laid cohere/TypeScript, at the pin"
else
	step upstream-clone git clone --depth 1 --branch "$(pin tag)" "$(pin repository)" "$work/upstream" || exit 66
	export ADAMIC_TYPESCRIPT_SOURCE="$work/upstream"
fi
[ "$(git -C "$ADAMIC_TYPESCRIPT_SOURCE" rev-parse HEAD)" = "$pinCommit" ] || { echo "three-slots: upstream isn't at $pinCommit"; exit 66; }

# stage3/api's lockfile-pinned typescript@6.0.3, which apply puts on NODE_PATH for every adaptation.
modules="$root/stage3/api/node_modules"
[ "$(node -p 'require(process.argv[1] + "/typescript/package.json").version' "$modules" 2> /dev/null)" = 6.0.3 ] || { echo "three-slots: stage3/api's typescript isn't laid at 6.0.3"; exit 66; }

# Before's stage3: this tree's apply inputs with 71-scratch-diagnostic-slots left out, and nothing else changed.
mutant="$work/mutant/stage3"
mkdir -p "$mutant/adapt" || exit 64
for input in apply.sh apply.py caches.py source.json api; do
	cp -a "$root/stage3/$input" "$mutant/" || exit 64
done
for entry in "$root/stage3/adapt"/* "$root/stage3/adapt"/.[!.]*; do
	[ -e "$entry" ] || [ -L "$entry" ] || continue
	[ "$(basename "$entry")" = 71-scratch-diagnostic-slots ] || cp -a "$entry" "$mutant/adapt/" || exit 64
done
diff -rq "$root/stage3/adapt" "$mutant/adapt" > "$out/before-inputs.diff" 2>&1
echo "three-slots: before's adapt differs from this tree's by: $(tr '\n' ' ' < "$out/before-inputs.diff")"
echo "three-slots: product keys: after $(cd "$root" && python3 stage3/apply.py --key), before $(python3 "$mutant/apply.py" --key)"

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
echo "three-slots: apply before exited $applyBeforeExit, apply after exited $applyAfterExit, census build exited $censusBuildExit"
[ "$applyBeforeExit" = 0 ] || { echo "three-slots: before's apply failed, so there is no before tree; its last lines:"; tail -n 30 "$out/apply-before.log"; exit 66; }
if [ "$applyAfterExit" != 0 ]; then
	# apply.py names the adaptation it was running in the failed command.
	failed=$(grep -o 'stage3/adapt/[0-9][0-9]-[^/]*/adapt.cjs' "$out/apply-after.log" | tail -n 1)
	say "three-slots: verdict apply FAIL: the after tree's apply exited $applyAfterExit${failed:+ in $failed}, so there is no after tree"
	tail -n 30 "$out/apply-after.log"
	say "three-slots: verdict (1) TS2540 SKIP: no after tree"
	say "three-slots: verdict (2) meter SKIP: no after tree"
	say "three-slots: verdict (3) lane SKIP: no after tree"
	exit 1
fi
say "three-slots: verdict apply PASS: both applies exited 0; after's patch-set row: $(grep -F '71-scratch-diagnostic-slots' "$after/patch-set.md")"

# The sources: src before and after differ only in types.ts, by `readonly ` before the three members.
python3 -I - "$before" "$after" "$out/sources.json" << 'PYTHON' | tee -a "$out/verdicts.log"
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
print(f"three-slots: verdict sources {'PASS' if ok else 'FAIL'}: {len(new)} src files, differing {differing}, "
      f"types.ts lines changed {[(line, b.strip()) for line, _, b in changed]}")
PYTHON

# (1) TS2540: the program 70's audit builds, on each tree, and every syntactic, global and semantic diagnostic.
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
            message: ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n') };
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
	say "three-slots: verdict (1) TS2540 FAIL: the program didn't build (exits $diagnosticsBeforeExit before, $diagnosticsAfterExit after)"
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
def ts2540(rows):
    return sum(row['code'] == 2540 for row in rows)
print(f"three-slots: (1) diagnostics: before {len(before['diagnostics'])} over {before['roots']} roots, "
      f"after {len(after['diagnostics'])} over {after['roots']} roots")
print(f"three-slots: (1) TS2540: before {ts2540(before['diagnostics'])}, after {ts2540(after['diagnostics'])}")
print(f"three-slots: (1) diagnostics new in after by code {dict(sorted(Counter(k[3] for k in added.elements()).items()))}, "
      f"gone from after by code {dict(sorted(Counter(k[3] for k in removed.elements()).items()))}")
fresh = sorted(k for k in added.elements() if k[3] == 2540)
for file, line, column, code, message in fresh:
    print(f"three-slots: (1)   new TS2540 {file}:{line}:{column} {message.replace(chr(10), ' / ')}")
for file, line, column, code, message in sorted(k for k in added.elements() if k[3] != 2540)[:40]:
    print(f"three-slots: (1)   new TS{code} {file}:{line}:{column} {message.replace(chr(10), ' / ')[:300]}")
print(f"three-slots: verdict (1) TS2540 {'PASS' if not fresh else 'FAIL'}: {len(fresh)} TS2540 new in after")
PYTHON
fi

# (2) The latent census as the meter runs it (first error, scoped tail, no output), before and after, in parallel.
if [ "$censusBuildExit" != 0 ]; then
	say "three-slots: verdict (2) meter FAIL: the latent census didn't build"
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
import json, re, sys
from collections import Counter
from pathlib import Path
meter = Path(sys.argv[1])
sys.path.insert(0, str(meter))
import report  # The meter's own reader: its coverage checks and its unique-site count.
beforeTree, beforeRun, afterTree, afterRun = map(Path, sys.argv[2:6])
out, codes = Path(sys.argv[6]), sys.argv[7:9]
# A family's reason names one of the four diagnostic views before ", which can write".
views = re.compile(r'\b(?:Diagnostic|DiagnosticRelatedInformation|DiagnosticWithLocation|DiagnosticWithDetachedLocation)\b')
families = {
    'SourceFile': ', which can write SourceFile | undefined where SourceFile is read',
    'number': ', which can write number | undefined where number is read',
}
# w6-job.sh's family, kept so its 192 reads the same here: the seen-as type names Diagnostic or DiagnosticRelatedInformation.
w6 = re.compile(r'seen as (.*), which can write SourceFile \| undefined where SourceFile is read$')
def family(reasons, ending):
    return {key: count for key, count in reasons.items()
            if key.endswith(ending) and views.search(key.split(': ', 1)[1][:-len(ending)])}
def measure(tree, run):
    summary = report.latent_summary(tree, run)
    reasons = Counter(summary['per_reason'])
    result = {'totals': summary['totals'], 'latent_total': summary['totals']['NotYet'] + summary['totals']['Refused'],
              'source_files': summary['source_files'],
              'families': {name: family(reasons, ending) for name, ending in families.items()},
              'sourcefile_any': sum(count for key, count in reasons.items() if key.endswith(families['SourceFile'])),
              'w6_family': sum(count for key, count in reasons.items()
                               if (m := w6.search(key)) and re.search(r'\bDiagnostic(RelatedInformation)?\b', m.group(1)))}
    return result, reasons
try:
    before, beforeReasons = measure(beforeTree, beforeRun)
    after, afterReasons = measure(afterTree, afterRun)
except Exception as error:
    print(f"three-slots: verdict (2) meter FAIL: a census didn't read whole (exits {codes}): {error}")
    sys.exit(0)
for name, r in (('before', before), ('after', after)):
    print(f"three-slots: (2) {name}: {r['source_files']} files, latent total {r['latent_total']} {r['totals']}")
    print(f"three-slots: (2) {name}: SourceFile family {sum(r['families']['SourceFile'].values())} "
          f"(w6's definition {r['w6_family']}, any reason ending so {r['sourcefile_any']}), "
          f"number family {sum(r['families']['number'].values())}")
    for label, members in r['families'].items():
        for key, count in sorted(members.items(), key=lambda item: (-item[1], item[0])):
            print(f"three-slots: (2)   {name} {label} {count} {key}")
delta = Counter(afterReasons)
delta.subtract(beforeReasons)
grown = sorted(((count, key) for key, count in delta.items() if count > 0), key=lambda item: (-item[0], item[1]))
shrunk = sorted(((count, key) for key, count in delta.items() if count < 0), key=lambda item: (item[0], item[1]))
print(f"three-slots: (2) reasons that appear or grow in after: {len(grown)}")
for count, key in grown:
    print(f"three-slots: (2)   +{count} {key} ({beforeReasons.get(key, 0)} to {afterReasons.get(key, 0)})")
print(f"three-slots: (2) reasons that shrink or go in after: {len(shrunk)}, {sum(c for c, _ in shrunk)} sites")
for count, key in shrunk:
    print(f"three-slots: (2)   {count} {key} ({beforeReasons.get(key, 0)} to {afterReasons.get(key, 0)})")
sourcefile = [sum(r['families']['SourceFile'].values()) for r in (before, after)]
number = [sum(r['families']['number'].values()) for r in (before, after)]
ok = sourcefile[1] == 0 and number[1] == 0 and not grown
print(f"three-slots: verdict (2) meter {'PASS' if ok else 'FAIL'}: latent total {before['latent_total']} to {after['latent_total']}, "
      f"SourceFile family {sourcefile[0]} to {sourcefile[1]}, number family {number[0]} to {number[1]}, {len(grown)} reasons grew")
(out / '2-summary.json').write_text(json.dumps({'before': before, 'after': after,
    'reason_delta': {key: count for key, count in delta.items() if count}}, indent=2) + '\n')
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
	echo "three-slots: (3) install check exited $?: $(tail -n 1 "$out/3-install-check.log")"
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
	echo "three-slots: 3-build-$label exited $code ($out/3-build-$label.log)"
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
    print(f"three-slots: verdict (3) lane {'PASS' if ok else 'FAIL'}: {len(rows)} units on the {label} tree, {len(failing)} failing, "
          f"counts {counts} where the plan holds {plan['counts']}")
else:
    print(f"three-slots: (3) control: {len(rows)} units rerun on the {label} tree, {len(failing)} failing there")
for row in failing:
    errors = ' | '.join(error.splitlines()[0][:300] for error in row['errors'][:4] if error)
    print(f"three-slots: (3)   {label} {row['name']} exit {row['exit']} {row['status']}, counts {row['counts']} "
          f"where the manifest expects {row['expect']}: {errors}")
if only is not None:
    for name in sorted(set(only) - {row['name'] for row in failing}):
        print(f"three-slots: (3)   {label} {name} passes")
PYTHON

if [ -z "$installReason" ]; then
	# Half the threads, from one to eight: each unit runs one upstream worker.
	concurrency=$(( $(nproc) / 2 ))
	[ "$concurrency" -ge 1 ] || concurrency=1
	[ "$concurrency" -le 8 ] || concurrency=8
	concurrency=${THREE_SLOTS_LANE_CONCURRENCY:-$concurrency}
	# shard.py's own limit is 30 s, the gate's timing contract; this run measures verdicts, so units get more room.
	seconds=${THREE_SLOTS_SHARD_SECONDS:-120}
	echo "three-slots: (3) the lane's product path: shards.json's units, $concurrency at a time, $seconds s each"
	if ! buildTree "$after" after; then
		say "three-slots: verdict (3) lane FAIL: the after tree didn't build with hereby; see $out/3-build-after.log"
		tail -n 30 "$out/3-build-after.log"
	else
		python3 -I "$out/lane.py" "$root" "$after" "$work/lane-after" "$out/3-lane-after" "$concurrency" "$seconds" after \
			> "$out/3-lane-after.log" 2>&1
		echo "three-slots: 3-lane-after exited $? ($out/3-lane-after.log)"
		grep '^three-slots: ' "$out/3-lane-after.log" | tee -a "$out/verdicts.log"
		grep -q '^three-slots: verdict (3) lane ' "$out/3-lane-after.log" ||
			say "three-slots: verdict (3) lane FAIL: the units' driver printed no verdict; see $out/3-lane-after.log"
		failing=$(cat "$out/3-lane-after/failing.txt" 2> /dev/null)
		# A control: the failing units again on the before tree, so a red reads as the three slots' or the box's.
		if [ -n "$failing" ]; then
			if buildTree "$before" before; then
				python3 -I "$out/lane.py" "$root" "$before" "$work/lane-before" "$out/3-lane-before" "$concurrency" "$seconds" before "$failing" \
					> "$out/3-lane-before.log" 2>&1
				grep '^three-slots: ' "$out/3-lane-before.log" | tee -a "$out/verdicts.log"
			else
				echo "three-slots: (3) control: the before tree didn't build; see $out/3-build-before.log"
			fi
		fi
	fi
elif [ "$hasNpm" = 1 ]; then
	platform=$(node -p 'process.platform + "/" + process.arch + "/" + process.version')
	if ! python3 -I -c 'import json, sys; sys.exit(0 if sys.argv[2] in json.load(open(sys.argv[1]))["platforms"] else 1)' "$root/stage3/lane/expected.json" "$platform"; then
		say "three-slots: verdict (3) lane SKIP: $installReason, and run.sh's expected.json holds no counts for $platform"
	elif ! npm ping > "$out/3-npm-ping.log" 2>&1; then
		say "three-slots: verdict (3) lane SKIP: $installReason, and run.sh's npm ci can't reach the registry (npm ping failed)"
	else
		# run.sh applies this tree's own stage3 again; the store holds after's product under the same key, so it's restored.
		(cd "$root" && STAGE3_CACHE="$work/cache-after" step 3-lane-run bash stage3/lane/run.sh "$work/lane-run")
		for file in verdict.txt verdict.json report.json execution.json failures.md; do
			cp "$work/lane-run/$file" "$out/3-lane-run-$file" 2> /dev/null
		done
		verdict=$(cat "$work/lane-run/verdict.txt" 2> /dev/null)
		case "$verdict" in
			PASS*) say "three-slots: verdict (3) lane PASS: run.sh: $verdict" ;;
			*) say "three-slots: verdict (3) lane FAIL: run.sh: ${verdict:-no verdict.txt}" ;;
		esac
	fi
else
	say "three-slots: verdict (3) lane SKIP: $installReason, and there is no npm for run.sh's npm ci; lay stage3/typescript/node_modules (the pinned devDependencies) to run the product path"
fi

echo "three-slots: summary, $SECONDS s:"
grep '^three-slots: verdict ' "$out/verdicts.log"
grep -q '^three-slots: verdict .* FAIL:' "$out/verdicts.log" && exit 1
grep -q '^three-slots: verdict .* SKIP:' "$out/verdicts.log" && exit 3
exit 0
