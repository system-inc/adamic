from pathlib import Path
import subprocess,json
repo=Path('/workspace/adamic');scratch=Path('/workspace/no-useless-catch-options');root=repo/'stage1/cohere/lint';cohere=repo/'cohere'

def run(label,args,cwd=repo,success=True):
 p=subprocess.run(args,cwd=cwd,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
 (scratch/(label+'.stdout')).write_bytes(p.stdout);(scratch/(label+'.stderr')).write_bytes(p.stderr)
 (scratch/(label+'.exit')).write_text(str(p.returncode)+'\n')
 if success and p.returncode:raise RuntimeError(label+': '+p.stderr.decode(errors='replace'))
 return p

replacements={};virtual=[]
def add(name,source):
 path=str(cohere/('adamic_lint_'+name+'.go'));replacements[path]=str(source);virtual.append(path)
add('oracle',root/'testdata/oracle.go');add('registry',root/'.generated/registry.go')
for directory in sorted((root/'rules').iterdir()):
 if (directory/'rule.json').exists():add(directory.name.replace('-','_'),directory/'oracle.go')
overlay=scratch/'overlay.json';overlay.write_text(json.dumps({'Replace':replacements}))
run('go-build',['go','build','-overlay='+str(overlay),'-o',str(scratch/'oracle'),*virtual],cohere)
run('native-build',['go','run','./cmd/adamic','build',str(root/'main.ts'),'-o',str(scratch/'native'),'--sanitize'])
emitted=run('js-build',['go','run','./cmd/adamic','js',str(root/'main.ts')]);(scratch/'driver.mjs').write_bytes(emitted.stdout)
runtime=scratch/'node_modules/adamic';runtime.mkdir(parents=True,exist_ok=True)
(runtime/'package.json').write_text('{"type":"module","exports":"./index.mjs"}');(runtime/'index.mjs').write_bytes((repo/'oracle/adamic.mjs').read_bytes())
fixture=scratch/'witness.a';fixture.write_bytes((root/'rules/no-useless-catch/testdata/rethrow.ts.txt').read_bytes())
options=['{}','{"Unused":true}','{"Nested":{"unused":[1,true,"x"]}}','null']
manifest=scratch/'configured.manifest';manifest.write_text('\n'.join(str(fixture)+'\tno-useless-catch\t\t\tfalse\t'+option for option in options)+'\n')
commands={'go':[str(scratch/'oracle')],'node':['node','--disable-warning=ExperimentalWarning',str(repo/'oracle/node.mjs'),str(root/'main.ts')],'javascript':['node',str(scratch/'driver.mjs')],'native':[str(scratch/'native')]}
truth=None
for label,args in commands.items():
 p=run(label,args+['--manifest',str(manifest)])
 if truth is None:truth=p.stdout
 assert p.stdout==truth,label+' configured rows differ'
 assert p.stderr==b'',label+' unexpected stderr'
print('four configured JSON rows: Go, Node, emitted JavaScript and sanitized native identical:',len(truth),'bytes')
source=(root/'rules/no-useless-catch/oracle.go').read_text();assert source.count('return options')==1
mutated=scratch/'mutant-oracle.go';mutated.write_text(source.replace('return options','return nil'))
replacements[str(cohere/'adamic_lint_no_useless_catch.go')]=str(mutated)
mutantOverlay=scratch/'mutant-overlay.json';mutantOverlay.write_text(json.dumps({'Replace':replacements}))
run('guard-mutant-build',['go','build','-overlay='+str(mutantOverlay),'-o',str(scratch/'guard-mutant'),*virtual],cohere)
p=run('guard-mutant',[str(scratch/'guard-mutant'),'--manifest',str(manifest)],success=False)
assert p.returncode!=0 and b'no-useless-catch: options {} reached an adapter that decodes none' in p.stderr
print('nil-options adapter mutant compiles and the unchanged shared guard catches its configured row')
