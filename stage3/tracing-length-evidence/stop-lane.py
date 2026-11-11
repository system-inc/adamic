from pathlib import Path
import os,signal,json
processes={}
for p in Path('/proc').iterdir():
 if not p.name.isdigit():continue
 try:
  args=(p/'cmdline').read_bytes().split(b'\0');parent=int(next(s for s in (p/'status').read_text().splitlines() if s.startswith('PPid:')).split()[1]);processes[int(p.name)]=(parent,args)
 except (OSError,StopIteration,ValueError):continue
matched=[pid for pid,(parent,args) in processes.items() if args[:2]==[b'go',b'test'] and b'/workspace/scratch/tracing-length-evidence/lane-overlay.json' in args];assert len(matched)==1,matched
selected={matched[0]}
while True:
 more={pid for pid,(parent,args) in processes.items() if parent in selected}
 if more<=selected:break
 selected|=more
for pid in sorted(selected-{matched[0]},reverse=True):
 try:os.kill(pid,signal.SIGTERM)
 except ProcessLookupError:pass
os.kill(matched[0],signal.SIGTERM)
Path('/workspace/scratch/tracing-length-evidence/lane-unguarded-cancel.json').write_text(json.dumps({'reason':'Unguarded splice loses reachable RangeError under reentrant args getter; replace implementation and repeat full lane','goPid':matched[0],'terminatedPids':sorted(selected)},indent=2)+'\n')
