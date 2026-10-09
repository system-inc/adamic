import pathlib,json,re,subprocess,time,os,difflib
r=pathlib.Path('/workspace/adamic');p=r/'review/test-audit/stage1-cohere-css-composition_shards';groups=json.loads((p/'row-members.json').read_text());menu=json.loads((p/'menu.json').read_text());runs=json.loads((p/'audit-runs.json').read_text()) if (p/'audit-runs.json').exists() else [];helpers=[]
paths={m['file'] for m in menu}|{'stage1/cohere/css/nodes.ts','stage1/cohere/css/main.ts','stage1/cohere/css/compose_main.ts','stage1/cohere/css/print_main.ts','stage1/cohere/css/css_test.go','stage1/cohere/css/composition_shards_test.go','stage1/cohere/css/printer_shards_test.go','stage1/cohere/css/parser_shards_test.go','internal/native/native.go','internal/native/emit.go','internal/lower/lower.go'};bases={f:(r/f).read_text() for f in paths}
def diff(f,new):return ''.join(difflib.unified_diff(bases[f].splitlines(True),new.splitlines(True),fromfile='a/'+f,tofile='b/'+f))
def run(cmd,log,id='clean',cache=None):
 if any(x['log']==log and x['exit']==0 for x in runs):return 0
 pathlib.Path('/tmp/u078-mutant').write_text(id);env=os.environ.copy();env['ADAMIC_MUTANT']=id
 if cache:env['ADAMIC_BUILD_CACHE_DIR']=cache
 start=time.monotonic()
 with (p/log).open('w') as out:q=subprocess.run(cmd,cwd=r,env=env,stdout=out,stderr=subprocess.STDOUT)
 rr=dict(command=cmd,log=log,selector=id,exit=q.returncode,wall=time.monotonic()-start,environment={k:v for k,v in env.items() if k.startswith('ADAMIC_CSS_') or k in ['ADAMIC_MUTANT','ADAMIC_BUILD_CACHE_DIR']});runs.append(rr);(p/'audit-runs.json').write_text(json.dumps(runs,indent=2));print(log,q.returncode,round(rr['wall'],3),flush=True);return q.returncode

def test(ts,log,id='clean',cache=None):return run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/css/','-run','^('+'|'.join(ts)+')$'],log,id,cache)
probes=[dict(id='PCompose',file='stage1/cohere/css/compose_main.ts',rows=['TestCompositionMatchesGo family']),dict(id='PPrint',file='stage1/cohere/css/print_main.ts',rows=['TestCSSPrinterOptimizedMatchesGo']),dict(id='PMain',file='stage1/cohere/css/main.ts',rows=['TestCSSThroughput'])]
for x in probes:(p/(x['id']+'.diff')).write_text(diff(x['file'],'export {};\n'));x['line']=1

goprobes=[dict(id='PLower',file='internal/lower/lower.go',signature='func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {',answer='return nil, nil',rows=['TestEachGapStandsWhereGapsMdSaysItDoes','TestClosedEmptyArrayUnionGap','TestClosedParserRegexGap','TestClosedOptionalBooleanConditionGap']),dict(id='PC',file='internal/native/emit.go',signature='func C(program *ir.Program) string {',answer='return ""',rows=['TestCSSParserOptimizedMatchesNode']),dict(id='PSetup',file='stage1/cohere/css/composition_shards_test.go',signature='func compositionSetup(t *testing.T) compositionProducts {',answer='return compositionProducts{}',rows=['TestCompositionMatchesGo setup family']),dict(id='PMode',file='stage1/cohere/css/printer_shards_test.go',signature='func cssModePlan(lines []string, count int) []cssModeShard {',answer='return nil',rows=['TestCSSPrinterAgreesWithGoUnion']),dict(id='PParserPlan',file='stage1/cohere/css/parser_shards_test.go',signature='func cssParserShards(keys []string) []cssParserShard {',answer='return nil',rows=['TestThePortParsesAsGoCohereDoesUnion'])]
for x in goprobes:
 x['line']=bases[x['file']][:bases[x['file']].index(x['signature'])].count('\n')+1;(p/(x['id']+'.diff')).write_text(diff(x['file'],bases[x['file']].replace(x['signature'],x['signature']+'\n emptyAnswer := true\n if emptyAnswer { '+x['answer']+' }\n',1)))
