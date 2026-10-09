import pathlib,subprocess,json,time,os,difflib,re,statistics
R=pathlib.Path('/workspace/adamic'); E=R/'review/test-audit/cmd-adamic-test262-corpus'; P=R/'cmd/adamic-test262'
rows='TestClassifyCorpus TestVerdictCorpus TestNormalizeReason TestRewriteHarnessCalls TestFrontmatterShapes TestMiniRunner TestLargeCompilerOutputIsComplete TestOutputOverflowIsReported TestEditCacheSeparation TestNodeHarnessIdentity TestWorkerLazyFallback TestLoweringSourceEdit TestRunnerLocationHelper TestRunnerLocationIdentity TestCompilerStartupMeasurement'.split()
base={f.name:f.read_text() for f in P.glob('*.go') if not f.name.endswith('_test.go')}
plans=[]
def add(id,f,old,new,kind='change constant',nth=0):
 s=base[f]; starts=[m.start() for m in re.finditer(re.escape(old),s)]; assert len(starts)>nth,(id,old);pos=starts[nth];plans.append(dict(id=id,file='cmd/adamic-test262/'+f,line=s[:pos].count('\n')+1,old=old,new=new,kind=kind,pos=pos))
add('M01','classify.go','return "negative parse or early error"','return ""')
add('M02','classify.go','case "Proxy", "Reflect"','case "Reflect"','change option')
add('M03','frontmatter.go','parsed.NegativeType = value','','drop statement')
add('M04','rewrite.go','builder.WriteByte(\'"\')','','drop statement')
add('M05','verdict.go','input.Native.Exit == 0 && input.Node.Stdout','input.Native.Exit != 0 && input.Node.Stdout','flip condition')
add('M06','verdict.go',' && input.Node.Stdout == input.Native.Stdout && input.Node.Stderr',' && input.Node.Stderr','drop condition')
add('M07','reason.go','"\'…\'"','"\'\'"')
add('M08','reason.go','return "crashed", "compiler panicked"','return "refused", "compiler panicked"')
add('M09','cache.go','cacheKey("test262-native-v1", code, library, command, context)','cacheKey("test262-native-v1", code, library, command, "")','change option')
add('M10','cache.go','cacheKey("test262-compiler-v1", program, compiler, command, context)','cacheKey("test262-compiler-v1", program, compiler, command, "")','change option')
add('M11','cache.go','parts = append(parts, "runner", cacheKey(string(contents)))','parts = append(parts, path, cacheKey(string(contents)))','change option')
add('M12','cache.go','parts = append(parts, name, string(contents))','parts = append(parts, name, string(contents[:0]))','change bound',1)
add('M13','cache.go','name == "compiler.go" || strings.HasSuffix','strings.HasSuffix','drop condition')
add('M14','run.go','buffer.exceeded = true','','drop statement')
add('M15','run.go','stdout.limit = limit','stdout.limit = 0')
s=base['run.go']; start=s.index('e.fallback.once.Do(func() {');end=s.index('\n\t\t})',start)+len('\n\t\t})');add('M16','run.go',s[start:end],'','drop statement')
add('M17','run.go','stdout.limit = limit','stdout.limit = min(limit, outputLimit)','change bound')
(E/'plan.json').write_text(json.dumps(plans,indent=2));(E/'scope.json').write_text(json.dumps(rows));(E/'functions.txt').write_text('\n'.join(re.findall(r'^func .*', '\n'.join(base.values()),re.M))+'\nConservative production inventory; scoped rows reach classification, rewriting, verdict, capture, preparation, cache and worker chains. main/CLI/report printing are reached by other package rows only.\n')
(E/'diffs').mkdir(exist_ok=True)
for p in plans:
 f=p['file'].split('/')[-1];s=base[f];changed=s[:p['pos']]+p['new']+s[p['pos']+len(p['old']):];(E/'diffs'/f"{p['id']}.diff").write_text(''.join(difflib.unified_diff(s.splitlines(True),changed.splitlines(True),fromfile='a/'+p['file'],tofile='b/'+p['file'])))
(E/'base.json').write_text(json.dumps(base))
# baseline timing before source mutation
for row in rows:
 for n in range(3):
  with (E/f'timing-{row}-{n}.log').open('w') as log:
   subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./cmd/adamic-test262/','-run','^'+row+'$'],cwd=R,stdout=log,stderr=subprocess.STDOUT)
(E/'timing.done').write_text('done')
