#!/usr/bin/env bash
# measure.sh: a scratch measurement for #38kdd4d, never to land. Adaptation 99-memory-factory-update-view changes one
# parameter in src/compiler/factory/nodeFactory.ts: createNodeFactory's `update` takes `updated: T`, not
# `updated: Mutable<T>`. This measures what it changes. Run from an adamic tree's root, on a sealed job as
#
#   source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh && sandboxed_typescript_source &&
#   sandboxed_npm stage3/api && bash scratch-upd/measure.sh
#
# Its instruments are objectflow.sh's (branch memory/scratch-77r): apply.sh from a fresh local product store, the
# latent census built by make_overlay.py and read by the meter's report.py, and adaptation 70's stock tsc program.
#
#   before: stage3/apply.sh over a copy of this tree's stage3 inputs without 99-memory-factory-update-view, main's tree
#   after:  this tree's own stage3/apply.sh, main's tree with the adaptation
#   mutant: the after tree with `updated: Mutable<T>` restored in update's one line, written here into a linked copy
#
#   (0) sources: PASS is nodeFactory.ts alone differing under src/, and there update's one line, `Mutable<T>` to `T`;
#       adapt.cjs run again over after's nodeFactory.ts prints already adapted and changes no byte, and run over
#       before's yields after's bytes exactly
#   (1) stock tsc: one ts.createProgram per tree over all upstream src, 70's options, stock TypeScript 6.0.3; PASS is
#       every diagnostic (code, file, line, message) the same before and after, each tree's root read as <tree>
#   (2) JavaScript: every src file the program reads, transpiled by stock TypeScript 6.0.3's transpileModule with 70's
#       module and target, on both trees; PASS is every output byte-identical
#   (3) the census: the latent census on both src/compiler trees, read by report.py's latent_summary; PASS is the
#       Refused sites seen as a Mutable<...> view at an `update(...)` call in nodeFactory.ts going from the count the
#       census gives (44 by #38kdd4d's W2 rows) to 0, and no site of any reason new in after
#   (4) the mutant: the census on the mutant tree; PASS is its src before's byte for byte, and the sites it holds that
#       after doesn't exactly the update sites (3) saw go, with nothing else differing from before
#
# Every line printed begins "upd:", verdict lines "upd: verdict". Logs, JSON and the summary go to
# $TMPDIR/measure-upd-<stamp>; long lists stay there, and at most about 40 sites print per section, since a job's output
# is cut near 150 KB. The heavy trees go to $TMPDIR too (UPD_WORK names another place), never beside the tree, which a
# sealed job lays read-only, and are removed at the end unless UPD_KEEP=1. Exit 0 when every part passes, 1 when any
# fails; 64 to 66 are preparation failures.
set -u

