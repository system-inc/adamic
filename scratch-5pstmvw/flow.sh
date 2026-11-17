#!/usr/bin/env bash
# flow.sh: a scratch measurement for #5pstmvw, never to land. Adaptation 74, stage3/adapt/74-readonly-flow-slots, makes
# node and antecedent readonly on FlowNodeBase and on every arm of the FlowNode union in src/compiler/types.ts, except
# FlowLabel's antecedent, and folds bindContainer's FlowStart statement pair in src/compiler/binder.ts into one
# createFlowNode call that takes the node at construction. This measures what it changes. Run from an adamic tree's root,
# on a sealed job as
#
#   source /opt/adamic-block/environment.sh && source cloud/sealed-job.sh && export GOFLAGS=... && sealed_typescript_source &&
#   sealed_npm stage3/api stage3/typescript && ... && bash scratch-5pstmvw/flow.sh
#
# never with sealed_git in the chain: a sealed tree already has a .git, sealed_git refuses it, and the && chain stops.
# Or on a house box after a lay line that lays cohere, cohere/TypeScript and both npm installs, and ~/adamic-tools/env.sh.
#
#   before: stage3/apply.sh over a copy of this tree's stage3 inputs with 74-readonly-flow-slots left out
#   after:  this tree's own stage3/apply.sh, which runs 74 after 72-readonly-diagnostic-slots and before 75
#   mutant: stage3/apply.sh over a copy of this tree's stage3 inputs whose 74 leaves FlowAssignment's node writable
#
#   (0) sources: diff -u of types.ts and binder.ts, before to after; PASS is those two files alone differing under src/
#       and binder.ts's change one hunk
#   (1) writes: scratch-5pstmvw/writes.cjs, every write to a flow node's node or antecedent found by the type checker;
#       PASS before is node written once, in binder.ts, through a FlowNode, every antecedent written through a FlowLabel,
#       and no Object.assign onto a flow node; PASS after is node written nowhere
#   (2) stock tsc: one ts.createProgram per tree over all upstream src, 70's options, stock TypeScript 6.0.3; PASS is no
#       new TS2540 and every diagnostic (code, file, line, message) the same before and after
#   (3) the meter: the latent census on both src/compiler trees, read by report.py's latent_summary; the family is every
#       view or union-write refusal over a flow node (flowlib.py below says why that and not a word match); PASS is the
#       family at 0 after and no reason grown
#   (4) the mutant: the census on the mutant tree; PASS is the family back above 0, so (3)'s 0 is the readonly slots' doing
#   (5) the lane, whole, on the after tree: every unit of stage3/lane/shards.json run by shard.py against the after tree
#       built by hereby with stage3/typescript's install, as the gate's product path runs it; or, on a box with npm and
#       no laid install, stage3/lane/run.sh; SKIP with the reason when neither can run here. 74 is internal, so
#       stage3-lane-byte-comparisons must pass: the published typescript.d.ts is byte-identical
#   (6) the plan: stage3/lane/test_check.py as the lane's own Go wrapper runs it, shard-plan.json's runtime hash, and
#       git status of stage3/lane after everything, which must be empty
#
# Every line printed begins "flow:", verdict lines "flow: verdict". Logs, JSON and the summary go to $TMPDIR/flow-<stamp>;
# long lists stay there, and at most about 40 sites print per section, since a job's output is cut near 150 KB. The heavy
# trees go to $TMPDIR too (FLOW_WORK names another place), never beside the tree, which a sealed job lays read-only, and
# are removed at the end unless FLOW_KEEP=1. Exit 0 when every part passes, 1 when any fails, 3 when none fails and one
# is skipped; 64 to 66 are preparation failures.
set -u

