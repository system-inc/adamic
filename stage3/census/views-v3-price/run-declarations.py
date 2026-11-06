#!/usr/bin/env python3
"""Sequential full-mode declaration pieces, bounded and immediately published."""
from collections import defaultdict
import datetime,hashlib,io,json,os,pathlib,re,signal,subprocess,sys,tarfile,time
from declaration_common import P,PLAN,load_plan,validate,split_piece,audit_partition,normalized
D=P/'declaration-plan';plan=load_plan();plan_hash=hashlib.sha256(PLAN.read_bytes()).hexdigest()
index_path=D/'RESULTS.json'
state=json.loads(index_path.read_text()) if index_path.exists() else dict(schema=1,plan_sha256=plan_hash,results={},refinements={})
assert state['plan_sha256']==plan_hash
source_manifest=json.loads((P/'source-manifest.json').read_text())
budget_path=os.getenv('PRICE_BOX_BUDGET')
budget=json.loads(pathlib.Path(budget_path).read_text()) if budget_path else {'stop_before_utc':plan['stop_before_utc'],'box':os.uname().nodename}
assert budget['box']==os.uname().nodename,'box budget belongs to a different box'
stop_at=datetime.datetime.fromisoformat(budget['stop_before_utc']).timestamp()
state['box_budget']=budget
limit=int(sys.argv[1]) if len(sys.argv)>1 else 999

def dump():index_path.write_text(json.dumps(state,indent=2)+'\n')
def leaves(piece):
 if piece['id'] in state['refinements']:
  return [x for c in state['refinements'][piece['id']] for x in leaves(c)]
 return [piece]
def active():
 rows=[x for p in plan['pieces'] for x in leaves(p)];audit_partition(plan,rows);return rows

def pack(stem,scratch):
 files={x.name:x.read_bytes() for x in scratch.iterdir() if x.is_file()}
 files['SHA256SUMS']=''.join(f'{hashlib.sha256(data).hexdigest()}  {name}\n' for name,data in sorted(files.items())).encode()
 out=io.BytesIO()
 with tarfile.open(fileobj=out,mode='w:gz') as t:
  for name,data in sorted(files.items()):
   info=tarfile.TarInfo(name);info.size=len(data);info.mtime=0;t.addfile(info,io.BytesIO(data))
 # Rolling bundles keep the final repository well under 200 files.
 number=len(state['results'])//12;bundle=P/f'declaration-data-{number:03d}.tar.gz'
 members={}
 if bundle.exists():
  with tarfile.open(bundle) as t:members={i.name:t.extractfile(i).read() for i in t if i.isfile() and i.name!='SHA256SUMS'}
 payload=out.getvalue();members[stem+'.tar.gz']=payload
 members['SHA256SUMS']=''.join(f'{hashlib.sha256(data).hexdigest()}  {name}\n' for name,data in sorted(members.items())).encode()
 with tarfile.open(bundle,'w:gz') as t:
  for name,data in sorted(members.items()):
   info=tarfile.TarInfo(name);info.size=len(data);info.mtime=0;t.addfile(info,io.BytesIO(data))
 return dict(bundle=bundle.name,member=stem+'.tar.gz',sha256=hashlib.sha256(payload).hexdigest())

def publish(message):
 dump();subprocess.run(['python3','-B',str(P/'update-report.py')],check=True)
 with pathlib.Path('/workspace/scratch/v3-declaration-publication.log').open('a') as log:
  for args in [['git','add',str(P)],['git','diff','--cached','--check'],['git','commit','-m',message],['git','push','origin','HEAD:refs/heads/typescript/views-v3-price']]:
   subprocess.run(args,check=True,stdout=log,stderr=log,timeout=90)

# Old whole-file observations are exact completed seed leaves, never rerun.
for piece in plan['pieces']:
 if piece['seed'] and piece['id'] not in state['results']:
  original=json.loads((P/piece['seed']).read_text());records=original.get('records',[original])
  state['results'][piece['id']]=dict(status='complete',seed=piece['seed'],records=records,expected_units=piece['expected_units'])
dump()

# Score only changes execution order, never the frozen assignment.
with tarfile.open(P/'source-and-build.tar.gz') as t:
 import gzip
 source_rows=list(map(json.loads,gzip.decompress(t.extractfile('v3-source-bound.jsonl.gz').read()).splitlines()))
read_positions=defaultdict(list)
for name in plan['inventory']:
 text=(pathlib.Path('/workspace/v3-main-adapted-complete')/name).read_bytes().decode();lines=text.splitlines(keepends=True);starts=[0]
 for line in lines:starts.append(starts[-1]+len(line.encode()))
 for r in source_rows:
  if r['file']==name:
   prefix=lines[r['line']-1].encode('utf-16-le')[:2*(r['column']-1)].decode('utf-16-le')
   read_positions[name].append(starts[r['line']-1]+len(prefix.encode()))
def score(piece):
 reads=sum(a['start']<=pos<a['end'] for a in piece['top_atoms'] for pos in read_positions[a['file']])
 return (max((a['end']-a['start'] for a in piece['top_atoms']),default=0)>64000,-reads/(len(piece['expected_units'])+12),len(piece['expected_units']),piece['id'])

