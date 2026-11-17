#!/usr/bin/env bash
# #rv75m9k road A on the adapted tree at this commit: scan (before), pin owners, adapt, scan again (after), replay (zero
# writes), restore with the mutant and check every touched file is back byte for byte. Prints owners.json whole at the
# end, so the adaptation's pins come from the tree the box built, never from a Mac's upstream copy.
# Sealed: it sources the image's environment and cloud/sealed-job.sh itself, since their helpers are shell functions.
set -uo pipefail
source /opt/adamic-block/environment.sh && source cloud/sealed-job.sh || { echo "FAIL sealed environment"; exit 1; }
here=$(cd "$(dirname "$0")" && pwd)
sealed_typescript_source >/dev/null && echo "PASS typescript source" || { echo "FAIL typescript source"; exit 1; }
sealed_npm stage3/api >/dev/null && echo "PASS stage3/api installed" || { echo "FAIL stage3/api"; exit 1; }
export STAGE3_PRODUCT_STORE="$TMPDIR/store" STAGE3_CACHE="$TMPDIR/stage3-cache" NODE_PATH="$PWD/stage3/api/node_modules"
tree="$TMPDIR/tree"
bash stage3/apply.sh --build-product "$tree" > "$TMPDIR/apply.log" 2>&1 && echo "PASS adapted tree built" || { echo "FAIL apply"; tail -30 "$TMPDIR/apply.log"; exit 1; }
echo "== before"
node "$here/scan.cjs" "$tree" "$here/names.json" "$TMPDIR/before.json" | grep -v '^A ' || { echo "FAIL scan before"; exit 1; }
work="$TMPDIR/adaptation"; mkdir -p "$work"; cp "$here/adapt.cjs" "$here/mutant.cjs" "$work/"
node "$here/owners.cjs" "$tree" "$TMPDIR/before.json" "$work/owners.json" > "$TMPDIR/owners.log" || { echo "FAIL owners"; exit 1; }
grep -E '^(owners|refused)' "$TMPDIR/owners.log"
files=$(node -e 'console.log([...new Set(require(process.argv[1]).map(o => o.file))].join(" "))' "$work/owners.json")
( cd "$tree" && sha256sum $files ) > "$TMPDIR/files-before.sha256"
echo "== adapt"
node "$work/adapt.cjs" "$tree" && echo "PASS adapt" || { echo "FAIL adapt"; exit 1; }
echo "== after"
node "$here/scan.cjs" "$tree" "$here/names.json" "$TMPDIR/after.json" | grep -v '^A ' || { echo "FAIL scan after"; exit 1; }
echo "== replay"
node "$work/adapt.cjs" "$tree" | tr -d '\n ' ; echo
echo "== restore"
node "$work/mutant.cjs" "$tree" > /dev/null && ( cd "$tree" && sha256sum -c --quiet "$TMPDIR/files-before.sha256" ) && echo "PASS mutant restores every touched file" || echo "FAIL mutant restore"
echo "== owners.log"
cat "$TMPDIR/owners.log"
echo "== owners.json base64"
gzip -9c "$work/owners.json" | base64 -w0; echo
echo "== end"
