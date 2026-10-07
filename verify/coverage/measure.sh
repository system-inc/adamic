#!/usr/bin/env bash
# measure.sh measures what the randomized test-program generator (internal/fuzz, cmd/adamic-fuzz)
# makes the compiler and the C runtime execute, and what the hand-written oracle fixtures do, and
# writes verify/coverage/REPORT.md and REPORT.json from the two. Meant to run nightly.
#
#	verify/coverage/measure.sh [output directory]
#
# Environment: SEED (1), COUNT (2000), PARALLEL (6), FIXTURES (1; 0 skips the oracle side), and
# ANALYZE_ONLY=1 to skip both runs and only read an output directory a finished run left.
#
# How each side is instrumented, without changing what it measures:
#   - Go: adamic-fuzz drives the compiler as a subprocess (adamic c, adamic js, fuzz.Prepare builds it
#     with go build), so GOFLAGS gives that build -cover and -coverpkg, and GOCOVERDIR collects every
#     run's counters. cmd/adamic is in -coverpkg only because this toolchain links the coverage
#     writer into a binary only when its main package is covered; the report reads the other four.
#     The oracle fixtures compile in process, so go test gets the same -cover -coverpkg.
#   - C: ADAMIC_C_COVERAGE=1 turns on native.Options.Coverage in every build (clang's
#     -fprofile-instr-generate -fcoverage-mapping, the sanitizers left on, the runtime cached under
#     its own key). adamic-fuzz gives each run its own LLVM_PROFILE_FILE under
#     ADAMIC_C_COVERAGE_DIRECTORY; the oracle's runs inherit one pattern with the process id and the
#     binary's signature in it (%p, %m), so no two runs overwrite each other's profile.
#
# The fuzzer runs with -shrink=false: shrinking runs extra candidate programs, as many as findings
# call for, which a fixed budget of seeds can't hold. Every other flag is a normal run's, with every
# family that's on by default.
#
# On macOS, ASan has no LeakSanitizer, and run.go's leak recheck (ASAN_OPTIONS=detect_leaks=1) then
# fails every program that finishes. Until run.go knows that, patch it locally (not committed) to
# ignore "detect_leaks is not supported on this platform", or every finishing program is a finding.
set -euo pipefail

repository="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
output="${1:-${TMPDIR:-/tmp}/adamic-coverage}"
seed="${SEED:-1}"
count="${COUNT:-2000}"
parallel="${PARALLEL:-6}"
fixtures="${FIXTURES:-1}"
coverpkg="./internal/lower/...,./internal/native/...,./internal/ir/...,./internal/flow/..."

if [ "$(uname)" = Darwin ]; then
	llvm_profdata=(xcrun llvm-profdata)
	llvm_cov=(xcrun llvm-cov)
	if ! grep -q "detect_leaks is not supported on this platform" "$repository/internal/fuzz/run.go"; then
		echo "measure.sh: warning: internal/fuzz/run.go doesn't ignore macOS's missing LeakSanitizer; every finishing program will be a leak finding" >&2
	fi
else
	llvm_profdata=(llvm-profdata)
	llvm_cov=(llvm-cov)
fi

analyze_only="${ANALYZE_ONLY:-0}"
mkdir -p "$output"
if [ "$analyze_only" != 1 ]; then
	for side in generator fixtures; do
		if [ -e "$output/$side" ]; then
			echo "measure.sh: $output/$side exists; give a fresh output directory" >&2
			exit 1
		fi
	done
	mkdir -p "$output/generator/go" "$output/generator/c" "$output/generator/work" "$output/fixtures/c"
fi
commands="$output/commands.md"
cd "$repository"
commit="$(git rev-parse HEAD)"

