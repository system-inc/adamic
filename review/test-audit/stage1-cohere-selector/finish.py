import pathlib,subprocess,time,json,os,shlex
root=pathlib.Path('/workspace/adamic');pkg='stage1/cohere/selector';p=root/pkg;out=root/'review/test-audit/stage1-cohere-selector';files=list(p.glob('*.ts'))+list(p.glob('*_test.go'))+[p/'testdata/library.mjs',p/'testdata/nontermination.mjs',root/'internal/lower/object.go',root/'internal/lower/lower.go'];base={str(f.relative_to(root)):f.read_text() for f in files};checks=[];env=os.environ.copy();env.update({'ADAMIC_SELECTOR_LIBRARY':'/tmp/u134/library','ADAMIC_SELECTOR_BENCH':'1'})
def restore():
 for f,s in base.items():(root/f).write_text(s)
try:
 # A complete entry replacement requires dropping its now-unused imports.
 f=root/'internal/lower/lower.go';s=f.read_text();start=s.index('func Lower(');body=s.index('{',start);end=s.index('\n}\n',body)+2;s=s[:body+1]+'\n return nil, nil\n}'+s[end:];s=s.replace('\n\t"fmt"','').replace('\n\t"path/filepath"','');f.write_text(s);(out/'P2.diff').write_bytes(subprocess.check_output(['git','diff','--','internal/lower/lower.go'],cwd=root))
 for row in ['.','TestEachGapStandsWhereGapsMdSaysItDoes','TestThePortParsesAsGoCohereDoes','TestSelectorThroughput','TestCorpusKeepsEveryParseableFile','TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars']:
  id='P2' if row=='.' else 'P2-'+row;cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./'+pkg+'/','-run','.' if row=='.' else '^'+row+'$'];e=env.copy();e['ADAMIC_BUILD_CACHE_DIR']='/tmp/u134/cache/'+id+'-valid';started=time.monotonic()
  with (out/(id+'.log')).open('w')as fd:r=subprocess.run(cmd,cwd=root,env=e,stdout=fd,stderr=subprocess.STDOUT)
  checks.append({'id':id,'kind':'corrected probe','command':shlex.join(cmd),'exit':r.returncode,'seconds':time.monotonic()-started});print(checks[-1],flush=True)
 restore()
 for id in ['M1','M2','M3','M4','P1','P2','W1','S1','S2','P3','P4']:
  patch=out/(id+'.diff');subprocess.run(['git','apply','--check',str(patch)],cwd=root,check=True);subprocess.run(['git','apply',str(patch)],cwd=root,check=True)
  if id in ['M1','M2','M3','P1']:cmd=['timeout','90','go','run','./cmd/adamic','build',str(p/'main.ts'),'-o','/tmp/u134/'+id+'-native','--sanitize']
  elif id in ['M4','P2']:cmd=['go','vet','./internal/lower/']
  elif id=='W1':cmd=['go','vet','./'+pkg+'/']
  else:cmd=['node','--check',str(p/'testdata'/('library.mjs' if id in ['S1','P3'] else 'nontermination.mjs'))]
  started=time.monotonic();e=env.copy();e['ADAMIC_BUILD_CACHE_DIR']='/tmp/u134/cache/validate-'+id
  with (out/(id+'-build.log')).open('w')as fd:r=subprocess.run(cmd,cwd=root,env=e,stdout=fd,stderr=subprocess.STDOUT)
  checks.append({'id':id,'kind':'standalone compile','command':shlex.join(cmd),'exit':r.returncode,'seconds':time.monotonic()-started});print(checks[-1],flush=True)
  if id=='P1':
   cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./'+pkg+'/','-run','^TestSelectorThroughput$'];e=env.copy();e.pop('ADAMIC_SELECTOR_LIBRARY',None);e['ADAMIC_BUILD_CACHE_DIR']='/tmp/u134/cache/P1-no-library';t=time.monotonic()
   with (out/'P1-no-library.log').open('w') as fd:r=subprocess.run(cmd,cwd=root,env=e,stdout=fd,stderr=subprocess.STDOUT)
   checks.append({'id':'P1-no-library','kind':'supplemental environment probe','command':shlex.join(cmd),'exit':r.returncode,'seconds':time.monotonic()-t});print(checks[-1],flush=True)
  restore()
 (out/'validation.json').write_text(json.dumps(checks,indent=2))
finally:restore()
