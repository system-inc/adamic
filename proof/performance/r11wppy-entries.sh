# #r11wppy class 2, adaptation 99-readonly-script-target-entries in its minimal form, sealed: the parent of this
# commit is the head. Applies the tree (with it), then: before (every argument the adaptation wrote taken out),
# outer (the outer argument taken out) and one copy per mixed map with its inner argument taken out. Scans sys.ts's
# closure in each, runs stock tsc over src/compiler before and after, then the adapted tree's product test and
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
grep -A2 '"maps"' "$TMPDIR/apply.log" | head -6
variants="before outer Promise BigInt64Array BigUint64Array"
for variant in $variants; do
	cp -a "$AFTER" "$SCAN_ALL_SCRATCH/$variant"
	python3 - "$SCAN_ALL_SCRATCH/$variant/src/compiler/utilities.ts" "$variant" <<'PYTHON'
import re, sys
p, variant = sys.argv[1:3]
s = open(p, newline='').read()
start = s.index('export const getScriptTargetFeatures')
end = s.index('\nexport ', start + 10)
body = s[start:end]
outer = 'Object.entries<ReadonlyMap<string, readonly string[]>>('
inner = 'Object.entries<readonly string[]>('
found = (body.count(outer), body.count(inner))
if variant in ('before', 'outer'):
    assert body.count(outer) == 1, 'outer anchor'
    body = body.replace(outer, 'Object.entries(')
if variant == 'before':
    body = body.replace(inner, 'Object.entries(')
if variant not in ('before', 'outer'):
    anchor = re.compile(r'(\b' + re.escape(variant) + r':\s*new Map\()Object\.entries<readonly string\[\]>\(')
    assert len(anchor.findall(body)) == 1, f'{variant} anchor'
    body = anchor.sub(lambda m: m.group(1) + 'Object.entries(', body)
open(p, 'w', newline='').write(s[:start] + body + s[end:])
print(f"{variant}: found outer {found[0]}, inner {found[1]}; left outer {body.count(outer)}, inner {body.count(inner)}")
PYTHON
done
for side in after $variants; do
	"$BIN" "$SCAN_ALL_SCRATCH/$side/src/compiler/sys.ts" > "$TMPDIR/$side.json" 2>/dev/null
	echo "SCAN $side exit $?"
done
python3 - "$TMPDIR" $variants <<'PYTHON'
import json, re, sys
root, variants = sys.argv[1], sys.argv[2:]
def load(name):
    rows = json.load(open(f"{root}/{name}.json")).get('Stops') or []
    return set((s['Rule'], re.sub(r'^.*/scan/[A-Za-z0-9]+/', '', s['Site']), s['Words']) for s in rows)
after = load('after')
inside = lambda d: [x for x in d if x[1].startswith('src/compiler/utilities.ts:') and 1380 <= int(x[1].split(':')[1]) <= 2000]
print(f"after: distinct {len(after)}, in getScriptTargetFeatures {len(inside(after))}")
for v in variants:
    d = load(v)
    gone, new = sorted(after - d), sorted(d - after)
    print(f"{v}: distinct {len(d)}, in getScriptTargetFeatures {len(inside(d))}; against after {len(new)} back, {len(gone)} gone")
    for x in new: print(f"  + {x[1]}: {x[2][:120]}")
PYTHON
TSC="stage3/api/node_modules/typescript/bin/tsc"
for side in before after; do
	node "$TSC" -p "$SCAN_ALL_SCRATCH/$side/src/compiler" --noEmit --pretty false > "$TMPDIR/tsc-$side.log" 2>&1
	echo "TSC $side errors $(grep -c 'error TS' "$TMPDIR/tsc-$side.log")"
done
new=$(comm -13 <(grep 'error TS' "$TMPDIR/tsc-before.log" | sed 's#^.*/scan/[a-zA-Z0-9]*/##' | sort) <(grep 'error TS' "$TMPDIR/tsc-after.log" | sed 's#^.*/scan/[a-zA-Z0-9]*/##' | sort))
[ -z "$new" ] && echo "PASS no tsc error new with the adaptation" || { echo "FAIL tsc errors new:"; echo "$new" | head -6 | cut -c1-220; bad=1; }
go test -count=1 -timeout 30m -run '^TestProduct_Stage3AdaptedTree$' ./stage3/lane/adaptedtree/ > "$TMPDIR/adapted.log" 2>&1
code=$?; echo "$([ $code = 0 ] && echo PASS || echo FAIL) adapted tree product exit $code"; [ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|.*_test.go:[0-9]+:)' "$TMPDIR/adapted.log" | head -8 | cut -c1-260; bad=1; }
go test -count=1 -timeout 30m -skip 'TestShard_|TestProduct_' ./stage3/lane/... > "$TMPDIR/lane.log" 2>&1
code=$?; echo "$([ $code = 0 ] && echo PASS || echo FAIL) lane exit $code"; [ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|.*_test.go:[0-9]+:)' "$TMPDIR/lane.log" | head -8 | cut -c1-260; bad=1; }
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