# The runtime the coverage builds link, sanitized as the fuzzer and the oracle compile it. Its objects
# carry the coverage mapping llvm-cov reads.
library="$(ADAMIC_C_COVERAGE=1 go run -trimpath ./verify/coverage/runtimelibrary)"
runtime_objects=("$(dirname "$library")"/*.o)

if [ "$analyze_only" != 1 ]; then
go build -trimpath -o "$output/adamic-fuzz" ./cmd/adamic-fuzz

started=$(date +%s)
set +e
ADAMIC_C_COVERAGE=1 ADAMIC_C_COVERAGE_DIRECTORY="$output/generator/c" GOCOVERDIR="$output/generator/go" \
	GOFLAGS="-trimpath -cover -coverpkg=./cmd/adamic,$coverpkg" \
	"$output/adamic-fuzz" -root "$repository" -seed "$seed" -count "$count" -parallel "$parallel" -shrink=false -v \
	-work "$output/generator/work" > "$output/generator/run.log" 2>&1
generator_exit=$?
set -e
echo "$generator_exit $(($(date +%s) - started))" > "$output/generator/status"

if [ "$fixtures" = 1 ]; then
	started=$(date +%s)
	set +e
	ADAMIC_GATE_UNCACHED=1 ADAMIC_C_COVERAGE=1 LLVM_PROFILE_FILE="$output/fixtures/c/oracle-%p-%m.profraw" \
		go test -trimpath -count=1 -v -cover -coverpkg="$coverpkg" -coverprofile="$output/fixtures/go.txt" \
		-run '^TestNativeAgreesWithNode$' -parallel "$parallel" -timeout 90m ./internal/oracle \
		> "$output/fixtures/test.log" 2>&1
	fixtures_exit=$?
	set -e
	echo "$fixtures_exit $(($(date +%s) - started))" > "$output/fixtures/status"
fi
fi
read -r generator_exit generator_seconds < "$output/generator/status"
if [ -f "$output/fixtures/status" ]; then
	read -r fixtures_exit fixtures_seconds < "$output/fixtures/status"
fi

go tool covdata textfmt -i="$output/generator/go" -o="$output/generator/go.txt"

for side in generator fixtures; do
	if [ ! -f "$output/$side/go.txt" ]; then
		continue
	fi
	find "$output/$side/c" -name '*.profraw' > "$output/$side/profraw.list"
	"${llvm_profdata[@]}" merge -sparse -f "$output/$side/profraw.list" -o "$output/$side/c.profdata" > "$output/$side/merge.log" 2>&1
	objects=()
	for object in "${runtime_objects[@]:1}"; do
		objects+=(-object "$object")
	done
	"${llvm_cov[@]}" export -format=text "${runtime_objects[0]}" "${objects[@]}" -instr-profile "$output/$side/c.profdata" > "$output/$side/c.json" 2> "$output/$side/export.log"
	"${llvm_cov[@]}" export -format=lcov "${runtime_objects[0]}" "${objects[@]}" -instr-profile "$output/$side/c.profdata" > "$output/$side/c.lcov" 2>> "$output/$side/export.log"
done

# What each side was, for the report's header.
python3 - "$output" "$seed" "$count" "$parallel" "$generator_exit" "$generator_seconds" "${fixtures_exit:-}" "${fixtures_seconds:-}" "$commit" <<'EOF'
import json, os, re, sys
output, seed, count, parallel, generator_exit, generator_seconds, fixtures_exit, fixtures_seconds, commit = sys.argv[1:10]
log = open(os.path.join(output, "generator", "run.log")).read()
verdicts = {}
for verdict in ("crash", "agreed", "finding", "checked", "not yet", "invalid", "unfit"):
    match = re.search(r"^  " + verdict + r" +(\d+)$", log, re.M)
    if match:
        verdicts[verdict] = int(match.group(1))
profiles = sum(1 for _ in open(os.path.join(output, "generator", "profraw.list")))
meta = {"seeds": f"{seed} to {int(seed) + int(count) - 1}", "parallel": int(parallel), "commit": commit[:10],
        "verdicts": ", ".join(f"{key} {value}" for key, value in verdicts.items()),
        "C profiles": profiles, "exit": int(generator_exit), "seconds": int(generator_seconds)}
json.dump(meta, open(os.path.join(output, "generator", "meta.json"), "w"))
if fixtures_exit:
    test = open(os.path.join(output, "fixtures", "test.log")).read()
    profiles = sum(1 for _ in open(os.path.join(output, "fixtures", "profraw.list")))
    failed = re.findall(r"--- FAIL: TestNativeAgreesWithNode/(\S+)", test)
    meta = {"test": "go test ./internal/oracle -run TestNativeAgreesWithNode", "subtests passed": len(re.findall(r"--- PASS: TestNativeAgreesWithNode/", test)),
            "subtests failed": len(failed) and f"{len(failed)} ({', '.join(failed)})", "C profiles": profiles,
            "exit": int(fixtures_exit), "seconds": int(fixtures_seconds)}
    json.dump(meta, open(os.path.join(output, "fixtures", "meta.json"), "w"))
EOF

cat > "$commands" <<EOF
Run by \`verify/coverage/measure.sh\` at $commit. Reproduce with \`verify/coverage/measure.sh <fresh directory>\`
(SEED, COUNT, PARALLEL and FIXTURES in the environment). What it ran:

\`\`\`
# the coverage runtime: sanitized, -fprofile-instr-generate -fcoverage-mapping, its own cache key
ADAMIC_C_COVERAGE=1 go run -trimpath ./verify/coverage/runtimelibrary
go build -trimpath -o adamic-fuzz ./cmd/adamic-fuzz

# the generator: seeds $seed to $((seed + count - 1)), every default family
ADAMIC_C_COVERAGE=1 ADAMIC_C_COVERAGE_DIRECTORY=generator/c GOCOVERDIR=generator/go \\
  GOFLAGS="-trimpath -cover -coverpkg=./cmd/adamic,$coverpkg" \\
  adamic-fuzz -root . -seed $seed -count $count -parallel $parallel -shrink=false -v -work generator/work

# the oracle fixtures, uncached so every binary really runs
ADAMIC_GATE_UNCACHED=1 ADAMIC_C_COVERAGE=1 LLVM_PROFILE_FILE=fixtures/c/oracle-%p-%m.profraw \\
  go test -trimpath -count=1 -v -cover -coverpkg=$coverpkg -coverprofile=fixtures/go.txt \\
  -run '^TestNativeAgreesWithNode\$' -parallel $parallel -timeout 90m ./internal/oracle

go tool covdata textfmt -i=generator/go -o=generator/go.txt
llvm-profdata merge -sparse -f <side>/profraw.list -o <side>/c.profdata
llvm-cov export -format=text <runtime objects> -instr-profile <side>/c.profdata > <side>/c.json
llvm-cov export -format=lcov <runtime objects> -instr-profile <side>/c.profdata > <side>/c.lcov
python3 verify/coverage/analyze.py ...
\`\`\`

Caveats, as measured:

- Go has block coverage, not branch coverage; the Go "branch" shares are block shares.
- adamic-fuzz runs the compiler twice per program (\`adamic c\` and \`adamic js\`), so the Go side
  counts the JavaScript backend's lowering too; lowering is the same for both.
- The C runtime is read through the sanitized runtime's objects. The oracle also runs a release
  build (-O2, the size-class allocator on) and a counted build of each fixture. Their runs are
  credited to the runtime's external functions whose structure matches the sanitized build's;
  their static functions are named after the directory they were compiled in, so they aren't
  credited, and llvm-cov skips any function whose profile doesn't match (export.log counts them).
- The fuzzer runs with -shrink=false, so exactly the seeds' programs are measured.
- cmd/adamic is in the fuzzer's -coverpkg only because this toolchain links the coverage writer
  into a binary only when its main package is covered; the report doesn't read it.
- adamic.h's five static inline functions are counted only for calls from inside the runtime; their
  copies inlined into each program's main.c aren't read.
- A run that ends in a signal (an ASan abort, a deadline kill) writes no C profile.
EOF

python3 "$repository/verify/coverage/analyze.py" "$output" "$repository" "$repository/verify/coverage/REPORT.md" "$repository/verify/coverage/REPORT.json" "$commands"
echo "measure.sh: wrote verify/coverage/REPORT.md and REPORT.json from $output (generator exit $generator_exit, fixtures exit ${fixtures_exit:-skipped})"
