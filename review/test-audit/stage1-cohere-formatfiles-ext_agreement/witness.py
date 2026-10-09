import json,subprocess,difflib,time
from pathlib import Path
p=Path('review/test-audit/stage1-cohere-formatfiles-ext_agreement');file='stage1/cohere/formatfiles/ext_agreement_test.go';src=subprocess.check_output(['git','show','HEAD:'+file],text=True)
new=src.replace('bytes.Equal(', 'auditAgreement(')+'\n// Audit weakening: every pair is declared equal.\nfunc auditAgreement(left, right []byte) bool { return true }\n'
scratch=Path('/workspace/u090-tmp/witness.go');scratch.write_text(new);overlay=Path('/workspace/u090-tmp/witness.json');overlay.write_text(json.dumps({'Replace':{str(Path(file).resolve()):str(scratch)}}))
(p/'W01.diff').write_text(''.join(difflib.unified_diff(src.splitlines(True),new.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
cmd=['timeout','120','go','test','-overlay',str(overlay),'-json','-count=1','-timeout','90s','./stage1/cohere/formatfiles/','-run','^TestExtAgreesWithGo$'];s=time.monotonic()
with (p/'W01.log').open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT)
with (p/'W01-vet.log').open('w') as f:v=subprocess.run(['go','vet','-overlay',str(overlay),'./stage1/cohere/formatfiles/'],stdout=f,stderr=subprocess.STDOUT)
(p/'witness-status.json').write_text(json.dumps(dict(id='W01',status=r.returncode,wall=time.monotonic()-s,command=cmd,compile_status=v.returncode,origin_line=src[:src.index('if !bytes.Equal(gotLines')].count('\n')+1),indent=2))
