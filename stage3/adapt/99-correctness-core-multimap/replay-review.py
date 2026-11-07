"""Validate source-coordinate remapping and owned fingerprints, preserving raw reports."""
import difflib,json,pathlib,re,subprocess,os,sys
before_repo,after_repo,before_tree,after_tree,output=map(lambda x:pathlib.Path(x).resolve(),sys.argv[1:]);a=json.loads((output/'before.json').read_text());b=json.loads((output/'after.json').read_text());left={r['adapter']:r for r in a};right={r['adapter']:r for r in b};maps={};remaps=[]
def remap(file,line,column=None,repos=False):
 key=(file,repos)
 if key not in maps:
  roots=(before_repo,after_repo) if repos else (before_tree,after_tree)
  x=(roots[0]/file).read_text().splitlines();y=(roots[1]/file).read_text().splitlines();mapping={}
  for tag,i,j,k,l in difflib.SequenceMatcher(None,x,y,autojunk=False).get_opcodes():
   if tag=='equal':mapping.update({v+1:u+1 for u,v in zip(range(i,j),range(k,l))})
  maps[key]=(x,y,mapping)
 x,y,mapping=maps[key];assert line in mapping,(file,line,'not an unchanged source line');old=mapping[line];assert x[old-1]==y[line-1];remaps.append({'file':file,'afterLine':line,'beforeLine':old,'identicalCompleteSourceLine':True,'column':column});return old
with (output/'hash-proof.log').open('w') as log:subprocess.run(['node',str(after_repo/'stage3/adapt/99-correctness-core-multimap/replay-hash-proof.cjs'),str(before_tree),str(after_tree),str(output)],env=dict(os.environ,NODE_PATH=str(after_repo/'stage3/api/node_modules')),stdout=log,stderr=log,check=True)
hashes=json.loads((output/'hash-proof.json').read_text());allowed_hash={(r['file'],r['function']):(r['before'],r['after']) for r in hashes}
def normalize(x,parent_file=None):
 if isinstance(x,dict):
  file=x.get('file',parent_file);result={k:normalize(v,file) for k,v in x.items()}
  if isinstance(x.get('line'),int):assert file;result['line']=remap(file,x['line'],x.get('column'))
  if 'runtime_sha256' in x and (file,x.get('function')) in allowed_hash:
   old,new=allowed_hash[(file,x['function'])];assert x['runtime_sha256']==new;result['runtime_sha256']=old
  return result
 if isinstance(x,list):return [normalize(v,parent_file) for v in x]
 if isinstance(x,str):return re.sub(r'(src/[^:\s]+\.ts):(\d+):(\d+)',lambda m:m[1]+':'+str(remap(m[1],int(m[2]),int(m[3])))+':'+m[3],x)
 return x
new={'94-correctness-multimap','99-correctness-core-multimap'};assert right.keys()-left.keys()==new;changed=[]
for name,x in left.items():
 y=right[name];assert x['exit']==y['exit'] and x['changedFiles']==y['changedFiles'] and x['filesWritten']==y['filesWritten'],name
 if x==y:continue
 if name=='95-checker-axolotl-h':
  assert x['exit']!=0 and 'writer owner shape: createProgram' in x['stderr'];assert x['stdout']==y['stdout']==''
  file='stage3/adapt/95-checker-axolotl-h/adapt.cjs';normalized=re.sub(re.escape('REPO/'+file)+r':(\d+):(\d+)',lambda m:'REPO/'+file+':'+str(remap(file,int(m[1]),int(m[2]),True))+':'+m[2],y['stderr']);assert normalized==x['stderr'],'95h refusal changed beyond source-line positions';changed.append({'adapter':name,'only':'amended adapter source-line locations; identical refusal hashes and no writes'});continue
 assert x['stderr']==y['stderr'],name+' stderr changed';old=json.loads(x['stdout']);actual=json.loads(y['stdout']);default='src/compiler/sys.ts' if name=='92-tty-host-results' else None;normalized=normalize(actual,default);assert old==normalized,name+' has a non-owned outcome change';changed.append({'adapter':name,'only':'validated source locations or owned runtime fingerprint'})
for name in new:assert right[name]['exit']==0 and not right[name]['changedFiles'] and not right[name]['filesWritten']
result={'passed':True,'baselineAdapters':len(a),'adaptedAdapters':len(b),'rawReportsRetained':True,'sourceCoordinateProofs':len(remaps),'ownedRuntimeFingerprintChanges':hashes,'sourceDerivedReportDifferences':changed,'newAdaptersReplayWithoutWrites':sorted(new),'allOtherOutcomesIdentical':True,'requiresFullApplySourceEquality':True};(output/'summary.json').write_text(json.dumps(result,indent=2)+'\n');(output/'coordinate-proof.json').write_text(json.dumps(remaps,indent=2)+'\n');print(json.dumps({k:result[k] for k in ['passed','baselineAdapters','adaptedAdapters','sourceCoordinateProofs','allOtherOutcomesIdentical']}))
