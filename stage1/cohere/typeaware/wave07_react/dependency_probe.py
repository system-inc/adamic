#!/usr/bin/env python3
"""Check the published JSX dependency without changing shared branch files."""
import gzip
import hashlib
import json
import pathlib
import subprocess

ROOT = pathlib.Path(__file__).resolve().parents[4]
UNIT = pathlib.Path(__file__).resolve().parent
OUT = pathlib.Path('/workspace/wave-07-jsx-dependency')
REVISION = 'a8a62d62ca49db7415e14c3887dd305022b17309'
OUT.mkdir(exist_ok=True)
records = []

def git(*arguments):
    return subprocess.check_output(['git', *arguments], cwd=ROOT)

def run(name, command):
    stdout = OUT / (name + '.stdout')
    stderr = OUT / (name + '.stderr')
    with stdout.open('wb') as output, stderr.open('wb') as errors:
        result = subprocess.run([str(value) for value in command], cwd=ROOT,
                                stdout=output, stderr=errors)
    records.append({'name': name, 'command': [str(value) for value in command],
                    'exit': result.returncode, 'stdout': stdout.read_text(),
                    'stderr': stderr.read_text()})
    if result.returncode:
        raise RuntimeError(name + ' failed; see ' + str(stderr))
    return stdout.read_text()

files = git('ls-tree', '-r', '--name-only', REVISION,
            'stage1/typescript').decode().splitlines()
source_hashes = {}
for name in files:
    if not name.endswith(('.ts', '.a')):
        continue
    content = git('show', REVISION + ':' + name)
    target = OUT / name
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_bytes(content)
    source_hashes[name] = hashlib.sha256(content).hexdigest()
probe = (UNIT / 'parser_probe.a').read_text().replace(
    '../../../typescript/parser/parser.ts',
    str(OUT / 'stage1/typescript/parser/parser.ts'))
assert str(OUT / 'stage1/typescript/parser/parser.ts') in probe
(OUT / 'parser_probe.a').write_text(probe)
run('build', ['/workspace/wave-07-next-rest/adamic', 'build',
              OUT / 'parser_probe.a', '-o', OUT / 'parser'])
fixtures = [
    'function Widget(){const Inner=make();return <Inner/>;}\nexport {};\n',
    'function Widget(){const [state,setState]=useState(0);setState(1);return <div/>;}\nexport {};\n',
    'function Widget(){const [state,setState]=useState(0);useEffect(()=>{setState(1);});return <div/>;}\nexport {};\n',
    'function Widget(){return 1;}\nexport {};\n',
]
for index, source in enumerate(fixtures):
    path = OUT / ('control-' + str(index) + '.tsx')
    path.write_text(source)
    output = run('control-' + str(index), [OUT / 'parser', path])
    expected = 'jsx 1\n' if index < 3 else 'jsx 0\n'
    assert output == expected, (index, output, expected)
result = {'dependency_revision': REVISION, 'source_sha256': source_hashes,
          'fixtures': fixtures, 'commands': records,
          'scope': 'Parser capability only; no React rule implementation or comparison.'}
(OUT / 'results.json').write_text(json.dumps(result, indent=2) + '\n')
with gzip.open(UNIT / 'evidence/jsx-dependency.json.gz', 'wt') as output:
    json.dump(result, output, indent=2)
print('Published JSX dependency: three JSX controls and one ordinary control pass.')
print('No shared branch source files changed; no React rule gates run.')
