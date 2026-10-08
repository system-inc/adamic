#!/usr/bin/env bash
set -euo pipefail
sha=$1 seed=$2 count=$3 box=$4 ready=$5
touch "$ready"
source ~/adamic-tools/env.sh
# Separate caches keep idle compilation from contending on gate cache locks.
export XDG_CACHE_HOME="$HOME/idle/cache"
export GOCACHE="$HOME/idle/cache/go" TMPDIR="$HOME/idle/tmp"
mkdir -p "$TMPDIR" "$GOCACHE"
base="$HOME/idle/out/${sha:0:12}/$seed"
mkdir -p "$base"
out=$(mktemp -d "$base/attempt.XXXXXXXX")
printf '%s\n' "$sha" > "$out/main"
printf '%s\n' "$seed" > "$out/seed"
exec >"$out/stdout.log" 2>&1
[ -d ~/idle/tree/.git ] || git clone -q https://github.com/system-inc/adamic.git ~/idle/tree
git -C ~/idle/tree fetch -q origin main "$sha"
git -C ~/idle/tree checkout -q --detach "$sha"
git -C ~/idle/tree submodule update -q --init --recursive
cd ~/idle/tree
go build -o "$out/adamic-fuzz" ./cmd/adamic-fuzz
go build -o "$out/adamic-reduce" ./cmd/adamic-reduce
code=0
"$out/adamic-fuzz" -seed "$seed" -count "$count" -parallel 1 -v -work "$out/work" -findings "$out/raw" > "$out/fuzz.log" 2>&1 || code=$?
# Exit 1 also means findings; the summary proves the entire range ran.
grep -q "^$count programs from seed $seed " "$out/fuzz.log" || exit 1
[ "$code" -le 1 ] || exit "$code"
for program in "$out"/raw/seed*.a; do
  [ -f "$program" ] || continue
  n=${program##*/seed}; n=${n%.a}
  temp=$(mktemp -d "$out/finding.XXXXXXXX")
  # Keep the generated original as well as the fuzzer's shrunk input.
  "$out/adamic-fuzz" -seed "$n" -print > "$temp/program.a"
  "$out/adamic-reduce" -parallel 1 -work "$temp/work" -o "$temp/reduced.a" "$program" 2> "$temp/reduce.log" || exit 1
  sed -n 's/^adamic-reduce: signature //p' "$temp/reduce.log" > "$temp/signature.txt"
  [ -s "$temp/signature.txt" ] || exit 1
  printf '%s\n' "$n" > "$temp/seed"
  printf '%s\n' "$sha" > "$temp/main"
  printf '%s\n' "$box" > "$temp/box"
  hash=$(cat "$temp/signature.txt" "$temp/reduced.a" | sha256sum | cut -d' ' -f1)
  mkdir -p ~/idle/findings
  # A directory appears to collectors only once all its files are complete.
  rm -rf "$temp/work"
  if [ ! -d "$HOME/idle/findings/$hash" ]; then mv "$temp" "$HOME/idle/findings/$hash"; fi
done
# Atomic completion, written last. SIGKILL leaves no completion marker.
printf '%s\n' "$count" > "$base/done.tmp"
mv "$base/done.tmp" "$base/done"
