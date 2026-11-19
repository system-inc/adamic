#!/usr/bin/env bash
set -u
root=/tmp/adamic-g02
evidence=$root/evidence/stage3/census/family5-c02-sites
cd "$root/main" || exit 1
run() {
    local label=$1
    shift
    "$@" > "$evidence/$label.log" 2>&1
    local code=$?
    printf '%s\t%s\n' "$label" "$code" | tee -a "$evidence/exits.tsv"
    return "$code"
}
while [[ ! -f $root/base/patch-set.md ]]; do sleep 2; done
run scratch-copy cp -a --reflink=auto "$root/base" "$root/edit" || exit
run scratch-edit python3 "$evidence/edit.py" "$root/edit" || exit
for mode in project strict; do
    for side in base edit; do
        strict=0
        [[ $mode != strict ]] || strict=1
        run "stock-$side-$mode" env FAMILY5_STOCK_STRICT=$strict node "$root/stock.cjs" "$root/$side" "$evidence/stock-$side-$mode.json"
        mv "$evidence/stock-$side-$mode.json.js-tree" "$root/js-$side-$mode"
    done
    run "stock-js-cmp-$mode" diff -qr "$root/js-base-$mode" "$root/js-edit-$mode"
done
for side in base edit; do
    ln -s "$root/main/stage3/typescript/node_modules" "$root/$side/node_modules"
    run "public-build-$side" bash -c 'cd "$1" && node "$2" local' bash "$root/$side" "$root/main/stage3/typescript/node_modules/hereby/bin/hereby.js"
done
run public-api-cmp cmp "$root/base/built/local/typescript.d.ts" "$root/edit/built/local/typescript.d.ts"
run public-api-sha256 sha256sum "$root/base/built/local/typescript.d.ts" "$root/edit/built/local/typescript.d.ts"
