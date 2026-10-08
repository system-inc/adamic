#!/usr/bin/env bash
# One pinned one-file change per area, build only, on the gate box: prints CSV rows
# (timestamp_utc,main_sha,box,file,build_seconds). Run by cloud/build-one-file.sh on every new main.
#
# The edit is the same each time, a new function at the end of the file, but it carries a fresh
# number so the build cache can never answer it from an earlier run. The Go files are timed with
# go build ./...; the stage 1 file with cmd/adamic building its program (the per-unit object cache on,
# as a worker's local build has it). Everything is built once untouched first, so the timing is one
# changed file over a warm cache.
set -euo pipefail
sha=$1
mkdir -p ~/bench
[ -d ~/bench/tree ] || cp -a ~/fast-gate/tree ~/bench/tree
source ~/adamic-tools/env.sh
cd ~/bench/tree
git fetch -q origin "${sha}"
git switch -q --detach "${sha}"
git submodule update -q --init --recursive
go build ./...
go build -o ~/bench/adamic ./cmd/adamic
~/bench/adamic build stage1/cohere/json/main.ts -o ~/bench/json.out > /dev/null
nonce=$(date +%s%N)
measure() {
  local file=$1 language=$2 started ended
  cp "${file}" ~/bench/original
  if [ "${language}" = go ]; then
    printf '\nfunc buildOneFileBenchmark%s() int { return %s }\n' "${nonce}" "${nonce:0:9}" >> "${file}"
  else
    printf '\nexport function buildOneFileBenchmark%s(): number {\n    return %s;\n}\n' "${nonce}" "${nonce:0:9}" >> "${file}"
  fi
  started=$(date +%s.%N)
  if [ "${language}" = go ]; then go build ./... ; else go build -o ~/bench/adamic ./cmd/adamic && ~/bench/adamic build stage1/cohere/json/main.ts -o ~/bench/json.out > /dev/null; fi
  ended=$(date +%s.%N)
  cp ~/bench/original "${file}"
  printf '%s,%s,%s,%s,%.2f\n' "$(date -u +%FT%TZ)" "${sha}" "$(hostname) $(nproc)" "${file}" "$(echo "${ended} - ${started}" | bc)"
}
measure internal/lower/optional_widening.go go
measure internal/native/native.go go
measure stage1/cohere/json/parser.ts adamic
git status --short | grep -q . && { echo "the bench tree is not clean after the run" >&2; exit 1; }
