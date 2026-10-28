set -u
source /opt/adamic-tools/env.sh
export TMPDIR=/tmp/adamic-gate
cd /home/user/adamic
L=/tmp/ev
F=internal/lower/predicates.go
T='^(TestAssignedSymbolAnswersAsAFreshLowering|TestPredicateCallbackContracts)$'
{ echo "sha=$(git rev-parse HEAD)"; go version; echo "nproc=$(nproc)"; clang --version | head -1; node --version; uname -a; } > $L/box.log 2>&1
go test ./internal/lower -run "$T" -count=1 -v > $L/run1-tests.log 2>&1; echo "exit=$?" >> $L/run1-tests.log
mutate() { name=$1; shift
  python3 - "$F" "$@" <<'PY'
import sys
p,old,new=sys.argv[1:4]
s=open(p).read(); assert s.count(old)==1,(old,s.count(old)); open(p,'w').write(s.replace(old,new))
PY
  git diff > $L/mutant-$name.diff
  go test ./internal/lower -run "$T" -count=1 -v > $L/mutant-$name.log 2>&1; echo "exit=$?" >> $L/mutant-$name.log
  git checkout -- $F
  echo "git diff after restore: [$(git diff | wc -c) bytes]" >> $L/mutant-$name.log
}
mutate a 'for _, file := range l.program.Files() {' 'for _, file := range l.program.Files()[:1] {'
mutate b '	if l.assignedSymbols == nil {
		l.assignedSymbols = map' '	if l.assignedSymbols != nil {
		return false
	}
	if l.assignedSymbols == nil {
		l.assignedSymbols = map'
mutate c 'return l.assignedSymbols[symbol]' 'return len(l.assignedSymbols) > 0'
go test ./internal/lower -count=1 > $L/run3-lower-whole.log 2>&1; echo "exit=$?" >> $L/run3-lower-whole.log
go test ./stage3/census/latent/... -count=1 > $L/run4-latent.log 2>&1; echo "exit=$?" >> $L/run4-latent.log
echo done > $L/done
