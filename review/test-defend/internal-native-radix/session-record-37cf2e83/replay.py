import pathlib, subprocess, json, time
root=pathlib.Path('/workspace/adamic'); out=root/'review/test-defend/internal-native-radix/session-record-37cf2e83'
plans=[dict(mutant='D01',file='internal/native/runtime/map.c',before='\t\tfor (size_t index = 0; index < string->length; index++) {\n\t\t\thash = (hash ^ (unsigned char)string->bytes[index]) * 1099511628211ull;\n\t\t}',after='\t\tfor (size_t repeat = 0; repeat < 8; repeat++) {\n\t\tfor (size_t index = 0; index < string->length; index++) {\n\t\t\thash = (hash ^ (unsigned char)string->bytes[index]) * 1099511628211ull;\n\t\t}\n\t\t}',change='Repeat string hashing eight times, increasing work while retaining consistent equal-key hashes; cost-rule repeat-loop attempt'),dict(mutant='D02',file='internal/native/runtime/record.c',before='while (next < keys->length)',after='while (next + 1 < keys->length)',change='Off-by-one iterator exhaustion bound drops the last snapshot key'),dict(mutant='D03',file='internal/native/runtime/map.c',before='\tmap->count--;',after='',change='Drop deletion count decrement')]
for p in plans:
 s=(root/p['file']).read_text(); assert s.count(p['before'])==1,p
 p['line']=s[:s.index(p['before'])].count('\n')+1
(out/'plan.json').write_text(json.dumps(plans,indent=2))
for name in ['REPORT.md','rows.json','friction-and-limits.md','mutant-table.md']:
 (out/('audit-'+name)).write_bytes(subprocess.check_output(['git','show','origin/test-audit/internal-native-radix:review/test-audit/internal-native-radix/'+name],cwd=root))
selector='^(TestRegExp.*|TestRecord.*|TestRuntimeStringEquality|TestMapHash.*|TestLibraryMapSetIteratorResources)$'
results=[]
for p in plans:
 path=root/p['file']; original=path.read_text()
 try:
  path.write_text(original.replace(p['before'],p['after']))
  (out/(p['mutant']+'.diff')).write_bytes(subprocess.check_output(['git','diff','--',p['file']],cwd=root))
  with (out/(p['mutant']+'-compile.log')).open('w') as log:
   code=subprocess.run(['clang','-std=c11','-Wall','-Wextra','-Werror','-Wno-unused-function','-DADAMIC_COUNT','-fsanitize=address,undefined','-Iinternal/native/runtime','-fsyntax-only',p['file']],cwd=root,stdout=log,stderr=subprocess.STDOUT).returncode
  assert code==0,(p,code)
  cmd=f"source /workspace/adamic-tools/env.sh\nADAMIC_RECORD_BENCH=1 ADAMIC_BUILD_CACHE_DIR=/tmp/record-defense/cache/{p['mutant']} timeout 330 go test -json -count=1 -timeout 300s ./internal/native/ -run '{selector}'"
  (out/(p['mutant']+'-command.txt')).write_text(cmd)
  start=time.monotonic()
  with (out/(p['mutant']+'.log')).open('w') as log:
   code=subprocess.run(['bash','-c',cmd],cwd=root,stdout=log,stderr=subprocess.STDOUT).returncode
  results.append(dict(mutant=p['mutant'],exit=code,wall_seconds=time.monotonic()-start))
  (out/'runs.json').write_text(json.dumps(results,indent=2))
 finally:path.write_text(original)
