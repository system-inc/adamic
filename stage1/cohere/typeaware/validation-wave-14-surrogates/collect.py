import gzip, hashlib, json, shutil, subprocess
from pathlib import Path
repo = Path('/workspace/adamic')
root = Path('/workspace/wave-14-resume')
source = root / 'final'
out = repo / 'stage1/cohere/typeaware/validation-wave-14-surrogates'
out.mkdir(exist_ok=True)
(out / '.gitattributes').write_text('controls/*.a -text whitespace=cr-at-eol\n*.log -text\n')
for p in source.glob('*.stdout'):
    (out / (p.name + '.gz')).write_bytes(gzip.compress(p.read_bytes(), mtime=0))
for p in source.glob('*.stderr'):
    shutil.copyfile(p, out / p.name)
controls = out / 'controls'
controls.mkdir(exist_ok=True)
for p in sorted(source.glob('control-*.a')):
    shutil.copyfile(p, controls / p.name)
(out / 'controls.manifest').write_text(''.join('controls/' + p.name + '\n' for p in sorted(controls.glob('*.a'))))
for name, path in {'final.log': '/tmp/wave-14-resume-final.log', 'first.log': '/tmp/wave-14-resume-first.log', 'setup.log': '/tmp/wave-14-resume-setup.log', 'fetch.log': '/tmp/wave-14-resume-fetch.log', 'diffcheck.log': '/tmp/wave-14-resume-diffcheck.log'}.items():
    shutil.copyfile(path, out / name)
shutil.copytree(root / 'benchmark', out / 'benchmark', dirs_exist_ok=True)
shutil.copyfile(root / 'measure.py', out / 'measure.py')
files = ['stage1/cohere/typeaware/no_invalid_regexp.a', 'stage1/cohere/typeaware/wave_14_third_test.go', 'stage1/cohere/typeaware/wave_14_third.a', 'stage1/cohere/typeaware/testdata/oracle_wave_14_third.go']
(out / 'sources.sha256.json').write_text(json.dumps({f: hashlib.sha256((repo/f).read_bytes()).hexdigest() for f in files}, indent=2) + '\n')
refs = subprocess.check_output(['git', 'for-each-ref', '--format=%(refname) %(objectname)', 'refs/remotes/origin'], cwd=repo)
(out / 'origin-refs.txt').write_bytes(refs)
results = {}
for name in ['controls', 'compiler', 'repository']:
    digests = []
    for suffix in ['go', 'native']:
        for label in [name, name + '-asan']:
            paths = sorted(source.glob('*-' + label + '-' + suffix + '.stdout'))
            assert len(paths) == 1, (label, suffix, paths)
            data = paths[0].read_bytes()
            digest = hashlib.sha256(data).hexdigest()
            digests.append(digest)
            results[label + '-' + suffix] = {'bytes': len(data), 'sha256': digest, 'summary': data.splitlines()[-1].decode()}
    assert len(set(digests)) == 1, (name, digests)
(out / 'agreement.json').write_text(json.dumps(results, indent=2) + '\n')
print(json.dumps(results, indent=2))
