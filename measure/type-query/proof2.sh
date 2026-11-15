#!/bin/bash
# #7yz6rtp's second proof: the row test at the head, internal/lower whole with room, the fixture's
# families and its count row, and a mutant for each of the three inTypeQuery clauses. Never lands.
set -u
T=$(mktemp -d)
echo "box $(hostname) $(nproc) cpus, at $(git rev-parse --short=10 HEAD)"
go vet ./internal/lower > $T/vet.log 2>&1 && echo "PASS vet lower" || { echo "FAIL vet"; head -20 $T/vet.log; }
rows() { L="$T/rows.$(echo "$1" | tr -c 'a-zA-Z0-9' '_').log"; go test ./internal/lower -run '^TestTypeQueryOfHostMemberIsNoUse$' -count=1 -v > "$L" 2>&1; echo "$1: exit $? | failing rows: $(grep -oE -- '--- FAIL: TestTypeQueryOfHostMemberIsNoUse/[^ ]+' "$L" | sed 's#.*/##' | tr '\n' ' ')"; grep -E 'main\.(a|ts):[0-9]+' "$L" | sed -E 's#^.*(main\.(a|ts))#\1#' | cut -c1-220 | sort -u | head -12; }
if [ "${1:-}" = whole ]; then
  go test ./internal/lower -count=1 -timeout 40m > $T/l.log 2>&1 && echo "PASS lower whole ($(tail -1 $T/l.log))" || { echo "FAIL lower whole"; grep -E -- '--- FAIL|panic: test timed out|^FAIL' $T/l.log | head -20; grep -B2 -A15 -- '--- FAIL' $T/l.log | head -60; }
  echo "== end"; exit 0
fi
rows head
go test ./internal/oracle -run '^Test[A-Za-z0-9]+_[0-9]{3}$/type_query' -count=1 > $T/o.log 2>&1 && echo "PASS type_query families" || { echo "FAIL type_query families"; grep -B3 -A20 -- '--- FAIL' $T/o.log | head -60; }

mutate() { # file old new
  python3 - "$1" "$2" "$3" <<'PY'
import sys
p,old,new=sys.argv[1:]; s=open(p).read(); assert s.count(old)==1, (p, old); open(p,'w').write(s.replace(old,new))
PY
}
A='ast.IsPartOfTypeNode(node) || inTypeQuery(node) || called(node)'; Am='ast.IsPartOfTypeNode(node) || called(node)'
B='ast.IsPartOfTypeNode(node) || inTypeQuery(node) {
		return nil
	}
	if err := l.nodeHostCallbackArgument'; Bm='ast.IsPartOfTypeNode(node) {
		return nil
	}
	if err := l.nodeHostCallbackArgument'
C='ast.IsPartOfTypeNode(node) || inTypeQuery(node) {
		return nil
	}
	if ast.IsAssignmentTarget'; Cm='ast.IsPartOfTypeNode(node) {
		return nil
	}
	if ast.IsAssignmentTarget'
# the timers clause is gone: the gate is the one guard
mutate internal/lower/library_node.go "$B" "$Bm"; rows "mutant B (gate clause dropped)"; mutate internal/lower/library_node.go "$Bm" "$B"
mutate internal/lower/host_bindings.go "$C" "$Cm"; rows "mutant C (host binding clause dropped)"; mutate internal/lower/host_bindings.go "$Cm" "$C"
git diff --quiet internal/lower && echo "PASS mutants undone" || { echo "FAIL mutants not undone"; git diff --stat internal/lower; }
# internal/lower whole runs as its own job: proof2.sh whole
echo "== end"
