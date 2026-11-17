#!/usr/bin/env bash
# #rv75m9k, Lower's deciding run: parser.ts's census with eepPropertyPresence traced (RV75M9K_TRACE, evidence branch only),
# on the tree with road A installed and again with road B applied to parser.ts alone.
set -uo pipefail
source /opt/adamic-block/environment.sh && source cloud/sealed-job.sh || { echo "FAIL sealed environment"; exit 1; }
here=$(cd "$(dirname "$0")" && pwd)
sealed_typescript_source >/dev/null && sealed_npm stage3/api >/dev/null && echo "PASS sealed inputs" || exit 1
export STAGE3_PRODUCT_STORE="$TMPDIR/store" STAGE3_CACHE="$TMPDIR/stage3-cache" NODE_PATH="$PWD/stage3/api/node_modules"
base="$TMPDIR/base"
bash stage3/apply.sh --build-product "$base" > "$TMPDIR/apply.log" 2>&1 && echo "PASS tree" || { tail -20 "$TMPDIR/apply.log"; exit 1; }
overlay="$TMPDIR/overlay"
{ python3 stage3/census/latent/make_overlay.py "$PWD" "$overlay" && gofmt -w "$overlay"/*.go && go build -buildvcs=false -overlay="$overlay/overlay.json" -o "$TMPDIR/latent-census" ./stage3/census/latent/tool; } > "$TMPDIR/cb.log" 2>&1 && echo "PASS census tool" || { tail -20 "$TMPDIR/cb.log"; exit 1; }
node "$here/scan.cjs" "$base" "$here/names.json" "$TMPDIR/sites.json" > /dev/null
only="$TMPDIR/only"; cp -R "$base" "$only"
ROAD_B_ONLY=src/compiler/parser.ts node "$here/codemod-b.cjs" "$only" "$TMPDIR/sites.json" | grep '^codemod:'
for tree in "$base" "$only"; do
	label=$(basename "$tree")
	RV75M9K_TRACE=1 LATENT_FULL=1 LATENT_ONLY_FILE="$tree/src/compiler/parser.ts" LATENT_ASSERT_NO_OUTPUT=1 timeout 1500 "$TMPDIR/latent-census" "$tree/src/compiler" "$TMPDIR/$label.jsonl" > "$TMPDIR/$label.log" 2>&1
	echo "== $label census exit=$?: $(grep -c rv75m9k-trace "$TMPDIR/$label.log") traced reads"
	grep rv75m9k-trace "$TMPDIR/$label.log" | sed "s|$tree/||" | sort | uniq -c | sort -k3 | cut -c1-200
	echo "-- $label: the unit holding 3812, its records"
	python3 - "$TMPDIR/$label.jsonl" <<'PY'
import json, sys
rows = [json.loads(line) for line in open(sys.argv[1]) if line.strip()]
def walk(value):
    if isinstance(value, dict):
        if 'phase' in value and isinstance(value.get('unit'), str) and ':3' in value['unit'] and isinstance(value.get('where'), str):
            line = int(value['where'].rsplit(':', 2)[-2]) if value['where'].count(':') >= 2 else 0
            unit = int(value['unit'].rsplit(':', 2)[-2])
            if unit <= 3812 <= unit + 200 or 3790 <= line <= 3920:
                print(value['unit'].split('src/')[-1], value['where'].split('src/')[-1], value['kind'], value['phase'], value.get('reason', '')[:110])
        for child in value.values(): walk(child)
    elif isinstance(value, list):
        for child in value: walk(child)
for row in rows: walk(row)
PY
done
echo "== end"
