# #0b9xqhh's census, stage 2, in the latent census: each rewritten site's unit, lowered whole or its stop named.
# A measurement on a prototype head (census/apple-72), never landed. Asked sealed through the door:
#   source /opt/adamic-block/environment.sh && bash stage3/census/apple-72/run2.sh MODE GROUP
# MODE A is the plain tree with the prototype off, B the rewritten tree with it on. GROUP is checker (checker.ts)
# or rest (every other file with a rewritten site).
MODE=${1:?MODE}; GROUP=${2:?GROUP}
source cloud/sealed-job.sh && sealed_npm stage3/api && sealed_typescript_source || { echo "FAIL setup"; exit 66; }
set -u
R=$PWD; W="$TMPDIR/c72"; C="$R/stage3/census/apple-72"; mkdir -p "$W"
export NODE_PATH="$R/stage3/api/node_modules"
copy() { python3 -I -c 'import shutil, sys; shutil.copytree(sys.argv[1], sys.argv[2], symlinks=True)' "$1" "$2"; }
bash stage3/apply.sh --build-product "$W/tree-a" > "$W/apply.log" 2>&1 || { tail -30 "$W/apply.log"; echo "FAIL apply"; exit 1; }
echo "PASS apply"
go test -count=1 -run '^TestProduct_ScanAllInstrumentedProbe$' -v ./stage3/census/scan-all/ > "$W/probe.log" 2>&1
P=$(sed -n 's/.*scan-all-instrumented probe: //p' "$W/probe.log" | tail -1); [ -x "$P" ] || { tail -30 "$W/probe.log"; echo "FAIL probe"; exit 1; }
ADAMIC_CENSUS_72=0 "$P" "$W/tree-a/src/tsc/tsc.ts" > "$W/off.json" 2> "$W/off.err" || { tail -8 "$W/off.err"; head -c 3000 "$W/off.json"; echo "FAIL scan"; exit 1; }
echo "PASS scan of the plain tree"
clean=0
for round in 1 2 3 4; do
  [ -d "$W/tree-b" ] && python3 -I -c 'import shutil, sys; shutil.rmtree(sys.argv[1])' "$W/tree-b"
  copy "$W/tree-a" "$W/tree-b"
  node "$C/rewrite72.cjs" "$W/tree-b" "$W/off.json" "$W/sites.json" "$W/exclude.json" || { echo "FAIL rewrite round $round"; exit 1; }
  if node "$C/diag72.cjs" "$W/tree-a" "$W/tree-b" "$W/sites.json" "$W/exclude.json" > "$W/diag-$round.log"; then clean=1; echo "PASS rewritten tree clean at round $round"; break; fi
done
[ "$clean" = 1 ] || { echo "FAIL the rewritten tree adds stock diagnostics after four rounds"; exit 1; }
python3 stage3/census/latent/make_overlay.py "$R" "$W/overlay" > "$W/overlay.log" 2>&1 || { tail -20 "$W/overlay.log"; echo "FAIL overlay"; exit 1; }
gofmt -w "$W"/overlay/*.go
go build -buildvcs=false -overlay="$W/overlay/overlay.json" -o "$W/latent" ./stage3/census/latent/tool > "$W/latent-build.log" 2>&1 || { tail -30 "$W/latent-build.log"; echo "FAIL latent build"; exit 1; }
echo "PASS latent build"
if [ "$MODE" = A ]; then T="$W/tree-a"; SWITCH=0; else T="$W/tree-b"; SWITCH=1; fi
FILES=$(python3 -I - "$W/sites.json" "$T" "$GROUP" <<'LIST'
import json, sys, os
sites, tree, group = json.load(open(sys.argv[1])), sys.argv[2], sys.argv[3]
files = sorted({s['relative'] for s in sites if s.get('rewritten')})
files = [f for f in files if (f == 'src/compiler/checker.ts') == (group == 'checker')]
print('\n'.join(os.path.join(tree, f) for f in files))
LIST
)
echo "latent over $(echo "$FILES" | wc -l) files: $(echo "$FILES" | xargs -n1 basename | tr '\n' ' ')"
start=$(date +%s)
ADAMIC_CENSUS_72=$SWITCH LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_ONLY_FILE="$FILES" "$W/latent" "$T/src/compiler" "$W/latent.jsonl" > "$W/latent-run.log" 2>&1
echo "latent exit=$? in $(( $(date +%s) - start ))s"; tail -5 "$W/latent-run.log"
node "$C/unit72.cjs" "$MODE" "$T" "$W/sites.json" "$W/latent.jsonl" || { echo "FAIL units"; exit 1; }
echo DONE
