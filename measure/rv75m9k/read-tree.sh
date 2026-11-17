#!/usr/bin/env bash
# #rv75m9k: builds stage3's adapted tree at this commit, then scans its src/compiler for methods read as values (scan.cjs),
# printing each road-A member's declaration as the adapted tree holds it and the stock checker's diagnostic count.
# Sealed: source /opt/adamic-block/environment.sh and cloud/sealed-job.sh first.
set -uo pipefail
here=$(cd "$(dirname "$0")" && pwd)
sealed_typescript_source && echo "PASS typescript source" || { echo "FAIL typescript source"; exit 1; }
sealed_npm stage3/api && echo "PASS stage3/api installed" || { echo "FAIL stage3/api"; exit 1; }
export STAGE3_PRODUCT_STORE="$TMPDIR/store" STAGE3_CACHE="$TMPDIR/stage3-cache"
tree="$TMPDIR/tree"
if bash stage3/apply.sh --build-product "$tree" > "$TMPDIR/apply.log" 2>&1; then echo "PASS adapted tree built"; else echo "FAIL apply"; tail -30 "$TMPDIR/apply.log"; exit 1; fi
NODE_PATH="$PWD/stage3/api/node_modules" node "$here/scan.cjs" "$tree" "$here/names.json" "$TMPDIR/sites.json" && echo "PASS scan" || { echo "FAIL scan"; exit 1; }
echo "sites, one a line:"
node -e 'for (const s of require(process.argv[1])) console.log(`site ${s.site} ${s.member} [${s.use}] ${s.internal ? "internal" : "public"} ${s.kinds.join(",")} :: ${s.text}`)' "$TMPDIR/sites.json"
