import json, os, statistics, subprocess, time
from pathlib import Path
root = Path('/workspace/wave-14-exhausted-third')
out = root / 'benchmark'
out.mkdir(exist_ok=True)
repo = Path('/workspace/adamic')
programs = {'native': root / 'third', 'go': root / 'third-oracle'}
corpora = {
    'compiler': ('/workspace/wave-14-typescript/src/compiler/tsconfig.json', '/workspace/wave-14-artifacts/compiler.manifest'),
    'repository': (str(repo / 'tsconfig.json'), '/workspace/wave-14-artifacts/repository.manifest'),
    'controls': (str(repo / 'stage1/cohere/typeaware/testdata/tsconfig.json'), str(root / 'controls.manifest')),
    'constructors': (str(repo / 'stage1/cohere/typeaware/testdata/tsconfig.json'), str(root / 'constructors.manifest')),
    'upstream': (str(repo / 'stage1/cohere/typeaware/testdata/tsconfig.json'), '/workspace/wave-14-constructor-upstream/valid.manifest'),
}
corpora['patterns']=(str(repo / 'stage1/cohere/typeaware/testdata/tsconfig.json'),str(root / 'patterns.manifest'))
results = []
medians = {}
for corpus, (config, manifest) in corpora.items():
    expected = None
    elapsed = {'native': [], 'go': []}
    for round in range(1, 4):
        for implementation in (['native', 'go'] if round % 2 else ['go', 'native']):
            stem = out / f'{corpus}-{round}-{implementation}'
            command = [str(programs[implementation]), config, manifest, '--count']
            if corpus in ['constructors', 'upstream']:
                command.append('--class-only')
            start = time.perf_counter()
            with stem.with_suffix('.stdout').open('wb') as stdout, stem.with_suffix('.stderr').open('wb') as stderr:
                subprocess.run(command, cwd=repo, env={**os.environ, 'ADAMIC_TSGO_TIMING': '1'}, stdout=stdout, stderr=stderr, check=True)
            duration = time.perf_counter() - start
            output = stem.with_suffix('.stdout').read_bytes()
            if expected is None:
                expected = output
            assert output == expected, (corpus, round, implementation, output, expected)
            elapsed[implementation].append(duration)
            results.append({'corpus': corpus, 'round': round, 'implementation': implementation, 'seconds': duration, 'command': command, 'stdout': output.decode()})
    native = statistics.median(elapsed['native'])
    go = statistics.median(elapsed['go'])
    medians[corpus] = {'nativeSeconds': native, 'goSeconds': go, 'nativeOverGo': native / go, 'findings': int(expected.decode().split()[1])}
(out / 'results.json').write_text(json.dumps(results, indent=2) + '\n')
(out / 'medians.json').write_text(json.dumps(medians, indent=2) + '\n')
print(json.dumps(medians, indent=2))
