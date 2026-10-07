#!/usr/bin/env python3
"""Rebuild the standalone cache and cleanly merged string-views runtime variants."""
import io,pathlib,shutil,subprocess,tarfile
root=pathlib.Path.cwd();scratch=pathlib.Path('/tmp/decode-ascii')
base='4eb5187c2faa312d1945c7a439d60ad27bfeee5b';views='ba9c9ef87e38d5369e0eaa1b0b6d24f93bbd227a'
tree=subprocess.check_output(['git','merge-tree','--write-tree',base,views],text=True).splitlines()[0]
archive=subprocess.check_output(['git','archive',tree,'internal/native/runtime'])
for label in ['only','combined']:(scratch/f'cache-{label}-runtime').mkdir(parents=True,exist_ok=True)
with tarfile.open(fileobj=io.BytesIO(archive)) as tar:
    for item in tar:
        if item.isfile() and pathlib.Path(item.name).parent==pathlib.Path('internal/native/runtime'):
            (scratch/'cache-combined-runtime'/pathlib.Path(item.name).name).write_bytes(tar.extractfile(item).read())
for source in (root/'internal/native/runtime').iterdir():
    if source.suffix in ('.c','.h'):shutil.copy2(source,scratch/'cache-only-runtime'/source.name)
shutil.copy2(root/'internal/native/runtime/input.c',scratch/'cache-combined-runtime/input.c')
subprocess.run(['go','build','-o',str(scratch/'build-service'),'cloud/reports/decode-ascii/build.go'],check=True)
# build.sh prepares the unchanged service/host/command sources; it is not run during measurement.
for label in ['only','combined']:
    runtime=scratch/f'cache-{label}-runtime'
    for target,source,output in [('wasi','service.a',f'cache-{label}.wasm'),('native','command.a',f'cache-{label}-native')]:
        subprocess.run([str(scratch/'build-service'),str(runtime),str(scratch/'service'/source),target,str(scratch/output)],check=True)
# The existing harness calls these two variants before/after; preserve historical named files.
for stem in ['service-before.wasm','service-after.wasm','native-before','native-after']:
    path=scratch/stem
    if path.exists():shutil.copy2(path,scratch/('prior-'+stem))
for source,destination in [('cache-only.wasm','service-before.wasm'),('cache-combined.wasm','service-after.wasm'),('cache-only-native','native-before'),('cache-combined-native','native-after')]:
    shutil.copy2(scratch/source,scratch/destination)
print('before = cache alone; after = cache plus string-views; merged tree '+tree)
