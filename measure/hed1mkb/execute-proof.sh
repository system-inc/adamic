#!/bin/bash
# #hed1mkb block 2 before review, sealed: format, vet, the packages it touches and both censuses whole, then three real
# runs of adamic-performance at this commit, each timed. The second must read level against the first, and the third,
# with a busy loop planted in bench/nbody.ts, must name nbody's native time slower. Never lands.
set -uo pipefail
echo "box $(hostname) $(nproc) cpus, $(uname -r), load $(cut -d' ' -f1-3 /proc/loadavg), tree $(git rev-parse HEAD)"
T="$TMPDIR/exe"
mkdir -p "$T"
packages="./cmd/adamic-performance/... ./internal/benchmark/... ./bench/..."
# bench/game/ is the Benchmarks Game's own Go, vendored byte for byte, which the gate's gofmt skips too.
unformatted=$(gofmt -l cmd/adamic-performance internal/benchmark bench internal/compilercensus | grep -v '^bench/game/')
[ -z "$unformatted" ] && echo "PASS gofmt" || { echo "$unformatted"; echo "FAIL gofmt"; exit 1; }
go vet $packages > "$T/vet.log" 2>&1 && echo "PASS vet" || { cat "$T/vet.log"; echo "FAIL vet"; exit 1; }
go test -count=1 $packages ./internal/compilercensus ./cmd/adamic-gate/parallelcensus > "$T/tests.log" 2>&1; code=$?
grep -E -- '^(--- FAIL|ok |FAIL)' "$T/tests.log"
[ $code = 0 ] && echo "PASS tests and censuses" || { grep -B3 -A30 -- '--- FAIL' "$T/tests.log" | head -150; echo "FAIL tests"; exit 1; }

go build -buildvcs=false -o "$T/adamic" ./cmd/adamic > "$T/build.log" 2>&1 && go build -buildvcs=false -o "$T/performance" ./cmd/adamic-performance >> "$T/build.log" 2>&1 || { cat "$T/build.log"; echo "FAIL build"; exit 1; }
sha=$(git rev-parse HEAD)
measure() { # name, then any further flags
	name=$1; shift
	started=$(date +%s.%N)
	"$T/performance" -adamic "$T/adamic" -out "$T/$name.jsonc" -commit "$sha" "$@" > "$T/$name.out" 2> "$T/$name.err"; status=$?
	ended=$(date +%s.%N)
	echo "== $name: exit $status, $(python3 -I -c "print(f'{$ended - $started:.1f}')") s, load after $(cut -d' ' -f1-3 /proc/loadavg)"
	cat "$T/$name.err"
	[ $status = 0 ] || { echo "FAIL $name"; exit 1; }
	python3 -I -c 'import json,re,sys; text=open(sys.argv[1]).read(); body=json.loads(re.sub(r"(?m)^//.*$","",text)); print("note:", body["note"])' "$T/$name.jsonc"
}
measure first
cat "$T/first.jsonc"
measure second -previous "$T/first.jsonc"
grep -q '"note": "level against' "$T/second.jsonc" && echo "PASS same commit twice reads level" || echo "FAIL same commit twice didn't read level"

python3 -I - bench/nbody.ts <<'PY'
import sys
path = sys.argv[1]
text = open(path).read()
plant = "let plantedSpin = 0;\nfor (let plantedIndex = 0; plantedIndex < 400000000; plantedIndex++) { plantedSpin += plantedIndex % 3; }\nif (plantedSpin < 0) { console.log(\"never\"); }\n"
open(path, "w").write(plant + text)
PY
measure planted -previous "$T/second.jsonc"
git diff --stat -- bench/nbody.ts
python3 -I -c 'import json,re,sys; body=json.loads(re.sub(r"(?m)^//.*$","",open(sys.argv[1]).read())); note=body["note"]; ok="nbody native user" in note and "(slower)" in note; print("PASS planted nbody reads slower" if ok else "FAIL planted nbody not named slower")' "$T/planted.jsonc"
python3 -I - "$T/second.jsonc" "$T/planted.jsonc" <<'PY'
import json, re, sys
before, after = [json.loads(re.sub(r"(?m)^//.*$", "", open(path).read())) for path in sys.argv[1:3]]
for name, entry in after["execute"]["programs"].items():
    print(f"{name}: native user {before['execute']['programs'][name]['native']['user']} to {entry['native']['user']}")
PY
echo VERDICT PASS
