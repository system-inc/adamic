import json,os,time,subprocess,difflib,re,statistics
from pathlib import Path
p=Path('/tmp/u098');root=Path('/workspace/adamic');d=root/'stage1/cohere/graphql/printer'
while not (p/'matrix-done').exists():
 if (p/'costs.json').exists():raise RuntimeError('matrix stopped without completing')
 time.sleep(2)
env=os.environ.copy();env.update(ADAMIC_GRAPHQL_PRINTER_BENCH='1',ADAMIC_GRAPHQL_PRETTIER='/tmp/u082/library')
# Clean timing of an observed subsumer, after all scratch mutations are restored.
samples=[]
for i in range(3):
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/graphql/printer/','-run','^TestPrinterAsGoCohere_[0-9]{3}$']
 name=f'subsumer-time-{i+1}';start=time.monotonic()
 with (p/(name+'.log')).open('w') as f:r=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
 seconds=None
 for l in (p/(name+'.log')).read_text().splitlines():
  try:x=json.loads(l)
  except:continue
  m=re.search(r'\t([0-9.]+)s',x.get('Output',''))
  if m:seconds=float(m.group(1))
 assert r.returncode==0 and seconds is not None
 samples.append(seconds);(p/(name+'.meta')).write_text(json.dumps(dict(command=cmd,exit=r.returncode,wall=time.monotonic()-start,seconds=seconds),indent=2))
(p/'subsumer-timing.json').write_text(json.dumps(dict(test='TestPrinterAsGoCohere family',members=[f'TestPrinterAsGoCohere_{i:03d}' for i in range(4)],samples=samples,median=statistics.median(samples)),indent=2))
# Independent no-selector harness diffs make each allowed probe reviewable.
original={n:(d/n).read_text() for n in ['grain_mutant_test.go','shards_test.go','printer_test.go','printer.ts']}
grain=original['grain_mutant_test.go'];shards=original['shards_test.go'];cases=original['printer_test.go']
def function_body(s,needle):
 a=s.index(needle);start=s.index('{',a);depth=1;pos=start+1;quote=None;escape=False
 while depth:
  c=s[pos]
  if quote:
   if escape:escape=False
   elif c=='\\':escape=True
   elif c==quote:quote=None
  elif c in ['"',"'",'`']:quote=c
  elif c=='{':depth+=1
  elif c=='}':depth-=1
  pos+=1
 return start,pos

def body_replaced(s,needle,body):
 a,b=function_body(s,needle);return s[:a]+'{\n'+body+'\n}'+s[b:]
mods={}
mods['W1']=('grain_mutant_test.go',grain.replace('difference := firstDifference(string(result.stdout), want)','difference := ""',1),'disable output difference detection')
old='\tif difference == "" {\n\t\treturn "", fmt.Errorf("shard-%03d %s mutant escaped oracle", number, side)\n\t}\n'
assert old in grain;mods['W2']=('grain_mutant_test.go',grain.replace(old,'',1),'drop unchanged-answer rejection')
old='shards[number] = printerShard{mode: "defaults", path: cases, cases: printerMutantCases(number, enumeration)}'
mods['S1']=('grain_mutant_test.go',grain.replace(old,old.replace('printerMutantCases(number, enumeration)','printerMutantCases(number, enumeration)[1:]'),1),'off-by-one shard construction omits first case')
mods['S2']=('grain_mutant_test.go',grain.replace('\t\t\tprinterDirectoryAt(t, dir, mutation.file, mutation.from, mutation.to)\n','',1),'drop mutant source construction statement')
a=shards.index('func printerLoweredProduct');pos=shards.index('func(dir string) error {',a)
changed=body_replaced(shards[pos:],'func(dir string) error','\t\treturn nil');s3=shards[:pos]+changed;s3=s3.replace('\t"github.com/system-inc/adamic/internal/javascript"\n','')
mods['S3']=('shards_test.go',s3,'replace whole lowering construction callback with empty success; remove unused javascript import')
a=shards.index('func printerCompiledProduct');pos=shards.index('func(dir string) error {',a);mods['S4']=('shards_test.go',shards[:pos]+body_replaced(shards[pos:],'func(dir string) error','\t\treturn nil'),'replace whole native construction callback with empty success')
a=grain.index('func printerMutantOracle');pos=grain.index('func(dir string) error {',a);mods['S5']=('grain_mutant_test.go',grain[:pos]+body_replaced(grain[pos:],'func(dir string) error','\t\t\treturn nil'),'replace whole oracle product construction callback with empty success, without modifying oracle')
pos=cases.index('func(directory string) error {');s6=cases[:pos]+body_replaced(cases[pos:],'func(directory string) error','\t\treturn nil')
for imp in ['context','encoding/json','fmt','os/exec']:s6=s6.replace('\t"'+imp+'"\n','')
mods['S6']=('printer_test.go',s6,'replace corpus construction callback with empty success; remove four unused imports')
for fun,mid,ret in [('printerMutantSource','P2','""'),('printerMutantLowered','P3','""'),('printerMutantSanitized','P4','""'),('printerMutantOracle','P5','""'),('printerMutantCorpus','P6','"", ""'),('printerMutantCases','P7','nil')]:
 s=body_replaced(grain,'func '+fun+'(','\treturn '+ret);mods[mid]=('grain_mutant_test.go',s,'empty answer at '+fun+' entry')
records=[]
try:
 for mid,(name,changed,desc) in mods.items():
  (d/name).write_text(changed);subprocess.run(['gofmt','-w',str(d/name)],check=True)
  changed=(d/name).read_text();fn='stage1/cohere/graphql/printer/'+name
  patch=''.join(difflib.unified_diff(original[name].splitlines(True),changed.splitlines(True),fromfile='a/'+fn,tofile='b/'+fn))
  (p/(mid+'.diff')).write_text(patch)
  cmd=['go','vet','./stage1/cohere/graphql/printer/'];start=time.monotonic()
  with (p/(mid+'-vet.log')).open('w') as f:r=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
  meta=dict(command=cmd,exit=r.returncode,wall=time.monotonic()-start,description=desc);(p/(mid+'-vet.meta')).write_text(json.dumps(meta,indent=2));records.append(dict(id=mid,**meta));(d/name).write_text(original[name]);assert r.returncode==0,(mid,meta)
 # The production empty-answer probe is built with the port's own tool.
 name='printer.ts';s=body_replaced(original[name],'export function format(','    return { kind: \'Ok\', text: \'\' };');(d/name).write_text(s)
 fn='stage1/cohere/graphql/printer/'+name;(p/'P1.diff').write_text(''.join(difflib.unified_diff(original[name].splitlines(True),s.splitlines(True),fromfile='a/'+fn,tofile='b/'+fn)))
 cmd=['timeout','90','go','run','./cmd/adamic','build','stage1/cohere/graphql/printer/main.ts','-o','/tmp/u098/standalone/P1'];start=time.monotonic()
 with (p/'P1-build.log').open('w') as f:r=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
 meta=dict(command=cmd,exit=r.returncode,wall=time.monotonic()-start,artifact_exists=(p/'standalone/P1').is_file());(p/'P1-build.meta').write_text(json.dumps(meta,indent=2));assert r.returncode==0 and meta['artifact_exists']
finally:
 for name,s in original.items():(d/name).write_text(s)
(p/'probe-validation.json').write_text(json.dumps(records,indent=2));(p/'post-done').write_text('done')
