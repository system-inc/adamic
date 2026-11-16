from pathlib import Path
import subprocess as s,json,shutil
D=Path('/workspace/scratch/b7phwab-amend5');P=D/'work/stage3/adapt/99-wasm-crypto-import';tree=D/'final-composed';file=tree/'src/compiler/sys.ts';saved=file.read_bytes();rows=[]
def run(name,args,expected):
 with (D/(name+'.log')).open('w') as f:p=s.run(list(map(str,args)),stdout=f,stderr=s.STDOUT)
 row=dict(name=name,command=list(map(str,args)),exit=p.returncode,expected=expected);rows.append(row);(D/'extra-mutants.json').write_text(json.dumps(rows,indent=2)+'\n');print(json.dumps(row),flush=True);assert p.returncode==expected,name
for name,text in [('moved-owner',saved.replace(b'const isMacOs = process.platform === "darwin";',b'const isMacOs = process.platform === "darw1n";')),('moved-imports',saved+b'\r\nimport * as cryptoScopedProbe from "node:util";\r\n')]:
 try:
  assert text!=saved;file.write_bytes(text);mt=file.stat().st_mtime_ns
  run(name,['node',P/'adapt.cjs',tree],1);assert file.read_bytes()==text and file.stat().st_mtime_ns==mt
 finally:file.write_bytes(saved)
try:
 text=saved.replace(b'return process.stdout.columns ?? 0;',b'return process.stdout.columns ?? 1;');assert text!=saved;file.write_bytes(text)
 run('runtime-scope-mutant',['node',P/'proofs/check.cjs',D/'before99',tree],1)
finally:file.write_bytes(saved)
files=[tree/p for p in ['src/compiler/sys.ts','src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts','src/harness/watchUtils.ts']];old={p:p.read_bytes() for p in files}
for mode in ['unnormalized','optional-contract']:
 try:
  run('watch-directory-'+mode,['node',P/'watch-directory-mutant.cjs',tree,mode],0)
  run('watch-directory-'+mode+'-refused',['node',P/'adapt.cjs',tree],1)
 finally:
  for p,b in old.items():p.write_bytes(b)
probe=D/'span-disabled';shutil.copytree(D/'before99/src',probe/'src')
script="""const fs=require('node:fs'),path=require('node:path'),Module=require('node:module');const filename=process.argv[1],tree=process.argv[2],rr=require(path.join(path.dirname(filename),'recordreplay.cjs'));const line=\"        assert(inOwner || importInsertion, 'edit outside getNodeSystem and pinned import edits');\";const source=fs.readFileSync(filename,'utf8');if(source.split(line).length!==2)throw Error('span guard moved');const old=rr.edits;rr.edits=(...args)=>[...old(...args),{start:args[0].text.length,end:args[0].text.length,text:'\\nexport const extra = true;\\n'}];const m=new Module(filename,module);m.filename=filename;m.paths=Module._nodeModulePaths(path.dirname(filename));m._compile(source.replace(line,'        // Deliberate mutation: no span guard.'),filename);console.log(JSON.stringify(m.exports.adapt(tree)));"""
run('span-disabled-control',['node','-e',script,P/'adapt.cjs',probe],0)
assert (probe/'src/compiler/sys.ts').read_bytes()==(D/'planned-after.ts').read_bytes()+b'\nexport const extra = true;\n'
run('span-disabled-scope-refused',['node',P/'proofs/check.cjs',D/'before99',probe],1)
assert file.read_bytes()==saved
