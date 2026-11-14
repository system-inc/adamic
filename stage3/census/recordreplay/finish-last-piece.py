"""Finish only the interrupted B group; never resume ordinary tables."""
from pathlib import Path
import os,json,time,subprocess as s,shutil,gzip
D=Path('/workspace/scratch/recordreplay');R=D/'base';O=R/'stage3/census/recordreplay';dest=O/'first-error/B-group-5';tree=D/'candidate2'
g=next(g for g in json.loads((O/'entry-groups.json').read_text())['groups'] if g['id']==5)
history=dest/'interrupted';history.mkdir(exist_ok=True)
for name in ['base-resource.json','base.jsonl.gz','base.log.gz','head.partial.jsonl.gz']:
 p=dest/name
 if p.exists():p.rename(history/name)
start=time.monotonic();resources={}
for side in ['base','head']:
 label='B-group-5-resumed-first-'+side;out=dest/(side+'.jsonl')
 env=dict(os.environ,LATENT_FULL='0',LATENT_SPECULATIVE='0',LATENT_ASSERT_NO_OUTPUT='1',CENSUS_MODULE_FORMAT='CommonJSBundle',LATENT_ONLY_FILE='\n'.join(str(tree/f) for f in g['files']),RUN_LIMIT=str(min(1700,int(3500-(time.monotonic()-start)))))
 s.run(['python3',str(D/'run.py'),label,str(D/('first-entry-'+side+'-bin')),str(tree/'src/tsc/tsc.ts'),str(out)],cwd=D/side,env=env,check=True)
 resources[side]=json.loads((D/(label+'.json')).read_text());shutil.copyfile(D/(label+'.json'),dest/(side+'-resource.json'));(dest/(side+'.log.gz')).write_bytes(gzip.compress((D/(label+'.log')).read_bytes(),mtime=0))
 rows=[json.loads(l) for l in out.read_text().splitlines()];assert rows[0]['status']=='measurement' and not rows[0]['checker_rejected'];assert rows[-1]['status']=='relation_report';assert {r['file'].split(str(tree)+'/')[-1] for r in rows[1:-1]}==set(g['files'])
s.run(['python3',str(R/'stage3/census/pair/pair.py'),str(dest/'base.jsonl'),str(dest/'head.jsonl'),str(tree),'--typescript',str(D/'head/stage3/api/node_modules/typescript'),'--json',str(dest/'delta.json'),'--report',str(dest/'delta.md')],cwd=R,check=True)
v=json.loads((dest/'delta.json').read_text())['distinct_reads'];piece={'pair':'B','group':5,'files':g['files'],'deadline_seconds':3600,'complete':True,'seconds':time.monotonic()-start,'unit_first_stop_changes':len(v['unit_first_stop_delta']),'regressions':len(v['regressions']),'new_process_findings':[r for r in v['new_findings'] if 'process' in (r['finding'].get('reason','')+' '+r['finding'].get('text','')).lower()]};assert piece['seconds']<3600
(dest/'piece.json').write_text(json.dumps(piece,indent=2)+'\n')
for side in ['base','head']:
 p=dest/(side+'.jsonl');(dest/(side+'.jsonl.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0));p.unlink()
p=O/'first-error/summary.json';summary=json.loads(p.read_text());assert len(summary['pieces'])==11;summary['pieces'].append(piece);summary['complete']=True;summary.pop('stopping_reason',None);p.write_text(json.dumps(summary,indent=2)+'\n')
s.run(['python3',str(O/'native-source.py')],cwd=R,check=True)
s.run(['git','add','stage3/census/recordreplay'],cwd=R,check=True);s.run(['git','commit','-m','stage3: finish the final paired recordreplay all-unit piece'],cwd=R,check=True);s.run(['git','push','origin','HEAD:refs/heads/typescript/recordreplay-census'],cwd=R,check=True)
print(json.dumps(piece),flush=True)
