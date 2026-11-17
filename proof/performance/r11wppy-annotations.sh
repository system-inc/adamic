# #r11wppy class 1 and 4's annotations, adaptation 99-readonly-inferred-bindings, sealed: the parent of this commit
# is the head. Applies the tree (with it), then makes two copies: before, every slot's `readonly ` taken back out (the
# tree without it), and the mutant, getUnionSignatures' return alone taken back out. Scans sys.ts's closure in all
# three, runs stock tsc over src/compiler in all three, then the adapted tree's product test and stage3/lane. Run
# after environment.sh, sealed-job.sh, sealed_typescript_source and sealed_npm stage3/api stage3/typescript.
# Never merged.
set -u
bad=0
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
export SCAN_ALL_SCRATCH="$TMPDIR/scan"; mkdir -p "$SCAN_ALL_SCRATCH"
start=$(date +%s)
go test -v -count=1 -timeout=20m -run '^TestProduct_ScanAllInstrumentedProbe$' ./stage3/census/scan-all > "$TMPDIR/product.log" 2>&1 || { echo "VERDICT FAIL product"; tail -20 "$TMPDIR/product.log"; exit 3; }
BIN=$(grep -oE 'probe: /[^ ]+' "$TMPDIR/product.log" | tail -1 | cut -d' ' -f2)
[ -x "$BIN" ] || { echo "VERDICT FAIL no probe binary"; exit 3; }
echo "product $(( $(date +%s) - start )) s"
AFTER="$SCAN_ALL_SCRATCH/after"
start=$(date +%s)
bash stage3/apply.sh --build-product "$AFTER" > "$TMPDIR/apply.log" 2>&1 || { echo "VERDICT FAIL apply"; tail -12 "$TMPDIR/apply.log"; exit 3; }
echo "apply $(( $(date +%s) - start )) s"
grep -A3 '"change": "annotations"' "$TMPDIR/apply.log" | head -4
echo "slots edited by the adaptation: $(grep -c '"edited": true' "$TMPDIR/apply.log")"
BEFORE="$SCAN_ALL_SCRATCH/before"; MUTANT="$SCAN_ALL_SCRATCH/mutant"
cp -a "$AFTER" "$BEFORE"; cp -a "$AFTER" "$MUTANT"
python3 - "$BEFORE" "$MUTANT" proof/performance/r11wppy-annotation-edits.diff <<'PYTHON'
import sys
before, mutant, diff = sys.argv[1:4]
lines = open(diff).read().split('\n')
edits = [(lines[i].rsplit(':', 1)[0], lines[i + 1][1:].strip(), lines[i + 2][1:].strip()) for i in range(len(lines) - 2) if lines[i].startswith('src/')]
def reverse(root, only=None):
    count = 0
    for k, (file, old, new) in enumerate(edits):
        if only is not None and k != only: continue
        p = f"{root}/{file}"; text = open(p, newline='').read().split('\n')
        hits = [n for n, line in enumerate(text) if line.rstrip('\r').strip() == new]
        if len(hits) != 1: print(f"  reverse: {file} {new[:80]} found {len(hits)} times"); continue
        n = hits[0]; indent = text[n][:len(text[n]) - len(text[n].lstrip())]
        text[n] = indent + old + ('\r' if text[n].endswith('\r') else '')
        open(p, 'w', newline='').write('\n'.join(text)); count += 1
    return count
print(f"before: {reverse(before)} of {len(edits)} lines reversed")
print(f"mutant: {reverse(mutant, 0)} line reversed, the first: {edits[0][2][:100]}")
PYTHON
for side in before after mutant; do
	start=$(date +%s)
	"$BIN" "$SCAN_ALL_SCRATCH/$side/src/compiler/sys.ts" > "$TMPDIR/$side.json" 2> "$TMPDIR/$side.stderr"
	echo "SCAN $side exit $?, $(( $(date +%s) - start )) s"
done
python3 - "$TMPDIR/before.json" "$TMPDIR/after.json" "$TMPDIR/mutant.json" <<'PYTHON'
import json, re, sys
def load(p):
    r = json.load(open(p))
    rows = r.get('Stops') or []
    return r, rows, set((s['Rule'], re.sub(r'^.*/scan/(before|after|mutant)/', '', s['Site']), s['Words']) for s in rows)
sides = {name: load(p) for name, p in zip(['before', 'after', 'mutant'], sys.argv[1:])}
for name, (r, rows, distinct) in sides.items():
    print(f"{name}: events {len(rows)}, distinct {len(distinct)}, never[] distinct {sum(1 for x in distinct if 'never[]' in x[2])}, completed {r.get('Completed')}, boundary {r.get('Boundary')}")
def compare(a, b):
    gone, new = sorted(sides[a][2] - sides[b][2]), sorted(sides[b][2] - sides[a][2])
    print(f"{a} to {b}: {len(gone)} gone, {len(new)} new")
    for x in gone: print(f"  - {x[1]}: {x[2][:150]}")
    for x in new: print(f"  + {x[1]}: {x[0]} {x[2][:150]}")
compare('before', 'after')
compare('after', 'mutant')
PYTHON
TSC="stage3/api/node_modules/typescript/bin/tsc"
for side in before after mutant; do
	node "$TSC" -p "$SCAN_ALL_SCRATCH/$side/src/compiler" --noEmit --pretty false > "$TMPDIR/tsc-$side.log" 2>&1
	echo "TSC $side exit $?, errors $(grep -c 'error TS' "$TMPDIR/tsc-$side.log")"
done
new=$(comm -13 <(grep 'error TS' "$TMPDIR/tsc-before.log" | sed 's#^.*/scan/[a-z]*/##' | sort) <(grep 'error TS' "$TMPDIR/tsc-after.log" | sed 's#^.*/scan/[a-z]*/##' | sort))
[ -z "$new" ] && echo "PASS no tsc error new with the adaptation" || { echo "FAIL tsc errors new:"; echo "$new" | head -6 | cut -c1-220; bad=1; }
start=$(date +%s)
go test -count=1 -timeout 30m -run '^TestProduct_Stage3AdaptedTree$' ./stage3/lane/adaptedtree/ > "$TMPDIR/adapted.log" 2>&1
code=$?
echo "$([ $code = 0 ] && echo PASS || echo FAIL) adapted tree product exit $code, $(( $(date +%s) - start )) s"; [ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|panic|.*_test.go:[0-9]+:)' "$TMPDIR/adapted.log" | head -10 | cut -c1-260; bad=1; }
go test -count=1 -timeout 30m -skip 'TestShard_|TestProduct_' ./stage3/lane/... > "$TMPDIR/lane.log" 2>&1
code=$?
echo "$([ $code = 0 ] && echo PASS || echo FAIL) lane exit $code"; [ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|.*_test.go:[0-9]+:)' "$TMPDIR/lane.log" | head -10 | cut -c1-260; bad=1; }
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
