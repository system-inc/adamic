#!/bin/bash
# #2v3xtgm wave 3 before review, sealed: 99-z-dictionary-view-kinds, Compiler's ruling (b). Apply into $TMPDIR, then the
# adaptation's check, its witness plain, under each mutant (each must fail) and with the premise's count, then
# stage3/test_apply.py, the adapted tree's product and owner-pin tests, and the sanctioned API's freshness. The census
# runs apart. Never lands.
#
# It runs after the gate inputs are in place: sandboxed_repositories && sandboxed_npm stage3/api stage3/typescript &&
# sandboxed_typescript_source
set -uo pipefail
echo "box $(hostname) $(nproc) cpus, load $(cut -d' ' -f1-3 /proc/loadavg), tree $(git rev-parse HEAD)"
T="$TMPDIR/wave3"
mkdir -p "$T"
TREE="$T/adapted"
ADAPTATION=stage3/adapt/99-z-dictionary-view-kinds
export NODE_PATH="$PWD/stage3/api/node_modules"

stage3/apply.sh "$TREE" > "$T/apply.log" 2>&1 && echo "PASS apply" || { tail -60 "$T/apply.log"; echo "FAIL apply"; exit 1; }
grep -F '99-z-dictionary-view-kinds' "$TREE/patch-set.md"
git -C "$TREE" diff --stat refs/stage3/99-z-dictionary-view-kinds-before refs/stage3/99-z-dictionary-view-kinds-after

# A second application writes nothing.
node "$ADAPTATION/adapt.cjs" "$TREE" > "$T/again.log" 2>&1; code=$?
cat "$T/again.log"
if [ $code = 0 ] && grep -qF '"files":0' "$T/again.log"; then
	echo "PASS a second application writes nothing"
else
	echo "FAIL a second application"; exit 1
fi

node "$ADAPTATION/check.cjs" "$TREE" > "$T/check.log" 2>&1; code=$?
grep -E '^check: (PASS|FAIL|before|after)' "$T/check.log"
[ $code = 0 ] && echo "PASS check" || { tail -40 "$T/check.log"; echo "FAIL check"; exit 1; }

node "$ADAPTATION/witness.cjs" "$TREE" > "$T/witness.log" 2>&1; code=$?
cat "$T/witness.log"
[ $code = 0 ] && echo "PASS witness" || { echo "FAIL witness"; exit 1; }

# mutant NAME WANT: the witness under WITNESS_MUTANT=NAME exits 1, with a FAIL line holding WANT.
mutant() {
	WITNESS_MUTANT="$1" node "$ADAPTATION/witness.cjs" "$TREE" > "$T/mutant-$1.log" 2>&1
	local code=$?
	grep -E '^witness: (FAIL|mutant)' "$T/mutant-$1.log"
	if [ $code = 1 ] && grep -E '^witness: FAIL' "$T/mutant-$1.log" | grep -qF -- "$2"; then
		echo "PASS mutant $1 fails"
	else
		tail -30 "$T/mutant-$1.log"; echo "FAIL mutant $1: exit $code"; exit 1
	fi
}
mutant symmetric 'isJsonEqual:'
mutant object-guard-inverted 'loadModuleFromExports:'

WITNESS_COUNT="$TREE" node "$ADAPTATION/witness.cjs" "$TREE" > "$T/count.log" 2>&1; code=$?
grep -E '^witness: (count:|PASS count|FAIL count)' "$T/count.log"
[ $code = 0 ] && echo "PASS the premise's count" || { tail -40 "$T/count.log"; echo "FAIL the premise's count"; exit 1; }

python3 stage3/test_apply.py > "$T/test-apply.log" 2>&1; code=$?
tail -4 "$T/test-apply.log"
[ $code = 0 ] && echo "PASS test_apply" || { grep -E '^(FAIL|ERROR)' -A20 "$T/test-apply.log" | head -80; echo "FAIL test_apply"; exit 1; }

PINS='^(TestProduct_Stage3AdaptedTree|TestPerformanceCoreYewmv46PinsEqualAdaptedProductBytes|TestBuiltinImportsScopeFollowsOwnerPins|TestTwoImplementsTagsKeepBothInTheDeclaration)$'
go test -count=1 -timeout 0 -v -run "$PINS" ./stage3/lane/adaptedtree > "$T/adaptedtree.log" 2>&1; code=$?
grep -E -- '^\s*--- (PASS|FAIL)' "$T/adaptedtree.log"
[ $code = 0 ] && echo "PASS the adapted tree's product and owner pins" || { grep -E '_test.go:[0-9]+:' -A4 "$T/adaptedtree.log" | head -60; tail -40 "$T/adaptedtree.log"; echo "FAIL adaptedtree"; exit 1; }

python3 stage3/lane/test_sanctioned_api_freshness.py > "$T/freshness.log" 2>&1; code=$?
tail -4 "$T/freshness.log"
[ $code = 0 ] && echo "PASS the sanctioned API is fresh" || { tail -60 "$T/freshness.log"; echo "FAIL freshness"; exit 1; }
echo VERDICT PASS
