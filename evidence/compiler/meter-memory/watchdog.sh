#!/usr/bin/env bash
# kill the step-4 latent-census if combined census RSS exceeds 12 GiB
while pgrep -x latent-census >/dev/null; do
  total=$(ps -eo rss=,comm= | awk '/census/ {s+=$1} END {print s+0}')
  if [ "$total" -gt 12582912 ]; then echo "watchdog: killed latent-census at total ${total} KB $(date +%s)" >> /tmp/m/run3/watchdog.log; pkill -TERM -x latent-census; fi
  sleep 2
done