root=$PWD
adaptation=74-readonly-flow-slots
[ -f "$root/stage3/apply.sh" ] && [ -f "$root/stage3/adapt/$adaptation/adapt.cjs" ] && [ -f "$root/scratch-5pstmvw/writes.cjs" ] || {
	echo "flow: $root is not an adamic tree's root carrying stage3/adapt/$adaptation and scratch-5pstmvw/writes.cjs, so nothing ran"
	exit 64
}
stamp=$(date -u +%Y%m%dT%H%M%SZ)
scratch="${TMPDIR:-/tmp}"
out="${scratch%/}/flow-$stamp"
work="${FLOW_WORK:-${scratch%/}/5pstmvw-$stamp.flow}"
mkdir -p "$out" || exit 64
mkdir "$work" || { echo "flow: $work is there already or can't be made, so nothing ran"; exit 64; }
jobEnd() {
	local code=$?
	trap - EXIT
	if [ "${FLOW_KEEP:-0}" != 1 ]; then
		case "$work" in /*/*.flow) chmod -R u+w -- "$work" 2> /dev/null; rm -rf -- "$work" ;; esac
	fi
	echo "flow: evidence in $out"
	exit "$code"
}
trap jobEnd EXIT
trap 'exit 143' TERM
trap 'exit 129' HUP
trap 'exit 130' INT
mkdir "$work/tmp" || exit 64
export TMPDIR="$work/tmp"
# No python this script starts writes a __pycache__ into the tree: stage3/meter's report and stage3/lane's modules are
# imported from it, and step (6) reads stage3/lane's git status.
export PYTHONDONTWRITEBYTECODE=1

if [ -f "$HOME/adamic-tools/env.sh" ]; then
	. "$HOME/adamic-tools/env.sh" || { echo "flow: $HOME/adamic-tools/env.sh didn't source, so nothing ran"; exit 65; }
fi
missing=""
for tool in go gofmt node python3 git diff; do
	command -v "$tool" > /dev/null 2>&1 || missing="$missing $tool"
done
[ -z "$missing" ] || { echo "flow: missing from the toolchain:$missing"; exit 65; }
hasNpm=0; command -v npm > /dev/null 2>&1 && hasNpm=1
node --version | grep -q '^v24\.' || { echo "flow: node is $(node --version), adaptations need Node 24"; exit 65; }
[ -n "$(ls -A "$root/cohere" 2> /dev/null)" ] || { echo "flow: cohere isn't laid in $root"; exit 66; }
echo "flow: $(uname -n), $(nproc) threads, $(git -C "$root" rev-parse HEAD 2> /dev/null || echo 'no checkout'), node $(node --version), $(go version), npm $([ "$hasNpm" = 1 ] && npm --version || echo absent)"

# step <name> <command...>: the command's whole output in $out/<name>.log, and its exit.
step() {
	local name=$1
	shift
	"$@" > "$out/$name.log" 2>&1
	local code=$?
	echo "flow: $name exited $code ($out/$name.log)"
	return "$code"
}

# say <line>: a line printed and kept in $out/verdicts.log for the summary and the exit.
say() { echo "$*" | tee -a "$out/verdicts.log"; }

# quote <lines> <file...>: a file's last lines, each printed under the flow: prefix.
quote() {
	local lines=$1
	shift
	tail -n "$lines" "$@" 2> /dev/null | sed 's/^/flow:   | /'
}

# The pinned upstream checkout apply reads: the caller's, the laid cohere/TypeScript when it is the pin, else a clone.
pin() { python3 -I -c 'import json, sys; print(json.load(open(sys.argv[1]))[sys.argv[2]])' "$root/stage3/source.json" "$1"; }
pinCommit=$(pin commit)
if [ -n "${ADAMIC_TYPESCRIPT_SOURCE:-}" ] && [ "$(git -C "$ADAMIC_TYPESCRIPT_SOURCE" rev-parse HEAD 2> /dev/null)" = "$pinCommit" ]; then
	echo "flow: upstream is the caller's ADAMIC_TYPESCRIPT_SOURCE"
elif [ "$(git -C "$root/cohere/TypeScript" rev-parse HEAD 2> /dev/null)" = "$pinCommit" ]; then
	export ADAMIC_TYPESCRIPT_SOURCE="$root/cohere/TypeScript"
	echo "flow: upstream is the laid cohere/TypeScript, at the pin"
else
	step upstream-clone git clone --depth 1 --branch "$(pin tag)" "$(pin repository)" "$work/upstream" || exit 66
	export ADAMIC_TYPESCRIPT_SOURCE="$work/upstream"
fi
[ "$(git -C "$ADAMIC_TYPESCRIPT_SOURCE" rev-parse HEAD)" = "$pinCommit" ] || { echo "flow: upstream isn't at $pinCommit"; exit 66; }

# stage3/api's lockfile-pinned typescript@6.0.3, which apply puts on NODE_PATH for every adaptation.
modules="$root/stage3/api/node_modules"
[ "$(node -p 'require(process.argv[1] + "/typescript/package.json").version' "$modules" 2> /dev/null)" = 6.0.3 ] || { echo "flow: stage3/api's typescript isn't laid at 6.0.3"; exit 66; }

# stageInputs <directory> <left out>: a copy of this tree's apply inputs as <directory>/stage3, every adaptation but the
# one named (none when it is empty), and nothing else changed.
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

# Before's stage3: this tree's apply inputs with 74 left out.
beforeStage="$work/before-inputs/stage3"
stageInputs "$work/before-inputs" "$adaptation" || exit 64
diff -rq "$root/stage3/adapt" "$beforeStage/adapt" > "$out/before-inputs.diff" 2>&1
echo "flow: before's adapt differs from this tree's by: $(tr '\n' ' ' < "$out/before-inputs.diff")"

# The mutant's stage3: this tree's apply inputs whole, with 74's adapt.cjs copied through one sed that gives
# FlowAssignment's flowSlots entry writable: 'node', so 74 inserts no `readonly ` before FlowAssignment's node and every
# other edit stays. The sed's address is the line that opens FlowAssignment's entry; its substitution appends
# `, writable: 'node'` before that line's closing ` },`. The check after it is the proof that it is exact: the copy
# differs from 74's adapt.cjs in that one line alone, and the line reads as written here.
mutantStage="$work/mutant-inputs/stage3"
stageInputs "$work/mutant-inputs" "" || exit 64
mutantAdapt="$mutantStage/adapt/$adaptation/adapt.cjs"
sed "/^    FlowAssignment: { /s/, antecedent: 'FlowNode' },\$/, antecedent: 'FlowNode', writable: 'node' },/" \
	"$root/stage3/adapt/$adaptation/adapt.cjs" > "$mutantAdapt" || exit 64
diff "$root/stage3/adapt/$adaptation/adapt.cjs" "$mutantAdapt" > "$out/4-mutant-adapt.diff" 2>&1
python3 -I - "$root/stage3/adapt/$adaptation/adapt.cjs" "$mutantAdapt" << 'PYTHON' || exit 64
import sys
old, new = (open(path, encoding='utf-8').read().split('\n') for path in sys.argv[1:3])
was = "    FlowAssignment: { node: 'Expression | VariableDeclaration | BindingElement', antecedent: 'FlowNode' },"
now = "    FlowAssignment: { node: 'Expression | VariableDeclaration | BindingElement', antecedent: 'FlowNode', writable: 'node' },"
changed = [(index + 1, a, b) for index, (a, b) in enumerate(zip(old, new)) if a != b]
if len(old) != len(new) or len(changed) != 1 or changed[0][1:] != (was, now):
    print(f"flow: the mutant's sed isn't exact: {len(old)} lines to {len(new)}, changed {changed}; nothing ran")
    sys.exit(1)
print(f"flow: the mutant's adapt.cjs differs from 74's at line {changed[0][0]} alone: {now.strip()}")
PYTHON
quote 10 "$out/4-mutant-adapt.diff"
diff -rq "$root/stage3/adapt" "$mutantStage/adapt" > "$out/mutant-inputs.diff" 2>&1
echo "flow: the mutant's adapt differs from this tree's by: $(tr '\n' ' ' < "$out/mutant-inputs.diff")"
echo "flow: product keys: after $(cd "$root" && python3 stage3/apply.py --key), before $(python3 "$beforeStage/apply.py" --key), mutant $(python3 "$mutantStage/apply.py" --key)"

# Phase 1, in parallel: the three applies from a fresh local product store, and the census tool. A local store makes
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
(STAGE3_CACHE="$work/cache-mutant" step apply-mutant bash "$mutantStage/apply.sh" "$mutant") &
applyMutant=$!
(cd "$root" && step census-overlay python3 stage3/census/latent/make_overlay.py "$root" "$work/latent-overlay" &&
	step census-gofmt bash -c 'gofmt -w "$1"/*.go' _ "$work/latent-overlay" &&
	step census-build go build -buildvcs=false -overlay="$work/latent-overlay/overlay.json" -o "$work/latent-census" ./stage3/census/latent/tool) &
censusBuild=$!
wait "$applyBefore"
applyBeforeExit=$?
wait "$applyAfter"
applyAfterExit=$?
wait "$applyMutant"
applyMutantExit=$?
wait "$censusBuild"
censusBuildExit=$?
echo "flow: apply before exited $applyBeforeExit, apply after exited $applyAfterExit, apply mutant exited $applyMutantExit, census build exited $censusBuildExit"
[ "$applyBeforeExit" = 0 ] || { echo "flow: before's apply failed, so there is no before tree; its last lines:"; quote 30 "$out/apply-before.log"; exit 66; }
if [ "$applyAfterExit" != 0 ]; then
	# apply.py names the adaptation it was running in the failed command.
	failed=$(grep -o 'stage3/adapt/[0-9][0-9]-[^/]*/adapt.cjs' "$out/apply-after.log" | tail -n 1)
	say "flow: verdict apply FAIL: the after tree's apply exited $applyAfterExit${failed:+ in $failed}, so there is no after tree"
	quote 30 "$out/apply-after.log"
	for part in '(0) sources' '(1) writes' '(2) stock tsc' '(3) meter' '(4) mutant' '(5) lane' '(6) plan'; do
		say "flow: verdict $part SKIP: no after tree"
	done
	exit 1
fi
say "flow: verdict apply PASS: before and after applies exited 0; after's patch-set row: $(grep -F "$adaptation" "$after/patch-set.md")"

# The judgments more than one part reads, in one home: the family, the meter's sites, and before's lines in after's
# numbering.
cat > "$out/flowlib.py" << 'PYTHON'
import difflib, hashlib, json, re
from collections import Counter
from pathlib import Path


def differing(before, after):
    # The tree-relative names of the files under src/ whose bytes differ between two trees, or that one lacks.
    def manifest(tree):
        return {p.relative_to(tree).as_posix(): hashlib.sha256(p.read_bytes()).hexdigest()
                for p in sorted((tree / 'src').rglob('*')) if p.is_file()}
    old, new = manifest(before), manifest(after)
    return sorted(name for name in old.keys() | new.keys() if old.get(name) != new.get(name))

# The flow interfaces adaptation 74 edits, declared in src/compiler/types.ts, and FlowNode, the union of its arms.
FLOW = frozenset(['FlowNode', 'FlowNodeBase', 'FlowUnreachable', 'FlowStart', 'FlowLabel', 'FlowAssignment', 'FlowCall',
                  'FlowCondition', 'FlowSwitchClause', 'FlowArrayMutation', 'FlowReduceLabel'])
MENTION = re.compile(r'\b(?:' + '|'.join(sorted(FLOW, key=len, reverse=True)) + r')\b')

# The family. A per_reason key is "<kind>: <reason>", and a reason is the lowering's What with its types printed by
# TypeToString (stage3/census/latent/lower.go.txt sets Reason from Refused.What and NotYet.What). A refusal the flow
# slots cause takes one of these shapes:
#   internal/lower/invariance.go, wideningRefusal: "a value of type A seen as B, which can write W where R is read",
#     or ", whose readonly field F becomes writable: ...", ", whose method M becomes ...", ", a type parameter ...";
#   internal/lower/union_writes.go, unionSlotWrite: "a write through a union of V into a slot that holds S in M", or
#     "a write through a union to a slot M doesn't name".
# The "can write" shape names types, never the field, so a match on a flow name beside the words node or antecedent
# finds none of the refusals adaptation 74 answers: the meter's committed evidence reads them as "a value of type
# FlowCall seen as FlowNode, which can write BinaryExpression | CallExpression where CallExpression is read". So the
# family is a reason of those shapes whose A or B holds a flow interface as a whole union or intersection arm, or whose
# M is one. Every arm extends FlowNodeBase and redeclares node and antecedent (74's flowSlots refuses otherwise), so a
# view between flow nodes writes through those two slots or through what they hold. A flow name nested deeper, as an
# array's element (never[] seen as FlowNode[]) or another object's field ({ flowNode: FlowNode; ... } seen as
# FlowGraphNode), is another slot's refusal and stays out; every reason naming a flow interface that the family leaves
# out is printed with its count by left_out, so the line drawn here can be read and argued with.
VIEW = re.compile(r'a value of type (?P<own>.+?) seen as (?P<view>.+?), (?:which can write |whose readonly field |whose method |a type parameter )')
UNION_WRITE = re.compile(r'a write through a union of .+ into a slot that holds .+ in (?P<member>.+)')
UNION_NAMELESS = re.compile(r"a write through a union to a slot (?P<member>.+) doesn't name")


def arms(text):
    # A type's top-level union and intersection arms, as TypeToString prints them; a function type's => is no bracket.
    parts, depth, start, index = [], 0, 0, 0
    while index < len(text):
        character = text[index]
        if character in '([{<':
            depth += 1
        elif character in ')]}' or (character == '>' and text[index - 1:index] != '='):
            depth -= 1
        elif depth == 0 and (text.startswith(' | ', index) or text.startswith(' & ', index)):
            parts.append(text[start:index])
            index += 3
            start = index
            continue
        index += 1
    parts.append(text[start:])
    return [part.strip() for part in parts]


def reason(key):
    return key.split(': ', 1)[1] if key.startswith(('NotYet: ', 'Refused: ')) else key


def family(key):
    text = reason(key)
    if (match := VIEW.match(text)):
        return any(arm in FLOW for arm in arms(match['own']) + arms(match['view']))
    if (match := UNION_WRITE.fullmatch(text)) or (match := UNION_NAMELESS.fullmatch(text)):
        return match['member'].strip() in FLOW
    return False


def left_out(key):
    return not family(key) and MENTION.search(reason(key)) is not None


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
              'source_files': summary['source_files'],
              'family': {key: count for key, count in reasons.items() if family(key)},
              'left_out': {key: count for key, count in reasons.items() if left_out(key)}, 'site_drift': drift}
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
    # A census site's where, src/compiler/binder.ts:120:9, in after's numbering.
    parts = where.rsplit(':', 2)
    if len(parts) != 3 or not parts[1].isdigit():
        return where
    line = mapped_line(maps, parts[0], int(parts[1]), side)
    return f"{parts[0]}:{line}:{parts[2] if isinstance(line, int) else '?'}"


