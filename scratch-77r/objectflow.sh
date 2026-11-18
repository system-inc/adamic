#!/usr/bin/env bash
# objectflow.sh: a scratch measurement for #38kdd4d, never to land. Adaptation 77, stage3/adapt/77-readonly-object-slots,
# is on main with its creator (#5pstmvw, #2gf0k9w). #38kdd4d changes it in resolveAnonymousTypeMembers alone: the call
# `setStructuredTypeMembers(type, members, emptyArray, emptyArray, indexInfos || emptyArray);` keeps its result in a new
# `const resolved`, and the writes of callSignatures and constructSignatures after it, which 77 wrote through a type
# assertion to ObjectTypeUnderConstruction, go through `resolved`. setStructuredTypeMembers casts its parameter to
# ResolvedType and returns it, so `resolved` is the type itself. The JavaScript gains one const and two receivers,
# value-identical. This measures what it changes. Run from an adamic tree's root, on a sealed job as
#
#   source /opt/adamic-block/environment.sh && source cloud/sealed-job.sh && export GOFLAGS=... && sealed_typescript_source &&
#   sealed_npm stage3/api stage3/typescript && ... && bash scratch-77r/objectflow.sh
#
# never with sealed_git in the chain: a sealed tree already has a .git, sealed_git refuses it, and the && chain stops.
# Or on a house box after a lay line that lays cohere, cohere/TypeScript and both npm installs, and ~/adamic-tools/env.sh.
#
#   before: stage3/apply.sh over a copy of this tree's stage3 inputs with 77's adapt.cjs as main holds it
#           (scratch-77r/main-adapt.cjs, main 0efb2d04's), main's tree
#   after:  this tree's own stage3/apply.sh, main's tree with #38kdd4d
#   mutant: the after tree with resolveAnonymousTypeMembers back in 77's asserted form, which is main's form
#   witness: the after tree plus one line, `Debug.assert(resolved === type, "setStructuredTypeMembers returns its
#           parameter");` right after the const, inserted here into a copy and never by the adaptation
#
#   (0) sources: diff -u of checker.ts, before to after; PASS is checker.ts alone differing under src/, and checker.ts
#       before's with exactly three lines changed, all in resolveAnonymousTypeMembers: the call now kept in the const,
#       and the two writes now through `resolved`; each hunk's contents listed
#   (1) writes: scratch-77r/ow.cjs, every write to one of the five slots on an ObjectType receiver found by the type
#       checker, with the interface whose declaration of the slot the checker resolves on each receiver arm; PASS is 24
#       writes on both trees, no key that isn't a literal, no any receiver, no Object.assign onto an ObjectType, none
#       on plain ObjectType, each slot declared writable where it is written; before 18 through the view and 6 through
#       ResolvedType, after 16 and 8, the two anonymous writes through the view before and through ResolvedType after,
#       and the same 24 writes on both (`(x as ObjectTypeUnderConstruction)` and the anonymous `resolved` read as the
#       bare receiver)
#   (1b) calls: by AST over every src file, each call of createObjectType and of createObjectTypeUnderConstruction with
#       the function it stands in; PASS is the same calls on both trees, the creator called from createObjectType and
#       the two writers alone
#   (2) stock tsc: one ts.createProgram per tree over all upstream src, 70's options, stock TypeScript 6.0.3; PASS is no
#       new TS2540 and every diagnostic (code, file, line, message) the same before and after, each tree's root read as <tree>
#   (3) the meter: the latent census on both src/compiler trees, read by report.py's latent_summary; PASS is the object
#       family at 0 after, "AnonymousType seen as ObjectTypeUnderConstruction" gone (2 sites before, 0 after), no reason
#       grown, and setStructuredTypeMembers' own sites, its cast among them, the same before and after; each tree's
#       root reads as <tree> in every reason and every site's text before they're compared
#   (4) the mutant: the census on the mutant tree; PASS is its src the same as before's, and "AnonymousType seen as
#       ObjectTypeUnderConstruction" back at 2 sites, both in resolveAnonymousTypeMembers
#   (5) the lane, whole, on the witness tree: every unit of stage3/lane/shards.json run by shard.py against the witness
#       tree built by hereby with stage3/typescript's install, as the gate's product path runs it, so Node checks the
#       identity at the const wherever the lane reaches it; SKIP with the reason when stage3/typescript's install isn't
#       laid, since run.sh applies its own tree and can't run the witness. 77 is internal, so
#       stage3-lane-byte-comparisons must pass: the published typescript.d.ts is byte-identical
#   (6) the plan: stage3/lane/test_check.py as the lane's own Go wrapper runs it, shard-plan.json's runtime hash, and
#       git status of stage3/lane after everything, which must be empty
#
# Every line printed begins "flow:", verdict lines "flow: verdict". Logs, JSON and the summary go to
# $TMPDIR/objectflow-77r-<stamp>; long lists stay there, and at most about 40 sites print per section, since a job's
# output is cut near 150 KB. The heavy trees go to $TMPDIR too (FLOW_WORK names another place), never beside the tree,
# which a sealed job lays read-only, and are removed at the end unless FLOW_KEEP=1. Exit 0 when every part passes, 1
# when any fails, 3 when none fails and one is skipped; 64 to 66 are preparation failures.
set -u

root=$PWD
adaptation=77-readonly-object-slots
[ -f "$root/stage3/apply.sh" ] && [ -f "$root/stage3/adapt/$adaptation/adapt.cjs" ] && [ -f "$root/scratch-77r/ow.cjs" ] &&
	[ -f "$root/scratch-77r/main-adapt.cjs" ] || {
	echo "flow: $root is not an adamic tree's root carrying stage3/adapt/$adaptation, scratch-77r/ow.cjs and scratch-77r/main-adapt.cjs, so nothing ran"
	exit 64
}
stamp=$(date -u +%Y%m%dT%H%M%SZ)
scratch="${TMPDIR:-/tmp}"
out="${scratch%/}/objectflow-77r-$stamp"
work="${FLOW_WORK:-${scratch%/}/77r-$stamp.flow}"
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

# Before's stage3: this tree's apply inputs whole, with 77's adapt.cjs as main holds it, scratch-77r/main-adapt.cjs,
# which is main 0efb2d04's: so before is main's tree, and before's adapt differs from this tree's in that file alone.
beforeStage="$work/before-inputs/stage3"
stageInputs "$work/before-inputs" "" || exit 64
cp "$root/scratch-77r/main-adapt.cjs" "$beforeStage/adapt/$adaptation/adapt.cjs" || exit 64
diff -rq "$root/stage3/adapt" "$beforeStage/adapt" > "$out/before-inputs.diff" 2>&1
echo "flow: before's adapt differs from this tree's by: $(tr '\n' ' ' < "$out/before-inputs.diff")"
echo "flow: main's 77 adapt.cjs sha256 $(sha256sum "$root/scratch-77r/main-adapt.cjs" 2> /dev/null | cut -c1-16), this tree's $(sha256sum "$root/stage3/adapt/$adaptation/adapt.cjs" 2> /dev/null | cut -c1-16)"
[ "$(wc -l < "$out/before-inputs.diff")" = 1 ] && grep -q "$adaptation/adapt.cjs" "$out/before-inputs.diff" ||
	{ echo "flow: before's inputs differ from this tree's by more than 77's adapt.cjs, so nothing ran"; exit 64; }
