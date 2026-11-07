#!/bin/bash
# Final run for #zjndrx5 head 1's view rule on the pushed tip, then its two mutants as patched copies.
set -u
echo "== tip $(git rev-parse HEAD)"
echo "== gofmt"; gofmt -l internal/lower internal/oracle; echo gofmt-done
echo "== vet"; go vet ./internal/lower ./internal/oracle && echo vet-ok
if [ "${COLV_LIGHT:-}" = 1 ]; then
	echo "== lower Intl|SlotType|RefuseNode"; timeout 900 go test ./internal/lower -run 'Intl|SlotType|RefuseNode' -count=1 2>&1 | tail -60
	echo "== oracle Intl|Concurrency"; timeout 900 go test ./internal/oracle -run 'Intl|Concurrency' -count=1 2>&1 | tail -60
else
	echo "== lower whole"; timeout 900 go test ./internal/lower -count=1 2>&1 | tail -60
	echo "== oracle Intl|Concurrency"; timeout 900 go test ./internal/oracle -run 'Intl|Concurrency' -count=1 2>&1 | tail -60
	echo "== admission-delta tests"; timeout 600 go test ./cmd/adamic-admission-delta -count=1 2>&1 | tail -30
	echo "== flow fresh"; timeout 900 go test ./internal/flow ./internal/fresh -count=1 2>&1 | tail -30
	echo "== native"; timeout 900 go test ./internal/native -run 'Statics|HostShape|ShapeKinds|Intl|Layout' -count=1 2>&1 | tail -30
fi
fixtures=internal/oracle/testdata/intl_collator
mutant() { # $1 label, $2 sed expression on library_intl.go, $3 lower test pattern, $4 oracle test pattern
	cp internal/lower/library_intl.go /tmp/colv-tip-intl.go
	sed -i "$2" internal/lower/library_intl.go
	echo "== mutant $1: $(git diff --stat internal/lower/library_intl.go | tail -1)"
	timeout 600 go test ./internal/lower -run "$3" -count=1 -v 2>&1 | grep -E '^(=== RUN|\s*--- (PASS|FAIL)|ok|FAIL)' | grep -E -- '--- |^ok|^FAIL'
	if [ -n "$4" ]; then timeout 600 go test ./internal/oracle -run "$4" -count=1 -v 2>&1 | grep -E -- '--- (PASS|FAIL)|^ok|^FAIL|NotYet|got '; fi
	go build -o /tmp/colv-mutant ./cmd/adamic && for file in $5; do
		/tmp/colv-mutant build $file -o /tmp/colv-mutant-run > /tmp/colv-build.err 2>&1 && echo "native $file: $(/tmp/colv-mutant-run 2>&1 | tr '\n' '|')" || echo "native $file: compile $(head -c 300 /tmp/colv-build.err | tr '\n' '|')"
	done
	cp /tmp/colv-tip-intl.go internal/lower/library_intl.go
}
mutant view '/refusalRules.register(160, "intlCollatorViewRefusal"/d' 'TestIntlCollatorSeenAsAnotherTypeIsNotYet' 'TestIntlCollatorWritesAndCopiesStopByName' "$fixtures/view_field.a $fixtures/view_parameter.a"
cat > /tmp/colv-generic.a <<'PROGRAM'
const collator = new Intl.Collator(undefined, { usage: 'sort', sensitivity: 'variant', numeric: true });
function swap<T extends { compare: (x: string, y: string) => number }>(target: T): string {
	try {
		target.compare = (x: string, y: string): number => 0;
		return 'wrote';
	} catch (error) {
		return 'threw';
	}
}
console.log(swap(collator));
console.log(String(collator.compare('a', 'b')));
PROGRAM
mutant binding '/refusalRules.register(170, "intlCollatorBindingRefusal"/d' 'TestIntlCollatorSeenAsAnotherTypeIsNotYet' '' "/tmp/colv-generic.a"
echo "== tree after"; git status --short
echo "== job end"
