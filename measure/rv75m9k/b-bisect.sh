#!/usr/bin/env bash
# #rv75m9k: which of road B's edits moves parser.ts's census? A full parser.ts census on the tree with road A installed,
# then with road B applied to parser.ts alone, then to every other file alone, each from a fresh copy of the tree.
set -uo pipefail
source /opt/adamic-block/environment.sh && source cloud/sealed-job.sh || { echo "FAIL sealed environment"; exit 1; }
here=$(cd "$(dirname "$0")" && pwd)
sealed_typescript_source >/dev/null && sealed_npm stage3/api >/dev/null && echo "PASS sealed inputs" || exit 1
export STAGE3_PRODUCT_STORE="$TMPDIR/store" STAGE3_CACHE="$TMPDIR/stage3-cache" NODE_PATH="$PWD/stage3/api/node_modules"
base="$TMPDIR/base"
bash stage3/apply.sh --build-product "$base" > "$TMPDIR/apply.log" 2>&1 && echo "PASS tree" || { tail -20 "$TMPDIR/apply.log"; exit 1; }
overlay="$TMPDIR/overlay"
{ python3 stage3/census/latent/make_overlay.py "$PWD" "$overlay" && gofmt -w "$overlay"/*.go && go build -buildvcs=false -overlay="$overlay/overlay.json" -o "$TMPDIR/latent-census" ./stage3/census/latent/tool; } > "$TMPDIR/cb.log" 2>&1 && echo "PASS census tool" || exit 1
node "$here/scan.cjs" "$base" "$here/names.json" "$TMPDIR/sites.json" > /dev/null
census() { LATENT_FULL=1 LATENT_ONLY_FILE="$1/src/compiler/parser.ts" LATENT_ASSERT_NO_OUTPUT=1 timeout 1500 "$TMPDIR/latent-census" "$1/src/compiler" "$2" > "$2.log" 2>&1; echo "census $2 exit=$?"; }
census "$base" "$TMPDIR/none.jsonl"
for variant in only except; do
	tree="$TMPDIR/$variant"; cp -R "$base" "$tree"
	if [ $variant = only ]; then ROAD_B_ONLY=src/compiler/parser.ts node "$here/codemod-b.cjs" "$tree" "$TMPDIR/sites.json" | grep '^codemod:'
	else ROAD_B_EXCEPT=src/compiler/parser.ts node "$here/codemod-b.cjs" "$tree" "$TMPDIR/sites.json" | grep '^codemod:'; fi
	census "$tree" "$TMPDIR/$variant.jsonl"
	python3 "$here/compare.py" "parser.ts ($variant)" "$TMPDIR/none.jsonl" "$TMPDIR/$variant.jsonl" | grep -E '^(census.*paired|removed|added)' | cut -c1-200
done
echo "== end"