(p/'probes.json').write_text(json.dumps(probes+goprobes,indent=2))
range_guards={ 'stage1/cohere/css/nodes.ts':"        fields.set('range', quote(`Range [${node.range[0]} ${node.range[1]}] is not source's [${node.start.offset} ${node.end.offset}]`));\n", 'stage1/cohere/css/tree.ts':"            parts.splice(0, 0, `${quote('<range>')}:${quote(`Range [${node.range[0]} ${node.range[1]}] disagrees with the source offsets`)}`);\n"}
special=[dict(id='W1',file='stage1/cohere/css/css_test.go',signature='func firstDifference(got string, want string) string {',change='disable byte disagreement',rows=['TestCompositionMatchesGoUnion','TestCSSPrinterAgreesWithGo family','TestCSSParserPlantedDisagreement']),dict(id='WMemory',file='internal/native/native.go',old='if options.Sanitize {',new='if options.Sanitize && false {',change='disable sanitizer instrumentation flags',rows=['TestComposedMemoryChecksCanFail']),dict(id='SSetup',file='stage1/cohere/css/composition_shards_test.go',old='oracle: oracle}',new='oracle: oracle[:0]}',change='erase constructed oracle product identity',rows=['TestCompositionMatchesGo setup family']),dict(id='SMode',file='stage1/cohere/css/printer_shards_test.go',old='parts := cssPartition(lines, count/8)',new='parts := cssPartition(lines, count/4)',change='double mode partition count',rows=['TestCSSPrinterAgreesWithGoUnion']),dict(id='SParser',file='stage1/cohere/css/parser_shards_test.go',old='for variant := -1; variant < len(mutants); variant++ {',new='for variant := 0; variant < len(mutants); variant++ {',change='omit agreement variant in shard construction',rows=['TestThePortParsesAsGoCohereDoesUnion'])]
for s in special:
 old=s.get('old',s.get('signature'));s['line']=bases[s['file']][:bases[s['file']].index(old)].count('\n')+1
 if s['id']=='W1':new=bases[s['file']].replace(s['signature'],s['signature']+'\n acceptEverything := true\n if acceptEverything { return "" }\n',1)
 else:new=bases[s['file']].replace(s['old'],s['new'],1)
 (p/(s['id']+'.diff')).write_text(diff(s['file'],new))
