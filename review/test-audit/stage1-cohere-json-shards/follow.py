exec(open('/tmp/u104-audit.py').read().split('try:\n')[0])
try:
 for id,f,name in [('W2','top_upstream_proof_test.go','jsonUpstreamBlockCheck'),('W3','upstream_parity_split_test.go','upstreamParityBlockCheck')]:
  diff(id,f,body(original[f],name,'return nil'))
  with (p/(id+'-vet.log')).open('w') as log:subprocess.run(['go','vet','./stage1/cohere/json/'],cwd=root,stdout=log,stderr=subprocess.STDOUT,check=True)
  run(id,'^(TestJSONPortShardDisagreement|TestJSONUpstreamShardDisagreement|TestUpstreamRepositoryCorpusParityShardProof)$');(src/f).write_text(original[f])
 for id,f,a,b in [('S1','shards_test.go','end-start < 16','end-start < 15'),('S2','shards_test.go','make([]nativeChunk, count)','make([]nativeChunk, count-1)'),('S3','upstream_parity_split_test.go','SplitShards = 16','SplitShards = 15'),('S4','shards_test.go','start = end\n','start = end + 1\n'),('S5','upstream_parity_split_test.go','filepath.Join(dir, "oracle")','filepath.Join(dir, "missing-oracle")')]:
  assert original[f].count(a)==1;diff(id,f,original[f].replace(a,b));
  with (p/(id+'-vet.log')).open('w') as log:subprocess.run(['go','vet','./stage1/cohere/json/'],cwd=root,stdout=log,stderr=subprocess.STDOUT,check=True)
  run(id,'^(TestJSONPortShardUnion|TestJSONHashShardsStayStable|TestPortMatchesGoCohereUnion|TestUpstreamRepositoryCorpusParityUnion|TestUpstreamRepositoryCorpusParity_Setup|TestUpstreamRepositoryCorpusParity|TestProduct_JSONUpstreamOracle)$',cache='/tmp/u104/cache/'+id if id=='S5' else None);(src/f).write_text(original[f])
finally:
 for f,s in original.items():(src/f).write_text(s)
