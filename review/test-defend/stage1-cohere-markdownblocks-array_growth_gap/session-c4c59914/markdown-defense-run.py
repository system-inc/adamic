import pathlib,subprocess,json,os,time
R=pathlib.Path('/workspace/adamic');P=R/'review/test-defend/stage1-cohere-markdownblocks-array_growth_gap/session-c4c59914';pkg='./stage1/cohere/markdownblocks/'
env=os.environ.copy();env['TMPDIR']='/workspace/defend-markdown-tmp'
records=[]
def run(name,selector,cover=False,mutant=None):
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s',pkg,'-run',selector]
 if cover:cmd+=['-coverpkg=github.com/system-inc/adamic/stage1/cohere/markdownblocks','-coverprofile='+str(P/(name+'.cover'))]
 e=env.copy()
 if mutant:e['ADAMIC_BUILD_CACHE_DIR']='/workspace/defend-markdown-cache/'+mutant
 start=time.monotonic()
 with (P/(name+'.log')).open('w') as f:rc=subprocess.run(cmd,cwd=R,env=e,stdout=f,stderr=subprocess.STDOUT).returncode
 ev=[]
 for l in (P/(name+'.log')).read_text().splitlines():
  try:ev.append(json.loads(l))
  except:pass
 states={x['Test']:x['Action'] for x in ev if '/' not in x.get('Test','') and x.get('Test') and x.get('Action') in ['pass','skip','fail']}
 cooked=any('test timed out after' in x.get('Output','') for x in ev)
 assertion=[x.get('Output','').strip() for x in ev if any(y in x.get('Output','') for y in ['mismatch','difference','survived','exit 70','fatal error:']) and x.get('Test')]
 rec=dict(name=name,command=cmd,cache=e.get('ADAMIC_BUILD_CACHE_DIR'),seconds=time.monotonic()-start,exit=rc,cooked=cooked,states=states,assertions=assertion)
 records.append(rec);(P/'runs.json').write_text(json.dumps(records,indent=2));print(name,round(rec['seconds'],2),rc,cooked,states,flush=True);return rec
if __name__=='__main__':
 for n,sel in [('coverage-ast','^TestMarkdownASTPreprocessing$'),('coverage-chunks','^TestMicromarkInputChunks$'),('coverage-decode','^TestMarkdownSourceDecoding$'),('coverage-probes','^TestParserRepresentationProbes$'),('coverage-events','^TestTokenizerEvents(Union|_[0-9]{3})$')]:
  r=run(n,sel,cover=True)
  if r['exit'] and not r['cooked']:raise RuntimeError('RED baseline '+n)