def flat(text, limit=400):
    line = ' '.join(text.split())
    return line if len(line) <= limit else f"{line[:limit]} ... ({len(line)} characters)"
PYTHON

# (0) The sources: before and after differ under src/ in types.ts and binder.ts alone, and binder.ts's change is one hunk.
diff -u "$before/src/compiler/types.ts" "$after/src/compiler/types.ts" > "$out/0-types.diff" 2>&1
diff -u "$before/src/compiler/binder.ts" "$after/src/compiler/binder.ts" > "$out/0-binder.diff" 2>&1
python3 -I - "$before" "$after" "$out" 2>> "$out/python.stderr" << 'PYTHON' | tee -a "$out/verdicts.log"
import difflib, json, re, sys
from pathlib import Path
before, after, out = map(Path, sys.argv[1:4])
sys.dont_write_bytecode = True
sys.path.insert(0, str(out))
import flowlib
differing = flowlib.differing(before, after)
files = sum(1 for p in (after / 'src').rglob('*') if p.is_file())
binder = (out / '0-binder.diff').read_text().rstrip('\n').split('\n')
hunks = [index for index, line in enumerate(binder) if line.startswith('@@')]
body = binder[hunks[0]:] if hunks else []
print(f"flow: (0) binder.ts: {len(hunks)} hunks ({out / '0-binder.diff'}):")
for line in body[:40]:
    print(f"flow: (0)   | {line}")
