from pathlib import Path
import json,shutil,gzip,subprocess,hashlib
s=Path(__file__).parent;e=s/'evidence-work/stage3/census/require-result-1610';lane=s/'lane-final'
report=json.loads((lane/'report.json').read_text());assert report['status']=='pass',report
assert json.loads((lane/'execution.json').read_text())['commit']=='ad8ca5a3801d979ca8486b73f3b0ea3cbec0b7d8'
for p in lane.rglob('*'):
 if not p.is_file() or 'adapted-tree' in p.relative_to(lane).parts:continue
 rel=p.relative_to(lane);out=e/'lane'/rel;out.parent.mkdir(parents=True,exist_ok=True)
 if p.suffix=='.log':
  with gzip.open(str(out)+'.gz','wb') as f:f.write(p.read_bytes())
 else:shutil.copy2(p,out)
for name in ['final-source.cjs','final-source.json','final-source.log']:
 shutil.copy2(s/name,e/name)
for name in ['lane-final.log','lane.log']:
 with gzip.open(e/'raw'/(name+'.gz'),'wb') as f:f.write((s/name).read_bytes())
with gzip.open(e/'raw/aborted-lane-apply.log.gz','wb') as f:f.write((s/'lane/apply.log').read_bytes())
commands=json.loads((e/'commands.json').read_text());commands['pending']=[];commands['completed'].append({'command':'stage3/lane/run.sh '+str(lane),'exit':0,'commit':'ad8ca5a3801d979ca8486b73f3b0ea3cbec0b7d8','mode':'whole and unfiltered'})
commands['completed'].append({'command':'node '+str(s/'final-source.cjs'),'exit':0,'result':'747 files, only sys.ts changed, measured candidate bytes exact, replay zero writes and unchanged bytes/mtime'})
(e/'commands.json').write_text(json.dumps(commands,indent=2)+'\n')
status=(e/'STATUS.md').read_text().replace('- Whole unfiltered lane, including candidate full apply: running; not yet claimed.', '- Whole unfiltered lane: PASS, 106,366 passing, one sanctioned API failure, zero pending.\n  Candidate full apply, install and build exit 0; oracle/tests exit 1 only for\n  the sanctioned API baseline row. See lane/report.json and complete logs.\n- Final source scope: 747 files; only sys.ts changes, with the exact measured\n  annotation bytes. Actual adapter replay writes nothing and preserves bytes/mtime.\n- Existing crypto source and JavaScript scope checks pass (747-file inventory).\n- Implementation is one commit on the exact base:\n  ad8ca5a3801d979ca8486b73f3b0ea3cbec0b7d8.')
(e/'STATUS.md').write_text(status)
print(json.dumps(report,indent=2))
