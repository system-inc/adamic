# #az92fg0: how fast the built adamic binary compiles a .a file. Builds cmd/adamic once, then for hello world and two
# real files: one cold build (empty build and runtime caches), five warm builds (best wall kept), and five runs of
# `adamic c` (load, check, lower and emit C alone). Every number's instrument is the shell's clock around the process
# (date +%s%N) or the compiler's own $ADAMIC_BUILD_LOG lines (phase load, lower, emit; cBuild, cCompile, cLink; the
# runtime archive's hit or miss). Run after environment.sh and sandboxed-job.sh. Never merged.
set -u
W=$(pwd); T=$(mktemp -d "${TMPDIR:-/tmp}/clock.XXXXXX")
echo "on $(hostname), $(nproc) cores, $(grep -m1 'model name' /proc/cpuinfo | cut -d: -f2 | sed 's/^ //'), load $(cut -d' ' -f1-3 /proc/loadavg)"
echo "$(clang --version | head -1); $(go version)"
s=$(date +%s%N); go build -o "$T/adamic" ./cmd/adamic || { echo "VERDICT FAIL build adamic"; exit 1; }; echo "adamic built in $(( ($(date +%s%N) - s) / 1000000 )) ms (not part of any number below)"
mkdir -p "$T/hello"; printf 'console.log("hello, world");\n' > "$T/hello/hello.a"
ms() { echo $(( ($2 - $1) / 1000000 )); }
summary() { # log file -> one line of the compiler's own phase and clang seconds
  awk '{k=$1; n=$2; s=$NF} k=="phase"{p[n]+=s} k=="cBuild"||k=="cCompile"||k=="cLink"{c[k]+=s} k!="phase"&&k!="cBuild"&&k!="cCompile"&&k!="cLink"{o[k" "$4]++; os[k]+=s}
    END{printf "load %.3f s, lower %.3f s, emit %.3f s; ", p["load"], p["lower"], p["emit"];
        printf "cBuild %.3f s, cCompile %.3f s, cLink %.3f s; ", c["cBuild"], c["cCompile"], c["cLink"];
        for (x in o) printf "%s x%d; ", x, o[x]; for (x in os) printf "%s %.3f s; ", x, os[x]}' "$1"; }
for f in "$T/hello/hello.a" bench/holdout/revcomp.a bench/holdout/fannkuchredux.a stage3/namespaces/shapes/Parser.a; do
  case $f in /*) path=$f;; *) path=$W/$f;; esac
  dir=$(dirname "$path"); base=$(basename "$path")
  echo "== $f, $(wc -l < "$path") lines, $(wc -c < "$path") bytes"
  export XDG_CACHE_HOME="$T/cache-$base" ADAMIC_BUILD_CACHE_DIR="$T/buildcache-$base"; mkdir -p "$XDG_CACHE_HOME" "$ADAMIC_BUILD_CACHE_DIR"
  log="$T/cold-$base.log"; s=$(date +%s%N)
  (cd "$dir" && ADAMIC_BUILD_LOG=$log "$T/adamic" build "$base" -o "$T/out-$base") > "$T/cold-$base.out" 2>&1; code=$?; e=$(date +%s%N)
  if [ $code != 0 ]; then echo "  build exit $code: $(head -3 "$T/cold-$base.out" | tr '\n' ' ' | cut -c1-200)"; continue; fi
  echo "  cold: $(ms $s $e) ms wall; $(summary $log)"
  best=999999999; bestlog=""
  for i in 1 2 3 4 5; do log="$T/warm-$base-$i.log"; s=$(date +%s%N)
    (cd "$dir" && ADAMIC_BUILD_LOG=$log "$T/adamic" build "$base" -o "$T/out-$base") > /dev/null 2>&1; e=$(date +%s%N)
    t=$(ms $s $e); walls="${walls:-} $t"; [ $t -lt $best ] && { best=$t; bestlog=$log; }; done
  echo "  warm build: best $best ms of 5 (${walls# }); best run's log: $(summary $bestlog)"; walls=""
  cbest=999999999; for i in 1 2 3 4 5; do s=$(date +%s%N); (cd "$dir" && "$T/adamic" c "$base") > "$T/c-$base.c" 2>/dev/null; e=$(date +%s%N); t=$(ms $s $e); cw="${cw:-} $t"; [ $t -lt $cbest ] && cbest=$t; done
  echo "  adamic c (front end alone, to C): best $cbest ms of 5 (${cw# }); C is $(wc -l < "$T/c-$base.c") lines, $(wc -c < "$T/c-$base.c") bytes"; cw=""
  echo "  output runs: $("$T/out-$base" < /dev/null > /dev/null 2>&1; echo exit $?)"
done
echo "load at end $(cut -d' ' -f1-3 /proc/loadavg)"
echo "VERDICT OK"