root=$PWD
adaptation=99-memory-factory-update-view
nodeFactory=src/compiler/factory/nodeFactory.ts
[ -f "$root/stage3/apply.sh" ] && [ -f "$root/stage3/adapt/$adaptation/adapt.cjs" ] || {
	echo "upd: $root is not an adamic tree's root carrying stage3/adapt/$adaptation, so nothing ran"
	exit 64
}
stamp=$(date -u +%Y%m%dT%H%M%SZ)
scratch="${TMPDIR:-/tmp}"
out="${scratch%/}/measure-upd-$stamp"
work="${UPD_WORK:-${scratch%/}/upd-$stamp.flow}"
mkdir -p "$out" || exit 64
mkdir "$work" || { echo "upd: $work is there already or can't be made, so nothing ran"; exit 64; }
jobEnd() {
	local code=$?
	trap - EXIT
	if [ "${UPD_KEEP:-0}" != 1 ]; then
		case "$work" in /*/*.flow) chmod -R u+w -- "$work" 2> /dev/null; rm -rf -- "$work" ;; esac
	fi
	echo "upd: evidence in $out"
	exit "$code"
}
trap jobEnd EXIT
trap 'exit 143' TERM
trap 'exit 129' HUP
trap 'exit 130' INT
mkdir "$work/tmp" || exit 64
export TMPDIR="$work/tmp"
# No python this script starts writes a __pycache__ into the tree: stage3/meter's report is imported from it.
export PYTHONDONTWRITEBYTECODE=1

if [ -f "$HOME/adamic-tools/env.sh" ]; then
	. "$HOME/adamic-tools/env.sh" || { echo "upd: $HOME/adamic-tools/env.sh didn't source, so nothing ran"; exit 65; }
fi
missing=""
for tool in go gofmt node python3 git diff cmp; do
	command -v "$tool" > /dev/null 2>&1 || missing="$missing $tool"
done
[ -z "$missing" ] || { echo "upd: missing from the toolchain:$missing"; exit 65; }
node --version | grep -q '^v24\.' || { echo "upd: node is $(node --version), adaptations need Node 24"; exit 65; }
[ -n "$(ls -A "$root/cohere" 2> /dev/null)" ] || { echo "upd: cohere isn't laid in $root"; exit 66; }
echo "upd: $(uname -n), $(nproc) threads, $(git -C "$root" rev-parse HEAD 2> /dev/null || echo 'no checkout'), node $(node --version), $(go version)"

# step <name> <command...>: the command's whole output in $out/<name>.log, and its exit.
step() {
	local name=$1
	shift
	"$@" > "$out/$name.log" 2>&1
	local code=$?
	echo "upd: $name exited $code ($out/$name.log)"
	return "$code"
}

# say <line>: a line printed and kept in $out/verdicts.log for the summary and the exit.
say() { echo "$*" | tee -a "$out/verdicts.log"; }

# quote <lines> <file...>: a file's last lines, each printed under the upd: prefix.
quote() {
	local lines=$1
	shift
	tail -n "$lines" "$@" 2> /dev/null | sed 's/^/upd:   | /'
}

# The pinned upstream checkout apply reads: the caller's, the laid cohere/TypeScript when it is the pin, else a clone.
pin() { python3 -I -c 'import json, sys; print(json.load(open(sys.argv[1]))[sys.argv[2]])' "$root/stage3/source.json" "$1"; }
pinCommit=$(pin commit)
if [ -n "${ADAMIC_TYPESCRIPT_SOURCE:-}" ] && [ "$(git -C "$ADAMIC_TYPESCRIPT_SOURCE" rev-parse HEAD 2> /dev/null)" = "$pinCommit" ]; then
	echo "upd: upstream is the caller's ADAMIC_TYPESCRIPT_SOURCE"
elif [ "$(git -C "$root/cohere/TypeScript" rev-parse HEAD 2> /dev/null)" = "$pinCommit" ]; then
	export ADAMIC_TYPESCRIPT_SOURCE="$root/cohere/TypeScript"
	echo "upd: upstream is the laid cohere/TypeScript, at the pin"
else
	step upstream-clone git clone --depth 1 --branch "$(pin tag)" "$(pin repository)" "$work/upstream" || exit 66
	export ADAMIC_TYPESCRIPT_SOURCE="$work/upstream"
fi
[ "$(git -C "$ADAMIC_TYPESCRIPT_SOURCE" rev-parse HEAD)" = "$pinCommit" ] || { echo "upd: upstream isn't at $pinCommit"; exit 66; }

# stage3/api's lockfile-pinned typescript@6.0.3, which apply puts on NODE_PATH for every adaptation.
modules="$root/stage3/api/node_modules"
[ "$(node -p 'require(process.argv[1] + "/typescript/package.json").version' "$modules" 2> /dev/null)" = 6.0.3 ] || { echo "upd: stage3/api's typescript isn't laid at 6.0.3"; exit 66; }

# stageInputs <directory> <left out>: a copy of this tree's apply inputs as <directory>/stage3, every adaptation but the
# one named, and nothing else changed.
stageInputs() {
	local stage=$1/stage3 leave=$2 input entry
	mkdir -p "$stage/adapt" || return 1
	for input in apply.sh apply.py caches.py source.json api; do
		cp -a "$root/stage3/$input" "$stage/" || return 1
	done
	for entry in "$root/stage3/adapt"/* "$root/stage3/adapt"/.[!.]*; do
		[ -e "$entry" ] || [ -L "$entry" ] || continue
		[ -n "$leave" ] && [ "$(basename "$entry")" = "$leave" ] && continue
		cp -a "$entry" "$stage/adapt/" || return 1
	done
}

# Before's stage3: this tree's apply inputs whole but the adaptation's directory, so before is main's tree.
beforeStage="$work/before-inputs/stage3"
stageInputs "$work/before-inputs" "$adaptation" || exit 64
diff -rq "$root/stage3/adapt" "$beforeStage/adapt" > "$out/before-inputs.diff" 2>&1
echo "upd: before's adapt differs from this tree's by: $(tr '\n' ' ' < "$out/before-inputs.diff")"
[ "$(wc -l < "$out/before-inputs.diff")" = 1 ] && grep -q "Only in $root/stage3/adapt: $adaptation\$" "$out/before-inputs.diff" ||
	{ echo "upd: before's inputs differ from this tree's by more than the adaptation's directory, so nothing ran"; exit 64; }
echo "upd: product keys: after $(cd "$root" && python3 stage3/apply.py --key), before $(python3 "$beforeStage/apply.py" --key)"

# Phase 1, in parallel: the two applies from a fresh local product store, and the census tool. A local store makes
# each apply build its tree from the pinned upstream here; nothing is fetched as a product.
before="$work/before"
after="$work/after"
mutant="$work/mutant"
export STAGE3_PRODUCT_STORE="$work/products"
mkdir -p "$STAGE3_PRODUCT_STORE"
(STAGE3_CACHE="$work/cache-before" step apply-before bash "$beforeStage/apply.sh" "$before") &
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
echo "upd: apply before exited $applyBeforeExit, apply after exited $applyAfterExit, census build exited $censusBuildExit"
[ "$applyBeforeExit" = 0 ] || { echo "upd: before's apply failed, so there is no before tree; its last lines:"; quote 30 "$out/apply-before.log"; exit 66; }
if [ "$applyAfterExit" != 0 ]; then
	failed=$(grep -o 'stage3/adapt/[0-9][0-9]-[^/]*/adapt.cjs' "$out/apply-after.log" | tail -n 1)
	say "upd: verdict apply FAIL: the after tree's apply exited $applyAfterExit${failed:+ in $failed}, so there is no after tree"
	quote 30 "$out/apply-after.log"
	for part in '(0) sources' '(1) stock tsc' '(2) javascript' '(3) census' '(4) mutant'; do
		say "upd: verdict $part SKIP: no after tree"
	done
	exit 1
fi
say "upd: verdict apply PASS: before and after applies exited 0; after's patch-set row: $(grep -F "$adaptation" "$after/patch-set.md")"
say "upd: the adaptation's own line in after's apply: $(grep -F 'update updated:' "$out/apply-after.log" | head -n 1)"

# The judgments more than one part reads, in one home: the files that differ, line maps, the meter's sites, and which
# call a site stands in.
cat > "$out/updlib.py" << 'PYTHON'
import difflib, hashlib, json, re
from collections import Counter
from pathlib import Path

NODE_FACTORY = 'src/compiler/factory/nodeFactory.ts'
UPSTREAM = '    function update<T extends Node>(updated: Mutable<T>, original: T): T {'
ADAPTED = '    function update<T extends Node>(updated: T, original: T): T {'
# #38kdd4d's W2 rows (memory/scratch-w2-scout, scratch-w2/w2-rows-85.tsv): the 44 update sites, in main's numbering.
W2_UPDATE_SITES = sorted(f'{NODE_FACTORY}:{line}:22' for line in [
    1640, 2669, 2709, 2754, 3629, 3871, 3920, 4091, 4110, 4131, 4174, 4617, 4722, 4753, 4771, 4788, 4804, 4824, 4841,
    4857, 4874, 4890, 5102, 5112, 5135, 5151, 5165, 5187, 5227, 5249, 5271, 5293, 5315, 5331, 5347, 5370, 5384, 5417,
    5432, 5447, 5456, 5524, 5563, 5577])


def differing(before, after):
    # The tree-relative names of the files under src/ whose bytes differ between two trees, or that one lacks.
    def manifest(tree):
        return {p.relative_to(tree).as_posix(): hashlib.sha256(p.read_bytes()).hexdigest()
                for p in sorted((tree / 'src').rglob('*')) if p.is_file()}
    old, new = manifest(before), manifest(after)
    return sorted(name for name in old.keys() | new.keys() if old.get(name) != new.get(name))


def lines(tree, name):
    # Each line without its carriage return, so a CRLF file compares as its text.
    return [line.rstrip('\r') for line in (tree / name).read_bytes().decode().split('\n')]


def reason(key):
    return key.split(': ', 1)[1] if key.startswith(('NotYet: ', 'Refused: ')) else key


def sites(site_where, run):
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
    for kind, where, why, text in unique:
        byReason.setdefault(kind + ': ' + why, []).append((site_where(where), text))
    return {key: sorted(found) for key, found in byReason.items()}


def measure(report, tree, run):
    summary = report.latent_summary(tree, run)
    reasons = Counter(summary['per_reason'])
    found = sites(report.site_where, run)
    # The sites read here are the ones the meter counts, or the part says so rather than print a different set.
    drift = sorted(key for key in reasons.keys() | found.keys() if reasons.get(key, 0) != len(found.get(key, [])))
    result = {'totals': summary['totals'], 'latent_total': summary['totals']['NotYet'] + summary['totals']['Refused'],
              'source_files': summary['source_files'], 'site_drift': drift}
    return result, reasons, found


def line_maps(before, after, names):
    # For each tree-relative file name that differs, before's line numbers to after's where the line stands in an
    # unchanged run, and the set of after's lines that do. A line in a changed run maps to 'changed' on both sides.
    maps = {}
    for name in names:
        old = (before / name).read_bytes().decode().split('\n') if (before / name).is_file() else []
        new = (after / name).read_bytes().decode().split('\n') if (after / name).is_file() else []
        forward = {}
        for tag, i1, i2, j1, j2 in difflib.SequenceMatcher(None, old, new, autojunk=False).get_opcodes():
            if tag == 'equal':
                for offset in range(i2 - i1):
                    forward[i1 + offset + 1] = j1 + offset + 1
        maps[name] = (forward, set(forward.values()))
    return maps


def mapped_line(maps, name, line, side):
    if name not in maps or not isinstance(line, int):
        return line
    forward, kept = maps[name]
    if side == 'before':
        return forward.get(line, 'changed')
    return line if line in kept else 'changed'


def mapped_where(maps, where, side):
    # A census site's where, src/compiler/checker.ts:5552:9, in after's numbering.
    parts = where.rsplit(':', 2)
    if len(parts) != 3 or not parts[1].isdigit():
        return where
    line = mapped_line(maps, parts[0], int(parts[1]), side)
    return f"{parts[0]}:{line}:{parts[2] if isinstance(line, int) else '?'}"


def rooted(mapping, tree):
    # A reason, and a site's text, that names a declaration prints its tree's root (read at <tree>/src/...), so each
    # tree's root reads as <tree>, resolved or as given, before reasons or sites are compared; otherwise the same refusal
    # counts as one gone and one grown. A count adds up under the rooted key; a list of (where, text) sites joins.
    roots = sorted({str(tree.resolve()), str(tree)}, key=len, reverse=True)
    def root(text):
        for spelling in roots:
            text = text.replace(spelling, '<tree>')
        # A site's text opens with its own location, <tree>/file:line:column:, in its own tree's numbering; the site's
        # where, mapped to after's numbering, already says where.
        return re.sub(r'^<tree>/[^ :]+:\d+:\d+: ', '', text)
    result = type(mapping)()
    for key, value in mapping.items():
        if isinstance(value, list):
            result[root(key)] = result.get(root(key), []) + [(where, root(text)) for where, text in value]
        else:
            result[root(key)] = result.get(root(key), 0) + value
    return result


def keyed(sites, maps, side):
    # Every site as one comparable (where in after's numbering, reason, text).
    return {(mapped_where(maps, where, side), key, text) for key, found in sites.items() for where, text in found}


SEEN = re.compile(r'a value of type (?P<own>.+?) seen as (?P<view>.+?), (?:whose|which) ')


def mutable_view(key):
    # A Refused view reason whose view is a Mutable<...>.
    if not key.startswith('Refused: '):
        return False
    match = SEEN.match(reason(key))
    return bool(match) and match['view'].startswith('Mutable<')


CALLS = {'update': re.compile(r'(?<![\w$.])update\(\s*$'), 'finishUpdate': re.compile(r'(?<![\w$.])finishUpdate\w*\(\s*$')}


class Callee:
    # Which of nodeFactory.ts's helpers a site stands as the first argument of, read from its own tree's line: the text
    # before the site's column ends `update(` or `finishUpdate...(`.
    def __init__(self, tree):
        self.tree, self.cache = tree, {}

    def __call__(self, where):
        parts = where.rsplit(':', 2)
        if len(parts) != 3 or parts[0] != NODE_FACTORY or not parts[1].isdigit() or not parts[2].isdigit():
            return None
        if parts[0] not in self.cache:
            self.cache[parts[0]] = lines(self.tree, parts[0])
        text = self.cache[parts[0]]
        line = int(parts[1])
        if line > len(text):
            return None
        before = text[line - 1][:int(parts[2]) - 1]
        return next((name for name, call in CALLS.items() if call.search(before)), None)


def at(sites, tree, name):
    # The Mutable view sites standing at one helper's calls, as (where, reason, text) in the tree's own numbering.
    callee = Callee(tree)
    return sorted((where, key, text) for key, found in sites.items() if mutable_view(key) for where, text in found
                  if callee(where) == name)


def flat(text, limit=400):
    line = ' '.join(text.split())
    return line if len(line) <= limit else f"{line[:limit]} ... ({len(line)} characters)"
PYTHON

# (0) The sources: before and after differ under src/ in nodeFactory.ts alone, and there in update's one line.
diff -u "$before/$nodeFactory" "$after/$nodeFactory" > "$out/0-nodeFactory.diff" 2>&1
python3 -I - "$before" "$after" "$out" 2>> "$out/python.stderr" << 'PYTHON' | tee -a "$out/verdicts.log"
import sys
from pathlib import Path
before, after, out = map(Path, sys.argv[1:4])
sys.dont_write_bytecode = True
sys.path.insert(0, str(out))
import updlib
differing = updlib.differing(before, after)
files = sum(1 for p in (after / 'src').rglob('*') if p.is_file())
diff = (out / '0-nodeFactory.diff').read_text().rstrip('\n').split('\n')
hunks = [line for line in diff if line.startswith('@@')]
print(f"upd: (0) nodeFactory.ts: {len(hunks)} hunks ({out / '0-nodeFactory.diff'}):")
for line in diff[2:42]:
    print(f"upd: (0)   | {line.rstrip(chr(13))}")
old, new = updlib.lines(before, updlib.NODE_FACTORY), updlib.lines(after, updlib.NODE_FACTORY)
found = [index for index, line in enumerate(old) if line == updlib.UPSTREAM]
rebuilt = list(old)
if len(found) == 1:
    rebuilt[found[0]] = updlib.ADAPTED
rebuiltOk = len(found) == 1 and rebuilt == new
print(f"upd: (0) before's nodeFactory.ts holds update's upstream line {len(found)} times"
      f"{f', at line {found[0] + 1}' if len(found) == 1 else ''}; with that line alone adapted it is after's: {'yes' if rebuiltOk else 'no'}")
ok = differing == [updlib.NODE_FACTORY] and rebuiltOk
print(f"upd: verdict (0) sources {'PASS' if ok else 'FAIL'}: {files} src files, differing {differing}, nodeFactory.ts in "
      f"{len(hunks)} hunks, after's nodeFactory.ts before's with update's `updated: Mutable<T>` as `updated: T` and "
      f"nothing else: {'yes' if rebuiltOk else 'no'}")
PYTHON

# (0b) adapt.cjs on its own: a second run over after's nodeFactory.ts changes nothing, and a run over before's yields
# after's bytes. Each runs in a small tree holding that one file, a copy, so neither real tree is touched.
mkdir -p "$work/again/src/compiler/factory" "$work/fresh/src/compiler/factory"
cp "$after/$nodeFactory" "$work/again/$nodeFactory"
cp "$before/$nodeFactory" "$work/fresh/$nodeFactory"
NODE_PATH="$modules" step 0-adapt-again node "$root/stage3/adapt/$adaptation/adapt.cjs" "$work/again"
againExit=$?
NODE_PATH="$modules" step 0-adapt-fresh node "$root/stage3/adapt/$adaptation/adapt.cjs" "$work/fresh"
freshExit=$?
quote 2 "$out/0-adapt-again.log" "$out/0-adapt-fresh.log"
againSame=no; cmp -s "$after/$nodeFactory" "$work/again/$nodeFactory" && againSame=yes
freshSame=no; cmp -s "$after/$nodeFactory" "$work/fresh/$nodeFactory" && freshSame=yes
againSaid=no; grep -q 'already adapted' "$out/0-adapt-again.log" && againSaid=yes
if [ "$againExit" = 0 ] && [ "$freshExit" = 0 ] && [ "$againSame" = yes ] && [ "$freshSame" = yes ] && [ "$againSaid" = yes ]; then
	say "upd: verdict (0b) idempotent PASS: a second run exits 0, says already adapted and leaves after's bytes; a run over before's nodeFactory.ts yields after's bytes"
else
	say "upd: verdict (0b) idempotent FAIL: second run exit $againExit, already adapted said $againSaid, bytes kept $againSame; run over before's exit $freshExit, yields after's bytes $freshSame"
fi

# (1) Stock tsc: the program 70's audit builds, on each tree, and every syntactic, global and semantic diagnostic.
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
	say "upd: verdict (1) stock tsc FAIL: the program didn't build (exits $diagnosticsBeforeExit before, $diagnosticsAfterExit after)"
	quote 20 "$out/1-diagnostics-before.log" "$out/1-diagnostics-after.log"
else
	python3 -I - "$out" "$before" "$after" 2>> "$out/python.stderr" << 'PYTHON' | tee -a "$out/verdicts.log"
import json, sys
from collections import Counter
from pathlib import Path
out, beforeTree, afterTree = map(Path, sys.argv[1:4])
sys.dont_write_bytecode = True
sys.path.insert(0, str(out))
import updlib
before, after = (json.loads((out / f'1-diagnostics-{label}.json').read_text()) for label in ('before', 'after'))
# Each differing file's lines compare in after's numbering; the change is one line in place, so no line moves.
maps = updlib.line_maps(beforeTree, afterTree, updlib.differing(beforeTree, afterTree))
roots = {'before': str(beforeTree.resolve()), 'after': str(afterTree.resolve())}
def key(row, side):
    file = row.get('file') or ''
    message = row['message'].replace(roots[side], '<tree>')
    return (row['code'], file, updlib.mapped_line(maps, file, row.get('line') or 0, side), message)
old = Counter(key(row, 'before') for row in before['diagnostics'])
new = Counter(key(row, 'after') for row in after['diagnostics'])
added, removed = new - old, old - new
print(f"upd: (1) diagnostics: before {len(before['diagnostics'])} over {before['roots']} roots, after {len(after['diagnostics'])} over {after['roots']} roots")
for label, rows in (('new', added), ('gone', removed)):
    listed = sorted(rows.elements(), key=str)
    for code, file, line, message in listed[:40]:
        print(f"upd: (1)   {label} TS{code} {file}:{line} {updlib.flat(message, 300)}")
    if len(listed) > 40:
        print(f"upd: (1)   {len(listed) - 40} more {label} diagnostics in {out}/1-diagnostics-difference.json")
(out / '1-diagnostics-difference.json').write_text(json.dumps({'new': sorted(map(list, added.elements()), key=str),
    'gone': sorted(map(list, removed.elements()), key=str)}, indent=2) + '\n')
ok = not added and not removed
print(f"upd: verdict (1) stock tsc {'PASS' if ok else 'FAIL'}: {len(after['diagnostics'])} diagnostics after, "
      f"{sum(added.values())} new and {sum(removed.values())} gone, by (code, file, line, message) in after's numbering")
PYTHON
fi

# (2) The JavaScript: every src file the program reads, transpiled on each tree with 70's module and target, its bytes'
# sha256 kept by file.
cat > "$out/javascript.cjs" << 'NODE'
'use strict';
const crypto = require('node:crypto');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
const tree = path.resolve(process.argv[2]);
const output = path.resolve(process.argv[3]);
const compilerOptions = { module: ts.ModuleKind.ESNext, target: ts.ScriptTarget.ES2024, verbatimModuleSyntax: true };
const roots = ts.sys.readDirectory(path.join(tree, 'src'), ['.ts'], ['**/lib/**']).filter(name => !name.endsWith('.d.ts'));
const files = {};
for (const name of roots.sort()) {
    const text = fs.readFileSync(name, 'utf8');
    const result = ts.transpileModule(text, { compilerOptions, fileName: path.relative(tree, name), reportDiagnostics: false });
    files[path.relative(tree, name)] = crypto.createHash('sha256').update(result.outputText).digest('hex');
}
fs.writeFileSync(output, JSON.stringify({ typescript: ts.version, files }) + '\n');
console.log(`${roots.length} files transpiled`);
NODE
(NODE_PATH="$modules" step 2-javascript-before node "$out/javascript.cjs" "$before" "$out/2-javascript-before.json") &
javascriptBefore=$!
(NODE_PATH="$modules" step 2-javascript-after node "$out/javascript.cjs" "$after" "$out/2-javascript-after.json") &
javascriptAfter=$!
wait "$javascriptBefore"
javascriptBeforeExit=$?
wait "$javascriptAfter"
javascriptAfterExit=$?
if [ "$javascriptBeforeExit" != 0 ] || [ "$javascriptAfterExit" != 0 ]; then
	say "upd: verdict (2) javascript FAIL: the transpile didn't run (exits $javascriptBeforeExit before, $javascriptAfterExit after)"
	quote 20 "$out/2-javascript-before.log" "$out/2-javascript-after.log"
else
	python3 -I - "$out" 2>> "$out/python.stderr" << 'PYTHON' | tee -a "$out/verdicts.log"
import json, sys
from pathlib import Path
out = Path(sys.argv[1])
before, after = (json.loads((out / f'2-javascript-{label}.json').read_text())['files'] for label in ('before', 'after'))
differ = sorted(name for name in before.keys() | after.keys() if before.get(name) != after.get(name))
for name in differ[:40]:
    print(f"upd: (2)   differs {name}")
node = 'src/compiler/factory/nodeFactory.ts'
ok = not differ and len(before) > 0 and node in after
print(f"upd: verdict (2) javascript {'PASS' if ok else 'FAIL'}: {len(after)} files transpiled on each tree, nodeFactory.ts "
      f"among them: {'yes' if node in after else 'no'}, {len(differ)} differ in their JavaScript")
PYTHON
fi

# census <label> <tree>: the latent census as the meter runs it (first error, scoped tail, no output), into $work/census-<label>.
census() {
	mkdir -p "$work/census-$1"
	LATENT_TIMES="$work/census-$1/latent-times.jsonl" LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TAIL=1 \
		step "census-$1" "$work/latent-census" "$2/src/compiler" "$work/census-$1/latent.jsonl"
	local code=$?
	gzip -c "$work/census-$1/latent.jsonl" > "$out/census-$1.jsonl.gz" 2> /dev/null
	return "$code"
}

# (3) The census before and after, in parallel.
if [ "$censusBuildExit" != 0 ]; then
	say "upd: verdict (3) census FAIL: the latent census didn't build"
	quote 30 "$out/census-overlay.log" "$out/census-gofmt.log" "$out/census-build.log"
else
	census before "$before" &
	censusBefore=$!
	census after "$after" &
	censusAfter=$!
	wait "$censusBefore"
	censusBeforeExit=$?
	wait "$censusAfter"
	censusAfterExit=$?
	python3 -I - "$root/stage3/meter" "$out" "$before" "$work/census-before" "$after" "$work/census-after" "$censusBeforeExit" "$censusAfterExit" 2>> "$out/python.stderr" << 'PYTHON' | tee -a "$out/verdicts.log"
import json, sys
from collections import Counter
from pathlib import Path
sys.dont_write_bytecode = True
meter, out = Path(sys.argv[1]), Path(sys.argv[2])
sys.path[:0] = [str(out), str(meter)]
import report  # The meter's own reader: its coverage checks, its unique-site count and its site_where.
import updlib
beforeTree, beforeRun, afterTree, afterRun = map(Path, sys.argv[3:7])
codes = sys.argv[7:9]
try:
    before, beforeReasons, beforeSites = updlib.measure(report, beforeTree, beforeRun)
    after, afterReasons, afterSites = updlib.measure(report, afterTree, afterRun)
except Exception as error:
    print(f"upd: verdict (3) census FAIL: a census didn't read whole (exits {codes}): {error}")
    sys.exit(0)
beforeReasons, beforeSites = updlib.rooted(beforeReasons, beforeTree), updlib.rooted(beforeSites, beforeTree)
afterReasons, afterSites = updlib.rooted(afterReasons, afterTree), updlib.rooted(afterSites, afterTree)
for name, r in (('before', before), ('after', after)):
    print(f"upd: (3) {name}: {r['source_files']} files, latent total {r['latent_total']} {r['totals']}")
    if r['site_drift']:
        print(f"upd: (3) {name}: the sites read here disagree with per_reason for {len(r['site_drift'])} reasons, first {r['site_drift'][:3]}")
# The update sites and the finishUpdate sites, each in its own tree's numbering.
updates = [updlib.at(beforeSites, beforeTree, 'update'), updlib.at(afterSites, afterTree, 'update')]
finishes = [updlib.at(beforeSites, beforeTree, 'finishUpdate'), updlib.at(afterSites, afterTree, 'finishUpdate')]
print(f"upd: (3) Mutable views refused at update(...) calls: {len(updates[0])} before, {len(updates[1])} after")
for label, rows in zip(('before', 'after'), updates):
    for where, key, text in rows[:50]:
        print(f"upd: (3)   {label} update site {where}: {updlib.flat(updlib.reason(key), 200)}")
beforeWheres = sorted({where for where, _, _ in updates[0]})
print(f"upd: (3) before's update sites are the W2 rows' 44 exactly: {'yes' if beforeWheres == updlib.W2_UPDATE_SITES else 'no'}"
      f"; in the rows, not the census: {sorted(set(updlib.W2_UPDATE_SITES) - set(beforeWheres))[:10]}"
      f"; in the census, not the rows: {sorted(set(beforeWheres) - set(updlib.W2_UPDATE_SITES))[:10]}")
print(f"upd: (3) Mutable views refused at finishUpdate(...) calls, left out: {len(finishes[0])} before, {len(finishes[1])} after")
for where, key, text in finishes[1][:10]:
    print(f"upd: (3)   after finishUpdate site {where}: {updlib.flat(updlib.reason(key), 200)}")
maps = updlib.line_maps(beforeTree, afterTree, updlib.differing(beforeTree, afterTree))
old, new = updlib.keyed(beforeSites, maps, 'before'), updlib.keyed(afterSites, maps, 'after')
fresh, gone = sorted(new - old), sorted(old - new)
updateGone = {(updlib.mapped_where(maps, where, 'before'), key, text) for where, key, text in updates[0]}
print(f"upd: (3) sites, every reason: {len(old)} before, {len(new)} after; {len(gone)} gone, {len(updateGone & set(gone))} of "
      f"them update sites; {len(fresh)} new")
for where, key, text in [site for site in gone if site not in updateGone][:40]:
    print(f"upd: (3)   gone, not an update site {where}: {updlib.flat(key, 200)}: {updlib.flat(text, 200)}")
for where, key, text in fresh[:40]:
    print(f"upd: (3)   new {where}: {updlib.flat(key, 200)}: {updlib.flat(text, 200)}")
delta = Counter(afterReasons)
delta.subtract(beforeReasons)
moved = sorted(((count, key) for key, count in delta.items() if count), key=lambda item: (-(item[0] > 0), -abs(item[0]), item[1]))
print(f"upd: (3) reasons whose count moves: {len(moved)}, {sum(1 for count, _ in moved if count > 0)} grow")
for count, key in moved[:60]:
    print(f"upd: (3)   {'+' if count > 0 else ''}{count} ({beforeReasons.get(key, 0)} to {afterReasons.get(key, 0)}) {updlib.flat(key)}")
(out / '3-sites.json').write_text(json.dumps({'update_before': updates[0], 'update_after': updates[1],
    'finish_before': finishes[0], 'finish_after': finishes[1], 'gone': gone, 'new': fresh,
    'reason_delta': {key: count for key, count in delta.items() if count}}, indent=2) + '\n')
drifted = before['site_drift'] or after['site_drift']
ok = len(updates[0]) > 0 and not updates[1] and not fresh and not drifted
print(f"upd: verdict (3) census {'PASS' if ok else 'FAIL'}: latent total {before['latent_total']} to {after['latent_total']}, "
      f"Mutable views at update calls {len(updates[0])} to {len(updates[1])} (44 by the W2 rows, 0 expected after), "
      f"finishUpdate's {len(finishes[0])} to {len(finishes[1])}, {len(fresh)} sites new (0 expected), "
      f"{len(gone)} gone{', and the sites read disagree with the meter' if drifted else ''}")
PYTHON
fi

# (4) The mutant: the after tree with update's `updated: T` back to `updated: Mutable<T>`, and nothing else; so its src
# must be before's byte for byte, and its census must bring back exactly the update sites (3) saw go. The tree is
# after's linked whole (cp -al, or a copy where links can't be made), and nodeFactory.ts is written as a new file in its
# place, so after's own nodeFactory.ts is never touched.
if [ "$censusBuildExit" != 0 ]; then
	say "upd: verdict (4) mutant FAIL: the latent census didn't build"
elif ! { cp -al "$after" "$mutant" 2> /dev/null || { rm -rf -- "$mutant"; cp -a "$after" "$mutant"; }; }; then
	say "upd: verdict (4) mutant FAIL: after's tree didn't copy to $mutant"
else
	chmod u+w "$mutant/src/compiler/factory"
	python3 -I - "$out" "$before" "$after" "$mutant" > "$out/4-mutant-sources.log" 2>> "$out/python.stderr" << 'PYTHON'
import os, sys
from pathlib import Path
out, before, after, mutant = map(Path, sys.argv[1:5])
sys.dont_write_bytecode = True
sys.path.insert(0, str(out))
import updlib
name = updlib.NODE_FACTORY
text = (after / name).read_bytes().decode()
newline = '\r\n' if '\r\n' in text else '\n'
lines = text.split(newline)
found = [index for index, line in enumerate(lines) if line == updlib.ADAPTED]
if len(found) != 1:
    print(f"upd: (4) after's nodeFactory.ts holds update's adapted line {len(found)} times, not once; no mutant")
    sys.exit(1)
lines[found[0]] = updlib.UPSTREAM
# A new file in the link's place, so after's nodeFactory.ts keeps its bytes.
(mutant / (name + '.mutant')).write_bytes(newline.join(lines).encode())
os.replace(mutant / (name + '.mutant'), mutant / name)
untouched = (after / name).read_bytes().decode() == text
fromAfter, fromBefore = updlib.differing(after, mutant), updlib.differing(before, mutant)
ok = fromAfter == [name] and fromBefore == [] and untouched
print(f"upd: (4) the mutant's src differs from after's in {fromAfter} and from before's in {fromBefore or 'nothing'}, after's own "
      f"nodeFactory.ts untouched: {'yes' if untouched else 'no'}: {'Mutable<T> restored, and nothing else' if ok else 'not the mutant'}")
sys.exit(0 if ok else 1)
PYTHON
	mutantSourcesExit=$?
	tee -a "$out/verdicts.log" < "$out/4-mutant-sources.log"
	if [ "$mutantSourcesExit" != 0 ]; then
		say "upd: verdict (4) mutant FAIL: the mutant tree isn't after's with Mutable<T> restored (exit $mutantSourcesExit), so its census would measure something else"
	else
		census mutant "$mutant"
		censusMutantExit=$?
		python3 -I - "$root/stage3/meter" "$out" "$before" "$work/census-before" "$after" "$work/census-after" "$mutant" "$work/census-mutant" "$censusMutantExit" 2>> "$out/python.stderr" << 'PYTHON' | tee -a "$out/verdicts.log"
import sys
from collections import Counter
from pathlib import Path
sys.dont_write_bytecode = True
meter, out = Path(sys.argv[1]), Path(sys.argv[2])
sys.path[:0] = [str(out), str(meter)]
import report
import updlib
beforeTree, beforeRun, afterTree, afterRun, mutantTree, mutantRun = map(Path, sys.argv[3:9])
try:
    _, _, beforeSites = updlib.measure(report, beforeTree, beforeRun)
    after, _, afterSites = updlib.measure(report, afterTree, afterRun)
    mutant, _, mutantSites = updlib.measure(report, mutantTree, mutantRun)
except Exception as error:
    print(f"upd: verdict (4) mutant FAIL: a census didn't read whole (the mutant's exit {sys.argv[9]}): {error}")
    sys.exit(0)
beforeSites = updlib.rooted(beforeSites, beforeTree)
afterSites = updlib.rooted(afterSites, afterTree)
mutantSites = updlib.rooted(mutantSites, mutantTree)
# The mutant's src is before's, so its sites read in before's numbering; after's map to it through the one changed line.
maps = updlib.line_maps(mutantTree, afterTree, updlib.differing(mutantTree, afterTree))
mutantKeyed = updlib.keyed(mutantSites, maps, 'before')
afterKeyed = updlib.keyed(afterSites, maps, 'after')
beforeKeyed = updlib.keyed(beforeSites, {}, 'before')
mutantOwn = updlib.keyed(mutantSites, {}, 'before')
back = sorted(mutantKeyed - afterKeyed)
lost = sorted(afterKeyed - mutantKeyed)
went = sorted((updlib.mapped_where(maps, where, 'before'), key, text) for where, key, text in updlib.at(beforeSites, beforeTree, 'update'))
print(f"upd: (4) mutant: latent total {mutant['latent_total']}, where after holds {after['latent_total']}; {len(back)} sites the "
      f"mutant holds and after doesn't, {len(lost)} after holds and the mutant doesn't")
for where, key, text in back[:50]:
    print(f"upd: (4)   back {where}: {updlib.flat(key, 200)}")
for where, key, text in lost[:20]:
    print(f"upd: (4)   lost {where}: {updlib.flat(key, 200)}")
same = mutantOwn == beforeKeyed
print(f"upd: (4) the mutant's sites are before's exactly: {'yes' if same else 'no'} "
      f"({len(mutantOwn - beforeKeyed)} only in the mutant, {len(beforeKeyed - mutantOwn)} only in before)")
ok = len(went) > 0 and back == went and not lost and same
print(f"upd: verdict (4) mutant {'PASS' if ok else 'FAIL'}: with Mutable<T> restored, {len(back)} sites come back, the "
      f"{len(went)} update sites (3) saw go exactly: {'yes' if back == went else 'no'}, none lost, the census before's: {'yes' if same else 'no'}")
PYTHON
	fi
fi

# The summary and the exit.
fails=$(grep -c 'verdict .* FAIL' "$out/verdicts.log" 2> /dev/null)
passes=$(grep -c 'verdict .* PASS' "$out/verdicts.log" 2> /dev/null)
echo "upd: summary: ${passes:-0} PASS, ${fails:-0} FAIL"
grep 'verdict ' "$out/verdicts.log" | sed 's/^/upd: summary: /' | cut -c1-400
[ "${fails:-0}" = 0 ] && [ "${passes:-0}" -ge 7 ] && exit 0
exit 1
