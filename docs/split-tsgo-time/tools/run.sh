#!/bin/bash
# run.sh <name> <test-regex>: runs one test with -v and a fresh build log, under /usr/bin/time -v, with the poller.
set -u
source /opt/adamic-tools/env.sh
E=/home/user/adamic/docs/split-tsgo-time/logs
name=$1 regex=$2
mkdir -p "$E"
cd /home/user/adamic
{
  echo "date $(date -u +%FT%TZ)"; echo "sha $(git rev-parse HEAD)"; echo "nproc $(nproc)"; grep -m1 'model name' /proc/cpuinfo
  grep MemTotal /proc/meminfo; clang --version | head -1; node --version; echo "WASI_SYSROOT=${WASI_SYSROOT:-}"
  echo "loadavg-before $(cat /proc/loadavg)"
  echo "units-cache $(ls ~/.cache/adamic-build/native-units 2>/dev/null | wc -l) entries"
} > "$E/$name.box"
rm -f "$E/$name.build" "$E/$name.timeline.stop"
: > "$E/$name.v"
python3 /tmp/claude-0/-home-user-adamic/dfa01dfb-d6d2-5c72-bbc4-beb3de8cdee4/scratchpad/tools/poll.py "$E/$name.v" "$E/$name.build" "$E/$name.timeline" "$TMPDIR" &
poller=$!
ADAMIC_BUILD_LOG="$E/$name.build" /usr/bin/time -v -o "$E/$name.time" go test ./internal/native -run "$regex" -count=1 -v -timeout 30m > "$E/$name.v" 2>&1
echo "exit=$?" >> "$E/$name.box"
echo "loadavg-after $(cat /proc/loadavg)" >> "$E/$name.box"
sleep 0.2; touch "$E/$name.timeline.stop"; wait $poller; rm -f "$E/$name.timeline.stop"
