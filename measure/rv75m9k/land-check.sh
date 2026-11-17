#!/usr/bin/env bash
# #rv75m9k: the landed form of road A. stage3/apply.sh builds the tree with 99-method-value-properties installed; the scan
# reads it (49 method value reads expected, stock tsc 0); its mutant restores the members (129 expected, stock tsc 0); the
# adapter re-applies; a replay writes nothing. Sealed: it sources the image's environment itself.
set -uo pipefail
source /opt/adamic-block/environment.sh && source cloud/sealed-job.sh || { echo "FAIL sealed environment"; exit 1; }
here=$(cd "$(dirname "$0")" && pwd)
sealed_typescript_source >/dev/null && sealed_npm stage3/api >/dev/null && echo "PASS sealed inputs" || { echo "FAIL sealed inputs"; exit 1; }
export STAGE3_PRODUCT_STORE="$TMPDIR/store" STAGE3_CACHE="$TMPDIR/stage3-cache" NODE_PATH="$PWD/stage3/api/node_modules"
tree="$TMPDIR/tree" adaptation=stage3/adapt/99-method-value-properties
bash stage3/apply.sh --build-product "$tree" > "$TMPDIR/apply.log" 2>&1 && echo "PASS adapted tree built with $adaptation" || { echo "FAIL apply"; tail -30 "$TMPDIR/apply.log"; exit 1; }
grep -E '99-method-value-properties' "$tree/patch-set.md"
echo "== landed"; node "$here/scan.cjs" "$tree" "$here/names.json" "$TMPDIR/landed.json" | grep -vE '^(A |written)'
echo "== restored"; node "$adaptation/mutant.cjs" "$tree" | tr -d '\n '; echo
node "$here/scan.cjs" "$tree" "$here/names.json" "$TMPDIR/restored.json" | grep -vE '^(A |written)'
echo "== re-adapted"; node "$adaptation/adapt.cjs" "$tree" | tr -d '\n '; echo
node "$here/scan.cjs" "$tree" "$here/names.json" "$TMPDIR/readapted.json" | grep -E '^scan'
echo "== replay"; node "$adaptation/adapt.cjs" "$tree" | tr -d '\n '; echo
echo "== end"
