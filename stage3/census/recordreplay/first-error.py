"""Paired complete-entry first-error census, disjoint file pieces under one hour."""
from pathlib import Path
import subprocess as s,json,os,time,gzip,shutil,hashlib
D=Path('/workspace/scratch/recordreplay');R=D/'base';O=R/'stage3/census/recordreplay';T=O/'first-error';T.mkdir(exist_ok=True)
groups=json.loads((O/'entry-groups.json').read_text())['groups'];inventory=json.loads((O/'entry-sources.json').read_text());files=[f for g in groups for f in g['files']]
assert len(files)==len(set(files))==81 and set(files)=={r['file'] for r in inventory}
summary={'schema_version':1,'coverage':{'implementation_sources':81,'groups_disjoint':True,'all_sources_covered':True},'pieces':[]}
def checkpoint(message):
 s.run(['git','add','stage3/census/recordreplay'],cwd=R,check=True)
 if s.run(['git','diff','--cached','--quiet'],cwd=R).returncode:
  s.run(['git','commit','-m',message],cwd=R,check=True)
  s.run(['git','push','origin','HEAD:refs/heads/typescript/recordreplay-census'],cwd=R,check=True)
try:
 for pair,tree in [('A',D/'control'),('B',D/'candidate2')]:
  for g in groups:
   start=time.monotonic();name=pair+f'-group-{g["id"]}';dest=T/name;dest.mkdir(exist_ok=True);piece={'pair':pair,'group':g['id'],'files':g['files'],'deadline_seconds':3600,'complete':False};summary['pieces'].append(piece)
   print('START',name,flush=True)
   for side in ['base','head']:
    out=dest/(side+'.jsonl');env=dict(os.environ,LATENT_FULL='0',LATENT_SPECULATIVE='0',LATENT_ASSERT_NO_OUTPUT='1',CENSUS_MODULE_FORMAT='CommonJSBundle',LATENT_ONLY_FILE='\n'.join(str(tree/f) for f in g['files']),RUN_LIMIT=str(max(1,min(1700,int(3500-(time.monotonic()-start))))))
    label=name+'-first-'+side
    result=s.run(['python3',str(D/'run.py'),label,str(D/('first-entry-'+side+'-bin') if side=='base' else D/'first-entry-head-bin'),str(tree/'src/tsc/tsc.ts'),str(out)],cwd=D/side if side=='base' else D/'head',env=env)
    shutil.copyfile(D/(label+'.json'),dest/(side+'-resource.json'));(dest/(side+'.log.gz')).write_bytes(gzip.compress((D/(label+'.log')).read_bytes(),mtime=0))
    if result.returncode:raise RuntimeError(f'{name} {side} stopped with {result.returncode}; see retained status and partial stream')
    rows=[json.loads(l) for l in out.read_text().splitlines()];assert rows[0]['status']=='measurement' and rows[-1]['status']=='relation_report';assert not rows[0]['checker_rejected']
    observed={r['file'].split(str(tree)+'/')[-1] for r in rows[1:-1]};assert observed==set(g['files']),(observed,set(g['files']))
    assert set(p.split(str(tree)+'/')[-1] for p in rows[0]['sources'])==set(files)
   s.run(['python3',str(R/'stage3/census/pair/pair.py'),str(dest/'base.jsonl'),str(dest/'head.jsonl'),str(tree),'--typescript',str(D/'head/stage3/api/node_modules/typescript'),'--json',str(dest/'delta.json'),'--report',str(dest/'delta.md')],cwd=R,check=True)
   delta=json.loads((dest/'delta.json').read_text());v=delta['distinct_reads'];piece.update(complete=True,seconds=time.monotonic()-start,unit_first_stop_changes=len(v['unit_first_stop_delta']),regressions=len(v['regressions']),new_process_findings=[r for r in v['new_findings'] if 'process' in (r['finding'].get('reason','')+' '+r['finding'].get('text','')).lower()])
   assert piece['seconds']<3600,piece['seconds']
   (dest/'piece.json').write_text(json.dumps(piece,indent=2)+'\n')
   for side in ['base','head']:
    raw=dest/(side+'.jsonl');(dest/(side+'.jsonl.gz')).write_bytes(gzip.compress(raw.read_bytes(),mtime=0));raw.unlink()
   (T/'summary.json').write_text(json.dumps(summary,indent=2)+'\n');checkpoint(f'stage3: recordreplay {name} paired first-error piece complete');print('CHECKPOINT',name,piece['seconds'],piece['unit_first_stop_changes'],len(piece['new_process_findings']),flush=True)
 summary['complete']=True;(T/'summary.json').write_text(json.dumps(summary,indent=2)+'\n');checkpoint('stage3: complete both recordreplay all-unit first-error pairs');print('first-error pairs complete',flush=True)
except Exception as err:
 summary['stopping_reason']=str(err);(T/'summary.json').write_text(json.dumps(summary,indent=2)+'\n');checkpoint('stage3: checkpoint the recordreplay first-error stopping reason');raise
finally:
 # The scheduler was paused between ordinary roots while the disjoint pieces ran.
 os.kill(13015,18)