echo "flow: product keys: after $(cd "$root" && python3 stage3/apply.py --key), before $(python3 "$beforeStage/apply.py" --key)"

# Phase 1, in parallel: the two applies from a fresh local product store, and the census tool. A local store makes
# each apply build its tree from the pinned upstream here; nothing is fetched as a product.
before="$work/before"
after="$work/after"
mutant="$work/mutant"
witness="$work/witness"
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
echo "flow: apply before exited $applyBeforeExit, apply after exited $applyAfterExit, census build exited $censusBuildExit"
[ "$applyBeforeExit" = 0 ] || { echo "flow: before's apply failed, so there is no before tree; its last lines:"; quote 30 "$out/apply-before.log"; exit 66; }
if [ "$applyAfterExit" != 0 ]; then
	# apply.py names the adaptation it was running in the failed command.
	failed=$(grep -o 'stage3/adapt/[0-9][0-9]-[^/]*/adapt.cjs' "$out/apply-after.log" | tail -n 1)
	say "flow: verdict apply FAIL: the after tree's apply exited $applyAfterExit${failed:+ in $failed}, so there is no after tree"
	quote 30 "$out/apply-after.log"
	for part in '(0) sources' '(1) writes' '(1b) calls' '(2) stock tsc' '(3) meter' '(4) mutant' '(5) lane' '(6) plan'; do
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

# The ObjectType family: ObjectType, the interfaces adaptation 77 reads or redeclares the five slots on, and the view
# it adds. configure() adds every interface ow.cjs found extending one of them, on either tree, so an extender
# (TupleType, MappedType, FreshObjectLiteralType) is a family arm by the program's own heritage, never by a name list.
CORE = ['ObjectType', 'ResolvedType', 'InterfaceType', 'AnonymousType', 'GenericType', 'TypeReference', 'ObjectTypeUnderConstruction']
FAMILY = set(CORE)
MENTION = None
# The five slots' types as ResolvedType reads them; ObjectType's are each of these or undefined.
NARROW = frozenset(['SymbolTable', 'Symbol[]', 'readonly Signature[]', 'readonly IndexInfo[]'])


def configure(out):
    # The family's extenders from ow.cjs's output on both trees, read when it is there; CORE alone otherwise.
    global MENTION
    for label in ('before', 'after'):
        path = Path(out) / f'1-writes-{label}.json'
        try:
            FAMILY.update(json.loads(path.read_text()).get('familyInterfaces') or [])
        except (OSError, ValueError):
            pass
    MENTION = re.compile(r'\b(?:' + '|'.join(sorted(FAMILY, key=len, reverse=True)) + r')\b')
    return sorted(FAMILY)

# The family. A per_reason key is "<kind>: <reason>", and a reason is the lowering's What with its types printed by
# TypeToString (stage3/census/latent/lower.go.txt sets Reason from Refused.What and NotYet.What). A refusal the five
# slots cause takes one of these shapes:
#   internal/lower/invariance.go, wideningRefusal: "a value of type A seen as B, which can write W where R is read";
#   internal/lower/union_writes.go, unionSlotWrite: "a write through a union of V into a slot that holds S in M".
# The "can write" shape names types, never the field, so a match on a family name beside a slot name finds none of the
# refusals adaptation 77 answers. So the family is parsed: a view whose A or B holds a family interface as a whole
# union or intersection arm, and whose W is exactly its R or undefined, R one of the five slots' narrow types
# (SymbolTable | undefined written where SymbolTable is read); or a union write whose M is a family interface, whose S
# is one of the narrow types and whose V holds undefined. Other shapes over the family, "whose readonly field F becomes
# writable" among them (an ObjectType seen as a ResolvedType), are another fact and stay out; every reason naming a
# family interface that the family leaves out is printed with its count by left_out, so the line drawn here can be
# read and argued with.
VIEW = re.compile(r'a value of type (?P<own>.+?) seen as (?P<view>.+?), which can write (?P<written>.+?) where (?P<read>.+?) is read(?:$|[,.;:])')
UNION_WRITE = re.compile(r'a write through a union of (?P<written>.+) into a slot that holds (?P<held>.+) in (?P<member>.+)')


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


def widened(written, read):
    # W is R or undefined, R one of the five slots' narrow types.
    readArms = arms(read)
    return len(readArms) == 1 and readArms[0] in NARROW and set(arms(written)) == {readArms[0], 'undefined'}


def family(key):
    if MENTION is None:
        raise RuntimeError('flowlib.configure(out) must run before family()')
    text = reason(key)
    if (match := VIEW.match(text)):
        return any(arm in FAMILY for arm in arms(match['own']) + arms(match['view'])) and widened(match['written'], match['read'])
    if (match := UNION_WRITE.fullmatch(text)):
        held = arms(match['held'])
        return match['member'].strip() in FAMILY and len(held) == 1 and held[0] in NARROW and 'undefined' in arms(match['written'])
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
    # A census site's where, src/compiler/checker.ts:5552:9, in after's numbering.
    parts = where.rsplit(':', 2)
    if len(parts) != 3 or not parts[1].isdigit():
        return where
    line = mapped_line(maps, parts[0], int(parts[1]), side)
    return f"{parts[0]}:{line}:{parts[2] if isinstance(line, int) else '?'}"


def rooted(mapping, tree):
    # A reason, and a site's text, that names a declaration prints its tree's root (read at <tree>/src/...), so each
    # tree's root reads as <tree>, resolved or as given, before reasons or sites are compared; otherwise the same refusal
    # counts as one gone and one grown, and a kept site as one gone and one new. A count adds up under the rooted key; a
    # list of (where, text) sites joins, each text rooted too.
    roots = sorted({str(tree.resolve()), str(tree)}, key=len, reverse=True)
    def root(text):
        for spelling in roots:
            text = text.replace(spelling, '<tree>')
        # A site's text opens with its own location, <tree>/file:line:column:, in its own tree's numbering, which moves
        # with any line the adaptation adds above it; the site's where, mapped to after's numbering, already says where.
        return re.sub(r'^<tree>/[^ :]+:\d+:\d+: ', '', text)
    result = type(mapping)()
    for key, value in mapping.items():
        if isinstance(value, list):
            result[root(key)] = result.get(root(key), []) + [(where, root(text)) for where, text in value]
        else:
            result[root(key)] = result.get(root(key), 0) + value
    return result


