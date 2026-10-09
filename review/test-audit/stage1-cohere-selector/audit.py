import pathlib,subprocess,time,json,os,shlex
root=pathlib.Path('/workspace/adamic');pkg='stage1/cohere/selector';p=root/pkg;out=root/'review/test-audit/stage1-cohere-selector';plan=json.loads((out/'plan.json').read_text());files=list(p.glob('*.ts'))+list(p.glob('*_test.go'))+[p/'testdata/library.mjs',p/'testdata/nontermination.mjs',root/'internal/lower/object.go',root/'internal/lower/lower.go'];base={str(f.relative_to(root)):f.read_text() for f in files};runs=[];os.environ['ADAMIC_SELECTOR_LIBRARY']='/tmp/u134/library';os.environ['ADAMIC_SELECTOR_BENCH']='1'
def restore():
 for f,s in base.items(): (root/f).write_text(s)
def diff(id): (out/(id+'.diff')).write_bytes(subprocess.check_output(['git','diff','--',pkg,'internal/lower/object.go','internal/lower/lower.go'],cwd=root))
def run(id,regex='.'):
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./'+pkg+'/','-run',regex];log=out/(id+'.log');t=time.monotonic();env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u134/cache/'+id
 with log.open('w')as fd:r=subprocess.run(cmd,cwd=root,env=env,stdout=fd,stderr=subprocess.STDOUT)
 runs.append({'id':id,'command':shlex.join(cmd)+' > '+shlex.quote(str(log))+' 2>&1','env':{'ADAMIC_SELECTOR_LIBRARY':'/tmp/u134/library','ADAMIC_SELECTOR_BENCH':'1','ADAMIC_BUILD_CACHE_DIR':env['ADAMIC_BUILD_CACHE_DIR']},'exit':r.returncode,'wall_seconds':time.monotonic()-t});(out/'runs.json').write_text(json.dumps(runs,indent=2));print(id,r.returncode,round(runs[-1]['wall_seconds'],3),flush=True)
rows=['TestEachGapStandsWhereGapsMdSaysItDoes','TestThePortParsesAsGoCohereDoes','TestSelectorThroughput','TestCorpusKeepsEveryParseableFile','TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars']
try:
 for row in rows:
  for i in range(1,4):run('timing-'+row+'-'+str(i),'^'+row+'$')
 for m in plan['mutants']:
  restore();f=root/m['file'];s=f.read_text();assert s.count(m['old'])==1;f.write_text(s.replace(m['old'],m['new']));diff(m['id']);run(m['id'])
 restore();f=p/'main.ts';s=f.read_text();f.write_text(s[:s.index('const args = programArguments();')]);diff('P1');run('P1')
 restore();f=root/'internal/lower/lower.go';s=f.read_text();start=s.index('func Lower(');body=s.index('{',start);end=s.index('\n}\n',body)+2;f.write_text(s[:body+1]+'\n return nil, nil\n}'+s[end:]);diff('P2');run('P2')
 # A nil IR panic aborts the binary. Run all rows independently for actual observations.
 for row in rows:run('P2-'+row,'^'+row+'$')
 restore();f=p/'selector_test.go';s=f.read_text();start=s.index('func firstDifference(');body=s.index('{',start);end=s.index('\n}\n',body)+2;f.write_text(s[:body+1]+'\n return ""\n}'+s[end:]);diff('W1');run('W1','^TestThePortParsesAsGoCohereDoes$')
 restore();f=p/'testdata/library.mjs';s=f.read_text();old='if(selector.trim()) selectors.push(selector);';assert s.count(old)==1;f.write_text(s.replace(old,''));diff('S1');run('S1','^TestCorpusKeepsEveryParseableFile$')
 restore();f=p/'testdata/nontermination.mjs';s=f.read_text();old='new Processor(() => {}).process(text);';assert s.count(old)==1;f.write_text(s.replace(old,''));diff('S2');run('S2','^TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars$')
 restore();f=p/'testdata/library.mjs';s=f.read_text();start=s.index("if(mode === 'corpus') {");end=s.index('\nelse if(mode',start);f.write_text(s[:start]+"if(mode === 'corpus') {}"+s[end:]);diff('P3');run('P3','^TestCorpusKeepsEveryParseableFile$')
 restore();f=p/'testdata/nontermination.mjs';s=f.read_text();f.write_text(s[:s.index('const [directory, text]')]);diff('P4');run('P4','^TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars$')
finally:restore()
