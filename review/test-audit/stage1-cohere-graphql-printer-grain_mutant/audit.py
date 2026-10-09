import os,json,time,subprocess,difflib,re
from pathlib import Path
root=Path('/workspace/adamic');p=Path('/tmp/u098');d=root/'stage1/cohere/graphql/printer'
assert (p/'timing-done').exists()
scope=json.loads((p/'scope.json').read_text());rows=scope['rows'];menu=json.loads((p/'menu.json').read_text())
files=['printer.ts','doc.ts','grain_mutant_test.go','shards_test.go'];original={name:(d/name).read_text() for name in files}
extra=[f'TestPrinterAsGoCohere_{i:03d}' for i in range(4)]+['TestPrinterFileDriver','TestPrinterShardPlantedDisagreement']+[f'TestPrinterWhitespaceGap_{i:03d}' for i in range(5)]
names=scope['requested']+extra;pattern='^('+'|'.join(names)+')$';scope['matrix_members']=names;scope['reached_extra']=extra
(p/'scope.json').write_text(json.dumps(scope,indent=2))
env=os.environ.copy();env.update(ADAMIC_GRAPHQL_PRINTER_BENCH='1',ADAMIC_GRAPHQL_PRETTIER='/tmp/u082/library')
(p/'selector').write_text('')
(p/'code-and-oracle.txt').write_text('CODE UNDER TEST: GraphQL printer port printer.ts, doc.ts and main.ts, composed with parser/lexer, character classes, block string and JSON Unicode-width ports. Production entry format. Product rows check preparation construction, and mutant rows witness the comparison. ORACLE: unchanged Go cohere printer and Prettier 3.9.6 for throughput expected bytes; self-authored construction invariants and planted survivor IDs for setup/witness rows. Node runs the same port and is cross-backend evidence, not an independent expected-answer authority. Declared in commentary before any mutation.\n')
meta_all={}
def invoke(mid,pat,mode='',cache=None):
 (p/'selector').write_text(mid if mid.startswith(('M','P1')) else '')
 e=env.copy();e['ADAMIC_AUDIT_MODE']=mode
 if cache:e['ADAMIC_BUILD_CACHE_DIR']=cache
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/graphql/printer/','-run',pat]
 t=time.monotonic()
 with (p/(mid+'.log')).open('w') as f:r=subprocess.run(cmd,cwd=root,env=e,stdout=f,stderr=subprocess.STDOUT)
 es=[]
 for l in (p/(mid+'.log')).read_text().splitlines():
  try:es.append(json.loads(l))
  except:pass
 secs=None
 for x in es:
  m=re.search(r'\t([0-9.]+)s',x.get('Output',''))
  if m:secs=float(m.group(1))
 meta=dict(command=cmd,exit=r.returncode,wall=time.monotonic()-t,seconds=secs,mode=mode,cache=cache,selector=(p/'selector').read_text())
 (p/(mid+'.meta')).write_text(json.dumps(meta,indent=2));meta_all[mid]=meta
 print(mid,meta,flush=True)
 cooked=any('panic: test timed out' in x.get('Output','') for x in es) or r.returncode==124
 return es,cooked,r.returncode

def restore():
 for name,s in original.items():(d/name).write_text(s)
