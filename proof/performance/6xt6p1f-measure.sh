# #6xt6p1f's measure, sealed: the latent entry census on utilitiesPublic.ts alone (full mode, LATENT_ONLY_FILE), at the
# head and then with a nil cache handed to the census's lowerings, which is main's walk on every lowering, on one box
# one after the other, verdicts compared. A sealed tree has no history, so the base is this toggle, not a checkout.
# Run after environment.sh, sealed-job.sh, sealed_typescript_source and sealed_npm stage3/api. Never merged.
set -u
W=$(pwd)
O=$TMPDIR/6xt6p1f-measure
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
if [ "${MODE:-file}" = entry ]; then only=; full=0; bound=3300; else only=$target; full=1; bound=1800; fi
LATENT_ONLY_FILE=$only LATENT_FULL=$full LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES=$O/$side/times.jsonl LATENT_CPUPROFILE=$O/$side/cpu.pprof \
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
echo "head (cache handed) against the cache withheld, on $(hostname), $(nproc) cores"
side head
# A nil cache is production's path: every lowering walks, as main does.
for f in stage3/census/latent/lower.go.txt stage3/census/latent/units.go.txt; do sed -i 's/placeholderSlots := &placeholderSlotCache{}/var placeholderSlots *placeholderSlotCache/' $f; done
[ "$(grep -c 'var placeholderSlots \*placeholderSlotCache' stage3/census/latent/lower.go.txt stage3/census/latent/units.go.txt | awk -F: '{s+=$2} END {print s}')" = 2 ] || { echo "VERDICT FAIL the cache isn't withheld"; exit 3; }
side base
# Where the placeholder walk stands in each side's profile, and how the census's units spent their CPU.
for side in head base; do
  echo "== $side: total and the placeholder functions, cumulative"
  go tool pprof -top -cum -nodecount=2000 "$O/$side/entry-census" "$O/$side/cpu.pprof" 2> /dev/null | grep -E "^ *Showing|Duration|[Pp]laceholder" | head -6 | cut -c1-160
done
# The census's verdicts, every field but the clocks, head against base.
python3 - "$O/head/latent.jsonl" "$O/base/latent.jsonl" <<'PYTHON'
import json, sys
def rows(path):
    out = []
    for line in open(path):
        row = json.loads(line)
        out.append(json.dumps({k: v for k, v in row.items() if not any(w in k.lower() for w in ("second", "wall", "cpu", "time", "duration", "rss", "memory", "byte"))}, sort_keys=True))
    return out
head, base = rows(sys.argv[1]), rows(sys.argv[2])
same = head == base
print(f"verdicts: {len(head)} head rows, {len(base)} base rows, {'identical' if same else 'DIFFER'}")
if not same:
    for index, (a, b) in enumerate(zip(head, base)):
        if a != b:
            print(f"  first difference at row {index}:\n    head {a[:300]}\n    base {b[:300]}")
            break
    sys.exit(1)
PYTHON
echo "VERDICT OK logs in $O"
