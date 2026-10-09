from pathlib import Path
import subprocess,os,json,time,shlex
p=Path('review/test-defend/internal-native-split_units'); f=Path('internal/native/runtime/string_index.c'); original=f.read_text()
old='if (index == NULL || index == ADAMIC_LITERAL_INDEX) {'
start=original.index('size_t adamic_string_units_before('); at=original.index(old,start)
changed=original[:at]+original[at:].replace(old,'if (index == NULL) {',1)
regex='^(TestStringsMatchJavaScript|TestStringIndexMatchesNode|TestStringIndexCacheStatesMatchNode|TestStringViewAfterAppendMatchesNode|TestStringBuildingMatchesNode)$'
cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run',regex]
env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/native-defense/cache/D1'
try:
 f.write_text(changed)
 p.joinpath('D1.diff').write_bytes(subprocess.check_output(['git','diff','--',str(f)]))
 with p.joinpath('D1-compile.log').open('w') as log:
  subprocess.run(['clang','-std=c11','-Wall','-Wextra','-Werror','-pedantic','-fsyntax-only','-Iinternal/native/runtime',str(f)],stdout=log,stderr=subprocess.STDOUT,check=True)
 t=time.monotonic()
 with p.joinpath('D1.log').open('w') as log:r=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
 events=[json.loads(l) for l in p.joinpath('D1.log').read_text().splitlines() if l.startswith('{')]
 results={k:sorted({e['Test'] for e in events if e.get('Action')==k and e.get('Test') and '/' not in e['Test']}) for k in ['pass','fail','skip']}
 results.update(command='ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+shlex.join(cmd)+' > D1.log 2>&1',wall_seconds=time.monotonic()-t,exit=r.returncode)
 p.joinpath('D1-results.json').write_text(json.dumps(results,indent=2)+'\n');print(json.dumps(results))
finally:f.write_text(original)
