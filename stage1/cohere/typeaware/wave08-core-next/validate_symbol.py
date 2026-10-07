"""Isolated wave 08 core validation; leaves the shared harness untouched."""
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
for name in ['artifacts','stage0','archive','asan-archive','compiler-root']:
 parser.add_argument('--'+name,required=True)
args=parser.parse_args()
source=pathlib.Path(__file__).resolve().parent
repository=source.parents[3]
root=pathlib.Path(args.artifacts).resolve();root.mkdir(parents=True,exist_ok=True)
records=[]
def run(name,arguments,expected=0,cwd=repository,env=None):
 with open(root/(name+'.stdout'),'wb') as out,open(root/(name+'.stderr'),'wb') as err:
  start=time.perf_counter();process=subprocess.run([str(a) for a in arguments],stdout=out,stderr=err,cwd=cwd,env=env);elapsed=time.perf_counter()-start
 output=(root/(name+'.stdout')).read_bytes();errors=(root/(name+'.stderr')).read_bytes()
 assert process.returncode==expected,(name,process.returncode,errors.decode())
 return output,errors,elapsed
for p in (source/'testdata/controls').glob('*.a'):shutil.copyfile(p,root/p.name)
(root/'alias.d.ts').write_text((source/'testdata/alias.txt').read_text())
(root/'tsconfig.json').write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','module':'NodeNext','lib':['ES2022'],'noEmit':True},'include':['*.ts']}))
(root/'controls.manifest').write_text(''.join(str(p)+'\n' for p in sorted(root.glob('control-*.a'))))
for corpus,prefix in [('compiler',args.compiler_root),('repository',str(repository))]:
 paths=(source.parent/'validation-coverage'/f'{corpus}.manifest').read_text().splitlines()
 (root/f'{corpus}.manifest').write_text(''.join(str(pathlib.Path(prefix)/p)+'\n' for p in paths))
virtual=repository/'cohere/adamic_wave08_next_oracle.go'
(root/'oracle-overlay.json').write_text(json.dumps({'Replace':{str(virtual):str(source/'testdata/oracle.go')}}))
run('oracle-build',['go','build','-overlay',root/'oracle-overlay.json','-o',root/'oracle',virtual],cwd=repository/'cohere')
for binary,archive,sanitize in [('native',args.archive,False),('native-asan',args.asan_archive,True)]:
 command=[args.stage0,'build',source/'suite.a','-o',root/binary,'--tsgo',archive]
 if sanitize:command+=['--sanitize']
 run(binary+'-build',command)
for corpus,config in [('controls',root/'tsconfig.json'),('compiler',pathlib.Path(args.compiler_root)/'src/compiler/tsconfig.json'),('repository',repository/'tsconfig.json')]:
 manifest=root/f'{corpus}.manifest';truth,_,_=run(corpus+'-go',[root/'oracle',config,manifest])
 for variant in ['native','native-asan']:
  output,errors,elapsed=run(corpus+'-'+variant,[root/variant,config,manifest]);assert output==truth,(corpus,variant)
  assert errors==b'',(corpus,variant,errors)
  records.append(dict(corpus=corpus,variant=variant,bytes=len(output),sha256=hashlib.sha256(output).hexdigest(),findings=output.splitlines()[-1].decode()))
  print(corpus,variant,len(output),output.splitlines()[-1].decode(),flush=True)
 for round in range(3):
  for variant in (['oracle','native'] if round%2==0 else ['native','oracle']):
   output,errors,elapsed=run(f'timing-{corpus}-{round}-{variant}',[root/variant,config,manifest],env=dict(os.environ,ADAMIC_TSGO_TIMING='1'));assert output==truth
   records.append(dict(corpus=corpus,variant=variant,round=round,seconds=elapsed,stderr=errors.decode()))
mutant=root/'mutant-source';mutant.mkdir(exist_ok=True)
for p in source.glob('*.a'):
 text=p.read_text().replace("'../", "'"+str(source.parent)+"/")
 if p.name=='symbol_description.a':
  before="callee.text !== 'Symbol'";assert text.count(before)==1;text=text.replace(before,"callee.text === 'Symbol'")
 (mutant/p.name).write_text(text)
run('mutant-build',[args.stage0,'build',mutant/'suite.a','-o',root/'mutant','--tsgo',args.archive])
output,errors,_=run('mutant-run',[root/'mutant',root/'tsconfig.json',root/'controls.manifest']);truth=(root/'controls-go.stdout').read_bytes()
assert output!=truth and errors==b''
first=next((i for i,(a,b) in enumerate(zip(output,truth)) if a!=b),min(len(output),len(truth)))
print('symbol mutant compiled, exits 0, empty stderr; byte oracle catches',first,flush=True)
run('handle-build',[args.stage0,'build',source/'testdata/handle_probe.a','-o',root/'handle','--tsgo',args.archive])
for question in ['wave08-symbol-origins']:
 output,errors,_=run('released-'+question,[root/'handle',root/'tsconfig.json',root/'control-00.a',question,'--release'],expected=70)
 assert output==b'' and b'invalid or released checker handle' in errors
 print(question,'released handle refused with exit 70',flush=True)
(root/'results.json').write_text(json.dumps(records,indent=2)+'\n')
for corpus in ['compiler','repository']:
 med={v:statistics.median(r['seconds'] for r in records if r.get('round') is not None and r['corpus']==corpus and r['variant']==v) for v in ['oracle','native']}
 print(corpus,'median seconds',med,'native/Go',med['native']/med['oracle'],flush=True)
(root/'mutant').unlink()
print('PASS',flush=True)
