import pathlib,json,subprocess,shlex,time
p=pathlib.Path('review/test-defend/internal-oracle-interface_cast/session')
while 'completed' not in (p/'D8-new-driver.log').read_text():time.sleep(1)
main=json.loads((p/'matrix.json').read_text());ext=json.loads((p/'new-matrix.json').read_text());follow=json.loads((p/'follow-matrix.json').read_text());d8new=json.loads((p/'D8-new-matrix.json').read_text());combined=[]
for x in main+[follow[-1]]:
 mid=x['mutant'];statuses=dict(x['statuses']);new=d8new if mid=='D8' else next(n for n in ext if n['mutant']==mid);statuses.update(new['statuses'])
 if mid=='D5':
  for y in follow[:-1]:statuses.update(y['statuses'])
 combined.append({'mutant':mid,'bounded':True,'discarded':mid=='D6','discard_reason':'Generated C is syntactically invalid; no verdict rests on these failures' if mid=='D6' else None,'rows_failed':sorted(k for k,v in statuses.items() if v=='fail'),'rows_passed':sorted(k for k,v in statuses.items() if v=='pass'),'unknown_rows':sorted((set(main[0]['statuses'])|set(ext[0]['statuses']))-set(statuses)) if mid=='D5' else [],'statuses':statuses,'main':x,'new':new,'isolated':follow[:-1] if mid=='D5' else []})
by={x['mutant']:x for x in combined};menu=json.loads((p/'menu.json').read_text());menu.append({'mutant':'D8','target':'TestInterfaceCastOracle','file':'internal/lower/interface_cast.go','old':'for _, member := range declared.Types() {','new':'for _, member := range declared.Types()[:min(len(declared.Types()), 3)] {','change':'off by one for the four-tag input: retain at most three union alternatives'})
rows=[]
for test,sub,ids in [('TestInterfaceCastOracle','TestInterfaceCastImportedConstruction',['D1','D7','D8']),('TestInterfaceCastImportedConstruction','TestInterfaceCastChecksMalformedRead',['D3','D4','D5']),('TestInterfaceCastScalarTags','TestInterfaceCastImportedConstruction',['D2'])]:
 unique=next((i for i in ids if by[i]['rows_failed']==[test]),None);evidence=[];attempts=[]
 for mid in ids:
  m=next(m for m in menu if m['mutant']==mid);orig=subprocess.check_output(['git','show','origin/main:'+m['file']],text=True);line=orig[:orig.index(m['old'])].count('\n')+1
  attempts.append({'mutant':mid,'file_line':m['file']+':'+str(line),'change':m['change'],'rows_failed':by[mid]['rows_failed']})
  b=by[mid]['main'];log=(mid+'.log') if mid!='D8' else 'D8-bounded.log'
  if mid=='D5':b=next(b for b in follow[:-1] if test in b['statuses']);log=b['log']
  events=[]
  for l in (p/log).read_text().splitlines():
   try:events.append(json.loads(l))
   except:pass
  err=next((e.get('Output','').strip() for e in events if e.get('Test','').split('/')[0]==test and e.get('OutputType')=='error'),None)
  if not err:err=next((e['Output'].strip() for e in events if 'panic: runtime' in e.get('Output','')), 'requested row passed; see matrix')
  evidence.append(mid+': ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/interface-defense/cache/'+mid+' '+shlex.join(b['command'])+' > '+log+' 2>&1; '+err)
 rows.append({'test':test,'package':'internal/oracle','prior_verdict':'subsumed','subsumed_by':sub,'defense':'defended' if unique else 'not defended','bounded':True,'unique_mutant':next((a['mutant']+' '+a['file_line'] for a in attempts if a['mutant']==unique),None),'attempts':attempts,'evidence':'\n'.join(evidence)})
(p/'combined-matrix.json').write_text(json.dumps(combined,indent=2));(p/'rows.json').write_text(json.dumps(rows,indent=2));(p/'final-menu.json').write_text(json.dumps(menu,indent=2))
checks=[]
for m in menu:
 subprocess.run(['git','apply','--check',str(p/(m['mutant']+'.diff'))],check=True);checks.append({'mutant':m['mutant'],'apply_check':0,'discarded':m['mutant']=='D6'})
(p/'replay-checks.json').write_text(json.dumps(checks,indent=2))
for f in ['run.py','coverage.py','extend.py','follow.py','d8new.py','prepare.py','finalize.py']:(p/f).write_text(pathlib.Path('/tmp/interface-defense',f).read_text())
scope=json.loads((p/'scope.json').read_text());outside=sorted(set(scope['current'])-set(main[0]['statuses'])-set(ext[0]['statuses']));(p/'unobserved-rows.json').write_text(json.dumps(outside,indent=2))
wall=sum(x['wall'] for x in main+ext+follow+[d8new]);(p/'timing.json').write_text(json.dumps({'setup_seconds':0,'nproc':5,'matrix_build_and_run_wall_seconds':wall,'coverage_wall_seconds':sum(x['wall'] for x in json.loads((p/'coverage-runs.json').read_text())),'whole_baseline_binary_seconds':90.111},indent=2))
print([(r['test'],r['defense'],r['unique_mutant']) for r in rows]);print('outside',len(outside),'wall',wall)
