#!/usr/bin/env bash
# #rv75m9k road B's adaptation, built on the tree apply makes with road A installed: the codemod writes the sites, then
# gen-owners-b.cjs pins every changed owner for the receiver printer's engine. On a fresh copy of the tree the engine must
# reproduce the codemod's files byte for byte, write nothing on replay, and its mutant must restore the copy; stock tsc
# reads 0. owners.json is printed gzipped and base64 at the end. Sealed: it sources the image's environment itself.
set -uo pipefail
source /opt/adamic-block/environment.sh && source cloud/sealed-job.sh || { echo "FAIL sealed environment"; exit 1; }
here=$(cd "$(dirname "$0")" && pwd)
sealed_typescript_source >/dev/null && sealed_npm stage3/api >/dev/null && echo "PASS sealed inputs" || exit 1
export STAGE3_PRODUCT_STORE="$TMPDIR/store" STAGE3_CACHE="$TMPDIR/stage3-cache" NODE_PATH="$PWD/stage3/api/node_modules"
tree="$TMPDIR/tree"
bash stage3/apply.sh --build-product "$tree" > "$TMPDIR/apply.log" 2>&1 && echo "PASS tree" || { tail -20 "$TMPDIR/apply.log"; exit 1; }
fresh="$TMPDIR/fresh"; cp -R "$tree" "$fresh"
before="$TMPDIR/before-src"; mkdir -p "$before"; cp -R "$tree/src/compiler/." "$before/"
node "$here/scan.cjs" "$tree" "$here/names.json" "$TMPDIR/sites.json" | grep -E '^scan'
node "$here/codemod-b.cjs" "$tree" "$TMPDIR/sites.json" | grep -E '^(codemod:|left|refused)' || { echo "FAIL codemod"; exit 1; }
work="$TMPDIR/adaptation"; mkdir -p "$work"; cp "$here/engine-adapt.cjs" "$work/adapt.cjs"; cp "$here/engine-mutant.cjs" "$work/mutant.cjs"
node "$here/gen-owners-b.cjs" "$before" "$tree" "$work/owners.json" | cut -c1-200 || { echo "FAIL owners"; exit 1; }
node "$work/adapt.cjs" "$fresh" | tr -d '\n '; echo
same=0; differ=0; for f in $(cd "$tree" && find src/compiler -name '*.ts'); do if cmp -s "$tree/$f" "$fresh/$f"; then same=$((same + 1)); else differ=$((differ + 1)); echo "differs $f"; fi; done
[ "$differ" = 0 ] && echo "PASS the engine reproduces the codemod in all $same files" || echo "FAIL $differ files differ"
echo "replay: $(node "$work/adapt.cjs" "$fresh" | tr -d '\n ')"
node "$here/scan.cjs" "$fresh" "$here/names.json" "$TMPDIR/after.json" | grep -E '^(scan|diagnostic)'
node "$work/mutant.cjs" "$fresh" > /dev/null; restored=0; for f in $(cd "$before" && find . -name '*.ts'); do cmp -s "$before/$f" "$fresh/src/compiler/$f" || { restored=1; echo "not restored $f"; }; done
[ "$restored" = 0 ] && echo "PASS the mutant restores every file" || echo "FAIL mutant"
echo "== owners.json base64"; gzip -9c "$work/owners.json" | base64 -w0; echo
echo "== end"
