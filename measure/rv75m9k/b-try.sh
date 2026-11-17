#!/usr/bin/env bash
# #rv75m9k road B, first look: on the tree apply builds with road A installed, scan the method value reads left, run the
# road-B codemod over them, then scan again, check stock tsc, and print the emitted JavaScript's difference for every
# changed file (transpileModule, ES2020, CommonJS). Sealed: it sources the image's environment itself.
set -uo pipefail
source /opt/adamic-block/environment.sh && source cloud/sealed-job.sh || { echo "FAIL sealed environment"; exit 1; }
here=$(cd "$(dirname "$0")" && pwd)
sealed_typescript_source >/dev/null && sealed_npm stage3/api >/dev/null && echo "PASS sealed inputs" || { echo "FAIL sealed inputs"; exit 1; }
export STAGE3_PRODUCT_STORE="$TMPDIR/store" STAGE3_CACHE="$TMPDIR/stage3-cache" NODE_PATH="$PWD/stage3/api/node_modules"
tree="$TMPDIR/tree"
bash stage3/apply.sh --build-product "$tree" > "$TMPDIR/apply.log" 2>&1 && echo "PASS adapted tree built" || { echo "FAIL apply"; tail -30 "$TMPDIR/apply.log"; exit 1; }
node "$here/scan.cjs" "$tree" "$here/names.json" "$TMPDIR/b-sites.json" | grep -E '^scan'
cp -p "$tree"/src/compiler/*.ts "$TMPDIR/" 2>/dev/null; ( cd "$tree" && git stash list >/dev/null; git -C "$tree" add -A src >/dev/null 2>&1; git -C "$tree" -c user.name=job -c user.email=job@loom.invalid commit -qm before-b >/dev/null 2>&1 ) || true
node "$here/codemod-b.cjs" "$tree" "$TMPDIR/b-sites.json" && echo "PASS codemod" || { echo "FAIL codemod"; exit 1; }
node "$here/scan.cjs" "$tree" "$here/names.json" "$TMPDIR/b-after.json" | grep -vE '^(A |written)'
echo "== source difference"
git -C "$tree" diff --stat | tail -3
git -C "$tree" diff -U0 | grep -E '^[-+][^-+]' | cut -c1-260
echo "== emitted JavaScript difference"
for file in $(git -C "$tree" diff --name-only); do
	git -C "$tree" show "HEAD:$file" > "$TMPDIR/b.ts"
	node -e 'const ts = require("typescript"), fs = require("node:fs"); for (const [i, o] of [[process.argv[1], process.argv[2]], [process.argv[3], process.argv[4]]]) fs.writeFileSync(o, ts.transpileModule(fs.readFileSync(i, "utf8"), { compilerOptions: { target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.CommonJS } }).outputText)' "$TMPDIR/b.ts" "$TMPDIR/b.js" "$tree/$file" "$TMPDIR/a.js"
	echo "-- $file: $(diff "$TMPDIR/b.js" "$TMPDIR/a.js" | grep -c '^[<>]') changed lines"
	diff "$TMPDIR/b.js" "$TMPDIR/a.js" | grep '^[<>]' | cut -c1-220
done
echo "== end"