if len(body) > 40:
    print(f"flow: (0)   ... {len(body) - 40} more lines in {out / '0-binder.diff'}")
# types.ts: each changed line paired with its new text and its owner, the nearest interface above it.
oldLines = (before / 'src/compiler/types.ts').read_bytes().decode().split('\n')
newLines = (after / 'src/compiler/types.ts').read_bytes().decode().split('\n')
declaration = re.compile(r'^(?:export )?interface (\w+)\b')
def owner(line):
    for text in reversed(oldLines[:line]):
        if (match := declaration.match(text)):
            return match.group(1)
    return None
removed = added = 0
paired = []
for tag, i1, i2, j1, j2 in difflib.SequenceMatcher(None, oldLines, newLines, autojunk=False).get_opcodes():
    if tag == 'equal':
        continue
    removed += i2 - i1
    added += j2 - j1
    if tag == 'replace' and i2 - i1 == j2 - j1:
        paired += [(i1 + offset + 1, oldLines[i1 + offset], newLines[j1 + offset]) for offset in range(i2 - i1)]
def inserted(a, b):
    indent = len(a) - len(a.lstrip())
    return b == a[:indent] + 'readonly ' + a[indent:]
insertions = [(line, owner(line), a.strip().split(':')[0]) for line, a, b in paired if inserted(a, b)]
slots = [f"{name}.{slot}" for _, name, slot in insertions]
print(f"flow: (0) types.ts: {removed} lines removed and {added} added, {len(insertions)} of them a `readonly ` inserted "
      f"(74 expects 19, node and antecedent on ten interfaces but FlowLabel's antecedent): {', '.join(slots)}")
others = [(line, a.strip(), b.strip()) for line, a, b in paired if not inserted(a, b)]
if others or removed != added or removed != len(paired):
    print(f"flow: (0) types.ts changes other than a `readonly ` inserted: {others[:10]}, see {out / '0-types.diff'}")
json.dump({'files': files, 'differing': differing, 'binder_hunks': len(hunks), 'types_ts': {'removed': removed, 'added': added, 'readonly_insertions': slots}},
          open(out / '0-sources.json', 'w'), indent=2)
ok = differing == ['src/compiler/binder.ts', 'src/compiler/types.ts'] and len(hunks) == 1
print(f"flow: verdict (0) sources {'PASS' if ok else 'FAIL'}: {files} src files, differing {differing}, binder.ts in {len(hunks)} hunks, "
      f"types.ts {removed} lines changed")
PYTHON

# (1) The writes: writes.cjs on each tree, in parallel.
(NODE_PATH="$modules" step 1-writes-before node "$root/scratch-5pstmvw/writes.cjs" "$before" "$out/1-writes-before.json") &
writesBefore=$!
(NODE_PATH="$modules" step 1-writes-after node "$root/scratch-5pstmvw/writes.cjs" "$after" "$out/1-writes-after.json") &
writesAfter=$!
wait "$writesBefore"
writesBeforeExit=$?
wait "$writesAfter"
writesAfterExit=$?
python3 -I - "$out" "$writesBeforeExit" "$writesAfterExit" 2>> "$out/python.stderr" << 'PYTHON' | tee -a "$out/verdicts.log"
import json, sys
from pathlib import Path
out, exits = Path(sys.argv[1]), dict(zip(('before', 'after'), sys.argv[2:4]))
def place(row):
    return f"{row['file']}:{row['line']}:{row['column']}"
for label in ('before', 'after'):
    path = out / f'1-writes-{label}.json'
    try:
        if exits[label] != '0':
            raise RuntimeError(f'writes.cjs exited {exits[label]}, see {out}/1-writes-{label}.log')
        result = json.loads(path.read_text())
    except Exception as error:
        print(f"flow: verdict (1) writes {label} FAIL: {error}")
        continue
    writes, computed, roads, loose = result['writes'], result['computed'], result['roads'], result['anyReceivers']
    node = [row for row in writes if row['slot'] == 'node']
    antecedent = [row for row in writes if row['slot'] == 'antecedent']
    assigns = [row for row in roads if row['road'] == 'Object.assign']
    print(f"flow: (1) {label}: {result['roots']} roots, {result['files']} src files, {result['accesses']} node or antecedent accesses; "
          f"on a flow receiver node is written {len(node)} times and antecedent {len(antecedent)}; {len(computed)} writes on a flow "
          f"receiver whose key isn't a literal; Object.assign onto a flow node {len(assigns)} of {result['objectAssignCalls']} calls; "
          f"{len(roads) - len(assigns)} other property roads onto a flow node; {len(loose)} node or antecedent writes through an any receiver")
    print(f"flow: (1) {label}: FlowNode's arms {', '.join(result['flowNodeArms'])}")
    for row in writes[:40]:
        print(f"flow: (1)   {label} write {place(row)} {row['slot']} {row['how']} receiver {row['receiver']} (flow arms {', '.join(row['arms'])}): {row['text']}")
    if len(writes) > 40:
        print(f"flow: (1)   {label}: {len(writes) - 40} more writes in {path}")
    for row in computed[:10]:
        print(f"flow: (1)   {label} key not a literal {place(row)} key {row['key']} {row['how']} receiver {row['receiver']}: {row['text']}")
    for row in roads[:10]:
        print(f"flow: (1)   {label} road {place(row)} {row['road']} onto {row['target']}: {row['text']}")
    for row in loose[:10]:
        print(f"flow: (1)   {label} any receiver {place(row)} {row['slot']} {row['how']}: {row['text']}")
    for kind, rows in (('keys not a literal', computed), ('roads', roads), ('any receivers', loose)):
        if len(rows) > 10:
            print(f"flow: (1)   {label}: {len(rows) - 10} more {kind} in {path}")
    if label == 'before':
        # FlowNode as the receiver: printed so, or its arms are FlowNode's own, whatever TypeToString makes of them.
        flowNode = lambda row: row['receiver'] == 'FlowNode' or sorted(row['arms']) == sorted(result['flowNodeArms'])
        problems = []
        if len(node) != 1 or node[0]['file'] != 'src/compiler/binder.ts' or not flowNode(node[0]):
            problems.append(f"node is written {len(node)} times, not once in binder.ts through a FlowNode: "
                            f"{[(place(row), row['receiver']) for row in node[:5]]}")
        stray = [row for row in antecedent if row['arms'] != ['FlowLabel']]
        if stray:
            problems.append(f"{len(stray)} antecedent writes through a receiver other than FlowLabel, first {[(place(row), row['receiver']) for row in stray[:5]]}")
        if assigns:
            problems.append(f"{len(assigns)} Object.assign onto a flow node")
        # A write whose key isn't a literal could reach either slot, so the counts above couldn't be claimed whole.
        if computed:
            problems.append(f"{len(computed)} writes on a flow receiver whose key isn't a literal")
        print(f"flow: verdict (1) writes before {'PASS' if not problems else 'FAIL'}: node written {len(node)} times"
              f"{' (' + place(node[0]) + ', receiver ' + node[0]['receiver'] + ')' if len(node) == 1 else ''}, antecedent {len(antecedent)} times, "
              f"all through FlowLabel: {'yes' if not stray else 'no'}, Object.assign onto a flow node {len(assigns)}"
              f"{'; ' + '; '.join(problems) if problems else ''}")
    else:
        print(f"flow: verdict (1) writes after {'PASS' if not node else 'FAIL'}: node written {len(node)} times, antecedent {len(antecedent)} times")
