# #r11wppy item 1's alias change, adaptation 77, sealed: the parent of this commit is the head. Builds scan-all's
# instrumented probe and tsc's adapted tree, then scans sys.ts's closure twice: after (77 applied) and before (the
# same tree with 77's one `readonly ` taken back out, which is also the mutant that drops it). Then stock tsc over
# src/compiler in both trees, and the adapted tree's own product test. Run after environment.sh, sealed-job.sh,
# sealed_typescript_source and sealed_npm stage3/api. Never merged.
set -u
bad=0
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
export SCAN_ALL_SCRATCH="$TMPDIR/scan"; mkdir -p "$SCAN_ALL_SCRATCH"
start=$(date +%s)
go test -v -count=1 -timeout=20m -run '^TestProduct_ScanAllInstrumentedProbe$' ./stage3/census/scan-all > "$TMPDIR/product.log" 2>&1 || { echo "VERDICT FAIL product"; tail -20 "$TMPDIR/product.log"; exit 3; }
BIN=$(grep -oE 'probe: /[^ ]+' "$TMPDIR/product.log" | tail -1 | cut -d' ' -f2)
[ -x "$BIN" ] || { echo "VERDICT FAIL no probe binary"; grep -n probe "$TMPDIR/product.log" | tail -5; exit 3; }
echo "product $BIN, $(( $(date +%s) - start )) s"
AFTER="$SCAN_ALL_SCRATCH/after"
start=$(date +%s)
bash stage3/apply.sh --build-product "$AFTER" > "$TMPDIR/apply.log" 2>&1 || { echo "VERDICT FAIL apply"; tail -12 "$TMPDIR/apply.log"; exit 3; }
echo "apply $(( $(date +%s) - start )) s"
grep -n 'export type ScriptTargetFeatures' "$AFTER/src/compiler/utilities.ts" | cut -c1-160
grep -q 'ReadonlyMap<string, ReadonlyMap<string, readonly string\[\]>>' "$AFTER/src/compiler/utilities.ts" && echo "PASS the alias is readonly in apply's tree" || { echo "FAIL the alias isn't readonly in apply's tree"; bad=1; }
BEFORE="$SCAN_ALL_SCRATCH/before"
cp -a "$AFTER" "$BEFORE"
sed -i 's/ReadonlyMap<string, ReadonlyMap<string, readonly string\[\]>>/ReadonlyMap<string, ReadonlyMap<string, string[]>>/' "$BEFORE/src/compiler/utilities.ts"
if cmp -s "$AFTER/src/compiler/utilities.ts" "$BEFORE/src/compiler/utilities.ts"; then echo "FAIL the mutant didn't apply"; bad=1; fi
# Variant typed: after, plus an explicit type argument on getScriptTargetFeatures' outer `new Map(`, so the alias
# reaches the inner entries instead of the union tsc infers from them.
TYPED="$SCAN_ALL_SCRATCH/typed"
cp -a "$AFTER" "$TYPED"
python3 - "$TYPED/src/compiler/utilities.ts" <<'PYTHON'
import sys
p = sys.argv[1]; s = open(p, newline='').read()
old = "memoize((): ScriptTargetFeatures =>\r\n    new Map(Object.entries({"
if old not in s: old = old.replace("\r\n", "\n")
assert s.count(old) == 1, 'typed anchor'
s = s.replace(old, old.replace("new Map(", "new Map<string, ReadonlyMap<string, readonly string[]>>("))
open(p, 'w', newline='').write(s)
PYTHON
grep -n 'new Map<string, ReadonlyMap<string, readonly string' "$TYPED/src/compiler/utilities.ts" | cut -c1-140
for side in before after typed; do
	start=$(date +%s)
	if [ $side = before ]; then T=$BEFORE; elif [ $side = typed ]; then T=$TYPED; else T=$AFTER; fi
	"$BIN" "$T/src/compiler/sys.ts" > "$TMPDIR/$side.json" 2> "$TMPDIR/$side.stderr"
	echo "SCAN $side exit $?, $(( $(date +%s) - start )) s"
done
python3 - "$TMPDIR/before.json" "$TMPDIR/after.json" "$TMPDIR/typed.json" <<'PYTHON'
import collections, json, sys
import re
def load(p):
    r = json.load(open(p))
    rows = r.get('Stops') or []
    return r, rows, set((s['Rule'], re.sub(r'^.*/scan/(before|after|typed)/', '', s['Site']), s['Words']) for s in rows)
