import subprocess,pathlib,time,json,shlex
p=pathlib.Path('review/test-audit/internal-lower-census_overload_proof');names=['TestCensusOverloadReturnProof','TestCensusPredicateInteractionBoundary','TestCensusPredicateInteractionEscapes','TestCensusPredicateMarkerKeepsProofBoundaries','TestCensusPredicateMarkerCallStaysNotYet','TestCensusOverloadRelation','TestCensusSmallFiniteKeyRead','TestCensusRestMutableElements','TestCensusBooleanDeadBranch']
(p/'scope.json').write_text(json.dumps(names,indent=2)); cmds=[]
def run(label,args):
 t=time.monotonic()
 with (p/(label+'.log')).open('w') as f:r=subprocess.run(args,stdout=f,stderr=subprocess.STDOUT)
 cmds.append(dict(label=label,command=shlex.join(args),exit=r.returncode,wall=time.monotonic()-t));(p/'baseline-commands.json').write_text(json.dumps(cmds,indent=2))
run('coverage',['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverprofile='+str(p/'coverage.out'),'./internal/lower/','-run','^('+'|'.join(names)+')$'])
for n in names:
 for i in range(3):run(n+'.'+str(i+1),['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^'+n+'$'])