PYTHON

# (2) Stock tsc: the program 70's audit builds, on each tree, and every syntactic, global and semantic diagnostic.
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
(NODE_PATH="$modules" step 2-diagnostics-before node "$out/diagnostics.cjs" "$before" "$out/2-diagnostics-before.json") &
diagnosticsBefore=$!
(NODE_PATH="$modules" step 2-diagnostics-after node "$out/diagnostics.cjs" "$after" "$out/2-diagnostics-after.json") &
diagnosticsAfter=$!
wait "$diagnosticsBefore"
diagnosticsBeforeExit=$?
wait "$diagnosticsAfter"
diagnosticsAfterExit=$?
if [ "$diagnosticsBeforeExit" != 0 ] || [ "$diagnosticsAfterExit" != 0 ]; then
	say "flow: verdict (2) stock tsc FAIL: the program didn't build (exits $diagnosticsBeforeExit before, $diagnosticsAfterExit after)"
	quote 20 "$out/2-diagnostics-before.log" "$out/2-diagnostics-after.log"
else
	python3 -I - "$out" "$before" "$after" 2>> "$out/python.stderr" << 'PYTHON' | tee -a "$out/verdicts.log"
import json, sys
from collections import Counter
from pathlib import Path
out, beforeTree, afterTree = map(Path, sys.argv[1:4])
sys.dont_write_bytecode = True
sys.path.insert(0, str(out))
import flowlib
before, after = (json.loads((out / f'2-diagnostics-{label}.json').read_text()) for label in ('before', 'after'))
# binder.ts loses lines with the folded statement pair, so a diagnostic below it stands lower before than after: each
# differing file's lines are compared in after's numbering, and a line in a changed run compares as 'changed'.
maps = flowlib.line_maps(beforeTree, afterTree, flowlib.differing(beforeTree, afterTree))
def key(row, side):
    file = row.get('file') or ''
    return (row['code'], file, flowlib.mapped_line(maps, file, row.get('line') or 0, side), row['message'])
old = Counter(key(row, 'before') for row in before['diagnostics'])
new = Counter(key(row, 'after') for row in after['diagnostics'])
added, removed = new - old, old - new
def ts2540(rows):
    return sum(row['code'] == 2540 for row in rows)
print(f"flow: (2) diagnostics: before {len(before['diagnostics'])} over {before['roots']} roots, after {len(after['diagnostics'])} over {after['roots']} roots")
print(f"flow: (2) TS2540: before {ts2540(before['diagnostics'])}, after {ts2540(after['diagnostics'])}")
print(f"flow: (2) diagnostics new in after by code {dict(sorted(Counter(k[0] for k in added.elements()).items()))}, "
      f"gone from after by code {dict(sorted(Counter(k[0] for k in removed.elements()).items()))}")
fresh = sorted((k for k in added.elements() if k[0] == 2540), key=str)
for code, file, line, message in fresh[:40]:
    print(f"flow: (2)   new TS2540 {file}:{line} {flowlib.flat(message, 300)}")
for label, rows in (('new', added), ('gone', removed)):
    listed = sorted((k for k in rows.elements() if not (label == 'new' and k[0] == 2540)), key=str)
    for code, file, line, message in listed[:40]:
        print(f"flow: (2)   {label} TS{code} {file}:{line} {flowlib.flat(message, 300)}")
    if len(listed) > 40:
        print(f"flow: (2)   {len(listed) - 40} more {label} diagnostics in {out}/2-diagnostics-difference.json")
(out / '2-diagnostics-difference.json').write_text(json.dumps({'new': sorted(map(list, added.elements()), key=str),
    'gone': sorted(map(list, removed.elements()), key=str)}, indent=2) + '\n')
ok = not fresh and not added and not removed
print(f"flow: verdict (2) stock tsc {'PASS' if ok else 'FAIL'}: {len(fresh)} TS2540 new in after; "
      f"{sum(added.values())} diagnostics new and {sum(removed.values())} gone, by (code, file, line, message) in after's numbering")
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

# (3) The meter: the census before and after, in parallel.
if [ "$censusBuildExit" != 0 ]; then
	say "flow: verdict (3) meter FAIL: the latent census didn't build"
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
import flowlib
beforeTree, beforeRun, afterTree, afterRun = map(Path, sys.argv[3:7])
codes = sys.argv[7:9]
try:
    before, beforeReasons, beforeSites = flowlib.measure(report, beforeTree, beforeRun)
    after, afterReasons, afterSites = flowlib.measure(report, afterTree, afterRun)
except Exception as error:
    print(f"flow: verdict (3) meter FAIL: a census didn't read whole (exits {codes}): {error}")
    sys.exit(0)
cap = 40
for name, r in (('before', before), ('after', after)):
    print(f"flow: (3) {name}: {r['source_files']} files, latent total {r['latent_total']} {r['totals']}")
    print(f"flow: (3) {name}: the flow family {sum(r['family'].values())} sites over {len(r['family'])} reasons; "
          f"{sum(r['left_out'].values())} sites over {len(r['left_out'])} reasons name a flow interface outside it")
    if r['site_drift']:
        print(f"flow: (3) {name}: the sites read here disagree with per_reason for {len(r['site_drift'])} reasons, first {r['site_drift'][:3]}")
for key in sorted(set(before['family']) | set(after['family']), key=lambda key: (-beforeReasons.get(key, 0), key))[:cap]:
    print(f"flow: (3)   family {beforeReasons.get(key, 0)} to {afterReasons.get(key, 0)}: {flowlib.flat(key)}")
