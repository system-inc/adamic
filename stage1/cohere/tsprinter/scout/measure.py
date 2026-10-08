"""Measure complete files on source Node; manifest is mandatory, pins verified, no sampling."""
import argparse,collections,hashlib,json,subprocess
from pathlib import Path
root=Path(__file__).resolve().parents[4]
p=argparse.ArgumentParser();p.add_argument('manifest',type=Path);p.add_argument('output',type=Path);p.add_argument('--before',action='store_true');p.add_argument('--import-stage',choices=['default','namespace','named','type','attributes']);a=p.parse_args()
manifest=json.loads(a.manifest.read_text());records=[];counts=collections.Counter();escape=lambda s:s.replace('\\','\\\\').replace('\n','\\n').replace('\r','\\r').replace('\t','\\t')
for repo in manifest:
 if 'error' in repo: raise RuntimeError(repo)
 checkout=Path(repo['checkout'])
 actual=subprocess.check_output(['git','rev-parse','HEAD'],cwd=checkout,text=True).strip()
 if actual!=repo['pin']:raise RuntimeError('wrong pin '+repo['repo'])
 selected=subprocess.check_output(['git','ls-files','-z'],cwd=checkout).decode().split('\0')
 selected=[f for f in selected if f.endswith(('.ts','.tsx'))]
 if selected!=repo['files']:raise RuntimeError('manifest changed '+repo['repo'])
 for start in range(0,len(selected),1000):
  files=selected[start:start+1000];sources=[(checkout/f).read_bytes() for f in files]
  inputs=[];valid=[]
  for f,source in zip(files,sources):
   try:text=source.decode('utf8')
   except UnicodeDecodeError:
    records.append(dict(repo=repo['repo'],pin=actual,path=f,sha256=hashlib.sha256(source).hexdigest(),outcome='input-not-utf8'));counts['input-not-utf8']+=1;continue
   inputs.append(dict(source=text,path=f));valid.append((f,source))
  batch=a.output.with_suffix('.batch.json');batch.write_text(json.dumps(inputs))
  pending=valid
  pending_inputs=inputs
  while pending:
   batch.write_text(json.dumps(pending_inputs))
   try:
    result=subprocess.run(['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(root/('stage1/cohere/tsprinter/scout/stages.mjs' if a.import_stage else 'stage1/cohere/tsprinter/scout/before.mjs' if a.before else 'stage1/cohere/tsprinter/scout/measure.mjs')),str(batch),*( [a.import_stage] if a.import_stage else [] )],cwd=root,capture_output=True,text=True,timeout=120)
    answers=result.stdout.splitlines()
    failure='node-exit-'+str(result.returncode)+':'+(result.stderr.splitlines()[0] if result.stderr else 'no-stderr') if result.returncode else ''
   except subprocess.TimeoutExpired as error:
    output=error.stdout or b''
    if isinstance(output,bytes):output=output.decode('utf8',errors='replace')
    answers=output.splitlines();failure='node-timeout'
   if len(answers)>len(pending):raise RuntimeError('extra answers')
   for (f,source),answer in zip(pending,answers):
    answer=json.loads(answer);outcome=answer['reason'] if answer['kind']=='NotYet' else 'accepted'
    counts[outcome]+=1;records.append(dict(repo=repo['repo'],pin=actual,path=f,sha256=hashlib.sha256(source).hexdigest(),outcome=outcome,answer=answer['text'] if answer['kind']=='Ok' else None))
   consumed=len(answers)
   if consumed<len(pending):
    if not failure:raise RuntimeError('missing answer on successful run')
    f,source=pending[consumed]
    batch.write_text(json.dumps(pending_inputs[consumed:consumed+1]))
    try:
     isolated=subprocess.run(['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(root/('stage1/cohere/tsprinter/scout/stages.mjs' if a.import_stage else 'stage1/cohere/tsprinter/scout/before.mjs' if a.before else 'stage1/cohere/tsprinter/scout/measure.mjs')),str(batch),*( [a.import_stage] if a.import_stage else [] )],cwd=root,capture_output=True,text=True,timeout=120)
     if isolated.returncode==0:
      recovered=json.loads(isolated.stdout);outcome=recovered.get('reason','accepted');counts[outcome]+=1
      records.append(dict(repo=repo['repo'],pin=actual,path=f,sha256=hashlib.sha256(source).hexdigest(),outcome=outcome,answer=recovered.get('text')));consumed+=1
      pending=pending[consumed:];pending_inputs=pending_inputs[consumed:];continue
     failure='node-exit-'+str(isolated.returncode)+':'+(isolated.stderr.splitlines()[0] if isolated.stderr else 'no-stderr')
    except subprocess.TimeoutExpired:
     failure='node-timeout'
    counts[failure]+=1
    records.append(dict(repo=repo['repo'],pin=actual,path=f,sha256=hashlib.sha256(source).hexdigest(),outcome=failure,answer=None));consumed+=1
   pending=pending[consumed:];pending_inputs=pending_inputs[consumed:]
 print(repo['repo'],len(selected),dict(counts),flush=True)
 a.output.write_text(json.dumps(dict(corpus='23 public pinned repositories supplied by user',counts=counts,files=records),indent=2)+'\n')
print('complete',len(records),dict(counts),flush=True)
