#!/usr/bin/env python3
"""Validate unaffected memory/corpus/mutant behavior; preserve the failing full oracle."""
import json,subprocess,time,re
from pathlib import Path
from controls import controls
own=Path(__file__).resolve().parent;r=own.parents[4];out=Path('/workspace/wave04-context/check');compiler=Path('/workspace/typeaware-wave-04-landing-d65/adamic');archives=Path('/workspace/typeaware-wave-04-landing-d65/next');results={}
def run(label,args,cwd=r,expected=0):
 start=time.monotonic()
 with (out/(label+'.stdout')).open('wb') as so,(out/(label+'.stderr')).open('wb') as se:code=subprocess.run(list(map(str,args)),cwd=cwd,stdout=so,stderr=se).returncode
 results[label]={'exit':code,'seconds':time.monotonic()-start};(out/'remaining-results.json').write_text(json.dumps(results,indent=2))
 if code!=expected:raise RuntimeError(f'{label}: exit {code}')
 return (out/(label+'.stdout')).read_bytes(),(out/(label+'.stderr')).read_bytes()
records=json.loads((out/'individual-comparisons.json').read_text());faults=[x for x in records if not x['byte_equal']];assert len(faults)==1 and Path(faults[0]['path']).name=='control-035.tsx' and faults[0]['go_exit']==2 and faults[0]['native_exit']==0
assert b'interface conversion: ast.nodeData is *ast.SatisfiesExpression, not *ast.AsExpression' in (out/'individual-035-go.stderr').read_bytes()
run('remaining-sanitize-build',[compiler,'build',own/'source_main.a','-o',out/'sanitize','--tsgo',archives/'checker-asan.a','--sanitize'])
config=out/'tsconfig.json';manifest=out/'manifest'
for at,record in enumerate(records):
 actual,error=run(f'individual-{at:03}-sanitize',[out/'sanitize',config,manifest,'default',record['path']]);assert not error
 baseline=(out/f'individual-{at:03}-native.stdout').read_bytes();assert actual==baseline
 if record['byte_equal']:assert actual==(out/f'individual-{at:03}-go.stdout').read_bytes()
print('Memory checks: every one of 110 inputs executed under sanitizers; 109 agree with Go; the upstream-fault witness remains a failed oracle.',flush=True)
for name,cfg,roots in [('compiler',Path('/workspace/typeaware-wave-04/typescript/src/compiler/tsconfig.json'),Path('/workspace/typeaware-wave-04/compiler.manifest')),('repository',r/'tsconfig.json',Path('/workspace/typeaware-wave-04/repository.manifest'))]:
 truth,_=run(name+'-go',[out/'oracle',cfg,roots]);actual,error=run(name+'-native',[out/'native',cfg,roots]);assert truth==actual and not error
 actual,error=run(name+'-sanitize',[out/'sanitize',cfg,roots]);assert actual==truth and not error
 print(f'Corpus {name}: {len(truth)} complete finding bytes equal, normal and sanitized.',flush=True)
ids=set()
for at,record in enumerate(records):
 if record['byte_equal']:
  for row in (out/f'individual-{at:03}-go.stdout').read_text().splitlines():
   fields=row.split('\t')
   if len(fields)>3:ids.add(fields[3])
assert {'defaultMsg','defaultMsgFunc','withIdentifierMsg','withIdentifierMsgFunc','memoWithoutDependenciesMsg','unstableDependencyMsg'}<=ids
sources=controls();cases={'span':0,'depth':next(i for i,x in enumerate(sources) if 'function make(){return {};}' in x),'escape':next(i for i,x in enumerate(sources) if 'store.value=value' in x),'primitive':next(i for i,x in enumerate(sources) if 'function make():number' in x)}
mutations=[('span','source.a','origin.source.rules.byte(origin.source.node(origin.index).end)','origin.source.rules.byte(origin.source.node(origin.index).end)+1'),('depth','stability.a','if(depth>8||this.active.has(key)||function_.source.body','if(depth>1||this.active.has(key)||function_.source.body'),('escape','stability.a','if(result.fresh!==undefined&&result.other&&this.anyEscapes(own,false))','if(false)'),('primitive','stability.a','if(this.primitive(node))','if(false)')]
for label,module,before,after in mutations:
 directory=out/(label+'-modules');directory.mkdir(exist_ok=True)
 for source in own.glob('*.a'):
  text=re.sub(r"from '(\.\.[^']+)'",lambda m:"from '"+str((source.parent/m[1]).resolve())+"'",source.read_text())
  if source.name==module:assert text.count(before)==1;text=text.replace(before,after)
  (directory/source.name).write_text(text)
 run(label+'-build',[compiler,'build',directory/'source_main.a','-o',out/label,'--tsgo',archives/'checker-asan.a','--sanitize'])
 at=cases[label];assert records[at]['byte_equal'];actual,error=run(label+'-run',[out/label,config,manifest,'default',records[at]['path']]);assert not error and actual!=(out/f'individual-{at:03}-go.stdout').read_bytes()
 print(f'Mutant {label}: compiled, exited zero, clean sanitizer stderr; Go byte comparison catches it on control {at}.',flush=True)
driver=re.sub(r"from '(\.[^']+)'",lambda m:"from '"+str((own/m[1]).resolve())+"'",(own/'source_main.a').read_text());released=out/'released-main.a';released.write_text(driver.replace('let findings = 0;','tsgoRelease(program);\nlet findings = 0;'))
run('remaining-released-build',[compiler,'build',released,'-o',out/'released','--tsgo',archives/'checker-normal.a']);actual,error=run('remaining-released-run',[out/'released',config,manifest],expected=70);assert not actual and error==b'adamic: panic: invalid or released checker handle\n'
print('Released handle rejected with exit 70. Partial checks passed; full source oracle remains FAILED on the unchanged production Go satisfies panic.',flush=True)
raise SystemExit(1)
