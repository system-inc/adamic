set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
# #snypn8c c3 on main d87e6963: a program's own comparison of a const local proves an open-enum value into a
# member slot (enumComparedMember). PART=a: gofmt, vet, the lower pins verbose, mutants M1 to M5 (each one exact
# line replaced by python, checked replaced once, run, put back from the head copy and checked put back), the
# oracle's c3 pins with enum_tag_decomposed's. PART=d: lower whole. PART=b: oracle whole and native whole. PART=c:
# stage3/fixtures, ir, compilercensus and parallelcensus whole.
part=${PART:-a}
ok=1
step() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -v "gate cache\|^=== " "$TMPDIR/step.log" | tail -60; ok=0; fi; }
whole() { name=$1; shift; "$@" > "$TMPDIR/whole.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -E "^(--- FAIL|    --- FAIL|panic:)" "$TMPDIR/whole.log" | sort -u | head -40; grep -E "_test.go:[0-9]+:" "$TMPDIR/whole.log" | head -20; grep -qE "^(--- FAIL|panic:)" "$TMPDIR/whole.log" || tail -25 "$TMPDIR/whole.log"; ok=0; fi; }
subtests() { sed -nE 's/^ *--- (PASS|FAIL|SKIP): ([^ ]+).*/\1 \2/p' "$TMPDIR/step.log" | sed 's/^SKIP \(.*\)/FAIL \1 (skipped)/'; grep -q "^ *--- SKIP" "$TMPDIR/step.log" && ok=0; }

echo "box: $(hostname) at $(git rev-parse HEAD 2>/dev/null) PART=$part"
step npm sandboxed_npm stage3/api stage3/typescript

pins='TestEnumCompared'
if [ "$part" = a ]; then
	files="internal/lower/enum_compared.go internal/lower/invariance.go"
	for f in $files; do cp "$f" "$TMPDIR/$(basename $f).head"; done
	trap 'for f in $files; do cp "$TMPDIR/$(basename $f).head" "$f"; done' EXIT
	step gofmt bash -c 'difference=$(gofmt -d internal/lower/enum_compared.go internal/lower/enum_compared_test.go internal/lower/invariance.go internal/oracle/enum_compared_test.go); printf "%s\n" "$difference"; test -z "$difference"'
	step vet go vet ./internal/lower/ ./internal/oracle/
	step "lower pins" go test -count=1 -run "$pins" ./internal/lower/ -v
	subtests
	# Every pin's error words, so a stop's wording is read, not assumed.
	grep -E "_test.go:[0-9]+:" "$TMPDIR/step.log" | head -20

	# mutant name file old new: replace exactly one line, run the lower pins, report each failing subtest, put back.
	mutant() {
		name=$1 file=$2 old=$3 new=$4
		python3 -I - "$file" "$old" "$new" <<'PY'
import sys
path, old, new = sys.argv[1:4]
text = open(path).read()
if text.count(old) != 1:
    print(f"FAIL mutant: {old!r} occurs {text.count(old)} times"); sys.exit(1)
open(path, 'w').write(text.replace(old, new))
PY
		if [ $? -ne 0 ]; then ok=0; return; fi
		go test -count=1 -run "$pins" ./internal/lower/ -v > "$TMPDIR/mutant.log" 2>&1
		code=$?
		reds=$(sed -nE 's/^ *--- FAIL: ([^ ]+).*/\1/p' "$TMPDIR/mutant.log" | tr '\n' ' ')
		if [ $code -eq 0 ]; then echo "SURVIVED $name"; else echo "RED $name: $reds"; fi
		grep -E "^# |cannot|undefined:" "$TMPDIR/mutant.log" | head -5
		cp "$TMPDIR/$(basename $file).head" "$file"
		if cmp -s "$TMPDIR/$(basename $file).head" "$file"; then echo "PASS restored $file after $name"; else echo "FAIL restore $file"; ok=0; fi
	}
	mutant M1 internal/lower/invariance.go \
		'			if l.openNumericEnumType(declared) && !l.enumMemberOrigin(node, l.enumIdentity(declared), map[*ast.Node]bool{}) && !l.enumComparedMember(node, l.enumIdentity(declared), contextual) {' \
		'			if l.openNumericEnumType(declared) && !l.enumMemberOrigin(node, l.enumIdentity(declared), map[*ast.Node]bool{}) {'
	mutant M2 internal/lower/enum_compared.go \
		'	if members == nil {' \
		'	if members == nil { members = l.definedMembers(l.checker.GetTypeAtLocation(node)) } else if false {'
	mutant M3 internal/lower/enum_compared.go \
		'	if symbol.ValueDeclaration.Kind != ast.KindVariableDeclaration || symbol.ValueDeclaration.Parent.Flags&ast.NodeFlagsConst == 0 {' \
		'	if symbol.ValueDeclaration.Kind != ast.KindVariableDeclaration {'
	mutant M4 internal/lower/enum_compared.go \
		'		if ast.IsFunctionLikeOrClassStaticBlockDeclaration(parent) || ast.IsClassLike(parent) {' \
		'		if false {'
	mutant M5 internal/lower/enum_compared.go \
		'		if index < position && len(clauses[index].AsCaseOrDefaultClause().Statements.Nodes) != 0 {' \
		'		if false {'
	mutant M6 internal/lower/enum_compared.go \
		'	case ast.KindEqualsEqualsEqualsToken, ast.KindEqualsEqualsToken:' \
		'	case ast.KindEqualsEqualsEqualsToken, ast.KindEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:'
	# Emit's shared-value stop row, probed: does tsc accept it, and what does Adamic say?
	go run ./cmd/adamic types scratch-c3/shared.a > "$TMPDIR/shared.log" 2>&1; echo "shared.a types: exit $?"; grep -v "^\s*$" "$TMPDIR/shared.log" | head -12
	go run ./cmd/adamic c scratch-c3/shared.a > /dev/null 2> "$TMPDIR/shared-c.log"; echo "shared.a c: exit $?"; head -c 1500 "$TMPDIR/shared-c.log"

	step products sandboxed_products ./internal/oracle
	step "oracle pins" go test -count=1 -run 'TestEnumComparedMemberAgreesWithNode|TestEnumTagDecomposed' ./internal/oracle/ -v
	subtests
	grep -E "_test.go:[0-9]+:" "$TMPDIR/step.log" | head -20
fi
if [ "$part" = d ]; then
	# The census products read the TypeScript source, which must be the job's own copy (sandboxed_typescript_source).
	step "typescript source" sandboxed_typescript_source
	whole "lower whole" go test -count=1 -timeout 40m ./internal/lower/
fi
if [ "$part" = b ]; then
	step products sandboxed_products ./internal/oracle
	whole "oracle whole" go test -count=1 -timeout 40m ./internal/oracle/
	whole "native whole, ADAMIC_TEST_WASI=1" env ADAMIC_TEST_WASI=1 go test -count=1 -timeout 40m ./internal/native/
fi
if [ "$part" = c ]; then
	whole "stage3/fixtures whole" go test -count=1 -timeout 40m ./stage3/fixtures/
	whole "ir whole" go test -count=1 -timeout 40m ./internal/ir/
	whole "compilercensus whole" go test -count=1 -timeout 40m ./internal/compilercensus/
	whole "parallelcensus whole" go test -count=1 -timeout 40m ./cmd/adamic-gate/parallelcensus/
fi
[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }
