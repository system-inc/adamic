#!/usr/bin/env bash
# Emit's review of tail family 120 (#bbssdcy), never to land: the head's apply and test.cjs at the pin, then
# a two-identifier mutant planted in adapt.cjs's after text with the tree regenerated, so the witness itself
# must turn red rather than the recorded-bytes check.
set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
W="$TMPDIR/w120"; mkdir -p "$W" && echo "PASS scratch $W" || { echo "FAIL scratch"; exit 1; }
type sandboxed_npm >/dev/null 2>&1 && echo "PASS helpers defined" || { echo "FAIL helpers"; exit 1; }
sandboxed_npm stage3/api stage3/typescript >"$W/npm.log" 2>&1 && echo "PASS npm" || { echo "FAIL npm"; tail -20 "$W/npm.log"; exit 1; }
sandboxed_typescript_source >"$W/ts.log" 2>&1 && echo "PASS typescript source" || { echo "FAIL typescript source"; tail -20 "$W/ts.log"; exit 1; }
A=stage3/adapt/99-seats-tail24-120-context-object-map
export NODE_PATH="$PWD/stage3/api/node_modules"
run_side() {
	name=$1
	bash stage3/apply.sh "$W/$name" >"$W/apply-$name.log" 2>&1; code=$?
	echo "apply $name exit=$code"; [ $code -eq 0 ] || { tail -30 "$W/apply-$name.log"; return 1; }
	grep -n 'collectedContextObject\|contextObjectMap' "$W/$name/src/compiler/transformers/module/system.ts" | head -12
	node "$A/test.cjs" "$W/$name" >"$W/test-$name.log" 2>&1; code=$?
	echo "test.cjs $name exit=$code"; cat "$W/test-$name.log" | head -40
}
run_side head
cp "$A/adapt.cjs" "$W/adapt.cjs.head"
python3 - "$A/adapt.cjs" <<'PY'
import sys
path = sys.argv[1]
text = open(path).read()
old = 'contextObjectMap.set(id, collectedContextObject);'
new = 'contextObjectMap.set(id, factory.createUniqueName(\\"context\\"));'
assert text.count(old) == 1, text.count(old)
open(path, 'w').write(text.replace(old, new))
print("PASS mutant planted: the stored value is a second createUniqueName")
PY
run_side mutant
cp "$W/adapt.cjs.head" "$A/adapt.cjs" && echo "PASS restored adapt.cjs"
echo DONE
