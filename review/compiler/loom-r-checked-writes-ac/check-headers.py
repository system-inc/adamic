import pathlib,json,subprocess,time,hashlib,runpy,types,collections
ROOT=pathlib.Path.cwd();OUT=ROOT/'review/compiler/loom-r-checked-writes-ac'; BINARY=OUT/'adamic-acheck'
paths=json.loads((OUT/'paths.json').read_text())
# Update only first-line gate metadata, preserving every byte of the program below it.
previous=json.loads((OUT/'status-before.json').read_text());changes=[]
for row in previous:
 p=ROOT/row['path'];raw=p.read_bytes();first,sep,rest=raw.partition(b'\n');has=first.startswith(b'// a-check:');body=rest if has else raw
 if row['status']=='unclassified' and "can't lower" in row['reason'] and ' yet' in row['reason']:row['status']='NotYet'
 assert row['status']!='unclassified',row
 header=('// a-check: refused '+row['reason']).encode() if row['status']=='refused' else ('// a-check: type error '+row['reason']).encode() if row['status']=='type error' else None
 changed=header+b'\n'+body if header else body
 row['body_sha256']=hashlib.sha256(body).hexdigest();row['header_before']=first.decode() if has else None;row['header_after']=header.decode() if header else None
 if changed!=raw:
  p.write_bytes(changed);changes.append(dict(path=row['path'],before=row['header_before'],after=row['header_after'],body_sha256=row['body_sha256'],line_shift=(int(header is not None)-int(has))))
 assert (p.read_bytes().split(b'\n',1)[1] if header else p.read_bytes())==body
(OUT/'status-before.json').write_text(json.dumps(previous,indent=2)+'\n');(OUT/'headers.json').write_text(json.dumps(changes,indent=2)+'\n')
namespace=runpy.run_path(str(OUT/'gate.py'));Gate=next(x for x in namespace.values() if isinstance(x,type) and hasattr(x,'aCheck'));captured={}
class Process:
 def __init__(self,command):self.command=command;self.started=time.monotonic();self.process=subprocess.Popen(['timeout','30',str(BINARY),*command[1:]],cwd=ROOT,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
 def communicate(self):
  out,err=self.process.communicate(timeout=35);self.returncode=self.process.returncode;p=self.command[-1];captured[p]=dict(stdout=out,stderr=err,exit=self.returncode,seconds=round(time.monotonic()-self.started,3));i=paths.index(p);(OUT/f'final-{i:02d}.stdout.log').write_text(out);(OUT/f'final-{i:02d}.stderr.log').write_text(err);return out,err
class Harness:
 def __init__(self):self.arguments=types.SimpleNamespace(tree=str(ROOT));self.result={};self.steps={};self.exits={};self.failures=[]
 def step(self,name,command):assert name=='a-check-build';return True # The same branch binary was explicitly built already.
 def spawn(self,command,stdout,stderr):return Process(command)
 def fail(self,name,message):self.failures.append(dict(step=name,message=message))
h=Harness();Gate.aCheck(h,paths);assert h.exits['a-check']==0,h.failures
(OUT/'a-check.json').write_text(json.dumps(dict(results=h.result,exits=h.exits,failures=h.failures),indent=2)+'\n')
rows=[]
for row in previous:
 p=row['path'];obs=captured[p];raw=(ROOT/p).read_bytes();header=row['header_after'];body=raw.split(b'\n',1)[1] if header else raw
 assert hashlib.sha256(body).hexdigest()==row['body_sha256']
 row.update(exit=obs['exit'],seconds=obs['seconds'],diagnostic=obs['stderr'],sha256=hashlib.sha256(raw).hexdigest());rows.append(row)
(OUT/'status.json').write_text(json.dumps(rows,indent=2)+'\n')
# Captured real diagnostics isolate the unchanged Gate.aCheck predicate, as in the README.
class Captured:
 def __init__(self,row):self.row=row;self.returncode=row['exit']
 def communicate(self):return self.row['stdout'],self.row['stderr']
class Replay(Harness):
 def spawn(self,command,stdout,stderr):return Captured(captured[command[-1]])
selected=next(x['path'] for x in rows if x['status']=='refused');p=ROOT/selected;original=p.read_bytes();first,body=original.split(b'\n',1);mutants=[]
try:
 for name,data in [('removed-header',body),('wrong-reason',b'// a-check: refused deliberately unrelated header mutant\n'+body)]:
  p.write_bytes(data);g=Replay();Gate.aCheck(g,[selected]);assert g.exits['a-check']==1 and len(g.failures)==1,(name,g.result)
  mutants.append(dict(name=name,path=selected,caught_by='unchanged Gate.aCheck over captured real branch compiler diagnostics',exits=g.exits,failures=g.failures,observations=g.result));p.write_bytes(original)
finally:p.write_bytes(original)
assert p.read_bytes()==original
restored=Replay();Gate.aCheck(restored,paths);assert restored.exits['a-check']==0
(OUT/'header-mutants.json').write_text(json.dumps(mutants,indent=2)+'\n')
print('PASS actual Gate.aCheck',len(paths),'files;',dict(collections.Counter(x['status']for x in rows)),';',len(changes),'header edits; both mutants caught; restored predicate passes')
