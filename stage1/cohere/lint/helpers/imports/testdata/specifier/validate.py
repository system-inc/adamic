#!/usr/bin/env python3
import json, os, shutil, subprocess, hashlib,tempfile
from pathlib import Path
owned=Path(__file__).resolve().parent
repo=owned.parents[6]
scratch=Path(tempfile.mkdtemp(prefix='imports-capture-'))
def run(cmd,label,cwd=repo,env=None):
 with (scratch/(label+'.log')).open('wb') as output:
  result=subprocess.run(list(map(str,cmd)),cwd=cwd,env=env,stdout=output,stderr=subprocess.PIPE)
 (scratch/(label+'.stderr')).write_bytes(result.stderr)
 if result.returncode or result.stderr:raise RuntimeError(f'{label}: exit={result.returncode} {result.stderr[:1200]!r}')
 return (scratch/(label+'.log')).read_bytes()
original=repo/'cohere/internal/lint/ecmascript/imports/source.go'
source=original.read_text()
assert source.count('func SpecifierNode(')==1
source=source.replace('import (','import (\n "encoding/json"\n "os"\n "strconv"',1).replace('func SpecifierNode(','func wave15OriginalSpecifierNode(',1)
modified=scratch/'source.go';modified.write_text(source+'\n'+(owned/'oracle-wrapper.go.txt').read_text())
virtual=original.parent/'wave15_specifier_test.go'
overlay=scratch/'oracle-overlay.json';overlay.write_text(json.dumps({'Replace':{str(original):str(modified),str(virtual):str(owned/'oracle-controls.go.txt')}}))
rows=[];coverage={}
for package,filters,label in [('nexus','^TestBoundaryNoInternalImport','internal'),('nexus','^TestBoundaryNoNexusOutsideImport','outside'),('nexus','^TestBoundaryNoProjectImport','project'),('../ecmascript/imports','^(TestWave15SpecifierControls|TestSpecifierNode|TestHelpersSurviveNilInput)','controls'),('../ecmascript/imports','^TestWave15SpecifierControls','controls-97'),('../ecmascript/imports','^TestWave15SpecifierControls','controls-273'),('../ecmascript/imports','^TestWave15SpecifierControls','controls-214'),('../ecmascript/imports','^TestWave15SpecifierControls','controls-65535')]:
 capture=scratch/(label+'.jsonl');capture.write_text('')
 env=dict(os.environ,WAVE15_SPECIFIER_CAPTURE=str(capture),WAVE15_SPECIFIER_BASE=label.split('-')[-1] if label.startswith('controls-') else '0')
 path='./internal/lint/rules/'+package if package=='nexus' else './internal/lint/ecmascript/imports'
 run(['go','test','-overlay='+str(overlay),path,'-run',filters,'-count=1','-timeout=10m'],'upstream-'+label,repo/'cohere',env)
 records=[json.loads(line) for line in capture.read_text().splitlines()]
 assert records,label+' made no helper observations'
 coverage[label]=len(records);rows.extend(records)
print('Actual Go calls per consumer/control family:',coverage,flush=True)
importKind,callKind=rows[0][6:]
assert all(r[6:]==[importKind,callKind] for r in rows)
production=owned/'../../specifier_node.a'
assert production.exists(),'Initialize specifier_node.a with Go kinds '+str([importKind,callKind])
assert f'const importDeclarationKind = {importKind};' in production.read_text()
assert f'const callExpressionKind = {callKind};' in production.read_text()
cases=scratch/'cases.tsv';cases.write_text(''.join('\t'.join(map(str,r[:5]))+'\n' for r in rows))
expected=''.join(str(r[5])+'\n' for r in rows).encode();(scratch/'Go.log').write_bytes(expected)
builder=repo/'wave15_specifier_build.go';bo=scratch/'build-overlay.json';bo.write_text(json.dumps({'Replace':{str(builder):str(owned/'build.go.txt')}}))
def build(entry,binary,label):run(['go','run','-overlay='+str(bo),builder,entry,binary],label)
entry=owned/'main.a';binary=scratch/'port';build(entry,binary,'build')
for side,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',entry,cases]),('JavaScript',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',str(binary)+'.mjs',cases]),('native',[binary,cases])]:
 actual=run(cmd,side);assert actual==expected,side+' differs from actual Go';print(side+': identical '+str(len(actual))+' bytes',flush=True)
mutant=scratch/'mutant';mutant.mkdir(exist_ok=True)
for p in owned.glob('*.a'):shutil.copyfile(p,mutant/p.name)
shutil.copyfile(owned/'../../specifier_node.a',mutant/'specifier_node.a')
p=mutant/'main.a';p.write_text(p.read_text().replace('../../specifier_node.a','./specifier_node.a'))
p=mutant/'specifier_node.a';s=p.read_text();anchor='kind === callExpressionKind && hasFirstArgument';assert s.count(anchor)==1;p.write_text(s.replace(anchor,'kind === callExpressionKind && !hasFirstArgument'))
build(mutant/'main.a',scratch/'mutant-port','mutant-build')
for side,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',mutant/'main.a',cases]),('JavaScript',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',str(scratch/'mutant-port')+'.mjs',cases]),('native',[scratch/'mutant-port',cases])]:
 actual=run(cmd,'mutant-'+side);assert actual!=expected,'mutant survived';print('Call-argument mutant caught on '+side+' only by output comparison; successful compile and exit',flush=True)
(scratch/'coverage.json').write_text(json.dumps(coverage,indent=2))
(scratch/'outputs.sha256').write_text('\n'.join(hashlib.sha256((scratch/(name+'.log')).read_bytes()).hexdigest()+' '+name+'.log' for name in ['Go','Node','JavaScript','native'])+'\n')
print('PASS '+str(len(rows))+' actual Go helper observations; all three consumer test families pass',flush=True)
