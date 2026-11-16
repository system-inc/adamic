# #v5b6pzp's probes, sealed: each program through Node, the JavaScript backend and a native build, at this commit's
# parent. Run after environment.sh and sealed-job.sh. Never merged.
set -u
D=$(cd "$(dirname "$0")" && pwd)
O=$TMPDIR/v5b6pzp
mkdir -p "$O"
go build -o "$O/adamic" ./cmd/adamic > "$O/build.log" 2>&1 || { echo "VERDICT FAIL adamic build"; tail -5 "$O/build.log"; exit 3; }
echo "on $(hostname), node $(node --version)"
for f in "$D"/*.ts; do
  name=$(basename "$f" .ts)
  node_out=$(node "$f" 2>&1 | tail -1)
  "$O/adamic" js "$f" > "$O/$name.js" 2> "$O/$name.js.err"; js_code=$?
  if [ $js_code = 0 ]; then js_out=$(node "$O/$name.js" 2>&1 | tail -1); else js_out="lowering exit $js_code: $(tail -1 "$O/$name.js.err" | cut -c1-220)"; fi
  "$O/adamic" build "$f" -o "$O/$name" > "$O/$name.build" 2>&1; build_code=$?
  if [ $build_code = 0 ]; then native_out=$("$O/$name" 2>&1 | tail -1); else native_out="build exit $build_code: $(grep -m1 -E 'error:|refuses|NotYet|not yet' "$O/$name.build" | cut -c1-220)"; fi
  echo "== $name"
  echo "   node:   $node_out"
  echo "   js:     $js_out"
  echo "   native: $native_out"
done
echo "VERDICT OK"
