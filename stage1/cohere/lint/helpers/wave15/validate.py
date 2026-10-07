#!/usr/bin/env python3
import hashlib,json,os,shutil,subprocess
from pathlib import Path
owned=Path(__file__).resolve().parent
repo=owned.parents[4]
scratch=Path('/tmp/wave15-helper-variant');scratch.mkdir(exist_ok=True)
def run(cmd,label,cwd=repo,env=None):
 with (scratch/(label+'.log')).open('wb') as output:r=subprocess.run(list(map(str,cmd)),cwd=cwd,env=env,stdout=output,stderr=subprocess.PIPE)
 (scratch/(label+'.stderr')).write_bytes(r.stderr)
 if r.returncode or r.stderr:raise RuntimeError(f'{label}: exit={r.returncode} {r.stderr[:600]!r}')
 return (scratch/(label+'.log')).read_bytes()
virtual=repo/'cohere/wave15_variant.go';seam=repo/'cohere/internal/lint/rules/tailwind/collapse/wave15_variant.go'
overlay=scratch/'oracle-overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(owned/'oracle-main.go.txt'),str(seam):str(owned/'oracle.go.txt')}}))
run(['go','build','-overlay='+str(overlay),'-o',scratch/'oracle',virtual],'oracle-build',repo/'cohere')
run([scratch/'oracle','generate',scratch/'cases.tsv',repo/'cohere/internal/lint/rules/tailwind'],'generate')
coverage=json.loads((scratch/'cases.tsv.coverage.json').read_text());assert len(coverage)==6 and all(coverage.values());print('Consumer literals:',coverage,flush=True)
builder=repo/'wave15_helper_build.go';bo=scratch/'build-overlay.json';bo.write_text(json.dumps({'Replace':{str(builder):str(owned/'build.go.txt')}}))
def build(entry,binary,label):run(['go','run','-overlay='+str(bo),builder,entry,binary],label)
entry=owned/'main.a';binary=scratch/'port';build(entry,binary,'port-build')
expected=run([scratch/'oracle',scratch/'cases.tsv'],'Go')
for side,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',entry,scratch/'cases.tsv']),('JavaScript',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',str(binary)+'.mjs',scratch/'cases.tsv']),('native',[binary,scratch/'cases.tsv'])]:
 actual=run(cmd,side);assert actual==expected,side+' output differs';print(side+': byte-identical '+str(len(actual))+' bytes',flush=True)
mutant=scratch/'mutant';mutant.mkdir(exist_ok=True)
for p in owned.glob('*.a'):shutil.copyfile(p,mutant/p.name)
p=mutant/'variant_kind.a';s=p.read_text();assert s.count("?? 'static'")==1;p.write_text(s.replace("?? 'static'","?? 'functional'"))
build(mutant/'main.a',scratch/'mutant-port','mutant-build')
for side,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',mutant/'main.a',scratch/'cases.tsv']),('JavaScript',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',str(scratch/'mutant-port')+'.mjs',scratch/'cases.tsv']),('native',[scratch/'mutant-port',scratch/'cases.tsv'])]:
 actual=run(cmd,'mutant-'+side);assert actual!=expected,'mutant survived';print('Fallback mutant caught on '+side+' by output comparison only',flush=True)
filters='^(TestEnforceCanonicalClasses|TestEnforceConsistentClassOrder|TestEnforceConsistentVariantOrder|TestEnforceShorthandClasses|TestNoConflictingClasses|TestNoUnknownClasses)'
run(['go','test','./internal/lint/rules/tailwind','-count=1','-run',filters,'-timeout=10m'],'consumers-upstream',repo/'cohere')
print('Queries '+str(len(expected.splitlines()))+'; every consumer fixture file represented; upstream consumer tests PASS',flush=True)
