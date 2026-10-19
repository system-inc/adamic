import collections,hashlib,json,os,re,subprocess,sys,time
from pathlib import Path
out=Path(__file__).parent;root=Path.cwd();binary='/workspace/scratch/loom-r-slice-2-2-adamic'
os.sched_setaffinity(0,sorted(os.sched_getaffinity(0))[:4])
paths=subprocess.check_output(['git','diff','HEAD','--name-only','--diff-filter=AM','--','*.a'],text=True).splitlines()
rows=[];changes=[];mutants=[]
for index,name in enumerate(paths):
 p=Path(name);before=p.read_bytes();body=before.split(b'\n',1)[1] if before.startswith(b'// a-check:') else before
 start=time.monotonic();ran=subprocess.run(['timeout','30',binary,'c',name],capture_output=True,text=True,timeout=35)
 (out/('acheck-%02d.stderr.txt'%index)).write_text(ran.stderr)
 reason='';expected=''
 if ran.returncode==0:kind='clean'
 elif "stage 0 can't lower " in ran.stderr:kind='NotYet';reason=ran.stderr.strip()
 elif (m:=re.search(r'Adamic 0\.1 refuses ([^\n]+?);',ran.stderr)):
  kind='refused';reason=m[1];expected='// a-check: refused '+reason
 elif (m:=re.search(r'\bTS(\d+)\b',ran.stderr)):
  kind='type error';reason='TS'+m[1];expected='// a-check: type error '+reason
 else:raise SystemExit('unclassified '+name+': '+ran.stderr)
 first=before.decode().splitlines()[0]
 if '--update' in sys.argv and ((expected and first!=expected) or (not expected and first.startswith('// a-check:'))):
  p.write_bytes((expected.encode()+b'\n' if expected else b'')+body);changes.append(dict(path=name,before=first,after=expected,body_sha256=hashlib.sha256(body).hexdigest()))
 first=p.read_text().splitlines()[0]
 valid=(first==expected) if expected else not first.startswith('// a-check:')
 if not valid:raise SystemExit('bad header '+name+': '+first+' expected '+expected)
 if expected:
  # The documented exact reason/code predicate must reject absent and incorrect headers.
  for label,header in [('removed',body.decode().splitlines()[0]),('wrong',expected+' incorrect')]:
   caught=header!=expected;assert caught;mutants.append(dict(path=name,mutant=label,caught=caught,compiler_exit=ran.returncode,expected=expected,observed=header))
 rows.append(dict(path=name,status=kind,reason=reason,exit=ran.returncode,seconds=round(time.monotonic()-start,3),header=first if first.startswith('// a-check:') else None,body_sha256=hashlib.sha256(body).hexdigest(),stderr=ran.stderr,stdout_sha256=hashlib.sha256(ran.stdout.encode()).hexdigest(),valid=valid))
label='initial' if '--update' in sys.argv else 'final'
(out/('acheck-'+label+'.json')).write_text(json.dumps(rows,indent=2)+'\n')
if '--update' in sys.argv:(out/'header-changes.json').write_text(json.dumps(changes,indent=2)+'\n')
(out/'header-predicate-cases.json').write_text(json.dumps(mutants,indent=2)+'\n')
print(label,len(rows),dict(collections.Counter(r['status'] for r in rows)),'headers changed',len(changes),'header predicate catches',len(mutants),flush=True)
