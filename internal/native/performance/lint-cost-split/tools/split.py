from pathlib import Path
import re,subprocess,json,collections,math
root=Path('/workspace/lint-cost-baseline');binary=str(root/'scanner');events=[];bases={};cur=None
header=re.compile(r'^(\S+)\s+(\d+)\s+([\d.]+): task-clock:u:')
mmap=re.compile(r'^\S+\s+(\d+)\s+[\d.]+: PERF_RECORD_MMAP2 .*\[(0x[0-9a-f]+)\(0x[0-9a-f]+\) @ (0x[0-9a-f]+|0).* /workspace/lint-cost-baseline/scanner$')
frame=re.compile(r'^\s+([0-9a-f]+) (.*?) \((.*?)\)\s*$')
for line in (root/'perf-mapped-script.txt').open():
 m=mmap.match(line)
 if m:bases[int(m[1])]=int(m[2],16)-int(m[3],16);continue
 m=header.match(line)
 if m:
  cur={'comm':m[1],'pid':int(m[2]),'time':float(m[3]),'frames':[]};events.append(cur);continue
 m=frame.match(line)
 if m and cur:cur['frames'].append((int(m[1],16),m[2],binary if m[3]=='inlined' else m[3]))
addresses=set()
for e in events:
 for ip,sym,dso in e['frames']:
  if dso==binary and e['pid'] in bases:addresses.add(ip-bases[e['pid']])
argv=['/workspace/adamic-tools/llvm/bin/llvm-addr2line','-e',binary,'-a','-f','-i']
p=subprocess.run(argv,input='\n'.join(hex(v) for v in sorted(addresses))+'\n',text=True,stdout=subprocess.PIPE,stderr=subprocess.PIPE,check=True)
(root/'addr2line.txt').write_text(p.stdout);locations={};current=None;pending=[]
for line in p.stdout.splitlines():
 if re.fullmatch(r'0x[0-9a-fA-F]+',line):current=int(line,16);locations[current]=[];pending=[]
 elif current is not None:
  pending.append(line)
  if len(pending)==2:locations[current].append(tuple(pending));pending=[]
# Prefixes end before the first rule-specific action. Multiline also compares the cached kind
# between bodies; only its primary kind comparison and cached-kind destruction lines are tagged.
rules={42:'no-octal-escape',194:'no-unexpected-multiline',199:'no-unused-private-class-members',201:'no-useless-constructor',215:'prefer-template',216:'react/forward-ref-uses-ref',217:'react/jsx-no-comment-textnodes',218:'react/no-find-dom-node',219:'react/no-is-mounted',220:'react/no-redundant-should-component-update'}
source=(root/'main.c').read_text().splitlines()
bounds={};entry_names={i:name for i,name in rules.items()}
markers={42:'_Context_raw(',199:'_Context_members(',201:'_Context_children(',216:'_Context_child(',217:'_Context_node(',218:'_Context_child(',219:'_Context_child(',220:'_pure = false;'}
for i in rules:
 start=next(n for n,line in enumerate(source,1) if re.match(r'static [^;]*adamic_function_'+str(i)+r'_[^;]+\) \{',line))
 end=next(n for n in range(start+1,len(source)+1) if source[n-1]=='}')
 if i==194:
  stop=next(n for n in range(start,end) if 'adamic_string_equal(adamic_local_' in source[n-1])-1
 elif i==215:
  stop=next(n for n in range(start,end) if '_string(' in source[n-1])+1
 else:stop=next(n for n in range(start,end) if markers[i] in source[n-1])-1
 bounds[i]=(start,stop)
extra_multiline=set()
start,stop=bounds[194];end=next(n for n in range(stop+1,len(source)+1) if source[n-1]=='}')
kind=re.search(r'adamic_string \* (adamic_local_\d+_kind) =', '\n'.join(source[start:end]))[1]
for n in range(start,end):
 if 'adamic_string_equal('+kind+',' in source[n-1] or source[n-1].strip()=='adamic_release('+kind+');':extra_multiline.add(n)
fn_re=re.compile(r'adamic_function_(\d+)_')
def rule_for(sym):
 m=fn_re.search(sym)
 if not m:return None
 i=int(m[1]);owner={193:194,195:199,196:199,197:199,198:199,200:201,205:215,206:215,207:215,208:215,209:215,210:215,211:215,212:215,213:215,214:215}.get(i,i);return rules.get(owner)
def guard_loc(function,location):
 m=fn_re.search(function);ln=re.search(r'main\.c:(\d+)',location)
 if not m or not ln:return False
 i=int(m[1]);n=int(ln[1]);return i in bounds and (bounds[i][0]<=n<=bounds[i][1] or i==194 and n in extra_multiline)
def guarded(e):
 for ip,sym,dso in e['frames']:
  if dso!=binary or e['pid'] not in bases:continue
  for function,location in locations.get(ip-bases[e['pid']],[]):
   if guard_loc(function,location):return True
 return False