def enclosing(tree, where):
    # The unit a site stands in: the nearest function declared at or above its line, read from its own tree's source,
    # as file:name, or file:<top>.
    file, line = where.rsplit(':', 2)[0], where.rsplit(':', 2)[1]
    if not line.isdigit():
        return None
    text = (tree / file).read_text().split('\n')
    for index in range(int(line) - 1, -1, -1):
        found = re.match(r'\s*(?:export\s+)?(?:async\s+)?function\s*\*?\s*(\w+)\s*[<(]', text[index])
        if found:
            return f"{file}:{found.group(1)}"
    return f"{file}:<top>"


def units(tree, sites):
    # Each unit's refusal sites, counted over every reason.
    counts = Counter()
    for key, found in sites.items():
        for where, _ in found:
            unit = enclosing(tree, where)
            if unit:
                counts[unit] += 1
    return counts


SEEN = re.compile(r'a value of type (?P<own>.+?) seen as (?P<view>.+?),')


def seen_as(key):
    # A view reason's own type and the type it is seen as, or None for another shape.
    match = SEEN.match(reason(key))
    return (match['own'], match['view']) if match else None


def function_lines(lines, name):
    # The first and last index of the one function declared by that name in a file's lines, carriage returns stripped:
    # its declaration line and the closing brace at the declaration's own indentation. A search scoped to it can't meet
    # the same text elsewhere in the file (setStructuredTypeMembers writes `resolved.constructSignatures = ...` too).
    starts = [index for index, line in enumerate(lines) if re.match(r'\s*function\s+' + re.escape(name) + r'\s*[<(]', line)]
    if len(starts) != 1:
        raise ValueError(f"expected one function {name}, found {len(starts)}")
    start = starts[0]
    indent = lines[start][:len(lines[start]) - len(lines[start].lstrip())]
    end = next((index for index in range(start + 1, len(lines)) if lines[index] == indent + '}'), None)
    if end is None:
        raise ValueError(f"function {name} at line {start + 1} has no closing brace at its indentation")
    return start, end


def flat(text, limit=400):
    line = ' '.join(text.split())
    return line if len(line) <= limit else f"{line[:limit]} ... ({len(line)} characters)"
PYTHON

# (0) The sources: before and after differ under src/ in checker.ts alone, and there in three lines, all in
# resolveAnonymousTypeMembers: the call kept in the const and the two writes through it.
diff -u "$before/src/compiler/checker.ts" "$after/src/compiler/checker.ts" > "$out/0-checker.diff" 2>&1
python3 -I - "$before" "$after" "$out" 2>> "$out/python.stderr" << 'PYTHON' | tee -a "$out/verdicts.log"
import difflib, json, re, sys
from collections import Counter
from pathlib import Path
before, after, out = map(Path, sys.argv[1:4])
sys.dont_write_bytecode = True
sys.path.insert(0, str(out))
import flowlib
differing = flowlib.differing(before, after)
files = sum(1 for p in (after / 'src').rglob('*') if p.is_file())
SLOTS = ['members', 'properties', 'callSignatures', 'constructSignatures', 'indexInfos']
VIEW = 'ObjectTypeUnderConstruction'
declaration = re.compile(r'^(?:export )?interface (\w+)\b')
member = re.compile(r'^\s*(?:/\*\* @internal \*/ )?(readonly )?(\w+)\??:')

def lines(tree, name):
    # Each line without its carriage return, so a CRLF file compares as its text.
    return [line.rstrip('\r') for line in (tree / name).read_bytes().decode().split('\n')]

def owner(source, line):
    for text in reversed(source[:line]):
        if (match := declaration.match(text)):
            return match.group(1)
    return None

def inserted(a, b):
    # b is a with `readonly ` inserted once, wherever it stands.
    index = next((i for i, (x, y) in enumerate(zip(a, b)) if x != y), len(a))
    return b[index:index + 9] == 'readonly ' and b[:index] + b[index + 9:] == a

def opcodes(old, new):
    return [op for op in difflib.SequenceMatcher(None, old, new, autojunk=False).get_opcodes() if op[0] != 'equal']

problems = []

# checker.ts.
checkerDiff = (out / '0-checker.diff').read_text().rstrip('\n').split('\n')
hunks = [index for index, line in enumerate(checkerDiff) if line.startswith('@@')]
body = checkerDiff[hunks[0]:] if hunks else []
print(f"flow: (0) checker.ts: {len(hunks)} hunks ({out / '0-checker.diff'}):")
for line in body[:40]:
    print(f"flow: (0)   | {line.rstrip(chr(13))}")
if len(body) > 40:
    print(f"flow: (0)   ... {len(body) - 40} more lines in {out / '0-checker.diff'}")
oldChecker, newChecker = lines(before, 'src/compiler/checker.ts'), lines(after, 'src/compiler/checker.ts')
importLine = f'import type {{ {VIEW} }} from "./types.js";'
# The proof is a reconstruction: before's lines with each expected change applied, each found exactly once, must be
# after's lines exactly. Each change keeps the name and the after lines it stands on, for the hunk listing.
substitutions = [
    ("resolveAnonymousTypeMembers' call, kept in the const",
     ['        setStructuredTypeMembers(type, members, emptyArray, emptyArray, indexInfos || emptyArray);'],
     ['        const resolved = setStructuredTypeMembers(type, members, emptyArray, emptyArray, indexInfos || emptyArray);']),
    ("resolveAnonymousTypeMembers' callSignatures", [f'            (type as {VIEW}).callSignatures = getSignaturesOfSymbol(symbol);'],
     ['            resolved.callSignatures = getSignaturesOfSymbol(symbol);']),
    ("resolveAnonymousTypeMembers' constructSignatures", [f'            (type as {VIEW}).constructSignatures = constructSignatures;'],
     ['            resolved.constructSignatures = constructSignatures;']),
]
def occurrences(haystack, needle):
    return [index for index in range(len(haystack) - len(needle) + 1) if haystack[index:index + len(needle)] == needle]
strays = []
rebuilt = list(oldChecker)
# Each search runs inside resolveAnonymousTypeMembers alone, in each tree: the call's text and setStructuredTypeMembers'
# own `resolved.` writes stand elsewhere in checker.ts too. Every substitution keeps its line count, so before's range
# holds in the rebuilt lines.
UNIT = 'resolveAnonymousTypeMembers'
try:
    beforeRange, afterRange = flowlib.function_lines(oldChecker, UNIT), flowlib.function_lines(newChecker, UNIT)
except ValueError as error:
    beforeRange = afterRange = None
    strays.append(str(error))
def scoped(haystack, needle, bounds):
    start, end = bounds
    return [start + index for index in occurrences(haystack[start:end + 1], needle)]
for name, old, new in substitutions:
    if beforeRange is None:
        break
    found = scoped(rebuilt, old, beforeRange)
    if len(found) != 1:
        strays.append(f"{name}: before's text found {len(found)} times in {UNIT}, not once")
        continue
    rebuilt[found[0]:found[0] + len(old)] = new
