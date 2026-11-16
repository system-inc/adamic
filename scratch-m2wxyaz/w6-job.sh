#!/usr/bin/env bash
# w6-job.sh: the evidence for adaptation 70's public DiagnosticRelatedInformation.file target (#m2wxyaz, refusal
# family 4), run from an adamic tree's root on a house box, the tree's adapt.cjs carrying the target. The caller lays
# cohere and cohere/TypeScript; ~/adamic-tools/env.sh gives go, node, npm and clang.
#
#   a. the audit: adaptation 70's adapt.cjs on the exact tree apply hands it, built fresh from TypeScript 6.0.3
#   b. the published API: built/local/typescript.d.ts before and after, as a plain diff and through check-api.cjs
#   c. the meter's family on a latent census of the adapted tree, before and after, with the total
#   d. the mutant: the same apply with only the new target dropped, which is the before tree of c
#
# Before is the mutant: a copy of this tree's stage3 inputs with the one target line removed, so the two applies
# differ by that line alone. If the audit refuses, apply stops at 70, there is no after tree, b is skipped, and c
# measures an unaudited after: the before tree with only `readonly ` inserted at the slot through the stock AST,
# labeled so, to show what the edit would do to the family and the total.
#
# Logs, JSON and the summary go to /tmp/w6-m2wxyaz-<stamp>; the heavy trees go beside the worktree and are removed
# at the end unless W6_KEEP=1. Exit 0 only when a, b, c and d all pass; 64 to 66 are preparation failures.
set -u

