import re,json
from pathlib import Path

def count(path):
 s=Path(path).read_text();rows=[];function=''
 # Both field names have real conflicting generated layouts. fieldSlot can only
 # emit a naked slot load through the uniform proof; otherwise it emits a shape
 # condition or name lookup. This zero-direct assertion is specific to this C.
 layouts={name:set() for name in ['text','kind']}
 for fields in re.findall(r'static const char \*const adamic_shape_\d+_names\[\] = \{([^}]*)\};',s):
  names=re.findall(r'"([^"]*)"',fields)
  for name in layouts:
   if name in names:layouts[name].add(names.index(name))
 if any(len(indices)<2 for indices in layouts.values()):raise RuntimeError('direct-count proof no longer holds; inspect emitter again')
 for i,line in enumerate(s.splitlines(),1):
  m=re.match(r'static .* (adamic_function_\w+)\([^;]*\) \{',line)
  if m:function=m[1]
  for field in re.findall(r'adamic_object_field\([^\n]*?, "(text|kind)", &adamic_cache_\d+\)',line):
   rows.append(dict(line=i,function=function,field=field,access='store_address' if re.match(r'\s*adamic_value \*adamic_temporary_\d+ =',line) else 'load',path='shape_tested' if '->shape ==' in line else 'lookup',source=line.strip()))
 report={'generated_C':path,'generated_layout_offsets':{k:sorted(v) for k,v in layouts.items()},'sites':rows,'counts':{}}
 for scope in ['whole','scanner']:
  report['counts'][scope]={}
  for field in ['text','kind']:
   loads=[r for r in rows if r['field']==field and r['access']=='load' and (scope=='whole' or '_Scanner_' in r['function'])]
   report['counts'][scope][field]={'direct':0,'shape_tested':sum(r['path']=='shape_tested' for r in loads),'lookup':sum(r['path']=='lookup' for r in loads)}
 return report
if __name__=='__main__':
 for label,path in [('parse','scratch/emitter-speed/input-layout/before.c'),('full-batch8','scratch/emitter-speed/no-freeze/batch8-before.c')]:
  report=count(path);Path('scratch/emitter-speed/input-layout/'+label+'-counts.json').write_text(json.dumps(report,indent=2)+'\n');print(label,json.dumps(report['counts']),report['generated_layout_offsets'])
