set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
# #pyak7te: the JavaScript backend writes a __proto__ field computed. PART=a: gofmt, vet, the pins verbose,
# the mutant that restores the quoted key (red on every pin, beside them green at head), internal/javascript
# whole. PART=b: internal/oracle whole and stage3/fixtures whole.
part=${PART:-a}
ok=1
step() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -v "gate cache\|^=== " "$TMPDIR/step.log" | tail -40; ok=0; fi; }
whole() { name=$1; shift; "$@" > "$TMPDIR/whole.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -E "^(--- FAIL|    --- FAIL|panic:)" "$TMPDIR/whole.log" | sort -u | head -40; grep -qE "^(--- FAIL|panic:)" "$TMPDIR/whole.log" || tail -25 "$TMPDIR/whole.log"; ok=0; fi; }
subtests() { sed -nE 's/^ *--- (PASS|FAIL|SKIP): ([^ ]+).*/\1 \2/p' "$TMPDIR/step.log"; grep -q "^ *--- SKIP" "$TMPDIR/step.log" && ok=0; }

echo "box: $(hostname) at $(git rev-parse HEAD 2>/dev/null) PART=$part"
step npm sandboxed_npm stage3/api stage3/typescript
step products sandboxed_products ./internal/oracle

if [ "$part" = a ]; then
	step gofmt bash -c 'difference=$(gofmt -d internal/javascript/javascript.go internal/oracle/proto_shorthand_test.go); printf "%s\n" "$difference"; test -z "$difference"'
	step vet go vet ./internal/javascript/ ./internal/oracle/
	step pins go test -count=1 -run 'TestShorthandProtoIsAnOwnKey' ./internal/oracle/ -v
	subtests
	file=internal/javascript/javascript.go
	cp "$file" "$TMPDIR/javascript.go.head"
	python3 -I - "$file" <<'PY'
import sys
path = sys.argv[1]
text = open(path).read()
old = '\t\treturn "[" + quote(name) + "]"\n'
if text.count(old) != 1:
    print(f"FAIL mutant: {old!r} occurs {text.count(old)} times"); sys.exit(1)
open(path, 'w').write(text.replace(old, '\t\treturn quote(name)\n'))
PY
	if [ $? -eq 0 ]; then
		go test -count=1 -run 'TestShorthandProtoIsAnOwnKey' ./internal/oracle/ -v > "$TMPDIR/mutant.log" 2>&1
		code=$?
		reds=$(sed -nE 's/^ *--- FAIL: ([^ ]+).*/\1/p' "$TMPDIR/mutant.log" | tr '\n' ' ')
		if [ $code -eq 0 ]; then echo "FAIL M1 quoted key restored SURVIVED"; ok=0; else echo "PASS M1 quoted key restored: red, $reds"; fi
		grep -E "proto_shorthand_test.go:[0-9]+:" "$TMPDIR/mutant.log" | head -4 | cut -c1-300
	else
		ok=0
	fi
	cp "$TMPDIR/javascript.go.head" "$file"
	if cmp -s "$TMPDIR/javascript.go.head" "$file"; then echo "PASS restored"; else echo "FAIL restore"; ok=0; fi
	whole "javascript whole" go test -count=1 -timeout 40m ./internal/javascript/
fi
if [ "$part" = b ]; then
	whole "oracle whole" go test -count=1 -timeout 40m ./internal/oracle/
	whole "stage3/fixtures whole" go test -count=1 -timeout 40m ./stage3/fixtures/
fi
[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }
