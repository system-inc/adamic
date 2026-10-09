import pathlib,json,shlex
p=pathlib.Path('/tmp/defend-yaml-gaps');allrows=[x for x in (p/'list.log').read_text().splitlines() if x.startswith('Test')]
def parse(files):
 statuses={};errors={}
 for f in files:
  for s in f.read_text().splitlines():
   try:d=json.loads(s)
   except:continue
   name=d.get('Test','').split('/')[0]
   if not name:continue
   if d.get('Action') in ('pass','fail','skip') and '/' not in d['Test']:
    if statuses.get(name)!='fail':statuses[name]=d['Action']
   if d.get('OutputType')=='error':errors.setdefault(name,[]).append(d['Output'].strip())
 return statuses,errors
matrix={}
for mid in ['D1','D2','D3','D4','D5']:
 files=sorted(p.glob(mid+'-batch*.log')) if mid in ['D1','D2'] else [p/(mid+'.log')]
 statuses,errors=parse(files)
 matrix[mid]={'rows_failed':[r for r,s in statuses.items() if s=='fail'],'rows_passed':[r for r,s in statuses.items() if s=='pass'],'rows_skipped':[r for r,s in statuses.items() if s=='skip'],'unknown_rows':[r for r in allrows if r not in statuses],'errors':errors}
 if mid in ['D1','D2']:assert len(statuses)==72 and len(matrix[mid]['rows_failed'])==1 and not matrix[mid]['rows_skipped'],(mid,matrix[mid])
plan={x['mutant']:x for x in json.loads((p/'plan.json').read_text())}
def attempt(mid):
 r=plan[mid];return {'mutant':mid,'file_line':r['file']+':'+str(r['line']),'change':r['change'],'rows_failed':matrix[mid]['rows_failed']}
def evidence(mid,row):
 runs=json.loads((p/(mid+'-runs.json')).read_text()) if mid in ['D1','D2'] else json.loads((p/'port-runs.json').read_text())
 run=next(r for r in runs if r.get('id')==mid and row in r['rows'])
 log=mid+f"-batch{run['batch']}.log" if 'batch' in run else mid+'.log'
 cmd=f'ADAMIC_YAML_LIBRARY=/tmp/defend-yaml-gaps/library ADAMIC_BUILD_CACHE_DIR=/tmp/defend-yaml-gaps/cache/{mid} '+shlex.join(run['command'])+' > '+log+' 2>&1'
 return cmd+'; '+'; '.join(matrix[mid]['errors'].get(row,[]))
rows=[]
for row,subsumer,mid in [('TestLexerGaps','TestStructuralPositionRefusal','D1'),('TestStructuralPositionRefusal','TestLexerGaps','D2')]:
 assert matrix[mid]['rows_failed']==[row]
 rows.append({'test':row,'package':'stage1/cohere/yaml','prior_verdict':'subsumed','subsumed_by':[subsumer],'defense':'defended','unique_mutant':mid+' '+plan[mid]['file']+':'+str(plan[mid]['line']),'attempts':[attempt(mid)],'rows_passed':matrix[mid]['rows_passed'],'evidence':evidence(mid,row)})
row='TestLexerMatchesGo'
valid=all(row in matrix[m]['rows_failed'] and len(matrix[m]['rows_failed'])>1 for m in ['D3','D4','D5'])
rows.append({'test':row,'package':'stage1/cohere/yaml','prior_verdict':'subsumed','subsumed_by':['TestPropsMatchGo'],'defense':'not defended' if valid else 'cannot-judge','unique_mutant':None,'attempts':[attempt(m) for m in ['D3','D4','D5']],'bounded':True,'matrix_rows':json.loads((p/'port-runs.json').read_text())[0]['rows'],'evidence':evidence('D4',row),'reason':None if valid else 'D3 and D4 are shared semantic kills. D5 exhausted 90 seconds in CST before this row ran, and its narrowed lexer replay also did not complete within the binary budget. D5 row results remain unknown; three completed attempts were not obtained.'})
(p/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n')
(p/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
print(json.dumps([{k:v for k,v in r.items() if k not in ['rows_passed','evidence']} for r in rows],indent=2))
