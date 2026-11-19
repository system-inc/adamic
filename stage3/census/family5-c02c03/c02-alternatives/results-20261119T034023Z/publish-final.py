import sys,subprocess,json,shutil,gzip,datetime
from pathlib import Path
W=Path('/tmp/family5-c02-alternatives');R=Path('/tmp/family5-land')
assert (W/'alternatives-finished.json').exists() and (W/'node-controls-finished.json').exists(),'measurements still running'
subprocess.run([sys.executable,W/'draft-report.py'],check=True)
subprocess.run(['git','switch','typescript/family5-c02c03-evidence'],cwd=R,check=True)
subprocess.run(['git','pull','--ff-only','origin','typescript/family5-c02c03-evidence'],cwd=R,check=True)
root=R/'stage3/census/family5-c02c03';study=root/'c02-alternatives';dest=study/('results-'+datetime.datetime.now(datetime.timezone.utc).strftime('%Y%m%dT%H%M%SZ'));dest.mkdir(parents=True)
exclude={'tools','entry','entry-pair-base','entry-pair-c02','__pycache__'}
for p in W.rglob('*'):
 rel=p.relative_to(W)
 if any(x in exclude for x in rel.parts) or p.name=='workflow.lock' or not p.is_file():continue
 q=dest/rel;q.parent.mkdir(parents=True,exist_ok=True)
 if p.suffix in ['.log','.jsonl','.diff']:
  with p.open('rb') as a,gzip.GzipFile(filename=str(q)+'.gz',mode='wb',mtime=0) as b:shutil.copyfileobj(a,b)
 else:shutil.copyfile(p,q)
assert not list(dest.rglob('*.tar.gz'))
text=(W/'REPORT.md').read_text();(study/'REPORT.md').write_text('Complete proof receipts, patches and raw streams: ['+dest.name+']('+dest.name+'/REPORT.md).\n\n'+text.replace('../whole-entry/','whole-entry/'))
(root/'before-alternatives-report.md').write_text((root/'REPORT.md').read_text())
(root/'REPORT.md').write_text('# Family 5: C02 reach and alternatives\n\n[Current whole-entry and alternatives report](c02-alternatives/REPORT.md).\n\n[Every whole-entry callable-field site and difference](c02-alternatives/whole-entry/REPORT.md).\n\nNo landing branch was pushed. Earlier accepted lane and serial stopping evidence remain in their recorded directories and Git history.\n')
subprocess.run(['git','add',str(root)],cwd=R,check=True)
subprocess.run(['git','commit','-m','stage3: record C02 scratch alternatives and full census evidence'],cwd=R,check=True)
subprocess.run(['git','push','origin','HEAD:refs/heads/typescript/family5-c02c03-evidence'],cwd=R,check=True)
sha=subprocess.check_output(['git','rev-parse','HEAD'],cwd=R,text=True).strip();(W/'final-publication.json').write_text(json.dumps({'sha':sha,'report':str((study/'REPORT.md').relative_to(R))},indent=2)+'\n');print('FINAL EVIDENCE '+sha,flush=True)
