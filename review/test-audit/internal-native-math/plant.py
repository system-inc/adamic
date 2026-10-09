import pathlib,json,difflib,re,subprocess
root=pathlib.Path('internal/native/runtime');out=pathlib.Path('review/test-audit/internal-native-math'); originals={p.name:subprocess.check_output(["git","show","HEAD:"+str(p)],text=True) for p in root.glob('*.c')}
specs=[('M01','math.c','value - floored >= 0.5','value - floored > 0.5','flip a condition'),('M02','math.c','value > 0 ? 1 : -1','value < 0 ? 1 : -1','flip a condition'),('M03','math.c','signbit(left) && signbit(right)','signbit(left) || signbit(right)','flip a condition'),('M04','math.c','left < right ? left : right','left > right ? left : right','flip a condition'),('M05','number.c','sqrt(base + 0)','sqrt(base)','drop arithmetic operand'),('M06','number.c',"exact[keep] >= '5'","exact[keep] > '5'",'flip a condition'),('M07','number.c','bool negative = value < 0;','bool negative = signbit(value);','change an option'),('M08','maybe.c','bits = 0x7ff8000000000000u;','bits = 0x7ff8000000000001u;','change a constant'),('M09','maybe.c','if (value.present) {','if (!value.present) {','flip a condition'),('M10','maybe.c','bits == ADAMIC_UNDEFINED_BITS','bits != ADAMIC_UNDEFINED_BITS','flip a condition'),('M11','normalize.c','if (mapping == NULL || (mapping->compatibility && !compatibility)) {','if (mapping == NULL || (mapping->compatibility && compatibility)) {','flip a condition'),('M12','normalize.c','compatibility = true, composed = true;','compatibility = false, composed = true;','change an option'),('M13','normalize.c','compatibility = false, composed = false;','compatibility = false, composed = true;','change an option'),('M14','node_buffer.c',"return (int)(c - 'a') + 26;","return (int)(c - 'a') + 25;",'change a constant'),('M15','node_crypto.c','hash->slots[1].number = 1;','/* dropped finalized flag assignment */','drop a statement')]
# M05 is outside the fixed menu: replace it with a permitted constant change.
specs[4]=('M05','number.c','exponent == 0.5','exponent == 0.25','change a constant')
mutants=[];switched=dict(originals)
for mid,file,old,new,menu in specs:
 text=originals[file];assert text.count(old)==1,(mid,text.count(old));line=text[:text.index(old)].count('\n')+1
 changed=text.replace(old,new);diff=''.join(difflib.unified_diff(text.splitlines(True),changed.splitlines(True),fromfile='a/internal/native/runtime/'+file,tofile='b/internal/native/runtime/'+file));(out/(mid+'.diff')).write_text(diff)
 if old.startswith('bool negative ='):
  replacement='bool negative = audit_mutant("'+mid+'") ? signbit(value) : value < 0;'
 elif old.endswith(';'):
  replacement='if (audit_mutant("'+mid+'")) { '+new+' } else { '+old+' }'
 elif old.startswith('if ('):
  replacement='if (audit_mutant("'+mid+'") ? ('+new[4:-3]+') : ('+old[4:-3]+')) {'
 else:replacement='(audit_mutant("'+mid+'") ? ('+new+') : ('+old+'))'
 switched[file]=switched[file].replace(old,replacement)
 mutants.append(dict(id=mid,file='internal/native/runtime/'+file,line=line,old=old,new=new,menu=menu))
probes=[]
entries=[('math.c','adamic_math_round','return 0;'),('math.c','adamic_math_sign','return 0;'),('math.c','adamic_math_max','return 0;'),('math.c','adamic_math_min','return 0;'),('number.c','adamic_power','return 0;'),('number.c','adamic_number_to_fixed','return adamic_retain(&adamic_string_empty);'),('maybe.c','adamic_maybe_number_pack','return 0;'),('maybe.c','adamic_maybe_number_unpack','return (adamic_maybe_number){false, 0};'),('normalize.c','adamic_string_normalize','return adamic_retain(&adamic_string_empty);'),('node_buffer.c','adamic_node_buffer_copy','return NULL;'),('node_buffer.c','adamic_node_buffer_from','return NULL;'),('node_buffer.c','adamic_node_buffer_string','return adamic_retain(&adamic_string_empty);'),('node_crypto.c','adamic_node_hash_new','return NULL;'),('node_crypto.c','adamic_node_hash_update','return NULL;'),('node_crypto.c','adamic_node_hash_digest','return adamic_retain(&adamic_string_empty);')]
for i,(file,entry,ret) in enumerate(entries,1):
 pid='P%02d'%i;text=originals[file];match=re.search(r'[^\n]*\b'+entry+r'\([^;]*?\)\s*\{',text,re.S);assert match
 pos=match.end();line=text[:pos].count('\n')+1
 changed=text[:pos]+'\n\t'+ret+text[pos:];(out/(pid+'.diff')).write_text(''.join(difflib.unified_diff(text.splitlines(True),changed.splitlines(True),fromfile='a/internal/native/runtime/'+file,tofile='b/internal/native/runtime/'+file)))
 match=re.search(r'[^\n]*\b'+entry+r'\([^;]*?\)\s*\{',switched[file],re.S);pos=match.end();switched[file]=switched[file][:pos]+'\n\tif (audit_mutant("'+pid+'")) { '+ret+' }'+switched[file][pos:];probes.append(dict(id=pid,file=file,entry=entry,line=line))
for file in set(x[1] for x in specs):
 switched[file]=switched[file].replace('#include "adamic.h"','#include "adamic.h"\n#include "audit_mutant.h"');(root/file).write_text(switched[file])
(root/'audit_mutant.h').write_text('#include <stdlib.h>\n#include <string.h>\nstatic inline bool audit_mutant(const char *id) { const char *value = getenv("ADAMIC_MUTANT"); return value != NULL && strcmp(value, id) == 0; }\n')
(out/'plan.json').write_text(json.dumps(dict(mutants=mutants,probes=probes),indent=2));inventory=[]
for file in set(x[1] for x in specs):
 for m in re.finditer(r'^(?:static )?[\w *]+\b(\w+)\([^;]*?\)\s*\{',originals[file],re.M):inventory.append(dict(file=file,function=m[1],line=originals[file][:m.start()].count('\n')+1))
(out/'functions.json').write_text(json.dumps(inventory,indent=2))
