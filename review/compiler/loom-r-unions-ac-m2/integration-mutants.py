from pathlib import Path
import subprocess,json,time
out=Path(__file__).resolve().parent;root=Path.cwd();(out/'mutants').mkdir(exist_ok=True);rows=[]
cases=[('drop-map-adapter-exemption',' || l.nullableObjectAdaptedMapCallback(node, contextual)','','^TestCallbackWideningNullableObjectMapRelease$'),('drop-nested-storage-match','if !l.nullableObjectCallbackStorageMatches(','if false && !l.nullableObjectCallbackStorageMatches(','^TestNullableObjectMapKeepsArrayStorageViewNotYet$')]
file=root/'internal/lower/nullable_objects.go';original=file.read_text()
for name,old,new,pattern in cases:
 assert old in original;replacement=out/'mutants'/(name+'.go.txt');replacement.write_text(original.replace(old,new,1));overlay=out/'mutants'/(name+'.json');overlay.write_text(json.dumps({'Replace':{str(file):str(replacement)}}))
 cmd=['go','test','./internal/oracle','-overlay',str(overlay),'-run',pattern,'-count=1','-timeout','90s','-parallel','4','-p','1','-json'];start=time.monotonic()
 with (out/('mutant-'+name+'.log')).open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,timeout=180)
 events=[];text=(out/('mutant-'+name+'.log')).read_text()
 for s in text.splitlines():
  try:e=json.loads(s)
  except ValueError:continue
  if e.get('Action')=='fail' and 'Test' in e:events.append(e['Test'])
 rows.append(dict(name=name,command=cmd,exit=r.returncode,seconds=round(time.monotonic()-start,3),catchers=events));(out/'mutant-results.json').write_text(json.dumps(rows,indent=2));print(rows[-1],flush=True)
 assert r.returncode==1 and events and '[build failed]' not in text and 'panic: test timed out' not in text
assert file.read_text()==original
