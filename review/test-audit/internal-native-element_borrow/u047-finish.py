exec(open('/tmp/u047-run.py').read().split("if __name__")[0])
import difflib
f=root/'internal/native/heap_test.go';s=f.read_text();changed=s.replace('Options{Sanitize: true, slabs: build.slabs}','Options{Sanitize: false, slabs: build.slabs}',1);assert changed!=s
(out/'diffs'/'W1.diff').write_text(''.join(difflib.unified_diff(s.splitlines(True),changed.splitlines(True),fromfile='a/internal/native/heap_test.go',tofile='b/internal/native/heap_test.go')))
f.write_text(changed)
try:
 run('W1-vet',['go','vet','./internal/native/'])
 run('W1',['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run','^TestFreedValuesAreCaughtWithSlabs$'])
finally:f.write_text(s)
run('restored-baseline',['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run',pattern])

# Make empty-entry probe diffs standalone and vet-clean by replacing the whole entry body.
menu=json.loads((out/'menu.json').read_text())
for m in menu:
 if m['kind']!='probe':continue
 f=m['file'];before=(root/f).read_text();after=before.replace(m['old'],m['new'],1)
 if f.endswith('.go'):
  start=before.index(m['old']); body=start+len(m['old']); end=before.index('\n}\n',body)+3
  value={'P1':'nil, nil','P2':'""','P3':'nil','P4':'nil, nil'}[m['id']]
  after=before[:body]+'\treturn '+value+'\n}\n'+before[end:]
  if m['id']=='P4':after=after.replace('\n\t"fmt"','').replace('\n\t"path/filepath"','')
 (out/'diffs'/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
 (root/f).write_text(after)
 try:
  run('apply-'+m['id'],['git','apply','--reverse','--check',str(out/'diffs'/(m['id']+'.diff'))])
  if f.endswith('.go'):code=run('vet-'+m['id'],['go','vet','./'+str(pathlib.Path(f).parent)+'/'])
  else:
   source='internal/native/runtime/string.c' if f.endswith('.h') else f
   flags=['-std=c11','-Wall','-Wextra','-Werror','-Wcast-function-type-strict','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O2']
   code=run('clang-'+m['id'],['clang']+flags+['-c',source,'-o','/tmp/u047-probe.o'])
  if code:raise RuntimeError('probe validation failed '+m['id'])
 finally:(root/f).write_text(before)
