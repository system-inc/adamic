# #r11wppy class 2: the 16 inner entries of getScriptTargetFeatures' literal, seen as the union tsc infers from
# their siblings. Scans sys.ts's closure on main's apply tree and two variants: outer, an explicit type argument
# on the outer Object.entries; inner, that plus one on every inner Object.entries in the function. Then stock
# tsc over src/compiler in all three. Run after environment.sh, sealed-job.sh, sealed_typescript_source and
# sealed_npm stage3/api. Never merged.
set -u
bad=0
export SCAN_ALL_SCRATCH="$TMPDIR/scan"; mkdir -p "$SCAN_ALL_SCRATCH"
go test -v -count=1 -timeout=20m -run '^TestProduct_ScanAllInstrumentedProbe$' ./stage3/census/scan-all > "$TMPDIR/product.log" 2>&1 || { echo "VERDICT FAIL product"; exit 3; }
BIN=$(grep -oE 'probe: /[^ ]+' "$TMPDIR/product.log" | tail -1 | cut -d' ' -f2)
MAIN="$SCAN_ALL_SCRATCH/main"
bash stage3/apply.sh --build-product "$MAIN" > "$TMPDIR/apply.log" 2>&1 || { echo "VERDICT FAIL apply"; exit 3; }
for variant in outer inner; do
	cp -a "$MAIN" "$SCAN_ALL_SCRATCH/$variant"
	python3 - "$SCAN_ALL_SCRATCH/$variant/src/compiler/utilities.ts" "$variant" <<'PYTHON'
import re, sys
p, variant = sys.argv[1:3]
s = open(p, newline='').read()
start = s.index('export const getScriptTargetFeatures')
end = s.index('\nexport ', start + 10)
body = s[start:end]
outer = re.compile(r'(=>\s*new Map\()Object\.entries\(')
assert len(outer.findall(body)) == 1, 'outer anchor'
body = outer.sub(lambda m: m.group(1) + 'Object.entries<ReadonlyMap<string, readonly string[]>>(', body)
inner_count = 0
if variant == 'inner':
    inner = re.compile(r'(:\s*new Map\()Object\.entries\(')
    inner_count = len(inner.findall(body))
    body = inner.sub(lambda m: m.group(1) + 'Object.entries<readonly string[]>(', body)
open(p, 'w', newline='').write(s[:start] + body + s[end:])
print(f"{variant}: outer typed, {inner_count} inner typed")
PYTHON
done
for side in main outer inner; do
	"$BIN" "$SCAN_ALL_SCRATCH/$side/src/compiler/sys.ts" > "$TMPDIR/$side.json" 2>/dev/null
	echo "SCAN $side exit $?"
done
python3 - "$TMPDIR/main.json" "$TMPDIR/outer.json" "$TMPDIR/inner.json" <<'PYTHON'
import json, re, sys
def load(p):
    rows = json.load(open(p)).get('Stops') or []
    return set((s['Rule'], re.sub(r'^.*/scan/(main|outer|inner)/', '', s['Site']), s['Words']) for s in rows)
sides = {n: load(p) for n, p in zip(['main', 'outer', 'inner'], sys.argv[1:])}
for n, d in sides.items():
    body = [x for x in d if x[1].startswith('src/compiler/utilities.ts:') and 1380 <= int(x[1].split(':')[1]) <= 2000]
    print(f"{n}: distinct {len(d)}, in getScriptTargetFeatures {len(body)}")
for a, b in (('main', 'outer'), ('main', 'inner')):
    gone, new = sorted(sides[a] - sides[b]), sorted(sides[b] - sides[a])
    print(f"{a} to {b}: {len(gone)} gone, {len(new)} new")
    for x in new: print(f"  + {x[1]}: {x[0]} {x[2][:170]}")
left = sorted(x for x in sides['inner'] if x[1].startswith('src/compiler/utilities.ts:') and 1380 <= int(x[1].split(':')[1]) <= 2000)
for x in left: print(f"  inner leaves {x[1]}: {x[2][:170]}")
PYTHON
TSC="stage3/api/node_modules/typescript/bin/tsc"
for side in main outer inner; do
	node "$TSC" -p "$SCAN_ALL_SCRATCH/$side/src/compiler" --noEmit --pretty false > "$TMPDIR/tsc-$side.log" 2>&1
	echo "TSC $side errors $(grep -c 'error TS' "$TMPDIR/tsc-$side.log")"
	grep 'error TS' "$TMPDIR/tsc-$side.log" | grep utilities.ts | head -3 | cut -c1-220
done
echo "VERDICT DONE"