rebuiltOk = not strays and rebuilt == newChecker
if not strays and not rebuiltOk:
    first = next((index for index, (a, b) in enumerate(zip(rebuilt, newChecker)) if a != b), min(len(rebuilt), len(newChecker)))
    strays.append(f"after's checker.ts differs from before's with the expected changes at line {first + 1}: "
                  f"expected {rebuilt[first:first + 1]}, found {newChecker[first:first + 1]}")
# Each change's after lines, found once in after, placed in every hunk whose after range they overlap: diff -u can split
# one change across hunks where its unchanged lines run longer than its context.
changeLines = [(name, new) for name, _, new in substitutions]
ranges = []
for index in hunks:
    found = re.match(r'@@ -\d+(?:,\d+)? \+(\d+)(?:,(\d+))? @@', checkerDiff[index])
    start, count = int(found.group(1)), int(found.group(2) or 1)
    ranges.append((start, start + count))
placed = {}
for name, new in changeLines:
    found = scoped(newChecker, new, afterRange) if afterRange else []
    if len(found) != 1:
        placed.setdefault(None, []).append(name)
        continue
    first, last = found[0] + 1, found[0] + len(new)
    overlapping = [k for k, (start, end) in enumerate(ranges) if start <= last and first < end]
    for hunk in overlapping or [None]:
        placed.setdefault(hunk, []).append(name)
for hunk, names in sorted(placed.items(), key=lambda item: (item[0] is None, item[0] or 0)):
    print(f"flow: (0) checker.ts hunk {hunk + 1 if hunk is not None else '? (in no hunk)'}: {', '.join(names)}")
seen = list(dict.fromkeys(name for names in placed.values() for name in names))
hunksOk = None not in placed and sorted(placed) == list(range(len(hunks)))
# Every change stands in resolveAnonymousTypeMembers, read from after's own source.
units = sorted({flowlib.enclosing(after, f"src/compiler/checker.ts:{index + 1}:1") for name, new in changeLines
                for index in (scoped(newChecker, new, afterRange) if afterRange else [])})
# The rebuild compares whole files, so a change outside the function's range would still fail it.
print(f"flow: (0) checker.ts: the changes stand in {', '.join(units)}")
unitsOk = units == ['src/compiler/checker.ts:resolveAnonymousTypeMembers']
checkerOk = rebuiltOk and hunksOk and unitsOk
if not checkerOk:
    problems.append(f"checker.ts: {len(strays)} changes not expected, first {strays[:3]}; every change in a hunk: "
                    f"{'yes' if None not in placed else 'no'}; every hunk holding a change: {'yes' if sorted(k for k in placed if k is not None) == list(range(len(hunks))) else 'no'}")
json.dump({'files': files, 'differing': differing, 'checker_hunks': len(hunks), 'units': units, 'problems': problems},
          open(out / '0-sources.json', 'w'), indent=2)
for problem in problems[:20]:
    print(f"flow: (0)   problem: {flowlib.flat(problem, 300)}")
ok = differing == ['src/compiler/checker.ts'] and not problems
print(f"flow: verdict (0) sources {'PASS' if ok else 'FAIL'}: {files} src files, differing {differing}, checker.ts in {len(hunks)} hunks "
      f"holding {len(seen)} changes (3 expected, each listed above by hunk: the call kept in the const and the two writes "
      f"through it), all in resolveAnonymousTypeMembers: {'yes' if unitsOk else 'no'}, after's checker.ts before's with "
      f"exactly those: {'yes' if rebuiltOk else 'no'}, {len(problems)} problems")
PYTHON

# (1) The writes: ow.cjs on each tree, in parallel.
(NODE_PATH="$modules" step 1-writes-before node "$root/scratch-77r/ow.cjs" "$before" "$out/1-writes-before.json") &
writesBefore=$!
(NODE_PATH="$modules" step 1-writes-after node "$root/scratch-77r/ow.cjs" "$after" "$out/1-writes-after.json") &
writesAfter=$!
wait "$writesBefore"
writesBeforeExit=$?
wait "$writesAfter"
writesAfterExit=$?
python3 -I - "$out" "$writesBeforeExit" "$writesAfterExit" 2>> "$out/python.stderr" << 'PYTHON' | tee -a "$out/verdicts.log"
import json, re, sys
from collections import Counter
from pathlib import Path
out, exits = Path(sys.argv[1]), dict(zip(('before', 'after'), sys.argv[2:4]))
# Where each slot is written writable after 77: the interface whose own declaration the checker resolves on the arm.
WRITABLE = {
    'members': {'ResolvedType', 'ObjectTypeUnderConstruction'},
    'properties': {'ResolvedType', 'ObjectTypeUnderConstruction'},
    'callSignatures': {'ResolvedType', 'ObjectTypeUnderConstruction'},
    'constructSignatures': {'ResolvedType', 'ObjectTypeUnderConstruction'},
    'indexInfos': {'ResolvedType', 'ObjectTypeUnderConstruction'},
}
# AnonymousType and InterfaceType stay as upstream declares them, their slots readonly, so no write goes through either.
ARMS = {'ResolvedType', 'ObjectTypeUnderConstruction'}
# Each tree's writes by the family arm they go through. Before, main's: createObjectTypeUnderConstruction's five, its two
# writers' ten, the reset and resolveAnonymousTypeMembers' two through the view, and setStructuredTypeMembers' six.
# After, the two anonymous writes go through setStructuredTypeMembers' result, a ResolvedType.
EXPECTED = {'before': Counter({'ObjectTypeUnderConstruction': 18, 'ResolvedType': 6}),
            'after': Counter({'ObjectTypeUnderConstruction': 16, 'ResolvedType': 8})}
ANONYMOUS = {'before': 'ObjectTypeUnderConstruction', 'after': 'ResolvedType'}

def place(row):
    return f"{row['file']}:{row['line']}:{row['column']}"

def family(arm):
    # An arm's family name: its own, or for "Name (extends Family)" the family interface it extends.
    found = re.fullmatch(r'.+ \(extends (\w+)\)', arm)
    return found.group(1) if found else arm

def owners(row):
    return ', '.join(f"{o['arm']} via {o['declaredBy']}{' readonly' if o['readonly'] else ''}" for o in row.get('owners') or [])

results = {}
for label in ('before', 'after'):
    path = out / f'1-writes-{label}.json'
    try:
        if exits[label] != '0':
            raise RuntimeError(f'ow.cjs exited {exits[label]}, see {out}/1-writes-{label}.log')
        results[label] = json.loads(path.read_text())
    except Exception as error:
        print(f"flow: verdict (1) writes {label} FAIL: {error}")
