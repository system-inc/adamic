#!/usr/bin/env python3
"""Reconcile simulated events; rank explicit cycle-cost models, never hardware cycles."""
import collections,json,re,sys
from pathlib import Path
s=Path(sys.argv[1]); repo=Path.cwd()
def ranges(path):
 mapping={}; lines=path.read_text().splitlines()
 for i,line in enumerate(lines):
  if line.startswith((' ','\t','//')):continue
  m=re.search(r'\b(\w+)\([^;]*\)\s*\{',line)
  if not m:continue
  depth=0
  for j in range(i,len(lines)):
   mapping[j+1]=m[1];depth+=lines[j].count('{')-lines[j].count('}')
   if depth==0:break
 return mapping
header=ranges(repo/'internal/native/runtime/adamic.h'); heap=ranges(repo/'internal/native/runtime/heap.c'); source=ranges(s/'parse.c')
release={'release_last','adamic_release','adamic_object_free_children','release_field','adamic_map_free_children','adamic_weak_forget','adamic_string_free_index','free_one','let_go','destroy_last_reference'}
slab={'take','give','new_chunk','list_chunk','unlist_chunk','deallocate'}
character={'adamic_string_char_code','adamic_string_bmp_view','adamic_string_unit_view','adamic_string_units','adamic_string_locate','adamic_string_units_before','usable','build','width','decode','unit_at','sequence','adamic_string_code_point_at'}
field={'callee_cache_miss','callee_index','adamic_object_field','adamic_object_data_field','adamic_object_find','adamic_object_write_field','adamic_object_check_write','adamic_object_callee','adamic_static_field','adamic_object_check_data_write'}
def bucket(fn,file,line):
 inline=header.get(line,'') if file=='adamic.h' else ''
 effective=heap.get(line,fn) if file=='heap.c' else fn
 if effective in slab:return 'Slab allocator'
 if effective=='adamic_retain':return 'Retains'
 if effective in release:return 'Releases and child destruction'
 if effective=='adamic_allocate' or re.search(r'(?:^|_)malloc|(?:^|_)realloc|(?:^|_)calloc|(?:^|_)free$',fn):return 'Allocation entry and libc allocator'
 if fn=='adamic_string_equal':return 'String equality'
 if file=='input.c' or fn=='adamic_read_text_file':return 'File reading and UTF-8 input decode'
 if fn in character or inline in {'adamic_string_length','adamic_string_char_code_at'}:return 'Character reads and UTF-16 indexing'
 if fn in field or inline in field:return 'Object field and call plumbing'
 if fn.startswith('adamic_string_slice') or fn=='adamic_string_share':return 'Substrings'
 if fn.startswith('adamic_string_') or (file.startswith('string_') and not fn.startswith('adamic_function_')) or fn=='units_next':return 'Other string operations'
 if re.search(r'_(Scanner|Speculation)_',fn) or re.search(r'_(isIdentifierStart|isIdentifierPart|isLineBreak|isSpace|arrowAhead|skipTrivia)\b',fn):return 'Scanner generated control'
 if '_Context_mapParents' in fn:return 'Parent map generated control'
 if '_ParseNode_new' in fn or '_Parser_make' in fn:return 'Node construction generated control'
 if file=='parse.c' and re.search(r'_\d+_lines$',source.get(line,'')):return 'Line table generated control'
 return 'Remainder: parser, arrays, driver, libc and unattributed inline'
symbols={}; files={}; fn=None; file='???'; line=0; pending=False; buckets={}; events=[]; total=[]; footer=[]
for text in (s/'native-instrument.callgrind').read_text().splitlines():
 if text.startswith('events:'):events=text.split()[1:]
 elif text.startswith('summary:'):total=list(map(int,text.split()[1:]))
 elif text.startswith('totals:'):footer=list(map(int,text.split()[1:]))
 elif text.startswith(('fl=','fi=','fe=','cfl=','cfi=','fn=','cfn=')):
  key,value=text.split('=',1); table=symbols if key.endswith('fn') else files
  m=re.match(r'\((\d+)\)(?: (.*))?$',value)
  if m:
   if m[2] is not None:table[m[1]]=m[2]
   value=table[m[1]]
  if key=='fn':fn=value
  elif not key.startswith('c'):file=Path(value).name
 elif text.startswith('calls='):pending=True
 elif fn and text and text[0] in '0123456789+-*':
  parts=text.split();position=parts[0]
  if position!='*':line=line+int(position) if position[0] in '+-' else int(position)
  values=list(map(int,parts[1:]));values += [0]*(len(events)-len(values))
  if pending:pending=False;continue
  name=bucket(fn,file,line); vector=buckets.setdefault(name,[0]*len(events))
  for i,v in enumerate(values):vector[i]+=v
def reconcile(summary):
 if [sum(v[i] for v in buckets.values()) for i in range(len(events))]!=summary:raise RuntimeError('event self accounting differs from summary: '+str([sum(v[i] for v in buckets.values()) for i in range(len(events))])+ ' vs '+str(summary))
reconcile(footer)
def scores(l1,ll,branch):
 def score(v):
  x=dict(zip(events,v)); return x['Ir']+l1*sum(x.get(k,0) for k in ['I1mr','D1mr','D1mw'])+ll*sum(x.get(k,0) for k in ['ILmr','DLmr','DLmw'])+branch*sum(x.get(k,0) for k in ['Bcm','Bim'])
 return {name:score(v) for name,v in buckets.items()}
def ranks(values):return {name:i+1 for i,(name,_) in enumerate(sorted(values.items(),key=lambda p:-p[1]))}
ir={name:v[events.index('Ir')] for name,v in buckets.items()}; models={str(c):scores(*c) for c in [(4,50,15),(2,40,8),(10,200,25)]}
result={'events':events,'summary':dict(zip(events,total)),'self_totals':dict(zip(events,footer)),'summary_minus_self':dict(zip(events,[a-b for a,b in zip(total,footer)])),'buckets':{name:dict(zip(events,v)) for name,v in buckets.items()},'instruction_rank':ranks(ir),'model_scores':models,'model_ranks':{k:ranks(v) for k,v in models.items()},'assumptions':'Ir + L1_penalty*(I1mr+D1mr+D1mw) + LL_penalty*(ILmr+DLmr+DLmw) + branch_penalty*(Bcm+Bim). These are additive hypothetical cycle costs, not measured cycles.'}
# Every event, rather than just Ir, must reconcile; run an intentionally corrupted summary.
mutant=footer.copy();mutant[-1]+=1
try:reconcile(mutant)
except RuntimeError as error:print('cache event mutant caught:',error)
else:raise RuntimeError('accounting mutant survived')
result['mutant']='last footer event +1 rejected by vector reconciliation'
(s/'cache-rerank.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
