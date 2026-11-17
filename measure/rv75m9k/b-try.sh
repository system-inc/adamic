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
before="$TMPDIR/before-src"; mkdir -p "$before"; cp -R "$tree/src/compiler" "$before/"
node "$here/codemod-b.cjs" "$tree" "$TMPDIR/b-sites.json" && echo "PASS codemod" || { echo "FAIL codemod"; exit 1; }
node "$here/scan.cjs" "$tree" "$here/names.json" "$TMPDIR/b-after.json" | grep -vE '^(A |written)'
echo "== source difference"
changed=$(cd "$before" && for f in $(find compiler -name '*.ts'); do cmp -s "$f" "$tree/src/$f" || echo "src/$f"; done)
echo "changed files: $(echo $changed | wc -w)"
for file in $changed; do diff -U0 "$before/${file#src/}" "$tree/$file" | grep -E '^[-+][^-+]' | cut -c1-260; done
echo "== emitted JavaScript difference"
for file in $changed; do
	cp "$before/${file#src/}" "$TMPDIR/b.ts"
	node -e 'const ts = require("typescript"), fs = require("node:fs"); for (const [i, o] of [[process.argv[1], process.argv[2]], [process.argv[3], process.argv[4]]]) fs.writeFileSync(o, ts.transpileModule(fs.readFileSync(i, "utf8"), { compilerOptions: { target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.CommonJS } }).outputText)' "$TMPDIR/b.ts" "$TMPDIR/b.js" "$tree/$file" "$TMPDIR/a.js"
	echo "-- $file: $(diff "$TMPDIR/b.js" "$TMPDIR/a.js" | grep -c '^[<>]') changed lines"
	diff "$TMPDIR/b.js" "$TMPDIR/a.js" | grep '^[<>]' | cut -c1-220
done
echo "== end"
