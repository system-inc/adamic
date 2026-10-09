import pathlib,json,subprocess,os,time,difflib
root=pathlib.Path('/workspace/adamic'); p=root/'review/test-defend/stage1-cohere-css-composition_shards'
plan=json.loads((p/'plan.json').read_text()); results=[]
env=os.environ.copy(); env.pop('NODE_V8_COVERAGE',None)
base={m['file']:(root/m['file']).read_text() for m in plan}
def run(ident,group,pattern,coverage=False):
 e=env.copy();e['ADAMIC_BUILD_CACHE_DIR']='/tmp/defend-css/cache/'+ident
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/css/','-run',pattern]
 log=ident+'-'+group+'.log';t=time.monotonic()
 with (p/log).open('w') as out:q=subprocess.run(cmd,cwd=root,env=e,stdout=out,stderr=subprocess.STDOUT)
 ev=[]
 for line in (p/log).read_text().splitlines():
  try:ev.append(json.loads(line))
  except:pass
 rr=dict(mutant=ident,group=group,command=cmd,cache=e['ADAMIC_BUILD_CACHE_DIR'],wall=round(time.monotonic()-t,3),exit=q.returncode,failed=[x['Test'] for x in ev if x['Action']=='fail' and x.get('Test')],passed=[x['Test'] for x in ev if x['Action']=='pass' and x.get('Test')],skipped=[x['Test'] for x in ev if x['Action']=='skip' and x.get('Test')],cooked=any('test timed out' in x.get('Output','') for x in ev),outputs=[x['Output'].strip() for x in ev if x.get('Test')=='TestCSSThroughput' and ('css_test.go:' in x.get('Output',''))])
 results.append(rr);(p/'results.json').write_text(json.dumps(results,indent=2)+'\n');print(ident,group,rr['wall'],q.returncode,rr['failed'],flush=True);return rr
# Complete reached parser family, with its product constructors, was not reached before baseline timeout.
run('clean','parser-products','^TestProduct_CSSParser')
run('clean','parser-family','^TestThePortParsesAsGoCohereDoes_[0-9]+$')
# Proceed only if reached clean results have no assertion failures or setup errors.
if any(x['failed'] or (x['exit']!=0 and not x['cooked']) for x in results):raise SystemExit('red reached baseline')
try:
 for m in plan[:2]:
  f=m['file'];assert base[f].count(m['old'])==1
  new=base[f].replace(m['old'],m['new'],1);(root/f).write_text(new)
  (p/(m['mutant']+'.diff')).write_text(''.join(difflib.unified_diff(base[f].splitlines(True),new.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
  run(m['mutant'],'throughput','^TestCSSThroughput$')
  run(m['mutant'],'composition-family','^TestCompositionMatchesGo_[0-9]+$')
  if m['mutant']=='D1':
   run('D1','parser-products','^TestProduct_CSSParser')
   run('D1','parser-family','^TestThePortParsesAsGoCohereDoes_[0-9]+$')
   run('D1','parser-construction','^(TestThePortParsesAsGoCohereDoesUnion|TestCSSParserPlantedDisagreement)$')
  (root/f).write_text(base[f])
finally:
 for f,s in base.items():(root/f).write_text(s)