for label, result in results.items():
    writes, computed, roads, loose = result['writes'], result['computed'], result['roads'], result['anyReceivers']
    assigns = [row for row in roads if row['road'] == 'Object.assign']
    print(f"flow: (1) {label}: {result['roots']} roots, {result['files']} src files, {result['accesses']} slot accesses; "
          f"{len(writes)} slot writes on an ObjectType receiver; {len(computed)} writes on an ObjectType receiver whose key isn't "
          f"a literal; Object.assign onto an ObjectType {len(assigns)} of {result['objectAssignCalls']} calls; "
          f"{len(roads) - len(assigns)} other property roads onto an ObjectType; {len(loose)} slot writes through an any receiver")
    print(f"flow: (1) {label}: the family's interfaces in the program: {', '.join(result.get('familyInterfaces') or [])}")
    table = Counter((row['function'], ', '.join(row['arms'])) for row in writes)
    for (function, arms), count in sorted(table.items()):
        print(f"flow: (1)   {label} {count} in {function} through {arms}")
    for row in writes[:40]:
        print(f"flow: (1)   {label} write {place(row)} {row['slot']} {row['how']} in {row['function']} receiver {row['receiver']} "
              f"({owners(row)}): {row['text']}")
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
    problems = []
    if len(writes) != 24:
        problems.append(f"{len(writes)} writes, not 24")
    # A write whose key isn't a literal could reach any slot, so the counts above couldn't be claimed whole.
    if computed:
        problems.append(f"{len(computed)} writes on an ObjectType receiver whose key isn't a literal")
    if loose:
        problems.append(f"{len(loose)} slot writes through an any receiver")
    if assigns:
        problems.append(f"{len(assigns)} Object.assign onto an ObjectType")
    plain = [row for row in writes if 'ObjectType' in row['arms']]
    if plain:
        problems.append(f"{len(plain)} writes on plain ObjectType, first {[(place(row), row['slot']) for row in plain[:5]]}")
    stray = []
    for row in writes:
        rows = row.get('owners') or []
        if not rows or any(family(o['arm']) not in ARMS or o['declaredBy'] not in WRITABLE[row['slot']] or o['readonly'] is not False for o in rows):
            stray.append(row)
    if stray:
        problems.append(f"{len(stray)} writes whose arm isn't one that keeps the slot writable, first "
                        f"{[(place(row), row['slot'], owners(row)) for row in stray[:5]]}")
    arms = Counter(family(o['arm']) for row in writes for o in row.get('owners') or [])
    if arms != EXPECTED[label]:
        problems.append(f"writes by arm {dict(sorted(arms.items()))}, expected {dict(sorted(EXPECTED[label].items()))}")
    anonymous = sorted((row['slot'], ', '.join(family(o['arm']) for o in row.get('owners') or [])) for row in writes
                       if row['function'] == 'resolveAnonymousTypeMembers')
    if anonymous != [('callSignatures', ANONYMOUS[label]), ('constructSignatures', ANONYMOUS[label])]:
        problems.append(f"resolveAnonymousTypeMembers' writes {anonymous}, expected both through {ANONYMOUS[label]}")
    same = None
    if label == 'after':
        if 'before' in results:
            # A receiver asserted to the view, and resolveAnonymousTypeMembers' `resolved`, read as the bare receiver, so
            # the same write compares as itself.
            def key(row):
                text = re.sub(r'\((\w+) as ObjectTypeUnderConstruction\)', r'\1', row['text'])
                if row['function'] == 'resolveAnonymousTypeMembers':
                    text = re.sub(r'^resolved\.', 'type.', text)
                return (row['function'], row['slot'], text)
            old, new = Counter(map(key, results['before']['writes'])), Counter(map(key, writes))
            same = old == new
            if not same:
                problems.append(f"the writes differ from before's by (function, slot, text): gone {sorted((old - new).elements())[:5]}, "
                                f"new {sorted((new - old).elements())[:5]}")
        else:
            problems.append("before's writes didn't read, so after's can't be compared with them")
    print(f"flow: verdict (1) writes {label} {'PASS' if not problems else 'FAIL'}: {len(writes)} writes through {dict(sorted(arms.items()))}, "
          f"resolveAnonymousTypeMembers' two through {', '.join(sorted({arm for _, arm in anonymous})) or 'nothing'}, "
          f"each slot declared writable there: {'yes' if not stray else 'no'}; on plain ObjectType {len(plain)}, keys not a literal "
          f"{len(computed)}, any receivers {len(loose)}, Object.assign onto an ObjectType {len(assigns)}"
          f"{'' if same is None else ', the same 24 as before: ' + ('yes' if same else 'no')}"
          f"{'; ' + '; '.join(problems) if problems else ''}")
PYTHON

# (1b) The calls: every call of createObjectType and of createObjectTypeUnderConstruction under each tree's src, by AST,
# with the function each stands in, and every other reference to the creator.
cat > "$out/calls.cjs" << 'NODE'
'use strict';
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
const tree = path.resolve(process.argv[2]);
const output = path.resolve(process.argv[3]);
const names = ['createObjectType', 'createObjectTypeUnderConstruction'];
const calls = [];
const references = [];
for (const relative of fs.readdirSync(path.join(tree, 'src'), { recursive: true }).map(String).filter(name => name.endsWith('.ts')).sort()) {
    const fileName = path.join(tree, 'src', relative);
    const text = fs.readFileSync(fileName, 'utf8');
    if (!text.includes('createObjectType')) continue;
    const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    ts.forEachChild(source, function find(node) {
        if (ts.isIdentifier(node) && names.includes(node.text)) {
            let owner = node.parent;
            while (owner && !((ts.isFunctionDeclaration(owner) || ts.isMethodDeclaration(owner) || ts.isFunctionExpression(owner)) && owner.name && owner.name !== node)) owner = owner.parent;
            const at = source.getLineAndCharacterOfPosition(node.getStart(source));
            const row = { file: path.join('src', relative), line: at.line + 1, name: node.text, function: owner?.name?.getText(source) ?? '<top>' };
            if (ts.isCallExpression(node.parent) && node.parent.expression === node) calls.push(row);
            else if (!(ts.isFunctionDeclaration(node.parent) && node.parent.name === node)) references.push(row);
        }
        ts.forEachChild(node, find);
    });
}
fs.writeFileSync(output, JSON.stringify({ calls, references }, null, 2) + '\n');
console.log(`${calls.length} calls, ${references.length} other references`);
NODE
(NODE_PATH="$modules" step 1b-calls-before node "$out/calls.cjs" "$before" "$out/1b-calls-before.json") &
callsBefore=$!
(NODE_PATH="$modules" step 1b-calls-after node "$out/calls.cjs" "$after" "$out/1b-calls-after.json") &
callsAfter=$!
wait "$callsBefore"
callsBeforeExit=$?
wait "$callsAfter"
callsAfterExit=$?
python3 -I - "$out" "$callsBeforeExit" "$callsAfterExit" 2>> "$out/python.stderr" << 'PYTHON' | tee -a "$out/verdicts.log"
import json, sys
from collections import Counter
from pathlib import Path
out, exits = Path(sys.argv[1]), sys.argv[2:4]
if exits != ['0', '0']:
    print(f"flow: verdict (1b) calls FAIL: calls.cjs exited {exits[0]} before and {exits[1]} after, see {out}/1b-calls-*.log")
    sys.exit(0)
