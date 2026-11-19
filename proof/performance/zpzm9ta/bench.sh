# #zpzm9ta: price Emit's literal-run check (this commit's parent is the head, its grandparent main). Builds adamic at
# the head and, with the head's non-test internal/native change reversed, at main; compiles each bench variant with
# both; runs each binary 5 times on one box and prints the median net ns per read. Never merged.
set -u
O=${TMPDIR:-/tmp}/lr; mkdir -p "$O"; K=proof/performance/zpzm9ta
head=$(git rev-parse HEAD~1); base=$(git rev-parse HEAD~2)
echo "on $(hostname), $(nproc) cores, $(grep -m1 'model name' /proc/cpuinfo | cut -d: -f2), load $(cut -d' ' -f1-3 /proc/loadavg), clang $(clang --version | head -1)"
echo "head $head base $base"
go build -o "$O/adamic-head" ./cmd/adamic || { echo "VERDICT FAIL build head"; exit 1; }
git diff "$base" "$head" -- internal/native ':!*_test.go' ':!internal/native/testdata' > "$O/head.diff"
echo "reversing $(grep -c '^diff --git' "$O/head.diff") files: $(grep '^diff --git' "$O/head.diff" | awk '{print $3}' | tr '\n' ' ')"
git apply -R "$O/head.diff" || { echo "VERDICT FAIL reverse"; exit 1; }
go build -o "$O/adamic-base" ./cmd/adamic || { echo "VERDICT FAIL build base"; exit 1; }
git apply "$O/head.diff"
for N in 1 16 400; do for mode in view plain; do for read in 0 1; do
  python3 $K/gen.py $N $mode $read > "$O/b_${N}_${mode}_${read}.ts"
  for side in base head; do
    "$O/adamic-$side" build "$O/b_${N}_${mode}_${read}.ts" -o "$O/bin_${side}_${N}_${mode}_${read}" > "$O/build_${side}_${N}_${mode}_${read}.log" 2>&1 || { echo "FAIL build $side $N $mode $read"; tail -3 "$O/build_${side}_${N}_${mode}_${read}.log"; }
  done
done; done; done
median() { sort -n | sed -n 3p; }
run() { local b=$1 ts=(); for i in 1 2 3 4 5; do s=$(date +%s.%N); "$b" > /dev/null; e=$(date +%s.%N); ts+=("$(awk -v a="$s" -v b="$e" 'BEGIN { printf "%.6f", b - a }')"); done; printf '%s\n' "${ts[@]}" | median; }
printf '%-6s %-5s %-6s %12s %12s %10s\n' N mode side read_s noread_s ns_per_read
for N in 1 16 400; do for mode in view plain; do for side in base head; do
  r=$(run "$O/bin_${side}_${N}_${mode}_1"); z=$(run "$O/bin_${side}_${N}_${mode}_0")
  ns=$(awk -v r="$r" -v z="$z" 'BEGIN { printf "%.3f", (r - z) * 1e9 / (1024 * 40000) }')
  printf '%-6s %-5s %-6s %12.4f %12.4f %10.2f\n' $N $mode $side "$r" "$z" "$ns"
done; done; done
for side in base head; do "$O/adamic-$side" c "$O/b_400_view_1.ts" > "$O/c_$side.c" 2>/dev/null; echo "$side C: $(grep -c 'adamic_same_number' "$O/c_$side.c") same_number, $(grep -c 'trunc' "$O/c_$side.c") trunc"; done
echo "VERDICT OK"
