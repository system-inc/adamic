#!/usr/bin/env python3
import hashlib,json,shutil,subprocess
from pathlib import Path
owned=Path(__file__).resolve().parent
repo=owned.parents[5]
scratch=Path('/tmp/wave15-regexp-flags');scratch.mkdir(exist_ok=True)
def run(cmd,label,cwd=repo):
 with (scratch/(label+'.log')).open('wb') as out:
  result=subprocess.run(list(map(str,cmd)),cwd=cwd,stdout=out,stderr=subprocess.PIPE)
 (scratch/(label+'.stderr')).write_bytes(result.stderr)
 if result.returncode or result.stderr:raise RuntimeError(f'{label}: exit={result.returncode} {result.stderr[:800]!r}')
 return (scratch/(label+'.log')).read_bytes()
virtual=repo/'cohere/wave15_regexp_flags.go';seam=repo/'cohere/internal/lint/ecmascript/regexp/wave15_flags.go'
overlay=scratch/'oracle-overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(owned/'oracle-main.go.txt'),str(seam):str(owned/'oracle.go.txt')}}))
run(['go','build','-overlay='+str(overlay),'-o',scratch/'oracle',virtual],'oracle-build',repo/'cohere')
run([scratch/'oracle','tables',scratch/'printable_ranges.a'],'tables')
assert (scratch/'printable_ranges.a').read_bytes()==(owned/'printable_ranges.a').read_bytes(),'Go printability table drift'
run([scratch/'oracle','generate',scratch/'cases.tsv',repo/'cohere/internal/lint/rules'],'generate')
coverage=json.loads((scratch/'cases.tsv.coverage.json').read_text());assert len(coverage)==4 and all(coverage.values());print('Consumer fixture strings:',coverage,flush=True)
builder=repo/'wave15_regexp_flags_build.go';bo=scratch/'build-overlay.json';bo.write_text(json.dumps({'Replace':{str(builder):str(owned/'build.go.txt')}}))
def build(entry,binary,label):run(['go','run','-overlay='+str(bo),builder,entry,binary],label)
entry=owned/'main.a';binary=scratch/'port';build(entry,binary,'port-build')
expected=run([scratch/'oracle',scratch/'cases.tsv'],'Go')
for side,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',entry,scratch/'cases.tsv']),('JavaScript',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',str(binary)+'.mjs',scratch/'cases.tsv']),('native',[binary,scratch/'cases.tsv'])]:
 actual=run(cmd,side);assert actual==expected,side+' output differs';print(side+': byte-identical '+str(len(actual))+' bytes',flush=True)
mutant=scratch/'mutant';mutant.mkdir(exist_ok=True)
for p in owned.glob('*.a'):shutil.copyfile(p,mutant/p.name)
p=mutant/'parse_flags.a';s=p.read_text();anchor='if((seen & bit) !== 0)';assert s.count(anchor)==1;p.write_text(s.replace(anchor,'if((seen & bit) < 0)'))
build(mutant/'main.a',scratch/'mutant-port','mutant-build')
witness=scratch/'mutant.tsv';witness.write_text(''.join('x'+s.encode('utf-16-be').hex()+'\n' for s in ['ii','imm','gg','dd','imsuy','iim']))
truth=run([scratch/'oracle',witness],'mutant-Go')
for side,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',mutant/'main.a',witness]),('JavaScript',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',str(scratch/'mutant-port')+'.mjs',witness]),('native',[scratch/'mutant-port',witness])]:
 actual=run(cmd,'mutant-'+side);assert actual!=truth,'mutant survived';print('Duplicate-flag mutant caught on '+side+' only by output comparison; successful compile and exit',flush=True)
for package,filters in [('next','^TestNoHtmlLinkForPages'),('typescript','^TestNoEmptyObjectType'),('core','^TestNoRestricted(Exports|Imports)'),('../ecmascript/regexp','.')]:
 path='./internal/lint/rules/'+package if not package.startswith('../') else './internal/lint/ecmascript/regexp'
 run(['go','test',path,'-count=1','-run',filters,'-timeout=10m'],'upstream-'+package.split('/')[-1],repo/'cohere')
print('Queries '+str(len(expected.splitlines()))+'; every consumer fixture string represented; upstream Go rule and regexp tests PASS',flush=True)
(scratch/'outputs.sha256').write_text('\n'.join(hashlib.sha256((scratch/(n+'.log')).read_bytes()).hexdigest()+'  '+n+'.log' for n in ['Go','Node','JavaScript','native'])+'\n')
