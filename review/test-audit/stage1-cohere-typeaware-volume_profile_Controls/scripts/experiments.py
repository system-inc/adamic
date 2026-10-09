import pathlib,subprocess,os,json,difflib,time,re
repo=pathlib.Path('/workspace/adamic');r=repo/'review/test-audit/stage1-cohere-typeaware-volume_profile_Controls';env=os.environ.copy();env.update(ADAMIC_TYPESCRIPT_SOURCE='/tmp/u147/typescript',ADAMIC_VOLUME_REPOSITORY_MANIFEST='/tmp/u147/repository.manifest',ADAMIC_VOLUME_COMPILER_MANIFEST='/tmp/u147/compiler.manifest');results=[]
def replace_body(s,name,body):
 a=s.index('func '+name+'(');start=s.index('{',a);depth=1;b=start+1
 while depth:
  if s[b]=='{':depth+=1
  elif s[b]=='}':depth-=1
  b+=1
 return s[:start+1]+'\n'+body+'\n'+s[b-1:]
def experiment(mid,changes,pattern):
 originals={}
 diffs=[]
 for file,change in changes:
  p=repo/file;s=p.read_text();originals[p]=s;v=change(s);p.write_text(v);diffs.extend(difflib.unified_diff(s.splitlines(True),v.splitlines(True),fromfile='a/'+file,tofile='b/'+file))
 (r/(mid+'.diff')).write_text(''.join(diffs))
 try:
  with (r/'logs'/(mid+'-vet.log')).open('w') as f:rc=subprocess.run(['go','vet','./stage1/cohere/typeaware/'],stdout=f,stderr=subprocess.STDOUT,env=env).returncode
  if rc:raise RuntimeError('vet '+mid)
  if mid=='W3':
   with (r/'logs'/(mid+'-bridge-vet.log')).open('w') as f:subprocess.run(['go','vet','./bridge/tsgo/checker/'],stdout=f,stderr=subprocess.STDOUT,env=env)
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/typeaware/','-run',pattern];start=time.monotonic()
  with (r/'logs'/(mid+'.log')).open('w') as f:rc=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env=env).returncode
  results.append(dict(id=mid,pattern=pattern,wall=time.monotonic()-start,returncode=rc));(r/'experiment-runs.json').write_text(json.dumps(results,indent=2)+'\n')
 finally:
  for p,s in originals.items():p.write_text(s)
p='stage1/cohere/typeaware/'
# Construction comparisons; none is a production mutant.
experiment('K1',[(p+'volume_profile_Controls_test.go',lambda s:s.replace('const testVolumeProfileControlsShards = 2','const testVolumeProfileControlsShards = 3'))],'^TestVolumeProfileControlsUnion$')
experiment('K2',[(p+'volume_profile_Controls_test.go',lambda s:s.replace('const testShadowIndexMissingBindingShards = 1','const testShadowIndexMissingBindingShards = 2'))],'^TestShadowIndexMissingBindingUnion$')
experiment('K3',[(p+'volume_profile_Controls_test.go',lambda s:s.replace('func TestVolumeProfileControls_001(', 'func TestVolumeProfileControls_999('))],'^TestVolumeProfilePartition$')
experiment('K4',[(p+'volume_type_symbol_test.go',lambda s:s.replace('ctx, cancel := context.WithTimeout(context.Background(), limit)','ctx, cancel := context.WithCancel(context.Background())'))],'^TestVolumeTypeSymbolCommandDeadline$')
experiment('W1',[(p+'volume_type_symbol_test.go',lambda s:s.replace('if bytes.Equal(observed.stdout, truth.stdout) {','if true {'))],'^TestVolumeTypeSymbol(PlantedSurvivor|Union|_000)$')
experiment('W2',[(p+'volume_profile_Mutants_test.go',lambda s:s.replace('bytes.Equal(got.stdout, truth.stdout)','true').replace('\n\t"bytes"',''))],'^TestVolumeProfileMutants(Union|_[0-9]+)$')
# Removing the whole external checker test loosens every guarded comparison, no unused imports.
experiment('W3',[('bridge/tsgo/checker/index_test.go',lambda s:'package checker\nimport "testing"\nfunc TestExactIndexMatchesCompilerNodes(t *testing.T) {}\n')],'^TestVolumeProfileOverlays(Union|_[0-9]+)$')

experiment('K5',[(p+'volume_profile_Controls_test.go',lambda s:s.replace('"typeaware volume lowered"','"typeaware volume lowered broken recipe"').replace('load.Load([]string{filepath.Join(h.repository, "stage1/cohere/typeaware/volume_suite.ts")})','load.Load([]string{filepath.Join(h.repository, "stage1/cohere/typeaware/missing-entry.ts")})'))],'^TestVolumeProfileControlsLower$')
experiment('K6',[(p+'volume_profile_Corpora_test.go',lambda s:s.replace('[]string{products.Binary, products.Asan, products.Oracle}','[]string{products.Binary, products.Asan, ""}'))],'^TestVolumeProfileCorpora_Setup$')
experiment('K7',[(p+'volume_config_guard_self_prepare_test.go',lambda s:s.replace('volumeGuardShared.products = buildVolumeConfigGuardProducts(t)','volumeGuardShared.products = volumeGuardProducts{}'))],'^TestVolumeConfigGuardAndMutant$')
experiment('PsetupLower',[(p+'volume_profile_Controls_test.go',lambda s:replace_body(s,'volumeProfileControlsLowered','return ""').replace('\n\t"github.com/system-inc/adamic/internal/load"','').replace('\n\t"github.com/system-inc/adamic/internal/lower"','').replace('\n\t"github.com/system-inc/adamic/internal/native"',''))],'^TestVolumeProfileControlsLower$')
experiment('PsetupCorpus',[(p+'volume_profile_Corpora_test.go',lambda s:replace_body(s,'volumeProfileCorporaPrepare','return volumeProfileCorporaProductPaths{}'))],'^TestVolumeProfileCorpora_Setup$')
experiment('PsetupConfig',[(p+'volume_config_guard_self_prepare_test.go',lambda s:replace_body(s,'volumeGuardProductsFor','return volumeGuardProducts{}'))],'^TestVolumeConfigGuardAndMutant$')

experiment('W4',[(p+'volume_type_symbol_test.go',lambda s:s.replace('if bytes.Equal(observed.stdout, truth.stdout) {','if false {'))],'^TestVolumeTypeSymbolPlantedSurvivor$')
