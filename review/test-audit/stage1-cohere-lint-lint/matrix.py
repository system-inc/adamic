from run import run,E
from pathlib import Path
S=Path('/tmp/adamic-u110-mutant')
node='^TestNodeTableIsLinkOnly_[0-9]{3}$'
owned='^TestOwnedWitnesses(_Setup|Union|_[0-9]{3})$'
for label,pattern in [('switch-node-control',node),('switch-owned-control',owned)]:
 r=run(label,pattern)
 if r['cooked']:r=run(label+'-warm',pattern)
 if r['exit']!=0:raise SystemExit('red switch control')
prod='^(TestThroughput|TestNodeTableIsLinkOnly|TestNodeTableIsLinkOnly_[0-9]{3}|TestOwnedWitnessesAssignmentStable|TestOwnedWitnesses(_Setup|Union|_[0-9]{3}))$'
for mid in ['M1','M2','M3','E1']:
 S.write_text(mid)
 r=run(mid,prod)
 if r['cooked']:raise SystemExit('cooked production matrix')
S.write_text('')
fast='^(TestLegacyMutants|TestDecorationOptionMutant|TestCountGuardMutant|TestThroughput|TestNodeTableIsLinkOnlyUnionAndPlantedFailure|TestOwnedWitnessesAssignmentStable|TestOwnedWitnessesPlantedDisagreement)$'
for wid in ['W1','W2','W3','W4','W5','W6']:
 run(wid,fast,{'ADAMIC_U110_WITNESS':wid})
 if wid=='W4':run('W4-Mutants','^TestMutants$',{'ADAMIC_U110_WITNESS':wid},['-failfast'])
run('final-control',prod)