rb, b, tb = load(sys.argv[1]); ra, a, ta = load(sys.argv[2])
rt, t, tt = load(sys.argv[3])
print(f"typed:  events {len(t)}, distinct {len(tt)}, never[] distinct {sum(1 for x in tt if 'never[] seen as' in x[2])}")
for x in sorted(ta - tt): print(f"  typed clears {x[1]}: {x[2][:150]}")
for x in sorted(tt - ta): print(f"  typed adds {x[1]}: {x[0]} {x[2][:150]}")
never = lambda t: 'never[] seen as' in t[2]
print(f"before: events {len(b)}, distinct {len(tb)}, never[] distinct {sum(map(never, tb))}, completed {rb.get('Completed')}, boundary {rb.get('Boundary')}")
print(f"after:  events {len(a)}, distinct {len(ta)}, never[] distinct {sum(map(never, ta))}, completed {ra.get('Completed')}, boundary {ra.get('Boundary')}")
gone, new = sorted(tb - ta), sorted(ta - tb)
print(f"gone with 77: {len(gone)} ({sum(map(never, gone))} never[]); new with 77: {len(new)}")
for t in gone: print(f"  - {t[1]}: {t[2][:150]}")
for t in new: print(f"  + {t[1]}: {t[0]} {t[2][:150]}")
PYTHON
TSC="stage3/api/node_modules/typescript/bin/tsc"
for side in before after typed; do
	if [ $side = before ]; then T=$BEFORE; elif [ $side = typed ]; then T=$TYPED; else T=$AFTER; fi
	node "$TSC" -p "$T/src/compiler" --noEmit --pretty false > "$TMPDIR/tsc-$side.log" 2>&1
	echo "TSC $side exit $?, errors $(grep -c 'error TS' "$TMPDIR/tsc-$side.log")"
done
diff <(grep 'error TS' "$TMPDIR/tsc-before.log" | sort) <(grep 'error TS' "$TMPDIR/tsc-after.log" | sort) > "$TMPDIR/tsc.diff"; echo "tsc errors that differ: $(grep -c '^[<>]' "$TMPDIR/tsc.diff")"; head -10 "$TMPDIR/tsc.diff" | cut -c1-220
diff <(grep 'error TS' "$TMPDIR/tsc-before.log" | sort) <(grep 'error TS' "$TMPDIR/tsc-typed.log" | sort) | head -6 | cut -c1-220
[ "$(grep -c '^>' "$TMPDIR/tsc.diff")" = 0 ] && echo "PASS no tsc error new with 77" || { echo "FAIL tsc errors new with 77"; bad=1; }
echo "VERDICT FAIL product"; tail -20 "$TMPDIR/product.log"; exit 3; }
BIN=$(grep -oE 'probe: /[^ ]+' "$TMPDIR/product.log" | tail -1 | cut -d' ' -f2)
[ -x "$BIN" ] || { echo "VERDICT FAIL no probe binary"; grep -n probe "$TMPDIR/product.log" | tail -5; exit 3; }
echo "product $BIN, $(( $(date +%s) - start )) s"
AFTER="$SCAN_ALL_SCRATCH/after"
start=$(date +%s)
bash stage3/apply.sh --build-product "$AFTER" > "$TMPDIR/apply.log" 2>&1 || { echo "VERDICT FAIL apply"; tail -12 "$TMPDIR/apply.log"; exit 3; }
echo "apply $(( $(date +%s) - start )) s"
grep -n 'export type ScriptTargetFeatures' "$AFTER/src/compiler/utilities.ts" | cut -c1-160
grep -q 'ReadonlyMap<string, ReadonlyMap<string, readonly string\[\]>>' "$AFTER/src/compiler/utilities.ts" && echo "PASS the alias is readonly in apply's tree" || { echo "FAIL the alias isn't readonly in apply's tree"; bad=1; }
BEFORE="$SCAN_ALL_SCRATCH/before"
cp -a "$AFTER" "$BEFORE"
sed -i 's/ReadonlyMap<string, ReadonlyMap<string, readonly string\[\]>>/ReadonlyMap<string, ReadonlyMap<string, string[]>>/' "$BEFORE/src/compiler/utilities.ts"
if cmp -s "$AFTER/src/compiler/utilities.ts" "$BEFORE/src/compiler/utilities.ts"; then echo "FAIL the mutant didn't apply"; bad=1; fi
# Variant typed: after, plus an explicit type argument on getScriptTargetFeatures' outer `new Map(`, so the alias
# reaches the inner entries instead of the union tsc infers from them.
TYPED="$SCAN_ALL_SCRATCH/typed"
cp -a "$AFTER" "$TYPED"
python3 - "$TYPED/src/compiler/utilities.ts" <<'PYTHON'
import sys
p = sys.argv[1]; s = open(p, newline='').read()
old = "memoize((): ScriptTargetFeatures =>\r\n    new Map(Object.entries({"
if old not in s: old = old.replace("\r\n", "\n")
assert s.count(old) == 1, 'typed anchor'
s = s.replace(old, old.replace("new Map(", "new Map<string, ReadonlyMap<string, readonly string[]>>("))
open(p, 'w', newline='').write(s)
PYTHON
grep -n 'new Map<string, ReadonlyMap<string, readonly string' "$TYPED/src/compiler/utilities.ts" | cut -c1-140
for side in before after typed; do
	start=$(date +%s)
	if [ $side = before ]; then T=$BEFORE; elif [ $side = typed ]; then T=$TYPED; else T=$AFTER; fi
	"$BIN" "$T/src/compiler/sys.ts" > "$TMPDIR/$side.json" 2> "$TMPDIR/$side.stderr"
	echo "SCAN $side exit $?, $(( $(date +%s) - start )) s"