for key in sorted(set(before['left_out']) | set(after['left_out']))[:20]:
    print(f"flow: (3)   left out of the family {beforeReasons.get(key, 0)} to {afterReasons.get(key, 0)}: {flowlib.flat(key)}")
delta = Counter(afterReasons)
delta.subtract(beforeReasons)
moved = sorted(((count, key) for key, count in delta.items() if count), key=lambda item: (-(item[0] > 0), -abs(item[0]), item[1]))
grown = [(count, key) for count, key in moved if count > 0]
print(f"flow: (3) reasons whose count moves: {len(moved)}, {len(grown)} grow, {len(moved) - len(grown)} shrink")
for count, key in moved[:80]:
    print(f"flow: (3)   {'+' if count > 0 else ''}{count} ({beforeReasons.get(key, 0)} to {afterReasons.get(key, 0)}) {flowlib.flat(key)}")
if len(moved) > 80:
    print(f"flow: (3)   {len(moved) - 80} more moving reasons in {out}/3-summary.json")
# Sites, where and text, before then after, in after's numbering, so a site at the same where and text in both is kept,
# one only before is gone, one only after is new: the family's, then every grown reason's, each section capped.
maps = flowlib.line_maps(beforeTree, afterTree, flowlib.differing(beforeTree, afterTree))
listing = []
def section(title, keys):
    printed = 0
    print(f"flow: (3) sites of {title}, {len(keys)} reasons:")
    for key in keys:
        old = [(flowlib.mapped_where(maps, where, 'before'), text) for where, text in beforeSites.get(key, [])]
        new = [(flowlib.mapped_where(maps, where, 'after'), text) for where, text in afterSites.get(key, [])]
        rows = [f"reason {flowlib.flat(key)}: before {len(old)}, after {len(new)}"]
        for label, mine, theirs, other in (('before', old, set(new), 'gone'), ('after', new, set(old), 'new')):
            rows += [f"  {label} {'kept' if site in theirs else other} {site[0]} {flowlib.flat(site[1], 300)}" for site in mine]
        listing.extend([f"{title}: {row}" for row in rows])
        for row in rows:
            if printed < cap:
                print(f"flow: (3)   {row}")
            printed += 1
    if printed > cap:
        print(f"flow: (3)   {printed - cap} more lines in {out}/3-sites.txt")
section('the flow family', sorted(set(before['family']) | set(after['family'])))
section('every reason that grew', [key for _, key in grown])
(out / '3-sites.txt').write_text('\n'.join(listing) + '\n')
family = [sum(r['family'].values()) for r in (before, after)]
drifted = before['site_drift'] or after['site_drift']
ok = family[1] == 0 and not grown and not drifted
print(f"flow: verdict (3) meter {'PASS' if ok else 'FAIL'}: latent total {before['latent_total']} to {after['latent_total']}, "
      f"the flow family {family[0]} to {family[1]}, {len(grown)} reasons grew"
      f"{', and the sites read disagree with the meter' if drifted else ''}")
(out / '3-summary.json').write_text(json.dumps({'before': before, 'after': after,
    'reason_delta': {key: count for key, count in delta.items() if count}}, indent=2) + '\n')
PYTHON
fi

# (4) The mutant: its src differs from after's in FlowAssignment's node alone, readonly there and not here; then the
# census on it, whose family must come back above 0.
if [ "$applyMutantExit" != 0 ]; then
	failed=$(grep -o 'stage3/adapt/[0-9][0-9]-[^/]*/adapt.cjs' "$out/apply-mutant.log" | tail -n 1)
	say "flow: verdict (4) mutant FAIL: the mutant's apply exited $applyMutantExit${failed:+ in $failed}, so there is no mutant tree"
	quote 30 "$out/apply-mutant.log"
elif [ "$censusBuildExit" != 0 ]; then
	say "flow: verdict (4) mutant FAIL: the latent census didn't build"
else
	# The exit is python's, read from its own run, never a pipeline's, whose exit would be tee's.
	python3 -I - "$out" "$after" "$mutant" > "$out/4-mutant-sources.log" 2>> "$out/python.stderr" << 'PYTHON'
import re, sys
from pathlib import Path
out, after, mutant = map(Path, sys.argv[1:4])
sys.dont_write_bytecode = True
sys.path.insert(0, str(out))
import flowlib
differing = flowlib.differing(after, mutant)
afterLines = (after / 'src/compiler/types.ts').read_bytes().decode().split('\n')
mutantLines = (mutant / 'src/compiler/types.ts').read_bytes().decode().split('\n')
changed = [(index + 1, a, b) for index, (a, b) in enumerate(zip(afterLines, mutantLines)) if a != b]
declaration = re.compile(r'^(?:export )?interface (\w+)\b')
owner = next((match.group(1) for text in reversed(afterLines[:changed[0][0]]) if (match := declaration.match(text))), None) if changed else None
slot = 'node: Expression | VariableDeclaration | BindingElement;'
ok = (differing == ['src/compiler/types.ts'] and len(afterLines) == len(mutantLines) and len(changed) == 1 and owner == 'FlowAssignment'
      and changed[0][1].strip() == 'readonly ' + slot and changed[0][2].strip() == slot)
print(f"flow: (4) the mutant's src differs from after's in {differing}, types.ts lines {[(line, owner, b.strip()) for line, _, b in changed[:5]]}: "
      f"{'FlowAssignment.node alone, writable' if ok else 'not the one-slot mutant'}")
sys.exit(0 if ok else 1)
PYTHON
	mutantSourcesExit=$?
	tee -a "$out/verdicts.log" < "$out/4-mutant-sources.log"
	if [ "$mutantSourcesExit" != 0 ]; then
		say "flow: verdict (4) mutant FAIL: the mutant tree isn't after's with FlowAssignment's node writable (exit $mutantSourcesExit), so its census would measure something else"
	else
		census mutant "$mutant"
		censusMutantExit=$?
		python3 -I - "$root/stage3/meter" "$out" "$after" "$work/census-after" "$mutant" "$work/census-mutant" "$censusMutantExit" 2>> "$out/python.stderr" << 'PYTHON' | tee -a "$out/verdicts.log"
import json, sys
from collections import Counter
from pathlib import Path
sys.dont_write_bytecode = True
meter, out = Path(sys.argv[1]), Path(sys.argv[2])
sys.path[:0] = [str(out), str(meter)]
import report
import flowlib
afterTree, afterRun, mutantTree, mutantRun = map(Path, sys.argv[3:7])
try:
    after, afterReasons, _ = flowlib.measure(report, afterTree, afterRun)
    mutant, mutantReasons, mutantSites = flowlib.measure(report, mutantTree, mutantRun)
