#!/bin/bash
# Runs the quiet sequence, waiting for the 1-minute load to fall under 1 before each run.
S=/tmp/claude-0/-home-user-adamic/dfa01dfb-d6d2-5c72-bbc4-beb3de8cdee4/scratchpad
settle() { while awk '{exit !($1 >= 1.0)}' /proc/loadavg; do sleep 5; done; }
for run in "$@"; do
  name=${run%%:*} unit=${run##*:}
  settle
  $S/tools/run.sh "$name" "^TestSplitTSGoAgreesUnit$unit\$"
done
