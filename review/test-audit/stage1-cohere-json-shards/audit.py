import pathlib,subprocess,json,time,difflib,re,os
root=pathlib.Path('/workspace/adamic'); p=root/'review/test-audit/stage1-cohere-json-shards'; src=root/'stage1/cohere/json'
env=os.environ.copy(); env.update(ADAMIC_JSON_PRETTIER='/tmp/u104/prettier',ADAMIC_JSON_BENCH='1')
rows=json.loads((p/'production-matrix-rows.txt').read_text()); selector='^('+'|'.join(rows)+')$'
original={f:(src/f).read_text() for f in ['doc.ts','formatter.ts','port_test.go','top_upstream_proof_test.go','upstream_parity_split_test.go','shards_test.go']}
def body(s,name,replacement):
 start=s.index('func '+name+'('); opening=s.index('{',start); depth=1; i=opening+1
 # Go checker bodies contain brace strings; brace-count works here because quoted braces balanced.
 while depth:
  if s[i]=='{': depth+=1
  elif s[i]=='}': depth-=1
  i+=1
 return s[:opening+1]+'\n'+replacement+'\n'+s[i-1:]
def run(id,sel=selector,cache=None):
 e=env.copy()
 if cache:e['ADAMIC_BUILD_CACHE_DIR']=cache
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/json/','-run',sel]
 began=time.monotonic()
 with (p/(id+'.log')).open('w') as f:r=subprocess.run(cmd,cwd=root,env=e,stdout=f,stderr=subprocess.STDOUT)
 result={'id':id,'command':cmd,'env':{k:e[k] for k in ['ADAMIC_JSON_PRETTIER','ADAMIC_JSON_BENCH']+(['ADAMIC_BUILD_CACHE_DIR'] if cache else [])},'wall_seconds':time.monotonic()-began,'exit':r.returncode}
 with (p/'runs.txt').open('a') as f:f.write(json.dumps(result)+'\n')
 print(id,result['exit'],round(result['wall_seconds'],3),flush=True)
def diff(id,f,s):
 (src/f).write_text(s)
 if f.endswith('.go'):subprocess.run(['gofmt','-w',str(src/f)],check=True);s=(src/f).read_text()
 (p/(id+'.diff')).write_text(''.join(difflib.unified_diff(original[f].splitlines(True),s.splitlines(True),fromfile='a/stage1/cohere/json/'+f,tofile='b/stage1/cohere/json/'+f)))
try:
 for id,f,a,b in [('M1','doc.ts','command.indent + 2','command.indent + 1'),('M2','formatter.ts','value = `0${value}`;','value = `1${value}`;'),('M3','formatter.ts',"base === 'package.json'","base === 'other-package.json'")]:
  assert original[f].count(a)==1;diff(id,f,original[f].replace(a,b));run(id,cache='/tmp/u104/cache/'+id);(src/f).write_text(original[f])
 f='formatter.ts'; s=original[f];start=s.index('export function format(');s=s[:s.index('{',start)+1]+"\n    return '';\n}\n";diff('P1',f,s);run('P1',cache='/tmp/u104/cache/P1');(src/f).write_text(original[f])
 for id,f,name in [('W1','port_test.go','comparisonError'),('W2','top_upstream_proof_test.go','jsonUpstreamBlockCheck'),('W3','upstream_parity_split_test.go','upstreamParityBlockCheck')]:
  diff(id,f,body(original[f],name,'return nil'))
  # Remove newly unused imports so every standalone diff compiles.
  if id=='W2':s=(src/f).read_text().replace('\n\t"fmt"','');diff(id,f,s)
  with (p/(id+'-vet.log')).open('w') as log:subprocess.run(['go','vet','./stage1/cohere/json/'],cwd=root,stdout=log,stderr=subprocess.STDOUT,check=True)
  run(id,'^(TestJSONPortShardDisagreement|TestJSONUpstreamShardDisagreement|TestUpstreamRepositoryCorpusParityShardProof)$');(src/f).write_text(original[f])
 for id,f,a,b in [('S1','shards_test.go','end-start < 16','end-start < 15'),('S2','shards_test.go','make([]nativeChunk, count)','make([]nativeChunk, count-1)'),('S3','upstream_parity_split_test.go','SplitShards = 16','SplitShards = 15')]:
  assert original[f].count(a)==1;diff(id,f,original[f].replace(a,b));
  with (p/(id+'-vet.log')).open('w') as log:subprocess.run(['go','vet','./stage1/cohere/json/'],cwd=root,stdout=log,stderr=subprocess.STDOUT,check=True)
  run(id,'^(TestJSONPortShardUnion|TestJSONHashShardsStayStable|TestPortMatchesGoCohereUnion|TestUpstreamRepositoryCorpusParityUnion|TestUpstreamRepositoryCorpusParity_Setup|TestUpstreamRepositoryCorpusParity|TestProduct_JSONUpstreamOracle)$');(src/f).write_text(original[f])
finally:
 for f,s in original.items():(src/f).write_text(s)
