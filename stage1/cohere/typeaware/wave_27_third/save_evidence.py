"""Preserve complete isolated gate evidence, without modifying shared harnesses."""
from pathlib import Path
import sys,json,shutil,hashlib,gzip,subprocess
repo=Path(__file__).resolve().parents[4]
source=Path(__file__).resolve().parent
art=Path(sys.argv[1]);out=source/'evidence';out.mkdir(exist_ok=True)
for p in sorted(art.iterdir()):
 if p.name[:3].isdigit() and p.suffix in ['.stdout','.stderr']:
  data=p.read_bytes()
  if len(data)>100000:(out/(p.name+'.gz')).write_bytes(gzip.compress(data,mtime=0))
  else:(out/p.name).write_bytes(data)
for name in ['measurements.json','excluded-controls.json','control-candidates.manifest','controls.manifest','tsconfig.json']:
 shutil.copyfile(art/name,out/name)
paths=(art/'control-candidates.manifest').read_text().splitlines()
(out/'control-sources.json').write_text(json.dumps({p:Path(p).read_text() for p in paths},indent=2)+'\n')
workspace=repo.parent
for p in sorted(workspace.glob('wave-27-third-*.log')):
 shutil.copyfile(p,out/p.name)
for name in ['wave-27-third-selection.json','wave-27-setup.log']:
 shutil.copyfile(workspace/name,out/name)
(out/'origin-heads.txt').write_bytes(subprocess.check_output(['git','for-each-ref','--format=%(refname) %(objectname)','refs/remotes/origin'],cwd=repo))
for name in ['compiler','repository']:
 shutil.copyfile(repo/'stage1/cohere/typeaware/validation-coverage'/(name+'.manifest'),out/(name+'.manifest'))
 manifest=workspace/'wave-27-scratch'/(name+'.manifest')
 files=[Path(p) for p in manifest.read_text().splitlines()]
 (out/(name+'-source-hashes.json')).write_text(json.dumps({str(p):hashlib.sha256(p.read_bytes()).hexdigest() for p in files},indent=2)+'\n')
sources=list(source.glob('*.a'))+list(source.glob('*.py'))+[source/'oracle.go.txt']
sources+=list((source.parent/'wave_27_next').glob('*.a'))
sources+=list(source.parent.glob('*.ts'))+list(source.parent.glob('*.a'))
sources+=list((repo/'stage1/typescript/parser').glob('*.ts'))+list((repo/'stage1/typescript/scanner').glob('*.ts'))
sources+=list((repo/'bridge/tsgo/checker').glob('*.go'))
(out/'source-hashes.json').write_text(json.dumps({str(p.relative_to(repo)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(set(sources))},indent=2)+'\n')
(out/'SHA256.json').write_text(json.dumps({p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(out.iterdir()) if p.is_file() and p.name!='SHA256.json'},indent=2)+'\n')
print(len(list(out.iterdir())),'evidence files')