before, after = (json.loads((out / f'1b-calls-{label}.json').read_text()) for label in ('before', 'after'))
CREATOR = 'createObjectTypeUnderConstruction'
WRITERS = ['getOrCreateTypeFromSignature', 'getTypeWithoutSignatures']
def by(result, name):
    return [row for row in result['calls'] if row['name'] == name]
for label, result in (('before', before), ('after', after)):
    for name in ('createObjectType', CREATOR):
        rows = by(result, name)
        print(f"flow: (1b) {label}: {len(rows)} calls of {name}")
        for row in rows:
            print(f"flow: (1b)   {label} {name} called in {row['function']} at {row['file']}:{row['line']}")
    for row in result['references']:
        print(f"flow: (1b)   {label} other reference to {row['name']} in {row['function']} at {row['file']}:{row['line']}")
problems = []
signature = lambda result: Counter((row['name'], row['function']) for row in result['calls'])
if signature(before) != signature(after):
    problems.append(f"the calls differ: gone {sorted((signature(before) - signature(after)).elements())}, new {sorted((signature(after) - signature(before)).elements())}")
for label, result in (('before', before), ('after', after)):
    callers = sorted(row['function'] for row in by(result, CREATOR))
    if callers != sorted(['createObjectType'] + WRITERS):
        problems.append(f"{label} calls {CREATOR} from {callers}, not createObjectType and the two writers alone")
    strays = [row for row in result['references'] if row['name'] == CREATOR]
    if strays:
        problems.append(f"{len(strays)} other references to {CREATOR} {label}")
print(f"flow: verdict (1b) calls {'PASS' if not problems else 'FAIL'}: createObjectType called {len(by(before, 'createObjectType'))} times "
      f"before and {len(by(after, 'createObjectType'))} after, {CREATOR} {len(by(before, CREATOR))} and {len(by(after, CREATOR))}, "
      f"the same calls on both trees: {'yes' if signature(before) == signature(after) else 'no'}"
      f"{'; ' + '; '.join(problems) if problems else ''}")
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
# Each differing file's lines are compared in after's numbering, and a line in a changed run compares as 'changed'; #38kdd4d
# changes three lines in place, so no line moves.
maps = flowlib.line_maps(beforeTree, afterTree, flowlib.differing(beforeTree, afterTree))
# A message that names a declaration prints its tree's root (import("<tree>/src/...")), so each side's root reads as
# <tree> before the two are compared: the same diagnostic from the same source says the same thing on both sides.
roots = {'before': str(beforeTree.resolve()), 'after': str(afterTree.resolve())}
def key(row, side):
    file = row.get('file') or ''
    message = row['message'].replace(roots[side], '<tree>')
    return (row['code'], file, flowlib.mapped_line(maps, file, row.get('line') or 0, side), message)
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
print(f"flow: (3) the family's interfaces: {', '.join(flowlib.configure(out))}")
beforeTree, beforeRun, afterTree, afterRun = map(Path, sys.argv[3:7])
codes = sys.argv[7:9]
try:
    before, beforeReasons, beforeSites = flowlib.measure(report, beforeTree, beforeRun)
    after, afterReasons, afterSites = flowlib.measure(report, afterTree, afterRun)
except Exception as error:
    print(f"flow: verdict (3) meter FAIL: a census didn't read whole (exits {codes}): {error}")
    sys.exit(0)
# Each tree's root reads as <tree>, in every reason and in every site's text (flowlib.rooted says why).
rooted = flowlib.rooted
beforeReasons, beforeSites = rooted(beforeReasons, beforeTree), rooted(beforeSites, beforeTree)
afterReasons, afterSites = rooted(afterReasons, afterTree), rooted(afterSites, afterTree)
for r, tree in ((before, beforeTree), (after, afterTree)):
    r['family'], r['left_out'] = rooted(r['family'], tree), rooted(r['left_out'], tree)
cap = 40
for name, r in (('before', before), ('after', after)):
    print(f"flow: (3) {name}: {r['source_files']} files, latent total {r['latent_total']} {r['totals']}")
    print(f"flow: (3) {name}: the object family {sum(r['family'].values())} sites over {len(r['family'])} reasons; "
          f"{sum(r['left_out'].values())} sites over {len(r['left_out'])} reasons name a family interface outside it")
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
        # A kept site goes to the listing alone: printed, it fills the cap before a grown reason's new sites show.
        for row in rows:
            if ' kept ' in row:
                continue
            if printed < cap:
                print(f"flow: (3)   {row}")
            printed += 1
    if printed > cap:
        print(f"flow: (3)   {printed - cap} more lines in {out}/3-sites.txt")
section('the object family', sorted(set(before['family']) | set(after['family'])))
section('every reason that grew', [key for _, key in grown])
(out / '3-sites.txt').write_text('\n'.join(listing) + '\n')
# Compiler's question on #5pstmvw: is a grown site created by the edit, or uncovered in a unit that lowered whole before?
# Each new site of a grown reason is placed in its enclosing function, read from its own tree's source; a function that
# held no refusal of any reason before had lowered whole, and a new site there is a regression.
beforeUnits = flowlib.units(beforeTree, beforeSites)
afterUnits = flowlib.units(afterTree, afterSites)
for _, key in grown:
    for where, text in afterSites.get(key, []):
        if any(flowlib.mapped_where(maps, w, 'before') == flowlib.mapped_where(maps, where, 'after') and t == text for w, t in beforeSites.get(key, [])):
            continue
        unit = flowlib.enclosing(afterTree, where)
        held = beforeUnits.get(unit, 0)
        print(f"flow: (3) unit of a grown site {where}: {unit}, {held} refusals there before: "
              f"{'a move, the unit never lowered whole' if held else 'a regression, the unit lowered whole before'}")
# #38kdd4d's measure. The stop "AnonymousType seen as ObjectTypeUnderConstruction", 77's two asserted writes, goes; and
# setStructuredTypeMembers' own sites, its `as ResolvedType` among them, are the same before and after, each where it was
# and saying what it said: the change reads that cast's result and mustn't move what the census says at it.
STOP = ('AnonymousType', 'ObjectTypeUnderConstruction')
stops = [sum(count for key, count in reasons.items() if flowlib.seen_as(key) == STOP) for reasons in (beforeReasons, afterReasons)]
print(f"flow: (3) AnonymousType seen as ObjectTypeUnderConstruction: {stops[0]} sites before, {stops[1]} after")
for label, tree, sites in (('before', beforeTree, beforeSites), ('after', afterTree, afterSites)):
    for key in sorted(key for key in sites if flowlib.seen_as(key) == STOP):
        for where, text in sites[key]:
            print(f"flow: (3)   {label} {where} in {flowlib.enclosing(tree, where)}: {flowlib.flat(text, 300)}")
STRUCTURED = 'src/compiler/checker.ts:setStructuredTypeMembers'
def unitSites(tree, sites, side):
    return sorted((flowlib.mapped_where(maps, where, side), key, text) for key, found in sites.items() for where, text in found
                  if flowlib.enclosing(tree, where) == STRUCTURED)
