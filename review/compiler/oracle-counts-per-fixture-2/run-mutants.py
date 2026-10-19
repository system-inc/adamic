from pathlib import Path
import subprocess,json,os
root=Path(__file__).resolve().parents[3];out=Path(__file__).resolve().parent;results=[]
def run(name,selector,needle,overlay=None):
 cmd=['go','test','./internal/oracle','-run',selector,'-count=1','-parallel=4','-timeout','90s','-json']
 if overlay:cmd+=['-overlay',str(overlay)]
 path=out/(name+'.jsonl')
 with path.open('w') as log:r=subprocess.run(cmd,cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=120)
 events=[]
 for line in path.read_text().splitlines():
  try:v=json.loads(line)
  except:continue
  events.append(v)
 text=''.join(v.get('Output','') for v in events);failed=[v.get('Test') for v in events if v.get('Test') and v.get('Action')=='fail'];assert r.returncode==1 and needle in text and failed,(name,r.returncode,text[-3000:]);assert not any(v.get('Action')=='build-fail' for v in events)
 results.append(dict(name=name,command=cmd,exit=r.returncode,caught_by=failed,required_diagnostic=needle));(out/'mutant-results.json').write_text(json.dumps(results,indent=2)+'\n');print(name,'CAUGHT',flush=True)
file=root/'internal/oracle/testdata/counts/runtime/dedication/dedication.a.counts';saved=file.read_bytes()
try:
 file.unlink();run('missing-counts-mutant','^TestCountsFilesCensus$','dedication.a.counts')
finally:file.write_bytes(saved)
orphan=root/'internal/oracle/testdata/counts/runtime/internal/oracle/testdata/ra6ah5s_orphan.a.counts';assert not orphan.exists()
try:
 orphan.write_text('| internal/oracle/testdata/ra6ah5s_orphan.a | 0 | 0 | 0 | 0 | 0 | 0 |\n');run('orphan-counts-mutant','^TestCountsFilesCensus$','orphan counts file')
finally:orphan.unlink()
try:
 cells=saved.decode().split('|');cells[2]=' '+str(int(cells[2])+1)+' ';file.write_text('|'.join(cells));run('changed-count-mutant','^TestCountsAreRecorded_[0-9]+$/^dedication/dedication[.]a$','the counts changed and testdata/counts/runtime/dedication/dedication.a.counts')
finally:file.write_bytes(saved)
source=root/'internal/oracle/count_files_test.go';s=source.read_text();anchor='files[filepath.Join(root, kind, filepath.FromSlash(key)+".counts")] = data';assert s.count(anchor)==1;mutant=out/'crlf-writer.go.txt';mutant.write_text(s.replace(anchor,anchor.replace('= data','= bytes.ReplaceAll(data, []byte("\\n"), []byte("\\r\\n"))')));overlay=out/'crlf-writer-overlay.json';overlay.write_text(json.dumps({'Replace':{str(source):str(mutant)}})+'\n');run('crlf-writer-mutant','^TestPredicateCountsRegenerateDeterministically$','regenerated counts differ',overlay)
