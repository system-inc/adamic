import pathlib,json,subprocess,time,os,difflib
R=pathlib.Path.cwd();E=R/'review/test-audit/stage1-cohere-lint-shards_agree_split';D=E/'diffs';D.mkdir(exist_ok=True);env=os.environ.copy();env['ADAMIC_TYPESCRIPT_SOURCE']='/tmp/u112/TypeScript-050880ce59e30b356b686bd3144efe24f875ebc8';results=[]
def run(name,regex,extra=None,limit=90):
 en=env.copy();en.update(extra or {});cmd=['timeout','120','go','test','-json','-count=1','-timeout',str(limit)+'s','./stage1/cohere/lint/','-run',regex];t=time.monotonic()
 with (E/(name+'.log')).open('w') as f:p=subprocess.run(cmd,env=en,stdout=f,stderr=subprocess.STDOUT)
 events=[]
 for l in (E/(name+'.log')).read_text().splitlines():
  try:events.append(json.loads(l))
  except:pass
 x={'id':name,'command':cmd,'environment':extra or {},'exit':p.returncode,'wall':time.monotonic()-t,'events':[e for e in events if e.get('Action') in ['pass','fail','skip']],'timeout':any('test timed out' in e.get('Output','') for e in events) or p.returncode==124};results.append(x);(E/'audit-runs.json').write_text(json.dumps(results,indent=2));print(name,p.returncode,round(x['wall'],3),flush=True);return x
# Guard: an assertion-red clean baseline stops the entire audit.
clean=json.loads((E/'clean-runs.json').read_text())
if any(x['exit'] and not x['timeout'] for x in clean):raise SystemExit('red baseline, no mutants')
for i in range(3):
 x=run('clean-kind-family-'+str(i+1), '^TestWitnessScriptKind(?:_[0-9]+)?$')
 if x['exit']:raise SystemExit('kind family baseline failed')
paths=['stage1/cohere/lint/main.ts','stage1/cohere/lint/shards/shards.go'];orig={p:(R/p).read_text() for p in paths}
def diff(id,p,new):
 (D/(id+'.diff')).write_text(''.join(difflib.unified_diff(orig[p].splitlines(True),new.splitlines(True),fromfile='a/'+p,tofile='b/'+p)))
main=orig[paths[0]];go=orig[paths[1]];loop='\tfor number := 0; number < rows; number++ {\n\t\tblock, ok := blocks[number]\n\t\tif !ok {\n\t\t\treturn nil, fmt.Errorf("case %d missing from every shard", number)\n\t\t}\n\t\tmerged.Write(block)\n\t}\n';assert loop in go
variants={'M1':(paths[0],main.replace('caseNumber % shardCount === shardIndex','caseNumber % (shardCount + 1) === shardIndex')),'M2':(paths[0],main.replace('`range ${start} ${end}','`range ${start + 1} ${end}')),'M3':(paths[1],go.replace('total := 0','total := 1')),'M4':(paths[1],go.replace(loop,''))}
# Entry probes replace complete standalone bodies, avoiding unreachable-code vet failures.
def body(s,signature,replacement):
 start=s.index('{',s.index(signature));depth=1;i=start+1
 while depth:
  if s[i]=='{':depth+=1
  if s[i]=='}':depth-=1
  i+=1
 return s[:start+1]+'\n'+replacement+'\n'+s[i-1:]
