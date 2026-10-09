import pathlib,subprocess,json,os,time,shlex
p=pathlib.Path('review/test-audit/internal-lower-non_null_impossible');names=['TestAdamicNullishAssertionsAreRefused','TestNonNullAssertionLowersToNullishPanic','TestNonNullAssertionOnPresentTypeIsErased','TestOptionalIndexingMapShapeRefused','TestOptionalIndexingKeepsIndexSignaturesRefused','TestOptionalIndexingKeepsUnsupportedStorageNotYet','TestOptionalWideningCensus','TestOptionalWideningRefused','TestOptionalWideningAllowed','TestOptionalWideningSpreadOverwrite','TestOverloadInferenceWitnesses','TestAbfe962OverrideDefaultAddedRefusesNativeSignature','TestOverrideOptionalNumberRefusesNativeSignature','TestOverrideOptionalRefusesNativeSignature'];members=names[-3:];rows=names[:-3]+['TestOverrideNativeSignature family'];(p/'scope.json').write_text(json.dumps({'names':names,'rows':rows,'family':members},indent=2));env=os.environ.copy();env['OPTIONAL_WIDENING_CONFIG']=str(p.resolve()/'census-project/tsconfig.json');env['OPTIONAL_WIDENING_OUTPUT']='/tmp/u040-census.json';records=[]
def run(label,args):
 t=time.monotonic()
 with (p/(label+'.log')).open('w') as f:r=subprocess.run(args,env=env,stdout=f,stderr=subprocess.STDOUT)
 records.append(dict(label=label,command=shlex.join(args),exit=r.returncode,wall=time.monotonic()-t));(p/'baseline-commands.json').write_text(json.dumps(records,indent=2))
run('coverage',['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverprofile='+str(p/'coverage.out'),'./internal/lower/','-run','^('+'|'.join(names)+')$'])
for n in rows:
 regex='^('+'|'.join(members)+')$' if n==rows[-1] else '^'+n+'$'
 for i in range(3):run(n.replace(' ','_')+'.'+str(i+1),['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run',regex])
