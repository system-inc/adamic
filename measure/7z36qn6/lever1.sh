#!/bin/bash
# #7z36qn6 lever 1 before Loom, sealed: the reset's own tests and mutants, the statics scanner and layout test,
# the oracle's TestSignal rows unchanged, one oracle mutant (a masked signal skipped) shown red end to end,
# and hello world's rt_sigaction count at the head against the sweep. Never lands.
set -uo pipefail
echo "box $(hostname) $(nproc) cpus, $(uname -r), tree $(git rev-parse HEAD)"
grep -E '^SigIgn' /proc/self/status
H=internal/native_runtime/c/inherited_signals_impl.h
cp "$H" "$TMPDIR/header.orig"
restore() { cp "$TMPDIR/header.orig" "$H"; }
mutate() {
	python3 -I - "$H" "$1" "$2" <<'PY'
import sys
path, before, after = sys.argv[1], sys.argv[2], sys.argv[3]
text = open(path).read()
if text.count(before) != 1:
    sys.exit("seam not found once: " + before)
open(path, "w").write(text.replace(before, after, 1))
PY
}

go vet ./internal/native > "$TMPDIR/vet.log" 2>&1 && echo "PASS vet internal/native" || { cat "$TMPDIR/vet.log"; echo "FAIL vet"; exit 1; }

go test -count=1 -v -run '^(TestInheritedIgnores|TestRuntimeStatics|TestRuntimeFieldLayouts)' ./internal/native > "$TMPDIR/native.log" 2>&1; code=$?
grep -E -- '^\s*--- (PASS|FAIL|SKIP)|caught |^(ok|FAIL)\s' "$TMPDIR/native.log" | head -80
[ $code = 0 ] && echo "PASS native: the reset's cases, its mutants, statics and layouts" || { grep -B3 -A30 -- '--- FAIL' "$TMPDIR/native.log" | head -150; echo "FAIL native"; exit 1; }

go test -count=1 -run '^TestSignal' ./internal/oracle > "$TMPDIR/signals.log" 2>&1; code=$?
grep -E -- '^\s*--- FAIL|^(ok|FAIL)\s' "$TMPDIR/signals.log" | head -40
[ $code = 0 ] && echo "PASS oracle TestSignal rows unchanged" || { grep -B3 -A30 -- '--- FAIL' "$TMPDIR/signals.log" | head -150; echo "FAIL oracle signals"; exit 1; }

# The oracle mutant: the mask road skips SIGALRM, so a parent's ignore of it survives into the program.
mutate 'if ((mask >> (signal_number - 1) & 1) != 0 && signal_number != SIGKILL' 'if ((mask >> (signal_number - 1) & 1) != 0 && signal_number != SIGALRM && signal_number != SIGKILL' || { restore; echo "FAIL mutant seam"; exit 1; }
go test -count=1 -run '^TestSignalALRM(Inherited|Default)Native$' -v ./internal/oracle > "$TMPDIR/mutant.log" 2>&1
restore
grep -E -- '^\s*--- (PASS|FAIL)' "$TMPDIR/mutant.log"
if grep -qE -- '--- FAIL: TestSignalALRMInheritedNative' "$TMPDIR/mutant.log" && grep -qE -- '--- PASS: TestSignalALRMDefaultNative' "$TMPDIR/mutant.log"; then
	grep -m1 -A2 -- '--- FAIL: TestSignalALRMInheritedNative' "$TMPDIR/mutant.log" | cut -c1-240
	echo "PASS oracle mutant skip_alrm red at TestSignalALRMInheritedNative, its default row green"
else
	tail -40 "$TMPDIR/mutant.log"; echo "FAIL oracle mutant escaped"; exit 1
fi
git diff --quiet -- "$H" && echo "PASS header restored" || { echo "FAIL header not restored"; exit 1; }

# hello world's rt_sigaction calls: the head, and the head with its mask road turned off (main's sweep), each run
# plainly and under a parent that ignores USR2 and ALRM.
if ! command -v strace > /dev/null; then
	echo "GAP no strace on this box: the counts are owed"
	echo VERDICT PASS
	exit 0
fi
printf 'console.log("hello world");\n' > "$TMPDIR/hello.a"
count() {
	strace -f -o "$TMPDIR/trace" "$@" > /dev/null 2>&1
	all=$(grep -c 'rt_sigaction(' "$TMPDIR/trace")
	reads=$(grep -c 'rt_sigaction([A-Z0-9_]*, NULL' "$TMPDIR/trace")
	opens=$(grep -c 'open.*/proc/self/status' "$TMPDIR/trace")
	echo "$all rt_sigaction, $reads of them read only, $opens opens of /proc/self/status"
}
for road in head sweep; do
	[ $road = sweep ] && { mutate 'if (inherited_ignores_mask(path, &mask)) {' 'if (false && inherited_ignores_mask(path, &mask)) {' || { restore; echo "FAIL sweep seam"; exit 1; }; }
	go build -buildvcs=false -o "$TMPDIR/adamic-$road" ./cmd/adamic > "$TMPDIR/build.log" 2>&1 || { restore; cat "$TMPDIR/build.log"; echo "FAIL adamic build"; exit 1; }
	restore
	"$TMPDIR/adamic-$road" build "$TMPDIR/hello.a" -o "$TMPDIR/hello-$road" > "$TMPDIR/b.log" 2>&1 || { cat "$TMPDIR/b.log"; echo "FAIL hello build"; exit 1; }
	echo "== $road, plain: $(count "$TMPDIR/hello-$road")"
	echo "== $road, USR2 and ALRM ignored by the parent: $(count sh -c 'trap "" USR2 ALRM; exec "$0"' "$TMPDIR/hello-$road")"
done
echo VERDICT PASS
