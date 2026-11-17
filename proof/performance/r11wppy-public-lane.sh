# #r11wppy item 1's public change re-cut on main: regenerate sanctioned-api.json and shard-plan.json's runtime,
# print the patch, then run stage3/lane. Run after environment.sh, sealed-job.sh, sealed_typescript_source and
# sealed_npm stage3/api stage3/typescript. Never merged.
set -u
NODE_PATH="$PWD/stage3/api/node_modules" python3 stage3/lane/generate-api.py "$ADAMIC_TYPESCRIPT_SOURCE" || { echo "VERDICT FAIL generate-api"; exit 3; }
( cd stage3/lane && python3 -c "import json, shard_plan; p = shard_plan.load_plan(); p['runtime'] = shard_plan.runtime_hash(); (shard_plan.LANE / 'shard-plan.json').write_text(json.dumps(p, indent=2) + '\n')" ) || { echo "VERDICT FAIL shard runtime"; exit 3; }
echo "=== LANE PATCH BEGIN"; git diff -- stage3/lane/sanctioned-api.json stage3/lane/shard-plan.json; echo "=== LANE PATCH END"
go test -count=1 -timeout 30m -run '^TestProduct_Stage3AdaptedTree$' ./stage3/lane/adaptedtree/ > "$TMPDIR/adapted.log" 2>&1; echo "adapted tree exit $?"
go test -count=1 -timeout 30m -skip 'TestShard_|TestProduct_' ./stage3/lane/... > "$TMPDIR/lane.log" 2>&1; code=$?; echo "lane exit $code"
grep -E '^(--- FAIL|FAIL|.*_test.(go|py):[0-9]+)' "$TMPDIR/lane.log" "$TMPDIR/adapted.log" | head -10 | cut -c1-260
echo "VERDICT DONE"
