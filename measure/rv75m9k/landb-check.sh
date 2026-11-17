#!/usr/bin/env bash
# #rv75m9k road B's landed form, stacked on road A: apply builds with both directories; the scan reads the 6 left reads
# with stock tsc 0; road B's mutant brings back the 49 (tsc 0) and the adapter re-applies; utilities.ts's census before
# (B's mutant) and after (B); TestProduct_Stage3AdaptedTree; TestGetterReadThroughView. Sealed.
set -uo pipefail
source /opt/adamic-block/environment.sh && source cloud/sealed-job.sh || { echo "FAIL sealed environment"; exit 1; }
here=$(cd "$(dirname "$0")" && pwd)
sealed_typescript_source >/dev/null && sealed_npm stage3/api stage3/typescript >/dev/null && echo "PASS sealed inputs" || exit 1
export STAGE3_PRODUCT_STORE="$TMPDIR/store" STAGE3_CACHE="$TMPDIR/stage3-cache" NODE_PATH="$PWD/stage3/api/node_modules"
tree="$TMPDIR/tree" adaptation=stage3/adapt/99-method-value-views
bash stage3/apply.sh --build-product "$tree" > "$TMPDIR/apply.log" 2>&1 && echo "PASS tree with both" || { tail -20 "$TMPDIR/apply.log"; exit 1; }
grep -E 'method-value' "$tree/patch-set.md"
node "$here/scan.cjs" "$tree" "$here/names.json" "$TMPDIR/landed.json" | grep -E '^(scan|diagnostic)'
overlay="$TMPDIR/overlay"
{ python3 stage3/census/latent/make_overlay.py "$PWD" "$overlay" && gofmt -w "$overlay"/*.go && go build -buildvcs=false -overlay="$overlay/overlay.json" -o "$TMPDIR/latent-census" ./stage3/census/latent/tool; } > "$TMPDIR/cb.log" 2>&1 && echo "PASS census tool" || exit 1
census() { LATENT_FULL=1 LATENT_ONLY_FILE="$tree/src/compiler/transformers/utilities.ts" LATENT_ASSERT_NO_OUTPUT=1 timeout 1200 "$TMPDIR/latent-census" "$tree/src/compiler" "$TMPDIR/$1.jsonl" > "$TMPDIR/$1.log" 2>&1; echo "census $1 exit=$?"; }
census after
node "$adaptation/mutant.cjs" "$tree" | tr -d '\n '; echo
node "$here/scan.cjs" "$tree" "$here/names.json" "$TMPDIR/restored.json" | grep -E '^(scan|diagnostic)'
census before
python3 "$here/compare.py" transformers/utilities.ts "$TMPDIR/before.jsonl" "$TMPDIR/after.jsonl" | grep -E '^(census.*paired|removed|added)' | cut -c1-220
node "$adaptation/adapt.cjs" "$tree" | tr -d '\n '; echo
node "$here/scan.cjs" "$tree" "$here/names.json" "$TMPDIR/readapted.json" | grep -E '^scan'
echo "replay: $(node "$adaptation/adapt.cjs" "$tree" | tr -d '\n ')"
go test -count=1 -timeout 0 -run '^TestProduct_Stage3AdaptedTree$' -v ./stage3/lane/adaptedtree/ > "$TMPDIR/product.log" 2>&1; grep -E -- '--- (PASS|FAIL)|products_test' "$TMPDIR/product.log" | head -5
sealed_products ./internal/oracle > "$TMPDIR/products.log" 2>&1 && echo "PASS oracle products" || { tail -20 "$TMPDIR/products.log"; echo "FAIL oracle products"; }
go test -count=1 -v -run '^TestGetterReadThroughView$' ./internal/oracle/ > "$TMPDIR/pin.log" 2>&1; grep -E -- '--- (PASS|FAIL)' "$TMPDIR/pin.log"
echo "== end"
