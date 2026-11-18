# #0b9xqhh's census, stage 1, at the refusal scan: rulings.md:72 spelled at every family 5 relation.
# A measurement on a prototype head (census/apple-72), never landed. Asked sealed through the door:
#   source /opt/adamic-block/environment.sh && bash stage3/census/apple-72/run1.sh
source cloud/sealed-job.sh && sealed_npm stage3/api && sealed_typescript_source || { echo "FAIL setup"; exit 66; }
set -u
R=$PWD; W="$TMPDIR/c72"; C="$R/stage3/census/apple-72"; mkdir -p "$W"
export NODE_PATH="$R/stage3/api/node_modules"
copy() { python3 -I -c 'import shutil, sys; shutil.copytree(sys.argv[1], sys.argv[2], symlinks=True)' "$1" "$2"; }
bash stage3/apply.sh --build-product "$W/tree-a" > "$W/apply.log" 2>&1 || { tail -30 "$W/apply.log"; echo "FAIL apply"; exit 1; }
echo "PASS apply"
go test -count=1 -run '^TestProduct_ScanAllInstrumentedProbe$' -v ./stage3/census/scan-all/ > "$W/probe.log" 2>&1
P=$(sed -n 's/.*scan-all-instrumented probe: //p' "$W/probe.log" | tail -1); [ -x "$P" ] || { tail -30 "$W/probe.log"; echo "FAIL probe"; exit 1; }
echo "PASS probe"
scan() { env "$1" "$P" "$2/src/tsc/tsc.ts" > "$3" 2> "$3.err" || { echo "FAIL scan $3"; tail -8 "$3.err"; head -c 3000 "$3"; exit 1; }; }
scan ADAMIC_CENSUS_72=0 "$W/tree-a" "$W/off.json"
scan ADAMIC_CENSUS_72=1 "$W/tree-a" "$W/on.json"
echo "PASS scans of the plain tree"
# Rewrite, check with stock tsc, exclude what broke it, and again, until the rewritten tree is clean.
clean=0
for round in 1 2 3 4; do
  [ -d "$W/tree-b" ] && python3 -I -c 'import shutil, sys; shutil.rmtree(sys.argv[1])' "$W/tree-b"
  copy "$W/tree-a" "$W/tree-b"
  node "$C/rewrite72.cjs" "$W/tree-b" "$W/off.json" "$W/sites.json" "$W/exclude.json" || { echo "FAIL rewrite round $round"; exit 1; }
  if node "$C/diag72.cjs" "$W/tree-a" "$W/tree-b" "$W/sites.json" "$W/exclude.json"; then clean=1; echo "PASS rewritten tree clean at round $round"; break; fi
done
[ "$clean" = 1 ] || { echo "FAIL the rewritten tree adds stock diagnostics after four rounds"; exit 1; }
scan ADAMIC_CENSUS_72=1 "$W/tree-b" "$W/b.json"
echo "PASS scan of the rewritten tree"
node "$C/classify72.cjs" "$W/tree-b" "$W/sites.json" "$W/off.json" "$W/on.json" "$W/b.json" "$W/out.json" || { echo "FAIL classify"; exit 1; }
python3 -I - "$W/out.json" "$W/sites.json" <<'PYTHON'
import json, sys, collections
out, sites = json.load(open(sys.argv[1])), json.load(open(sys.argv[2]))
files = collections.Counter(s['relative'] for s in sites)
print('census: family 5 sites by file: ' + ', '.join(f'{f.split("/")[-1]} {n}' for f, n in files.most_common()))
print('census: rewritten sites by file: ' + ', '.join(f'{f.split("/")[-1]} {n}' for f, n in collections.Counter(s['relative'] for s in sites if s.get('rewritten')).most_common()))
for s in out['sites']:
    print(f"site {s['key']} [{s['unit'].split('#')[-1]}] as ({s['slot'][:70]}) next: {' || '.join(n[:150] for n in s['next'][:3]) or '-'}")
PYTHON
echo DONE
