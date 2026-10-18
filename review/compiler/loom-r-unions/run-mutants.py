import concurrent.futures,difflib,json,subprocess,time
from pathlib import Path
root=Path.cwd();e=root/'review/compiler/loom-r-unions';old=root/'review/compiler/loom-unions-3';out=e/'mutants';out.mkdir(exist_ok=True)
cases=json.loads((old/'mutants/cases.json').read_text());prepared=[]
for c in cases:
 original=subprocess.check_output(['git','show','2d9d5504:'+c['file']],text=True);mutated=(old/'mutants'/(c['topic']+'.go.txt')).read_text();current=(root/c['file']).read_text();changed=current
 before_lines=original.splitlines(True);after_lines=mutated.splitlines(True)
 for op,i,j,k,l in reversed(list(difflib.SequenceMatcher(None,before_lines,after_lines,autojunk=False).get_opcodes())):
  if op=='equal':continue
  before=''.join(before_lines[max(0,i-3):min(len(before_lines),j+3)]);after=''.join(before_lines[max(0,i-3):i])+''.join(after_lines[k:l])+''.join(before_lines[j:j+3])
  if not before:
   before=''.join(before_lines[max(0,i-2):i]);after=before+after
  if changed.count(before)!=1:raise RuntimeError((c['topic'],'ambiguous changed seam',changed.count(before)))
  changed=changed.replace(before,after,1)
 p=out/(c['topic']+'.go.txt');p.write_text(changed)
 overlay=out/(c['topic']+'.json');overlay.write_text(json.dumps({'Replace':{str(root/c['file']):str(p)}}))
 (out/(c['topic']+'.patch')).write_text(''.join(difflib.unified_diff(current.splitlines(True),changed.splitlines(True),fromfile='a/'+c['file'],tofile='b/'+c['file'])))
 c={**c,'overlay':str(overlay)};prepared.append(c)
(out/'cases.json').write_text(json.dumps(prepared,indent=2)+'\n')
def run(c):
 cmd=['go','test','-overlay',c['overlay'],*c['packages'],'-run',c['pattern'],'-count=1','-timeout','90s','-parallel','2','-json'];start=time.monotonic();log=out/(c['topic']+'.log')
 with log.open('w') as f:
  try:code=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,timeout=180).returncode
  except subprocess.TimeoutExpired:code=124
 text=log.read_text();events=[]
 for line in text.splitlines():
  try:r=json.loads(line)
  except json.JSONDecodeError:continue
  if r.get('Action')=='fail' and 'Test' in r:events.append(r)
 killed=code==1 and bool(events) and '[build failed]' not in text and 'panic: test timed out' not in text
 row=dict(topic=c['topic'],command=cmd,exit=code,seconds=round(time.monotonic()-start,3),killed=killed,catchers=events);print(c['topic'],code,row['seconds'],killed,flush=True)
 with (out/'results.jsonl').open('a') as f:f.write(json.dumps(row)+'\n')
 return row
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:rows=list(pool.map(run,prepared))
assert len(rows)==22 and all(r['killed'] for r in rows)
