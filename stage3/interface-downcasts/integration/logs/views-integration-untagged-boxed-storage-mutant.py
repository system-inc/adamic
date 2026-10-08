from pathlib import Path
import subprocess
p=Path('internal/native/emit_objects.go');original=p.read_bytes()
try:
 s=original.decode();a='if field.Uninitialized && field.Value.Type() == ir.Union {';assert s.count(a)==1
 p.write_text(s.replace(a,'if false && field.Uninitialized && field.Value.Type() == ir.Union {'))
 with open('/tmp/views-integration-untagged-cd32-boxed-storage-mutant.log','wb') as out:
  r=subprocess.run(['go','test','./internal/oracle','-run','^TestDefaultTaggedSourceViews$/default-boxed-write$','-count=1','-v'],stdout=out,stderr=subprocess.STDOUT)
 log=Path('/tmp/views-integration-untagged-cd32-boxed-storage-mutant.log').read_text()
 assert r.returncode!=0 and '--- FAIL: TestDefaultTaggedSourceViews' in log and 'exit codes differ' in log and '[build failed]' not in log and 'clang failed' not in log,log
 print('reserved boxed storage omission caught by Node initialization control')
finally:p.write_bytes(original)
