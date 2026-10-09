import pathlib,json,re,statistics,subprocess,difflib
root=pathlib.Path('/workspace/adamic');p=root/'review/test-audit/stage1-cohere-json-shards';pkg='github.com/system-inc/adamic/stage1/cohere/json'
def events(name):return [json.loads(s) for s in (p/(name+'.log')).read_text().splitlines() if s.startswith('{')]
ids=['M1','M2','M3','P1','W1','W2','W3','S1','S2','S3','S4','S5'];ev={i:events(i) for i in ids};members={}
names=[s for s in (p/'list.log').read_text().splitlines() if s.startswith('Test')]
members['R1']=['TestJSONPortShardUnion'];members['R2']=['TestJSONPortShardDisagreement'];members['R3']=['TestJSONHashShardsStayStable'];members['R4']=[n for n in names if re.fullmatch('TestPortMatchesGoCohere_[0-9]{4}',n) and int(n[-4:])<2048]+['TestPortMatchesGoCohereUnion'];members['R5']=[n for n in names if re.fullmatch('TestPortMatchesGoCohere_[0-9]{4}',n) and int(n[-4:])>=2048];members['R6']=[n for n in names if re.fullmatch('TestUpstreamRepositoryCorpusParity_[0-9]{4}',n)]+['TestUpstreamRepositoryCorpusParityUnion'];members['R7']=['TestJSONUpstreamShardDisagreement'];members['R8']=['TestUpstreamRepositoryCorpusParity_Setup','TestUpstreamRepositoryCorpusParity'];members['R9']=['TestUpstreamRepositoryCorpusParityShardProof'];members['R10']=[n for n in names if re.fullmatch('TestUpstreamRepositoryCorpusParity_[0-9]{3}',n)];members['R11']=['TestProduct_JSONUpstreamOracle']
assert sum(map(len,members.values()))==10266
(p/'family-members.txt').write_text(json.dumps(members,indent=2)+'\n')
lookup={n:r for r,ns in members.items() for n in ns};matrix={i:sorted({lookup[e['Test']] for e in es if e['Action']=='fail' and e.get('Test') in lookup}) for i,es in ev.items()}
(p/'kill-matrix.txt').write_text(json.dumps({'production':{i:matrix[i] for i in ids[:3]},'probes':{'P1':matrix['P1']},'witness':{i:matrix[i] for i in ids[4:7]},'construction':{i:matrix[i] for i in ids[7:]},'failed_members':{i:sorted({e['Test'] for e in es if e['Action']=='fail' and e.get('Test') in lookup}) for i,es in ev.items()}},indent=2)+'\n')
# Parse original/new line correspondence from standalone diffs, without modifying sources.
def mapping(id):
 maps={};lines=(p/(id+'.diff')).read_text().splitlines();f=None;old=new=0
 for line in lines:
  if line.startswith('+++ b/'):f=line[6:];maps[f]={}
  elif line.startswith('@@'):
   m=re.match(r'@@ -(\d+)(?:,\d+)? \+(\d+)',line);old,new=map(int,m.groups())
  elif f and line.startswith(' '):maps[f][new]=old;old+=1;new+=1
  elif f and line.startswith('-') and not line.startswith('---'):old+=1
  elif f and line.startswith('+') and not line.startswith('+++'):new+=1
 return maps
# Errors in unchanged blocks after changed function are offset by final hunk delta.
def normalized(id,text):
 d=(p/(id+'.diff')).read_text();matches=list(re.finditer(r'@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@',d));file=re.search(r'\+\+\+ b/(.+)',d).group(1);base=file.rsplit('/',1)[-1];mp=mapping(id).get(file,{})
 def fix(m):
  n=int(m.group(1))
  if n in mp: return base+':'+str(mp[n])
  offset=0
  for h in matches:
   old,oc,new,nc=int(h[1]),int(h[2] or 1),int(h[3]),int(h[4] or 1)
   if n>=new+nc:offset=(old+oc)-(new+nc)
  return base+':'+str(n+offset)
 return re.sub(re.escape(base)+r':(\d+)',fix,text)
def fail(id,row):
 failed={e['Test'] for e in ev[id] if e['Action']=='fail' and e.get('Test') in members[row]}
 for e in ev[id]:
  if e.get('Test') in failed and e.get('OutputType')=='error':return normalized(id,e['Output'].strip())
 return None
secs={}
for r in members:
 vals=[]
 for k in range(1,4):
  es=events(f'{r}-time{k}');found=[float(m.group(1)) for e in es if (m:=re.search(r'^(?:ok\s+|FAIL\s+)'+re.escape(pkg)+r'\s+(\d+(?:\.\d+)?)s',e.get('Output','')))]
  assert found and all(e['Action']!='fail' for e in es),r;vals.append(found[-1])
 secs[r]=statistics.median(vals)
