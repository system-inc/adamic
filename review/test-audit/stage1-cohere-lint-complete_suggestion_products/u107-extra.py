import sys,os,pathlib,json,difflib,subprocess,time
sys.path.insert(0,'/workspace');from importlib.machinery import SourceFileLoader
r=SourceFileLoader('audit','/workspace/u107-run.py').load_module();p=r.p;root=r.root
rows=[('TestProduct_CompleteSuggestion family',['TestProduct_CompleteSuggestionGoOracle','TestProduct_CompleteSuggestionLowered','TestProduct_CompleteSuggestionMutantLowered','TestProduct_CompleteSuggestionNative']),('TestCompleteSuggestionSerialization_000',['TestCompleteSuggestionSerialization_000']),('TestCompleteSuggestionSerialization agreement family',['TestCompleteSuggestionSerialization_001','TestCompleteSuggestionSerialization_002','TestCompleteSuggestionSerialization_003','TestCompleteSuggestionSerialization']),('TestCompleteSuggestionSerialization mutant family',['TestCompleteSuggestionSerialization_004','TestCompleteSuggestionSerialization_005'])]+[(x,[x]) for x in ['TestCompleteSuggestionSerialization_Setup','TestCompleteSuggestionSerialization_PlantedFailure','TestCompilerCorpusSourcesExcludeOnlyNamedFolders','TestCompilerGuardBackend','TestChildCPUHangGuard','TestChildWallBackstop','TestChildCPUWaitGuard','TestEmittedJavaScriptMismatch_000']]
(p/'rows.json').write_text(json.dumps(rows,indent=2))
# These edits are fixed before observing their catches.
mods=[('S0','stage1/cohere/lint/complete_suggestion_serialization_shards_test.go','p.want = completeSuggestionExecute(ctx, t, "", p.oracle, "--manifest", p.path).output','p.want = nil',['TestCompleteSuggestionSerialization_000'],'setup-check'),('S1','stage1/cohere/lint/complete_suggestion_serialization_shards_test.go','completeSuggestionProductsReady = true','completeSuggestionProductsReady = false',['TestCompleteSuggestionSerialization_Setup'],'setup-check'),('S2','stage1/cohere/lint/complete_suggestion_serialization_shards_test.go','return os.WriteFile(filepath.Join(out, "lint.c"), []byte(native.C(lowered)), 0644)','return os.WriteFile(filepath.Join(out, "missing.c"), []byte(native.C(lowered)), 0644)',['TestProduct_CompleteSuggestionNative'],'setup-check'),('S3','stage1/cohere/lint/corpus_sources_test.go','slices.Contains(compilerCorpusFixtureFolders, directory)','!slices.Contains(compilerCorpusFixtureFolders, directory)',['TestCompilerCorpusSourcesExcludeOnlyNamedFolders'],'setup-check'),('W1','stage1/cohere/lint/complete_suggestion_serialization_shards_test.go','return caseID != plantedID || bytes.Equal(got, want)','return true',['TestCompleteSuggestionSerialization_PlantedFailure'],'witness'),('W2','stage1/cohere/lint/complete_suggestion_serialization_shards_test.go','if bytes.Equal(got, p.want) {','if true {',['TestCompleteSuggestionSerialization_004','TestCompleteSuggestionSerialization_005'],'witness'),('W3','internal/testguard/guard.go','return fmt.Errorf("%s: child CPU hang guard exceeded: CPU %s, budget %s: %w", label, cpu, budget, err)','return nil',['TestChildCPUHangGuard'],'witness'),('W4','internal/testguard/guard.go','return fmt.Errorf("%s: wall backstop exceeded: ceiling %s", label, ceiling)','return nil',['TestChildWallBackstop'],'witness'),('M5','internal/testguard/guard.go','time.NewTimer(ceiling)','time.NewTimer(budget)',['TestChildCPUWaitGuard'],'change option'),('M6','internal/testguard/guard.go','uint64((budget + time.Second - 1) / time.Second)','uint64(1)',['TestChildCPUWaitGuard'],'change constant')]
(p/'extra-plan.json').write_text(json.dumps(mods,indent=2))
if __name__=='__main__':
 with (p/'M6-witness-before.log').open('w') as log:subprocess.run(['go','run',str(p/'guard-probe.go')],cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=20)
 # Individual family timing uses the binary package elapsed line, includes every member, no summed timings.
 for name,members in rows:
  if name=='TestEmittedJavaScriptMismatch_000':continue
  pattern='^('+'|'.join(members)+')$'
  for i in range(3):
   label='timing-'+str(rows.index((name,members)))+'-'+str(i)
   if not (p/(label+'-time.json')).exists():r.run(label,pattern)
 for mid,file,old,new,members,kind in mods:
  if (p/(mid+'-time.json')).exists():continue
  f=root/file;before=f.read_text();assert old in before,(mid,old);after=before.replace(old,new)
  (p/(mid+'.diff')).write_text(''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
  try:
   f.write_text(after)
   # Verify every Go construction, witness and production diff compiles with vet.
   pkg='./internal/testguard/' if file.startswith('internal/') else './stage1/cohere/lint/'
   with (p/(mid+'-vet.log')).open('w') as log:v=subprocess.run(['go','vet',pkg],cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=90)
   assert v.returncode==0,mid
   e=os.environ.copy()
   if mid=='S2':e['ADAMIC_BUILD_CACHE_DIR']='/workspace/u107-cache/S2'
   r.run(mid,'^('+'|'.join(members)+')$',e)
   if mid=='M6':
    with (p/'M6-witness-after.log').open('w') as log:subprocess.run(['go','run',str(p/'guard-probe.go')],cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=20)
  finally:f.write_text(before)
