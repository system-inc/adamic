import pathlib,json,re,subprocess,statistics,difflib,shutil
p=pathlib.Path('/tmp/u129/evidence');root=pathlib.Path('stage1/cohere/markdownblocks');base='d054e3578d4e53bef8cb8c9ab82d6ace69cdafeb';groups=json.loads((p/'groups.json').read_text());plan=json.loads((p/'plan.json').read_text());runs=json.loads((p/'matrix-timings.json').read_text());member_group={n:g for g,ns in groups for n in ns}
scratch={};original={}
# Reconstruct the exact switch source from its saved unified diff in an isolated worktree-free directory.
for item in plan:
 file=item['file'];original[file]=subprocess.check_output(['git','show',base+':'+file],text=True)
# Reconstruct original-to-switch line mappings from the retained diff.
chunks=(p/'switch.diff').read_text().split('diff --git ')[1:]
for chunk in chunks:
 lines=chunk.splitlines(True);file=lines[0].split(' b/')[1].strip();old=original[file].splitlines(True);out=[];cursor=0;active=False
 for line in lines[1:]:
  h=re.match(r'@@ -(\d+)(?:,\d+)? \+\d+(?:,\d+)? @@',line)
  if h:
   start=int(h.group(1))-1;out.extend(old[cursor:start]);cursor=start;active=True;continue
  if not active:continue
  if line.startswith(' '):out.append(line[1:]);cursor+=1
  elif line.startswith('-'):cursor+=1
  elif line.startswith('+'):out.append(line[1:])
 out.extend(old[cursor:]);scratch[file]=''.join(out)
def original_failure(line):
 match=re.search(r'([\w]+_test\.go):(\d+): (.*)',line)
 if not match:return line.strip()
 path=str(root/match.group(1));observed=int(match.group(2));original_line=observed
 if path in scratch:
  sm=difflib.SequenceMatcher(None,original[path].splitlines(),scratch[path].splitlines(),autojunk=False)
  for tag,a,b,c,d in sm.get_opcodes():
   if c<=observed-1<d:
    original_line=a+(observed-1-c)+1 if tag=='equal' else a+1;break
 return line[:match.start()]+match.group(1)+':'+str(original_line)+': '+match.group(3)
def events(id):return [json.loads(l) for l in (p/(id+'-matrix.log')).read_text().splitlines()]
matrix=[];rows=[]
for item in plan:
 es=events(item['id']);observed={}
 for g,ns in groups:
  statuses={n:[e['Action'] for e in es if e.get('Test')==n and e['Action'] in ('pass','fail','skip')] for n in ns}
  if not any(statuses.values()):continue
  observed[g]='fail' if any('fail' in s for s in statuses.values()) else 'pass' if all(s and s[-1]=='pass' for s in statuses.values()) else 'unknown'
 matrix.append(dict(id=item['id'],kind=item['kind'],observed=observed,selected_rows=[groups[i][0] for i in item['groups']],selected_members=sum([groups[i][1] for i in item['groups']],[]),timeouts=any('test timed out after' in e.get('Output','') for e in es)))
(p/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n')
for i,(g,ns) in enumerate(groups):
 secs=[]
 for j in range(1,4):
  for e in [json.loads(l) for l in (p/f'group-{i}-timing-{j}.log').read_text().splitlines()]:
   if e['Action']=='pass' and 'Test' not in e:secs.append(e['Elapsed'])
 applicable=[m for m in matrix if g in m['observed'] and m['kind']=='construction'];kills=[m['id'] for m in applicable if m['observed'][g]=='fail'];probed=[m for m in matrix if g in m['observed'] and m['kind']=='probe'];pk=[m['id'] for m in probed if m['observed'][g]=='fail'];failure=None;evidence='No construction matrix: subprocess helper for TestMarkdownLayoutSetupHasNoDeadline.'
 if kills:
  id=kills[-1];es=events(id);member=next(e['Test'] for e in es if e['Action']=='fail' and e.get('Test') in ns)
  outputs=[e.get('Output','').strip() for e in es if e.get('Test')==member and e['Action']=='output'];output=next((s for s in outputs if re.search(r'\w+_test.go:\d+:',s) and re.search(r'failed|no such file|setup has deadline|linking|Is a directory',s)),'--- FAIL: '+member)
  failure=id+': '+original_failure(output);command=next(r['command'] for r in runs if r['id']==id and r['kind']=='matrix');evidence='ADAMIC_MUTANT='+id+' ADAMIC_BUILD_CACHE_DIR=/tmp/u129/cache/'+id+' '+' '.join(command)+'; '+failure
 matrixrows=sorted(set(x for m in applicable for x in m['selected_rows']))
 row=dict(test=g,package=str(root),file=str(root/'setup_deadline_markdownblocks_test.go'),seconds=statistics.median(secs),oracle='self: successful product construction and artifact reads; Go oracle executables are built but not run or compared. Empty helper return is not asserted by this wrapper.' if i>=2 else 'self: no setup deadline, subprocess must end with its work deadline, and setup context must remain uncanceled.' if i==0 else 'Subprocess entry activated by ADAMIC_MARKDOWN_LAYOUT_DEADLINE_WORKER=1; parent checks its termination.',oracle_kind='self',kills=kills,unique_kills=[],last_proven_fail=failure,verdict='helper' if i==1 else 'setup-check' if kills else 'cannot-judge',subsumed_by=[],mutants_in_matrix=len(applicable),probe_kills=pk,subsumer_seconds=None,vacuous=None if not probed else all(m['observed'][g]=='pass' for m in probed),bounded=True,matrix_rows=matrixrows,evidence=evidence,members=ns)
 if i==1:row['parent']='TestMarkdownLayoutSetupHasNoDeadline'
 if i==9:row['oracle']='self: deserialize the cached product manifest, read backend files, and require nonempty Go binary path; no executable or formatting output is compared.'
 rows.append(row)
(p/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
# Conservative lexical inventory using original files, not the added switch helpers.
nodes={}
for f in root.glob('*.go'):
 text=subprocess.check_output(['git','show',base+':'+str(f)],text=True);matches=list(re.finditer(r'^func (?:\([^\n]+\) )?(\w+)\(',text,re.M))
 for j,m in enumerate(matches):nodes[m.group(1)]={'file':str(f),'line':text[:m.start()].count('\n')+1,'body':text[m.start():matches[j+1].start() if j+1<len(matches) else len(text)]}
edges={k:{n for n in nodes if re.search(r'\b'+re.escape(n)+r'\s*\(',v['body']) and n!=k} for k,v in nodes.items()}
def closure(n):
 seen=set();pending=[n]
 while pending:
  k=pending.pop()
  if k in seen:continue
  seen.add(k);pending.extend(edges.get(k,set())-seen)
 return seen
inventory={g:sorted(set.union(*(closure(n) for n in ns))) for g,ns in groups};callers={fn:sorted(n for n in nodes if n.startswith('Test') and fn in closure(n)) for fn in {x['function'] for x in plan}}
(p/'static-function-inventory.json').write_text(json.dumps({'base':base,'method':'Conservative lexical package call graph, not dynamic coverage; external library internals not resolved','row_functions':inventory,'candidate_callers':callers,'locations':{n:{k:v for k,v in nodes[n].items() if k!='body'} for n in set.union(*map(set,inventory.values()))}},indent=2))
# Report below is completed by the root after parsing observations.
print(json.dumps([{'test':r['test'],'seconds':r['seconds'],'kills':r['kills'],'vacuous':r['vacuous']} for r in rows],indent=2))
