import pathlib,subprocess,time,json,difflib,os
repo=pathlib.Path('/workspace/adamic');p=repo/'review/test-audit/stage1-cohere-typeaware-inspect_shards';tmp=pathlib.Path('/tmp/u145');base='stage1/cohere/typeaware/'
menu=[('M1',base+'frames.ts','value = value * 10 + digit;','value = value * 9 + digit;',"value = value * (audit('M1') ? 9 : 10) + digit;",'change constant in decimal'),('M2',base+'frames.ts','return value === 1;','return value === 0;',"return value === (audit('M2') ? 0 : 1);",'change boolean return constant'),('M3',base+'facts.ts','tuple < -1','tuple < 0',"tuple < (audit('M3') ? 0 : -1)",'off-by-one tuple bound'),('M4',base+'testdata/fact_cost.ts','count < 2','count < 3',"count < (audit('M4') ? 3 : 2)",'off-by-one minimum count')]
originals={file:(repo/file).read_text() for _,file,*_ in menu}; manifest=[]
for id,file,old,new,switch,desc in menu:
 s=originals[file];assert s.count(old)==1
 manifest.append({'id':id,'file':file,'line':s[:s.index(old)].count('\n')+1,'change':new,'menu':desc})
(p/'production-menu.json').write_text(json.dumps(manifest,indent=2))
(p/'code-under-test.md').write_text('CODE UNDER TEST: stage1 TypeScript facts decoder and fact_cost native entry. ORACLE: manually written frames, stdout 64, exit 70 and error substrings for decoder; request validity/refusals use exit codes and named errors. Go cohere exact findings decide Six agreement.\n\nDecoder reached functions: facts.contains, facts.types; Frames.constructor,next,field,number,natural,yes,ids,end; frames.decimal,header; Types.constructor,type,root; TypeFact.constructor; testdata/facts_decode top-level. Request entry: testdata/fact_cost top-level plus linked bridge tsgoProgram,tsgoInspect,tsgoRelease, and built-in guard variants. Other requested entries include Rules/UnaryMinus runs, Shadow.run/report and compiler/parser dependencies. Full dynamic reachability for the timed-out complete Six family was not established. No mutations are selected from unestablished reachability.\n\nFour production mutants were frozen before their outcomes. Construction/witness controls are separate.\n')
# Freeze only when called with --plan.
import sys
if '--plan' in sys.argv: print('menu frozen');raise SystemExit()
runs=[];builds=[]
try:
 cmd=['go','build','-o',str(tmp/'adamic'),'./cmd/adamic'];start=time.monotonic()
 with (p/'audit-compiler-build.log').open('w') as f:subprocess.run(cmd,cwd=repo,stdout=f,stderr=subprocess.STDOUT,check=True)
 builds.append({'id':'compiler','seconds':time.monotonic()-start,'command':cmd,'exit':0})
 cmd=['go','build','-buildmode=c-archive','-o',str(tmp/'checker.a'),'./bridge/tsgo/archive'];start=time.monotonic()
 with (p/'audit-archive-build.log').open('w') as f:subprocess.run(cmd,cwd=repo,stdout=f,stderr=subprocess.STDOUT,check=True)
 builds.append({'id':'archive','seconds':time.monotonic()-start,'command':cmd,'exit':0})
 for id,file,old,new,switch,desc in menu:
  s=originals[file];modified=s.replace(old,new,1);diff=''.join(difflib.unified_diff(s.splitlines(True),modified.splitlines(True),fromfile='a/'+file,tofile='b/'+file));(p/(id+'.diff')).write_text(diff)
  subprocess.run(['git','apply','--check',str(p/(id+'.diff'))],cwd=repo,check=True);(repo/file).write_text(modified)
  entry=base+'testdata/fact_cost.ts' if id=='M4' else base+'testdata/facts_decode.ts';cmd=['timeout','90',str(tmp/'adamic'),'build',entry,'-o',str(tmp/id)]
  cmd+=['--tsgo',str(tmp/'checker.a')] if id=='M4' else ['--sanitize'];start=time.monotonic()
  with (p/(id+'-build.log')).open('w') as f:r=subprocess.run(cmd,cwd=repo,stdout=f,stderr=subprocess.STDOUT)
  builds.append({'id':id,'seconds':time.monotonic()-start,'command':cmd,'exit':r.returncode});(repo/file).write_text(s)
  if r.returncode: raise RuntimeError('standalone build failed '+id)
 switched={}
 for id,file,old,new,switch,desc in menu:switched[file]=switched.get(file,originals[file]).replace(old,switch,1)
 marker='export function types(text: string, question: string): Types {'
 switched[base+'facts.ts']=switched[base+'facts.ts'].replace(marker,marker+"\n    if(audit('P1')) { return new Types(false, false, [], [], []); }",1)
 probe=originals[base+'facts.ts'].replace(marker,marker+'\n    if(text.length >= 0) { return new Types(false, false, [], [], []); }',1)
 (p/'P1.diff').write_text(''.join(difflib.unified_diff(originals[base+'facts.ts'].splitlines(True),probe.splitlines(True),fromfile='a/'+base+'facts.ts',tofile='b/'+base+'facts.ts')))
 for file,s in switched.items():
  relative='../audit.ts' if '/testdata/' in file else './audit.ts';(repo/file).write_text("import { audit } from '"+relative+"';\n"+s)
 helper=repo/base/'audit.ts';helper.write_text("import { readTextFile } from 'adamic';\nexport function audit(id: string): boolean { const value = readTextFile('/tmp/u145/selector'); return value.kind === 'Error' ? false : value.text === id; }\n")
 with (p/'scratch-switch.diff').open('w') as f:subprocess.run(['git','diff','--',*originals.keys()],cwd=repo,stdout=f)
 (p/'scratch-audit.ts.txt').write_text(helper.read_text())
 for id in ['M1','M2','M3','M4','P1']:
  (tmp/'selector').write_text(id); pattern='^(TestInspectRequestRefusals|TestSixRuleAgreementAndMutants_000)$' if id=='M4' else '^(TestFactsDecoderGuards|TestSixRuleAgreementAndMutants_000)$'
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/typeaware/','-run',pattern];start=time.monotonic()
  with (p/(id+'.log')).open('w') as f:r=subprocess.run(cmd,cwd=repo,env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE='/tmp/u145/typescript',ADAMIC_TYPEAWARE_BENCH='1'),stdout=f,stderr=subprocess.STDOUT)
  runs.append({'id':id,'command':cmd,'seconds':time.monotonic()-start,'exit':r.returncode});(p/'production-runs.json').write_text(json.dumps(runs,indent=2))
finally:
 for file,s in originals.items():(repo/file).write_text(s)
 helper=repo/base/'audit.ts'
 if helper.exists():helper.unlink()
 (tmp/'selector').write_text('')
 (p/'build-times.json').write_text(json.dumps(builds,indent=2))
print('production matrix complete')