(p/'WRange.diff').write_text(''.join(diff(f,bases[f].replace(guard,'',1)) for f,guard in range_guards.items()));special.append(dict(id='WRange',file='stage1/cohere/css/nodes.ts + tree.ts',line=102,change='drop both canonical range mismatch reporting statements',rows=['TestTheCanonicalRangeChecksCanFail']));(p/'special-edits.json').write_text(json.dumps(special,indent=2))
try:
 assert run(['timeout','90','go','build','-o','/tmp/u078-adamic','./cmd/adamic'],'compiler-build.log')==0
 for m in menu:
  f=m['file'];(r/f).write_text(bases[f].replace(m['old'],m['new'],1));assert run(['timeout','90','/tmp/u078-adamic','build','stage1/cohere/css/'+m['entry'],'-o','/tmp/u078-'+m['id'],'--sanitize'],m['id']+'-standalone-build.log')==0;(r/f).write_text(bases[f])
 for x in probes:
  (r/x['file']).write_text('export {};\n');assert run(['timeout','90','/tmp/u078-adamic','build',x['file'],'-o','/tmp/u078-'+x['id'],'--sanitize'],x['id']+'-standalone-build.log')==0;(r/x['file']).write_text(bases[x['file']])
 for x in goprobes:
  f=x['file'];(r/f).write_text(bases[f].replace(x['signature'],x['signature']+'\n emptyAnswer := true\n if emptyAnswer { '+x['answer']+' }\n',1));assert run(['timeout','90','go','vet','./'+str(pathlib.Path(f).parent)+'/'],x['id']+'-vet.log')==0;(r/f).write_text(bases[f])
 for s in special:
  if s['id']=='WRange':continue
  f=s['file'];new=bases[f].replace(s['signature'],s['signature']+'\n acceptEverything := true\n if acceptEverything { return "" }\n',1) if s['id']=='W1' else bases[f].replace(s['old'],s['new'],1);(r/f).write_text(new);assert run(['timeout','90','go','vet','./'+str(pathlib.Path(f).parent)+'/'],s['id']+'-vet.log')==0;(r/f).write_text(bases[f])
 for f,guard in range_guards.items():(r/f).write_text(bases[f].replace(guard,'',1))
 assert run(['timeout','90','/tmp/u078-adamic','build','stage1/cohere/css/testdata/composition_range_guard.ts','-o','/tmp/u078-WRange','--sanitize'],'WRange-standalone-build.log')==0
 for f in range_guards:(r/f).write_text(bases[f])
 # One TS controller file is read by each product at process start; changing its contents never changes the oracle.
 selection=r/'stage1/cohere/css/u078_selection.ts';selection.write_text("import { readTextFile } from 'adamic';\nfunction selectAudit(): string {\n const read = readTextFile('/tmp/u078-mutant');\n if(read.kind === 'Error') return '';\n return read.text.trim();\n}\nexport const auditSelection = selectAudit();\n");helpers.append(selection);(p/'selection.ts.txt').write_text(selection.read_text())
 gohelper=r/'stage1/cohere/css/u078_audit_test.go';gohelper.write_text('package css\nimport "os"\nfunc auditUnitMutant()string{return os.Getenv("ADAMIC_MUTANT")}\nfunc auditUnitString(id,a,b string)string{if auditUnitMutant()==id{return a};return b}\n');helpers.append(gohelper);(p/'go-switch-helper.go.txt').write_text(gohelper.read_text())
 for f,base in bases.items():
  if f.endswith('.ts'):
   needs=f in {m['file'] for m in menu}|set(range_guards)|{x['file'] for x in probes}
   if needs:base="import { auditSelection } from './u078_selection.ts';\n"+base
   for m in menu:
    if m['file']!=f:continue
    repl={'M1':"        if(auditSelection !== 'M1') parser.parse();\n",'M2':"node.setBoolean('import', auditSelection !== 'M2');",'M3':"return quote(auditSelection === 'M3' ? '' : node.literalValue);",'M4':"n.kind === 'indent' ? (auditSelection === 'M4' ? 0 : 1) : -1"}[m['id']];base=base.replace(m['old'],repl,1)
   if f in range_guards:base=base.replace(range_guards[f],"if(auditSelection !== 'WRange') {\n"+range_guards[f]+"}\n",1)
   for x in probes:
    if x['file']==f:
     off=base.index('const args = programArguments();');base=base[:off]+"function auditPortMain(): void {\n if(auditSelection === '"+x['id']+"') return;\n"+base[off:]+"\n}\nauditPortMain();\n"
  else:
   for x in goprobes:
    if x['file']==f:base=base.replace(x['signature'],x['signature']+'\n if '+('os.Getenv("ADAMIC_MUTANT")' if f.startswith('internal/') else 'auditUnitMutant()')+' == "'+x['id']+'" { '+x['answer']+' }\n',1)
   for s in special:
    if s['file']!=f:continue
    if s['id']=='W1':base=base.replace(s['signature'],s['signature']+'\n if auditUnitMutant() == "W1" { return "" }\n',1)
    elif s['id']=='WMemory':base=base.replace(s['old'],'if options.Sanitize && os.Getenv("ADAMIC_MUTANT") != "WMemory" {',1)
    elif s['id']=='SSetup':base=base.replace(s['old'],'oracle: auditUnitString("SSetup", "", oracle)}',1)
    elif s['id']=='SMode':base=base.replace(s['old'],'divisor := 8\n if auditUnitMutant() == "SMode" { divisor = 4 }\n parts := cssPartition(lines, count/divisor)',1)
    elif s['id']=='SParser':base=base.replace(s['old'],'startVariant := -1\n if auditUnitMutant() == "SParser" { startVariant = 0 }\n for variant := startVariant; variant < len(mutants); variant++ {',1)
   if f in ['internal/lower/lower.go','internal/native/emit.go'] and '"os"' not in base:base=base.replace('import (','import (\n "os"',1)
  (r/f).write_text(base);(p/(f.replace('/','_')+'.switch.txt')).write_text(base)
 assert run(['gofmt','-w']+[f for f in bases if f.endswith('.go')]+[str(gohelper)],'switch-gofmt.log')==0
 # Preserve original-line mapping for every reported failing test line.
 maps={}
 for f in bases:
  if not f.endswith('.go'):continue
  original=bases[f].splitlines();scratch=(r/f).read_text().splitlines();mapping={}
  for a,b,n in difflib.SequenceMatcher(None,original,scratch,autojunk=False).get_matching_blocks():
   for k in range(n):mapping[str(b+k+1)]=a+k+1
  maps[pathlib.Path(f).name]=mapping
 (p/'origin-line-maps.json').write_text(json.dumps(maps,indent=2))
 assert run(['timeout','90','go','test','-c','-o','/tmp/u078-test','./stage1/cohere/css/'],'switch-build.log')==0
 # Instrumented clean checks are bounded to production callers. The coupled Node oracle stays clean.
 for row in ['TestCompositionMatchesGo family','TestCSSThroughput','TestCSSPrinterOptimizedMatchesGo','TestCSSParserOptimizedMatchesNode']:
  code=test(groups[row],'switch-baseline-'+row.replace(' ','_')+'.log')
  if code!=0:raise RuntimeError('instrumented clean failure or budget: '+row)
 for s in special:
  for row in s['rows']:
   ts=groups[row]
   if row=='TestCSSPrinterAgreesWithGo family':ts=[t for t in ts if int(t.rsplit('_',1)[1])>=16]
   test(ts,s['id']+'-'+row.replace(' ','_')+'.log',s['id'])
 for m in menu:
  for row in m['rows']:test(groups[row],m['id']+'-'+row.replace(' ','_')+'.log',m['id'])
 for x in probes+goprobes:
  for row in x['rows']:
   # Isolate production entry probes and individual setup-family members.
   if len(groups[row])==2 and 'setup family' in row:
    for t in groups[row]:test([t],x['id']+'-'+t+'.log',x['id'])
   else:test(groups[row],x['id']+'-'+row.replace(' ','_')+'.log',x['id'], '/tmp/u078/cache/'+x['id'] if x['id'] in ['PC','PLower'] else None)
finally:
 for f,base in bases.items():(r/f).write_text(base)
 for f in helpers:
  if f.exists():f.unlink()
 pathlib.Path('/tmp/u078-mutant').write_text('clean')
run(['go','vet','./stage1/cohere/css/'],'final-vet.log');run(['git','diff','--check'],'source-diff-check.log')
