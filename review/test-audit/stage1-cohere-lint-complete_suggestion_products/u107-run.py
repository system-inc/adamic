import pathlib,json,subprocess,time,os,difflib
root=pathlib.Path('/workspace/adamic');p=root/'review/test-audit/stage1-cohere-lint-complete_suggestion_products'
def run(label,pattern,env=None):
 args=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/','-run',pattern];s=time.monotonic()
 with (p/(label+'.log')).open('w') as f:r=subprocess.run(args,cwd=root,stdout=f,stderr=subprocess.STDOUT,env=env)
 d={'label':label,'command':' '.join(args),'seconds':time.monotonic()-s,'exit':r.returncode}
 (p/(label+'-time.json')).write_text(json.dumps(d));print(label,d,flush=True)
 return d
if __name__=='__main__':
 assert json.loads((p/'narrow-baseline-time.json').read_text())['exit']==0
 pattern=json.loads((p/'narrow-baseline-time.json').read_text())['pattern']
 # Exclude the expensive self-spawning witness from production matrices. Its comparison edit is audited separately.
 pattern=pattern.replace('|TestCompleteSuggestionSerialization_PlantedFailure','')
 for m in json.loads((p/'plan.json').read_text()):
  f=root/m['file'];before=f.read_text();after=before.replace(m['old'],m['new']);assert after!=before
  (p/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
  try:
   f.write_text(after);e=os.environ.copy();e['ADAMIC_BUILD_CACHE_DIR']='/workspace/u107-cache/'+m['id'];e['ADAMIC_BUILD_CACHE']='on';run(m['id'],pattern,e)
  finally:f.write_text(before)
