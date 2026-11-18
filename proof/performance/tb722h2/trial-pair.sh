# #tb722h2's trial pair: oracle's runtime archives built into a fresh cache, then one non-test edit to
# internal/native (a function nothing calls, appended to native.go), then built again into the same cache. On H2 every
# archive's key must stay and hit; on H2's base, main, the same edit must move them, which shows the pair can fail.
# Run after environment.sh and sandboxed-job.sh. Never merged.
set -u
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg), at $(git rev-parse --short=12 HEAD~1)"
mkdir -p "$TMPDIR"
export ADAMIC_BUILD_CACHE_DIR="$(mktemp -d "$TMPDIR/trial-cache.XXXXXX")"
products='^TestProduct_(NativeRuntimes|ProgramRuntimes)(Release|Sanitized)[0-2]$'
side() {
	local start=$(date +%s)
	ADAMIC_BUILD_LOG="$TMPDIR/log-$1" go test -count=1 -timeout 40m -run "$products" ./internal/oracle/ > "$TMPDIR/test-$1.log" 2>&1
	local code=$?
	grep -E '^build [^ ]*runtime[^ ]* ' "$TMPDIR/log-$1" | awk '{print $2, $3, $4}' | sort > "$TMPDIR/keys-$1"
	echo "SIDE $1 exit $code, $(( $(date +%s) - start )) s, $(wc -l < "$TMPDIR/keys-$1") runtime builds: $(awk '{print $3}' "$TMPDIR/keys-$1" | sort | uniq -c | tr -s ' \n' ' ')"
	[ $code = 0 ] || grep -E '^(--- FAIL|FAIL|panic)' "$TMPDIR/test-$1.log" | head -5
}
side before
printf '\n// trialPairEdit is a non-test edit to internal/native for #tb722h2'"'"'s trial pair, never merged.\nfunc trialPairEdit() int { return 1 }\n' >> internal/native/native.go
git diff --stat -- internal/native/native.go
side after
python3 - "$TMPDIR/keys-before" "$TMPDIR/keys-after" <<'PYTHON'
import sys
def load(p): return [tuple(l.split()[:2]) for l in open(p) if l.strip()]
before, after = load(sys.argv[1]), load(sys.argv[2])
b, a = set(before), set(after)
names = sorted({n for n, _ in b | a})
print(f"keys before {len(b)}, after {len(a)}, kept {len(b & a)}, moved {len(b - a)}, new {len(a - b)}")
for n in names: print(f"  {n}: before {sum(1 for x in b if x[0] == n)}, after {sum(1 for x in a if x[0] == n)}, kept {sum(1 for x in b & a if x[0] == n)}")
print("VERDICT", "KEPT" if b and b == a else ("MOVED" if b and not (b & a) else "MIXED"))
PYTHON
