import pathlib,json,re,difflib,subprocess,shutil
root=pathlib.Path('review/test-audit/internal-native-split_units'); menu=json.loads((root/'menu.json').read_text()); originals={}
def original(f):
 if f not in originals: originals[f]=pathlib.Path(f).read_text()
 return originals[f]
# Confirm all fixed diffs apply while production is clean.
for m in menu['mutations']:
 subprocess.run(['git','apply','--check',str(root/(m['id']+'.diff'))],check=True)
# Empty entry probes are separate from the fixed mutant menu.
names=['allocate','append','at','char_code','char_code_at','code_point_at','code_points','compare','concat','ends_with','equal','from_number','index_of','index_of_from','length','pad','repeat','share','slice','starts_with','trim','units']
probes=[]
for n in names:
 name='adamic_string_'+n; matches=[]
 for p in pathlib.Path('internal/native/runtime').glob('*'):
  if p.suffix not in ['.c','.h']:continue
  s=original(str(p))
  for match in re.finditer(r'^((?:static inline )?(?:adamic_string \*|adamic_array \*|adamic_maybe_number |size_t |double |bool |int )'+name+r'\([^;\n]*\) \{)',s,re.M):matches.append((str(p),match))
 assert len(matches)==1,(name,len(matches))
 f,mat=matches[0];s=original(f);sig=mat.group(1);pos=mat.end();id='P'+str(len(probes)+1)
 ret='return NULL;' if '*' in sig.split(name)[0] else 'return (adamic_maybe_number){false, 0};' if 'adamic_maybe_number' in sig else 'return 0;'
 b=s[:pos]+'\n\t'+ret+s[pos:];diff=''.join(difflib.unified_diff(s.splitlines(True),b.splitlines(True),fromfile='a/'+f,tofile='b/'+f))
 (root/(id+'.diff')).write_text(diff);probes.append(dict(id=id,entry=name,file=f,line=s[:pos].count('\n')+1,signature=sig,statement=ret))
(root/'probes.json').write_text(json.dumps(probes,indent=2)+'\n')
# Save clean sources before selector instrumentation.
pathlib.Path('/tmp/u053/originals.json').write_text(json.dumps(originals))
for m in menu["mutations"]: original(m["file"])
changes=dict(originals)
header='internal/native/runtime/adamic.h'
changes[header]=changes[header].replace('#include <stdbool.h>','#include <stdlib.h>\n#include <string.h>\nstatic inline bool_PLACEHOLDER u053_selected(const char *id) { const char *value = getenv("ADAMIC_MUTANT"); return value != NULL && strcmp(value, id) == 0; }\n#include <stdbool.h>')
# Helper uses int before bool include.
changes[header]=changes[header].replace('bool_PLACEHOLDER','int')
for m in menu['mutations']:
 id=m['id'];f=m['file'];s=changes[f];a=m['before'];b=m['after']
 if id=='M1':new='#define SHARE_FRACTION (u053_selected("M1") ? 7 : 8)'
 elif id=='M2':new='result->units = units + (u053_selected("M2") ? 2 : 1);'
 elif id=='M3':new='index->view[at++] = (uint16_t)((u053_selected("M3") ? 0xdc01 : 0xdc00) +'
 elif id=='M4':new='index = isnan(index) ? (u053_selected("M4") ? 1 : 0) : trunc(index);'
 elif id=='S1':new='for first := func() int { if os.Getenv("ADAMIC_MUTANT") == "S1" { return 1 }; return 0 }(); first < total; first += width {'
 elif id=='S2':new='count < func() int { if os.Getenv("ADAMIC_MUTANT") == "S2" { return 2 }; return 1 }() || index < 0'
 elif id=='S3':new='return (piece%s.count == s.index) != (os.Getenv("ADAMIC_MUTANT") == "S3")'
 else:continue
 changes[f]=s.replace(a,new)
for p in probes:
 f=p['file'];sig=p['signature'];assert sig in changes[f]
 changes[f]=changes[f].replace(sig,sig+'\n\tif (u053_selected("'+p['id']+'")) { '+p['statement']+' }',1)
for f,s in changes.items():
 if s!=originals[f]:pathlib.Path(f).write_text(s)
pathlib.Path('/tmp/u053/originals.json').write_text(json.dumps(originals))
