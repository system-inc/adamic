# #3wv0ykh's probe, sealed: the whole tsc entry census as the meter runs it (full mode, the scoped tail on) at main, with ADAMIC_PROBE_NEVER_SOURCE set,
# so every optional-widening refusal whose source prints as never writes its site, the source's flags before and after
# TypeToString, the intersection's parts and the site's own type. Run after environment.sh, sealed-job.sh,
# sealed_typescript_source and sealed_npm stage3/api. Never merged.
set -u
W=$(pwd)
O=$TMPDIR/3wv0ykh-probe
mkdir -p "$O"
fail() { echo "VERDICT FAIL $1 (logs in $O)"; tail -5 "$O/$2" 2>/dev/null; exit 3; }
[ -n "${ADAMIC_TYPESCRIPT_SOURCE:-}" ] && [ -d stage3/api/node_modules ] || fail "the sealed TypeScript source or stage3/api install is missing" env.log
bash stage3/apply.sh "$O/adapted" > "$O/apply.log" 2>&1 || fail "stage3/apply.sh" apply.log
side() {
side=$1
mkdir -p "$O/$side"
python3 stage3/census/latent/make_overlay.py "$W" "$O/$side/latent" > "$O/latent-overlay.log" 2>&1 || fail "latent overlay" latent-overlay.log
gofmt -w "$O/$side/latent/"*.go
python3 stage3/meter/entry_overlay.py "$O/$side/latent" "$O/$side/entry" > "$O/entry-overlay.log" 2>&1 || fail "entry overlay" entry-overlay.log
# The profile: a file in package main that starts the CPU profile at init and stops it when main returns.
cat > "$O/$side/entry/profile.go" <<'GO'
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
python3 - "$O/$side/entry" "$W" <<'PYTHON' || fail "profile overlay" entry-overlay.log
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
gofmt -w "$O/$side/entry/"*.go
go build -buildvcs=false -overlay="$O/$side/entry/overlay.json" -o "$O/$side/entry-census" ./stage3/census/latent/tool > "$O/build.log" 2>&1 || fail "entry census build" build.log
echo "built; $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
target=$O/adapted/src/compiler/utilitiesPublic.ts
[ -f "$target" ] || fail "no utilitiesPublic.ts in the adapted tree" apply.log
start=$(date +%s)
# As the daily meter runs its tsc entry pass (twice-daily.sh:113): full mode, the scoped tail on.
only=; full=1; bound=3500
LATENT_ONLY_FILE=$only LATENT_TAIL=1 LATENT_FULL=$full LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES=$O/$side/times.jsonl LATENT_CPUPROFILE=$O/$side/cpu.pprof ADAMIC_PROBE_NEVER_SOURCE=$O/probe.txt \
    timeout --kill-after=10s $bound "$O/$side/entry-census" "$O/adapted/src/tsc/tsc.ts" "$O/$side/latent.jsonl" > "$O/$side/run.log" 2>&1
code=$?
echo "$side census exit $code after $(( $(date +%s) - start )) s; load $(cut -d' ' -f1-3 /proc/loadavg)"
[ -s "$O/$side/cpu.pprof" ] || fail "no profile written at $side (census exit $code)" $side/run.log
python3 - "$O/$side/times.jsonl" "$side" <<'PYTHON'
import json, sys
rows = [json.loads(l) for l in open(sys.argv[1])]
up = [r for r in rows if 'utilitiesPublic.ts:' in r['where']]
w = sorted(r['wall_seconds'] for r in up)
print(f"{sys.argv[2]}: units {len(rows)}, utilitiesPublic {len(up)}, sum {sum(w):.1f} s, median {w[len(w)//2] if w else 0:.2f} s")
for line in (1550, 1536, 1416, 870, 1422):
    hit = [r for r in up if r['where'].endswith(f'.ts:{line}:1')]
    print(f"  line {line}: " + (f"{hit[0]['wall_seconds']:.2f} s wall, {hit[0].get('cpu_seconds', 0):.2f} s cpu" if hit else "no unit"))
PYTHON
}
echo "probe on $(hostname), $(nproc) cores"
side head
P=$O/probe.txt
[ -f "$P" ] || { echo "VERDICT FAIL the probe wrote nothing"; exit 3; }
echo "probe lines: $(wc -l < "$P"), distinct sites: $(cut -d'|' -f1 "$P" | sort -u | wc -l)"
echo "== before/after patterns, by count"
awk -F' [|] ' '{print $6 " -> " $7}' "$P" | sort | uniq -c | sort -rn | head
echo "== site kinds"
awk -F' [|] ' '{print $2}' "$P" | sort | uniq -c | sort -rn | head
echo "== distinct parts, by count"
awk -F' [|] ' '{print $8}' "$P" | sort | uniq -c | sort -rn | head -12 | cut -c1-300
echo "== the first site of each distinct parts list"
sort -t'|' -k8,8 -u "$P" | head -12 | cut -c1-600
echo "VERDICT OK"