except Exception as error:
    print(f"flow: verdict (4) mutant FAIL: a census didn't read whole (mutant's exit {sys.argv[7]}): {error}")
    sys.exit(0)
total = sum(mutant['family'].values())
print(f"flow: (4) mutant: {mutant['source_files']} files, latent total {mutant['latent_total']} {mutant['totals']}; the flow family "
      f"{total} sites over {len(mutant['family'])} reasons, where after holds {sum(after['family'].values())}")
for key, count in sorted(mutant['family'].items(), key=lambda item: (-item[1], item[0]))[:40]:
    print(f"flow: (4)   family {count}: {flowlib.flat(key)}")
printed = 0
for key in sorted(mutant['family']):
    for where, text in mutantSites.get(key, []):
        if printed < 40:
            print(f"flow: (4)   site {where} {flowlib.flat(text, 300)}")
        printed += 1
if printed > 40:
    print(f"flow: (4)   {printed - 40} more family sites in {out}/census-mutant.jsonl.gz")
delta = Counter(mutantReasons)
delta.subtract(afterReasons)
moved = sorted(((count, key) for key, count in delta.items() if count), key=lambda item: (-item[0], item[1]))
print(f"flow: (4) reasons whose count moves from after to the mutant: {len(moved)}")
for count, key in moved[:40]:
    print(f"flow: (4)   {'+' if count > 0 else ''}{count} ({afterReasons.get(key, 0)} to {mutantReasons.get(key, 0)}) {flowlib.flat(key)}")
(out / '4-summary.json').write_text(json.dumps({'after': after, 'mutant': mutant,
    'reason_delta': {key: count for key, count in delta.items() if count}}, indent=2) + '\n')
print(f"flow: verdict (4) mutant {'PASS' if total > 0 else 'FAIL'}: with FlowAssignment's node writable the flow family is {total}"
      f"{', back above 0' if total > 0 else ', so the census may not see these slots at all and (3) proves nothing'}")
PYTHON
	fi
fi

# (5) The lane, whole, on the after tree. The gate runs it as shards.json's units, each shard.py against an adapted tree
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
	cat > "$out/5-install-check.py" << 'PYTHON'
import sys
from pathlib import Path
stage = Path(sys.argv[1])
sys.path.insert(0, str(stage))
import apply
apply.installed_packages(stage / 'typescript')
print('stage3/typescript/node_modules is an install of its lockfile')
PYTHON
	python3 -I -B "$out/5-install-check.py" "$root/stage3" > "$out/5-install-check.log" 2>&1
	echo "flow: (5) install check exited $?: $(tail -n 1 "$out/5-install-check.log")"
fi

# buildTree <tree> <label>: hereby local and tests, as the adapted tree product builds, then the link is removed.
buildTree() {
	local tree=$1 label=$2 real code
	real=$(cd "$install" && pwd -P)
	ln -s "$real" "$tree/node_modules" || return 1
	mkdir -p "$work/home-$label"
	HOME="$work/home-$label" XDG_CONFIG_HOME="$work/home-$label" NODE_DISABLE_COMPILE_CACHE=1 \
		bash -c 'cd "$1" && node "$2/hereby/bin/hereby.js" local && node "$2/hereby/bin/hereby.js" tests' _ "$tree" "$real" \
		> "$out/5-build-$label.log" 2>&1
	code=$?
	unlink "$tree/node_modules"
	echo "flow: 5-build-$label exited $code ($out/5-build-$label.log)"
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
    # 74 is internal: no member of the published API changes, so the byte comparisons must pass as the plan expects.
    byte = next((row for row in rows if row['name'] == 'stage3-lane-byte-comparisons'), None)
    byteState = ('was not among the units' if byte is None else 'passed' if byte['exit'] == 0 and byte['status'] == 'pass'
                 else f"failed ({byte['status']}, exit {byte['exit']})")
    ok = not failing and counts == plan['counts'] and byte is not None
    print(f"flow: verdict (5) lane {'PASS' if ok else 'FAIL'}: {len(rows)} units on the {label} tree, {len(failing)} failing, "
          f"counts {counts} where the plan holds {plan['counts']}; 74 is internal, so stage3-lane-byte-comparisons must pass "
          f"with the published typescript.d.ts byte-identical, and it {byteState}")
else:
    print(f"flow: (5) control: {len(rows)} units rerun on the {label} tree, {len(failing)} failing there")
for row in failing[:40]:
    errors = ' | '.join(error.splitlines()[0][:300] for error in row['errors'][:4] if error)
    print(f"flow: (5)   {label} {row['name']} exit {row['exit']} {row['status']}, counts {row['counts']} "
          f"where the manifest expects {row['expect']}: {errors}")
if len(failing) > 40:
    print(f"flow: (5)   {len(failing) - 40} more failing units in {evidence}/failing.txt")
if only is not None:
    for name in sorted(set(only) - {row['name'] for row in failing}):
        print(f"flow: (5)   {label} {name} passes")
PYTHON

if [ -z "$installReason" ]; then
	# Half the threads, from one to eight: each unit runs one upstream worker.
	concurrency=$(( $(nproc) / 2 ))
	[ "$concurrency" -ge 1 ] || concurrency=1
	[ "$concurrency" -le 8 ] || concurrency=8
	concurrency=${FLOW_LANE_CONCURRENCY:-$concurrency}
	# shard.py's own limit is 30 s, the gate's timing contract; this run measures verdicts, so units get more room.
	seconds=${FLOW_SHARD_SECONDS:-120}
	echo "flow: (5) the lane's product path: shards.json's units, $concurrency at a time, $seconds s each"
	if ! buildTree "$after" after; then
		say "flow: verdict (5) lane FAIL: the after tree didn't build with hereby; see $out/5-build-after.log"
		quote 30 "$out/5-build-after.log"
	else
		python3 -I "$out/lane.py" "$root" "$after" "$work/lane-after" "$out/5-lane-after" "$concurrency" "$seconds" after \
			> "$out/5-lane-after.log" 2>&1
		echo "flow: 5-lane-after exited $? ($out/5-lane-after.log)"
		grep '^flow: ' "$out/5-lane-after.log" | tee -a "$out/verdicts.log"
		grep -q '^flow: verdict (5) lane ' "$out/5-lane-after.log" ||
			say "flow: verdict (5) lane FAIL: the units' driver printed no verdict; see $out/5-lane-after.log"
		failing=$(cat "$out/5-lane-after/failing.txt" 2> /dev/null)
		# A control: the failing units again on the before tree, so a red reads as the slots' or the box's.
		if [ -n "$failing" ]; then
			if buildTree "$before" before; then
				python3 -I "$out/lane.py" "$root" "$before" "$work/lane-before" "$out/5-lane-before" "$concurrency" "$seconds" before "$failing" \
					> "$out/5-lane-before.log" 2>&1
				grep '^flow: ' "$out/5-lane-before.log" | tee -a "$out/verdicts.log"
			else
				echo "flow: (5) control: the before tree didn't build; see $out/5-build-before.log"
			fi
		fi
	fi
elif [ "$hasNpm" = 1 ]; then
	platform=$(node -p 'process.platform + "/" + process.arch + "/" + process.version')
	if ! python3 -I -c 'import json, sys; sys.exit(0 if sys.argv[2] in json.load(open(sys.argv[1]))["platforms"] else 1)' "$root/stage3/lane/expected.json" "$platform"; then
		say "flow: verdict (5) lane SKIP: $installReason, and run.sh's expected.json holds no counts for $platform"
	elif ! npm ping > "$out/5-npm-ping.log" 2>&1; then
		say "flow: verdict (5) lane SKIP: $installReason, and run.sh's npm ci can't reach the registry (npm ping failed)"
	else
		# run.sh applies this tree's own stage3 again; the store holds after's product under the same key, so it's restored.
		(cd "$root" && STAGE3_CACHE="$work/cache-after" step 5-lane-run bash stage3/lane/run.sh "$work/lane-run")
		for file in verdict.txt verdict.json report.json execution.json failures.md; do
			cp "$work/lane-run/$file" "$out/5-lane-run-$file" 2> /dev/null
		done
		verdict=$(cat "$work/lane-run/verdict.txt" 2> /dev/null)
		# run.sh's check.py compares the published API whole, so its PASS holds typescript.d.ts byte-identical too.
		case "$verdict" in
			PASS*) say "flow: verdict (5) lane PASS: run.sh: $verdict; 74 is internal, and check.py found the published typescript.d.ts as sanctioned" ;;
			*) say "flow: verdict (5) lane FAIL: run.sh: ${verdict:-no verdict.txt}; 74 is internal, so an API failure in $out/5-lane-run-failures.md is the slots'" ;;
		esac
	fi
