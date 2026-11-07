import hashlib,json,shutil,tarfile,subprocess
from collections import Counter
from pathlib import Path
out=Path(__file__).resolve().parent;repo=out.parents[2];s=Path('/workspace/scratch/elementat-declaration')
read=lambda p:json.loads(p.read_text())
assert read(out/'lane-summary.json')['complete'] and read(out/'reconciliation.json')['complete']
def source(tree):return {str(p.relative_to(tree/'src')):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree/'src').rglob('*')) if p.is_file()}
original=read(s/'source-manifest.json');assert source(s/'baseline')==original
candidate=source(s/'candidate');changed=[name for name in sorted(set(original)|set(candidate)) if original.get(name)!=candidate.get(name)];assert changed==['compiler/core.ts']
assert candidate['compiler/core.ts']==read(s/'source-delta.json')['candidate']
actual={str(p.relative_to(s/'counted')):hashlib.sha256(p.read_bytes()).hexdigest() for base in ['src','built/local'] for p in sorted((s/'counted'/base).rglob('*')) if p.is_file()}
assert actual==read(s/'counted-manifest.json')
for name,digest in read(out/'instrumentation.json')['observerHashes'].items():assert hashlib.sha256((out/name).read_bytes()).hexdigest()==digest,name
emit=read(out/'declaration-emit.json');assert emit['compilerBuildExit']==0
assert all(emit['files'][name]['identical'] for name in ['typescript.d.ts','tsserverlibrary.d.ts'])
diff=emit['files']['typescript.internal.d.ts']['diff'];added=Counter(line[1:] for line in diff.splitlines() if line.startswith('+') and not line.startswith('+++'));removed=Counter(line[1:] for line in diff.splitlines() if line.startswith('-') and not line.startswith('---'))
addOnly=added-removed;removeOnly=removed-added;assert sum(addOnly.values())==sum(removeOnly.values())==1 and all('elementAt' in line for line in [*addOnly,*removeOnly])
production=subprocess.check_output(['git','diff','--name-only','origin/main','--','.',':(exclude)stage3/census/elementat-declaration'],cwd=repo,text=True).splitlines();assert production==[],production
validation={'baselineSourceFrozen':True,'sourceFileCount':len(original),'candidateChangedSourceFiles':changed,'candidateSourceFrozenAfterBuildAndCensus':True,'observedSourceAndBuildFrozen':True,'observerHashesFrozen':True,'productionTrackedDiff':production,'publicDeclarationsIdentical':True,'internalDeclarationOnlyElementAtKindChanged':True,'internalDiffHunks':diff.count('\n@@'),'internalAddedSignature':list(addOnly),'internalRemovedSignature':list(removeOnly)}
(out/'validation.json').write_text(json.dumps(validation,indent=2)+'\n')
raw=s/'retained-raw';raw.mkdir()
# Exact stdout/stderr files and status records, including rejected preparation attempts.
for p in s.iterdir():
 if p.is_file() and p.name!='p2-javascript-build.log' and (p.suffix in ['.log','.json','.jsonl','.txt','.ts','.py']):shutil.copyfile(p,raw/p.name)
for dirname in ['latent-overlay','entry-overlay','excluded-seeds','initial-probe-setup','second-probe-setup','value-only-probes']:
 if (s/dirname).exists():shutil.copytree(s/dirname,raw/dirname,symlinks=False,ignore=shutil.ignore_patterns('p2-javascript-build.log'))
probe=s/'dependency-probe-results'
if probe.exists():shutil.copytree(probe,raw/'excluded-dependency-probe',ignore=shutil.ignore_patterns('adapted-tree'))
inputs=raw/'source-inputs';inputs.mkdir()
for side in ['baseline','candidate','counted']:shutil.copyfile(s/side/'src/compiler/core.ts',inputs/(side+'-core.ts.txt'))
for side in ['baseline','counted']:
 for name in ['compiler/transformers/es2015.ts','services/codefixes/convertToAsyncFunction.ts']:
  shutil.copyfile(s/side/'src'/name,inputs/(side+'-'+name.replace('/','_')+'.txt'))
shutil.copyfile(repo/'stage3/adapt/99-correctness-axolotl-z/owners.json',inputs/'99z-owners.json')
shutil.copyfile(repo/'stage3/lane/shard-plan.json',raw/'shard-plan.json')
for name in ['functions.go','modules.go']:shutil.copyfile(repo/'internal/lower'/name,inputs/(name+'.txt'))
for p in raw.rglob('*.go'):raise AssertionError('Go source must have .go.txt name: '+str(p))
files=sorted(p for p in raw.rglob('*') if p.is_file());(raw/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+str(p.relative_to(raw))+'\n' for p in files))
archive=out/'raw.tar.gz'
with tarfile.open(archive,'w:gz') as tar:tar.add(raw,arcname='raw')
(out/'raw.sha256').write_text(hashlib.sha256(archive.read_bytes()).hexdigest()+'  '+archive.name+'\n')
print(json.dumps({'validation':validation,'rawArchiveBytes':archive.stat().st_size,'rawArchiveSha256':hashlib.sha256(archive.read_bytes()).hexdigest()}),flush=True)