processed=0
while processed<limit:
 pending=[p for p in active() if not p['seed'] and p['id'] not in state['results']]
 if not pending:break
 piece=min(pending,key=score)
 # Reserve full piece budget plus checkpoint/shutdown time.
 if time.time()+piece['timeout_seconds']+90>stop_at:
  state['stop']=dict(reason='box time reserve',utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),next_piece=piece['id'])
  publish('census: checkpoint views-v3 declaration progress before box deadline');print('BOX TIME STOP',flush=True);break
 scratch=pathlib.Path('/workspace/scratch/v3-declaration-runs')/piece['id'];scratch.mkdir(parents=True,exist_ok=False)
 (scratch/'selection.json').write_text(json.dumps(piece,indent=2)+'\n')
 started=time.monotonic();deadline=started+piece['timeout_seconds'];rows_by_side={};timings={};error=None;interrupted=False
 print('START '+piece['id']+' '+','.join(f['file'] for f in piece['files'])+' units='+str(len(piece['expected_units'])),flush=True)
 try:
  for side in ['main','head']:
   tree=pathlib.Path(f'/workspace/v3-{side}-adapted-complete')
   for selected in piece['files']:assert hashlib.sha256((tree/selected['file']).read_bytes()).hexdigest()==source_manifest[selected['file']]
   env=dict(os.environ,LATENT_FULL='1',LATENT_SPECULATIVE='0',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE='\n'.join(str(tree/f['file']) for f in piece['files']),PRICE_DECLARATION_ROOT=str(tree),PRICE_DECLARATION_PIECE=str(scratch/'selection.json'),LATENT_TIMES=str(scratch/(side+'-times.jsonl')))
   for k in list(env):
    if k.startswith('LATENT_MUTANT_') or k in ['PRICE_INVENTORY','V3_MUTANT_SKIP_NODE','LATENT_PIECE_PLAN','LATENT_PIECE_ID']:env.pop(k)
   output=scratch/(side+'.jsonl');leg_start=time.monotonic()
   with (scratch/(side+'.log')).open('w') as log:
    child=subprocess.Popen([f'/workspace/scratch/v3-{side}-declaration-bin',str(tree/'src/tsc/tsc.ts'),str(output)],env=env,stdout=log,stderr=log,start_new_session=True)
    try:code=child.wait(timeout=max(1,deadline-time.monotonic()))
    except subprocess.TimeoutExpired:
     interrupted=True;os.killpg(child.pid,signal.SIGQUIT)
     try:child.wait(timeout=5)
     except subprocess.TimeoutExpired:os.killpg(child.pid,signal.SIGKILL);child.wait()
     raise
    finally:
     if child.poll() is None:os.killpg(child.pid,signal.SIGKILL);child.wait()
   timings[side]=time.monotonic()-leg_start;assert code==0,(side,code)
   raw=[json.loads(x) for x in output.read_text().splitlines()];rows_by_side[side]=validate(plan,piece,raw,str(tree))
   timers=[json.loads(x) for x in (scratch/(side+'-times.jsonl')).read_text().splitlines()]
   actual=[t['where'].replace(str(tree)+'/','') for t in timers]
   expected=[u['where'] for r in rows_by_side[side][1:] for u in r['units'] if u['status'] in ['attempted','panic']]
   assert sorted(actual)==sorted(expected),'timer/unit inventory differs'
   pair=[]
   for row in rows_by_side[side]:
    row=json.loads(json.dumps(row))
    for unit in row.get('units',[]):
     for field in ['view_elements','view_element_checks','ir_measured']:unit.pop(field)
    pair.append(row)
   (scratch/(side+'-pair.jsonl')).write_text(''.join(json.dumps(r,separators=(',',':'))+'\n' for r in pair))
  with (scratch/'pair.log').open('w') as log:
   subprocess.run(['python3',str(P.parent/'pair/pair.py'),str(scratch/'main-pair.jsonl'),str(scratch/'head-pair.jsonl'),'/workspace/v3-main-adapted-complete','--typescript',str(P.parents[1]/'api/node_modules/typescript'),'--json',str(scratch/'pair.json'),'--report',str(scratch/'pair.md')],stdout=log,stderr=log,check=True,timeout=max(1,deadline-time.monotonic()))
 except Exception as exc:
  error=repr(exc);interrupted=interrupted or isinstance(exc,subprocess.TimeoutExpired)
 result=dict(status='interrupted' if interrupted else 'failed' if error else 'complete',expected_units=piece['expected_units'],top_units=[a['where'] for a in piece['top_atoms']],seconds=time.monotonic()-started,timings=timings,error=error,plan_sha256=plan_hash)
 if not error:
  perfile={f['file']:{'file':f['file'],'sides':{}} for f in piece['files']}
  for side,rows in rows_by_side.items():
   for row in rows[1:]:
    units=row['units'];blocked={f['unit'] for f in row['findings'] if f['phase']=='lowering'};free=[u for u in units if u['status']=='attempted' and u['where'] not in blocked]
    perfile[row['file']]['sides'][side]=dict(view_elements=sum(u['view_elements'] for u in units),view_element_checks=sum(u['view_element_checks'] for u in units),units=len(units),attempted=sum(u['status'] in ['attempted','panic'] for u in units),panicked_units=sum(u['status']=='panic' for u in units),ir_observed=sum(u['ir_measured'] for u in units),finding_free_attempted=len(free),finding_free_view_elements=sum(u['view_elements'] for u in free),finding_free_view_element_checks=sum(u['view_element_checks'] for u in free))
  result['records']=list(perfile.values())
 else:
  if interrupted and len(piece['top_atoms'])>1:state['refinements'][piece['id']]=split_piece(plan,piece)
  elif interrupted:result['atomic_blocked']=True
 (scratch/'result.json').write_text(json.dumps(result,indent=2)+'\n');result['raw']=pack(piece['id'],scratch);state['results'][piece['id']]=result
 active();publish('census: views-v3 declaration '+piece['id']+' '+result['status']);processed+=1
 print(json.dumps(dict(id=piece['id'],status=result['status'],seconds=result['seconds'],records=result.get('records'),error=error)),flush=True)
 if error and not interrupted:raise SystemExit('validation failure retained and pushed')
print('RUNNER EXIT',flush=True)
