import pathlib,subprocess,json,time,os,difflib
R=pathlib.Path('/workspace/adamic');P=R/'review/test-defend/cmd-adamic-test262-corpus';PKG='./cmd/adamic-test262/'
plans=[
 dict(id='D01',test='TestVerdictCorpus',file='cmd/adamic-test262/verdict.go',old='nativeFailed := input.Native.Exit != 0',new='nativeFailed := input.Native.Exit == 0',menu='flip condition',difference='decideNegative covered by VerdictCorpus, not EditCacheSeparation'),
 dict(id='D02',test='TestFrontmatterShapes',file='cmd/adamic-test262/frontmatter.go',old='\t\t\taddField(&parsed, listKey, unquote(strings.TrimSpace(trim[2:])))',new='',menu='drop statement',difference='block-list addField branch covered only by Shapes versus ClassifyCorpus'),
 dict(id='D03',test='TestMiniRunner',file='cmd/adamic-test262/run.go',old='\t\ttarget = filepath.Join(root, filepath.FromSlash(filter))',new='',menu='drop statement',difference='nonempty-filter path covered only by MiniRunner versus ParallelCachedMatchesSerial'),
 dict(id='D04',test='TestEditCacheSeparation',file='cmd/adamic-test262/run.go',old='nativeResultKey(lowered.Stdout, e.runtimeKey, nativeCommand, e.context)',new='nativeResultKey(lowered.Stdout, "", nativeCommand, e.context)',menu='change identity argument to empty constant',difference='EditCacheSeparation changes runtimeKey while LoweringSourceEdit changes context; statement coverage does not record these argument histories'),
 dict(id='D05',test='TestCompilerStartupMeasurement',file='cmd/adamic-test262/run.go',old='\tcommand.Stdout = &stdout',new='',menu='drop statement',difference='No exclusive blocks; test exact subprocess/in-process C comparison by dropping captured stdout'),
 dict(id='D06',test='TestCompilerStartupMeasurement',file='cmd/adamic-test262/run.go',old='buffer.buf.Write(data)',new='buffer.buf.Write(data[:max(0, len(data)-1)])',menu='off-by-one write bound',difference='No exclusive blocks; lose one final byte from each capture write, preserving successful compiler exit'),
 dict(id='D07',test='TestCompilerStartupMeasurement',file='cmd/adamic-test262/run.go',old='if stdout.exceeded || stderr.exceeded {',new='if !stdout.exceeded || stderr.exceeded {',menu='flip condition',difference='No exclusive blocks; reject complete stdout capture as overflow')]
original={x['file']:(R/x['file']).read_text() for x in plans}
for x in plans:
 s=original[x['file']];assert s.count(x['old'])==1,(x['id'],s.count(x['old']));x['line']=s[:s.index(x['old'])].count('\n')+1
 diff=''.join(difflib.unified_diff(s.splitlines(True),s.replace(x['old'],x['new'],1).splitlines(True),fromfile='a/'+x['file'],tofile='b/'+x['file']));(P/(x['id']+'.diff')).write_text(diff)
(P/'plan.json').write_text(json.dumps(plans,indent=2));res=[]
def run(id,cmd,env=None):
 t=time.monotonic()
 with (P/(id+'.log')).open('w') as f:code=subprocess.run(cmd,cwd=R,stdout=f,stderr=subprocess.STDOUT,env=env).returncode
 result=dict(id=id,command=cmd,seconds=time.monotonic()-t,exit=code,selector=(env or {}).get('ADAMIC_MUTANT'),build_cache=(env or {}).get('ADAMIC_BUILD_CACHE_DIR'));res.append(result);(P/'runs.json').write_text(json.dumps(res,indent=2));print(id,code,round(result['seconds'],2),flush=True);return code
for x in plans:
 f=R/x['file'];f.write_text(original[x['file']].replace(x['old'],x['new'],1))
 try:
  if run(x['id']+'-vet',['timeout','90','go','vet',PKG]):raise SystemExit('standalone does not compile')
 finally:f.write_text(original[x['file']])
# Instrument production code only. Replay diffs above contain no selectors.
for f,s in original.items():
 for x in [x for x in plans if x['file']==f]:
  if x['id']=='D01':new='nativeFailed := input.Native.Exit != 0; if os.Getenv("ADAMIC_MUTANT") == "D01" { nativeFailed = input.Native.Exit == 0 }'
  elif x['id'] in ['D02','D03','D05']:new='if os.Getenv("ADAMIC_MUTANT") != "'+x['id']+'" { '+x['old'].strip()+' }'
  elif x['id']=='D04':new='nativeResultKey(lowered.Stdout, defenseRuntimeKey(e.runtimeKey), nativeCommand, e.context)'
  elif x['id']=='D06':new='buffer.buf.Write(defenseBytes(data))'
  elif x['id']=='D07':new='if defenseOverflow(stdout.exceeded, stderr.exceeded) {'
  s=s.replace(x['old'],new,1)
 if f.endswith('verdict.go') or f.endswith('frontmatter.go'):s=s.replace('import (','import (\n "os"',1)
 (R/f).write_text(s)
helper=R/'cmd/adamic-test262/defense_selector.go';helper.write_text('''package main
import "os"
func defenseRuntimeKey(key string) string { if os.Getenv("ADAMIC_MUTANT") == "D04" { return "" }; return key }
func defenseBytes(data []byte) []byte { if os.Getenv("ADAMIC_MUTANT") == "D06" { return data[:max(0,len(data)-1)] }; return data }
func defenseOverflow(stdout,stderr bool) bool { if os.Getenv("ADAMIC_MUTANT") == "D07" { return !stdout || stderr }; return stdout || stderr }
''')
try:
 if run('switch-vet',['timeout','90','go','vet',PKG]):raise SystemExit('switch does not compile')
 for x in plans:
  env=os.environ.copy();env['ADAMIC_MUTANT']=x['id'];env['ADAMIC_BUILD_CACHE_DIR']='/workspace/defend-test262-cache/'+x['id']
  run(x['id'],['timeout','120','go','test','-json','-count=1','-timeout','90s',PKG,'-run','.'],env)
  text=(P/(x['id']+'.log')).read_text()
  events=[]
  for line in text.splitlines():
   try:events.append(json.loads(line))
   except:pass
  if any('panic: test timed out' in e.get('Output','') for e in events):
   print('COOKED: full package, narrowing to row and named subsumer',flush=True)
   subs={'TestVerdictCorpus':'TestEditCacheSeparation','TestFrontmatterShapes':'TestClassifyCorpus','TestMiniRunner':'TestParallelCachedMatchesSerial','TestEditCacheSeparation':'TestLoweringSourceEdit','TestCompilerStartupMeasurement':'TestLargeCompilerOutputIsComplete'}
   regex='^('+x['test']+'|'+subs[x['test']]+')$';run(x['id']+'-narrow',['timeout','120','go','test','-json','-count=1','-timeout','90s',PKG,'-run',regex],env)
finally:
 for f,s in original.items():(R/f).write_text(s)
 helper.unlink()
