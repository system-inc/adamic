#!/usr/bin/env python3
import json,os,shutil,subprocess,hashlib,tempfile
from pathlib import Path
owned=Path(__file__).resolve().parent;repo=owned.parents[6]
scratch=Path(tempfile.mkdtemp(prefix='imports-capture-'))
def run(cmd,label,cwd=repo,env=None):
 with (scratch/(label+'.log')).open('wb') as log:r=subprocess.run(list(map(str,cmd)),cwd=cwd,env=env,stdout=log,stderr=subprocess.PIPE)
 (scratch/(label+'.stderr')).write_bytes(r.stderr)
 if r.returncode or r.stderr:raise RuntimeError(f'{label}: exit {r.returncode}: {r.stderr[:1200]!r}')
 return (scratch/(label+'.log')).read_bytes()
original=repo/'cohere/internal/lint/ecmascript/imports/source.go';source=original.read_text()
assert source.count('func SourceVisitors(')==1
source=source.replace('import (','import (\n "encoding/json"\n "os"\n "sync"',1).replace('func SourceVisitors(','func wave15OriginalSourceVisitors(',1)
source=source.replace('source, isImport := CallExpressionSource(node)','source, isImport := wave15DelegatedCallSource(node)',1)
modified=scratch/'source.go';modified.write_text(source+'\n'+(owned/'oracle-wrapper.go.txt').read_text())
virtual=original.parent/'wave15_source_visitors_test.go';overlay=scratch/'oracle-overlay.json';overlay.write_text(json.dumps({'Replace':{str(original):str(modified),str(virtual):str(owned/'oracle-controls.go.txt')}}))
rows=[];coverage={}
for package,filters,label in [('nexus','^TestBoundaryNoInternalImport','internal'),('nexus','^TestBoundaryNoNexusOutsideImport','outside'),('nexus','^TestBoundaryNoProjectImport','project'),('imports','^(TestSourceVisitors|TestWave15SourceVisitorControls)','controls')]:
 capture=scratch/(label+'.jsonl');capture.write_text('')
 env=dict(os.environ,WAVE15_VISITORS_CAPTURE=str(capture));path='./internal/lint/rules/nexus' if package=='nexus' else './internal/lint/ecmascript/imports'
 run(['go','test','-overlay='+str(overlay),path,'-run',filters,'-count=1','-timeout=10m'],'upstream-'+label,repo/'cohere',env)
 records=[json.loads(line) for line in capture.read_text().splitlines()];assert records,label+' made no calls';coverage[label]=len(records);rows+=records
print('Actual Go listener calls:',coverage,flush=True)
hextext=lambda s:s.encode('utf-16-be',errors='surrogatepass').hex()
cases=scratch/'cases.tsv';cases.write_text(''.join('\t'.join([str(r['Kind']),str(int(r['ModulePresent'])),str(int(r['ModuleLiteral'])),hextext(r['ModuleText']),str(int(r['CallMatches'])),hextext(r['CallSource']),str(int(r['Nil']))])+'\n' for r in rows))
expected=''
for i,r in enumerate(rows):
 assert r['Kind'] in [274,215]
 for source in r['Sources'] or []:expected+=f'{i}:report:{hextext(source)}@{i}\n'
 expected+=f'{i}:counts:{len(r["Sources"] or [])}:{r["Delegations"]}\n'
expected=expected.encode();(scratch/'Go.log').write_bytes(expected)
builder=repo/'wave15_visitors_build.go';bo=scratch/'build-overlay.json';bo.write_text(json.dumps({'Replace':{str(builder):str(owned/'build.go.txt')}}))
def build(entry,binary,label):run(['go','run','-overlay='+str(bo),builder,entry,binary],label)
entry=owned/'main.a';binary=scratch/'port';build(entry,binary,'build')
for side,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',entry,cases]),('JavaScript',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',str(binary)+'.mjs',cases]),('native',[binary,cases])]:
 actual=run(cmd,side);assert actual==expected,side+' differs from Go';print(side+': identical '+str(len(actual))+' bytes',flush=True)
mutant=scratch/'mutant';mutant.mkdir(exist_ok=True)
for p in owned.glob('*.a'):shutil.copyfile(p,mutant/p.name)
shutil.copyfile(owned/'../../source_visitors.a',mutant/'source_visitors.a')
p=mutant/'main.a';p.write_text(p.read_text().replace('../../source_visitors.a','./source_visitors.a'))
p=mutant/'source_visitors.a';s=p.read_text();anchor='if(!result.matches)';assert s.count(anchor)==1;p.write_text(s.replace(anchor,'if(result.matches)'))
build(mutant/'main.a',scratch/'mutant-port','mutant-build')
for side,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',mutant/'main.a',cases]),('JavaScript',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',str(scratch/'mutant-port')+'.mjs',cases]),('native',[scratch/'mutant-port',cases])]:
 actual=run(cmd,'mutant-'+side);assert actual!=expected,'mutant survived';print('Call-result gate mutant caught on '+side+' only by output comparison; normal exit and successful compile',flush=True)
(scratch/'coverage.json').write_text(json.dumps(coverage,indent=2));(scratch/'outputs.sha256').write_text('\n'.join(hashlib.sha256((scratch/(n+'.log')).read_bytes()).hexdigest()+' '+n+'.log' for n in ['Go','Node','JavaScript','native'])+'\n')
print('PASS '+str(len(rows))+' Go listener observations; three upstream consumer test families pass',flush=True)
