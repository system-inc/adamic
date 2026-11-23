# #hjspb3b: with CENSUS_COMPARE=1, every census cache's compare mode on (placeholder slots, viewed elements,
# inheritance, accessor names): a hit re-walks and panics on any difference, which a unit records as a panic. Else the
# whole tsc entry census, first-error mode, at this tree, sealed: a digest of every finding row with its
# timing and relation-cache fields dropped, the cache sizes apart, and the census's own prepass times. Run after
# environment.sh, sandboxed-job.sh, sandboxed_typescript_source and sandboxed_npm stage3/api. Never merged.
set -u
W=$(pwd)
O=$TMPDIR/entry-profile
mkdir -p "$O"
fail() { echo "VERDICT FAIL $1"; tail -5 "$O/$2" 2>/dev/null; exit 3; }
bash stage3/apply.sh --build-product "$O/adapted" > "$O/apply.log" 2>&1 || fail "stage3/apply.sh" apply.log
python3 stage3/census/latent/make_overlay.py "$W" "$O/latent" > "$O/latent-overlay.log" 2>&1 || fail "latent overlay" latent-overlay.log
gofmt -w "$O/latent/"*.go
python3 stage3/meter/entry_overlay.py "$O/latent" "$O/entry" > "$O/entry-overlay.log" 2>&1 || fail "entry overlay" entry-overlay.log
# The profile: a file in package main that starts the CPU profile at init and stops it when main returns.
cat > "$O/entry/profile.go" <<'GO'
package main

import (
	"os"
	"runtime/pprof"
)

var latentProfile *os.File

func init() {
	if path := os.Getenv("LATENT_CPUPROFILE"); path != "" {
		file, err := os.Create(path)
		if err != nil {
			panic(err)
		}
		latentProfile = file
		if err := pprof.StartCPUProfile(file); err != nil {
			panic(err)
		}
	}
}

func latentStopProfile() {
	if latentProfile != nil {
		pprof.StopCPUProfile()
		latentProfile.Close()
	}
}
GO
python3 - "$O/entry" "$W" <<'PYTHON' || fail "profile overlay" entry-overlay.log
import json, pathlib, sys
entry, repository = pathlib.Path(sys.argv[1]), pathlib.Path(sys.argv[2])
main = entry / 'entry_main.go'
text = main.read_text()
assert text.count('func main() {') == 1
main.write_text(text.replace('func main() {', 'func main() {\n\tdefer latentStopProfile()', 1))
overlay = json.loads((entry / 'overlay.json').read_text())
overlay['Replace'][str(repository / 'stage3/census/latent/tool/profile.go')] = str(entry / 'profile.go')
(entry / 'overlay.json').write_text(json.dumps(overlay, indent=2) + '\n')
PYTHON
gofmt -w "$O/entry/"*.go
if [ "${CENSUS_COMPARE:-0}" = 1 ]; then
# Every census cache's compare flag this tree declares, found by name, so a cache added later (expansiveGenerics,
# #bmqm0k1) is compared without editing this script. viewWalkCompare is the view site index's, not a census cache's.
flags=$(grep -ho '^var [A-Za-z]*Compare atomic.Bool' $(ls internal/lower/*.go | grep -v '_test\.go$') | awk '{print $2}' | grep -v '^viewWalkCompare$' | sort)
{ echo 'package lower'; echo; echo '// The census caches answer as a fresh walk would, or the walk beside each hit panics (#hjspb3b).'; echo 'func init() {'; for flag in $flags; do printf '\t%s.Store(true)\n' "$flag"; done; echo '}'; } > "$O/entry/census_compare.go"
echo "compare flags: $(echo $flags)"
python3 - "$O/entry" "$W" <<'PYTHON' || fail "compare overlay" entry-overlay.log
import json, pathlib, sys
entry, repository = pathlib.Path(sys.argv[1]), pathlib.Path(sys.argv[2])
overlay = json.loads((entry / 'overlay.json').read_text())
overlay['Replace'][str(repository / 'internal/lower/census_compare.go')] = str(entry / 'census_compare.go')
(entry / 'overlay.json').write_text(json.dumps(overlay, indent=2) + '\n')
PYTHON
echo "compare modes on"
fi
go build -buildvcs=false -overlay="$O/entry/overlay.json" -o "$O/entry-census" ./stage3/census/latent/tool > "$O/build.log" 2>&1 || fail "entry census build" build.log
echo "built; $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
start=$(date +%s)
LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES=$O/times.jsonl LATENT_PREPASS_TIMES=$O/prepass.tsv \
    timeout --kill-after=10s 2400 "$O/entry-census" "$O/adapted/src/tsc/tsc.ts" "$O/latent.jsonl" > "$O/run.log" 2>&1
code=$?
echo "census exit $code after $(( $(date +%s) - start )) s; load $(cut -d' ' -f1-3 /proc/loadavg)"
python3 - "$O/latent.jsonl" "$O/times.jsonl" "$O/prepass.tsv" <<'PYTHON'
import hashlib, json, re, sys
drop = re.compile(r'(second|time|cache_size|duration|elapsed)', re.I)
def strip(value):
    if isinstance(value, dict):
        return {k: strip(v) for k, v in value.items() if not drop.search(k)}
    if isinstance(value, list):
        return [strip(v) for v in value]
    return value
rows, sizes = [], []
for line in open(sys.argv[1]):
    line = line.strip()
    if not line: continue
    row = json.loads(line)
    for key in ('cache_sizes', 'relation_overflow'):
        if isinstance(row, dict) and key in row: sizes.append(json.dumps(row[key], sort_keys=True)[:300])
    rows.append(json.dumps(strip(row), sort_keys=True))
rows.sort()
digest = hashlib.sha256('\n'.join(rows).encode()).hexdigest()
print(f"DIGEST rows {len(rows)} sha256 {digest}")
for s in sizes[:4]: print(f"SIZES {s}")
times = [json.loads(l) for l in open(sys.argv[2]) if l.strip()]
units = sorted(r.get('wall_seconds', 0) for r in times)
print(f"UNITS {len(times)} sum {sum(units):.1f} s")
import collections, os
calls, total = collections.Counter(), collections.Counter()
if os.path.exists(sys.argv[3]):
    for line in open(sys.argv[3]):
        parts = line.rstrip('\n').split('\t')
        if len(parts) == 3: calls[parts[0]] += 1; total[parts[0]] += int(parts[2])
for method in sorted(total, key=lambda m: -total[m]): print(f"PREPASS {method}: {calls[method]} calls, {total[method] / 1e9:.1f} s")
PYTHON
panics=$(grep -c "comparison (#" "$O/latent.jsonl" 2>/dev/null)
echo "COMPARE panics in findings: ${panics:-0}"
echo "VERDICT $([ $code = 0 ] && [ "${panics:-0}" = 0 ] && echo OK || echo FAIL) census exit $code"
