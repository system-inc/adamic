#!/bin/bash
# #7z36qn6 libm --as-needed before review, sealed: vet, TestMathLinksLibmOnlyWhereCalled verbose in every build, the
# mutant that names -lm plainly again (hello's subtests must go red), internal/native_runtime and internal/native whole.
# Never lands.
set -uo pipefail
echo "box $(hostname) $(nproc) cpus, $(uname -r), load $(cut -d' ' -f1-3 /proc/loadavg), tree $(git rev-parse HEAD)"
T="$TMPDIR/asneeded"
mkdir -p "$T"
L=internal/native_runtime/library.go
cp "$L" "$T/library.go.orig"
go vet ./internal/native ./internal/native_runtime ./internal/oracle > "$T/vet.log" 2>&1 && echo "PASS vet" || { cat "$T/vet.log"; echo "FAIL vet"; exit 1; }
go test -count=1 -v -run '^TestMathLinksLibmOnlyWhereCalled$' ./internal/oracle > "$T/pin.log" 2>&1; code=$?
grep -E -- '^\s*--- (PASS|FAIL|SKIP)|math_link_test.go' "$T/pin.log"
[ $code = 0 ] && echo "PASS the pin, every build" || { tail -40 "$T/pin.log"; echo "FAIL the pin"; exit 1; }

python3 -I - "$L" <<'PY' || { cp "$T/library.go.orig" "$L"; echo "FAIL mutant seam"; exit 1; }
import sys
path = sys.argv[1]
text = open(path).read()
seam = '\treturn []string{"-Wl,--as-needed", "-lm", "-Wl,--no-as-needed"}\n'
if text.count(seam) != 1:
    sys.exit("seam not found once")
open(path, "w").write(text.replace(seam, '\treturn []string{"-lm"}\n', 1))
PY
go test -count=1 -v -run '^TestMathLinksLibmOnlyWhereCalled$' ./internal/oracle > "$T/mutant.log" 2>&1
cp "$T/library.go.orig" "$L"
git diff --quiet -- "$L" || { echo "FAIL library.go not restored"; exit 1; }
grep -E -- '^\s*--- (PASS|FAIL)' "$T/mutant.log"
red=$(grep -cE -- '--- FAIL: TestMathLinksLibmOnlyWhereCalled/hello.a/(plain|counted|release|split)' "$T/mutant.log")
if [ "$red" = 4 ] && ! grep -qE -- '--- FAIL: TestMathLinksLibmOnlyWhereCalled/fmod.a' "$T/mutant.log"; then
	grep -m1 'calls nothing of libm' "$T/mutant.log" | cut -c1-220
	echo "PASS mutant -lm named plainly is red at hello's four held builds, and fmod's stay green"
else
	tail -30 "$T/mutant.log"; echo "FAIL mutant: $red of hello's four held builds red"; exit 1
fi

for package in ./internal/native_runtime ./internal/native; do
	go test -count=1 -timeout 0 "$package" > "$T/whole.log" 2>&1; code=$?
	grep -E -- '^(--- FAIL|ok |FAIL)' "$T/whole.log" | head -20
	[ $code = 0 ] && echo "PASS $package whole" || { grep -B3 -A25 -- '--- FAIL' "$T/whole.log" | head -120; echo "FAIL $package"; exit 1; }
done
echo VERDICT PASS