p1=body(main,'function run(', '    return 0;')
for line in ["import { written } from '../../typescript/parser/nodes.ts';\n","import { Parser } from '../../typescript/parser/parser.ts';\n","import { Scanner } from '../../typescript/scanner/scanner.ts';\n","import { Linter } from './lint.ts';\n","import { Settings } from './settings.ts';\n"]:p1=p1.replace(line,'')
p2=body(go,'func Run(', '\treturn nil, nil').replace('\t"os/exec"\n','').replace('\t"sync"\n','').replace('\t"os"\n','')
variants['P1']=(paths[0],p1);variants['P2']=(paths[1],p2)
for id,(p,s) in variants.items():diff(id,p,s)
# Stable runtime selector in the port, one native build per existing recipe.
sm=main.replace("// Test only: see Linter.junkRows.","const auditSelection = readTextFile('/tmp/u112/selector');\nconst auditMutant = auditSelection.kind === 'Error' ? '' : auditSelection.text.trim();\n\n// Test only: see Linter.junkRows.")
sm=sm.replace('function run(row: string, countOnly: boolean): number {','function run(row: string, countOnly: boolean): number {\n    if(auditMutant === \'P1\') { return 0; }')
sm=sm.replace('caseNumber % shardCount === shardIndex',"caseNumber % (shardCount + (auditMutant === 'M1' ? 1 : 0)) === shardIndex").replace('`range ${start} ${end}',"`range ${start + (auditMutant === 'M2' ? 1 : 0)} ${end}")
sg=go.replace('func Run(binary, manifest string, count int, countOnly bool) ([]byte, error) {','func Run(binary, manifest string, count int, countOnly bool) ([]byte, error) {\n\tif os.Getenv("ADAMIC_MUTANT") == "P2" { return nil, nil }').replace('total := 0','total := 0\n\t\tif os.Getenv("ADAMIC_MUTANT") == "M3" { total = 1 }').replace(loop,'\tif os.Getenv("ADAMIC_MUTANT") != "M4" {\n'+loop+'\t}\n')
try:
 (R/paths[0]).write_text(sm);(R/paths[1]).write_text(sg);subprocess.run(['gofmt','-w',paths[1]],check=True);(E/'scratch-main.ts.txt').write_text(sm);(E/'scratch-shards.go.txt').write_text((R/paths[1]).read_text());pathlib.Path('/tmp/u112/selector').write_text('')
 for g,rx in [('suggestion','^TestSuggestionAlongsideAutomaticFix_[0-9]+$'),('kind','^TestWitnessScriptKind(?:_[0-9]+)?$')]:
  x=run('switched-control-'+g,rx)
  if x['exit']:raise RuntimeError('switched control failed')
 for id in ['M1','M2','M3','M4','P1','P2']:
  pathlib.Path('/tmp/u112/selector').write_text(id)
  for g,rx in [('suggestion','^TestSuggestionAlongsideAutomaticFix_[0-9]+$'),('kind','^TestWitnessScriptKind(?:_[0-9]+)?$')]:run(id+'-'+g,rx,{'ADAMIC_MUTANT':id})
 # Changed-output witnesses do not count as row kills.
 pathlib.Path('/tmp/u112/two-manifest.txt').write_text('/tmp/u112/source.ts\tno-debugger\n/tmp/u112/source.ts\tno-debugger\n')
 for id in ['', 'M1', 'M2']:
  pathlib.Path('/tmp/u112/selector').write_text(id)
  cmd=['node','--disable-warning=ExperimentalWarning','oracle/node.mjs','stage1/cohere/lint/main.ts','--manifest','/tmp/u112/two-manifest.txt']
  with (E/('witness-'+(id or 'clean')+'.log')).open('w') as f:subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT,check=True)
 pathlib.Path('/tmp/u112/producer').write_text('#!/bin/sh\ncase "$*" in *--count*) printf \'7\\n\';; *) printf \'case 0\\nbody\\n\';; esac\n')
 pathlib.Path('/tmp/u112/producer').chmod(0o755)
 pathlib.Path('/tmp/u112/shard-witness.go').write_text('''package main
import("fmt"; "github.com/system-inc/adamic/stage1/cohere/lint/shards")
func main(){count,err:=shards.Run("/tmp/u112/producer","/tmp/u112/manifest.txt",1,true);fmt.Printf("Run count=%q error=%v\\n",count,err);out,err:=shards.Merge([][]byte{[]byte("case 0\\nbody\\n")},1);fmt.Printf("Merge output=%q error=%v\\n",out,err)}
''')
 for id in ['', 'M3', 'M4']:
  en=env.copy();en['ADAMIC_MUTANT']=id
  with (E/('witness-shards-'+(id or 'clean')+'.log')).open('w') as f:subprocess.run(['go','run','/tmp/u112/shard-witness.go'],env=en,stdout=f,stderr=subprocess.STDOUT,check=True)
 # Shards family needs its full corpus; record one bounded matrix attempt per state if baseline finished.
 if any(x['group']=='shards' and x['exit']==0 for x in clean):
  for id in ['M1','M2','M3','M4','P2']:
   pathlib.Path('/tmp/u112/selector').write_text(id);run(id+'-shards','^TestShardsAgree_[0-9]+$',{'ADAMIC_MUTANT':id})
finally:
 for p,s in orig.items():(R/p).write_text(s)
 pathlib.Path('/tmp/u112/selector').write_text('')
# Standalone replay compilation, each diff applies independently.
for id,(p,s) in variants.items():
 t=time.monotonic()
 try:
  (R/p).write_text(s)
  cmd=['go','vet','./stage1/cohere/lint/shards/'] if p.endswith('.go') else ['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/','-run','^TestProduct_WitnessScriptKindNative$']
  with (E/(id+'-compile.log')).open('w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
  results.append({'id':id+'-compile','command':cmd,'exit':r.returncode,'wall':time.monotonic()-t});(E/'audit-runs.json').write_text(json.dumps(results,indent=2));print(id,'compile',r.returncode,flush=True)
 finally:(R/p).write_text(orig[p])
