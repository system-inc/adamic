exec(open('/tmp/u047-run.py').read().split("if __name__")[0])
import difflib
menu=json.loads((out/'menu.json').read_text()); files={m['file'] for m in menu}; files.add('internal/native/runtime/adamic.h')
original={f:(root/f).read_text() for f in files}
(out/'scratch').mkdir(exist_ok=True)
for f,s in original.items():
 p=out/'scratch'/f;p.parent.mkdir(parents=True,exist_ok=True);p.write_text(s)
# Validate every production mutation independently against the unmodified starting tree.
for m in menu:
 if m['kind']!='mutant':continue
 f=m['file'];(root/f).write_text(original[f].replace(m['old'],m['new'],1))
 try:
  run('apply-'+m['id'],['git','apply','--reverse','--check',str(out/'diffs'/(m['id']+'.diff'))])
  if f.endswith('.go'):code=run('vet-'+m['id'],['timeout','120','go','vet','./'+str(pathlib.Path(f).parent)+'/'])
  else:
   flags=['-std=c11','-Wall','-Wextra','-Werror','-Wcast-function-type-strict','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O2']
   code=run('clang-'+m['id'],['clang']+flags+['-c',f,'-o','/tmp/u047-'+m['id']+'.o'])
   if code==0:code=run('clang-sanitize-'+m['id'],['clang']+flags+['-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all','-DADAMIC_SLABS','-c',f,'-o','/tmp/u047-'+m['id']+'-san.o'])
  if code:raise RuntimeError('validation failed '+m['id'])
 finally:(root/f).write_text(original[f])
# Install the complete predeclared selector once.
for f,s in original.items():
 for m in menu:
  if m['file']==f and m['kind']!='setup':s=s.replace(m['old'],m['switch'],1)
 (root/f).write_text(s)
helper='package %s\nimport "os"\nfunc u047Mutant(id string) bool { return os.Getenv("ADAMIC_MUTANT") == id }\nfunc u047Number(id string, normal, changed int) int { if u047Mutant(id) { return changed }; return normal }\n'
for pkg in ['native','lower']:(root/('internal/'+pkg+'/u047_switch.go')).write_text(helper%pkg)
f=root/'internal/native/runtime/adamic.h';s=f.read_text();s=s.replace('#include <stdbool.h>','#include <stdlib.h>\n#include <string.h>\nstatic inline bool_dummy_not_used_placeholder') if False else s
s=s.replace('#include <math.h>','#include <math.h>\n#include <stdlib.h>\n#include <string.h>\nstatic inline bool adamic_u047(const char *id) { const char *selected = getenv("ADAMIC_MUTANT"); return selected != NULL && strcmp(selected, id) == 0; }');f.write_text(s)
run('switch-vet',['go','vet','./internal/native/','./internal/lower/'])
run('switch-build',['go','test','-c','./internal/native/','-o','/tmp/u047-native.test'])
# Environment-selected C behavior shares one content-hashed runtime build; generated products get separate cache dirs.
for m in menu:
 if m['kind']=='setup':continue
 env=os.environ.copy();env['ADAMIC_MUTANT']=m['id'];env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u047/cache/'+m['id']
 code=run(m['id'],['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run',pattern],env)
 # Convert verbose binary output with go tool test2json to retain the exact observed matrix.
 (out/(m['id']+'.json')).write_bytes((out/(m['id']+'.log')).read_bytes())
 # A panic is a binary abort. Only individually rerun the requested rows not already observed to finish.
 if 'panic:' in (out/(m['id']+'.log')).read_text():
  events=[json.loads(l) for l in (out/(m['id']+'.json')).read_text().splitlines()];finished={e.get('Test') for e in events if e['Action'] in ['pass','fail','skip']}
  for r in rows:
   if r not in finished:run(m['id']+'-alone-'+r,['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run','^'+r+'$'],env)
# Special permitted setup check, only the helper under construction is changed.
m=next(x for x in menu if x['id']=='S1');f=m['file'];(root/f).write_text(original[f].replace(m['old'],m['new'],1))
run('S1',['go','test','-json','-count=1','-timeout','90s','./internal/native/','-run','^TestResidentSetUnits$'])
for f,s in original.items():(root/f).write_text(s)
for pkg in ['native','lower']:(root/('internal/'+pkg+'/u047_switch.go')).unlink()
