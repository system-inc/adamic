from pathlib import Path
import subprocess,os,json,time,shlex
p=Path('review/test-defend/internal-native-split_units/followup')
rows=['TestStringsMatchJavaScript','TestStringIndexMatchesNode','TestStringIndexCacheStatesMatchNode','TestStringViewAfterAppendMatchesNode','TestStringBuildingMatchesNode','TestRuntimeStringViews']
regex='^('+'|'.join(rows)+')$';cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run',regex]
menu=[('D2','internal/native/runtime/string_repeat_impl.h','return repeat_unchecked(string, count);','return repeat_unchecked(string, count - 1);','internal/native/runtime/string.c'),('D3','internal/native/runtime/string_append.c','if (string->heap.references == 1 && !itself','if (string->heap.references == 2 && !itself','internal/native/runtime/string_append.c'),('D4','internal/native/runtime/string_index.c','bytes[0] & 0x07','bytes[0] & 0x03','internal/native/runtime/string_index.c')]
p.joinpath('menu.json').write_text(json.dumps(menu,indent=2)+'\n')
for mid,name,old,new,unit in menu:
 f=Path(name);original=f.read_text();assert original.count(old)==1
 env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/native-followup/cache/'+mid
 try:
  f.write_text(original.replace(old,new))
  p.joinpath(mid+'.diff').write_bytes(subprocess.check_output(['git','diff','--',name]))
  with p.joinpath(mid+'-compile.log').open('w') as log:subprocess.run(['clang','-std=c11','-Wall','-Wextra','-Werror','-pedantic','-fsyntax-only','-Iinternal/native/runtime',unit],stdout=log,stderr=subprocess.STDOUT,check=True)
  start=time.monotonic()
  with p.joinpath(mid+'.log').open('w') as log:r=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
  es=[json.loads(l) for l in p.joinpath(mid+'.log').read_text().splitlines() if l.startswith('{')]
  result={k:sorted({e['Test'] for e in es if e.get('Action')==k and e.get('Test') and '/' not in e['Test']}) for k in ['pass','fail','skip']}
  result.update(command='ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+shlex.join(cmd)+' > '+mid+'.log 2>&1',wall_seconds=time.monotonic()-start,exit=r.returncode,file=name,line=original[:original.index(old)].count('\n')+1,change=old+' -> '+new)
  p.joinpath(mid+'-results.json').write_text(json.dumps(result,indent=2)+'\n');print(mid,json.dumps(result),flush=True)
 finally:f.write_text(original)
