#!/usr/bin/env bash
set -euo pipefail
S=/tmp/claude-0/-home-user-adamic/8a2602c2-ac96-5be7-8daf-c96e4d3f586c/scratchpad
repository=/home/user/adamic; cd $repository
ev=$S/ev; scratch=$S/work; label=main
commit=$(git rev-parse HEAD)
git show "$commit:stage3/source.json" > $S/source.json
url=$(python3 -c 'import json;print(json.load(open("'$S'/source.json"))["repository"])')
tag=$(python3 -c 'import json;print(json.load(open("'$S'/source.json"))["tag"])')
git clone --depth 1 --branch "$tag" "$url" $scratch/upstream > $ev/upstream.log 2>&1
export ADAMIC_TYPESCRIPT_SOURCE=$scratch/upstream
mkdir -p $scratch
(cd $repository && go build -o $scratch/census ./stage3/census/tool) > $ev/build.log 2>&1
python3 stage3/census/latent/make_overlay.py $repository $scratch/latent-overlay > $ev/latent-overlay.log 2>&1
gofmt -w $scratch/latent-overlay/*.go
python3 stage3/meter/entry_overlay.py $scratch/latent-overlay $scratch/entry-overlay > $ev/entry-overlay.log 2>&1
gofmt -w $scratch/entry-overlay/*.go
(cd $repository && go build -buildvcs=false -overlay=$scratch/entry-overlay/overlay.json -o $scratch/entry-census ./stage3/census/latent/tool) > $ev/entry-build.log 2>&1
mkdir -p $scratch/$label-source
git archive "$commit" stage3 | tar -x -C $scratch/$label-source
(cd $scratch/$label-source/stage3/api && npm ci --ignore-scripts --no-audit --no-fund) > $ev/packages.log 2>&1
bash $scratch/$label-source/stage3/apply.sh $scratch/$label-adapted > $ev/apply.log 2>&1
entry=$scratch/$label-adapted/src/tsc/tsc.ts
[[ -f $entry ]] || { echo missing entry; exit 1; }
mkdir -p $ev/tsc
sync; echo "MEASURE_START $(date -u +%FT%TZ)" > $ev/phase.txt
LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 /usr/bin/time -v -o $ev/entry-census.time $scratch/entry-census "$entry" "$ev/tsc/latent.jsonl" > $ev/tsc/latent.log 2>&1 &
pid=$!
echo "epoch,rss_kb,vsz_kb" > $ev/rss.csv
while kill -0 $pid 2>/dev/null; do
  for p in $(pgrep -f "$scratch/entry-census" || true); do
    ps -o rss=,vsz= -p $p 2>/dev/null | awk -v t=$(date +%s) -v p=$p '{print t","$1","$2","p}' >> $ev/rss.csv || true
  done
  sleep 1
done
wait $pid && echo "exit=0" > $ev/exit.txt || echo "exit=$?" > $ev/exit.txt
echo "MEASURE_END $(date -u +%FT%TZ)" >> $ev/phase.txt
echo DONE >> $ev/phase.txt
