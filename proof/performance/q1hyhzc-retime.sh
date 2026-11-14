# #q1hyhzc's re-time of #40hkx9g's predicates: the latent entry census on utilitiesPublic.ts alone (full mode,
# LATENT_ONLY_FILE), at the head and then at its main base, on one box, one after the other. Built from
# proof/performance/pprof-40hkx9g.sh (8060440a). Never merged.
set -u
W=$(pwd)
O=$HOME/scratch/system_adamic_compiler_performance/q1hyhzc-retime-$(git rev-parse --short HEAD)-$(date +%s)
mkdir -p "$O"
fail() { echo "VERDICT FAIL $1 (logs in $O)"; tail -5 "$O/$2" 2>/dev/null; exit 3; }
. "$HOME/adamic-tools/env.sh" > "$O/env.log" 2>&1 || fail "the box's adamic-tools env" env.log
git -c protocol.file.allow=always submodule update --init --recursive --depth 1 > "$O/submodules.log" 2>&1 || fail submodules submodules.log
read -r url tag <<< "$(python3 -c 'import json; s=json.load(open("stage3/source.json")); print(s["repository"], s["tag"])')"
git clone -q --depth 1 --branch "$tag" "$url" "$O/upstream" > "$O/upstream.log" 2>&1 || fail "TypeScript $tag clone" upstream.log
export ADAMIC_TYPESCRIPT_SOURCE=$O/upstream
if ! command -v npm > /dev/null 2>&1; then
    b=cloud/markdown-width/npm-bootstrap.json
    n=$O/npm; mkdir -p "$n/bin"
    curl -fsSL "$(python3 -I -c 'import json,sys; print(json.load(open(sys.argv[1]))["url"])' $b)" -o "$n/npm.tgz" || fail "npm download" env.log
    python3 -I - "$b" "$n/npm.tgz" <<'PYTHON' || fail "npm integrity" env.log
import base64, hashlib, json, sys
algorithm, expected = json.load(open(sys.argv[1]))['integrity'].split('-', 1)
actual = base64.b64encode(hashlib.sha512(open(sys.argv[2], 'rb').read()).digest()).decode()
sys.exit(0 if algorithm == 'sha512' and actual == expected else 'npm bootstrap integrity mismatch')
PYTHON
    tar --no-same-owner -xzf "$n/npm.tgz" -C "$n"
    printf '#!/bin/sh\nexec node %q "$@"\n' "$n/package/bin/npm-cli.js" > "$n/bin/npm"; chmod 755 "$n/bin/npm"
    export PATH="$n/bin:$PATH"
fi
if [ ! -d stage3/api/node_modules ]; then
    (cd stage3/api && npm ci --ignore-scripts --no-audit --no-fund) > "$O/packages.log" 2>&1 || fail "npm ci in stage3/api" packages.log
fi
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
LATENT_ONLY_FILE=$target LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES=$O/$side/times.jsonl LATENT_CPUPROFILE=$O/$side/cpu.pprof \
    timeout --kill-after=10s 1800 "$O/$side/entry-census" "$O/adapted/src/tsc/tsc.ts" "$O/$side/latent.jsonl" > "$O/$side/run.log" 2>&1
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
head=$(git log --format=%H -1 --invert-grep --grep='^proof: ' HEAD)
base=5d334653fa
echo "head $head base $(git rev-parse --short=12 $base) on $(hostname), $(nproc) cores"
side head
git diff --name-status "$base" "$head" | while read -r status path; do
  case $status in A) rm -f "$path" ;; M) git show "$base:$path" > "$path" ;; esac
done
[ ! -e internal/lower/view_sites.go ] && git diff --quiet "$base" -- internal/lower/lower.go stage3/census || { echo "VERDICT FAIL base not restored"; exit 3; }
side base
git checkout -q HEAD -- internal stage3
echo "VERDICT OK logs in $O"
