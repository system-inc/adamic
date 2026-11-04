#!/usr/bin/env bash
# 1 Hz sample: rss of census binaries, plus total used memory.
out=$1
while true; do
  ts=$(date +%s)
  ps -eo rss=,comm= | awk -v ts=$ts '$2 ~ /census/ {print ts","$2","$1}' >> $out
  sleep 1
done
