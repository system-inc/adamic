from pathlib import Path
import subprocess,json,difflib,time
p=Path('review/test-audit/internal-oracle-wasi');file='internal/oracle/oracle_test.go'
src=subprocess.check_output(['git','show','HEAD:'+file],text=True)
start=src.index('func disagreement(oracle run, native run) string {');end=src.index('\nfunc TestNativeAgreesWithNode',start)
new=src[:start]+'func disagreement(oracle run, native run) string {\n\treturn ""\n}\n'+src[end:]
scratch=Path('/workspace/u072-tmp/witness.go');scratch.write_text(new)
(p/'W01.diff').write_text(''.join(difflib.unified_diff(src.splitlines(True),new.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
overlay=Path('/workspace/u072-tmp/witness.json');overlay.write_text(json.dumps({'Replace':{str(Path(file).resolve()):str(scratch)}}))
cmd=['timeout','120','go','test','-overlay',str(overlay),'-json','-count=1','-timeout','90s','./internal/oracle/','-run','^(TestWASIOracleCatchesMutants|TestWASIRunnerCatchesMutants)$']
start=time.monotonic()
with open(p/'W01.log','w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT)
(p/'witness-status.json').write_text(json.dumps(dict(id='W01',command=cmd,seconds=time.monotonic()-start,status=r.returncode,file=file,line=717),indent=2))
