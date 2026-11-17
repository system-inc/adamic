# #r11wppy class 2, adaptation 99-readonly-script-target-entries, sealed: the parent of this commit is the head.
# Applies the tree (with it), then three copies: before (every type argument the adaptation wrote taken back out),
# outer (only the outer argument taken out) and inner (one inner argument, the first, taken out). Scans sys.ts's
# closure in all four, runs stock tsc over src/compiler in all four, then the adapted tree's product test and
# stage3/lane. Run after environment.sh, sealed-job.sh, sealed_typescript_source and sealed_npm stage3/api
# stage3/typescript. Never merged.
set -u
bad=0
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
export SCAN_ALL_SCRATCH="$TMPDIR/scan"; mkdir -p "$SCAN_ALL_SCRATCH"
go test -v -count=1 -timeout=20m -run '^TestProduct_ScanAllInstrumentedProbe$' ./stage3/census/scan-all > "$TMPDIR/product.log" 2>&1 || { echo "VERDICT FAIL product"; exit 3; }
BIN=$(grep -oE 'probe: /[^ ]+' "$TMPDIR/product.log" | tail -1 | cut -d' ' -f2)
AFTER="$SCAN_ALL_SCRATCH/after"
bash stage3/apply.sh --build-product "$AFTER" > "$TMPDIR/apply.log" 2>&1 || { echo "VERDICT FAIL apply"; tail -12 "$TMPDIR/apply.log"; exit 3; }
grep -E 'inner|outer|entries' "$TMPDIR/apply.log" | grep -i 'script' | head -4 | cut -c1-200
for variant in before outer inner innerall innerempty; do
	cp -a "$AFTER" "$SCAN_ALL_SCRATCH/$variant"
	python3 - "$SCAN_ALL_SCRATCH/$variant/src/compiler/utilities.ts" "$variant" <<'PYTHON'
import sys
p, variant = sys.argv[1:3]
s = open(p, newline='').read()
start = s.index('export const getScriptTargetFeatures')
end = s.index('\nexport ', start + 10)
body = s[start:end]
outer = 'Object.entries<ReadonlyMap<string, readonly string[]>>('
inner = 'Object.entries<readonly string[]>('
counts = (body.count(outer), body.count(inner))
if variant in ('before', 'outer'):
    assert body.count(outer) == 1, 'outer anchor'
    body = body.replace(outer, 'Object.entries(')
if variant == 'before':
    body = body.replace(inner, 'Object.entries(')
if variant == 'inner':
    assert body.count(inner) >= 1, 'inner anchor'
    body = body.replace(inner, 'Object.entries(', 1)
if variant == 'innerall':
    body = body.replace(inner, 'Object.entries(')
if variant == 'innerempty':
    # Take the inner argument off every inner map whose own literal holds the shared emptyArray.
    out, i, taken = [], 0, 0
    while True:
        j = body.find(inner, i)
        if j < 0:
            out.append(body[i:])
            break
        k = body.find('}))', j)
        if 'emptyArray' in body[j:k]:
            out.append(body[i:j] + 'Object.entries(')
            taken += 1
        else:
            out.append(body[i:j] + inner)
        i = j + len(inner)
    body = ''.join(out)
    print(f"innerempty: {taken} inner arguments taken off maps holding emptyArray")
open(p, 'w', newline='').write(s[:start] + body + s[end:])
print(f"{variant}: found outer {counts[0]}, inner {counts[1]}; left outer {body.count(outer)}, inner {body.count(inner)}")
PYTHON
done
for side in before after outer inner innerall innerempty; do
	"$BIN" "$SCAN_ALL_SCRATCH/$side/src/compiler/sys.ts" > "$TMPDIR/$side.json" 2>/dev/null
	echo "SCAN $side exit $?"
done
python3 - "$TMPDIR/before.json" "$TMPDIR/after.json" "$TMPDIR/outer.json" "$TMPDIR/inner.json" "$TMPDIR/innerall.json" "$TMPDIR/innerempty.json" <<'PYTHON'
import json, re, sys
def load(p):
    rows = json.load(open(p)).get('Stops') or []
    return set((s['Rule'], re.sub(r'^.*/scan/(before|after|outer|inner|innerall|innerempty)/', '', s['Site']), s['Words']) for s in rows)
sides = {n: load(p) for n, p in zip(['before', 'after', 'outer', 'inner', 'innerall', 'innerempty'], sys.argv[1:])}
def inside(d):
    return sorted(x for x in d if x[1].startswith('src/compiler/utilities.ts:') and 1380 <= int(x[1].split(':')[1]) <= 2000)
for n, d in sides.items():
    print(f"{n}: distinct {len(d)}, in getScriptTargetFeatures {len(inside(d))}")
for a, b in (('before', 'after'), ('after', 'outer'), ('after', 'inner'), ('after', 'innerall'), ('after', 'innerempty')):
    gone, new = sorted(sides[a] - sides[b]), sorted(sides[b] - sides[a])
    print(f"{a} to {b}: {len(gone)} gone, {len(new)} new")
    if a != 'before':
        for x in new: print(f"  + {x[1]}: {x[2][:150]}")
    else:
        for x in new: print(f"  NEW {x[1]}: {x[0]} {x[2][:150]}")
PYTHON
TSC="stage3/api/node_modules/typescript/bin/tsc"
for side in before after; do
	node "$TSC" -p "$SCAN_ALL_SCRATCH/$side/src/compiler" --noEmit --pretty false > "$TMPDIR/tsc-$side.log" 2>&1
	echo "TSC $side errors $(grep -c 'error TS' "$TMPDIR/tsc-$side.log")"
done
new=$(comm -13 <(grep 'error TS' "$TMPDIR/tsc-before.log" | sed 's#^.*/scan/[a-z]*/##' | sort) <(grep 'error TS' "$TMPDIR/tsc-after.log" | sed 's#^.*/scan/[a-z]*/##' | sort))
[ -z "$new" ] && echo "PASS no tsc error new with the adaptation" || { echo "FAIL tsc errors new:"; echo "$new" | head -6 | cut -c1-220; bad=1; }
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