done
python3 - "$TMPDIR/before.json" "$TMPDIR/after.json" "$TMPDIR/typed.json" <<'PYTHON'
import collections, json, sys
import re
def load(p):
    r = json.load(open(p))
    rows = r.get('Stops') or []
    return r, rows, set((s['Rule'], re.sub(r'^.*/scan/(before|after|typed)/', '', s['Site']), s['Words']) for s in rows)
rb, b, tb = load(sys.argv[1]); ra, a, ta = load(sys.argv[2])
rt, t, tt = load(sys.argv[3])
print(f"typed:  events {len(t)}, distinct {len(tt)}, never[] distinct {sum(1 for x in tt if 'never[] seen as' in x[2])}")
for x in sorted(ta - tt): print(f"  typed clears {x[1]}: {x[2][:150]}")
for x in sorted(tt - ta): print(f"  typed adds {x[1]}: {x[0]} {x[2][:150]}")
never = lambda t: 'never[] seen as' in t[2]
print(f"before: events {len(b)}, distinct {len(tb)}, never[] distinct {sum(map(never, tb))}, completed {rb.get('Completed')}, boundary {rb.get('Boundary')}")
print(f"after:  events {len(a)}, distinct {len(ta)}, never[] distinct {sum(map(never, ta))}, completed {ra.get('Completed')}, boundary {ra.get('Boundary')}")
gone, new = sorted(tb - ta), sorted(ta - tb)
print(f"gone with 77: {len(gone)} ({sum(map(never, gone))} never[]); new with 77: {len(new)}")
for t in gone: print(f"  - {t[1]}: {t[2][:150]}")
for t in new: print(f"  + {t[1]}: {t[0]} {t[2][:150]}")
PYTHON
TSC="stage3/api/node_modules/typescript/bin/tsc"
for side in before after typed; do
	if [ $side = before ]; then T=$BEFORE; elif [ $side = typed ]; then T=$TYPED; else T=$AFTER; fi
	node "$TSC" -p "$T/src/compiler" --noEmit --pretty false > "$TMPDIR/tsc-$side.log" 2>&1
	echo "TSC $side exit $?, errors $(grep -c 'error TS' "$TMPDIR/tsc-$side.log")"
done
diff <(grep 'error TS' "$TMPDIR/tsc-before.log" | sort) <(grep 'error TS' "$TMPDIR/tsc-after.log" | sort) > "$TMPDIR/tsc.diff"; echo "tsc errors that differ: $(grep -c '^[<>]' "$TMPDIR/tsc.diff")"; head -10 "$TMPDIR/tsc.diff" | cut -c1-220
diff <(grep 'error TS' "$TMPDIR/tsc-before.log" | sort) <(grep 'error TS' "$TMPDIR/tsc-typed.log" | sort) | head -6 | cut -c1-220
[ "$(grep -c '^>' "$TMPDIR/tsc.diff")" = 0 ] && echo "PASS no tsc error new with 77" || { echo "FAIL tsc errors new with 77"; bad=1; }
start=$(date +%s)
go test -count=1 -timeout 30m -run '^TestProduct_Stage3AdaptedTree$' ./stage3/... > "$TMPDIR/adapted.log" 2>&1
code=$?
echo "$([ $code = 0 ] && echo PASS || echo FAIL) adapted tree product exit $code, $(( $(date +%s) - start )) s"; [ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|panic|.*_test.go:[0-9]+:)' "$TMPDIR/adapted.log" | head -10 | cut -c1-260; bad=1; }
grep -E '^(ok|FAIL|---)' "$TMPDIR/adapted.log" | grep -v 'no test files' | head -6
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
