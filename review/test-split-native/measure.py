import pathlib,re,subprocess,os,time,json,tempfile
root=pathlib.Path(__file__).resolve().parents[2];out=root/'review/test-split-native/cold';out.mkdir(parents=True,exist_ok=True)
names=['TestArtifactInputsAndReuse','TestShardSelection','TestRecordReadMutants/prototype-membership-restored','TestRecordReadMutants/missing-read-silent','TestRecordReadMutants/own-read-checked-as-missing','TestRetainedSplitCoverage','TestRetainedTopLevelCoverage']
for prefix,n in [('TestRegExpBytecodeRandomNodeUnit',40),('TestRecordMutantsUnit',6),('TestSplitTSGoAgreesUnit',2),('TestWASIUnit',36)]:names += [prefix+f'{i:02}' for i in range(n)]
import sys
if len(sys.argv)>1: names=[n for n in names if any(n.startswith(prefix) for prefix in sys.argv[1:])]
go_cache=subprocess.check_output(['go','env','GOCACHE'],text=True).strip()
for name in names:
 env=os.environ.copy();env.update(GOCACHE=go_cache,GOMAXPROCS='4',ADAMIC_GATE_UNCACHED='1',XDG_CACHE_HOME=tempfile.mkdtemp(prefix='native-cold-'),ADAMIC_BUILD_CACHE_DIR=tempfile.mkdtemp(prefix='native-products-'))
 if name.startswith('TestWASIUnit'):
  env['ADAMIC_TEST_WASI']='1'
  env['PATH']=str(pathlib.Path(env['WASI_SYSROOT']).parents[1]/'bin')+os.pathsep+env['PATH']
 cmd=['go','test','./internal/native','-run','/'.join('^'+re.escape(part)+'$' for part in name.split('/')),'-count=1','-parallel=4','-timeout=90s','-json'];start=time.monotonic()
 with (out/(name.replace('/', '__')+'.jsonl')).open('w') as f:r=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=f)
 events=[]
 for line in (out/(name.replace('/', '__')+'.jsonl')).read_text().splitlines():
  try:events.append(json.loads(line))
  except:pass
 terminal=[e for e in events if e.get('Test')==name and e.get('Action') in ['pass','skip','fail']]
 row=dict(test=name,wall=time.monotonic()-start,exit=r.returncode,result=terminal[-1] if terminal else None,command=cmd)
 with (out/'results.jsonl').open('a') as f:f.write(json.dumps(row)+'\n')
 print(name,row['wall'],row['result'],flush=True)