alloc=re.compile(r'^(adamic_allocate|adamic_free|adamic_region_alloc|adamic_region_allocate|adamic_array_new|adamic_object_new|adamic_cell_new|adamic_closure_new|adamic_map_new|new_chunk|give|drop_chunk|malloc|calloc|realloc|cfree|free|_int_malloc|_int_free|_int_realloc|malloc_consolidate|sysmalloc|munmap_chunk|.*malloc.*|.*free.*)$')
own=re.compile(r'^(adamic_retain|adamic_release|release_last|let_go|adamic_.*_free_children)$')
string=lambda s:s.startswith('adamic_string_') or s in {'units_next','units_before','locate','utf8_next','decode','scalar','utf8_decode','adamic_equal_strings','unit_at','index_get','units_at','string_view','adamic_to_string'}
scanner=lambda s:'_Scanner_' in s or re.search(r'adamic_function_(?:[0-9]|1[0-9]|20|21|38)_',s) is not None
parser=lambda s:any(x in s for x in ['_Parser_','_Statements_','_ParseNode_','_Speculation_']) or re.search(r'adamic_function_(?:6[7-9]|7[01])_',s) is not None
# Generic runtime instructions are assigned by the nearest semantic caller after physical
# ownership, allocation, dispatch and string functions have been peeled off.
def bucket(e,isguard):
 if not e['frames']:return 'remainder: missing stack'
 if e['comm']!='scanner':return 'startup and I/O: harness'
 rawframes=e['frames'];leaf=rawframes[0][1];dso=rawframes[0][2];frames=[]
 for ip,sym,library in rawframes:
  if library==binary and e['pid'] in bases:
   for function,location in locations.get(ip-bases[e['pid']],[]):frames.append((ip,function,library))
  frames.append((ip,sym,library))
 if own.fullmatch(leaf):return 'retains and releases'
 if alloc.fullmatch(leaf) and not own.fullmatch(leaf):return 'allocation and freeing'
 if leaf in ['adamic_virtual','adamic_object_callee','adamic_closure_call']:return 'virtual dispatch'
 if string(leaf):return 'string work'
 if any(x in dso for x in ['ld-linux','ld.so']):return 'startup and I/O'
 for ip,sym,library in frames:
  if own.fullmatch(sym):return 'retains and releases'
  if alloc.fullmatch(sym):return 'allocation and freeing'
  if sym in ['adamic_virtual','adamic_object_callee','adamic_closure_call']:return 'virtual dispatch'
  if string(sym):return 'string work'
  if sym in ['adamic_read_text_file','adamic_arguments','adamic_write_line','adamic_file_close'] or any(x in sym for x in ['_readTextFile','_programArguments']):return 'startup and I/O'
  if '_Context_enabled' in sym:return 'rule entry and selection'
  if any(x in sym for x in ['_Context_node','_Context_kind','_Context_mapParents','_Context_parent','_Context_child']):return 'tree walk and node access'
  if '_Parser_node' in sym:
   return 'tree walk and node access' if any('_Context_' in s or rule_for(s) or '_visit' in s for _,s,_ in frames) else 'scanner and parser'
  r=rule_for(sym)
  if r:return 'rule entry and selection' if isguard else 'rule logic: '+r
  if '_visit' in sym:return 'tree walk and node access'
  if scanner(sym) or parser(sym):return 'scanner and parser'
  if re.search(r'adamic_function_(222|223|224|225)_',sym) or sym in ['main','_start','__libc_start_main']:return 'startup and I/O'
 return 'remainder: unclassified'
counts=collections.Counter();guard_counts=collections.Counter();symbols=collections.Counter();by_pid=collections.defaultdict(collections.Counter);rows=[]
for e in events:
 g=guarded(e);b=bucket(e,g);counts[b]+=1;by_pid[e['pid']][b]+=1
 if g:guard_counts[b]+=1
 if e['frames']:symbols[(b,e['frames'][0][1],e['frames'][0][2])]+=1
 rows.append({'pid':e['pid'],'bucket':b,'guard':g,'leaf':e['frames'][0][1] if e['frames'] else '','time':e['time']})
assert sum(counts.values())==len(events)
timing=json.loads(Path('/workspace/lint-cost-timing.json').read_text());run=min((r for r in timing['runs'] if r['name']=='today'),key=lambda r:r['wall']);user_fraction=run['user']/run['wall'];total=len(events)
result={'samples':total,'counts':dict(counts),'shares':{k:v/total for k,v in counts.items()},'guard_counts':dict(guard_counts),'guard_samples':sum(guard_counts.values()),'guard_share':sum(guard_counts.values())/total,'user_fraction_from_best_unprofiled_run':user_fraction,'unsampled_wall_fraction':1-user_fraction,'best_unprofiled_run':run,'seconds_at_2_11':{k:2.11*user_fraction*v/total for k,v in counts.items()},'unsampled_seconds_at_2_11':2.11*(1-user_fraction),'per_pid':{str(k):dict(v) for k,v in by_pid.items()},'guard_boundaries':{str(k):v for k,v in bounds.items()},'extra_multiline_lines':sorted(extra_multiline),'symbols':[{'bucket':b,'symbol':s,'dso':d,'samples':n} for (b,s,d),n in symbols.most_common()]}
Path('/workspace/lint-cost-split.json').write_text(json.dumps(result,indent=2)+'\n')
with Path('/workspace/lint-cost-sample-buckets.jsonl').open('w') as out:
 for row in rows:out.write(json.dumps(row)+'\n')
print('samples',total,'guard',result['guard_share'],'user',user_fraction)
for k,v in counts.most_common():print(k,v,round(v/total*100,3),round(result['seconds_at_2_11'][k],4))
