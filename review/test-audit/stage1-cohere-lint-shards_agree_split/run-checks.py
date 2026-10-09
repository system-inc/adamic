import pathlib,json,subprocess,time,os,difflib
R=pathlib.Path.cwd();E=R/'review/test-audit/stage1-cohere-lint-shards_agree_split';D=E/'diffs';D.mkdir(exist_ok=True);W=pathlib.Path('/tmp/u112/weak');W.mkdir(exist_ok=True);env=os.environ.copy();env['ADAMIC_TYPESCRIPT_SOURCE']='/tmp/u112/TypeScript-050880ce59e30b356b686bd3144efe24f875ebc8';res=[]
def body(s,sig,repl):
 a=s.index('{',s.index(sig));i=a+1;d=1
 while d:
  if s[i]=='{':d+=1
  if s[i]=='}':d-=1
  i+=1
 return s[:a+1]+'\n'+repl+'\n'+s[i-1:]
def overlay(id,changes):
 replaces={};diffs=''
 for name,s in changes.items():
  p=R/'stage1/cohere/lint'/name;old=p.read_text();q=W/(id+'-'+name);q.write_text(s);subprocess.run(['gofmt','-w',str(q)],check=True);s=q.read_text();replaces[str(p)]=str(q);diffs+=''.join(difflib.unified_diff(old.splitlines(True),s.splitlines(True),fromfile='a/stage1/cohere/lint/'+name,tofile='b/stage1/cohere/lint/'+name))
 path=W/(id+'.json');path.write_text(json.dumps({'Replace':replaces}));(D/(id+'.diff')).write_text(diffs);return path
name='lint_test.go';s=(R/'stage1/cohere/lint'/name).read_text();weak=overlay('W1',{name:body(s,'func difference(', '\treturn ""')})
# Construction probes: replace product fetches with their empty directory.
name='witness_script_kind_products_test.go';s=(R/'stage1/cohere/lint'/name).read_text()
for f in ['witnessScriptKindLoweredProduct','witnessScriptKindNativeProduct','witnessScriptKindOracleProduct']:s=body(s,'func '+f+'(', '\treturn ""')
# Only product wrapper functions remain; remove implementation-only imports.
s=s[:s.index('import (')]+'import "testing"\n'+s[s.index('// Build-phase'):]
products=overlay('S1',{name:s})
name='suggestion_alongside_shards_test.go';s=(R/'stage1/cohere/lint'/name).read_text();setup=overlay('S2',{name:s.replace('suggestionAlongside.ready = true','suggestionAlongside.ready = false')})
name='witness_script_kind_shards_test.go';s=(R/'stage1/cohere/lint'/name).read_text();kindsetup=overlay('S3',{name:s.replace('filepath.Ext(strings.TrimSuffix(path, ".txt"))', '".ts"')});standalone=overlay('S4',{name:s.replace('"-test.run=^TestWitnessScriptKind_000$"','"-test.run=^NoSuchLeaf$"')})
# Cold setup-required witness: a selected leaf now skips construction.
name='suggestion_alongside_shards_test.go';s=(R/'stage1/cohere/lint'/name).read_text();required=overlay('S5',{name:body(s,'func suggestionAlongsideReady(', '\tt.Helper()')})
checks=[('W1-suggestion','^TestSuggestionAlongsideAutomaticFixPlantedFailure$',weak),('W1-kind','^TestWitnessScriptKindPlantedFailure$',weak),('S1-products','^TestProduct_WitnessScriptKind',products),('S2-suggestion-setup','^TestSuggestionAlongsideAutomaticFix_Setup$',setup),('S3-kind-setup','^TestWitnessScriptKind_Setup$',kindsetup),('S4-kind-standalone','^TestWitnessScriptKindStandalone$',standalone),('S5-required','^TestSuggestionAlongsideAutomaticFixSetupIsRequired$',required)]
for id,rx,path in checks:
 cmd=['timeout','120','go','test','-overlay='+str(path),'-json','-count=1','-timeout','90s','./stage1/cohere/lint/','-run',rx];t=time.monotonic()
 with (E/(id+'.log')).open('w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
 events=[]
 for l in (E/(id+'.log')).read_text().splitlines():
  try:events.append(json.loads(l))
  except:pass
 res.append({'id':id,'command':cmd,'exit':r.returncode,'wall':time.monotonic()-t,'events':[e for e in events if e.get('Action') in ['pass','fail','skip']],'timeout':any('test timed out' in e.get('Output','') for e in events)});(E/'checks.json').write_text(json.dumps(res,indent=2));print(id,r.returncode,flush=True)
 with (E/(id+'-vet.log')).open('w') as f:r=subprocess.run(['go','vet','-overlay='+str(path),'./stage1/cohere/lint/'],env=env,stdout=f,stderr=subprocess.STDOUT)
 print(id,'vet',r.returncode,flush=True)
