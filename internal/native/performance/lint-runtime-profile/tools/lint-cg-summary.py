import sys,re,json
from collections import defaultdict
from pathlib import Path
for arg in sys.argv[1:]:
 p=Path(arg);names={};callids=defaultdict(int);selfids=defaultdict(int);current=None;callee=None;edge=False;total=0
 for line in (p/'callgrind.out').read_text().splitlines():
  m=re.match(r'(c?fn)=\((\d+)\)(?: (.*))?$',line)
  if m:
   kind,identity,name=m.groups()
   if name is not None:names[identity]=name
   if kind=='fn':current=identity
   else:callee=identity
  elif line.startswith('calls='):
   callids[callee]+=int(line.split()[0][6:]);edge=True
  elif line.startswith('summary:'):total=int(line.split()[1])
  elif re.match(r'^[+*\-\d]',line):
   if edge:edge=False
   else:selfids[current]+=int(line.split()[-1])
 result={'total_instructions':total,'calls':{},'self_instructions':{}}
 for namekey,source in [('calls',callids),('self_instructions',selfids)]:
  for identity,n in source.items():
   name=names.get(identity,'?');name=re.sub(r"'\d+$",'',name)
   result[namekey][name]=result[namekey].get(name,0)+n
 (p/'callgrind-summary.json').write_text(json.dumps(result,indent=2)+'\n')
 print(p.name,total)
 for name in ('adamic_release','adamic_retain','adamic_string_equal','adamic_virtual','adamic_object_find','adamic_number_format','fmod','adamic_string_last_index_of','units_next','adamic_string_locate'):
  print(name,result['calls'].get(name,0),result['self_instructions'].get(name,0))
