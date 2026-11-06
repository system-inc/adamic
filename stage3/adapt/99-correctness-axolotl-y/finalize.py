from pathlib import Path
import gzip,json,subprocess,os
repo=Path('/workspace/adamic');out=repo/'stage3/adapt/99-correctness-axolotl-y'
assert Path('/workspace/scratch/y-groups.log').read_text().endswith('group 4 committed and pushed\n'), 'census is still running'
with (out/'census-summary.log').open('wb') as log:subprocess.run(['python3',str(out/'summarize.py')],stdout=log,stderr=log,check=True)
result=json.loads((out/'census-results.json').read_text());assert result['movedPastCore966']==73
for i in range(1,5):
 n=19 if i==4 else 20
 rows=[json.loads(line) for line in (out/f'group-{i}.jsonl').read_text().splitlines()];assert len(rows)==2*n
 for a,b in zip(rows[::2],rows[1::2]):assert {k:v for k,v in a.items() if k!='seconds'}=={k:v for k,v in b.items() if k!='seconds'}
 p=out/f'group-{i}.log'
 data=p.read_bytes() if p.exists() else gzip.decompress((out/f'group-{i}.log.gz').read_bytes())
 assert len(data.splitlines())==2*n,(i,len(data.splitlines()))
proof=out/'PROOF.md';proof.write_text(proof.read_text().replace('In progress; checkpoint groups retained','Complete; groups of 20, 20, 20, 19 each committed and pushed'))
with gzip.open(out/'build-main-before-setup.log.gz','wb') as stream:stream.write(Path('/workspace/scratch/y-build.log').read_bytes())
with proof.open('a') as stream:stream.write('\nAn initial compiler build attempted before the nested checkout was ready failed\non its missing go.mod. The unchanged build passed after that checkout completed.\nThe early failure is retained in build-main-before-setup.log.gz. Group 1 logging\nwas archived early; its complete 40-line log was recovered from the open file\ndescriptor before it closed. All group logs and raw row pairs were validated.\n')
for p in out.glob('*.log'):
 with gzip.open(p.with_suffix('.log.gz'),'wb') as stream:stream.write(p.read_bytes())
 p.unlink()
with gzip.open(out/'groups-driver.log.gz','wb') as stream:stream.write(Path('/workspace/scratch/y-groups.log').read_bytes())
# Enforce the requested artifact naming and punctuation conventions.
for p in out.rglob('*'):
 if not p.is_file():continue
 assert p.suffix!='.go',p
 if p.suffix!='.gz':
  text=p.read_text();assert '\u2013' not in text and '\u2014' not in text,p
# The current finalizer is the only live task worker permitted at this stage.
ps=subprocess.check_output(['ps','-eo','pid,ppid,state,comm'],text=True)
live=[]
for line in ps.splitlines()[1:]:
 pid,ppid,state,name=line.split(maxsplit=3)
 if name in ['node','adamic','tool','go','python3'] and state!='Z' and int(pid)!=os.getpid():live.append(line)
assert not live,live
(out/'process-audit.json').write_text(json.dumps({'liveTaskWorkers':live,'allStartedTaskWorkersExitedOrReaped':True},indent=2)+'\n')
history=[]
for line in subprocess.check_output(['git','log','--format=%H %s','--grep=99y: checkpoint compiler census group'],cwd=repo,text=True).splitlines():
 sha,subject=line.split(' ',1);history.append({'sha':sha,'subject':subject})
assert len(history)==4,history
(out/'checkpoint-history.json').write_text(json.dumps(list(reversed(history)),indent=2)+'\n')
print('Final evidence validated; no live task workers')
