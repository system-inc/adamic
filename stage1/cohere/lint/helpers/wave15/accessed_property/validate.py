#!/usr/bin/env python3
import json,os,shutil,subprocess,hashlib
from pathlib import Path
owned=Path(__file__).resolve().parent;repo=owned.parents[5]
scratch=Path('/tmp/wave15-accessed-property');scratch.mkdir(exist_ok=True)
def run(cmd,label,cwd=repo,env=None):
 with (scratch/(label+'.log')).open('wb') as log:r=subprocess.run(list(map(str,cmd)),cwd=cwd,env=env,stdout=log,stderr=subprocess.PIPE)
 (scratch/(label+'.stderr')).write_bytes(r.stderr)
 if r.returncode or r.stderr:raise RuntimeError(f'{label}: exit {r.returncode}: {r.stderr[:1200]!r}')
 return (scratch/(label+'.log')).read_bytes()
original=repo/'cohere/internal/lint/ecmascript/property/name.go';source=original.read_text();assert source.count('func AccessedName(')==1
source=source.replace('import (','import (\n "encoding/json"\n "os"\n "sync"',1).replace('func AccessedName(','func wave15OriginalAccessedName(',1)
source=source.replace('return Name(node.AsPropertyAccessExpression().Name(), accept)','return wave15AccessDependencyName(node.AsPropertyAccessExpression().Name(), accept)',1).replace('return Name(argument, accept)','return wave15AccessDependencyName(argument, accept)',1)
modified=scratch/'name.go';modified.write_text(source+'\n'+(owned/'oracle-wrapper.go.txt').read_text());virtual=original.parent/'wave15_accessed_test.go'
overlay=scratch/'oracle-overlay.json';overlay.write_text(json.dumps({'Replace':{str(original):str(modified),str(virtual):str(owned/'oracle-controls.go.txt')}}))
rows=[];coverage={}
for package,filters,label in [('core','^TestNoPrototypeBuiltins','prototype'),('react','^TestNoRenderReturnValue','render'),('core','^TestYoda','yoda'),('property','^(TestAccessedName|TestNilIsNotAName|TestWave15AccessedControls)','controls')]:
 capture=scratch/(label+'.jsonl');capture.write_text('');env=dict(os.environ,WAVE15_ACCESSED_CAPTURE=str(capture))
 path='./internal/lint/rules/'+package if package!='property' else './internal/lint/ecmascript/property'
 run(['go','test','-overlay='+str(overlay),path,'-run',filters,'-count=1','-timeout=10m'],'upstream-'+label,repo/'cohere',env)
 records=[json.loads(line) for line in capture.read_text().splitlines()];assert records,label+' made no observations';coverage[label]=len(records);rows+=records
print('Actual Go calls:',coverage,flush=True)
kinds=rows[0]['Kinds'];assert all(r['Kinds']==kinds for r in rows)
production=owned/'accessed_name.a';assert production.exists(),'Initialize accessed_name.a with Go kinds '+str(kinds)
hextext=lambda s:s.encode('utf-16-be',errors='surrogatepass').hex()
cases=scratch/'cases.tsv';cases.write_text(''.join('\t'.join([str(int(r['Present'])),str(r['Kind']),str(r['KeyHandle']),str(r['KeyKind']),str(r['Accept']),hextext(r['NameText']),str(int(r['NameOK']))])+'\n' for r in rows))
expected=''.join('\t'.join([str(int(r['OK'])),hextext(r['Text']),str(r['Calls']),str(r['DelegatedAccept']),str(r['DelegatedKey'])])+'\n' for r in rows).encode();(scratch/'Go.log').write_bytes(expected)
builder=repo/'wave15_accessed_build.go';bo=scratch/'build-overlay.json';bo.write_text(json.dumps({'Replace':{str(builder):str(owned/'build.go.txt')}}))
def build(entry,binary,label):run(['go','run','-overlay='+str(bo),builder,entry,binary],label)
entry=owned/'main.a';binary=scratch/'port';build(entry,binary,'build')
for side,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',entry,cases]),('JavaScript',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',str(binary)+'.mjs',cases]),('native',[binary,cases])]:
 actual=run(cmd,side);assert actual==expected,side+' differs from Go';print(side+': identical '+str(len(actual))+' bytes',flush=True)
mutant=scratch/'mutant';mutant.mkdir(exist_ok=True)
for p in owned.glob('*.a'):shutil.copyfile(p,mutant/p.name)
p=mutant/'accessed_name.a';s=p.read_text();anchor='keyKind !== identifierKind && keyKind !== privateIdentifierKind';assert s.count(anchor)==1;p.write_text(s.replace(anchor,'keyKind !== -123 && keyKind !== -124'))
build(mutant/'main.a',scratch/'mutant-port','mutant-build')
for side,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',mutant/'main.a',cases]),('JavaScript',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',str(scratch/'mutant-port')+'.mjs',cases]),('native',[scratch/'mutant-port',cases])]:
 actual=run(cmd,'mutant-'+side);assert actual!=expected,'mutant survived';print('Variable-subscript mutant caught on '+side+' only by output comparison; successful compile and normal exit',flush=True)
(scratch/'coverage.json').write_text(json.dumps(coverage,indent=2));(scratch/'outputs.sha256').write_text('\n'.join(hashlib.sha256((scratch/(n+'.log')).read_bytes()).hexdigest()+' '+n+'.log' for n in ['Go','Node','JavaScript','native'])+'\n')
print('PASS '+str(len(rows))+' actual Go observations; three consumer test families pass',flush=True)
