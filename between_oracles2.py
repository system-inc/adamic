import pathlib,subprocess,json,os,hashlib
repo=pathlib.Path('/workspace/adamic');out=pathlib.Path(__file__).parent;base=pathlib.Path('/workspace/scratch');env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE=str(base/'typescript-6.0.3'));rows=[]
def run(name,args,cwd=repo):
 with (out/(name+'.log')).open('w') as log:child=subprocess.run(args,cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT)
 rows.append(dict(name=name,argv=args,exit=child.returncode));(out/'between-oracles2-commands.json').write_text(json.dumps(rows,indent=2)+'\n');print(name,'child exit='+str(child.returncode),flush=True)
 if child.returncode:raise SystemExit(child.returncode)
run('guard-fix-commit',['git','add','stage3/adapt/47-host-errors/tracing-fs-import.cjs']);run('guard-fix-amend',['git','commit','--amend','--no-edit'])
run('after3-apply',['bash','stage3/apply.sh',str(base/'tracing-after3')])
a=base/'tracing-after2';b=base/'tracing-after3';files=sorted(p.relative_to(a) for p in (a/'src').rglob('*') if p.is_file());assert files==sorted(p.relative_to(b) for p in (b/'src').rglob('*') if p.is_file());assert all((a/p).read_bytes()==(b/p).read_bytes() for p in files)
assert (a/'package-lock.json').read_bytes()==(b/'package-lock.json').read_bytes();(b/'node_modules').symlink_to(a/'node_modules',target_is_directory=True)
run('after3-build',['npm','run','build'],b);run('after3-harness',['npm','exec','hereby','--','tests'],b)
names=sorted(p.relative_to(a/'built/local') for p in (a/'built/local').rglob('*.js'));assert names==sorted(p.relative_to(b/'built/local') for p in (b/'built/local').rglob('*.js'));assert all((a/'built/local'/p).read_bytes()==(b/'built/local'/p).read_bytes() for p in names)
roots=sorted(str(p) for p in (b/'src/compiler').rglob('*.ts'));assert len(roots)==79;run('after3-types',[str(base/'tracing2-adamic'),'types',*roots])
run('after3-replay',['bash','stage3/apply.sh',str(base/'tracing-after3-replay')]);replay=base/'tracing-after3-replay';assert all((b/p).read_bytes()==(replay/p).read_bytes() for p in files)
(out/'after3-identity.json').write_text(json.dumps(dict(sourceFiles=len(files),sourceIdenticalToAfter2=True,emittedFiles=len(names),emittedByteIdenticalToAfter2=True,officialReplayByteIdentical=True,guardChange='boolean identity assertion avoids formatting cyclic symbol objects on rejection; generated TypeScript unchanged'),indent=2)+'\n')
