"""Bounded, logged checks on the refreshed delivery head."""
import json, os, pathlib, subprocess, time
root=pathlib.Path('/workspace/adamic')
out=root/'review/compiler/generic-values-next/lint-chain-refresh'
tmp=pathlib.Path('/workspace/generic-values-proof-tmp')
results=[]
def run(name,args,cwd=root,limit=90,expected=0):
    start=time.monotonic()
    with (out/(name+'.log')).open('w') as log:
        result=subprocess.run(['timeout',str(limit),*args],cwd=cwd,stdout=log,stderr=subprocess.STDOUT)
    row=dict(name=name,command=args,cwd=str(cwd),exit=result.returncode,seconds=round(time.monotonic()-start,3),expected=expected)
    results.append(row); (out/'checks-results.json').write_text(json.dumps(results,indent=2)+'\n')
    print(json.dumps(row),flush=True)
    if result.returncode!=expected: raise SystemExit('unexpected exit: '+name)
for package in ['oracle','lower','ir','native']:
    run('build-'+package,['go','test','-c','./internal/'+package,'-o',str(tmp/('lint-refresh-'+package+'.test'))],limit=300)
run('generic-values',[str(tmp/'lint-refresh-oracle.test'),'-test.run=^TestGenericValue','-test.count=1','-test.timeout=85s','-test.v'],root/'internal/oracle')
run('json-identity',[str(tmp/'lint-refresh-native.test'),'-test.run=^TestGenericValueLibraryIdentityViews$','-test.count=1','-test.timeout=85s','-test.v'],root/'internal/native')
run('reader-guard',[str(tmp/'lint-refresh-ir.test'),'-test.run=^TestCallTargetReaders$','-test.count=1','-test.timeout=85s','-test.v'],root/'internal/ir')
names=subprocess.check_output([str(tmp/'lint-refresh-lower.test'),'-test.list=^Test'],cwd=root/'internal/lower',text=True).splitlines()
names=[n for n in names if n.startswith('Test')]
shards=[names[::2],names[1::2]]
(out/'lower-shards.json').write_text(json.dumps(shards,indent=2)+'\n')
for i,shard in enumerate(shards):
    run('lower-'+str(i+1),[str(tmp/'lint-refresh-lower.test'),'-test.run=^('+'|'.join(shard)+')$','-test.count=1','-test.timeout=85s','-test.v'],root/'internal/lower')
run('counts',[str(tmp/'lint-refresh-oracle.test'),'-test.run=^TestCountsAreRecorded$','-test.count=1','-test.timeout=600s','-test.v'],root/'internal/oracle',limit=650)
