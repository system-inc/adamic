#!/usr/bin/env bash
# #0q0cv0m scout, never to land: the null-receiver ?. read probes at main, on Node and both backends.
set -u
P="$TMPDIR/probes0q0"; mkdir -p "$P" && echo "PASS scratch" || { echo "FAIL scratch"; exit 1; }
for f in scratch-0q0/probes/*.txt; do cp "$f" "$P/$(basename "${f%.txt}")"; done
ls "$P"
bash cloud/probe/run.sh "$P" --logs "$P/logs" 2>&1 | tail -60
for f in "$P"/logs/*; do echo "== $f"; head -c 1500 "$f"; echo; done 2>/dev/null | head -120
echo DONE
