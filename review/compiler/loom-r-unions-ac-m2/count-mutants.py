from pathlib import Path
import subprocess,json,time,difflib
out=Path(__file__).resolve().parent;path=Path('internal/oracle/counts.md');original=path.read_bytes();key='internal/oracle/testdata/scout_class_boolean_fields.a';text=original.decode();line=next(s for s in text.splitlines(True) if s.startswith('| '+key+' | '));fields=line.split(' | ');fields[1]=str(int(fields[1])+1);wrong=' | '.join(fields)
rows=json.loads((out/'mutant-results.json').read_text()) if (out/'mutant-results.json').exists() else []
try:
 for name,changed,pattern in [('removed-count-row',text.replace(line,'',1),'^TestCountsAreRecordedTable$'),('wrong-count-row',text.replace(line,wrong,1),'^TestCountsAreRecorded_[0-9]+$/'+key+'$')]:
  path.write_text(changed);(out/(name+'.patch')).write_text(''.join(difflib.unified_diff(text.splitlines(True),changed.splitlines(True),fromfile='a/internal/oracle/counts.md',tofile='b/internal/oracle/counts.md')))
  cmd=['go','test','./internal/oracle','-run',pattern,'-count=1','-timeout','90s','-parallel','4','-p','1','-json'];start=time.monotonic()
  with (out/('mutant-'+name+'.log')).open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,timeout=180)
  events=[];output=(out/('mutant-'+name+'.log')).read_text()
  for s in output.splitlines():
   try:e=json.loads(s)
   except ValueError:continue
   if e.get('Action')=='fail' and 'Test' in e:events.append(e['Test'])
  rows.append(dict(name=name,exit=r.returncode,command=cmd,seconds=round(time.monotonic()-start,3),catchers=events));(out/'mutant-results.json').write_text(json.dumps(rows,indent=2));print(rows[-1],flush=True)
  assert r.returncode==1 and events and '[build failed]' not in output and 'panic: test timed out' not in output
  if name=='wrong-count-row':assert any(key in e for e in events)
  path.write_bytes(original)
finally:path.write_bytes(original)
assert path.read_bytes()==original