try:
 # Standalone menu selected before outcomes. Build each independently against base.
 for m in menu:
  name=Path(m['file']).name;s=original[name];changed=s.replace(m['old'],m['new'],1)
  patch=''.join(difflib.unified_diff(s.splitlines(True),changed.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file']))
  (p/(m['id']+'.diff')).write_text(patch);(d/name).write_text(changed)
  target=p/'standalone'/m['id'];target.parent.mkdir(exist_ok=True)
  cmd=['timeout','90','go','run','./cmd/adamic','build','stage1/cohere/graphql/printer/main.ts','-o',str(target)]
  start=time.monotonic()
  with (p/(m['id']+'-build.log')).open('w') as f:r=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
  b=dict(command=cmd,exit=r.returncode,wall=time.monotonic()-start,artifact_exists=target.is_file())
  (p/(m['id']+'-build.meta')).write_text(json.dumps(b,indent=2));print('build',m['id'],b,flush=True)
  (d/name).write_text(s);assert r.returncode==0 and target.is_file()
 # One selector-based port build per existing product; preserve built-in mutant anchors.
 doc=original['doc.ts'].replace("import { panic } from 'adamic';","import { panic, readTextFile } from 'adamic';\nconst auditRead = readTextFile('/tmp/u098/selector');\nexport const auditMode = auditRead.kind === 'Ok' ? auditRead.text : '';",1)
 doc=doc.replace("output.push('\\n');","output.push(auditMode === 'M3' ? '\\r\\n' : '\\n');",1)
 doc=doc.replace('while(remaining >= 0)',"while(auditMode === 'M4' ? remaining > 0 : remaining >= 0)",1)
 printer=original['printer.ts'].replace('Documents, defaults, type SettingsOptions','Documents, defaults, auditMode, type SettingsOptions',1)
 printer=printer.replace("flag(node, 'value') ? 'true' : 'false'","flag(node, 'value') ? (auditMode === 'M1' ? 'TRUE' : 'true') : 'false'",1)
 printer=printer.replace('point = point * 16 +',"point = point * (auditMode === 'M2' ? 15 : 16) +",1)
 printer=printer.replace('export function format(source: string, settings: SettingsOptions = defaults): ResultType {',"export function format(source: string, settings: SettingsOptions = defaults): ResultType {\n    if(auditMode === 'P1') return { kind: 'Ok', text: '' };",1)
 grain=original['grain_mutant_test.go']
 grain=grain.replace('difference := firstDifference(string(result.stdout), want)','difference := firstDifference(string(result.stdout), want)\n\tif os.Getenv("ADAMIC_AUDIT_MODE") == "W1" { difference = "" }',1)
 grain=grain.replace('if difference == "" {','if difference == "" && os.Getenv("ADAMIC_AUDIT_MODE") != "W2" {',1)
 # Construction probes, with fresh caches for actual absent-artifact evidence.
 grain=grain.replace('printerDirectoryAt(t, dir, mutation.file, mutation.from, mutation.to)','if os.Getenv("ADAMIC_AUDIT_MODE") != "S2" { printerDirectoryAt(t, dir, mutation.file, mutation.from, mutation.to) }',1)
 grain=grain.replace('binary := printerOracle(t)','if os.Getenv("ADAMIC_AUDIT_MODE") == "S5" { return nil }\n\t\t\tbinary := printerOracle(t)',1)
 grain=grain.replace('shards[number] = printerShard{mode: "defaults", path: cases, cases: printerMutantCases(number, enumeration)}','shards[number] = printerShard{mode: "defaults", path: cases, cases: printerMutantCases(number, enumeration)}\n\t\tif os.Getenv("ADAMIC_AUDIT_MODE") == "S1" { shards[number].cases = shards[number].cases[1:] }',1)
 # Empty construction entries.
 for fun,mid,ret in [('printerMutantSource','P2','""'),('printerMutantLowered','P3','""'),('printerMutantSanitized','P4','""'),('printerMutantOracle','P5','""'),('printerMutantCorpus','P6','"", ""'),('printerMutantCases','P7','nil')]:
  a=grain.index('func '+fun+'(');b=grain.index('{',a)+1
  grain=grain[:b]+f'\n\tif os.Getenv("ADAMIC_AUDIT_MODE") == "{mid}" {{ return {ret} }}'+grain[b:]
 shards=original['shards_test.go']
 a=shards.index('func printerLoweredProduct');b=shards.index('}, func(dir string) error {',a)+len('}, func(dir string) error {')
 shards=shards[:b]+'\n\t\tif os.Getenv("ADAMIC_AUDIT_MODE") == "S3" { return nil }'+shards[b:]
 a=shards.index('func printerCompiledProduct');b=shards.index('func(dir string) error {',a)+len('func(dir string) error {')
 shards=shards[:b]+'\n\t\tif os.Getenv("ADAMIC_AUDIT_MODE") == "S4" { return nil }'+shards[b:]
 a=shards.index('func printerCases') if 'func printerCases' in shards else -1
 # S6 suppresses corpus generation callback in printer_test.go, not Go oracle implementation.
 extra_name='printer_test.go';original[extra_name]=(d/extra_name).read_text();cases=original[extra_name]
 a=cases.index('}, func(directory string) error {')+len('}, func(directory string) error {')
 cases=cases[:a]+'\n\t\tif os.Getenv("ADAMIC_AUDIT_MODE") == "S6" { return nil }'+cases[a:]
 switched={'doc.ts':doc,'printer.ts':printer,'grain_mutant_test.go':grain,'shards_test.go':shards,extra_name:cases}
 for name,s in switched.items():(d/name).write_text(s)
 subprocess.run(['gofmt','-w',str(d/'grain_mutant_test.go'),str(d/'shards_test.go'),str(d/extra_name)],check=True)
 diff=subprocess.run(['git','diff'],cwd=root,capture_output=True,text=True,check=True).stdout;(p/'scratch-switch.diff').write_text(diff)
 with (p/'vet.log').open('w') as f:vet=subprocess.run(['go','vet','./stage1/cohere/graphql/printer/'],cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
 assert vet.returncode==0
 es,cooked,rc=invoke('clean-switch',pattern,cache='/tmp/u098/cache/switch')
 if cooked:
  names=scope['requested'];pattern='^('+'|'.join(names)+')$';scope['matrix_members']=names;(p/'scope.json').write_text(json.dumps(scope,indent=2))
  es,cooked,rc=invoke('clean-switch-slice',pattern,cache='/tmp/u098/cache/switch')
 assert not cooked and rc==0,'instrumented clean matrix baseline did not pass'
 for mid in ['M1','M2','M3','M4','P1']:
  es,cooked,rc=invoke(mid,pattern,cache='/tmp/u098/cache/switch')
  if cooked:
   # Rerun each requested grouped row alone; aborted runs never infer unseen results.
   for ri,r in enumerate(rows):invoke(mid+'-row-'+str(ri),r['pattern'],cache='/tmp/u098/cache/switch')
 (p/'selector').write_text('')
 pats={r['test']:r['pattern'] for r in rows}
 special=[('W1',pats['TestPrinterMutants family']),('W2',pats['TestPrinterMutantPlantedSurvivor']),('S1',pats['TestPrinterMutantUnion']),('S2',pats['TestProduct_GraphQLMutantSource family']),('S3',pats['TestProduct_GraphQLMutantLowered family']),('S4',pats['TestProduct_GraphQLMutantSanitized family']),('S5',pats['TestProduct_GraphQLMutantOracle']),('S6',pats['TestProduct_GraphQLMutantDefaults'])]
 special += [('P2',pats['TestProduct_GraphQLMutantSource family']),('P3',pats['TestProduct_GraphQLMutantLowered family']),('P4',pats['TestProduct_GraphQLMutantSanitized family']),('P5',pats['TestProduct_GraphQLMutantOracle']),('P6',pats['TestProduct_GraphQLMutantDefaults']),('P7',pats['TestPrinterMutantUnion'])]
 for mid,pat in special:
  cache='/tmp/u098/cache/'+mid if mid.startswith('S') and mid not in ['S1'] else '/tmp/u098/cache/switch'
  es,cooked,rc=invoke(mid,pat,mode=mid,cache=cache)
  if mid in ['S2','S3','S4','S5']:
   expected={'S2':'main.ts','S3':'port.c','S4':'port','S5':'oracle'}[mid]
   (p/(mid+'-artifacts.json')).write_text(json.dumps([str(x) for x in Path(cache).rglob(expected)],indent=2))
 (p/'matrix-done').write_text('done')
finally:
 restore()
 (p/'costs.json').write_text(json.dumps(meta_all,indent=2))