else
	say "flow: verdict (5) lane SKIP: $installReason, and there is no npm for run.sh's npm ci; lay stage3/typescript/node_modules (the pinned devDependencies) to run the product path"
fi

# (6) The plan: test_check.py as the lane's Go wrapper runs it, stage3/lane/python_test.go's TestLaneCheckPython through
# runPythonCommand: python3 -E -S -B -m unittest test_check from stage3/lane, with a fresh STAGE3_CACHE and
# PYTHONDONTWRITEBYTECODE=1. The wrapper bounds it at 55 s; this gives it 300 where timeout exists. test_check reads the
# committed sanctioned-api.json and expected.json; shard-plan.json's runtime hash is read beside it with shard_plan's
# own runtime_hash, since 74 must move neither.
mkdir -p "$work/cache-test-check"
bound=()
command -v timeout > /dev/null 2>&1 && bound=(timeout 300)
(cd "$root/stage3/lane" && STAGE3_CACHE="$work/cache-test-check" step 6-test-check ${bound[@]+"${bound[@]}"} python3 -E -S -B -m unittest test_check)
testCheckExit=$?
python3 -I -B - "$root/stage3/lane" "$out/6-test-check.log" "$testCheckExit" 2>> "$out/python.stderr" << 'PYTHON' | tee -a "$out/verdicts.log"
import json, sys
from pathlib import Path
lane, log, code = Path(sys.argv[1]), Path(sys.argv[2]), sys.argv[3]
sys.dont_write_bytecode = True
sys.path.insert(0, str(lane))
tail = [line for line in log.read_text(errors='replace').splitlines() if line.startswith(('Ran ', 'OK', 'FAILED'))]
for line in log.read_text(errors='replace').splitlines():
    if line.startswith(('FAIL:', 'ERROR:')):
        print(f"flow: (6)   {line[:300]}")
print(f"flow: (6) test_check exited {code}: {' '.join(tail) or 'no unittest summary'}")
try:
    import shard_plan
    committed = json.loads((lane / 'shard-plan.json').read_text())['runtime']
    actual = shard_plan.runtime_hash()
    runtime = committed == actual
    print(f"flow: (6) shard-plan.json's runtime {'holds' if runtime else 'is stale'}: committed {committed[:16]}, the lane's files hash to {actual[:16]}")
except Exception as error:
    runtime = False
    print(f"flow: (6) shard-plan.json's runtime didn't read: {error}")
ok = code == '0' and runtime
print(f"flow: verdict (6) plan {'PASS' if ok else 'FAIL'}: test_check {'passed' if code == '0' else 'failed, exit ' + code}, "
      f"shard-plan.json's runtime {'holds' if runtime else 'does not hold'}")
PYTHON
# git status with no optional locks, so reading it writes nothing into a tree a sealed job lays read-only.
if git -C "$root" rev-parse --is-inside-work-tree > /dev/null 2>&1; then
	GIT_OPTIONAL_LOCKS=0 git -C "$root" status --porcelain -- stage3/lane > "$out/6-lane-status.txt" 2>&1
	statusExit=$?
	if [ "$statusExit" = 0 ] && [ ! -s "$out/6-lane-status.txt" ]; then
		say "flow: verdict (6) lane clean PASS: git status --porcelain stage3/lane is empty after everything"
	else
		say "flow: verdict (6) lane clean FAIL: git status exited $statusExit, $(wc -l < "$out/6-lane-status.txt") lines under stage3/lane"
		quote 20 "$out/6-lane-status.txt"
	fi
else
	say "flow: verdict (6) lane clean SKIP: $root is no git checkout, so git status can't read stage3/lane"
fi

# A part whose reader crashed printed no verdict; that is a failure, never a silent pass.
for part in 'apply' '(0) sources' '(1) writes before' '(1) writes after' '(2) stock tsc' '(3) meter' '(4) mutant' '(5) lane' '(6) plan' '(6) lane clean'; do
	grep -qF "flow: verdict $part " "$out/verdicts.log" ||
		say "flow: verdict $part FAIL: this part printed no verdict; the readers' errors are in $out/python.stderr"
done
if [ -s "$out/python.stderr" ]; then
	echo "flow: the readers wrote to stderr, last lines:"
	quote 20 "$out/python.stderr"
fi

echo "flow: summary, $SECONDS s:"
grep '^flow: verdict ' "$out/verdicts.log"
grep -q '^flow: verdict .* FAIL:' "$out/verdicts.log" && exit 1
grep -q '^flow: verdict .* SKIP:' "$out/verdicts.log" && exit 3
exit 0
