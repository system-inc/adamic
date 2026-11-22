#!/bin/bash
# #7z36qn6 lever (a) before Loom, sealed: vet, the oracle's TestSignal rows whole with the two new ones, the runtime
# statics' signal tests, two mutants that must go red (SIGUSR1 or SIGXFSZ added to stop_signals) and one control that
# must stay green (the reset moved after the handlers, which no longer matters), and hello world's sigaction count.
# Never lands.
set -uo pipefail
echo "box $(hostname) $(nproc) cpus, $(uname -r), load $(cut -d' ' -f1-3 /proc/loadavg), tree $(git rev-parse HEAD)"
T="$TMPDIR/reads"
mkdir -p "$T"
A=internal/native_runtime/c/adamic.c
cp "$A" "$T/adamic.c.orig"
restore() { cp "$T/adamic.c.orig" "$A"; }
mutate() {
	python3 -I - "$A" "$1" "$2" <<'PY'
import sys
path, before, after = sys.argv[1], sys.argv[2], sys.argv[3]
text = open(path).read()
if text.count(before) != 1:
    sys.exit("seam not found once: " + before)
open(path, "w").write(text.replace(before, after, 1))
PY
}
go vet ./internal/oracle ./internal/native > "$T/vet.log" 2>&1 && echo "PASS vet" || { cat "$T/vet.log"; echo "FAIL vet"; exit 1; }
go test -count=1 -v -run '^TestSignal' ./internal/oracle > "$T/signals.log" 2>&1; code=$?
grep -cE -- '^--- PASS' "$T/signals.log" | sed 's/^/passed rows: /'
grep -E -- '^--- (PASS|FAIL): TestSignal(USR1|XFSZ)Ignored|^--- FAIL|^(ok|FAIL)\s' "$T/signals.log"
[ $code = 0 ] && echo "PASS oracle TestSignal rows" || { grep -B3 -A20 -- '--- FAIL' "$T/signals.log" | head -100; echo "FAIL oracle signals"; exit 1; }
go test -count=1 -run '^TestRuntimeStatics(SignalAndExit|StopWholeLines|StorageScanner|AreListed)$' ./internal/native > "$T/statics.log" 2>&1; code=$?
grep -E -- '^--- FAIL|^(ok|FAIL)\s' "$T/statics.log"
[ $code = 0 ] && echo "PASS runtime statics' signal tests" || { grep -B3 -A20 -- '--- FAIL' "$T/statics.log" | head -100; echo "FAIL statics"; exit 1; }

signals='	SIGTERM, SIGINT, SIGHUP, SIGQUIT, SIGUSR2,'
for added in SIGUSR1 SIGXFSZ; do
	name=${added#SIG}
	mutate "$signals" "$signals $added," || { restore; echo "FAIL mutant seam"; exit 1; }
	go test -count=1 -v -run "^TestSignal${name}IgnoredByTheProgramNative\$" ./internal/oracle > "$T/mutant-$name.log" 2>&1
	restore
	if grep -qE -- "^--- FAIL: TestSignal${name}IgnoredByTheProgramNative" "$T/mutant-$name.log"; then
		grep -m1 -A1 -- "--- FAIL" "$T/mutant-$name.log" | tail -1 | cut -c1-220
		echo "PASS mutant $added in stop_signals is red at TestSignal${name}IgnoredByTheProgramNative"
	else
		tail -20 "$T/mutant-$name.log"; echo "FAIL mutant $added escaped"; exit 1
	fi
done

# The control: the reset after the handlers. With no read in stop_with, the order no longer matters, so every
# inherited row stays green.
reset='	inherited_ignores_reset("/proc/self/status", &default_action);
'
mutate "$reset" "" || { restore; echo "FAIL control seam"; exit 1; }
mutate '		stop_with(signal_number);
	}
#endif
' '		stop_with(signal_number);
	}
#endif
	inherited_ignores_reset("/proc/self/status", &default_action);
' || { restore; echo "FAIL control seam"; exit 1; }
go test -count=1 -run '^TestSignal.*Inherited' ./internal/oracle > "$T/control.log" 2>&1; code=$?
restore
grep -E -- '^--- FAIL|^(ok|FAIL)\s' "$T/control.log"
[ $code = 0 ] && echo "PASS control: the reset after the handlers keeps every inherited row green" || echo "NOTE control red: the order is still a premise, read $T/control.log"
git diff --quiet -- "$A" && echo "PASS adamic.c restored" || { echo "FAIL adamic.c not restored"; exit 1; }

if command -v strace > /dev/null; then
	go build -buildvcs=false -o "$T/adamic" ./cmd/adamic > "$T/b.log" 2>&1 || { cat "$T/b.log"; echo "FAIL build"; exit 1; }
	printf 'console.log("hello world");\n' > "$T/hello.a"
	"$T/adamic" build "$T/hello.a" -o "$T/hello" > "$T/b.log" 2>&1 || { cat "$T/b.log"; echo "FAIL hello"; exit 1; }
	strace -f -o "$T/trace" "$T/hello" > /dev/null
	echo "hello world: $(grep -c 'rt_sigaction(' "$T/trace") rt_sigaction, $(grep -c 'rt_sigaction([A-Z0-9_]*, NULL' "$T/trace") of them read only"
fi
echo VERDICT PASS
