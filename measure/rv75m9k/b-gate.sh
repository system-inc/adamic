#!/usr/bin/env bash
# #rv75m9k road B's census gate: on the tree apply builds with road A installed, each named file's full latent census
# (LATENT_FULL=1, LATENT_ONLY_FILE) before and after the road-B codemod, paired by compare.py. With --whole it also
# checks `adamic types` over the 79 roots and stock tsc, before and after. Sealed: it sources the image's environment.
# Usage: bash b-gate.sh [--whole] <src/compiler/file.ts>...
set -uo pipefail
source /opt/adamic-block/environment.sh && source cloud/sealed-job.sh || { echo "FAIL sealed environment"; exit 1; }
here=$(cd "$(dirname "$0")" && pwd)
whole=0; [ "${1:-}" = --whole ] && { whole=1; shift; }
sealed_typescript_source >/dev/null && sealed_npm stage3/api >/dev/null && echo "PASS sealed inputs" || { echo "FAIL sealed inputs"; exit 1; }
export STAGE3_PRODUCT_STORE="$TMPDIR/store" STAGE3_CACHE="$TMPDIR/stage3-cache" NODE_PATH="$PWD/stage3/api/node_modules"
tree="$TMPDIR/tree"
bash stage3/apply.sh --build-product "$tree" > "$TMPDIR/apply.log" 2>&1 && echo "PASS adapted tree built" || { echo "FAIL apply"; tail -30 "$TMPDIR/apply.log"; exit 1; }
overlay="$TMPDIR/overlay"
{ python3 stage3/census/latent/make_overlay.py "$PWD" "$overlay" && gofmt -w "$overlay"/*.go && go build -buildvcs=false -overlay="$overlay/overlay.json" -o "$TMPDIR/latent-census" ./stage3/census/latent/tool; } > "$TMPDIR/census-build.log" 2>&1 && echo "PASS census tool built" || { echo "FAIL census tool"; tail -20 "$TMPDIR/census-build.log"; exit 1; }
[ "$whole" = 1 ] && { go build -buildvcs=false -o "$TMPDIR/adamic" ./cmd/adamic > "$TMPDIR/adamic-build.log" 2>&1 && echo "PASS adamic built" || { echo "FAIL adamic build"; exit 1; }; }
roots=$(cd "$tree" && ls src/compiler/*.ts src/compiler/*/*.ts src/compiler/*/*/*.ts 2>/dev/null | grep -v '\.d\.ts$')
files="$*"
measure() {
	local label=$1 file
	for file in $files; do
		LATENT_FULL=1 LATENT_ONLY_FILE="$tree/$file" LATENT_ASSERT_NO_OUTPUT=1 timeout 2400 "$TMPDIR/latent-census" "$tree/src/compiler" "$TMPDIR/$label.$(echo "$file" | tr / _).jsonl" > "$TMPDIR/$label.$(echo "$file" | tr / _).log" 2>&1
		echo "census $label $file exit=$?"
	done
	if [ "$whole" = 1 ]; then
		( cd "$tree" && "$TMPDIR/adamic" types $roots ) > /dev/null 2> "$TMPDIR/$label.types.err"
		echo "types $label exit=$? diagnostics=$(wc -l < "$TMPDIR/$label.types.err")"
	fi
}
echo "== before"; measure before
node "$here/scan.cjs" "$tree" "$here/names.json" "$TMPDIR/b-sites.json" | grep -E '^scan'
node "$here/codemod-b.cjs" "$tree" "$TMPDIR/b-sites.json" | grep -E '^(codemod:|refused)' || { echo "FAIL codemod"; exit 1; }
node "$here/scan.cjs" "$tree" "$here/names.json" "$TMPDIR/b-after.json" | grep -E '^(scan|diagnostic)'
echo "== after"; measure after
echo "== compare"
for file in $files; do base="$(echo "$file" | tr / _).jsonl"; python3 "$here/compare.py" "$file" "$TMPDIR/before.$base" "$TMPDIR/after.$base" | grep -v '^sample'; done
[ "$whole" = 1 ] && { sort "$TMPDIR/before.types.err" > "$TMPDIR/b.err"; sort "$TMPDIR/after.types.err" > "$TMPDIR/a.err"; cmp -s "$TMPDIR/b.err" "$TMPDIR/a.err" && echo "PASS adamic types diagnostics identical ($(wc -l < "$TMPDIR/a.err") lines)" || { echo "DIFF adamic types:"; diff "$TMPDIR/b.err" "$TMPDIR/a.err" | head -40; }; }
echo "== end"
