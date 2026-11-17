# #r11wppy item 1's alias change, adaptation 77, sealed, round 2: the parent of this commit carries the head. Scans
# sys.ts's closure three times: before (77's one `readonly ` taken back out, also the mutant that drops it), after
# (77 applied), and typed (after, plus an explicit type argument on getScriptTargetFeatures' outer `new Map(`, so the
# alias reaches the inner entries instead of the union tsc infers from them). Then stock tsc over src/compiler in all
# three. Run after environment.sh, sealed-job.sh, sealed_typescript_source and sealed_npm stage3/api. Never merged.
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
grep -q 'ReadonlyMap<string, ReadonlyMap<string, readonly string\[\]>>' "$AFTER/src/compiler/utilities.ts" && echo "PASS the alias is readonly in apply's tree" || { echo "FAIL the alias isn't readonly in apply's tree"; bad=1; }
BEFORE="$SCAN_ALL_SCRATCH/before"
cp -a "$AFTER" "$BEFORE"
sed -i 's/ReadonlyMap<string, ReadonlyMap<string, readonly string\[\]>>/ReadonlyMap<string, ReadonlyMap<string, string[]>>/' "$BEFORE/src/compiler/utilities.ts"
cmp -s "$AFTER/src/compiler/utilities.ts" "$BEFORE/src/compiler/utilities.ts" && { echo "FAIL the mutant didn't apply"; bad=1; }
TYPED="$SCAN_ALL_SCRATCH/typed"
cp -a "$AFTER" "$TYPED"
python3 - "$TYPED/src/compiler/utilities.ts" <<'PYTHON'
import re, sys
p = sys.argv[1]
s = open(p, newline='').read()
pattern = re.compile(r'(memoize\(\(\): ScriptTargetFeatures =>\s*new Map)\(')
assert len(pattern.findall(s)) == 1, 'typed anchor'
s = pattern.sub(lambda m: m.group(1) + '<string, ReadonlyMap<string, readonly string[]>>(', s)
open(p, 'w', newline='').write(s)
PYTHON
grep -c 'new Map<string, ReadonlyMap<string, readonly string\[\]>>(' "$TYPED/src/compiler/utilities.ts" | sed 's/^/typed insertions: /'
for side in before after typed; do
	start=$(date +%s)
	"$BIN" "$SCAN_ALL_SCRATCH/$side/src/compiler/sys.ts" > "$TMPDIR/$side.json" 2> "$TMPDIR/$side.stderr"
	echo "SCAN $side exit $?, $(( $(date +%s) - start )) s"
done
python3 - "$TMPDIR/before.json" "$TMPDIR/after.json" "$TMPDIR/typed.json" <<'PYTHON'
import json, re, sys
def load(p):
    r = json.load(open(p))
    rows = r.get('Stops') or []
    return r, rows, set((s['Rule'], re.sub(r'^.*/scan/(before|after|typed)/', '', s['Site']), s['Words']) for s in rows)
sides = {name: load(p) for name, p in zip(['before', 'after', 'typed'], sys.argv[1:])}
for name, (r, rows, distinct) in sides.items():
    print(f"{name}: events {len(rows)}, distinct {len(distinct)}, completed {r.get('Completed')}, boundary {r.get('Boundary')}")
def compare(a, b):
    gone, new = sorted(sides[a][2] - sides[b][2]), sorted(sides[b][2] - sides[a][2])
    print(f"{a} to {b}: {len(gone)} gone, {len(new)} new")
    for x in gone: print(f"  - {x[1]}: {x[2][:150]}")
    for x in new: print(f"  + {x[1]}: {x[0]} {x[2][:150]}")
compare('before', 'after')
compare('after', 'typed')
alias = [x for x in sides['typed'][2] if x[1].startswith('src/compiler/utilities.ts:') and 1384 <= int(x[1].split(':')[1]) <= 2000]
print(f"typed: stops left in getScriptTargetFeatures' body: {len(alias)}")
for x in sorted(alias): print(f"  = {x[1]}: {x[2][:150]}")
PYTHON
TSC="stage3/api/node_modules/typescript/bin/tsc"
for side in before after typed; do
	node "$TSC" -p "$SCAN_ALL_SCRATCH/$side/src/compiler" --noEmit --pretty false > "$TMPDIR/tsc-$side.log" 2>&1
	echo "TSC $side exit $?, errors $(grep -c 'error TS' "$TMPDIR/tsc-$side.log")"
done
for side in after typed; do
	new=$(comm -13 <(grep 'error TS' "$TMPDIR/tsc-before.log" | sed 's#^.*/scan/[a-z]*/##' | sort) <(grep 'error TS' "$TMPDIR/tsc-$side.log" | sed 's#^.*/scan/[a-z]*/##' | sort))
	[ -z "$new" ] && echo "PASS no tsc error new in $side" || { echo "FAIL tsc errors new in $side:"; echo "$new" | head -6 | cut -c1-220; bad=1; }
done
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
