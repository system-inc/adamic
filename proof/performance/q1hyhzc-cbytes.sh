# #q1hyhzc's C byte-identical check: bash proof/performance/q1hyhzc-cbytes.sh <shard> <shards>. The head is this
# commit's first ancestor not under proof/; the base is the head's main parent. Never merged.
set -u
shard=$1; shards=$2
. "$HOME/adamic-tools/env.sh" > /dev/null 2>&1 || { echo "VERDICT FAIL env"; exit 3; }
git -c protocol.file.allow=always submodule update --init --recursive --depth 1 > /tmp/q1c-sub.log 2>&1 || { echo "VERDICT FAIL submodules"; exit 3; }
head=$(git log --format=%H -1 --invert-grep --grep='^proof: ' HEAD)
base=7534c1d7575d
T=$(mktemp -d /tmp/q1c.XXXXXX)
echo "head $head base $(git rev-parse --short=12 $base) on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg), shard $shard of $shards"
git diff --quiet "$head" HEAD -- internal stage3 cmd || { echo "VERDICT FAIL the proof commits touch code"; exit 3; }
go build -o $T/head/adamic ./cmd/adamic > $T/build-head.log 2>&1 || { echo "VERDICT FAIL build head"; tail -5 $T/build-head.log; exit 1; }
git diff --name-status "$base" "$head" | while read -r status path; do
  case $status in A) rm -f "$path" ;; M) git show "$base:$path" > "$path" ;; *) echo "unexpected $status $path"; exit 3 ;; esac
done
go build -o $T/base/adamic ./cmd/adamic > $T/build-base.log 2>&1 || { echo "VERDICT FAIL build base"; tail -5 $T/build-base.log; exit 1; }
git checkout -q HEAD -- internal stage3
{ git ls-files internal/oracle/testdata | grep -E '\.(a|ts)$' | grep -v '\.d\.ts$'; git ls-files stage3/fixtures | grep -E '\.a$'; } | sort | awk -v s=$shard -v n=$shards 'NR % n == s' > $T/files
echo "files $(wc -l < $T/files)"
one() { local f=$1 key=${1//\//__}; for v in head base; do (cd "$(dirname "$f")" && timeout 120 $T/$v/adamic c "$(basename "$f")" > $T/$v.$key.out 2> $T/$v.$key.err; echo $? > $T/$v.$key.exit); done
  if cmp -s $T/head.$key.out $T/base.$key.out && cmp -s $T/head.$key.err $T/base.$key.err && cmp -s $T/head.$key.exit $T/base.$key.exit; then echo "same $(cat $T/head.$key.exit) $f"; else echo "DIFF $f"; fi; rm -f $T/*.$key.*; }
export -f one; export T
start=$(date +%s)
xargs -P "$(nproc)" -I{} bash -c 'one "$@"' _ {} < $T/files > $T/results
echo "compared in $(( $(date +%s) - start )) s"
echo "same $(grep -c '^same' $T/results), of which exit 0: $(grep -c '^same 0 ' $T/results), timed out: $(grep -c '^same 124 ' $T/results); differ $(grep -c '^DIFF' $T/results)"
grep '^DIFF' $T/results | head -10
[ "$(grep -c '^DIFF' $T/results)" = 0 ] && [ "$(wc -l < $T/results)" = "$(wc -l < $T/files)" ] && echo "VERDICT OK" || { echo "VERDICT FAIL"; exit 1; }
