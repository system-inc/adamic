from pathlib import Path
import re,json
root=Path.cwd();out=root/'review/compiler/loom-r-slice-5';records=[]
pattern=re.compile(r'^<<<<<<< HEAD\n(.*?)^=======\n(.*?)^>>>>>>> [^\n]*\n',re.M|re.S)
for name in ['cloud/admission-corpus/negative-witnesses.json','internal/javascript/javascript.go','internal/native/emit_objects.go','internal/native/reuse.go','internal/native/runtime/adamic.h','internal/native/runtime/class_static.c','internal/native/runtime/object.c','internal/native/runtime/union.c']:
 p=root/name;s=p.read_text();index=0
 def resolve(m):
  global index
  index+=1;ours,theirs=m.groups();reason=''
  if name.endswith('negative-witnesses.json'):
   result=theirs;reason='Retain main schema/comment and install the source eight exact-hash ruled witnesses.'
  elif name.endswith('javascript.go'):
   result=theirs.replace('if expression.Readiness != "" {','// Async programs skip readiness, as on main.\n\t\tif expression.Readiness != "" && e.program.Async == nil {');reason='Preserve placeholder global declaration check and main async readiness exclusion.'
  elif name.endswith(('emit_objects.go','reuse.go')):
   result=ours.replace('[field.Uninitialized]', '[field.Uninitialized && !field.Unset]');reason='Keep main slot-derived metadata index and source distinction between unset payload and uninitialized storage.'
  elif name.endswith('adamic.h'):
   result=ours.replace('slot == object->shape->count','slot == object->shape->count || !adamic_object_present(object, slot)');reason='Use one main packed atomic cache snapshot, adding source per-object presence check on its local slot.'
  elif name.endswith('class_static.c'):
   result=ours+'        adamic_object_publish(object, adamic_slot_index(object, slot));\n';reason='Publish the field while retaining main slot-derived readiness bookkeeping.'
  elif name.endswith('object.c'):
   if index==1:
    result='\t\tif (!adamic_object_present(source, index)) continue;\n'+ours;reason='Skip absent construction fields and retain main zero initializer for packed cache and owner-aware copy below.'
   elif index==2:
    result=ours.replace('slot == object->shape->count','slot == object->shape->count || !adamic_object_present(object, slot)');reason='Keep atomic packed lookup and add per-object presence without rereading shared cache.'
   else:
    result=ours+'\tadamic_object_publish(object, adamic_slot_index(object, slot));\n';reason='Initialize and publish the own slot, preserving main rule not to initialize a static parent.'
  elif name.endswith('union.c'):
   result=theirs.replace('entry = shape_types;', 'entry = __atomic_load_n(&shape_types, __ATOMIC_ACQUIRE);');reason='Retain source construction scalar tagging and main acquire publication of immutable shape metadata.'
  records.append(dict(file=name,hunk=index,main=ours,source=theirs,resolution=result,reason=reason));return result
 s=pattern.sub(resolve,s);assert '<<<<<<<' not in s;p.write_text(s)
# Two cleanly merged source additions still used the removed cache.index member.
p=root/'internal/native/runtime/library_object.c';s=p.read_text();assert s.count('adamic_object_publish(target, cache.index);') + s.count('adamic_object_publish(target, adamic_slot_index(target, slot));') == 1;s=s.replace('adamic_object_publish(target, cache.index);','adamic_object_publish(target, adamic_slot_index(target, slot));');p.write_text(s)
p=root/'internal/native/runtime/object.c';s=p.read_text();old='\tif (slot == NULL && cache->index < object->shape->count) { slot = &object->slots[cache->index]; }';assert s.count(old)==1
new='''\t// An absent reserved field still has writable storage. Find it without rereading a shared cache.
\tif (slot == NULL) {
\t\tfor (size_t index = 0; index < object->shape->count; index++) {
\t\t\tif (strcmp(object->shape->names[index], name) == 0) { slot = &object->slots[index]; break; }
\t\t}
\t}''';s=s.replace(old,new);p.write_text(s)
(out/'conflict-hunks.json').write_text(json.dumps(records,indent=2)+'\n')