structured = [unitSites(beforeTree, beforeSites, 'before'), unitSites(afterTree, afterSites, 'after')]
for label, rows in zip(('before', 'after'), structured):
    print(f"flow: (3) {STRUCTURED} {label}: {len(rows)} sites")
    for where, key, text in rows[:20]:
        print(f"flow: (3)   {label} {where}: {flowlib.flat(key, 200)}: {flowlib.flat(text, 200)}")
structuredSame = structured[0] == structured[1]
for name in ['resolveAnonymousTypeMembers', 'setStructuredTypeMembers']:
    unit = f'src/compiler/checker.ts:{name}'
    print(f"flow: (3) unit {unit}: {beforeUnits.get(unit, 0)} refusals before, {afterUnits.get(unit, 0)} after")
family = [sum(r['family'].values()) for r in (before, after)]
drifted = before['site_drift'] or after['site_drift']
# The family goes to 0 after. A reason that grew isn't failed here: each is listed above with its sites, before and
# after, for judgment, and the verdict says how many grew.
ok = family[1] == 0 and not drifted and stops == [2, 0] and not grown and structuredSame
print(f"flow: verdict (3) meter {'PASS' if ok else 'FAIL'}: latent total {before['latent_total']} to {after['latent_total']}, "
      f"the object family {family[0]} to {family[1]}, AnonymousType seen as ObjectTypeUnderConstruction {stops[0]} to {stops[1]} "
      f"(2 to 0 expected), {len(grown)} reasons grew (0 expected), setStructuredTypeMembers' sites the same before and after: "
      f"{'yes' if structuredSame else 'no'}"
      f"{', and the sites read disagree with the meter' if drifted else ''}")
(out / '3-summary.json').write_text(json.dumps({'before': before, 'after': after,
    'reason_delta': {key: count for key, count in delta.items() if count}}, indent=2) + '\n')
PYTHON
fi

# (4) The mutant: the after tree with resolveAnonymousTypeMembers back in 77's asserted form, the call without its const
# and both writes through (type as ObjectTypeUnderConstruction), and nothing else; so its src must be before's, byte for
# byte, and its census must show the two stops back. The tree is after's linked whole (cp -al, or a copy where links
# can't be made), and checker.ts is written as a new file in its place, so after's own checker.ts is never touched.
if [ "$censusBuildExit" != 0 ]; then
	say "flow: verdict (4) mutant FAIL: the latent census didn't build"
elif ! { cp -al "$after" "$mutant" 2> /dev/null || { rm -rf -- "$mutant"; cp -a "$after" "$mutant"; }; }; then
	say "flow: verdict (4) mutant FAIL: after's tree didn't copy to $mutant"
else
	chmod u+w "$mutant/src/compiler"
	python3 -I - "$out" "$before" "$after" "$mutant" > "$out/4-mutant-sources.log" 2>> "$out/python.stderr" << 'PYTHON'
import os, sys
from pathlib import Path
out, before, after, mutant = map(Path, sys.argv[1:5])
sys.dont_write_bytecode = True
sys.path.insert(0, str(out))
import flowlib
VIEW = 'ObjectTypeUnderConstruction'
restore = [('const resolved = setStructuredTypeMembers(type, members, emptyArray, emptyArray, indexInfos || emptyArray);',
            'setStructuredTypeMembers(type, members, emptyArray, emptyArray, indexInfos || emptyArray);'),
           ('resolved.callSignatures = getSignaturesOfSymbol(symbol);', f'(type as {VIEW}).callSignatures = getSignaturesOfSymbol(symbol);'),
           ('resolved.constructSignatures = constructSignatures;', f'(type as {VIEW}).constructSignatures = constructSignatures;')]
text = (after / 'src/compiler/checker.ts').read_bytes().decode()
newline = '\r\n' if '\r\n' in text else '\n'
lines = text.split(newline)
# Each line is found inside resolveAnonymousTypeMembers alone: setStructuredTypeMembers writes
# `resolved.constructSignatures = constructSignatures;` too.
try:
    start, end = flowlib.function_lines(lines, 'resolveAnonymousTypeMembers')
except ValueError as error:
    print(f"flow: (4) {error}; no mutant")
    sys.exit(1)
for adapted, asserted in restore:
    found = [index for index in range(start, end + 1) if lines[index].strip() == adapted]
    if len(found) != 1:
        print(f"flow: (4) resolveAnonymousTypeMembers holds `{adapted}` {len(found)} times, not once; no mutant")
        sys.exit(1)
    line = lines[found[0]]
    lines[found[0]] = line[:len(line) - len(line.lstrip())] + asserted
restored = newline.join(lines)
# A new file in the link's place, so after's checker.ts keeps its bytes.
(mutant / 'src/compiler/checker.ts.mutant').write_bytes(restored.encode())
os.replace(mutant / 'src/compiler/checker.ts.mutant', mutant / 'src/compiler/checker.ts')
untouched = (after / 'src/compiler/checker.ts').read_bytes().decode() == text
fromAfter, fromBefore = flowlib.differing(after, mutant), flowlib.differing(before, mutant)
ok = fromAfter == ['src/compiler/checker.ts'] and fromBefore == [] and untouched
print(f"flow: (4) the mutant's src differs from after's in {fromAfter} and from before's in {fromBefore or 'nothing'}, after's own "
      f"checker.ts untouched: {'yes' if untouched else 'no'}: {'the asserted form restored, and nothing else' if ok else 'not the mutant'}")
sys.exit(0 if ok else 1)
PYTHON
	mutantSourcesExit=$?
	tee -a "$out/verdicts.log" < "$out/4-mutant-sources.log"
	if [ "$mutantSourcesExit" != 0 ]; then
		say "flow: verdict (4) mutant FAIL: the mutant tree isn't after's with 77's asserted form restored (exit $mutantSourcesExit), so its census would measure something else"
	else
		census mutant "$mutant"
		censusMutantExit=$?
		python3 -I - "$root/stage3/meter" "$out" "$after" "$work/census-after" "$mutant" "$work/census-mutant" "$censusMutantExit" 2>> "$out/python.stderr" << 'PYTHON' | tee -a "$out/verdicts.log"
import sys
from collections import Counter
from pathlib import Path
sys.dont_write_bytecode = True
meter, out = Path(sys.argv[1]), Path(sys.argv[2])
sys.path[:0] = [str(out), str(meter)]
import report
import flowlib
flowlib.configure(out)
afterTree, afterRun, mutantTree, mutantRun = map(Path, sys.argv[3:7])
try:
    after, afterReasons, afterSites = flowlib.measure(report, afterTree, afterRun)
    mutant, mutantReasons, mutantSites = flowlib.measure(report, mutantTree, mutantRun)
except Exception as error:
    print(f"flow: verdict (4) mutant FAIL: a census didn't read whole (the mutant's exit {sys.argv[7]}): {error}")
    sys.exit(0)
