set -u
source /opt/adamic-tools/env.sh
export TMPDIR=/tmp/adamic-gate
cd /home/user/adamic
L=/tmp/ev
F=internal/lower/predicates.go
T='^(TestAssignedSymbolAnswersAsAFreshLowering|TestPredicateCallbackContracts)$'
sed -i '648s/range l.program.Files() {/range l.program.Files()[:1] {/' $F
git diff > $L/mutant-a.diff
go test ./internal/lower -run "$T" -count=1 -v > $L/mutant-a.log 2>&1; echo "exit=$?" >> $L/mutant-a.log
git checkout -- $F
echo "git diff after restore: [$(git diff | wc -c) bytes]" >> $L/mutant-a.log
FT='^(TestProgramRegionCensusMembership|TestProduct_ProgramRegionCensusPairwise|TestProduct_ProgramRegionCensusIndexed|TestCycleFinderCensusEquality|TestProduct_CycleScanningCensus)$'
git checkout -q 5f0b2152
echo "sha=$(git rev-parse HEAD)" > $L/run3-main-5f0b2152.log
go test ./internal/lower -run "$FT" -count=1 -v >> $L/run3-main-5f0b2152.log 2>&1; echo "exit=$?" >> $L/run3-main-5f0b2152.log
git checkout -q b65d7026ecebcca80ac12499103a40aa993345d3
echo done > $L/done2
