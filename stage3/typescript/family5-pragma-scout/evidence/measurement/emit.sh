#!/usr/bin/env bash
set -euo pipefail
d=/workspace/scratch/family5-measure
tsc=$d/main/stage3/api/node_modules/typescript/lib/tsc.js
e=$d/report/stage3/typescript/family5-pragma-scout/evidence/measurement
for label in before after; do
    cd "$d/$label"
    node "$tsc" -p src/compiler --noCheck --emitDeclarationOnly false --outDir "$d/$label-js" --tsBuildInfoFile "$d/$label-js.tsbuildinfo" --pretty false
    node "$tsc" -b src/typescript --noCheck --force --pretty false
    node "$d/upstream/scripts/dtsBundler.mjs" --entrypoint "$PWD/built/local/typescript/typescript.d.ts" --output "$PWD/built/local/typescript.d.ts"
done
diff -qr "$d/before/src" "$d/after/src" > "$d/trial-files.txt" || test "$?" = 1
for f in compiler/parser.js compiler/types.js; do
    if cmp -s "$d/before-js/$f" "$d/after-js/$f"; then
        printf '%s byte-identical\n' "$f"
    else
        printf '%s differs\n' "$f"
        cmp "$d/before-js/$f" "$d/after-js/$f" || test "$?" = 1
    fi
done > "$e/javascript-cmp.txt"
for label in before after; do
    (cd "$d/$label-js" && sha256sum compiler/parser.js compiler/types.js) | sed "s/  /  $label\//"
done > "$e/javascript.sha256"
for label in before after; do
    (cd "$d/$label/built/local" && sha256sum typescript.d.ts) | sed "s/  /  $label\//"
done > "$e/public-dts.sha256"
diff -u --label before/typescript.d.ts --label after/typescript.d.ts "$d/before/built/local/typescript.d.ts" "$d/after/built/local/typescript.d.ts" > "$e/public-dts.diff" || test "$?" = 1
if cmp -s "$d/before/built/local/typescript.d.ts" "$d/after/built/local/typescript.d.ts"; then
    printf 'typescript.d.ts byte-identical\n'
else
    printf 'typescript.d.ts differs\n'
fi > "$e/public-dts-cmp.txt"