(p/'median-seconds.txt').write_text(json.dumps(secs,indent=2)+'\n')
rowtitles=['TestJSONPortShardUnion','TestJSONPortShardDisagreement','TestJSONHashShardsStayStable','TestPortMatchesGoCohere agreement family','TestPortMatchesGoCohere benchmark family','TestUpstreamRepositoryCorpusParity top family','TestJSONUpstreamShardDisagreement','TestUpstreamRepositoryCorpusParity preparation family','TestUpstreamRepositoryCorpusParityShardProof','TestUpstreamRepositoryCorpusParity legacy family','TestProduct_JSONUpstreamOracle']
files=['shards_test.go','shards_test.go','shards_test.go','top_portmatchesgocohere_test.go + top_shared_test.go','top_portmatchesgocohere_test.go','top_upstreamrepositorycorpusparity_test.go + top_shared_test.go','top_upstream_proof_test.go','upstream_parity_split_test.go','upstream_parity_split_test.go','upstream_parity_split_test.go','upstream_parity_split_test.go']
verdicts=['setup-check','witness','setup-check','subsumed','subsumed','cannot-judge','witness','setup-check','witness','cannot-judge','untrue'];deciders=['S4','W1','S2','M3','M3',None,'W2','S3','W3',None,'S5']
oracles=['Self-written exact-union/selector invariants. Does not assert the 16-case bound.','Self-written planted protocol mismatch and ownership invariant.','Self-written fixed bucket count, case conservation and ownership stability.','Runs Go cohere and compares complete output/error protocol against source Node, release native, JavaScript backend, ASan/UBSan and LeakSanitizer.','Runs Go cohere and compares complete protocol against release native and source Node.','Runs Go cohere and Prettier 3.9.6; expected disagreements are our checked-in report. Both-error acceptance uses sameAnswer and ignores differing diagnostic text.','Self-written planted difference must be detected by jsonUpstreamBlockCheck.','Self-written declaration census and corpus construction checks; also builds an oracle without executing it.','Self-written planted difference must be detected by upstreamParityBlockCheck; also detects missing corpus case.','Runs Go cohere and Prettier 3.9.6; expected disagreements are our checked-in report. Both-error acceptance ignores diagnostic text.','Build-exit and nonempty product-directory checks only; does not verify the expected oracle executable exists.']
result=[]
prodrows=json.loads((p/'production-matrix-rows.txt').read_text());timing=json.loads((p/'timing-selectors.txt').read_text())
for index,(r,ns) in enumerate(members.items()):
 kind='external-run' if r in ['R4','R5'] else ['external-run','self'] if r in ['R6','R10'] else 'self';kills=[i for i in ids[:3] if r in matrix[i]];probe=['P1'] if r in ['R4','R5'] and r in matrix['P1'] else [];dec=deciders[index];line=fail(dec,r) if dec else None;sub=['R5'] if r=='R4' else ['R4'] if r=='R5' else [];sr=sub[0] if sub else None
 mrows=prodrows if r in ['R4','R5'] else ['TestJSONPortShardDisagreement','TestJSONUpstreamShardDisagreement','TestUpstreamRepositoryCorpusParityShardProof'] if r in ['R2','R7','R9'] else ['TestJSONPortShardUnion','TestJSONHashShardsStayStable','TestPortMatchesGoCohereUnion','TestUpstreamRepositoryCorpusParityUnion','TestUpstreamRepositoryCorpusParity_Setup','TestUpstreamRepositoryCorpusParity','TestProduct_JSONUpstreamOracle'] if r in ['R1','R3','R8','R11'] else []
 o={'test':rowtitles[index],'package':pkg,'file':'stage1/cohere/json/'+files[index],'seconds':None if r in ['R4','R5','R6'] else secs[r],'oracle':oracles[index],'oracle_kind':kind,'kills':kills,'unique_kills':[],'last_proven_fail':dec+' '+line if line else None,'verdict':verdicts[index],'subsumed_by':[rowtitles[int(x[1:])-1] for x in sub],'mutants_in_matrix':3 if r in ['R4','R5','R2','R7','R9'] else 5 if r in ['R1','R3','R8','R11'] else 0,'probe_kills':probe,'subsumer_seconds':None,'vacuous':False if probe else None,'bounded':True,'matrix_rows':mrows,'evidence':{'command':next((json.loads(x) for x in (p/'runs.txt').read_text().splitlines() if json.loads(x)['id']==dec), {'command':['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/json/','-run',(p/'selector.txt').read_text()], 'env':{'ADAMIC_JSON_PRETTIER':'/tmp/u104/prettier','ADAMIC_JSON_BENCH':'1'}}),'failing_output':line},'members_file':'family-members.txt#'+r,'member_count':len(ns),'timing_runs':3,'timing_scope':'bounded subset of family' if r in ['R4','R5','R6'] else 'entire row','bounded_seconds':secs[r] if r in ['R4','R5','R6'] else None,'construction_kills':[i for i in ids[7:] if r in matrix[i]],'witness_kills':[i for i in ids[4:7] if r in matrix[i]],'vacuous_subcases':[]}
 if sr:o['subsumer_bounded_seconds']=secs[sr];o['subsumption_basis']='3 port mutants; only the selected members were replayed'
 if r in ['R6','R10']:o['cannot_judge_reason']='These rows run Go cohere and Prettier, not the Adamic port. Mutating either external oracle is prohibited; no production mutant was judged for these rows.'
 if r=='R11':o['evidence']['passing_output']='S5: --- PASS: TestProduct_JSONUpstreamOracle; expected oracle absent, missing-oracle present (S5-artifact-witness.txt).'
 result.append(o)
(p/'results.json.txt').write_text(json.dumps(result,indent=2)+'\n')
(p/'mutant-table.txt').write_text(json.dumps([{'id':i,'change':(p/(i+'.diff')).read_text(),'failed_rows':[rowtitles[int(r[1:])-1] for r in matrix[i]],'kind':'production' if i.startswith('M') else 'probe' if i.startswith('P') else 'witness' if i.startswith('W') else 'construction'} for i in ids],indent=2)+'\n')
print(json.dumps({'medians':secs,'matrix':matrix},indent=2))
