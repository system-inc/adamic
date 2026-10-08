from pathlib import Path
import subprocess
p=Path('internal/native/runtime/view_dictionaries.c');original=p.read_bytes()
try:
 s=original.decode();a='else if (actual == 13) value.kind = adamic_view_union_undefined;';assert s.count(a)==1
 p.write_text(s.replace(a,'else if (actual == 13) value.kind = adamic_view_union_unknown;'))
 with open('/tmp/views-integration-untagged-cd32-dictionary-undefined-mutant.log','wb') as out:
  r=subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewDictionaryCompilerPaths$/compiler-paths-missing$','-count=1','-v'],stdout=out,stderr=subprocess.STDOUT)
 log=Path('/tmp/views-integration-untagged-cd32-dictionary-undefined-mutant.log').read_text()
 assert r.returncode!=0 and '--- FAIL: TestCheckedViewDictionaryCompilerPaths' in log and 'exit codes differ' in log and '[build failed]' not in log and 'clang failed' not in log,log
 print('dictionary semantic-undefined decoder omission caught by valid Node control')
finally:p.write_bytes(original)
