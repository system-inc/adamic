from pathlib import Path
import subprocess,shutil,json,hashlib,os
r=Path('/workspace/scratch/99r');repo=Path('/workspace/adamic');adapter=repo/'stage3/adapt/99-correctness-axolotl-r/adapt.cjs';after=r/'after'
brand=r/'brand-mutant';(brand/'src/compiler').mkdir(parents=True)
for name in ['core.ts','checker.ts']:shutil.copyfile(after/'src/compiler'/name,brand/'src/compiler'/name)
p=brand/'src/compiler/core.ts';b=p.read_bytes();old=b'emptyArray as unknown as SortedReadonlyArray<T>';assert b.count(old)==1;p.write_bytes(b.replace(old,b'emptyArray as any as SortedReadonlyArray<T>'))
def snapshot(tree):return {p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in (tree/'src/compiler').iterdir() if p.is_file()}
before=snapshot(brand)
with (r/'brand-mutant.log').open('w') as log:
 child=subprocess.run(['node',str(adapter),str(brand)],cwd=repo,env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules')),stdout=log,stderr=subprocess.STDOUT)
assert child.returncode!=0
assert before==snapshot(brand)
assert 'changed or half-adapted owner' in (r/'brand-mutant.log').read_text()
(r/'brand-mutant-result.json').write_text(json.dumps({'exit':child.returncode,'red':True,'sourceUnchanged':True,'reason':'deduplicateSorted: changed or half-adapted owner'},indent=2)+'\n')
relative=r/'relative-mutant';relative.mkdir();shutil.copytree(after/'src',relative/'src',symlinks=True)
for name in ['package.json','tsconfig.json']:
 if (after/name).exists():shutil.copyfile(after/name,relative/name)
roots=sorted(str(p) for p in (relative/'src/compiler').rglob('*.ts'));assert len(roots)==79
with (r/'relative-mutant-baseline.log').open('w') as log:baseline=subprocess.run([str(r/'bin/adamic'),'types',*roots],stdout=log,stderr=subprocess.STDOUT)
assert baseline.returncode==0,'mutant copy baseline must type-check'
p=relative/'src/compiler/core.ts';b=p.read_bytes();old=b'export function relativeComplement<T>(arrayA: T[], arrayB: T[], comparer: Comparer<T>): T[];\r\n';assert b.count(old)==1;p.write_bytes(b.replace(old,b''))
with (r/'relative-mutant.log').open('w') as log:child=subprocess.run([str(r/'bin/adamic'),'types',*roots],stdout=log,stderr=subprocess.STDOUT)
text=(r/'relative-mutant.log').read_text();assert child.returncode!=0
assert 'checker.ts' in text and ('not assignable' in text or 'TS2322' in text),text[-2000:]
(r/'relative-mutant-result.json').write_text(json.dumps({'baselineExit':baseline.returncode,'roots':79,'exit':child.returncode,'red':True,'mutation':'drop only the defined-array relativeComplement overload','diagnostics':text},indent=2)+'\n')
print('Both required mutants are red')