root=$PWD
[ -f "$root/stage3/apply.sh" ] && [ -f "$root/stage3/adapt/70-readonly-views/adapt.cjs" ] || {
	echo "w6: $root is not an adamic tree's root, so nothing ran"
	exit 64
}
target="    { file: 'src/compiler/types.ts', declaration: 'DiagnosticRelatedInformation', member: 'file', public: true },"
[ "$(grep -cxF -- "$target" "$root/stage3/adapt/70-readonly-views/adapt.cjs")" = 1 ] || {
	echo "w6: adapt.cjs doesn't carry the DiagnosticRelatedInformation.file target once, so nothing ran"
	exit 64
}
stamp=$(date -u +%Y%m%dT%H%M%SZ)
out=/tmp/w6-m2wxyaz-$stamp
work="${W6_WORK:-$root.w6}"
mkdir -p "$out" || exit 64
mkdir "$work" || { echo "w6: $work is there already or can't be made, so nothing ran"; exit 64; }
jobEnd() {
	local code=$?
	trap - EXIT
	if [ "${W6_KEEP:-0}" != 1 ]; then
		case "$work" in /*/*.w6) chmod -R u+w -- "$work" 2> /dev/null; rm -rf -- "$work" ;; esac
	fi
	echo "w6: evidence in $out"
	exit "$code"
}
trap jobEnd EXIT
trap 'exit 143' TERM
trap 'exit 129' HUP
trap 'exit 130' INT
mkdir "$work/tmp" || exit 64
export TMPDIR="$work/tmp"

. "$HOME/adamic-tools/env.sh" || { echo "w6: $HOME/adamic-tools/env.sh didn't source, so nothing ran"; exit 65; }
missing=""
for tool in go gofmt node python3 git; do
	command -v "$tool" > /dev/null 2>&1 || missing="$missing $tool"
done
[ -z "$missing" ] || { echo "w6: missing from the toolchain:$missing"; exit 65; }
hasNpm=0; command -v npm > /dev/null 2>&1 && hasNpm=1
[ "$hasNpm" = 1 ] || echo "w6: no npm on this box, so the pristine and b builds are skipped and stage3/api uses the laid node_modules"
node --version | grep -q '^v24\.' || { echo "w6: node is $(node --version), adaptations need Node 24"; exit 65; }
[ -n "$(ls -A "$root/cohere" 2> /dev/null)" ] || { echo "w6: cohere isn't laid in $root"; exit 66; }
echo "w6: $(uname -n), $(nproc) threads, $(git -C "$root" rev-parse HEAD), node $(node --version), $(go version)"

# step <name> <command...>: the command's whole output in $out/<name>.log, and its exit.
step() {
	local name=$1
	shift
	"$@" > "$out/$name.log" 2>&1
	local code=$?
	echo "w6: $name exited $code ($out/$name.log)"
	return "$code"
}

# say <line>: a verdict line, printed and kept in $out/verdicts.log for the summary and the exit.
say() { echo "$*" | tee -a "$out/verdicts.log"; }

# The pinned upstream checkout apply reads, cloned as the meter clones it when the caller supplies none.
pin() { python3 -I -c 'import json, sys; print(json.load(open(sys.argv[1]))[sys.argv[2]])' "$root/stage3/source.json" "$1"; }
pinCommit=$(pin commit)
if [ -z "${ADAMIC_TYPESCRIPT_SOURCE:-}" ] || [ "$(git -C "$ADAMIC_TYPESCRIPT_SOURCE" rev-parse HEAD 2> /dev/null)" != "$pinCommit" ]; then
	step upstream-clone git clone --depth 1 --branch "$(pin tag)" "$(pin repository)" "$work/upstream" || exit 66
	export ADAMIC_TYPESCRIPT_SOURCE="$work/upstream"
fi
[ "$(git -C "$ADAMIC_TYPESCRIPT_SOURCE" rev-parse HEAD)" = "$pinCommit" ] || { echo "w6: upstream isn't at $pinCommit"; exit 66; }

# stage3/api's lockfile-pinned typescript@6.0.3, which apply puts on NODE_PATH for every adaptation.
if [ "$hasNpm" = 1 ]; then (cd "$root/stage3/api" && step api-install npm ci --ignore-scripts --no-audit --no-fund) || exit 66; fi
modules="$root/stage3/api/node_modules"
[ "$(node -p 'require(process.argv[1] + "/typescript/package.json").version' "$modules")" = 6.0.3 ] || { echo "w6: stage3/api's typescript isn't 6.0.3"; exit 66; }

# The mutant's stage3: this tree's apply inputs with the one target line dropped, and nothing else changed.
mutant="$work/mutant/stage3"
mkdir -p "$mutant" || exit 64
for input in apply.sh apply.py caches.py source.json api adapt; do
	cp -a "$root/stage3/$input" "$mutant/" || exit 64
done
python3 -I - "$mutant/adapt/70-readonly-views/adapt.cjs" "$target" << 'PYTHON' || exit 64
import sys
path, line = sys.argv[1], sys.argv[2] + '\n'
text = open(path).read()
if text.count(line) != 1:
    sys.exit('the target line is not in the copy exactly once')
open(path, 'w').write(text.replace(line, '', 1))
PYTHON
diff "$root/stage3/adapt/70-readonly-views/adapt.cjs" "$mutant/adapt/70-readonly-views/adapt.cjs" > "$out/d-mutant-adapt.diff"
diff -rq --exclude=adapt.cjs "$root/stage3/adapt/70-readonly-views" "$mutant/adapt/70-readonly-views" > "$out/d-mutant-other-files.diff" 2>&1
echo "w6: d mutant adapt.cjs differs by $(grep -c '^<' "$out/d-mutant-adapt.diff") line(s), other 70 files differing: $(wc -l < "$out/d-mutant-other-files.diff")"
echo "w6: product keys: with target $(cd "$root" && python3 stage3/apply.py --key), mutant $(python3 "$mutant/apply.py" --key)"

# Phase 1, in parallel: both applies from a fresh local product store, the pristine control, and the census tool.
# A local store makes each apply build its tree from the pinned upstream here; nothing is fetched as a product.
before="$work/before"
after="$work/after"
pristine="$work/pristine"
export STAGE3_PRODUCT_STORE="$work/products"
mkdir -p "$STAGE3_PRODUCT_STORE"
(STAGE3_CACHE="$work/cache-mutant" step apply-before bash "$mutant/apply.sh" "$before") &
applyBefore=$!
(cd "$root" && STAGE3_CACHE="$work/cache" step apply-after bash stage3/apply.sh "$after") &
applyAfter=$!
if [ "$hasNpm" = 1 ]; then (step pristine-clone git clone "$ADAMIC_TYPESCRIPT_SOURCE" "$pristine" &&
	step pristine-checkout git -C "$pristine" checkout --detach "$pinCommit" &&
	cd "$pristine" && step pristine-install npm ci --no-audit --no-fund && step pristine-build npm run build) &
pristineBuild=$!; else (exit 0) & pristineBuild=$!; fi
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
[ "$applyBeforeExit" = 0 ] || { echo "w6: the mutant's apply failed, so there is no before tree; its last lines:"; tail -n 30 "$out/apply-before.log"; exit 66; }
[ "$censusBuildExit" = 0 ] || { echo "w6: the latent census didn't build; its last lines:"; tail -n 30 "$out/census-build.log"; exit 66; }

# a. The audit, run alone on the exact tree apply handed adaptation 70 (its refs/stage3 snapshot), with the target
# and, for d, without it. Adaptations 00 to 65 are the same in both applies, so either tree's snapshot serves.
snapshotFrom=$before
git -C "$after" rev-parse -q --verify refs/stage3/70-readonly-views-before > /dev/null 2>&1 && snapshotFrom=$after
audit70="$work/audit70"
auditMutant="$work/audit70-mutant"
mkdir -p "$audit70" "$auditMutant"
git -C "$snapshotFrom" archive refs/stage3/70-readonly-views-before | tar -x -C "$audit70" || exit 66
git -C "$snapshotFrom" archive refs/stage3/70-readonly-views-before | tar -x -C "$auditMutant" || exit 66
NODE_PATH="$modules" node "$root/stage3/adapt/70-readonly-views/adapt.cjs" "$audit70" > "$out/a-audit.json" 2> "$out/a-audit.log"
auditExit=$?
NODE_PATH="$modules" node "$mutant/adapt/70-readonly-views/adapt.cjs" "$auditMutant" > "$out/d-mutant-audit.json" 2> "$out/d-mutant-audit.log"
auditMutantExit=$?
python3 -I - "$out" "$auditExit" "$auditMutantExit" "$applyAfterExit" "$audit70" << 'PYTHON' | tee -a "$out/verdicts.log"
import json, re, sys
from pathlib import Path
out, code, mutantCode, applyCode, tree = Path(sys.argv[1]), sys.argv[2], sys.argv[3], sys.argv[4], Path(sys.argv[5])
roots = [p for p in (tree / 'src').rglob('*.ts') if '/lib/' not in p.as_posix()]
print(f"w6: a audit scope: {len(roots)} upstream src TypeScript files, "
      f"directories {sorted({p.relative_to(tree / 'src').parts[0] for p in roots})}")
result = {'apply_exit': int(applyCode), 'audit_exit': int(code), 'mutant_audit_exit': int(mutantCode)}
if code == '0':
    audit = json.loads((out / 'a-audit.json').read_text())
    owner = next((o for o in audit['owners'] if o['declaration'] == 'DiagnosticRelatedInformation' and o['member'] == 'file'), None)
    if owner is None:
        print('w6: a FAIL: the audit passed but reports no DiagnosticRelatedInformation.file owner')
        sys.exit(0)
    edited = 'readonly file: SourceFile | undefined;' in (tree / 'src/compiler/types.ts').read_text()
    result.update(owner={k: owner[k] for k in ('where', 'public', 'references', 'writes', 'escapes', 'aliases', 'copies')}, edited=edited)
    ok = not owner['writes'] and not owner['escapes'] and edited and applyCode == '0'
    print(f"w6: a {'PASS' if ok else 'FAIL'}: DiagnosticRelatedInformation.file at {owner['where']}, public {owner.get('public')}, "
          f"{owner['references']} references, {len(owner['writes'])} writes, {len(owner['escapes'])} escapes, "
          f"{len(owner['aliases'])} bounded aliases, {len(owner['copies'])} copies; slot edited {edited}; apply exited {applyCode}")
else:
    log = (out / 'a-audit.log').read_text()
    found = re.search(r'cannot prove (\S+) readonly: (\{.*\})', log)
    if found:
        refusal = json.loads(found.group(2))
        result.update(refused=found.group(1), writes=refusal['writes'], escapes=refusal['escapes'])
        print(f"w6: a FAIL: the audit refuses {found.group(1)} before any edit: {len(refusal['writes'])} writes, "
              f"{len(refusal['escapes'])} escapes; apply exited {applyCode}")
        for write in refusal['writes']:
            print(f"w6: a   write {write}")
        for escape in refusal['escapes']:
            print(f"w6: a   escape {escape['where']} into {escape['target']}")
    else:
        result['error'] = log.strip().splitlines()[-5:]
        print(f"w6: a FAIL: adapt.cjs exited {code}; its last lines: {' | '.join(result['error'])}")
print(f"w6: d mutant audit (target dropped) exited {mutantCode}, as main's apply does")
(out / 'a-summary.json').write_text(json.dumps(result, indent=2) + '\n')
PYTHON
auditPass=0
grep -q '^w6: a PASS' "$out/verdicts.log" && auditPass=1

# The tree c measures as after: apply's own when the audit passed, else the unaudited single edit.
if [ "$auditPass" = 1 ]; then
	measured=$after
	afterLabel=after
else
	measured="$work/unaudited"
	afterLabel=unaudited
	cp -a "$before" "$measured" || exit 66
	# The file comes by environment: a script read from stdin sees `-` as its first argument.
	W6_TYPES="$measured/src/compiler/types.ts" NODE_PATH="$modules" node - > "$out/c-unaudited-edit.log" 2>&1 << 'NODE' || { echo "w6: the unaudited edit failed"; cat "$out/c-unaudited-edit.log"; exit 66; }
const fs = require('node:fs');
const ts = require('typescript');
const file = process.env.W6_TYPES;
const text = fs.readFileSync(file, 'utf8');
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
const slots = [];
ts.forEachChild(source, function visit(node) {
    if (ts.isInterfaceDeclaration(node) && node.name.text === 'DiagnosticRelatedInformation') {
        for (const member of node.members) if (ts.isPropertySignature(member) && member.name.getText(source) === 'file') slots.push(member);
    }
    ts.forEachChild(node, visit);
});
if (slots.length !== 1) throw new Error(`expected one DiagnosticRelatedInformation.file, found ${slots.length}`);
if (ts.getCombinedModifierFlags(slots[0]) & ts.ModifierFlags.Readonly) throw new Error('the slot is readonly already');
const at = slots[0].getStart(source);
fs.writeFileSync(file, text.slice(0, at) + 'readonly ' + text.slice(at));
console.log(`inserted readonly at ${source.getLineAndCharacterOfPosition(at).line + 1}:${source.getLineAndCharacterOfPosition(at).character + 1}`);
NODE
	echo "w6: c measures an unaudited after: the before tree with only readonly at DiagnosticRelatedInformation.file ($(cat "$out/c-unaudited-edit.log"))"
fi

# d's identity: the compiler sources before and after differ only in types.ts, by the one readonly insertion.
python3 -I - "$before" "$measured" "$out/d-sources.json" << 'PYTHON' | tee -a "$out/verdicts.log"
import hashlib, json, sys
from pathlib import Path
before, after, report = Path(sys.argv[1]), Path(sys.argv[2]), Path(sys.argv[3])
def manifest(tree):
    root = tree / 'src/compiler'
    return {p.relative_to(root).as_posix(): hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(root.rglob('*')) if p.is_file()}
old, new = manifest(before), manifest(after)
differing = sorted(name for name in old.keys() | new.keys() if old.get(name) != new.get(name))
oldLines = (before / 'src/compiler/types.ts').read_text().split('\n')
newLines = (after / 'src/compiler/types.ts').read_text().split('\n')
changed = [(i + 1, a, b) for i, (a, b) in enumerate(zip(oldLines, newLines)) if a != b]
single = len(oldLines) == len(newLines) and len(changed) == 1 and changed[0][2] == changed[0][1].replace('file: SourceFile | undefined;', 'readonly file: SourceFile | undefined;', 1)
ok = differing == ['types.ts'] and single
json.dump({'files': len(new), 'differing': differing, 'types_ts_changed_lines': changed}, open(report, 'w'), indent=2)
print(f"w6: d sources {'PASS' if ok else 'FAIL'}: {len(new)} compiler files, differing {differing}, types.ts lines changed {[(l, b.strip()) for l, a, b in changed]}")
PYTHON

# c. The latent census as the meter runs it (first error, scoped tail, no output), before and after, in parallel.
mkdir -p "$work/census-before" "$work/census-$afterLabel"
(LATENT_TIMES="$work/census-before/latent-times.jsonl" LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TAIL=1 \
	step c-census-before "$work/latent-census" "$before/src/compiler" "$work/census-before/latent.jsonl") &
censusBefore=$!
(LATENT_TIMES="$work/census-$afterLabel/latent-times.jsonl" LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TAIL=1 \
	step "c-census-$afterLabel" "$work/latent-census" "$measured/src/compiler" "$work/census-$afterLabel/latent.jsonl") &
censusAfter=$!
wait "$censusBefore"
censusBeforeExit=$?
wait "$censusAfter"
censusAfterExit=$?
gzip -c "$work/census-before/latent.jsonl" > "$out/c-before.jsonl.gz" 2> /dev/null
gzip -c "$work/census-$afterLabel/latent.jsonl" > "$out/c-$afterLabel.jsonl.gz" 2> /dev/null
python3 -I - "$root/stage3/meter" "$before" "$work/census-before" "$measured" "$work/census-$afterLabel" "$afterLabel" "$out" "$censusBeforeExit" "$censusAfterExit" << 'PYTHON' | tee -a "$out/verdicts.log"
import json, re, sys
from collections import Counter
from pathlib import Path
meter = Path(sys.argv[1])
sys.path.insert(0, str(meter))
import report  # The meter's own reader: its coverage checks and its unique-site count.
beforeTree, beforeRun, afterTree, afterRun = map(Path, sys.argv[2:6])
label, out, codes = sys.argv[6], Path(sys.argv[7]), sys.argv[8:10]
family = re.compile(r'seen as (.*), which can write SourceFile \| undefined where SourceFile is read$')
headline = {
    'a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read': 'Diagnostic',
    'a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile | undefined where SourceFile is read': 'DiagnosticRelatedInformation',
}
def measure(tree, run):
    summary = report.latent_summary(tree, run)
    reasons = Counter()
    for key, count in summary['per_reason'].items():
        reasons[key] += count
    members = {key: count for key, count in reasons.items()
               if (m := family.search(key)) and re.search(r'\bDiagnostic(RelatedInformation)?\b', m.group(1))}
    heads = {name: reasons.get('Refused: ' + reason, 0) + reasons.get('NotYet: ' + reason, 0) for reason, name in headline.items()}
    total = summary['totals']['NotYet'] + summary['totals']['Refused']
    return {'totals': summary['totals'], 'latent_total': total, 'family': sum(members.values()), 'headline': heads,
            'family_reasons': members, 'source_files': summary['source_files']}, reasons
results = {}
try:
    results['before'], beforeReasons = measure(beforeTree, beforeRun)
    results[label], afterReasons = measure(afterTree, afterRun)
except Exception as error:
    print(f"w6: c FAIL: a census didn't read whole (exits {codes}): {error}")
    sys.exit(0)
for name in ('before', label):
    r = results[name]
    print(f"w6: c {name}: {r['source_files']} files, latent total {r['latent_total']} {r['totals']}, family {r['family']} "
          f"(DiagnosticWithLocation seen as Diagnostic {r['headline']['Diagnostic']}, as DiagnosticRelatedInformation {r['headline']['DiagnosticRelatedInformation']})")
    for key, count in sorted(r['family_reasons'].items(), key=lambda item: -item[1]):
        print(f"w6: c   {name} {count} {key}")
delta = Counter(afterReasons)
delta.subtract(beforeReasons)
moved = sorted(((count, key) for key, count in delta.items() if count), key=lambda item: (item[0], item[1]))
print(f"w6: c reason changes, {label} less before ({len(moved)} reasons moved):")
for count, key in moved if len(moved) <= 50 else moved[:25] + moved[-25:]:
    print(f"w6: c   {count:+d} {key}")
after = results[label]
ok = after['family'] == 0 and results['before']['family'] > 0
print(f"w6: c {'PASS' if ok else 'FAIL'}{' (unaudited after)' if label == 'unaudited' else ''}: family {results['before']['family']} to {after['family']}, "
      f"headline {results['before']['headline']['Diagnostic']} + {results['before']['headline']['DiagnosticRelatedInformation']} to "
      f"{after['headline']['Diagnostic']} + {after['headline']['DiagnosticRelatedInformation']}, latent total {results['before']['latent_total']} to {after['latent_total']}")
print(f"w6: d {'PASS' if results['before']['family'] > 0 else 'FAIL'}: with the target dropped the family stands at {results['before']['family']} "
      f"(the before tree is the mutant's apply), the latent total at {results['before']['latent_total']}")
results['reason_delta'] = {key: count for count, key in moved}
(out / 'c-summary.json').write_text(json.dumps(results, indent=2) + '\n')
PYTHON

# b. The published API: both adapted trees built by upstream's own build, their typescript.d.ts diffed, and
# check-api.cjs reconstructing each from the pristine snapshot, adaptation 20's ledger and 70's owner report.
wait "$pristineBuild"
pristineExit=$?
if [ "$hasNpm" != 1 ]; then
	say "w6: b SKIP: no npm on this box"
elif [ "$auditPass" != 1 ]; then
	say "w6: b SKIP: the audit refused, so apply made no after tree and no published API to compare"
else
	(cd "$before" && step b-before-install npm ci --no-audit --no-fund && step b-before-build npm run build) &
	buildBefore=$!
	(cd "$after" && step b-after-install npm ci --no-audit --no-fund && step b-after-build npm run build) &
	buildAfter=$!
	wait "$buildBefore"
	buildBeforeExit=$?
	wait "$buildAfter"
	buildAfterExit=$?
	diff -u "$before/built/local/typescript.d.ts" "$after/built/local/typescript.d.ts" > "$out/b-typescript.d.ts.diff"
	echo "w6: b typescript.d.ts, before to after (builds exited $buildBeforeExit and $buildAfterExit):"
	cat "$out/b-typescript.d.ts.diff"
	python3 -I - "$out/b-typescript.d.ts.diff" << 'PYTHON' | tee -a "$out/verdicts.log"
import sys
lines = open(sys.argv[1]).read().split('\n')
removed = [l[1:] for l in lines if l.startswith('-') and not l.startswith('---')]
added = [l[1:] for l in lines if l.startswith('+') and not l.startswith('+++')]
# One line, the slot's own, keeping its indentation: DiagnosticRelatedInformation is the only interface whose file is
# SourceFile | undefined in the snapshot, and check-api below pins the owner by its path.
ok = (len(removed) == len(added) == 1 and removed[0].strip() == 'file: SourceFile | undefined;' and
      added[0] == removed[0].replace('file: SourceFile | undefined;', 'readonly file: SourceFile | undefined;', 1))
print(f"w6: b diff {'PASS' if ok else 'FAIL'}: {len(removed)} line(s) removed, {len(added)} added")
PYTHON
	# Adaptation 20's ledger is its own output on the tree apply handed it.
	ledgerTree="$work/ledger20"
	mkdir -p "$ledgerTree"
	git -C "$after" archive refs/stage3/20-optional-declarations-before | tar -x -C "$ledgerTree"
	NODE_PATH="$modules" node "$root/stage3/adapt/20-optional-declarations/adapt.cjs" "$ledgerTree" > "$out/b-optional-owners.json" 2> "$out/b-optional-owners.log"
	echo "w6: b adaptation 20 ledger exited $? with $(python3 -I -c 'import json, sys; print(len(json.load(open(sys.argv[1]))["declarations"]))' "$out/b-optional-owners.json" 2> /dev/null) declarations; pristine build exited $pristineExit"
	NODE_PATH="$modules" node "$root/stage3/adapt/70-readonly-views/check-api.cjs" "$pristine" "$before" "$out/b-optional-owners.json" "$out/d-mutant-audit.json" > "$out/b-check-api-before.json" 2> "$out/b-check-api-before.log"
	checkBefore=$?
	NODE_PATH="$modules" node "$root/stage3/adapt/70-readonly-views/check-api.cjs" "$pristine" "$after" "$out/b-optional-owners.json" "$out/a-audit.json" > "$out/b-check-api-after.json" 2> "$out/b-check-api-after.log"
	checkAfter=$?
	python3 -I - "$out" "$checkBefore" "$checkAfter" << 'PYTHON' | tee -a "$out/verdicts.log"
import json, sys
from pathlib import Path
out = Path(sys.argv[1])
for name, code in (('before', sys.argv[2]), ('after', sys.argv[3])):
    if code != '0':
        last = (out / f'b-check-api-{name}.log').read_text().strip().splitlines()[-3:]
        print(f"w6: b check-api {name} FAIL: exited {code}: {' | '.join(last)}")
        continue
    r = json.loads((out / f'b-check-api-{name}.json').read_text())
    print(f"w6: b check-api {name} PASS: {r['snapshot_declarations']} optional owners, {r['readonly_declarations']} readonly owners "
          f"{[(o['path'], o['change'], o['line']) for o in r['readonly_owners']]}, {len(r['changed_lines'])} changed lines, "
          f"{r['other_reference_baselines_identical']} other reference baselines identical")
PYTHON
fi

echo "w6: summary:"
grep -E '^w6: [abcd] .*(PASS|FAIL|SKIP)' "$out/verdicts.log"
for part in a c d; do
	grep -qE "^w6: $part PASS" "$out/verdicts.log" || exit 1
done
grep -qE '^w6: b diff PASS' "$out/verdicts.log" || exit 1
! grep -qE '^w6: [abcd] .*(FAIL|SKIP)' "$out/verdicts.log"