afterReasons, afterSites = flowlib.rooted(afterReasons, afterTree), flowlib.rooted(afterSites, afterTree)
mutantReasons, mutantSites = flowlib.rooted(mutantReasons, mutantTree), flowlib.rooted(mutantSites, mutantTree)
STOP = ('AnonymousType', 'ObjectTypeUnderConstruction')
UNIT = 'src/compiler/checker.ts:resolveAnonymousTypeMembers'
stops = [(where, text) for key, found in mutantSites.items() if flowlib.seen_as(key) == STOP for where, text in found]
inUnit = [where for where, _ in stops if flowlib.enclosing(mutantTree, where) == UNIT]
print(f"flow: (4) mutant: latent total {mutant['latent_total']}, where after holds {after['latent_total']}; AnonymousType seen as "
      f"ObjectTypeUnderConstruction {len(stops)} sites, {len(inUnit)} of them in resolveAnonymousTypeMembers")
for where, text in stops:
    print(f"flow: (4)   site {where} in {flowlib.enclosing(mutantTree, where)}: {flowlib.flat(text, 300)}")
delta = Counter(mutantReasons)
delta.subtract(afterReasons)
for key, count in sorted(((key, count) for key, count in delta.items() if count), key=lambda item: (-item[1], item[0]))[:20]:
    print(f"flow: (4)   {'+' if count > 0 else ''}{count} ({afterReasons.get(key, 0)} to {mutantReasons.get(key, 0)}) {flowlib.flat(key)}")
ok = len(stops) == 2 and len(inUnit) == 2
print(f"flow: verdict (4) mutant {'PASS' if ok else 'FAIL'}: with 77's asserted form restored, AnonymousType seen as "
      f"ObjectTypeUnderConstruction is back at {len(stops)} sites, {len(inUnit)} in resolveAnonymousTypeMembers"
      f"{', so (3) sees the two go' if ok else ''}")
PYTHON
	fi
fi

# (5) The lane, whole, on the witness tree. The gate runs it as shards.json's units, each shard.py against an adapted tree
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
    # 77 is internal: no member of the published API changes, so the byte comparisons must pass as the plan expects.
    byte = next((row for row in rows if row['name'] == 'stage3-lane-byte-comparisons'), None)
    byteState = ('was not among the units' if byte is None else 'passed' if byte['exit'] == 0 and byte['status'] == 'pass'
                 else f"failed ({byte['status']}, exit {byte['exit']})")
    ok = not failing and counts == plan['counts'] and byte is not None
    print(f"flow: verdict (5) lane {'PASS' if ok else 'FAIL'}: {len(rows)} units on the {label} tree, {len(failing)} failing, "
          f"counts {counts} where the plan holds {plan['counts']}; 77 is internal, so stage3-lane-byte-comparisons must pass "
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
	echo "flow: (5) the lane's product path: shards.json's units, $concurrency at a time, $seconds s each, on the witness tree"
	# The witness: a whole copy of after, never linked, since hereby builds into it, with one line inserted after the
	# const, which the adaptation never writes. The check is that the witness differs from after in that line alone.
	cp -a "$after" "$witness" &&
		python3 -I - "$out" "$after" "$witness" > "$out/5-witness.log" 2>> "$out/python.stderr" << 'PYTHON'
import sys
from pathlib import Path
out, after, witness = map(Path, sys.argv[1:4])
sys.dont_write_bytecode = True
sys.path.insert(0, str(out))
import flowlib
const = 'const resolved = setStructuredTypeMembers(type, members, emptyArray, emptyArray, indexInfos || emptyArray);'
assertion = 'Debug.assert(resolved === type, "setStructuredTypeMembers returns its parameter");'
text = (after / 'src/compiler/checker.ts').read_bytes().decode()
newline = '\r\n' if '\r\n' in text else '\n'
lines = text.split(newline)
try:
    start, end = flowlib.function_lines(lines, 'resolveAnonymousTypeMembers')
except ValueError as error:
    print(f"flow: (5) {error}; no witness")
    sys.exit(1)
found = [index for index in range(start, end + 1) if lines[index].strip() == const]
if len(found) != 1:
    print(f"flow: (5) resolveAnonymousTypeMembers holds the const {len(found)} times, not once; no witness")
    sys.exit(1)
indent = lines[found[0]][:len(lines[found[0]]) - len(lines[found[0]].lstrip())]
lines.insert(found[0] + 1, indent + assertion)
(witness / 'src/compiler/checker.ts').write_bytes(newline.join(lines).encode())
afterLines = text.split(newline)
witnessLines = (witness / 'src/compiler/checker.ts').read_bytes().decode().split(newline)
ok = (flowlib.differing(after, witness) == ['src/compiler/checker.ts'] and len(witnessLines) == len(afterLines) + 1 and
      witnessLines[:found[0] + 1] + witnessLines[found[0] + 2:] == afterLines and witnessLines[found[0] + 1].strip() == assertion)
print(f"flow: (5) the witness tree is the after tree plus one line, checker.ts:{found[0] + 2}, `{assertion}`, right after the "
      f"const: {'yes' if ok else 'no'}")
sys.exit(0 if ok else 1)
PYTHON
	witnessExit=$?
	cat "$out/5-witness.log" 2> /dev/null | tee -a "$out/verdicts.log"
	if [ "$witnessExit" != 0 ]; then
		say "flow: verdict (5) lane FAIL: the witness tree isn't after plus the one assert (exit $witnessExit); see $out/5-witness.log"
	elif ! buildTree "$witness" witness; then
		say "flow: verdict (5) lane FAIL: the witness tree, after plus the one assert, didn't build with hereby; see $out/5-build-witness.log"
		quote 30 "$out/5-build-witness.log"
	else
		python3 -I "$out/lane.py" "$root" "$witness" "$work/lane-after" "$out/5-lane-after" "$concurrency" "$seconds" \
			"witness (after plus Debug.assert(resolved === type) right after the const)" > "$out/5-lane-after.log" 2>&1
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
else
	# run.sh applies this tree's own stage3 and builds what it applied, so it can't run the witness tree.
	say "flow: verdict (5) lane SKIP: $installReason, and the witness tree, after plus the one assert, runs only on the product path; lay stage3/typescript/node_modules (the pinned devDependencies) to run it"
fi

# (6) The plan: test_check.py as the lane's Go wrapper runs it, stage3/lane/python_test.go's TestLaneCheckPython through
# runPythonCommand: python3 -E -S -B -m unittest test_check from stage3/lane, with a fresh STAGE3_CACHE and
# PYTHONDONTWRITEBYTECODE=1. The wrapper bounds it at 55 s; this gives it 300 where timeout exists. test_check reads the
# committed sanctioned-api.json and expected.json; shard-plan.json's runtime hash is read beside it with shard_plan's
# own runtime_hash, since #38kdd4d must move neither.
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
for part in 'apply' '(0) sources' '(1) writes before' '(1) writes after' '(1b) calls' '(2) stock tsc' '(3) meter' '(4) mutant' '(5) lane' '(6) plan' '(6) lane clean'; do
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
