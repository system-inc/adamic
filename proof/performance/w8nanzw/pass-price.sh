# #w8nanzw (A): what its post-lowering pass costs. SIDE=head times checkSlotWrites; SIDE=base times the two passes it
# replaces (checkKindUnionWriteReach, checkViewOptionalWrites). A timer is patched around each call in lower.go, adamic
# is built, and `adamic c` runs over the oracle's and stage3's programs, each recording its pass time and its own wall.
# Run after environment.sh and sandboxed-job.sh. Never merged.
set -u
SIDE=${SIDE:?head or base}
T=$(mktemp -d "${TMPDIR:-/tmp}/w8.XXXXXX")
echo "side $SIDE on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
cat > internal/lower/zz_pass_time.go <<'GO'
package lower

import (
	"fmt"
	"os"
	"time"
)

func passNow() time.Time { return time.Now() }

func passTime(name string, start time.Time) {
	if path := os.Getenv("ADAMIC_PASS_TIMES"); path != "" {
		if file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
			fmt.Fprintf(file, "%s\t%d\n", name, time.Since(start).Nanoseconds())
			file.Close()
		}
	}
}
GO
python3 - "$SIDE" <<'PYTHON' || { echo "VERDICT FAIL patch"; exit 1; }
import sys
side = sys.argv[1]; path = 'internal/lower/lower.go'; s = open(path).read()
names = ['checkSlotWrites'] if side == 'head' else ['checkKindUnionWriteReach', 'checkViewOptionalWrites']
for name in names:
    old = f'\tif err := lowering.{name}(); err != nil {{'
    assert s.count(old) == 1, old
    new = f'\t{name}Start := passNow()\n\t{name}Err := lowering.{name}()\n\tpassTime("{name}", {name}Start)\n\tif err := {name}Err; err != nil {{'
    s = s.replace(old, new)
open(path, 'w').write(s)
print('timed', ' '.join(names))
PYTHON
go build -o $T/adamic ./cmd/adamic > $T/build.log 2>&1 || { echo "VERDICT FAIL build"; tail -5 $T/build.log; exit 1; }
{ git ls-files internal/oracle/testdata 2>/dev/null || find internal/oracle/testdata -type f; } | grep -E '\.(a|ts)$' | grep -v '\.d\.ts$' > $T/files
{ git ls-files stage3/fixtures 2>/dev/null || find stage3/fixtures -type f; } | grep -E '\.a$' >> $T/files
sort -u -o $T/files $T/files
echo "files $(wc -l < $T/files)"
one() { local f=$1 key=${1//\//__}; local s=$(date +%s%N); (cd "$(dirname "$f")" && ADAMIC_PASS_TIMES=$T/times.$key timeout 120 $T/adamic c "$(basename "$f")" > /dev/null 2>&1); local code=$?; echo "$f	$code	$(( $(date +%s%N) - s ))" >> $T/walls.$key; }
export -f one; export T
start=$(date +%s)
xargs -P "$(( $(nproc) / 2 ))" -I{} bash -c 'one "$@"' _ {} < $T/files
echo "ran in $(( $(date +%s) - start )) s"
python3 - "$T" <<'PYTHON'
import glob, os, sys, collections
T = sys.argv[1]
passes, walls, per = collections.Counter(), 0, []
calls = collections.Counter()
for path in glob.glob(os.path.join(T, 'walls.*')):
    key = path.split('walls.', 1)[1]
    f, code, wall = open(path).read().strip().split('\t')
    walls += int(wall)
    mine = 0
    tpath = os.path.join(T, 'times.' + key)
    if os.path.exists(tpath):
        for line in open(tpath):
            name, ns = line.split('\t'); passes[name] += int(ns); calls[name] += 1; mine += int(ns)
    per.append((mine, int(wall), f, code))
per.sort(reverse=True)
print(f"programs {len(per)}, lowered through the pass {max(calls.values()) if calls else 0}, total wall of adamic c {walls/1e9:.1f} s")
for name in passes: print(f"PASS {name}: {calls[name]} calls, {passes[name]/1e6:.1f} ms in all, {passes[name]/max(calls[name],1)/1e3:.1f} us a call")
print(f"share of adamic c's wall: {sum(passes.values())/walls*100:.3f}%")
print("slowest five by pass time:")
for mine, wall, f, code in per[:5]: print(f"  {mine/1e6:8.2f} ms of {wall/1e6:8.1f} ms  exit {code}  {f}")
PYTHON
echo "VERDICT OK"
