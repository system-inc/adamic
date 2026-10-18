import hashlib,json,os,statistics,subprocess,time
from pathlib import Path
e=Path(__file__).resolve().parent;art=e/'artifacts';env={**os.environ,'GOMAXPROCS':'4'};emission=[]
for side,root in [('base','/tmp/adamic-v4-2-base'),('v4','/tmp/adamic-json-82')]:
 compiler=art/(side+'-adamic')
 with (e/(side+'-cli-build.log')).open('w') as log:r=subprocess.run(['go','build','-o',str(compiler),'./cmd/adamic'],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT,timeout=180)
 assert r.returncode==0
 script=art/side/'program.mjs';start=time.monotonic()
 with script.open('wb') as out,(e/(side+'-javascript-emit.log')).open('wb') as err:r=subprocess.run([str(compiler),'js','stage1/cohere/json/main.ts'],cwd=root,env=env,stdout=out,stderr=err,timeout=90)
 assert r.returncode==0
 emission.append(dict(side=side,load_lower_emit_seconds=time.monotonic()-start,script_bytes=script.stat().st_size))
(e/'javascript-emission.json').write_text(json.dumps(emission,indent=2)+'\n')
empty=art/'empty-cases.txt';empty.write_text('');rows=[]
truth=json.loads((e/'run-times.json').read_text())[0]['stdout_sha256']
for mode,rounds,input_ in [('startup',10,empty),('corpus',1,art/'cases.txt')]:
 for round_ in range(rounds):
  for side in (['base','v4'] if round_%2==0 else ['v4','base']):
   root='/tmp/adamic-v4-2-base' if side=='base' else '/tmp/adamic-json-82';start=time.monotonic();output=art/(side+'-javascript-'+mode+'.stdout');error=art/(side+'-javascript-'+mode+'.stderr')
   with output.open('wb') as out,error.open('wb') as err:r=subprocess.run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(art/side/'program.mjs'),'--cases',str(input_)],cwd=root,env=env,stdout=out,stderr=err,timeout=180)
   row=dict(side=side,mode=mode,round=round_,seconds=time.monotonic()-start,exit=r.returncode,stdout_bytes=output.stat().st_size,stdout_sha256=hashlib.sha256(output.read_bytes()).hexdigest(),stderr_bytes=error.stat().st_size);rows.append(row);(e/'javascript-times.json').write_text(json.dumps(rows,indent=2)+'\n');print(row,flush=True)
   assert r.returncode==0 and error.stat().st_size==0,row
   assert row['stdout_sha256']==truth if mode=='corpus' else output.stat().st_size==0,row
print('All backend corpus outputs match source Node',flush=True)
