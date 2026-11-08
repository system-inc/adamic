import json,os,signal,time
from pathlib import Path
scratch=Path('/workspace/scratch/c3b22184');out=Path(__file__).resolve().parent
binaries={str(scratch/'census'),str(scratch/'entry-census')}
def active():
 result=[]
 for p in Path('/proc').iterdir():
  if not p.name.isdigit():continue
  try:argv=(p/'cmdline').read_bytes().split(b'\0')
  except (FileNotFoundError,PermissionError,ProcessLookupError):continue
  if argv and argv[0].decode(errors='replace') in binaries:result.append(dict(pid=int(p.name),argv=[a.decode(errors='replace') for a in argv if a]))
 return result
before=active()
for row in before:
 try:os.killpg(os.getpgid(row['pid']),signal.SIGTERM)
 except ProcessLookupError:pass
if before:time.sleep(1)
for row in active():
 try:os.kill(row['pid'],signal.SIGKILL)
 except ProcessLookupError:pass
remaining=active()
statuses={p.name:json.loads(p.read_text()) for p in scratch.glob('*-status.json') if p.name.startswith(('root-','entry-census-','entry-latent-'))}
result=dict(processes_found=before,remaining=remaining,statuses=statuses,all_supervised_groups_cleaned=all(s.get('process_group_cleaned') for s in statuses.values()))
(out/'cleanup.json').write_text(json.dumps(result,indent=2)+'\n')
assert not remaining and result['all_supervised_groups_cleaned']
print('remaining_census_processes=0')
