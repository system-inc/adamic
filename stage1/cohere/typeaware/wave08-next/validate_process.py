"""Native process-rule production comparisons and byte-only mutants.

Build checker archives with overlay_bridge.py first. All logs are files;
this script does not modify the shared bridge dispatcher or test harness.
"""
import argparse
import gzip
import hashlib
import json
import os
import pathlib
import shutil
import statistics
import subprocess
import time

parser=argparse.ArgumentParser()
for name in ['artifacts','fixtures','stage0','archive','asan-archive','compiler-root']:
    parser.add_argument('--'+name,required=True)
args=parser.parse_args()
source=pathlib.Path(__file__).resolve().parent
repository=source.parents[3]
root=pathlib.Path(args.artifacts).resolve();root.mkdir(parents=True,exist_ok=True)
fixtures=pathlib.Path(args.fixtures).resolve()
records=[]
def run(name,command,expected=0,cwd=repository,env=None):
    with open(root/(name+'.stdout'),'wb') as out,open(root/(name+'.stderr'),'wb') as err:
        start=time.perf_counter();p=subprocess.run([str(a) for a in command],cwd=cwd,stdout=out,stderr=err,env=env);elapsed=time.perf_counter()-start
    output=(root/(name+'.stdout')).read_bytes();errors=(root/(name+'.stderr')).read_bytes()
    assert p.returncode==expected,(name,p.returncode,errors.decode())
    return output,errors,elapsed
virtual=repository/'cohere/adamic_wave08_process_oracle.go'
(root/'oracle-overlay.json').write_text(json.dumps({'Replace':{str(virtual):str(source/'testdata/oracle_process.go')}}))
run('oracle-build',['go','build','-overlay',root/'oracle-overlay.json','-o',root/'oracle',virtual],cwd=repository/'cohere')
for binary,archive,sanitize in [('native',args.archive,False),('native-asan',args.asan_archive,True)]:
    command=[args.stage0,'build',source/'suite.a','-o',root/binary,'--tsgo',archive]
    if sanitize:command+=['--sanitize']
    run(binary+'-build',command)
for case in sorted((fixtures/'cases').iterdir()):
    meta=json.loads((case/'metadata.json').read_text());name='fixture-'+case.name
    config=case/'tsconfig.json';manifest=case/'roots.manifest'
    truth,_,_=run(name+'-go',[root/'oracle',config,manifest,'--all'])
    for variant in ['native','native-asan']:
        output,errors,_=run(name+'-'+variant,[root/variant,config,manifest,'--all'])
        assert output==truth,(meta['Name'],variant)
        assert errors==b'',(meta['Name'],variant,errors)
    records.append(dict(corpus='upstream',case=case.name,name=meta['Name'],bytes=len(truth),sha256=hashlib.sha256(truth).hexdigest(),findings=truth.splitlines()[-1].decode()))
print('all upstream fixture programs agree, normal and ASan/UBSan:',len(records),flush=True)
for corpus,prefix in [('compiler',args.compiler_root),('repository',repository)]:
    paths=(source.parent/'validation-coverage'/f'{corpus}.manifest').read_text().splitlines()
    manifest=root/(corpus+'.manifest');manifest.write_text(''.join(str(pathlib.Path(prefix)/path)+'\n' for path in paths))
    config=pathlib.Path(args.compiler_root)/'src/compiler/tsconfig.json' if corpus=='compiler' else repository/'tsconfig.json'
    truth,_,_=run(corpus+'-go',[root/'oracle',config,manifest,'--all'])
    for variant in ['native','native-asan']:
        output,errors,_=run(corpus+'-'+variant,[root/variant,config,manifest,'--all']);assert output==truth and errors==b'',(corpus,variant,errors)
        records.append(dict(corpus=corpus,variant=variant,files=len(paths),bytes=len(truth),sha256=hashlib.sha256(truth).hexdigest(),findings=truth.splitlines()[-1].decode()))
    print(corpus,len(paths),'files, byte equality, normal and ASan/UBSan:',len(truth),truth.splitlines()[-1].decode(),flush=True)
    for round in range(3):
        for variant in (['oracle','native'] if round%2==0 else ['native','oracle']):
            output,errors,elapsed=run(f'timing-{corpus}-{round}-{variant}',[root/variant,config,manifest,'--all'],env=dict(os.environ,ADAMIC_TSGO_TIMING='1'));assert output==truth
            records.append(dict(corpus=corpus,variant=variant,round=round,seconds=elapsed,stderr=errors.decode()))
mutants=[('process-callee','no_process_exit_after_output.a','if(writers.get(key) ?? false)', 'if(false)'),('blocking-order','require_blocking_standard_streams.a','this.ordered = index.canBlock.has(this.rules.path);','this.ordered = false;')]
for name,file,before,after in mutants:
    directory=root/(name+'-source');directory.mkdir(exist_ok=True)
    for path in source.glob('*.a'):
        text=path.read_text().replace("'../", "'"+str(source.parent)+"/")
        if path.name==file:
            assert text.count(before)==1,(name,before);text=text.replace(before,after)
        (directory/path.name).write_text(text)
    run(name+'-build',[args.stage0,'build',directory/'suite.a','-o',root/name,'--tsgo',args.archive])
    caught=[]
    for case in sorted((fixtures/'cases').iterdir()):
        output,errors,_=run(name+'-'+case.name,[root/name,case/'tsconfig.json',case/'roots.manifest','--all'])
        assert errors==b''
        truth=(root/('fixture-'+case.name+'-go.stdout')).read_bytes()
        if output!=truth:
            byte=next((i for i,(a,b) in enumerate(zip(output,truth)) if a!=b),min(len(output),len(truth)))
            caught.append(dict(case=case.name,byte=byte,name=json.loads((case/'metadata.json').read_text())['Name']))
    assert caught,(name,'production byte oracle did not catch mutant')
    records.append(dict(mutant=name,exit=0,stderr='',caught=caught));print(name,'compiled, exits 0, empty stderr; production byte oracle catches',len(caught),'programs, first byte',caught[0]['byte'],flush=True)
    (root/name).unlink()
run('handle-build',[args.stage0,'build',source/'testdata/handle_probe.a','-o',root/'handle','--tsgo',args.archive,'--sanitize'])
witness=next((fixtures/'cases').iterdir());path=pathlib.Path((witness/'roots.manifest').read_text().splitlines()[0])
for question in ['wave08-resolved-callee','wave08-program-loads','wave08-symbol-ancestry']:
    output,errors,_=run('released-'+question,[root/'handle',witness/'tsconfig.json',path,question,'--release'],expected=70)
    assert output==b'' and b'invalid or released checker handle' in errors
    print(question,'released handle rejected, exit 70',flush=True)
(root/'results.json').write_text(json.dumps(records,indent=2)+'\n')
for corpus in ['compiler','repository']:
    med={variant:statistics.median(record['seconds'] for record in records if record.get('round') is not None and record['corpus']==corpus and record['variant']==variant) for variant in ['oracle','native']}
    print(corpus,'median seconds',med,'native/Go',med['native']/med['oracle'],flush=True)
print('PASS',flush=True)
